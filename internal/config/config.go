package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
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
	_ = loadDotEnv()

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

func loadDotEnv() error {
	path, ok := findDotEnv()
	if !ok {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" || strings.TrimSpace(os.Getenv(key)) != "" {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}

func findDotEnv() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		path := filepath.Join(dir, ".env")
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
