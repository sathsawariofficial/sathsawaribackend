package driver

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

// creates a driver account
func RegisterDriverHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RegisterDriverHandler", sessionId)

	var request DriverRegistrationRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Registeration_Failed, "driver"),
		})
		return
	}

	logger.LogDebug2("Response received in RegisterDriverHandler", sessionId, request)

	err := ValidateDriverRegistration(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	driverId, otp, err := RegisterDriver(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "registration error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	registrationResp := registerDriverResp(driverId, otp)

	logger.LogInfo("Response received in RegisterDriverHandler", sessionId)
	logger.LogDebug2("Response received in RegisterDriverHandler", sessionId, registrationResp)

	ctx.JSON(http.StatusOK, registrationResp)
}

func SetDriverPinHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SetDriverPinHandler", sessionId)

	driverId := ctx.Query(constants.Driver_KEY)
	pin := ctx.Query(constants.Pin_Key)

	logger.LogDebug2("Response received in SetDriverPinHandler", sessionId, pin)

	err := ValidateDriverProfileInfo(driverId)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = ValidatePin(pin)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = SetDriverPin(ctx, sessionId, driverId, pin)
	if err != nil {
		logger.LogError(sessionId, "set driver pin error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response received in SetDriverPinHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Driver pin set")))
}

// registers a vehicle in the name of a driver
func RegisterVehicleHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RegisterVehicleHandler", sessionId)

	var request VehicleRegistrationRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Registeration_Failed, "vehicle"),
		})
		return
	}

	logger.LogDebug2("Response received in RegisterVehicleHandler", sessionId, request)

	err := ValidateVehicleRegistration(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	vehicleId, otp, err := RegisterVehicle(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "registration error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	registrationResp := registerVehicleResp(vehicleId, otp)

	logger.LogInfo("Response received in RegisterVehicleHandler", sessionId)
	logger.LogDebug2("Response received in RegisterVehicleHandler", sessionId, registrationResp)

	ctx.JSON(http.StatusOK, registrationResp)
}

// create a session for driver
func LoginDriverHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in LoginDriverHandler", sessionId)

	var request DriverLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: constants.General_Error,
		})
		return
	}

	logger.LogDebug2("Request received in LoginDriverHandler", sessionId, request)

	err := ValidateDriverLogin(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	driverSessionId, otp, driver, err := LoginDriver(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "login error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	driverLoginResp := loginDriverResp(driverSessionId, otp, driver)

	logger.LogInfo("Response received in LoginDriverHandler", sessionId)
	logger.LogDebug2("Response received in LoginDriverHandler", sessionId, driverLoginResp)

	if !utils.IsStringEmpty(otp) {
		ctx.JSON(http.StatusAccepted, driverLoginResp)
	} else {
		ctx.JSON(http.StatusOK, driverLoginResp)
	}
}

// deletes the driver session
func LogoutDriverHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in LogoutDriverHandler", sessionId)

	err := LogoutDriver(ctx, sessionId)
	if err != nil {
		logger.LogError(sessionId, "logout error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logoutResp := logoutResp()

	logger.LogInfo("Response received in LogoutDriverHandler", sessionId)
	logger.LogDebug2("Response received in LogoutDriverHandler", sessionId, logoutResp)

	ctx.JSON(http.StatusOK, logoutResp)
}

func RateDriverHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RateDriverHandler", sessionId)

	var request RateDriverRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: constants.General_Error,
		})
		return
	}

	logger.LogDebug2("Response received in RateDriverHandler", sessionId, request)

	err := ValidateRateDriver(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = RateDriver(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "registration error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	ratingResp := rateDriverResp()

	logger.LogInfo("Response received in RateDriverHandler", sessionId)
	logger.LogDebug2("Response received in RateDriverHandler", sessionId, ratingResp)

	ctx.JSON(http.StatusOK, ratingResp)
}

func DriverProfileInfoHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DriverProfileInfoHandler", sessionId)

	driverId := ctx.GetString(constants.Driver_KEY)
	err := ValidateDriverProfileInfo(driverId)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	driverDetails, err := DriverProfileInfo(ctx, sessionId, driverId)
	if err != nil {
		logger.LogError(sessionId, err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	ratingResp := dirverProfileResp(driverDetails)

	logger.LogInfo("Response received in DriverProfileInfoHandler", sessionId)
	logger.LogDebug2("Response received in DriverProfileInfoHandler", sessionId, ratingResp)

	ctx.JSON(http.StatusOK, ratingResp)
}

func UpdateProfileStatusHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateProfileStatusHandler", sessionId)

	driverId := ctx.GetString(constants.Driver_KEY)
	pin := ctx.Query(constants.Pin_Key)
	err := ValidateDriverProfileInfo(driverId)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	status := ctx.Query("status")
	err = ValidateUpdateProfileStatus(status)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = UpdateProfileStatus(ctx, sessionId, driverId, pin, status)
	if err != nil {
		logger.LogError(sessionId, err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response received in UpdateProfileStatusHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
	})
}

func UpdateVehicleHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateVehicleHandler", sessionId)

	var request VehicleUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "vehicle"),
		})
		return
	}

	logger.LogDebug2("Response received in UpdateVehicleHandler", sessionId, request)

	err := ValidateVehicleUpdate(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = UpdateVehicle(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "update error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	registrationResp := updateVehicleResp(request.VehicleId)

	logger.LogInfo("Response received in UpdateVehicleHandler", sessionId)
	logger.LogDebug2("Response received in UpdateVehicleHandler", sessionId, registrationResp)

	ctx.JSON(http.StatusOK, registrationResp)
}

func DeleteDriverProfileHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeleteDriverProfileHandler", sessionId)

	driverId := ctx.GetString(constants.Driver_KEY)
	pin := ctx.Query(constants.Pin_Key)

	err := ValidateDriverProfileInfo(driverId)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = ValidatePin(pin)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}
	status := constants.Status_InActive

	err = UpdateProfileStatus(ctx, sessionId, driverId, pin, status)
	if err != nil {
		logger.LogError(sessionId, err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response received in DeleteDriverProfileHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
	})
}

