package passenger

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"

	"github.com/gin-gonic/gin"
)

func BookSeat(ctx *gin.Context, sessionId string, request BookSeatRequest) (err error) {
	logger.LogInfo("Request returned from BookSeat", sessionId)

	if err := IncrementSeatsTaken(request.RideId, request.Code); err != nil {
		logger.LogError(sessionId, "failed to book ride error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "book the seat")
		return err
	}

	logger.LogInfo("Response returned from BookSeat", sessionId)

	return
}
