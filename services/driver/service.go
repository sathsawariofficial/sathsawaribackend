package driver

import (
	"errors"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// create a driver account
func RegisterDriver(ctx *gin.Context, sessionId string, request DriverRegistrationRequest) (driverId, otp string, err error) {
	logger.LogInfo("Request returned from RegisterDriver", sessionId)

	// Save the driver in the database
	var driver postgress.Driver

	driver, err = getDriver(ctx, request.MobileNumber)
	if err != nil {
		logger.LogError(sessionId, "get driver error: "+err.Error())
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New(constants.Unknown_Error)
			return
		}
	}
	if !utils.IsStringEmpty(driver.ID) && driver.Status == constants.Status_Active {
		logger.LogError(sessionId, "driver already exist error")
		err = fmt.Errorf(constants.Unable_To_Do_Job, "register driver")
		return
	}

	// Save driver info in the database
	driverId, err = saveDriverInfo(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to create driver error: "+err.Error())
		return
	}

	otp, err = sendOTP(ctx, sessionId, request.MobileNumber, constants.ACTIVATE_DRIVER_OPERATION)
	if err != nil {
		logger.LogError(sessionId, err)
	}

	logger.LogInfo("Response returned from RegisterDriver", sessionId)
	logger.LogDebug2("Response returned from RegisterDriver", sessionId, fmt.Sprintf("driverId: %v", driverId))

	return
}

func SetDriverPin(ctx *gin.Context, sessionId, driverId, pin string) (err error) {
	logger.LogInfo("Request returned from SetDriverPin", sessionId)

	err = saveDriverPin(ctx, sessionId, driverId, pin)
	if err != nil {
		logger.LogError(sessionId, "failed to set driver pin error: "+err.Error())
		return
	}

	driver, err := database.GetDriverById(ctx, driverId)
	if err != nil {
		logger.LogError(sessionId, err)
		err = nil
		return
	}

	message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PIN_CREATION, pin)
	if utils.IsStringEmpty(driver.ID) {
		utils.SendNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_PIN_CREATED, driverId, constants.NOTIFICATION_TITLE_PIN_CREATION, message, map[string]string{
			"mobileNumber": driver.DriverMobile,
			"message":      message,
		})
	}

	logger.LogInfo("Response returned from SetDriverPin", sessionId)

	return
}

func RegisterVehicle(ctx *gin.Context, sessionId string, request VehicleRegistrationRequest) (vehicleId, otp string, err error) {
	logger.LogInfo("Request returned from RegisterDriver", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	driver, err := database.GetActiveDriverById(ctx, driverId)
	if err != nil {
		logger.LogError(sessionId, "failed to get driver error: "+err.Error())
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	if !validatePin(ctx, sessionId, driver, request.Pin) {
		logger.LogError(sessionId, "error failed to validate pin")
		err = fmt.Errorf(constants.Invalid_Data, "pin")
		return
	}

	// Save the vehicle info in the database
	vehicleId, err = saveVehicleInfo(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to create vehicle error: "+err.Error())
		return
	}

	logger.LogInfo("Response returned from RegisterDriver", sessionId)
	logger.LogDebug2("Response returned from RegisterDriver", sessionId, fmt.Sprintf("vehicleId: %v, driverId: %v", vehicleId, driverId))

	return
}

func LoginDriver(ctx *gin.Context, sessionId string, request DriverLoginRequest) (token, otp string, driver postgress.Driver, err error) {
	logger.LogInfo("Request returned from LoginDriver", sessionId)

	driver, err = getDriver(ctx, request.MobileNumber)
	if err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Login_Failed)
		return
	}
	if utils.IsStringEmpty(driver.ID) {
		logger.LogError(sessionId, "driver not found error")
		err = errors.New(constants.Driver_Not_Found)
		return
	}
	if driver.Status == constants.Status_PendingApproval {
		otp, err = sendOTP(ctx, sessionId, request.MobileNumber, constants.ACTIVATE_DRIVER_OPERATION)
		if err != nil {
			logger.LogError(sessionId, err)
			err = errors.New(constants.Login_Failed)
			return
		}
		return
	}
	if driver.Status != constants.Status_Active {
		logger.LogError(sessionId, "driver is not active, status is "+driver.Status)
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	token, err = loginTokenCreation(sessionId, request, driver)
	if err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Login_Failed)
		return
	}

	if !utils.IsStringEmpty(request.FCM) {
		handleFCM(ctx, sessionId, driver.ID, request.FCM)
	}

	logger.LogInfo("Response returned from LoginDriver", sessionId)

	return
}

