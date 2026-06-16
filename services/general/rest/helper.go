package general

import (
	"context"
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
)

func mapSMSFcmRequest(request SMSFCMRequest) *postgress.SMSFCM {
	return &postgress.SMSFCM{
		ID:      utils.GenerateUUID(),
		App:     request.App,
		FCM:     request.FCM,
		APPHash: utils.ChoiseMaker(request.AppHash, constants.DEFAULT_APP_HASH),
	}
}

func mapContactData(request ApprochRequest) postgress.ApprochInfo {
	return postgress.ApprochInfo{
		ID:      utils.GenerateUUID(),
		Name:    request.Name,
		Number:  request.Number,
		Email:   request.Email,
		Type:    request.Type,
		Message: request.Message,
	}
}

func activateVehicleByDriverId(orgCtx *gin.Context, driverId string) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	if err = database.DatabaseConn.Postgres.WithContext(ctx).Model(&postgress.Vehicle{}).Where("driver_id = ?", driverId).
		Updates(postgress.Vehicle{
			Status: constants.Status_Active,
		}).Error; err != nil {
		return
	}

	return
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

func updatePasswordByMobile(orgCtx *gin.Context, mobileNumber, password string) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	if err = database.DatabaseConn.Postgres.WithContext(ctx).Model(&postgress.Driver{}).Where("driver_mobile = ?", mobileNumber).
		Updates(postgress.Driver{
			Password: password,
		}).Error; err != nil {
		return
	}

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

func activateDriverByMobile(orgCtx *gin.Context, mobileNumber string) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	if err = database.DatabaseConn.Postgres.WithContext(ctx).Model(&postgress.Driver{}).Where("driver_mobile = ?", mobileNumber).
		Updates(postgress.Driver{
			Status: constants.Status_Active,
		}).Error; err != nil {
		return
	}

	return
}

func getActiveDriver(orgCtx *gin.Context, sessionId, mobileNumber string) (driver postgress.Driver, err error) {
	logger.LogInfo("Request recevied in getActiveDriver", sessionId)
	logger.LogDebug("Request recevied in getActiveDriver", sessionId, mobileNumber)

	driver, err = utils.GetDriver(orgCtx, mobileNumber)
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
	if driver.Status != constants.Status_Active {
		logger.LogError(sessionId, "driver is not active error")
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	logger.LogInfo("Response returned from getActiveDriver", sessionId)
	return
}
