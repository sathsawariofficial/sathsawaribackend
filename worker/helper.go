package worker

import (
	"encoding/json"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
)

func getAllActiveRides() (rides []postgress.Ride, err error) {
	err = database.DatabaseConn.Postgres.Table("rides").
		Select("*").
		Where(`rides.is_active = ?`, true).
		Find(&rides).Error

	return
}

func getAllRecurringRides() (rides []postgress.Ride, err error) {
	err = database.DatabaseConn.Postgres.
		Table("rides r").
		Where("r.is_active = ?", true).
		Where("r.is_recurring = ?", true).

		// 4-hour window condition
		Where(`
			r.start_datetime::timestamp BETWEEN NOW() AND NOW() + INTERVAL '4 hours'
		`).

		// Avoid duplicates for today
		Where(`
			NOT EXISTS (
				SELECT 1 FROM rides r2
				WHERE r2.driver_id = r.driver_id
				AND DATE(r2.start_datetime::timestamp) = CURRENT_DATE
			)
		`).
		Find(&rides).Error

	return
}

func saveNotification(notification redis.NotificationRequest) (err error) {
	notifyReq := &postgress.NotificationRequest{
		ID:               utils.GenerateUUID(),
		UserId:           notification.UserId,
		UserType:         notification.UserType,
		Title:            notification.Title,
		Message:          notification.Message,
		NotificationType: notification.NotificationType,
	}

	if notification.Data != nil {
		var bVal []byte
		bVal, err = json.Marshal(notification.Data)
		if err != nil {
			logger.LogError(notifyReq.ID, err)
			return
		}

		notifyReq.Data = string(bVal)
	}

	err = database.DatabaseConn.Postgres.Create(notifyReq).Error
	if err != nil {
		logger.LogError(constants.WROKER_SESSION, fmt.Sprintf("Error: %s, Data: %v", err.Error(), notification))
	}
	return
}
