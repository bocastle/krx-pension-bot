package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeResponder struct{}

func (fakeResponder) HandleText(ctx context.Context, text string) (string, error) {
	return "ok", nil
}

func TestHealthz(t *testing.T) {
	t.Setenv("RENDER_GIT_COMMIT", "d953cc5abcdef123456")
	handler := NewHandler("secret", fakeResponder{}, func(ctx context.Context, chatID int64, text string) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "d953cc5") {
		t.Fatalf("body = %q, want short render commit", rec.Body.String())
	}
}
