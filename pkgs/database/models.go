package database

import (
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type DatabaseConnections struct {
	Postgres  *gorm.DB
	RedisConn *redis.Client
}

// PlaceAlertRecipient is one device a place alert goes to, with the id of the setting it
// was found through so the next page starts after it.
type PlaceAlertRecipient struct {
	SettingId string
	FCM       string
}
