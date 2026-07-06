package passenger

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
)

func BookSeat(ctx *gin.Context, sessionId string, request BookSeatRequest) (err error) {
	logger.LogInfo("Request returned from BookSeat", sessionId)

	driverId, vehicleNumber, driverMobile, err := database.GetDriverByRideId(ctx, request.RideId)
	if err != nil {
		logger.LogError(sessionId, "failed to get driver error: "+err.Error())
		err = fmt.Errorf(constants.Not_Found, "ride")
		return err
	}

	if err := bookRide(ctx, sessionId, request); err != nil {
		logger.LogError(sessionId, "failed to book ride error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "book the seat")
		return err
	}

	message := fmt.Sprintf(constants.NOTIFICATION_MESSSGE_RIDE_BOOKED_DRIVER, request.Seats, vehicleNumber)
	utils.SendNotification(ctx, sessionId, constants.NOTIFICATION_TITLE_RIDE_BOOKED, driverId, constants.NOTIFICATION_TITLE_RIDE_BOOKED, message, map[string]string{
		constants.SMS_KEY_MOBILE_NUMBER:    driverMobile,
		constants.SMS_KEY_MESSAGE:          message,
		constants.NOTIFICATION_KEY_RIDE_ID: request.RideId,
	})

	logger.LogInfo("Response returned from BookSeat", sessionId)

	return
}
