package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSendsMessage(t *testing.T) {
	var got struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot123:token/sendMessage" {
			t.Fatalf("path = %q, want Telegram sendMessage path", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient("123:token", server.URL, server.Client())
	if err := client.SendMessage(context.Background(), 42, "hello"); err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	if got.ChatID != 42 || got.Text != "hello" {
		t.Fatalf("request = (%d, %q), want (42, hello)", got.ChatID, got.Text)
	}
}
