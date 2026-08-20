package app

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Port, DatabaseURL, RedisURL, AccessSecret, RefreshSecret string
	QueueMediaPath, QueueDisplayURL                          string
	AccessTTL                                                timeDuration
	RefreshTTL                                               timeDuration
}
type timeDuration int64

func LoadConfig() (Config, error) {
	c := Config{Port: env("PORT", "4000"), DatabaseURL: os.Getenv("DATABASE_URL"), RedisURL: env("REDIS_URL", "redis://localhost:6379"), AccessSecret: os.Getenv("JWT_ACCESS_SECRET"), RefreshSecret: os.Getenv("JWT_REFRESH_SECRET"), QueueMediaPath: env("QUEUE_MEDIA_PATH", "/data/queue-media"), QueueDisplayURL: env("QUEUE_DISPLAY_URL", "http://localhost/display"), AccessTTL: timeDuration(15 * 60 * 1e9), RefreshTTL: timeDuration(30 * 24 * 60 * 60 * 1e9)}
	if c.DatabaseURL == "" {
		return c, errors.New("DATABASE_URL is required")
	}
	if len(c.AccessSecret) < 32 || len(c.RefreshSecret) < 32 || c.AccessSecret == c.RefreshSecret {
		return c, errors.New("JWT secrets must be different and at least 32 characters")
	}
	return c, nil
}
func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
