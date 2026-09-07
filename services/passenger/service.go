package passenger

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
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		err = fmt.Errorf(constants.Failed_To_Do_Job, "broadcasted request")
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

// RegisterPassenger creates a passenger account. Until now a passenger was only a
// name and a number typed into a booking, this gives them a real account so their
// standing weekly travel form can be kept and a manager can seat them on a shift.
func RegisterPassenger(ctx *gin.Context, sessionId string, request PassengerRegistrationRequest) (passengerId, otp string, err error) {
	logger.LogInfo("Request received in RegisterPassenger", sessionId)

	passenger, err := utils.GetPassenger(ctx, request.MobileNumber)
	if err != nil {
		logger.LogError(sessionId, "get passenger error: "+err.Error())
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New(constants.Unknown_Error)
			return
		}
	}
	if !utils.IsStringEmpty(passenger.ID) && passenger.Status == constants.Status_Active {
		logger.LogError(sessionId, "passenger already exist error")
		err = fmt.Errorf(constants.Unable_To_Do_Job, "register passenger")
		return
	}

	passengerId, err = savePassengerInfo(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to create passenger error: "+err.Error())
		return
	}

	otp, err = utils.SendOTP(ctx, sessionId, request.MobileNumber, constants.ACTIVATE_PASSENGER_OPERATION)
	if err != nil {
		logger.LogError(sessionId, err)
	}

	logger.LogInfo("Response returned from RegisterPassenger", sessionId)
	logger.LogDebug2("Response returned from RegisterPassenger", sessionId, fmt.Sprintf("passengerId: %v", passengerId))

	return
}

func LoginPassenger(ctx *gin.Context, sessionId string, request PassengerLoginRequest) (token, otp string, passenger postgress.Passenger, err error) {
	logger.LogInfo("Request received in LoginPassenger", sessionId)

	passenger, err = utils.GetPassenger(ctx, request.MobileNumber)
	if err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Login_Failed)
		return
	}
	if utils.IsStringEmpty(passenger.ID) {
		logger.LogError(sessionId, "passenger not found error")
		err = errors.New(constants.Passenger_Not_Found)
		return
	}
	// an account that never verified its number gets a fresh otp instead of a token
	if passenger.Status == constants.Status_PendingApproval {
		otp, err = utils.SendOTP(ctx, sessionId, request.MobileNumber, constants.ACTIVATE_PASSENGER_OPERATION)
		if err != nil {
			logger.LogError(sessionId, err)
			err = errors.New(constants.Login_Failed)
			return
		}
		return
	}
	if passenger.Status != constants.Status_Active {
		logger.LogError(sessionId, "passenger is not active, status is "+passenger.Status)
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	token, err = passengerLoginTokenCreation(sessionId, request, passenger)
	if err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Login_Failed)
		return
	}

	if !utils.IsStringEmpty(request.FCM) {
		handlePassengerFCM(ctx, sessionId, passenger.ID, request.FCM)
	}

	logger.LogInfo("Response returned from LoginPassenger", sessionId)

	return
}

