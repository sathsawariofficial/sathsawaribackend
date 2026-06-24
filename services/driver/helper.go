package driver

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
	"gorm.io/gorm/clause"
)

func mapDriverData(request DriverRegistrationRequest) postgress.Driver {
	return postgress.Driver{
		ID:           utils.GenerateUUID(),
		Status:       constants.Status_PendingApproval,
		DriverMobile: request.MobileNumber,
		DriverName:   request.Name,
		Password:     request.EXTPassword,
	}
}

func mapDriverDataWithPin(driver postgress.Driver, pin string) postgress.Driver {
	return postgress.Driver{
		ID:           utils.GenerateUUID(),
		Status:       constants.Status_PendingApproval,
		DriverMobile: driver.DriverMobile,
		DriverName:   driver.DriverName,
		Password:     driver.Password,
		Pin:          pin,
	}
}

func mapVehicleData(request VehicleRegistrationRequest, driverId string) postgress.Vehicle {
	return postgress.Vehicle{
		ID:            utils.GenerateUUID(),
		DriverId:      driverId,
		VehicleNumber: request.VehicleNumber,
		VehicleInfo:   request.VehicleInfo,
		Status:        constants.Status_Active,
	}
}

func mapVehicleUpdateData(
	request VehicleUpdateRequest,
	existingVehicle postgress.Vehicle,
	driverId string,
) postgress.Vehicle {
	updated := existingVehicle

	// Only update if value is provided (non-empty)
	if request.VehicleNumber != "" {
		updated.VehicleNumber = request.VehicleNumber
	}

	if request.VehicleInfo != "" {
		updated.VehicleInfo = request.VehicleInfo
	}

	if request.Status != "" {
		updated.Status = request.Status
		if request.Status == constants.Status_InActive {
			updated.VehicleNumber = fmt.Sprintf("DEL_%v_%s", time.Now(), updated.VehicleNumber)
		}
	}

	return updated
}

func mapFCMData(driverId, fcm string) postgress.DriverFCM {
	return postgress.DriverFCM{
		ID:       utils.GenerateUUID(),
		DriverId: driverId,
		FCM:      fcm,
	}
}

func updateDeviceId(orgCtx *gin.Context, driverId, deviceId string) error {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	var driverDevices []postgress.DriverDevice
	err := database.DatabaseConn.Postgres.WithContext(ctx).Where(`driver_id = ?`, driverId).Find(&driverDevices).Error
	if err != nil {
		return err
	}

	var found bool
	for _, driverDevice := range driverDevices {
		if driverDevice.DeviceId == deviceId {
			found = true
			break
		}
	}

	if !found {
		err := database.DatabaseConn.Postgres.WithContext(ctx).Create(postgress.DriverDevice{
			DriverId: driverId,
			DeviceId: deviceId,
		}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func updateDriverFCM(orgCtx *gin.Context, id, fcm string) error {
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	err := database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.DriverFCM{}).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "driver_id"}}, // conflict key
			DoUpdates: clause.AssignmentColumns([]string{"fcm"}),
		}).
		Create(&postgress.DriverFCM{
			DriverId: id,
			FCM:      fcm,
		}).Error

	return err
}

func updateDriverStatus(orgCtx *gin.Context, driverId, status string) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	return database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Driver{}).
		Where("id = ?", driverId).
		Update("status", status).
		Error
}

func getActiveVehiclesByDriverId(orgCtx *gin.Context, driverId, status string) (vehicles []postgress.Vehicle, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Where(`driver_id = ?`, driverId)
	if !utils.IsStringEmpty(status) {
		query = query.Where(`status = ?`, status)
	}
	err = query.Find(&vehicles).Error
	return
}

// init a db transaction saves driver data
func saveDriverInfo(orgCtx *gin.Context, sessionId string, request DriverRegistrationRequest) (driverId string, err error) {
	logger.LogInfo("Request received in saveDriverInfo", sessionId)

	// Begin a new transaction
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	excryptedPassword, err := utils.HashPassword(sessionId, request.Password)
	if err != nil {
		logger.LogError(sessionId, err)
		tx.Rollback() // Roll back the transaction if there is an error
		err = fmt.Errorf(constants.Invalid_Data, "password")
		return
	}
	request.EXTPassword = excryptedPassword

	// Save the driver
	driver := mapDriverData(request)
	if err = tx.Create(&driver).Error; err != nil {
		logger.LogError(sessionId, err)
		tx.Rollback() // Roll back the transaction if there is an error
		err = fmt.Errorf(constants.Registeration_Failed, "driver")
		return
	}
	driverId = driver.ID

	err = tx.Create(&postgress.DriverDevice{
		DriverId: driver.ID,
		DeviceId: request.DeviceId,
	}).Error
	if err != nil {
		logger.LogError(sessionId, err)
		tx.Rollback() // Roll back the transaction if there is an error
		err = fmt.Errorf(constants.Registeration_Failed, "driver")
		return
	}

	// Commit the transaction if all operations succeed
	if err = tx.Commit().Error; err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from saveDriverInfo", sessionId)

	return driverId, nil
}

