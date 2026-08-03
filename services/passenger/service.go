package passenger

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
)

func BookSeat(ctx *gin.Context, sessionId string, request BookSeatRequest) (uuid string, err error) {
	logger.LogInfo("Request returned from BookSeat", sessionId)

	driverId, vehicleNumber, driverMobile, err := database.GetDriverByRideId(ctx, request.RideId)
	if err != nil {
		logger.LogError(sessionId, "failed to get driver error: "+err.Error())
		err = fmt.Errorf(constants.Not_Found, "ride")
		return uuid, err
	}

	if uuid, err = bookRide(ctx, sessionId, request); err != nil {
		logger.LogError(sessionId, "failed to book ride error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "book the seat")
		return uuid, err
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

func RequestRide(ctx *gin.Context, sessionId string, request RideRequest) (requestId, openURL string, err error) {
	logger.LogInfo("Request returned from RequestRide", sessionId)

	rideRequest := mapRideRequest(request)
	if err = database.DatabaseConn.Postgres.Create(&rideRequest).Error; err != nil {
		logger.LogError(sessionId, "failed to create ride request error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "record request")
		return
	}
	requestId = rideRequest.ID

	shortCode := utils.GenerateShortCode(rideRequest.ID)
	openURL = utils.CreateOpenRideLink(constants.LIKE_TYPE_RIDE_REQUEST_URL, shortCode)
	redis.SetRedisValue(database.DatabaseConn.RedisConn, shortCode, rideRequest.ID)

	logger.LogInfo("Response returned from RequestRide", sessionId)

	return
}

func GetRequestedRides(ctx *gin.Context, sessionId string, request GetRideRequest) (rides []postgress.RideRequest, totalPages int, err error) {
	logger.LogInfo("Request returned from GetRequestedRides", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		return
	}

	if rides, totalPages, err = getFilterAndPaginateRideRequests(ctx,
		page,
		request.StartDatetime,
		request.EstimatedEndDatetime,
		request.StartLocation,
		request.EndLocation); err != nil {
		logger.LogError(sessionId, "failed to get requested rides error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "fetch request data")
		return
	}

	logger.LogInfo("Response returned from GetRequestedRides", sessionId)

	return
}

func GetRequestedRide(ctx *gin.Context, sessionId, rideId string) (rideRequest postgress.RideRequest, err error) {
	logger.LogInfo("Request returned from GetRequestedRide", sessionId)
	logger.LogDebug("Request returned from GetRequestedRide", sessionId, rideId)

	rideRequest, err = getRideRequestByID(ctx, rideId)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get ride")
		return
	}

	logger.LogInfo("Response returned from GetRequestedRide", sessionId)
	logger.LogDebug("Response returned from GetRequestedRide", sessionId, rideRequest)

	return
}