// reset's drivers password
func ChangePasswordHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in ChangePasswordHandler", sessionId)

	var request ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Unable_To_Do_Job, "change password"),
		})
		return
	}

	logger.LogDebug2("Response received in ChangePasswordHandler", sessionId, request)

	driverId := ctx.GetString(constants.Driver_KEY)
	err := ValidateChangePassword(driverId, &request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	otp, err := ChangePassword(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "reset password error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resetResp := changePasswordResp(otp)

	logger.LogInfo("Response received in ChangePasswordHandler", sessionId)
	logger.LogDebug2("Response received in ChangePasswordHandler", sessionId, resetResp)

	ctx.JSON(http.StatusOK, resetResp)
}

// recover driver's password
func ForgotPasswordHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in ForgotPasswordHandler", sessionId)

	mobileNumber := utils.HandleMobileNumberInQuery(ctx.Query(constants.MOBILE_NUMBER_QUERY))

	logger.LogDebug2("Response received in ForgotPasswordHandler", sessionId, mobileNumber)

	err := ValidateForgotPassword(mobileNumber)
	if err != nil {
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

	resetResp := forgotPasswordResp(otp)

	logger.LogInfo("Response received in ChangePasswordHandler", sessionId)
	logger.LogDebug2("Response received in ChangePasswordHandler", sessionId, resetResp)

	ctx.JSON(http.StatusOK, resetResp)
}

func GetVehicleHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetVehicleHandler", sessionId)

	driverId := ctx.GetString(constants.Driver_KEY)
	status := ctx.Query(constants.Status_Key)
	err := ValidateDriverProfileInfo(driverId)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	vehicleDetails, err := GetVehicles(ctx, sessionId, driverId, status)
	if err != nil {
		logger.LogError(sessionId, err.Error())
		if err.Error() == constants.Ride_Not_Found {
			ctx.JSON(http.StatusBadRequest, utils.APIResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
				Data: VehiclesResponse{
					Vehicles: []Vehicles{},
				},
			})
		} else {
			ctx.JSON(http.StatusBadRequest, utils.APIResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
				Data: VehiclesResponse{
					Vehicles: []Vehicles{},
				},
			})
		}
		return
	}

	vehicleResp := vehicleInfoResp(vehicleDetails)

	logger.LogInfo("Response received in GetVehicleHandler", sessionId)
	logger.LogDebug2("Response received in GetVehicleHandler", sessionId, vehicleResp)

	ctx.JSON(http.StatusOK, vehicleResp)
}

func ResendOTPHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in ResendOTPHandler", sessionId)

	mobileNumber := ctx.Query(constants.MOBILE_NUMBER_QUERY)

	logger.LogDebug2("Response received in ResendOTPHandler", sessionId, mobileNumber)

	err := ValidateSendOTP(mobileNumber)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	otp, err := ResendOTP(ctx, sessionId, mobileNumber)
	if err != nil {
		logger.LogError(sessionId, "send otp error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resetResp := sendOTPResp(otp)

	logger.LogInfo("Response received in ResendOTPHandler", sessionId)
	logger.LogDebug2("Response received in ResendOTPHandler", sessionId, resetResp)

	ctx.JSON(http.StatusOK, resetResp)
}

func SendOTPHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SendOTPHandler", sessionId)

	driverId := ctx.GetString(constants.Driver_KEY)

	logger.LogDebug2("Response received in SendOTPHandler", sessionId, driverId)

	otp, err := SendOTP(ctx, sessionId, driverId)
	if err != nil {
		logger.LogError(sessionId, "send otp error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resetResp := sendOTPResp(otp)

	logger.LogInfo("Response received in SendOTPHandler", sessionId)
	logger.LogDebug2("Response received in SendOTPHandler", sessionId, resetResp)

	ctx.JSON(http.StatusOK, resetResp)
}

// verify otp
func VerifyOTPHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in VerifyOTPHandler", sessionId)

	var request VerifyOTPRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Registeration_Failed, "driver"),
		})
		return
	}

	logger.LogDebug2("Response received in VerifyOTPHandler", sessionId, request)

	err := ValidateOTP(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	replyMessage, err := VerifyOTP(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "verify otp error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp := utils.GeneralSuccessResp(replyMessage)

	logger.LogInfo("Response received in VerifyOTPHandler", sessionId)
	logger.LogDebug2("Response received in VerifyOTPHandler", sessionId, resp)

	ctx.JSON(http.StatusOK, resp)
}
