package redis

import (
	"encoding/json"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"

	"github.com/go-redis/redis"
)

func SendNotification(rdb *redis.Client, data NotificationRequest) {
	err := PublishMessage(rdb, NOTIFICATION_CHANNEL, data)
	if err != nil {
		logger.LogError(constants.DEFAULT_SESSION, err)
	}
}

func PublishMessage(rdb *redis.Client, channel string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return rdb.Publish(channel, jsonData).Err()
}

func ProcessNotifications(rdb *redis.Client, handler func(data NotificationRequest)) {
	SubscribeChannel(rdb, NOTIFICATION_CHANNEL, handler)
}

func SubscribeChannel(rdb *redis.Client, channel string, handler func(data NotificationRequest)) {
	sub := rdb.Subscribe(channel)
	ch := sub.Channel()

	for msg := range ch {
		var data NotificationRequest

		err := json.Unmarshal([]byte(msg.Payload), &data)
		if err != nil {
			logger.LogError(constants.DEFAULT_SESSION, err)
			continue
		}

		handler(data)
	}
}
