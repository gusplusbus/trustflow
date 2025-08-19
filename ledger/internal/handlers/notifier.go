package handlers

import (
	"bytes"
	"context"
	"net/http"
	"time"
)

type Notifier struct {
	URL        string
	HTTPClient *http.Client
}

// ForwardRaw sends the original GitHub JSON + original headers to the API.
// The API will verify X-Hub-Signature-256 using its GitHub secret.
func (n Notifier) ForwardRaw(ctx context.Context, body []byte, hdrs map[string]string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	// pass-through relevant headers exactly as received
	for k, v := range hdrs {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	// optional: identify source
	req.Header.Set("X-Relay", "ledger")

	if n.HTTPClient == nil {
		n.HTTPClient = &http.Client{Timeout: 6 * time.Second}
	}
	res, err := n.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	_ = res.Body.Close()
	if res.StatusCode/100 != 2 {
		return &httpStatusError{Code: res.StatusCode}
	}
	return nil
}

type httpStatusError struct{ Code int }
func (e *httpStatusError) Error() string { return http.StatusText(e.Code) }
