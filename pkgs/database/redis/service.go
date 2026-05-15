package redis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func SetRedisValueTTL(rdb *redis.Client, key string, value string, ttl time.Duration) (err error) {
	err = rdb.Set(key, value, ttl).Err()
	return
}

func SetRedisValue(rdb *redis.Client, key string, value string) (err error) {
	err = rdb.Set(key, value, 0).Err()
	return
}

func DeleteRedisValue(rdb *redis.Client, key string) (err error) {
	err = rdb.Del(key).Err()
	return
}

func GetRedisValue(rdb *redis.Client, key string) (value string, err error) {
	value, err = rdb.Get(key).Result()
	if err == redis.Nil {
		err = fmt.Errorf("key does not exist")
		return
	} else if err != nil {
		return
	}
	return
}
