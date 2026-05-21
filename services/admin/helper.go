package admin

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

func getAdmin(orgCtx *gin.Context, mobile string) (admin postgress.Admin, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where(`username = ?`, mobile).Find(&admin).Error
	return
}

func loginTokenCreation(sessionId string, request AdminLoginRequest, admin postgress.Admin) (token string, err error) {
	if err = utils.ComparePassword(request.Password, admin.Password); err != nil {
		logger.LogError(sessionId, "invalid password error")
		err = errors.New(constants.Invalid_Password)
		return
	}

	excryptedAdminId, err := utils.EncryptAES(sessionId, admin.ID)
	if err != nil {
		logger.LogError(sessionId, "failed to set session error: "+err.Error())
		err = errors.New(constants.Login_Failed)
		return
	}

	err = redis.DeleteRedisValue(database.DatabaseConn.RedisConn, excryptedAdminId)
	if err != nil {
		logger.LogError(sessionId, "trying to delete session before new login error: "+err.Error())
	}

	token, err = utils.CreateJWT(sessionId, map[string]string{
		"adminId":   excryptedAdminId,
		"tokenType": constants.ADMIN_TOKEN,
	})
	if err != nil {
		logger.LogError(sessionId, "failed to created jwt token error: "+err.Error())
		err = errors.New(constants.Login_Failed)
		return
	}

	err = redis.SetRedisValueTTL(database.DatabaseConn.RedisConn, excryptedAdminId, token, time.Duration(configuration.ConfigurationData.Auth.ExpirationTime)*time.Second)
	if err != nil {
		logger.LogError(sessionId, "session deleted error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, "login the admin")
		return
	}

	return
}

func getAllActiveRides(orgCtx *gin.Context, page int) (rides []postgress.RideDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Table("rides").
		Select(`
			rides.id,
			rides.driver_id,
			drivers.driver_name,
			drivers.driver_mobile,
			vehicles.vehicle_number,
			vehicles.vehicle_info,
			rides.start_datetime,
			rides.estimated_end_datetime,
			rides.number_of_seats,
			rides.seats_taken,
			rides.start_location,
			rides.end_location,
			rides.fare,
			rides.route_details,
			rides.is_active,
			rides.created_at,
			rides.updated_at
		`).
		Joins("JOIN drivers ON rides.driver_id = drivers.id").
		Joins("JOIN vehicles ON rides.vehicle_id = vehicles.id").
		Where(`rides.is_active = ?`, true).
		Limit(pageSize).
		Offset(offset)

	err = query.Order("created_at desc").Find(&rides).Error
	err = query.Count(&totalRows).Error

	return
}

func getAllActiveVehicles(orgCtx *gin.Context, page int) (vehicles []postgress.Vehicle, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).
		Table("vehicles").
		Select(`
			id,
			driver_id,
			vehicle_number,
			vehicle_info,
			status,
			created_at,
			updated_at
		`).
		Where("status = ?", constants.Status_Active).
		Limit(pageSize).
		Offset(offset)

	countQuery := query

	err = query.Order("created_at desc").Find(&vehicles).Error
	if err != nil {
		return
	}

	err = countQuery.Count(&totalRows).Error
	return
}

func getAllActiveDriversWithVehicles(orgCtx *gin.Context, page int) (drivers []postgress.DriverWithVehicle, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	// Count query (NO preload)
	err = db.Model(&postgress.Driver{}).
		Where("status = ?", constants.Status_Active).
		Count(&totalRows).Error
	if err != nil {
		return
	}

	// Main query with preload
	err = db.
		Where("status = ?", "active").
		Preload("Vehicle").
		Order("created_at desc").
		Limit(pageSize).
		Offset(offset).
		Find(&drivers).Error

	return
}
