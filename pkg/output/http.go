package output

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// HTTPOutput sends logs to a remote URL via POST.
type HTTPOutput struct {
	url     string
	headers map[string]string
	client  *http.Client
}

func NewHTTPOutput(url string, headers map[string]string) *HTTPOutput {
	return &HTTPOutput{
		url:     url,
		headers: headers,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (h *HTTPOutput) WriteBatch(entries [][]byte) error {
	// 1. Join entries. For now, we use newline delimited.
	// In production, we might want a JSON array: [ "log1", "log2" ]
	joined := bytes.Join(entries, []byte("\n"))

	// 2. Create Request
	req, err := http.NewRequest("POST", h.url, bytes.NewReader(joined))
	if err != nil {
		return err
	}

	// 3. Set Headers
	req.Header.Set("Content-Type", "text/plain")
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	// 4. Send
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 5. Check Status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http output failed with status: %d", resp.StatusCode)
	}

	return nil
}
