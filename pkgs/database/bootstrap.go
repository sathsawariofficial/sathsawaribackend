package database

import (
	"log"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/database/redis"
)

var DatabaseConn DatabaseConnections

func init() {
	postgressConn, err := postgress.NewPortgress()
	if err != nil {
		log.Fatal(err)
	}

	redisConn := redis.NewRedis()

	DatabaseConn.Postgres = postgressConn
	DatabaseConn.RedisConn = redisConn
}
