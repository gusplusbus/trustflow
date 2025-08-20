package timeline

import (
	"context"
	"encoding/json"
	"log"
	"time"

	issuetimelinev1 "github.com/gusplusbus/trustflow/data_server/gen/issuetimelinev1"
	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/providers/github"
	"github.com/gusplusbus/trustflow/api/internal/queue"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type checkpoint struct {
	Cursor    string    // GraphQL endCursor; empty means "start"
	UpdatedAt time.Time // informational
}

// ----- DS checkpoint I/O -----
func getCheckpoint(ctx context.Context, owner, repo string, number int, ghIssueID int64) (*checkpoint, error) {
	resp, err := clients.TimelineClient().GetCheckpoint(ctx, &issuetimelinev1.GetCheckpointRequest{
		GhIssueId: ghIssueID,
	})
	if err != nil {
		log.Printf("[timeline] GetCheckpoint error: %v (defaulting to empty)", err)
		return &checkpoint{Cursor: ""}, nil
	}
	var updated time.Time
	if ts := resp.GetUpdatedAt(); ts != nil {
		updated = ts.AsTime()
	}
	return &checkpoint{
		Cursor:    resp.GetCursor(),
		UpdatedAt: updated,
	}, nil
}

func appendBatchAndAdvance(
	ctx context.Context,
	owner, repo string, number int,
	ghIssueID int64,
	issueNodeID string,
	items []Item,
	endCursor string,
) error {
	// map Items -> proto
	pbItems := make([]*issuetimelinev1.TimelineItem, 0, len(items))
	for _, it := range items {
		payload, _ := json.Marshal(it.Payload) // safe for interface{}
		pbItems = append(pbItems, &issuetimelinev1.TimelineItem{
			Provider:        it.Provider,        // "github"
			ProviderEventId: it.ProviderEventID, // GraphQL node id
			IssueNodeId:     it.IssueNodeID,
			Type:            it.Type,
			Actor:           it.Actor,
			CreatedAt:       timestamppb.New(it.CreatedAt),
			PayloadJson:     payload,
		})
	}

	// call DS
	_, err := clients.TimelineClient().AppendBatch(ctx, &issuetimelinev1.AppendBatchRequest{
		GhIssueId: ghIssueID,
		Items:     pbItems,
		EndCursor: endCursor,
	})
	if err != nil {
		return err
	}

	log.Printf("[timeline→ds] %s/%s#%d ghID=%d wrote=%d cursor=%q",
		owner, repo, number, ghIssueID, len(pbItems), endCursor)
	return nil
}

// ----- Worker -----

func Consumer(ctx context.Context, instr queue.RefreshInstruction) {
	owner, repo, number := instr.Owner, instr.Repo, instr.Number

	// 1) installation token (API already has this plumbing)
	ver, err := github.NewVerifierFromEnv()
	if err != nil {
		log.Printf("[worker] github verifier: %v", err)
		return
	}
	tok, err := ver.InstallationTokenForRepo(ctx, owner, repo)
	if err != nil {
		log.Printf("[worker] install token: %v", err)
		return
	}

	// 2) checkpoint
	ck, err := getCheckpoint(ctx, owner, repo, number, instr.GhIssueID)
	if err != nil {
		log.Printf("[worker] checkpoint: %v", err)
		return
	}

	// 3) fetch loop
	gql := github.NewGraphQLClient(12 * time.Second)
	cursor := ck.Cursor
	pageSize := 100
	total := 0
	issueNodeID := ""

	for {
		pg, err := gql.FetchIssueTimelinePage(ctx, tok, owner, repo, number, pageSize, cursor)
		if err != nil {
			log.Printf("[worker] graphql fetch error: %v", err)
			return
		}

		if issueNodeID == "" {
			issueNodeID = pg.IssueID
		}

		// 4) normalize
		items := make([]Item, 0, len(pg.Items))
		now := time.Now().UTC()
		for _, n := range pg.Items {
			typ, _ := n["__typename"].(string)
			id, _ := n["id"].(string)
			createdAt := parseTime(n["createdAt"])
			var actor string
			if a, ok := n["actor"].(map[string]any); ok {
				if v, ok := a["login"].(string); ok {
					actor = v
				}
			}
			if a, ok := n["author"].(map[string]any); ok && actor == "" {
				if v, ok := a["login"].(string); ok {
					actor = v
				}
			}
			// keep raw-ish payload (minus fields we lifted)
			payload := map[string]any{}
			for k, v := range n {
				if k == "__typename" || k == "id" || k == "createdAt" || k == "actor" || k == "author" {
					continue
				}
				payload[k] = v
			}
			items = append(items, Item{
				Provider:        "github",
				ProviderEventID: id,
				IssueNodeID:     issueNodeID,
				Type:            typ,
				Actor:           actor,
				CreatedAt:       createdAt,
				Payload:         payload,
			})
		}

		// 5) hand to DS (atomic append + checkpoint advance)
		if err := appendBatchAndAdvance(ctx, owner, repo, number, instr.GhIssueID, issueNodeID, items, pg.EndCursor); err != nil {
			log.Printf("[worker] ds append: %v", err)
			return
		}
		total += len(items)

		// 6) loop or stop
		if !pg.HasNextPage {
			break
		}
		cursor = pg.EndCursor

		// optional safety cap per run
		if total >= 1000 {
			// re-enqueue continuation
			queue.Enqueue(queue.RefreshInstruction{
				Owner: owner, Repo: repo, Number: number,
				GhIssueID:  instr.GhIssueID,
				DeliveryID: instr.DeliveryID,
				ReceivedAt: now,
			})
			break
		}
	}
}

func parseTime(v any) time.Time {
	if s, ok := v.(string); ok {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
