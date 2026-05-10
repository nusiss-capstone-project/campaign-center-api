package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/lianjin/campaign-center-api/server/config"
)

var Client *redis.Client

func InitRedis() (*redis.Client, error) {
	if config.Config == nil || config.Config.RedisConfig == nil || !config.Config.RedisConfig.Enabled {
		Client = nil
		return nil, nil
	}
	Client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.Config.RedisConfig.Host, config.Config.RedisConfig.Port),
		DB:   0,
	})
	if _, err := Client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}
	return Client, nil
}
