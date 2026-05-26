package telegram

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
)

type Responder interface {
	HandleText(ctx context.Context, text string) (string, error)
}

type SendMessageFunc func(ctx context.Context, chatID int64, text string) error

type Update struct {
	Message *Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ID int64 `json:"id"`
}

func NewWebhookHandler(secret string, responder Responder, send SendMessageFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if secret != "" && !sameSecret(r.Header.Get("X-Telegram-Bot-Api-Secret-Token"), secret) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var update Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "invalid update", http.StatusBadRequest)
			return
		}
		if update.Message == nil || update.Message.Text == "" {
			w.WriteHeader(http.StatusOK)
			return
		}

		reply, err := responder.HandleText(r.Context(), update.Message.Text)
		if err != nil {
			log.Printf("handle telegram message: %v", err)
			reply = "요청을 처리하는 중 오류가 발생했습니다. 잠시 후 다시 시도해 주세요."
		}
		if err := send(r.Context(), update.Message.Chat.ID, reply); err != nil {
			log.Printf("send telegram message: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	})
}

func sameSecret(got, want string) bool {
	if len(got) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