func saveDriverPin(orgCtx *gin.Context, sessionId, driverId, pin string) (err error) {
	driver, err := database.GetActiveDriverById(orgCtx, driverId)
	if err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	if !utils.IsStringEmpty(driver.Pin) {
		err = errors.New("pin in already set")
		logger.LogError(sessionId, err)
		return
	}

	excryptedPin, err := utils.HashPassword(sessionId, pin)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Invalid_Data, "pin")
		return
	}

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	if err := database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&driver).
		Where("id = ?", driver.ID).
		Updates(map[string]interface{}{
			"pin": excryptedPin,
		}).Error; err != nil {

		logger.LogError(sessionId, err)
		return fmt.Errorf(constants.Failed_To_Do_Job, "add pin")
	}

	return
}

// init a db transaction saves vehicle
func saveVehicleInfo(orgCtx *gin.Context, sessionId, driverId string, request VehicleRegistrationRequest) (vehicleId string, err error) {
	logger.LogInfo("Request received in saveVehicleInfo", sessionId)

	// Begin a new transaction
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Save the vehicle
	vehicle := mapVehicleData(request, driverId)
	if err = tx.Create(&vehicle).Error; err != nil {
		logger.LogError(sessionId, err)
		tx.Rollback() // Roll back the transaction if there is an error
		err = fmt.Errorf(constants.Registeration_Failed, "vehicle")
		return
	}
	vehicleId = vehicle.ID

	// Commit the transaction if all operations succeed
	if err = tx.Commit().Error; err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from saveVehicleInfo", sessionId)

	return vehicleId, nil
}

func updateVehicleInfo(orgCtx *gin.Context, sessionId, driverId string, request VehicleUpdateRequest) error {
	logger.LogInfo("Request received in updateVehicleInfo", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingVehicle postgress.Vehicle
	if err := tx.Where("id = ? AND driver_id = ? AND status = ?",
		request.VehicleId,
		driverId,
		constants.Status_Active).First(&existingVehicle).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return errors.New(constants.Vehicle_Not_Found)
	}

	updatedVehicle := mapVehicleUpdateData(request, existingVehicle, driverId)

	updatedVehicle.ID = existingVehicle.ID
	updatedVehicle.CreatedAt = existingVehicle.CreatedAt

	if err := tx.Model(&existingVehicle).Updates(updatedVehicle).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return fmt.Errorf(constants.Update_Failed, "vehicle")
	}

	if err := tx.Commit().Error; err != nil {
		logger.LogError(sessionId, err)
		return errors.New(constants.Unknown_Error)
	}

	logger.LogInfo("Response returned from updateVehicleInfo", sessionId)

	return nil
}

func mapRating(request RateDriverRequest) postgress.PassengerRating {
	return postgress.PassengerRating{
		ID:                    utils.GenerateUUID(),
		DriverID:              request.DriverId,
		RideID:                request.RideId,
		PassengerMobileNumber: request.MobileNumber,
	}
}

func updateRating(orgCtx *gin.Context, driverId, updatedRating, updatedCount string) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	if err = database.DatabaseConn.Postgres.WithContext(ctx).Model(&postgress.Driver{}).Where("id = ?", driverId).
		Updates(postgress.Driver{
			Rating:        updatedRating,
			NumberOfVotes: updatedCount,
		}).Error; err != nil {
		return
	}

	return
}

func calculateRating(sessionId string, oldAverage float64, oldCount int, newRating float64) (newAverage float64, newCount int) {
	newAverage = (oldAverage*float64(oldCount) + newRating) / float64(oldCount+1)
	newCount = oldCount + 1
	logger.LogDebug("Response returned from calculateRating", sessionId, fmt.Sprintf("new rating: %v, new count: %v", newAverage, newCount))
	return
}

