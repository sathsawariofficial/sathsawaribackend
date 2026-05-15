package passenger

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/logger"

	"github.com/gin-gonic/gin"
)

func BookSeatDemand(ctx *gin.Context, sessionId string, request BookSeatDemandRequest) (err error) {
	logger.LogInfo("Request returned from BookSeatDemand", sessionId)

	seatBookingRequest := mapBookSeatData(request)
	if err = database.DatabaseConn.Postgres.Create(&seatBookingRequest).Error; err != nil {
		logger.LogError(sessionId, "failed to book ride error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "book the seat")

		return
	}

	logger.LogInfo("Response returned from BookSeatDemand", sessionId)
	logger.LogDebug2("Response returned from BookSeatDemand", sessionId, fmt.Sprintf("booking request Id: %v", seatBookingRequest.ID))

	return
}
