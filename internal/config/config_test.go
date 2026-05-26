package config

import (
	"testing"
	"time"
)

func TestLoadReadsEnvironmentAndClampsCacheTTL(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "123:token")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "secret")
	t.Setenv("PUBLIC_BASE_URL", "https://example.com/")
	t.Setenv("CACHE_TTL_MINUTES", "2")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "8080" {
		t.Fatalf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.PublicBaseURL != "https://example.com" {
		t.Fatalf("PublicBaseURL = %q, want trimmed URL", cfg.PublicBaseURL)
	}
	if cfg.CacheTTL != 5*time.Minute {
		t.Fatalf("CacheTTL = %v, want 5m clamp", cfg.CacheTTL)
	}
}

func TestLoadRequiresTelegramToken(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "secret")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want missing token error")
	}
}
