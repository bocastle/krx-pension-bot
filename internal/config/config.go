package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort            = "8080"
	defaultCacheTTLMinutes = 10
	minCacheTTLMinutes     = 5
	maxCacheTTLMinutes     = 30
	defaultKRXBaseURL      = "https://data.krx.co.kr"
)

type Config struct {
	Port                  string
	TelegramBotToken      string
	TelegramWebhookSecret string
	PublicBaseURL         string
	CacheTTL              time.Duration
	KRXBaseURL            string
}

func Load() (Config, error) {
	cfg := Config{
		Port:                  getenv("PORT", defaultPort),
		TelegramBotToken:      strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		TelegramWebhookSecret: strings.TrimSpace(os.Getenv("TELEGRAM_WEBHOOK_SECRET")),
		PublicBaseURL:         strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/"),
		CacheTTL:              cacheTTL(),
		KRXBaseURL:            strings.TrimRight(getenv("KRX_BASE_URL", defaultKRXBaseURL), "/"),
	}

	if cfg.TelegramBotToken == "" {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.TelegramWebhookSecret == "" {
		return Config{}, errors.New("TELEGRAM_WEBHOOK_SECRET is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func cacheTTL() time.Duration {
	minutes := defaultCacheTTLMinutes
	if raw := strings.TrimSpace(os.Getenv("CACHE_TTL_MINUTES")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			minutes = parsed
		}
	}
	if minutes < minCacheTTLMinutes {
		minutes = minCacheTTLMinutes
	}
	if minutes > maxCacheTTLMinutes {
		minutes = maxCacheTTLMinutes
	}
	return time.Duration(minutes) * time.Minute
}
