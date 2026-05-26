package telegram

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeResponder struct{}

func (fakeResponder) HandleText(ctx context.Context, text string) (string, error) {
	return "reply: " + text, nil
}

type errorResponder struct{}

func (errorResponder) HandleText(ctx context.Context, text string) (string, error) {
	return "", errors.New("krx timeout")
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

func TestWebhookRepliesWithHelpfulErrorMessage(t *testing.T) {
	var gotText string
	handler := NewWebhookHandler("secret", errorResponder{}, func(ctx context.Context, chatID int64, text string) error {
		gotText = text
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", bytes.NewBufferString(`{"message":{"chat":{"id":42},"text":"/연기금 오늘"}}`))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	for _, want := range []string{"KRX 데이터", "잠시 후", "/help"} {
		if !strings.Contains(gotText, want) {
			t.Fatalf("error reply missing %q:\n%s", want, gotText)
		}
	}
}
