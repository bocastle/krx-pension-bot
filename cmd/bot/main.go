package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bocastle/krx-pension-bot/internal/config"
	"github.com/bocastle/krx-pension-bot/internal/krx"
	"github.com/bocastle/krx-pension-bot/internal/report"
	"github.com/bocastle/krx-pension-bot/internal/server"
	"github.com/bocastle/krx-pension-bot/internal/telegram"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	kst := time.FixedZone("KST", 9*60*60)
	source := krx.NewClient(cfg.KRXBaseURL, cfg.CacheTTL)
	reports := report.NewService(source, kst)
	bot := telegram.NewClient(cfg.TelegramBotToken, "", nil)
	handler := server.NewHandler(cfg.TelegramWebhookSecret, reports, bot.SendMessage)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
}
