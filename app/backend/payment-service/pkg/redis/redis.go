package redis

import (
	"context"
	"fmt"
	"log"
	"payment-service/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis() {
	cfg := config.AppConfig
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	log.Printf("Connecting to Redis: %s", addr)

	Client = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := Client.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️  Warning: Failed to connect to Redis: %v. Running in mock/console log lock mode.", err)
		Client = nil
		return
	}

	fmt.Println("✅ Successfully connected to Redis!")
}

// AcquireLock attempts to acquire a lock for a specific key
func AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	if Client == nil {
		log.Printf("[REDIS MOCK] AcquireLock: key=%s", key)
		return true, nil
	}

	success, err := Client.SetNX(ctx, "lock:"+key, "locked", expiration).Result()
	return success, err
}

// ReleaseLock releases a lock for a specific key
func ReleaseLock(ctx context.Context, key string) error {
	if Client == nil {
		log.Printf("[REDIS MOCK] ReleaseLock: key=%s", key)
		return nil
	}

	return Client.Del(ctx, "lock:"+key).Err()
}
