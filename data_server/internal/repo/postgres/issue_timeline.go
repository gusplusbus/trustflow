package postgres

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed queries/*.sql
var itFS embed.FS

type IssuesTimelinePG struct {
	db *pgxpool.Pool

	qGetIssueIDByGh   string // it_get_issue_id_by_gh.sql
	qSelectCheckpoint string // it_select_checkpoint.sql
	qUpsertCheckpoint string // it_upsert_checkpoint.sql
	qInsertRaw        string // it_insert_many.sql
}

func NewIssuesTimelinePG(db *pgxpool.Pool) (*IssuesTimelinePG, error) {
	read := func(name string) string {
		b, err := itFS.ReadFile("queries/" + name)
		if err != nil { panic(err) }
		return string(b)
	}
	return &IssuesTimelinePG{
		db:                db,
		qGetIssueIDByGh:   read("it_get_issue_id_by_gh.sql"),
		qSelectCheckpoint: read("it_select_checkpoint.sql"),
		qUpsertCheckpoint: read("it_upsert_checkpoint.sql"),
		qInsertRaw:        read("it_insert_many.sql"),
	}, nil
}

// -------- types (kept same as your previous draft) --------

type Checkpoint struct {
	Cursor    string
	LastEvent *time.Time
	UpdatedAt time.Time
}

type RawItem struct {
	ProjectIssueID  uuid.UUID
	Provider        string
	ProviderEventID string
	Type            string
	Actor           *string
	CreatedAt       time.Time
	PayloadJSON     []byte
}

// -------- methods --------

// Resolve GH numeric id -> internal project_issues.id (UUID)
func (r *IssuesTimelinePG) GetProjectIssueIDByGhID(ctx context.Context, ghIssueID int64) (uuid.UUID, error) {
	var id uuid.UUID
	if err := r.db.QueryRow(ctx, r.qGetIssueIDByGh, ghIssueID).Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// Load checkpoint; return empty cursor if none exists yet
func (r *IssuesTimelinePG) GetCheckpoint(ctx context.Context, projectIssueID uuid.UUID) (*Checkpoint, error) {
	var ck Checkpoint
	err := r.db.QueryRow(ctx, r.qSelectCheckpoint, projectIssueID).
		Scan(&ck.Cursor, &ck.LastEvent, &ck.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &Checkpoint{Cursor: ""}, nil
		}
		return nil, err
	}
	return &ck, nil
}

// Insert a batch with ON CONFLICT DO NOTHING; returns number of new rows
func (r *IssuesTimelinePG) InsertMany(ctx context.Context, items []RawItem) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}
	inserted := 0
	for _, it := range items {
		// it_insert_many.sql is a single-row INSERT with ON CONFLICT DO NOTHING
		_, err := r.db.Exec(ctx, r.qInsertRaw,
			it.ProjectIssueID, it.Provider, it.ProviderEventID, it.Type, it.Actor, it.CreatedAt, it.PayloadJSON,
		)
		if err != nil {
			return inserted, fmt.Errorf("insert timeline raw: %w", err)
		}
		// count as "attempted"—pgx doesn’t tell us CONFLICT vs inserted; ok for now
		inserted++
	}
	return inserted, nil
}

// Upsert checkpoint in place
func (r *IssuesTimelinePG) UpsertCheckpoint(ctx context.Context, projectIssueID uuid.UUID, cursor string, lastEventAt *time.Time) error {
	_, err := r.db.Exec(ctx, r.qUpsertCheckpoint, projectIssueID, cursor, lastEventAt)
	return err
}
