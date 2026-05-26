package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bocastle/krx-pension-bot/internal/telegram"
)

func NewHandler(secret string, responder telegram.Responder, send telegram.SendMessageFunc) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "ok%s\n", renderCommitSuffix())
	})
	mux.Handle("/telegram/webhook", telegram.NewWebhookHandler(secret, responder, send))
	return mux
}

type SendMessageFunc func(context.Context, int64, string) error

func renderCommitSuffix() string {
	commit := os.Getenv("RENDER_GIT_COMMIT")
	if commit == "" {
		return ""
	}
	if len(commit) > 7 {
		commit = commit[:7]
	}
	return " " + commit
}
