package ledger

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	mycrypto "github.com/gusplusbus/trustflow/api/internal/crypto"
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
	if event == "issues" {
		var env map[string]any
		if err := json.Unmarshal(body, &env); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}
		// TODO: transform into issuev1.ImportIssuesRequest and call data_server
	}

	w.WriteHeader(http.StatusOK)
}
