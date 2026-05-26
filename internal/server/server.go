package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bocastle/krx-pension-bot/internal/telegram"
)

type Diagnostic interface {
	CheckKOSPI(ctx context.Context) (int, error)
}

func NewHandler(secret string, responder telegram.Responder, diagnostic Diagnostic, send telegram.SendMessageFunc) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "ok%s\n", renderCommitSuffix())
	})
	mux.HandleFunc("/debug/krx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if diagnostic == nil {
			http.Error(w, "diagnostic unavailable", http.StatusServiceUnavailable)
			return
		}
		count, err := diagnostic.CheckKOSPI(r.Context())
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if err != nil {
			http.Error(w, fmt.Sprintf("krx_error=%v", err), http.StatusBadGateway)
			return
		}
		_, _ = fmt.Fprintf(w, "kospi_rows=%d\n", count)
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