func passengerLoginTokenCreation(sessionId string, request PassengerLoginRequest, passenger postgress.Passenger) (token string, err error) {
	if err = utils.ComparePassword(passenger.Password, request.Password); err != nil {
		logger.LogError(sessionId, "invalid password error")
		err = errors.New(constants.Invalid_Password)
		return
	}

	excryptedPassengerId, err := utils.EncryptAES(sessionId, passenger.ID)
	if err != nil {
		logger.LogError(sessionId, "failed to set session error: "+err.Error())
		err = errors.New(constants.Login_Failed)
		return
	}

	err = redis.DeleteRedisValue(database.DatabaseConn.RedisConn, excryptedPassengerId)
	if err != nil {
		logger.LogError(sessionId, "trying to delete session before new login error: "+err.Error())
	}

	token, err = utils.CreateJWT(sessionId, map[string]string{
		"passengerId": excryptedPassengerId,
		"tokenType":   constants.PASSENGER_TOKEN,
	})
	if err != nil {
		logger.LogError(sessionId, "failed to created jwt token error: "+err.Error())
		err = errors.New(constants.Login_Failed)
		return
	}

	err = redis.SetRedisValueTTL(database.DatabaseConn.RedisConn, excryptedPassengerId, token, time.Duration(configuration.ConfigurationData.Auth.ExpirationTime)*time.Second)
	if err != nil {
		logger.LogError(sessionId, "session deleted error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, "login the passenger")
		return
	}

	return
}

func handlePassengerFCM(ctx *gin.Context, sessionId, passengerId, fcm string) {
	logger.LogInfo("Request received in handlePassengerFCM", sessionId)

	passengerFCM, err := database.GetDriverFCM(ctx, passengerId)
	if err != nil || utils.IsStringEmpty(passengerFCM.ID) {
		if err != nil {
			logger.LogError(sessionId, err)
		}

		fcmRequest := mapPassengerFCMData(passengerId, fcm)
		if err = database.DatabaseConn.Postgres.Create(&fcmRequest).Error; err != nil {
			logger.LogError(sessionId, err)
		}
	} else if passengerFCM.FCM != fcm {
		if err = updatePassengerFCM(ctx, passengerId, fcm); err != nil {
			logger.LogError(sessionId, err)
		}
	}

	logger.LogInfo("Response returned from handlePassengerFCM", sessionId)
}

// logs the passenger out by removing the token from redis
func LogoutPassenger(ctx *gin.Context, sessionId string) (err error) {
	logger.LogInfo("Request received in LogoutPassenger", sessionId)

	passengerId := ctx.GetString(constants.Encrypted_User_KEY)

	err = redis.DeleteRedisValue(database.DatabaseConn.RedisConn, passengerId)
	if err != nil {
		logger.LogError(sessionId, "session deleted error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, "logout the passenger")
		return
	}

	logger.LogInfo("Response returned from LogoutPassenger", sessionId)

	return
}

func PassengerProfileInfo(ctx *gin.Context, sessionId, passengerId string) (passenger postgress.Passenger, err error) {
	logger.LogInfo("Request received in PassengerProfileInfo", sessionId)

	passenger, err = database.GetActivePassengerById(ctx, passengerId)
	if err != nil {
		logger.LogError(sessionId, "failed to get passenger info error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}
	if utils.IsStringEmpty(passenger.ID) {
		logger.LogError(sessionId, "passenger not found error")
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	logger.LogInfo("Response returned from PassengerProfileInfo", sessionId)

	return
}

// ChangePassword parks the new password in redis against the otp, the password is
// only really swapped once the otp comes back verified, exactly as it works for a
// driver.
func ChangePassword(ctx *gin.Context, sessionId, passengerId string, request PassengerChangePasswordRequest) (otp string, err error) {
	logger.LogInfo("Request received in ChangePassword", sessionId)

	passenger, err := database.GetActivePassengerById(ctx, passengerId)
	if err != nil {
		logger.LogError(sessionId, "get passenger error: "+err.Error())
		err = errors.New(constants.Passenger_Not_Found)
		return
	}
	if utils.IsStringEmpty(passenger.ID) {
		logger.LogError(sessionId, "passenger not found error")
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	err = utils.ComparePassword(passenger.Password, request.OldPassword)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Invalid_Data, "existing password")
		return
	}

	excryptedNewPassword, err := utils.HashPassword(sessionId, request.NewPassword)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Unable_To_Do_Job, "change password")
		return
	}

	otp, err = utils.SendOTP(ctx, sessionId, passenger.PassengerMobile, constants.PASSENGER_UPDATE_PASSWORD_OPERATION)
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

	return
}

func ForgotPassword(ctx *gin.Context, sessionId, mobileNumber string) (otp string, err error) {
	logger.LogInfo("Request received in ForgotPassword", sessionId)

	passenger, err := utils.GetPassenger(ctx, mobileNumber)
	if err != nil {
		logger.LogError(sessionId, "get passenger error: "+err.Error())
		// NOTE: we dont want to let the people know weather a number actuly exist or not
		err = nil
		return
	}
	if utils.IsStringEmpty(passenger.ID) {
		logger.LogError(sessionId, "passenger not found error")
		err = nil
		return
	}

	otp, err = utils.SendOTP(ctx, sessionId, passenger.PassengerMobile, constants.PASSENGER_FORGOT_PASSWORD_OPERATION)
	if err != nil {
		logger.LogError(sessionId, "failed to send otp: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	logger.LogInfo("Response returned from ForgotPassword", sessionId)

	return
}

func DeletePassengerProfile(ctx *gin.Context, sessionId, passengerId string) (err error) {
	logger.LogInfo("Request received in DeletePassengerProfile", sessionId)

	passenger, err := database.GetPassengerById(ctx, passengerId)
	if err != nil {
		logger.LogError(sessionId, "get passenger error: "+err.Error())
		err = errors.New(constants.Passenger_Not_Found)
		return
	}
	if utils.IsStringEmpty(passenger.ID) {
		logger.LogError(sessionId, "passenger not found error")
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	if err = deletePassengerProfile(ctx, sessionId, passenger); err != nil {
		logger.LogError(sessionId, "failed to delete passenger error: "+err.Error())
		err = fmt.Errorf(constants.DELETE_Failed, "passenger")
		return
	}

	if err = redis.DeleteRedisValue(database.DatabaseConn.RedisConn, ctx.GetString(constants.Encrypted_User_KEY)); err != nil {
		logger.LogError(sessionId, "session deleted error: "+err.Error())
		err = nil
	}

	logger.LogInfo("Response returned from DeletePassengerProfile", sessionId)

	return
}

// SetPassengerSchedule replaces the passenger's standing weekly form in one call.
// The form is what a group manager reads while deciding who to seat on which shift,
// it never creates a shift by itself.
func SetPassengerSchedule(ctx *gin.Context, sessionId, passengerId string, request PassengerScheduleRequest) (err error) {
	logger.LogInfo("Request received in SetPassengerSchedule", sessionId)

	passenger, err := database.GetActivePassengerById(ctx, passengerId)
	if err != nil || utils.IsStringEmpty(passenger.ID) {
		logger.LogError(sessionId, "get passenger error")
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	if err = savePassengerSchedule(ctx, sessionId, passengerId, request.Preferences); err != nil {
		logger.LogError(sessionId, "failed to save the schedule error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "schedule")
		return
	}

	logger.LogInfo("Response returned from SetPassengerSchedule", sessionId)

	return
}

func GetPassengerSchedule(ctx *gin.Context, sessionId, passengerId string) (preferences []postgress.PassengerLocationPreference, err error) {
	logger.LogInfo("Request received in GetPassengerSchedule", sessionId)

	preferences, err = getPassengerSchedule(ctx, passengerId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the schedule error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the schedule")
		return
	}

	logger.LogInfo("Response returned from GetPassengerSchedule", sessionId)

	return
}
