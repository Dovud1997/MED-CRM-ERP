package app

import (
	"context"
	"log/slog"

	"clinicos/api/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	cfg   Config
	log   *slog.Logger
	db    *pgxpool.Pool
	redis *redis.Client
	auth  *auth.Service
}

func New(ctx context.Context, cfg Config, log *slog.Logger) (*App, error) {
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		db.Close()
		return nil, err
	}
	rdb := redis.NewClient(opt)
	if err = rdb.Ping(ctx).Err(); err != nil {
		db.Close()
		return nil, err
	}
	svc := auth.New(db, rdb, []byte(cfg.AccessSecret), []byte(cfg.RefreshSecret), int64(cfg.AccessTTL), int64(cfg.RefreshTTL))
	return &App{cfg: cfg, log: log, db: db, redis: rdb, auth: svc}, nil
}
func (a *App) Close() { _ = a.redis.Close(); a.db.Close() }
