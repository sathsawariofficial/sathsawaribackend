package general

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

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

func SaveApprochDetails(ctx *gin.Context, sessionId string, request ApprochRequest) (approchId string, err error) {
	logger.LogInfo("Request received in SaveApprochDetails", sessionId)

	// Save the approch info in the database
	approch := mapContactData(request)
	if err = database.DatabaseConn.Postgres.Create(&approch).Error; err != nil {
		logger.LogError(sessionId, "failed to create contact error: "+err.Error())
		err = fmt.Errorf(constants.Creation_Failed, "ride")
		return
	}
	approchId = approch.ID

	logger.LogInfo("Response returned from SaveApprochDetails", sessionId)
	logger.LogDebug2("Response returned from SaveApprochDetails", sessionId, approchId)

	return
}

func ResendOTP(ctx *gin.Context, sessionId, mobileNumber, operation string) (otp string, err error) {
	logger.LogInfo("Request received in ResendOTP", sessionId)

	// make sure driver exist
	driver, err := utils.GetDriver(ctx, mobileNumber)
	if err != nil {
		logger.LogError(sessionId, "get driver error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}
	if utils.IsStringEmpty(driver.ID) {
		logger.LogError(sessionId, "driver not found error")
		err = errors.New(constants.Unknown_Error)
		return
	}

	key := fmt.Sprintf("%s:%s", driver.DriverMobile, operation)
	otp, err = utils.GetOTP(ctx, sessionId, key)
	if err != nil {
		logger.LogError(sessionId, "failed to send otp: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}
	if !utils.IsStringEmpty(otp) {
		// send otp
		otp, err = utils.SendOTP(ctx, sessionId, driver.DriverMobile, operation)
		if err != nil {
			logger.LogError(sessionId, "failed to send otp: "+err.Error())
			err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
			return
		}
	}

	logger.LogInfo("Response returned from ResendOTP", sessionId)
	logger.LogDebug2("Response returned from ResendOTP", sessionId, otp)

	return
}

func VerifyOTP(ctx *gin.Context, sessionId string, request VerifyOTPRequest) (replyMessage string, err error) {
	logger.LogInfo("Request received in VerifyOTP", sessionId)

	key := fmt.Sprintf("%s:%s", request.MobileNumber, request.Operation)
	sentOTP, err := utils.GetOTP(ctx, sessionId, key)
	if err != nil {
		logger.LogError(sessionId, "failed to get otp: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "OTP")
		return
	}

	if sentOTP != request.OTP {
		err = fmt.Errorf(constants.Invalid_Data, "OTP")
		logger.LogError(sessionId, err)
		return
	}

	switch request.Operation {
	case constants.FORGOT_PASSWORD_OPERATION:
		err = updateForgottonPassword(ctx, sessionId, request.MobileNumber, request.Password, sentOTP)
		if err != nil {
			logger.LogError(sessionId, "update forgotton password error: "+err.Error())
			err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
			return
		}
		replyMessage = "Your password was resetted successfully"
	case constants.UPDATE_PASSWORD_OPERATION:
		err = updatePassword(ctx, sessionId, request.MobileNumber, sentOTP)
		if err != nil {
			logger.LogError(sessionId, "update password error: "+err.Error())
			err = fmt.Errorf(constants.Update_Failed, "password")
			return
		}
		replyMessage = "Your password was updated successfully"
	case constants.FORGOT_PIN_OPERATION:
		err = updateForgottonPin(ctx, sessionId, request.MobileNumber, request.Pin, sentOTP)
		if err != nil {
			logger.LogError(sessionId, "update forgotton pin error: "+err.Error())
			err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
			return
		}
		replyMessage = "Your pin was resetted successfully"
	case constants.UPDATE_PIN_OPERATION:
		err = updatePin(ctx, sessionId, request.MobileNumber, sentOTP)
		if err != nil {
			logger.LogError(sessionId, "update pin error: "+err.Error())
			err = fmt.Errorf(constants.Update_Failed, "pin")
			return
		}
		replyMessage = "Your pin was updated successfully"
	case constants.ACTIVATE_DRIVER_OPERATION:
		err = activateDriverByMobile(ctx, request.MobileNumber)
		if err != nil {
			logger.LogError(sessionId, "driver activation error: "+err.Error())
			err = fmt.Errorf(constants.Unable_To_Do_Job, "activate the driver")
			return
		}
		replyMessage = "Driver activated successfully"
	case constants.ACTIVATE_VEHICLE_OPERATION:
		var driver postgress.Driver
		driver, err = getActiveDriver(ctx, sessionId, request.MobileNumber)
		if err != nil {
			logger.LogError(sessionId, "vehicle activation error: "+err.Error())
			err = errors.New(constants.Driver_Not_Found)
			return
		}

		err = activateVehicleByDriverId(ctx, driver.ID)
		if err != nil {
			logger.LogError(sessionId, "vehicle activation error: "+err.Error())
			err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
			return
		}
		replyMessage = "Vehicle activated successfully"
	default:
		logger.LogError(sessionId, "invalid operation error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "operation")
		return
	}

	logger.LogInfo("Response returned from VerifyOTP", sessionId)

	return
}
