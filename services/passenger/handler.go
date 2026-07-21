package passenger

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
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

	err = BookSeat(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to book seat error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bookingResp := utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Booked the seat"))

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
			Message: fmt.Sprintf(constants.Unable_To_Do_Job, "book ride"),
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

	err = RequestRide(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to save ride request error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bookingResp := utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Request recorded"))

	logger.LogInfo("Response returned from RideRequestHandler", sessionId)
	logger.LogDebug2("Response returned from RideRequestHandler", sessionId, bookingResp)

	ctx.JSON(http.StatusOK, bookingResp)
}

func GetRideRequestHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetRideRequestHandler", sessionId)

	request := getRideRequestFromQuery(ctx, sessionId)

	logger.LogDebug2("Response received in GetRideRequestHandler", sessionId, request)

	err := ValidateGetRideRequest(sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rides, totalPages, err := GetRequestedRide(ctx, sessionId, request)
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

	rideRequestDetailsResp := filteredRidesResp(rides, totalPages)

	logger.LogInfo("Response returned from GetRideRequestHandler", sessionId)
	logger.LogDebug2("Response returned from GetRideRequestHandler", sessionId, rideRequestDetailsResp)

	ctx.JSON(http.StatusOK, rideRequestDetailsResp)
}