func getDriverDetails(orgCtx *gin.Context, driverId string) (response postgress.DriverDetails, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Table("drivers").
		Select(`
			drivers.id,
			vehicles.id AS vehicle_id,
			drivers.driver_mobile,
			drivers.driver_name,
			drivers.rating,
			CASE
				WHEN drivers.pin IS NOT NULL AND drivers.pin <> '' THEN true
				ELSE false
			END AS has_pin
		`).
		Joins("LEFT JOIN vehicles ON drivers.id = vehicles.driver_id").
		Where("drivers.id = ?", driverId).
		Scan(&response).Error

	return
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

func updatePinByMobile(orgCtx *gin.Context, mobileNumber, Pin string) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	if err = database.DatabaseConn.Postgres.WithContext(ctx).Model(&postgress.Driver{}).Where("driver_mobile = ?", mobileNumber).
		Updates(postgress.Driver{
			Pin: Pin,
		}).Error; err != nil {
		return
	}

	return
}

func getActiveDriverAndVehicles(orgCtx *gin.Context, sessionId, driverId, status string) (vehicles []postgress.Vehicle, err error) {
	logger.LogInfo("Request recevied in getActiveDriverAndVehicles", sessionId)
	logger.LogDebug("Request recevied in getActiveDriverAndVehicles", sessionId, fmt.Sprintf("driverId: %s, status: %s", driverId, status))

	var driver postgress.Driver
	driver, err = database.GetActiveDriverById(orgCtx, driverId)
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

	vehicles, err = getActiveVehiclesByDriverId(orgCtx, driverId, status)
	if err != nil {
		logger.LogError(sessionId, "get vehicle error: "+err.Error())
		err = errors.New(constants.Vehicle_Not_Found)
		return
	}

	logger.LogInfo("Response returned from getActiveDriverAndVehicles", sessionId)
	return
}

func validatePin(orgCtx *gin.Context, sessionId string, driver postgress.Driver, pin string) bool {
	driverPin := driver.Pin
	if utils.IsStringEmpty(driverPin) {
		driver, err := database.GetActiveDriverById(orgCtx, driver.ID)
		if err != nil {
			logger.LogError(sessionId, err)
			err = fmt.Errorf(constants.Invalid_Data, "pin")
			return false
		}
		if utils.IsStringEmpty(driver.Pin) {
			err = errors.New("please set pin for driver " + driver.ID)
			logger.LogError(sessionId, err)
			return false
		}
		driverPin = driver.Pin
	}

	if err := utils.ComparePassword(driverPin, pin); err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Invalid_Data, "pin")
		return false
	}

	logger.LogDebug("pin is valid", sessionId, driver.ID)

	return true
}

func bookRide(orgCtx *gin.Context, sessionId, driverID string, request BookSeatRequest) error {
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()

	result := tx.Exec(`
		UPDATE rides
		SET seats_taken = seats_taken + $1
		WHERE id = $2
		  AND driver_id = $3
		  AND is_active = true
		  AND seats_taken + $1 <= number_of_seats
	`, request.Seats, request.RideId, driverID)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("ride not found or insufficient seats available")
	}

	if err := tx.Create(&postgress.RideBooking{
		ID:           utils.GenerateUUID(),
		RideID:       request.RideId,
		MobileNumber: request.MobileNumber,
		Name:         request.Name,
		Seats:        request.Seats,
	}).Error; err != nil {
		logger.LogError(sessionId, err)
		tx.Rollback()
		return fmt.Errorf(constants.Registeration_Failed, "driver")
	}

	return tx.Commit().Error
}

func unBookRide(orgCtx *gin.Context, sessionId, driverID string, request BookSeatRequest) error {
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()

	var booking postgress.RideBooking

	err := tx.Where(
		"ride_id = ? AND mobile_number = ?",
		request.RideId,
		request.MobileNumber,
	).First(&booking).Error

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("booking not found")
	}

	result := tx.Delete(&booking)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	result = tx.Exec(`
		UPDATE rides
		SET seats_taken = seats_taken - $1
		WHERE id = $2
		  AND driver_id = $3
		  AND is_active = true
		  AND seats_taken >= $1
	`, booking.Seats, request.RideId, driverID)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("ride not found or cannot unbook seats")
	}

	return tx.Commit().Error
}

func getBookedSeatsByRideID(orgCtx *gin.Context, rideID string) ([]postgress.RideBooking, error) {
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	var bookings []postgress.RideBooking

	err := database.DatabaseConn.Postgres.
		WithContext(ctx).
		Model(&postgress.RideBooking{}).
		Where("ride_id = ?", rideID).
		Order("name ASC").
		Find(&bookings).Error

	if err != nil {
		return nil, err
	}

	return bookings, nil
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
