package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/driftwatch/internal/alert"
)

func TestWebhook_Notify_Success(t *testing.T) {
	received := make([]alert.WebhookPayload, 0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p alert.WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Errorf("failed to decode payload: %v", err)
		}
		received = append(received, p)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wn := alert.NewWebhook(ts.URL, 5*time.Second)
	alerts := []alert.Alert{
		{ServiceName: "svc-a", Level: alert.LevelWarning, Message: "drift in svc-a", Fields: []string{"replicas"}},
		{ServiceName: "svc-b", Level: alert.LevelCritical, Message: "drift in svc-b", Fields: []string{"image"}},
	}
	if err := wn.Notify(alerts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 2 {
		t.Fatalf("expected 2 payloads, got %d", len(received))
	}
	if received[0].Service != "svc-a" {
		t.Errorf("expected svc-a, got %s", received[0].Service)
	}
	if received[1].Level != "critical" {
		t.Errorf("expected critical, got %s", received[1].Level)
	}
}

func TestWebhook_Notify_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	wn := alert.NewWebhook(ts.URL, 5*time.Second)
	alerts := []alert.Alert{
		{ServiceName: "svc-x", Level: alert.LevelWarning, Message: "drift", Fields: []string{"env"}},
	}
	if err := wn.Notify(alerts); err == nil {
		t.Fatal("expected error for non-OK status, got nil")
	}
}

func TestWebhook_Notify_Unreachable(t *testing.T) {
	wn := alert.NewWebhook("http://127.0.0.1:0", 500*time.Millisecond)
	alerts := []alert.Alert{
		{ServiceName: "svc-y", Level: alert.LevelCritical, Message: "drift", Fields: []string{"image"}},
	}
	if err := wn.Notify(alerts); err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
}

func TestWebhook_Notify_Empty(t *testing.T) {
	wn := alert.NewWebhook("http://127.0.0.1:0", time.Second)
	if err := wn.Notify(nil); err != nil {
		t.Fatalf("expected no error for empty alerts, got %v", err)
	}
}
