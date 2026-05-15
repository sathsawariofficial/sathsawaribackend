package database

import (
	"context"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"time"

	"github.com/gin-gonic/gin"
)

func GetDriverById(orgCtx *gin.Context, driverId string) (driver postgress.Driver, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Where(`id = ?`, driverId).Where(`status = ?`, constants.Status_Active).Find(&driver).Error
	return
}

func GetDriverFCM(orgCtx *gin.Context, id string) (driverFCM postgress.DriverFCM, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Where(`driver_Id = ?`, id).Find(&driverFCM).Error
	return
}

func GetSMSFCM(orgCtx *gin.Context) (smsfcm postgress.SMSFCM, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Find(&smsfcm).Limit(1).Error
	return
}

func GetAllNotificationByUserId(orgCtx *gin.Context, userId string, page int) (notifications []postgress.NotificationRequest, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	baseQuery := DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.NotificationRequest{}).
		Where("user_id = ?", userId)

	// Count total rows BEFORE applying limit/offset
	err = baseQuery.Count(&totalRows).Error
	if err != nil {
		return
	}

	// Apply pagination
	err = baseQuery.
		Limit(pageSize).
		Offset(offset).
		Order("created_at DESC"). // usually needed for notifications
		Find(&notifications).Error

	return
}

func SaveMissingLocation(orgCtx context.Context, request postgress.MissingLocations) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Create(&request).Error
	return
}
