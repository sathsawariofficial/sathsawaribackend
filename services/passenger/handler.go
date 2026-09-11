package passenger

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

func BookSeatHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in BookSeatHandler", sessionId)

	var request BookSeatRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Unable_To_Do_Job, "book ride"),
		})
		return
	}

	logger.LogDebug2("Response received in BookSeatHandler", sessionId, request)

	err := ValidateBookSeat(sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	uuid, err := BookSeat(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to book seat error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bookingResp := bookSeatResponse(fmt.Sprintf(constants.Success_Info, "Booked the seat"), uuid)

	logger.LogInfo("Response returned from GetBookSeatsHandler", sessionId)
	logger.LogDebug2("Response returned from GetBookSeatsHandler", sessionId, bookingResp)

	ctx.JSON(http.StatusOK, bookingResp)
}

func RideRequestHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RideRequestHandler", sessionId)

	var request RideRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Unable_To_Do_Job, "broadcasted request"),
		})
		return
	}

	logger.LogDebug2("Response received in RideRequestHandler", sessionId, request)

	err := ValidateRideRequest(sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	openURL, requestId, err := RequestRide(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to save ride request error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bookingResp := rideRequestResponse(fmt.Sprintf(constants.Success_Info, "Request broadcasted"), requestId, openURL)

	logger.LogInfo("Response returned from RideRequestHandler", sessionId)
	logger.LogDebug2("Response returned from RideRequestHandler", sessionId, bookingResp)

	ctx.JSON(http.StatusOK, bookingResp)
}

func GetRideRequestsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetRideRequestsHandler", sessionId)

	request := getRideRequestFromQuery(ctx, sessionId)

	logger.LogDebug2("Response received in GetRideRequestsHandler", sessionId, request)

	err := ValidateGetRideRequest(sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rides, totalPages, err := GetRequestedRides(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to save ride request error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if totalPages == 0 {
		logger.LogError(sessionId, "no rides found error")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusNoContent,
			Message: constants.Ride_Not_Found,
		})
		return
	}

	rideRequestDetailsResp := filteredRideRequestsResp(rides, totalPages)

	logger.LogInfo("Response returned from GetRideRequestsHandler", sessionId)
	logger.LogDebug2("Response returned from GetRideRequestsHandler", sessionId, rideRequestDetailsResp)

	ctx.JSON(http.StatusOK, rideRequestDetailsResp)
}

func GetRideRequestHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetRideRequestHandler", sessionId)

	rideRequestId := ctx.Query(constants.Ride_Request_Key)
	if utils.IsStringEmpty(rideRequestId) {
		err := fmt.Errorf(constants.Not_Found, "ride")
		logger.LogError(sessionId, err)
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if !utils.IsUUID(rideRequestId) {
		var err error
		rideRequestId, err = redis.GetRedisValue(database.DatabaseConn.RedisConn, rideRequestId)
		if err != nil {
			logger.LogError(sessionId, "error: "+err.Error())
			ctx.JSON(http.StatusBadRequest, utils.APIResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
	}

	rideRequest, err := GetRequestedRide(ctx, sessionId, rideRequestId)
	if err != nil {
		logger.LogError(sessionId, "error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rideRequestResp := rideRequestDetailsResp(rideRequest)

	logger.LogInfo("Response received in GetRideRequestHandler", sessionId)
	logger.LogDebug2("Response received in GetRideRequestHandler", sessionId, rideRequestResp)

	ctx.JSON(http.StatusOK, rideRequestResp)
}

// SaveNotificationSettingsHandler saves the places a passenger's device follows. Rides
// taking in one of them are pushed to the device with their link, and are never kept.
func SaveNotificationSettingsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SaveNotificationSettingsHandler", sessionId)

	var request NotificationSettingsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Unable_To_Do_Job, "save the notification settings"),
		})
		return
	}

	logger.LogDebug2("Request received in SaveNotificationSettingsHandler", sessionId, request)

	err := ValidateNotificationSettings(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	setting, err := SaveNotificationSettings(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to save notification settings error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	settingsResp := notificationSettingsResp(setting)

	logger.LogInfo("Response returned from SaveNotificationSettingsHandler", sessionId)
	logger.LogDebug2("Response returned from SaveNotificationSettingsHandler", sessionId, settingsResp)

	ctx.JSON(http.StatusOK, settingsResp)
}

func GetNotificationSettingsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetNotificationSettingsHandler", sessionId)

	deviceId := ctx.Query(constants.Device_Key)
	err := ValidateDeviceId(deviceId)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	setting, err := GetNotificationSettings(ctx, sessionId, deviceId)
	if err != nil {
		logger.LogError(sessionId, "failed to get notification settings error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	settingsResp := notificationSettingsResp(setting)

	logger.LogInfo("Response returned from GetNotificationSettingsHandler", sessionId)
	logger.LogDebug2("Response returned from GetNotificationSettingsHandler", sessionId, settingsResp)

	ctx.JSON(http.StatusOK, settingsResp)
}
