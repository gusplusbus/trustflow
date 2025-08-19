package webhook

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gusplusbus/trustflow/ledger/internal/config"
	mycrypto "github.com/gusplusbus/trustflow/ledger/internal/crypto"
	"github.com/gusplusbus/trustflow/ledger/internal/dataserver"
	"github.com/gusplusbus/trustflow/ledger/internal/handlers"
	gh "github.com/gusplusbus/trustflow/ledger/internal/github"
)

type githubHandler struct {
	cfg     config.Config
	checker dataserver.Checker
	notify  handlers.Notifier
}

func NewGitHubHandler(cfg config.Config) http.Handler {
  checker, err := dataserver.NewGRPCChecker(cfg.DataServerGRPCAddr, 900*time.Millisecond)
  if err != nil { log.Fatalf("checker: %v", err) }
  return githubHandler{
    cfg: cfg,
    checker: checker,
    notify: handlers.Notifier{ URL: cfg.APIURL, HTTPClient: &http.Client{Timeout: cfg.APITimeout} },
  }

}

func (h githubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1) read body once
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	// 2) verify GitHub HMAC
	got := r.Header.Get("X-Hub-Signature-256")
	if !mycrypto.VerifyGitHubSignature(h.cfg.GitHubWebhookSecret, body, got) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// 3) ACK immediately (non-blocking)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	// 4) capture headers we’ll forward to API
	hdrs := map[string]string{
		"X-GitHub-Event":      r.Header.Get("X-GitHub-Event"),
		"X-GitHub-Delivery":   r.Header.Get("X-GitHub-Delivery"),
		"X-Hub-Signature-256": got,
		"Content-Type":        r.Header.Get("Content-Type"),
	}

	// 5) background processing
	go h.process(body, hdrs)
}

func (h githubHandler) process(body []byte, hdrs map[string]string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event := hdrs["X-GitHub-Event"]
	delivery := hdrs["X-GitHub-Delivery"]

	// We handle issues.* only for now
	if event != "issues" {
		log.Printf("[ledger] drop event=%s delivery=%s (unsupported for now)", event, delivery)
		return
	}

	me, err := gh.ParseIssuesEvent(delivery, event, body)
	if err != nil {
		log.Printf("[ledger] parse error delivery=%s: %v", delivery, err)
		return
	}
	log.Printf("[ledger] received event=%s delivery=%s repo=%s/%s number=%d id=%d",
		me.Event, me.Delivery, me.Owner, me.Repo, me.Number, me.GHIssueID)

	managed, err := h.checker.IsManaged(ctx, me.Owner, me.Repo, me.GHIssueID, me.Number)
	if err != nil {
		log.Printf("[ledger] check error delivery=%s: %v", delivery, err)
		return
	}
	if !managed {
		log.Printf("[ledger] ignored unmanaged repo/issue delivery=%s", delivery)
		return
	}

	// forward ORIGINAL body + ORIGINAL GitHub headers to API
	if err := h.notify.ForwardRaw(ctx, body, hdrs); err != nil {
		log.Printf("[ledger] forward error delivery=%s: %v", delivery, err)
		return
	}
	log.Printf("[ledger] forwarded delivery=%s -> API ok", delivery)
}
