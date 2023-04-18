package utils

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	"go.uber.org/zap"
)

func IncreaseAndSetExpire(
	ctx context.Context,
	rdb redis.UniversalClient,
	key string,
	value int64,
	expireTime time.Duration,
	logger *zap.Logger,
) {
	result, err := rdb.IncrBy(ctx, key, value).Result()
	if err != nil {
		logger.Error("failed to save key:", zap.Any(key, err))
		return
	}
	if result == value && expireTime > 0 {
		if err != rdb.Expire(ctx, key, expireTime).Err() {
			logger.Error("failed to save key expire:", zap.Any(key, err))
		}
	}
}
