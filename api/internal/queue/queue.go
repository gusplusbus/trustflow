package queue

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"
)

type RefreshInstruction struct {
	Owner       string // github org
	Repo        string // github repo
	Number      int    // github issue number
	GhIssueID   int64  // github "database id" (from webhook)
	DeliveryID  string // X-GitHub-Delivery (trace/dedupe upstream if you want)
	ReceivedAt  time.Time
}

type consumer func(ctx context.Context, instr RefreshInstruction)

var (
	ch    chan RefreshInstruction
	alive = make(chan struct{})
)

func Enqueue(instr RefreshInstruction) {
	select {
	case ch <- instr:
	default:
		// prevent back-pressure on webhook handler; drop into log if full
		log.Printf("[queue] instruction queue full; dropping delivery=%s %s/%s#%d",
			instr.DeliveryID, instr.Owner, instr.Repo, instr.Number)
	}
}

func Start(ctx context.Context, consume consumer) {
	size := 1024
	if s := stringsEnv("QUEUE_SIZE", "1024"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			size = n
		}
	}
	ch = make(chan RefreshInstruction, size)

	workers := 4
	if s := stringsEnv("WORKERS", "4"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			workers = n
		}
	}

	log.Printf("[queue] starting %d workers (cap=%d)", workers, size)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer func() { recover() }()
			for {
				select {
				case <-ctx.Done():
					return
				case instr := <-ch:
					safeConsume(ctx, consume, instr, id)
				}
			}
		}(i + 1)
	}
	close(alive) // signal started
}

func WaitStarted() { <-alive }

func safeConsume(ctx context.Context, c consumer, instr RefreshInstruction, wid int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[worker %d] panic: %v", wid, r)
		}
	}()
	c(ctx, instr)
}

func stringsEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
