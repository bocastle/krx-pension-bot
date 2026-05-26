package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadReadsDotEnvFileWithoutOverridingExistingEnv(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, ".env"), []byte(strings.Join([]string{
		"TELEGRAM_BOT_TOKEN=from-file",
		"TELEGRAM_WEBHOOK_SECRET=file-secret",
		"PUBLIC_BASE_URL=https://file.example.com/",
		"CACHE_TTL_MINUTES=20",
	}, "\n")), 0o600)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Chdir(dir)
	t.Setenv("TELEGRAM_BOT_TOKEN", "from-env")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "")
	t.Setenv("PUBLIC_BASE_URL", "")
	t.Setenv("CACHE_TTL_MINUTES", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.TelegramBotToken != "from-env" {
		t.Fatalf("TelegramBotToken = %q, want existing env value", cfg.TelegramBotToken)
	}
	if cfg.TelegramWebhookSecret != "file-secret" {
		t.Fatalf("TelegramWebhookSecret = %q, want file value", cfg.TelegramWebhookSecret)
	}
	if cfg.PublicBaseURL != "https://file.example.com" {
		t.Fatalf("PublicBaseURL = %q, want trimmed file value", cfg.PublicBaseURL)
	}
	if cfg.CacheTTL != 20*time.Minute {
		t.Fatalf("CacheTTL = %v, want 20m", cfg.CacheTTL)
	}
}
