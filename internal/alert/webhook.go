package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookPayload is the JSON body sent to the webhook endpoint.
type WebhookPayload struct {
	Service string   `json:"service"`
	Level   string   `json:"level"`
	Message string   `json:"message"`
	Fields  []string `json:"fields"`
}

// WebhookNotifier sends alerts to an HTTP webhook.
type WebhookNotifier struct {
	URL    string
	client *http.Client
}

// NewWebhook creates a WebhookNotifier targeting url.
func NewWebhook(url string, timeout time.Duration) *WebhookNotifier {
	return &WebhookNotifier{
		URL:    url,
		client: &http.Client{Timeout: timeout},
	}
}

// Notify POSTs each alert to the configured webhook URL.
// It returns the first error encountered, if any.
func (wn *WebhookNotifier) Notify(alerts []Alert) error {
	for _, al := range alerts {
		payload := WebhookPayload{
			Service: al.ServiceName,
			Level:   string(al.Level),
			Message: al.Message,
			Fields:  al.Fields,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("alert: marshal payload for %q: %w", al.ServiceName, err)
		}
		resp, err := wn.client.Post(wn.URL, "application/json", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("alert: post to webhook for %q: %w", al.ServiceName, err)
		}
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("alert: webhook returned status %d for service %q", resp.StatusCode, al.ServiceName)
		}
	}
	return nil
}
