package general

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"

	"github.com/gin-gonic/gin"
)

func GetNotifications(ctx *gin.Context, sessionId, driverId string, page int) (notifications []postgress.NotificationRequest, totalRows int64, err error) {
	logger.LogInfo("Request received in GetNotifications", sessionId)
	logger.LogDebug("Request received in GetNotifications", sessionId, fmt.Sprintf("driverId: %s, page: %d", driverId, page))

	notifications, totalRows, err = database.GetAllNotificationByUserId(ctx, driverId, page)
	if err != nil {
		logger.LogError(sessionId, " get notifications error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetNotifications", sessionId)
	logger.LogDebug2("Response returned from GetNotifications", sessionId, fmt.Sprintf("notification count: %v, total rows: %v", len(notifications), totalRows))

	return
}

func SaveSMSFCM(ctx *gin.Context, sessionId string, request SMSFCMRequest) (err error) {
	logger.LogInfo("Request received in SaveSMSFCM", sessionId)

	err = database.DatabaseConn.Postgres.Create(mapSMSFcmRequest(request)).Error
	if err != nil {
		logger.LogError(sessionId, " get notifications error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from SaveSMSFCM", sessionId)

	return nil
}
