package database

import (
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type DatabaseConnections struct {
	Postgres  *gorm.DB
	RedisConn *redis.Client
}
