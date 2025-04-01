package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// Add these helper functions for Redis operations
func getProductFromRedis(ctx context.Context, asin string) (*SimplifiedResponse, error) {
	key := RedisKeyPrefix + asin
	data, err := redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("product not found in Redis")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get product from Redis: %v", err)
	}
	var simplifiedResponse SimplifiedResponse
	err = json.Unmarshal(data, &simplifiedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product from Redis: %v", err)
	}
	return &simplifiedResponse, nil
}

func saveProductToRedis(ctx context.Context, asin string, simplifiedResponse *SimplifiedResponse) error {
	key := RedisKeyPrefix + asin
	data, _ := json.Marshal(simplifiedResponse)
	return redisClient.Set(ctx, key, data, RedisTTL).Err()
}
