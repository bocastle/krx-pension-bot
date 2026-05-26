package telegram

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeResponder struct{}

func (fakeResponder) HandleText(ctx context.Context, text string) (string, error) {
	return "reply: " + text, nil
}

func TestWebhookRejectsWrongSecret(t *testing.T) {
	handler := NewWebhookHandler("secret", fakeResponder{}, func(ctx context.Context, chatID int64, text string) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", bytes.NewBufferString(`{"message":{"chat":{"id":1},"text":"/start"}}`))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "wrong")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestWebhookRepliesToMessage(t *testing.T) {
	var gotChatID int64
	var gotText string
	handler := NewWebhookHandler("secret", fakeResponder{}, func(ctx context.Context, chatID int64, text string) error {
		gotChatID = chatID
		gotText = text
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", bytes.NewBufferString(`{"message":{"chat":{"id":42},"text":"/start"}}`))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if gotChatID != 42 || gotText != "reply: /start" {
		t.Fatalf("reply = (%d, %q), want (42, reply: /start)", gotChatID, gotText)
	}
}
