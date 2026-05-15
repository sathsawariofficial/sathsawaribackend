package redis

import (
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"time"

	"github.com/go-redis/redis"
)

func NewRedis() (rdb *redis.Client) {
	redisConf := configuration.ConfigurationData.Database.Redis

	client := redis.NewClient(&redis.Options{
		Addr:     redisConf.Host + ":" + redisConf.Port,
		Password: redisConf.Password,
		DB:       redisConf.Database,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		PoolTimeout:  4 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	if client == nil {
		logger.LogFatal(constants.DEFAULT_SESSION, "failed to connect to redis")
	}

	return client
}
