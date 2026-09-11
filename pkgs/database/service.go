package database

import (
	"context"
	"errors"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm/clause"
)

func GetActiveDriverById(orgCtx *gin.Context, driverId string) (driver postgress.Driver, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Where(`id = ?`, driverId).Where(`status = ?`, constants.Status_Active).Find(&driver).Error
	return
}

func GetDriverById(orgCtx *gin.Context, driverId string) (driver postgress.Driver, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Where(`id = ?`, driverId).Find(&driver).Error
	return
}

func GetDriverFCM(orgCtx *gin.Context, id string) (driverFCM postgress.UserFCM, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Where(`user_id = ?`, id).Find(&driverFCM).Error
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

func GetAnnouncements(orgCtx *gin.Context, page int) (announcements []postgress.AnnouncementRequests, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	baseQuery := DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.AnnouncementRequests{})

	// Count total rows BEFORE applying limit/offset
	err = baseQuery.Count(&totalRows).Error
	if err != nil {
		return
	}

	// Apply pagination
	err = baseQuery.
		Limit(pageSize).
		Offset(offset).
		Order("created_at DESC").
		Find(&announcements).Error

	return
}

func SaveMissingLocation(orgCtx context.Context, request postgress.MissingLocations) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Create(&request).Error
	return
}

func DeleteDriver(orgCtx *gin.Context, driver postgress.Driver, updateById string) error {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	tx := DatabaseConn.Postgres.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	////////// DELETE DRIVER //////////
	// Copy driver to archive table
	delDriver := postgress.DELDriver{
		ID:            driver.ID,
		DriverName:    driver.DriverName,
		DriverMobile:  fmt.Sprintf("DEL_%s_%s", driver.DriverMobile, time.Now().String()),
		Status:        constants.Status_InActive,
		UpdateBy:      updateById,
		Password:      driver.Password,
		Pin:           driver.Pin,
		Rating:        driver.Rating,
		NumberOfVotes: driver.NumberOfVotes,
		CreatedAt:     driver.CreatedAt,
		UpdatedAt:     time.Now(),
	}

	if err := tx.Create(&delDriver).Error; err != nil {
		tx.Rollback()
		return err
	}

	////////// DELETE TEMPLATES //////////
	var templates []postgress.RideTemplate
	err := tx.Where("driver_id = ?", driver.ID).Find(&templates).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, template := range templates {
		if err := tx.Delete(&postgress.RideTemplate{}, "id = ?", template.ID).Error; err != nil {
			tx.Rollback()
			return errors.New(constants.General_Error)
		}
	}

	////////// DELETE RIDES //////////
	var rides []postgress.Ride
	err = tx.Where("driver_id = ?", driver.ID).Find(&rides).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, ride := range rides {
		delRide := postgress.DELRide{
			ID:                   ride.ID,
			DriverID:             ride.DriverID,
			VehicleID:            ride.VehicleID,
			StartDatetime:        ride.StartDatetime,
			EstimatedEndDatetime: ride.EstimatedEndDatetime,
			NumberOfSeats:        ride.NumberOfSeats,
			SeatsTaken:           ride.SeatsTaken,
			StartLocation:        ride.StartLocation,
			EndLocation:          ride.EndLocation,
			RoutePoints:          ride.RoutePoints,
			Fare:                 ride.Fare,
			RouteDetails:         ride.RouteDetails,
			IsActive:             false,
			ParentRideId:         ride.ParentRideId,
			Code:                 ride.Code,
			CreatedAt:            ride.CreatedAt,
			UpdatedAt:            time.Now(),
		}

		if err := tx.Create(&delRide).Error; err != nil {
			tx.Rollback()
			return errors.New(constants.General_Error)
		}

		if err := tx.Delete(&postgress.Ride{}, "id = ?", ride.ID).Error; err != nil {
			tx.Rollback()
			return errors.New(constants.General_Error)
		}
	}

	////////// DELETE VEHICLES //////////
	var vehicles []postgress.Vehicle
	err = DatabaseConn.Postgres.Where("driver_id = ?", driver.ID).Find(&vehicles).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, vehicle := range vehicles {
		delVehicle := postgress.DELVehicle{
			ID:            vehicle.ID,
			DriverId:      vehicle.DriverId,
			VehicleNumber: fmt.Sprintf("DEL_%s_%s", vehicle.VehicleNumber, time.Now().String()),
			VehicleInfo:   vehicle.VehicleInfo,
			Status:        constants.Status_InActive,
			CreatedAt:     vehicle.CreatedAt,
			UpdatedAt:     time.Now(),
		}

		if err := tx.Create(&delVehicle).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf(constants.Update_Failed, "vehicle")
		}

		if err := tx.Delete(&postgress.Vehicle{}, "id = ?", vehicle.ID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf(constants.Update_Failed, "vehicle")
		}
	}

	// Delete original record
	if err := tx.Delete(&postgress.Driver{}, "id = ?", driver.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func GetDriverByRideId(orgCtx *gin.Context, rideId string) (
	driverID string,
	vehicleNumber string,
	driverMobile string,
	err error,
) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	row := DatabaseConn.Postgres.WithContext(ctx).
		Table("drivers").
		Select("drivers.id, vehicles.vehicle_number, drivers.driver_mobile").
		Joins("JOIN rides ON rides.driver_id = drivers.id").
		Joins("JOIN vehicles ON vehicles.id = rides.vehicle_id").
		Where("rides.id = ?", rideId).
		Row()

	err = row.Scan(&driverID, &vehicleNumber, &driverMobile)
	return
}

// GetPlaceNotificationSetting returns the place notification setting of one user, a
// setting with no id when they never saved one.
func GetPlaceNotificationSetting(orgCtx *gin.Context, userType int, userId string) (setting postgress.PlaceNotificationSetting, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).
		Where(`user_type = ? AND user_id = ?`, userType, userId).
		Limit(1).
		Find(&setting).Error
	return
}

// SavePlaceNotificationSetting writes a user's whole setting in one statement, the places
// sent replace the ones kept before.
func SavePlaceNotificationSetting(orgCtx *gin.Context, setting *postgress.PlaceNotificationSetting) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_type"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"fcm", "enabled", "places", "updated_at"}),
		}).
		Create(setting).Error
	return
}

// GetPlaceAlertRecipients returns one page of the devices that follow any of the places:
// users of the type with place notifications switched on, and for drivers only active
// ones with an fcm on record. The places overlap is served by the GIN index on places.
// Pages are walked by setting id, pass the last id of a page to read the next one.
func GetPlaceAlertRecipients(orgCtx context.Context, userType int, places []string, afterId string, limit int) (recipients []PlaceAlertRecipient, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := DatabaseConn.Postgres.WithContext(ctx).
		Table("place_notification_settings AS s").
		Where("s.user_type = ? AND s.enabled = ? AND s.places && ?", userType, true, pq.Array(places)).
		Where("s.id > ?", afterId).
		Order("s.id").
		Limit(limit)

	if userType == constants.User_Driver {
		query = query.
			Select("s.id AS setting_id, f.fcm AS fcm").
			Joins("JOIN drivers d ON d.id = s.user_id AND d.status = ?", constants.Status_Active).
			Joins("JOIN user_fcms f ON f.user_id = s.user_id AND f.fcm <> ''")
	} else {
		query = query.
			Select("s.id AS setting_id, s.fcm AS fcm").
			Where("s.fcm <> ''")
	}

	err = query.Scan(&recipients).Error
	return
}
