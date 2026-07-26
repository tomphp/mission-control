package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPReporter reports events to the mission-control server over loopback
// HTTP. It is deliberately configured with a short timeout: reporting must
// never add perceptible latency to a Claude Code hook, and must fail fast
// (not hang) if the server isn't running.
type HTTPReporter struct {
	baseURL string
	client  *http.Client
}

// NewHTTPReporter builds an HTTPReporter posting to baseURL + "/ingest",
// bounded by timeout.
func NewHTTPReporter(baseURL string, timeout time.Duration) *HTTPReporter {
	return &HTTPReporter{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

func (r *HTTPReporter) Report(ctx context.Context, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("transport: encode event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/ingest", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("transport: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("transport: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("transport: unexpected status %d", resp.StatusCode)
	}
	return nil
}