func handleFCM(ctx *gin.Context, sessionId, driverId, fcm string) {
	logger.LogInfo("Request returned from handleFCM", sessionId)

	driverFCM, err := database.GetDriverFCM(ctx, driverId)
	if err != nil {
		logger.LogError(sessionId, err)
		fcmRequest := mapFCMData(driverId, fcm)
		err = database.DatabaseConn.Postgres.Create(&fcmRequest).Error
		if err != nil {
			logger.LogError(sessionId, err)
		}
	} else {
		if driverFCM.FCM != fcm {
			err = updateDriverFCM(ctx, driverId, fcm)
			if err != nil {
				logger.LogError(sessionId, err)
			}
		}
	}

	logger.LogInfo("Response returned from handleFCM", sessionId)
}

func loginTokenCreation(sessionId string, request DriverLoginRequest, driver postgress.Driver) (token string, err error) {
	if err = utils.ComparePassword(driver.Password, request.Password); err != nil {
		logger.LogError(sessionId, "invalid password error")
		err = errors.New(constants.Invalid_Password)
		return
	}

	excryptedDriverId, err := utils.EncryptAES(sessionId, driver.ID)
	if err != nil {
		logger.LogError(sessionId, "failed to set session error: "+err.Error())
		err = errors.New(constants.Login_Failed)
		return
	}

	err = redis.DeleteRedisValue(database.DatabaseConn.RedisConn, excryptedDriverId)
	if err != nil {
		logger.LogError(sessionId, "trying to delete session before new login error: "+err.Error())
	}

	token, err = utils.CreateJWT(sessionId, map[string]string{
		"driverId":  excryptedDriverId,
		"tokenType": constants.DRIVER_TOKEN,
	})
	if err != nil {
		logger.LogError(sessionId, "failed to created jwt token error: "+err.Error())
		err = errors.New(constants.Login_Failed)
		return
	}

	err = redis.SetRedisValueTTL(database.DatabaseConn.RedisConn, excryptedDriverId, token, time.Duration(configuration.ConfigurationData.Auth.ExpirationTime)*time.Second)
	if err != nil {
		logger.LogError(sessionId, "session deleted error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, "login the driver")
		return
	}

	return
}

// logs driver out by removing token from redis
func LogoutDriver(ctx *gin.Context, sessionId string) (err error) {
	logger.LogInfo("Request returned from CreateRideHandler", sessionId)

	driverId := ctx.GetString(constants.Encrypted_User_KEY)

	err = redis.DeleteRedisValue(database.DatabaseConn.RedisConn, driverId)
	if err != nil {
		logger.LogError(sessionId, "session deleted error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, "logout the driver")
		return
	}
	logger.LogDebug("Logged out", sessionId, driverId)

	logger.LogInfo("Response returned from CreateRideHandler", sessionId)

	return
}

func RateDriver(ctx *gin.Context, sessionId string, request RateDriverRequest) (err error) {
	logger.LogInfo("Request returned from RateDriver", sessionId)
	// get the driver details from rideid  and save the number of persion who gave the review
	driver, err := database.GetActiveDriverById(ctx, request.DriverId)
	if err != nil {
		logger.LogError(sessionId, "get driver error: "+err.Error())
		err = errors.New(constants.Driver_Not_Found)
		return
	}
	if utils.IsStringEmpty(driver.ID) {
		logger.LogError(sessionId, "driver not found error")
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	rating, count := calculateRating(sessionId, utils.ToFloat64(driver.Rating), utils.ToInt(driver.NumberOfVotes), float64(request.Rating))
	updatedRating := fmt.Sprintf("%.2f", rating)
	logger.LogDebug("rating after 2 decimal places", sessionId, updatedRating)

	err = updateRating(ctx, request.DriverId, updatedRating, utils.ToString(count))
	if err != nil {
		logger.LogError(sessionId, "failed to update rating: %s"+err.Error())
		return
	}

	rideRating := mapRating(request)
	if err = database.DatabaseConn.Postgres.Create(&rideRating).Error; err != nil {
		logger.LogError(sessionId, "failed to save passanger rating error: "+err.Error())
		err = errors.New(constants.General_Error)
		return
	}

	logger.LogInfo("Response returned from RateDriver", sessionId)

	return
}

func DriverProfileInfo(ctx *gin.Context, sessionId, driverId string) (driverDetails postgress.DriverDetails, err error) {
	logger.LogInfo("Request returned from DriverProfileInfo", sessionId)
	logger.LogDebug("Request returned from DriverProfileInfo", sessionId, driverId)

	driverDetails, err = getDriverDetails(ctx, driverId)
	if err != nil {
		logger.LogError(sessionId, "error failed to get driver info: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	logger.LogInfo("Response returned from DriverProfileInfo", sessionId)

	return
}

func UpdateProfileStatus(ctx *gin.Context, sessionId, driverId, pin, status string) (err error) {
	logger.LogInfo("Request returned from UpdateProfileStatus", sessionId)
	logger.LogDebug("Request returned from UpdateProfileStatus", sessionId, driverId)

	if !validatePin(ctx, sessionId, postgress.Driver{ID: driverId}, pin) {
		logger.LogError(sessionId, "error failed to validate pin")
		err = fmt.Errorf(constants.Invalid_Data, "pin")
		return
	}

	err = updateDriverStatus(ctx, driverId, status)
	if err != nil {
		logger.LogError(sessionId, "error failed to update driver status: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "finish operation")
		return
	}

	logger.LogInfo("Response returned from UpdateProfileStatus", sessionId)

	return
}

func DeleteProfile(ctx *gin.Context, sessionId, driverId, pin string) (err error) {
	logger.LogInfo("Request returned from DeleteProfile", sessionId)
	logger.LogDebug("Request returned from DeleteProfile", sessionId, driverId)

	if !validatePin(ctx, sessionId, postgress.Driver{ID: driverId}, pin) {
		logger.LogError(sessionId, "error failed to validate pin")
		err = fmt.Errorf(constants.Invalid_Data, "pin")
		return
	}

	driver, err := database.GetActiveDriverById(ctx, driverId)
	if err != nil {
		logger.LogError(sessionId, "error failed to get driver status: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "find driver")
		return
	}

	if !strings.EqualFold(driver.Status, constants.Status_Active) {
		err = fmt.Errorf(constants.Failed_To_Do_Job, "delete the driver")
		logger.LogError(sessionId, "error failed to delete driver: "+err.Error())
		return
	}

	err = database.DeleteDriver(ctx, driver, driverId)
	if err != nil {
		logger.LogError(sessionId, "error failed to delete driver status: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "finish operation")
		return
	}

	logger.LogInfo("Response returned from DeleteProfile", sessionId)

	return
}

// TODO: update user profile data, save old data and verify new data via otp
func UpdateDriverData(ctx *gin.Context, sessionId string, request UpdateDriverRequest) (err error) {
	logger.LogInfo("Request returned from UpdateDriverData", sessionId)

	driver, err := getDriver(ctx, request.MobileNumber)
	if err != nil {
		logger.LogError(sessionId, "get driver error: "+err.Error())
		err = errors.New(constants.Driver_Not_Found)
		return
	}
	if utils.IsStringEmpty(driver.ID) {
		logger.LogError(sessionId, "driver not found error")
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	// validate data

	// update driver password and mobile number

	// add forgot password functionality -  use otp to validate mobile number

	logger.LogInfo("Response returned from UpdateDriverData", sessionId)

	return
}

func UpdateVehicle(ctx *gin.Context, sessionId string, request VehicleUpdateRequest) (err error) {
	logger.LogInfo("Request returned from UpdateVehicle", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	if !validatePin(ctx, sessionId, postgress.Driver{ID: driverId}, request.Pin) {
		logger.LogError(sessionId, "error failed to validate pin")
		err = fmt.Errorf(constants.Invalid_Data, "pin")
		return
	}

	// Save the vehicle info in the database
	err = updateVehicleInfo(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to update vehicle error: "+err.Error())
		return
	}

	logger.LogInfo("Response returned from UpdateVehicle", sessionId)
	logger.LogDebug2("Response returned from UpdateVehicle", sessionId, fmt.Sprintf("vehicleId: %v, driverId: %v", request.VehicleId, driverId))

	return
}

func ChangePassword(ctx *gin.Context, sessionId, driverId string, request ChangePasswordRequest) (otp string, err error) {
	logger.LogInfo("Request received in ChangePassword", sessionId)

	// make sure driver exist
	driver, err := database.GetActiveDriverById(ctx, driverId)
	if err != nil {
		logger.LogError(sessionId, "get driver error: "+err.Error())
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	err = utils.ComparePassword(driver.Password, request.OldPassword)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Invalid_Data, "existing password")
		return
	}

	// save the password
	excryptedNewPassword, err := utils.HashPassword(sessionId, request.NewPassword)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Unable_To_Do_Job, "change password")
		return
	}

	// send otp
	otp, err = sendOTP(ctx, sessionId, driver.DriverMobile, constants.UPDATE_PASSWORD_OPERATION)
	if err != nil {
		logger.LogError(sessionId, "failed to send otp: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	err = redis.SetRedisValueTTL(database.DatabaseConn.RedisConn, otp, excryptedNewPassword, time.Duration(configuration.ConfigurationData.Database.Redis.TTL)*time.Second)
	if err != nil {
		logger.LogError(sessionId, "pasword not found in cache error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	logger.LogInfo("Response returned from ChangePassword", sessionId)
	logger.LogDebug2("Response returned from ChangePassword", sessionId, otp)

	return
}

func ForgotPassword(ctx *gin.Context, sessionId, mobileNumber string) (otp string, err error) {
	logger.LogInfo("Request received in ForgotPassword", sessionId)

	// make sure driver exist
	driver, err := getDriver(ctx, mobileNumber)
	if err != nil {
		logger.LogError(sessionId, "get driver error: "+err.Error())
		// NOTE: we dont want to let the people know weather a number actuly exist or not
		err = nil
		return
	}
	if utils.IsStringEmpty(driver.ID) {
		logger.LogError(sessionId, "driver not found error")
		err = errors.New(constants.Unknown_Error)
		return
	}

	// send otp
	otp, err = sendOTP(ctx, sessionId, driver.DriverMobile, constants.FORGOT_PASSWORD_OPERATION)
	if err != nil {
		logger.LogError(sessionId, "failed to send otp: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	logger.LogInfo("Response returned from ForgotPassword", sessionId)
	logger.LogDebug2("Response returned from ForgotPassword", sessionId, otp)

	return
}

func GetVehicles(ctx *gin.Context, sessionId, driverId, status string) (vehicles []postgress.Vehicle, err error) {
	logger.LogInfo("Request received in GetVehicles", sessionId)

	vehicles, err = getActiveDriverAndVehicles(ctx, sessionId, driverId, status)
	if err != nil {
		logger.LogError(sessionId, "get vehicle error: "+err.Error())
		return
	}

	if len(vehicles) == 0 {
		logger.LogInfo(sessionId, "no vehicles found")
		err = errors.New(constants.Vehicle_Not_Found)
		return
	}

	logger.LogInfo("Response returned from GetVehicles", sessionId)
	logger.LogDebug2("Response returned from GetVehicles", sessionId, vehicles)

	return
}

func ResendOTP(ctx *gin.Context, sessionId, mobileNumber, operation string) (otp string, err error) {
	logger.LogInfo("Request received in ResendOTP", sessionId)

	// make sure driver exist
	driver, err := getDriver(ctx, mobileNumber)
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
	otp, err = getOTP(ctx, sessionId, key)
	if err != nil {
		logger.LogError(sessionId, "failed to send otp: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}
	if !utils.IsStringEmpty(otp) {
		// send otp
		otp, err = sendOTP(ctx, sessionId, driver.DriverMobile, operation)
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

func ChangePin(ctx *gin.Context, sessionId, driverId string, request ChangePinRequest) (otp string, err error) {
	logger.LogInfo("Request received in ChangePin", sessionId)

	// make sure driver exist
	driver, err := database.GetActiveDriverById(ctx, driverId)
	if err != nil {
		logger.LogError(sessionId, "get driver error: "+err.Error())
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	err = utils.ComparePassword(driver.Pin, request.OldPin)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Invalid_Data, "existing Pin")
		return
	}

	// save the Pin
	excryptedNewPin, err := utils.HashPassword(sessionId, request.NewPin)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Unable_To_Do_Job, "change Pin")
		return
	}

	// send otp
	otp, err = sendOTP(ctx, sessionId, driver.DriverMobile, constants.UPDATE_PIN_OPERATION)
	if err != nil {
		logger.LogError(sessionId, "failed to send otp: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	err = redis.SetRedisValueTTL(database.DatabaseConn.RedisConn, otp, excryptedNewPin, time.Duration(configuration.ConfigurationData.Database.Redis.TTL)*time.Second)
	if err != nil {
		logger.LogError(sessionId, "pasword not found in cache error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	logger.LogInfo("Response returned from ChangePin", sessionId)
	logger.LogDebug2("Response returned from ChangePin", sessionId, otp)

	return
}

func ForgotPin(ctx *gin.Context, sessionId, mobileNumber string) (otp string, err error) {
	logger.LogInfo("Request received in ForgotPin", sessionId)

	// make sure driver exist
	driver, err := getDriver(ctx, mobileNumber)
	if err != nil {
		logger.LogError(sessionId, "get driver error: "+err.Error())
		// NOTE: we dont want to let the people know weather a number actuly exist or not
		err = nil
		return
	}
	if utils.IsStringEmpty(driver.ID) {
		logger.LogError(sessionId, "driver not found error")
		err = errors.New(constants.Unknown_Error)
		return
	}

	// send otp
	otp, err = sendOTP(ctx, sessionId, driver.DriverMobile, constants.FORGOT_PIN_OPERATION)
	if err != nil {
		logger.LogError(sessionId, "failed to send otp: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	logger.LogInfo("Response returned from ForgotPin", sessionId)
	logger.LogDebug2("Response returned from ForgotPin", sessionId, otp)

	return
}

func VerifyOTP(ctx *gin.Context, sessionId string, request VerifyOTPRequest) (replyMessage string, err error) {
	logger.LogInfo("Request received in VerifyOTP", sessionId)

	key := fmt.Sprintf("%s:%s", request.MobileNumber, request.Operation)
	sentOTP, err := getOTP(ctx, sessionId, key)
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
		err = updateForgottonPassword(ctx, sessionId, request.MobileNumber, request.Password, sentOTP)
		if err != nil {
			logger.LogError(sessionId, "update forgotton pin error: "+err.Error())
			err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
			return
		}
		replyMessage = "Your pin was resetted successfully"
	case constants.UPDATE_PIN_OPERATION:
		err = updatePassword(ctx, sessionId, request.MobileNumber, sentOTP)
		if err != nil {
			logger.LogError(sessionId, "update pin error: "+err.Error())
			err = fmt.Errorf(constants.Update_Failed, "password")
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

func updatePassword(ctx *gin.Context, sessionId, mobileNumber, otp string) (err error) {
	logger.LogInfo("Request recevied in updatePassword", sessionId)

	excryptedPassword, err := redis.GetRedisValue(database.DatabaseConn.RedisConn, otp)
	if err != nil {
		logger.LogError(sessionId, "otp not found error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	err = updatePasswordByMobile(ctx, mobileNumber, excryptedPassword)
	if err != nil {
		logger.LogError(sessionId, "update password error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, "update the password")
		return
	}

	logger.LogInfo("Response returned from updatePassword", sessionId)

	return nil
}

func updateForgottonPassword(ctx *gin.Context, sessionId, mobileNumber, password, otp string) (err error) {
	logger.LogInfo("Request recevied in updateForgottonPassword", sessionId)

	excryptedPassword, err := utils.HashPassword(sessionId, password)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Registeration_Failed, "vehicle")
		return
	}

	err = updatePasswordByMobile(ctx, mobileNumber, excryptedPassword)
	if err != nil {
		logger.LogError(sessionId, "update forgotton password error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	logger.LogInfo("Response returned from updateForgottonPassword", sessionId)

	return nil
}

func updatePin(ctx *gin.Context, sessionId, mobileNumber, otp string) (err error) {
	logger.LogInfo("Request recevied in updatePin", sessionId)

	excryptedPin, err := redis.GetRedisValue(database.DatabaseConn.RedisConn, otp)
	if err != nil {
		logger.LogError(sessionId, "otp not found error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	err = updatePinByMobile(ctx, mobileNumber, excryptedPin)
	if err != nil {
		logger.LogError(sessionId, "update Pin error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, "update the Pin")
		return
	}

	logger.LogInfo("Response returned from updatePin", sessionId)

	return nil
}

func updateForgottonPin(ctx *gin.Context, sessionId, mobileNumber, Pin, otp string) (err error) {
	logger.LogInfo("Request recevied in updateForgottonPin", sessionId)

	excryptedPin, err := utils.HashPassword(sessionId, Pin)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Registeration_Failed, "vehicle")
		return
	}

	err = updatePinByMobile(ctx, mobileNumber, excryptedPin)
	if err != nil {
		logger.LogError(sessionId, "update forgotton Pin error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	logger.LogInfo("Response returned from updateForgottonPin", sessionId)

	return nil
}
