package worker

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"time"
)

func CloseActiveRides() {
	sessionId := constants.WROKER_SESSION
	logger.LogInfo("Request received in CloseActiveRides", sessionId)

	rides, err := getAllActiveRides()
	if err != nil {
		logger.LogError(sessionId, "failed to get ride error: "+err.Error())
		return
	}

	now := time.Now().UTC()

	for _, ride := range rides {
		endTime, err := utils.ConvertStrToTime(ride.EstimatedEndDatetime)
		if err != nil {
			logger.LogError(sessionId, "invalid datetime for ride "+fmt.Sprint(ride.ID))
			continue
		}

		if endTime.Before(now) || endTime.Equal(now) {
			logger.LogDebug2("closing ride", sessionId, ride.ID)

			database.DatabaseConn.Postgres.
				Model(&postgress.Ride{}).
				Where("id = ?", ride.ID).
				Update("is_active", false)
		}
	}
}

func ProcessNotifications() {
	redis.ProcessNotifications(database.DatabaseConn.RedisConn, func(data redis.NotificationRequest) {
		var err error
		if data.NotificationType != constants.NOTIFICATION_TYPE_SMS_TO_SERVICE {
			err = saveNotification(data)
		}
		if err == nil {
			err = utils.SendPush(data.Token, data.Title, data.Message, data.NotificationType, data.Data)
			if err != nil {
				logger.LogError(constants.WROKER_SESSION, err)
			}
		}
		logger.LogError(constants.WROKER_SESSION, err)
	})
}
