package redis

import (
	"context"
	"downloader_gochat/configs"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func ConnectRedis() {
	time.Sleep(time.Duration(configs.GetConfigs().WaitForRedisConnectionSec) * time.Second)
	redisClient = redis.NewClient(&redis.Options{
		Addr:     configs.GetConfigs().RedisUrl,
		Password: configs.GetConfigs().RedisPassword,
		DB:       0,
	})
	ctx := context.Background()
	pong, err := redisClient.Ping(ctx).Result()
	fmt.Println("====> [[GoChat Redis Client:", pong, err, "]]")
}

func GetRedis(ctx context.Context, key string) (interface{}, error) {
	val, err := redisClient.Get(ctx, key).Result()
	return val, err
}

func SetRedis(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	err := redisClient.Set(ctx, key, value, duration).Err()
	return err
}
