package redis

import (
	"context"
	"github.com/qthang02/goticket/config"
	"github.com/qthang02/goticket/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return nil, err
	}

	logger.Info("Connected to Redis")
	return client, nil
}
