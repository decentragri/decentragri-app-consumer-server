package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var ctx = context.Background()

// InitRedis initializes the Redis connection
func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")
	db := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if parsedDB, err := strconv.Atoi(dbStr); err == nil {
			db = parsedDB
		}
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Warning: Failed to connect to Redis: %v\n", err)
		fmt.Println("Server will continue without caching. Install and start Redis for optimal performance.")
		RedisClient = nil
		return
	}

	fmt.Println("Connected to Redis successfully")
}

// Set stores a value in Redis with expiration
func Set(key string, value interface{}, expiration time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not available")
	}
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, jsonValue, expiration).Err()
}

// Get retrieves a value from Redis and unmarshals it
func Get(key string, dest interface{}) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not available")
	}
	value, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), dest)
}

// Delete removes a key from Redis
func Delete(key string) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not available")
	}
	return RedisClient.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis
func Exists(key string) bool {
	if RedisClient == nil {
		return false
	}
	result, _ := RedisClient.Exists(ctx, key).Result()
	return result > 0
}
