package ride

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

// create ride
func CreateRideHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in CreateRideHandler", sessionId)

	var request RideCreationRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Creation_Failed, "ride"),
		})
		return
	}

	logger.LogDebug2("Response received in CreateRideHandler", sessionId, request)

	driverId := ctx.GetString(constants.User_KEY)
	err := ValidateRideCreation(sessionId, driverId, &request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}
	request.EXTDriverId = driverId

	rideId, openURL, err := CreateRide(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "ride creation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rideResp := createRideResp(rideId, openURL)

	logger.LogInfo("Response returned from CreateRideHandler", sessionId)
	logger.LogDebug2("Response returned from CreateRideHandler", sessionId, rideResp)

	ctx.JSON(http.StatusOK, rideResp)
}

// create a driver account
func GetFilteredRidesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetFilteredRidesHandler", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	startTime, endTime, searchLoc, _, _ := getQueryParams(ctx)

	err = ValidateFilteredRides(page, startTime, endTime, searchLoc)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rides, totalRows, err := GetFilteredRides(ctx, sessionId, page, startTime, endTime, searchLoc)
	if err != nil {
		logger.LogError(sessionId, "get filtered rides error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if totalRows == 0 {
		logger.LogError(sessionId, "no rides found error")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusNoContent,
			Message: constants.Ride_Not_Found,
		})
		return
	}

	rideDetailsResp := filteredRidesResp(rides, totalRows)

	logger.LogInfo("Response returned from GetFilteredRidesHandler", sessionId)
	logger.LogDebug2("Response returned from GetFilteredRidesHandler", sessionId, rideDetailsResp)

	ctx.JSON(http.StatusOK, rideDetailsResp)
}

// driver's rides
func DriverRideHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DriverRideHandler", sessionId)

	rideStatus, _ := ctx.GetQuery(constants.Status_Key)
	driverId := ctx.GetString(constants.User_KEY)

	startTime, endTime, _, startLoc, endLoc := getQueryParams(ctx)

	err := ValidateDriverRide(driverId, rideStatus)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	totalPages, rides, err := DriverRide(ctx, sessionId, rideStatus, driverId, startTime, endTime, startLoc, endLoc)
	if err != nil {
		logger.LogError(sessionId, "driver rides error: "+err.Error())
		if err.Error() == constants.Ride_Not_Found {
			ctx.JSON(http.StatusBadRequest, utils.APIResponse{
				Code:    http.StatusNoContent,
				Message: err.Error(),
				Data: DriverRideResponse{
					Rides: []RideDetails{},
				},
			})
		} else {
			ctx.JSON(http.StatusBadRequest, utils.APIResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
				Data: DriverRideResponse{
					Rides: []RideDetails{},
				},
			})
		}
		return
	}

	driverRideResp := getDriverRidesResp(rides, totalPages, driverId)

	logger.LogInfo("Response returned from DriverRideHandler", sessionId)
	logger.LogDebug2("Response returned from DriverRideHandler", sessionId, driverRideResp)

	ctx.JSON(http.StatusOK, driverRideResp)
}

// update the ride status
func UpdateRideHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateRideHandler", sessionId)

	rideId, _ := ctx.GetQuery(constants.Ride_Key)

	var request UpdateRideRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: constants.General_Error,
		})
		return
	}

	logger.LogDebug2("Response received in UpdateRideHandler", sessionId, request)

	err := ValidateUpdateRide(rideId, request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = UpdateRide(ctx, sessionId, rideId, request)
	if err != nil {
		logger.LogError(sessionId, "driver rides error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})

		return
	}

	driverRideResp := updateRideResp()

	logger.LogInfo("Response returned from UpdateRideHandler", sessionId)
	logger.LogDebug2("Response returned from UpdateRideHandler", sessionId, driverRideResp)

	ctx.JSON(http.StatusOK, driverRideResp)
}

func BookSeatByDriverHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in BookSeatByDriverHandler", sessionId)

	var request BookSeatRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Creation_Failed, "ride"),
		})
		return
	}

	logger.LogDebug2("Response received in BookSeatByDriverHandler", sessionId, request)

	err := ValidateBookRide(sessionId, &request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = BookSeatByDriver(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "ride creation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bookingResp := utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Ride booked"))

	logger.LogInfo("Response returned from BookSeatByDriverHandler", sessionId)
	logger.LogDebug2("Response returned from BookSeatByDriverHandler", sessionId, bookingResp)

	ctx.JSON(http.StatusOK, bookingResp)
}

func GetBookSeatsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetBookSeatsHandler", sessionId)

	rideId := ctx.GetString(constants.Ride_Key)

	logger.LogDebug2("Response received in GetBookSeatsHandler", sessionId, rideId)

	err := utils.ValidateId(rideId)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bookedSeats, err := GetBookedSeats(ctx, sessionId, rideId)
	if err != nil {
		logger.LogError(sessionId, "failed to get booked seats error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bookingResp := getBookedSeatsResp(bookedSeats)

	logger.LogInfo("Response returned from GetBookSeatsHandler", sessionId)
	logger.LogDebug2("Response returned from GetBookSeatsHandler", sessionId, bookingResp)

	ctx.JSON(http.StatusOK, bookingResp)
}

func UpdateBookSeatHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateBookSeatHandler", sessionId)

	seatId := ctx.GetString(constants.Seat_Key)

	logger.LogDebug2("Response received in UpdateBookSeatHandler", sessionId, seatId)

	err := utils.ValidateId(seatId)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = UpdateBookedSeat(ctx, sessionId, seatId)
	if err != nil {
		logger.LogError(sessionId, "failed to cancel booked seat error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bookingResp := utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Cancelled the seat"))

	logger.LogInfo("Response returned from GetBookSeatsHandler", sessionId)
	logger.LogDebug2("Response returned from GetBookSeatsHandler", sessionId, bookingResp)

	ctx.JSON(http.StatusOK, bookingResp)
}

func GetRideHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetRideHandler", sessionId)

	rideId := ctx.Query(constants.Ride_Key)
	if utils.IsStringEmpty(rideId) {
		err := fmt.Errorf(constants.Not_Found, "ride")
		logger.LogError(sessionId, err)
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if !utils.IsUUID(rideId) {
		var err error
		rideId, err = redis.GetRedisValue(database.DatabaseConn.RedisConn, rideId)
		if err != nil {
			logger.LogError(sessionId, "error: "+err.Error())
			ctx.JSON(http.StatusBadRequest, utils.APIResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
	}

	ride, childRides, err := GetRide(ctx, sessionId, rideId)
	if err != nil {
		logger.LogError(sessionId, "error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rideResp := getRideResp(sessionId, ride, childRides)

	logger.LogInfo("Response received in GetRideHandler", sessionId)
	logger.LogDebug2("Response received in GetRideHandler", sessionId, rideResp)

	ctx.JSON(http.StatusOK, rideResp)
}

func GetOpenRideHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetOpenRideHandler", sessionId)

	code := ctx.Param("id")
	Id, err := redis.GetRedisValue(database.DatabaseConn.RedisConn, code)
	if err != nil || utils.IsStringEmpty(Id) {
		if err == nil {
			err = fmt.Errorf(constants.Not_Found, "ride")
		}

		logger.LogError(sessionId, err)

		redirectURL := constants.APP_BASE_URL + "?page=expired"
		ctx.Redirect(http.StatusFound, redirectURL)
		return
	}

	var redirectURL string

	if strings.HasPrefix(ctx.Request.URL.Path, "/rq/") {
		redirectURL = fmt.Sprintf(
			"%s?page=seatrequests&request_id=%s",
			constants.APP_BASE_URL,
			Id,
		)
	} else {
		redirectURL = fmt.Sprintf(
			"%s?page=reserve&ride_id=%s",
			constants.APP_BASE_URL,
			Id,
		)
	}

	logger.LogInfo("Redirecting user", sessionId)
	logger.LogDebug2("Redirect URL", sessionId, redirectURL)

	ctx.Redirect(http.StatusFound, redirectURL)
}

func GetRideTemplatesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetRideTemplatesHandler", sessionId)

	rideTemplates, err := GetRideTemplates(ctx, sessionId)
	if err != nil {
		logger.LogError(sessionId, "error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	templateResp := getRideTemplatesResp(sessionId, rideTemplates)

	logger.LogInfo("Response received in GetRideTemplatesHandler", sessionId)
	logger.LogDebug2("Response received in GetRideTemplatesHandler", sessionId, templateResp)

	ctx.JSON(http.StatusOK, templateResp)
}

func DeleteRideTemplatesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeleteRideTemplatesHandler", sessionId)

	rideTemplateId := ctx.Query(constants.Ride_Template_Key)

	err := utils.ValidateId(rideTemplateId)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = DeleteRideTemplate(ctx, sessionId, rideTemplateId)
	if err != nil {
		logger.LogError(sessionId, "error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	templateResp := utils.GeneralSuccessResp(constants.Success)

	logger.LogInfo("Response received in DeleteRideTemplatesHandler", sessionId)
	logger.LogDebug2("Response received in DeleteRideTemplatesHandler", sessionId, templateResp)

	ctx.JSON(http.StatusOK, templateResp)
}
