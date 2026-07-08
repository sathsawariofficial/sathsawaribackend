package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func closeActiveRides() {
	sessionId := constants.WROKER_SESSION
	logger.LogInfo("Request received in CloseActiveRides", sessionId)

	defer func() {
		if r := recover(); r != nil {
			logger.LogError(sessionId, fmt.Errorf("panic recovered: %v", r))
		}
	}()

	rides, err := getAllActiveRides()
	if err != nil {
		logger.LogError(sessionId, "failed to get ride error: "+err.Error())
		return
	}
	if len(rides) == 0 {
		logger.LogWarning(sessionId, "no ride found")
		return
	}

	now := time.Now()

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

func getAllActiveRides() (rides []postgress.Ride, err error) {
	err = database.DatabaseConn.Postgres.
		Where(`is_active = ?`, true).
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
		ID:               database.GenerateUUID(),
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

func processBroadcastBatch(ginCtx *gin.Context, sessionId string, batchSize int) {
	var requests []postgress.BroadcastNotificationRequests

	ctx, cancel := context.WithTimeout(
		ginCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	err := database.DatabaseConn.Postgres.
		WithContext(ctx).
		Where("user_type = ? AND processed = false", constants.User_Driver).
		Order("created_at asc").
		Limit(batchSize).
		Find(&requests).Error

	if err != nil {
		logger.LogError(sessionId, "failed to fetch broadcast requests: "+err.Error())
		return
	}

	if len(requests) == 0 {
		return
	}

	for _, req := range requests {
		processBroadcastRequest(ginCtx, sessionId, req)

		// mark as processed (IMPORTANT to avoid reprocessing every tick)
		database.DatabaseConn.Postgres.
			Model(&postgress.BroadcastNotificationRequests{}).
			Where("id = ?", req.ID).
			Update("processed", true)
	}
}

func processBroadcastRequest(ctx *gin.Context, sessionId string, req postgress.BroadcastNotificationRequests) {

	switch req.UserType {

	case constants.User_Driver:
		sendDriverBroadcast(ctx, sessionId, req)

	default:
		logger.LogInfo("unimplemented user type in broadcast: "+fmt.Sprint(req.UserType), sessionId)
	}
}

func sendDriverBroadcast(ginCtx *gin.Context, sessionId string, req postgress.BroadcastNotificationRequests) {
	const driverBatchSize = 50
	offset := 0

	for {

		var drivers []postgress.Driver

		ctx, cancel := context.WithTimeout(
			ginCtx,
			time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
		)

		err := database.DatabaseConn.Postgres.
			WithContext(ctx).
			Limit(driverBatchSize).
			Offset(offset).
			Where(`status = ?`, constants.Status_Active).
			Find(&drivers).Error

		cancel()

		if err != nil {
			logger.LogError(sessionId, "failed to fetch drivers: "+err.Error())
			return
		}

		if len(drivers) == 0 {
			break
		}

		for _, driver := range drivers {
			utils.SendNotification(
				ginCtx,
				sessionId,
				req.NotificationType,
				driver.ID,
				req.Title,
				req.Message,
				map[string]string{
					"driverId":                driver.ID,
					"title":                   req.Title,
					constants.SMS_KEY_MESSAGE: req.Message,
				},
			)
		}

		offset += driverBatchSize
	}
}
