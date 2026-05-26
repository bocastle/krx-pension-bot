package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeResponder struct{}

func (fakeResponder) HandleText(ctx context.Context, text string) (string, error) {
	return "ok", nil
}

func TestHealthz(t *testing.T) {
	handler := NewHandler("secret", fakeResponder{}, func(ctx context.Context, chatID int64, text string) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "ok\n" {
		t.Fatalf("body = %q, want ok", rec.Body.String())
	}
}
