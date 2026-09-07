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

// creates a passenger account
func RegisterPassengerHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RegisterPassengerHandler", sessionId)

	var request PassengerRegistrationRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Registeration_Failed, "passenger"),
		})
		return
	}

	logger.LogDebug2("Request received in RegisterPassengerHandler", sessionId, request)

	if err := ValidatePassengerRegistration(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	passengerId, otp, err := RegisterPassenger(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "registration error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	registrationResp := registerPassengerResp(passengerId, otp)

	logger.LogInfo("Response returned from RegisterPassengerHandler", sessionId)

	ctx.JSON(http.StatusOK, registrationResp)
}

// creates a session for a passenger
func LoginPassengerHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in LoginPassengerHandler", sessionId)

	var request PassengerLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: constants.Login_Failed,
		})
		return
	}

	logger.LogDebug2("Request received in LoginPassengerHandler", sessionId, request)

	if err := ValidatePassengerLogin(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	token, otp, passenger, err := LoginPassenger(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "login error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	loginResp := loginPassengerResp(token, otp, passenger)

	logger.LogInfo("Response returned from LoginPassengerHandler", sessionId)

	ctx.JSON(http.StatusOK, loginResp)
}

// removes the passenger session
func LogoutPassengerHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in LogoutPassengerHandler", sessionId)

	if err := LogoutPassenger(ctx, sessionId); err != nil {
		logger.LogError(sessionId, "logout error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from LogoutPassengerHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Loggedout_Successfully, "Passenger")))
}

// returns the profile of the logged in passenger
func PassengerProfileInfoHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in PassengerProfileInfoHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	if err := ValidatePassengerId(passengerId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	passenger, err := PassengerProfileInfo(ctx, sessionId, passengerId)
	if err != nil {
		logger.LogError(sessionId, "profile info error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	profileResp := passengerProfileResp(passenger)

	logger.LogInfo("Response returned from PassengerProfileInfoHandler", sessionId)

	ctx.JSON(http.StatusOK, profileResp)
}

// starts a password change, it completes once the otp is verified
func ChangePasswordHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in ChangePasswordHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	var request PassengerChangePasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "password"),
		})
		return
	}

	if err := ValidatePassengerChangePassword(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	otp, err := ChangePassword(ctx, sessionId, passengerId, request)
	if err != nil {
		logger.LogError(sessionId, "change password error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from ChangePasswordHandler", sessionId)

	ctx.JSON(http.StatusOK, passengerOTPResp(constants.SENT_OTP_Successfully, otp))
}

// sends the otp that lets a passenger reset a forgotten password
func ForgotPasswordHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in ForgotPasswordHandler", sessionId)

	mobileNumber := ctx.Query(constants.MOBILE_NUMBER_QUERY)

	if err := utils.IsValidMobileNumber(mobileNumber); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	otp, err := ForgotPassword(ctx, sessionId, mobileNumber)
	if err != nil {
		logger.LogError(sessionId, "forgot password error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from ForgotPasswordHandler", sessionId)

	ctx.JSON(http.StatusOK, passengerOTPResp(constants.SENT_OTP_Successfully, otp))
}

// deletes the passenger account and their weekly travel form
func DeletePassengerProfileHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeletePassengerProfileHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	if err := ValidatePassengerId(passengerId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := DeletePassengerProfile(ctx, sessionId, passengerId); err != nil {
		logger.LogError(sessionId, "delete profile error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DeletePassengerProfileHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Deleted_Successfully, "Passenger")))
}

// replaces the whole weekly travel form of the passenger in one call
func SetPassengerScheduleHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SetPassengerScheduleHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	var request PassengerScheduleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "schedule"),
		})
		return
	}

	logger.LogDebug2("Request received in SetPassengerScheduleHandler", sessionId, request)

	if err := ValidatePassengerSchedule(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := SetPassengerSchedule(ctx, sessionId, passengerId, request); err != nil {
		logger.LogError(sessionId, "set schedule error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from SetPassengerScheduleHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Updated_Successfully, "Schedule")))
}

// returns the weekly travel form of the passenger
func GetPassengerScheduleHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPassengerScheduleHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	preferences, err := GetPassengerSchedule(ctx, sessionId, passengerId)
	if err != nil {
		logger.LogError(sessionId, "get schedule error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	scheduleResp := passengerScheduleResp(preferences)

	logger.LogInfo("Response returned from GetPassengerScheduleHandler", sessionId)

	ctx.JSON(http.StatusOK, scheduleResp)
}
