package ledger

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	mycrypto "github.com/gusplusbus/trustflow/api/internal/crypto"
	"github.com/gusplusbus/trustflow/api/internal/queue"
)

func HandleNotify(w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "read error", http.StatusBadRequest)
    return
  }
  _ = r.Body.Close()

  // verify signature again
  secret := []byte(os.Getenv("GITHUB_WEBHOOK_SECRET"))
  sig := r.Header.Get("X-Hub-Signature-256")
  if !mycrypto.VerifyGitHubSignature(secret, body, sig) {
    http.Error(w, "unauthorized", http.StatusUnauthorized)
    return
  }

  // parse event type
  event := r.Header.Get("X-GitHub-Event")
  delivery := r.Header.Get("X-GitHub-Delivery")
  log.Printf("[api] got event=%s delivery=%s", event, delivery)

  // For now, handle issues only

  // For now, handle issues only
  if event == "issues" {
    var env map[string]any
    if err := json.Unmarshal(body, &env); err != nil {
      http.Error(w, "bad payload", http.StatusBadRequest)
      return
    }
    // Extract owner/repo/number/id from GH webhook
    var (
      owner string
      repo  string
      num   int
      ghid  int64
    )
    if repoMap, ok := env["repository"].(map[string]any); ok {
      if o, ok := repoMap["owner"].(map[string]any); ok {
        if l, _ := o["login"].(string); l != "" {
          owner = l
        }
      }
      if r, _ := repoMap["name"].(string); r != "" {
        repo = r
      }
    }
    if iss, ok := env["issue"].(map[string]any); ok {
      if n, ok := iss["number"].(float64); ok {
        num = int(n)
      }
      if id, ok := iss["id"].(float64); ok {
        ghid = int64(id)
      }
    }
    if owner == "" || repo == "" || num <= 0 {
      http.Error(w, "missing repo/issue identifiers", http.StatusBadRequest)
      return
    }
    // Enqueue refresh instruction & ACK fast
    queue.Enqueue(queue.RefreshInstruction{
      Owner:      owner,
      Repo:       repo,
      Number:     num,
      GhIssueID:  ghid,
      DeliveryID: delivery,
      ReceivedAt: time.Now().UTC(),
    })
  }
  w.WriteHeader(http.StatusAccepted) // ACK fast; worker runs async
}
