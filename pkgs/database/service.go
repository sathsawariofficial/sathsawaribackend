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

// checkDriverFleetTies stops a driver being deleted out from under a fleet that
// still depends on them. A fleet with a deleted owner would have nobody able to run
// it, and a shift whose driver no longer exists would strand its passengers, so
// both cases are refused and the caller is told what to sort out first.
func checkDriverFleetTies(ctx context.Context, driverId string) error {
	var ownedGroups int64
	if err := DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Group{}).
		Where("owner_driver_id = ?", driverId).
		Where("status = ?", constants.Status_Active).
		Count(&ownedGroups).Error; err != nil {
		return err
	}

	if ownedGroups > 0 {
		return fmt.Errorf("this driver owns %d active group(s), delete or hand those over first", ownedGroups)
	}

	var upcomingShifts int64
	if err := DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Shift{}).
		Where("driver_id = ?", driverId).
		Where("is_active = ?", true).
		Where("start_datetime > ?", time.Now().Format(constants.DateTimeLayout)).
		Count(&upcomingShifts).Error; err != nil {
		return err
	}

	if upcomingShifts > 0 {
		return fmt.Errorf("this driver is still driving %d upcoming shift(s), cancel or reassign those first", upcomingShifts)
	}

	return nil
}

func DeleteDriver(orgCtx *gin.Context, driver postgress.Driver, updateById string) error {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	if err := checkDriverFleetTies(ctx, driver.ID); err != nil {
		return err
	}

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

	////////// LEAVE EVERY FLEET //////////
	// the driver is on their way out, so their membership and the vehicles they
	// lent stop counting towards any fleet they belonged to
	if err := tx.Model(&postgress.GroupMember{}).
		Where("driver_id = ?", driver.ID).
		Updates(map[string]interface{}{
			"status":  constants.Membership_Status_Left,
			"role_id": "",
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&postgress.GroupVehicle{}).
		Where("driver_id = ?", driver.ID).
		Update("status", constants.Membership_Status_Left).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete original record
	if err := tx.Delete(&postgress.Driver{}, "id = ?", driver.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DeletePassenger archives a passenger account and unpicks everything it was still
// attached to: the seats it holds on upcoming shifts go back to the manager, its
// fleet memberships end, and its standing travel form is cleared. The mobile number
// is prefixed on the archive row so it is free to register again.
func DeletePassenger(orgCtx *gin.Context, passenger postgress.Passenger, updateById string) error {
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

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	delPassenger := postgress.DELPassenger{
		ID:              passenger.ID,
		PassengerName:   passenger.PassengerName,
		PassengerMobile: fmt.Sprintf("DEL_%s_%s", passenger.PassengerMobile, time.Now().String()),
		Password:        passenger.Password,
		Gender:          passenger.Gender,
		Status:          constants.Status_InActive,
		UpdateBy:        updateById,
		CreatedAt:       passenger.CreatedAt,
		UpdatedAt:       time.Now(),
	}

	if err := tx.Create(&delPassenger).Error; err != nil {
		tx.Rollback()
		return err
	}

	////////// FREE THEIR UPCOMING SEATS //////////
	var shiftIds []string
	if err := tx.Model(&postgress.Shift{}).
		Where("is_active = ?", true).
		Where("start_datetime > ?", time.Now().Format(constants.DateTimeLayout)).
		Pluck("id", &shiftIds).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(shiftIds) > 0 {
		if err := tx.Model(&postgress.ShiftSeat{}).
			Where("shift_id IN ?", shiftIds).
			Where("passenger_id = ?", passenger.ID).
			Updates(map[string]interface{}{
				"passenger_id": "",
				"stop_id":      "",
				"gender":       "",
				"status":       constants.Seat_Status_Empty,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Exec(`
			UPDATE shifts
			SET seats_taken = (
				SELECT COUNT(*) FROM shift_seats
				WHERE shift_seats.shift_id = shifts.id
				  AND shift_seats.status = ?
			)
			WHERE shifts.id IN ?
		`, constants.Seat_Status_Assigned, shiftIds).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	////////// LEAVE EVERY FLEET //////////
	if err := tx.Model(&postgress.GroupPassenger{}).
		Where("passenger_id = ?", passenger.ID).
		Update("status", constants.Membership_Status_Left).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("passenger_id = ?", passenger.ID).Delete(&postgress.PassengerLocationPreference{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&postgress.Passenger{}, "id = ?", passenger.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func GetActivePassengerById(orgCtx *gin.Context, passengerId string) (passenger postgress.Passenger, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Where(`id = ?`, passengerId).Where(`status = ?`, constants.Status_Active).Find(&passenger).Error
	return
}

func GetPassengerById(orgCtx *gin.Context, passengerId string) (passenger postgress.Passenger, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Where(`id = ?`, passengerId).Find(&passenger).Error
	return
}

func GetGroupById(orgCtx *gin.Context, groupId string) (group postgress.Group, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).Where(`id = ?`, groupId).Where(`status = ?`, constants.Status_Active).First(&group).Error
	return
}

// HasGroupPermission answers whether a driver holds a permission inside one group.
// The permission is read from the role attached to the driver's membership, so an
// admin can hand a new permission to a role, or invent a brand new role, without
// any code change here. A membership that is not approved never holds anything.
func HasGroupPermission(orgCtx *gin.Context, groupId, driverId, permissionCode string) (hasPermission bool, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	var count int64

	err = DatabaseConn.Postgres.WithContext(ctx).
		Table("group_members").
		Joins("JOIN role_permissions ON role_permissions.role_id = group_members.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("group_members.group_id = ?", groupId).
		Where("group_members.driver_id = ?", driverId).
		Where("group_members.status = ?", constants.Membership_Status_Approved).
		Where("permissions.code = ?", permissionCode).
		Count(&count).Error

	return count > 0, err
}

// IsApprovedGroupMember reports whether a driver is an approved member of a group,
// with no regard to any role that driver may hold.
func IsApprovedGroupMember(orgCtx *gin.Context, groupId, driverId string) (isMember bool, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	var count int64

	err = DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.GroupMember{}).
		Where("group_id = ?", groupId).
		Where("driver_id = ?", driverId).
		Where("status = ?", constants.Membership_Status_Approved).
		Count(&count).Error

	return count > 0, err
}

// IsApprovedGroupPassenger reports whether a passenger is an approved member of a
// group, a passenger can only be seated on the shifts of a group they belong to.
func IsApprovedGroupPassenger(orgCtx *gin.Context, groupId, passengerId string) (isMember bool, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	var count int64

	err = DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.GroupPassenger{}).
		Where("group_id = ?", groupId).
		Where("passenger_id = ?", passengerId).
		Where("status = ?", constants.Membership_Status_Approved).
		Count(&count).Error

	return count > 0, err
}

// VehicleHasShiftDuringTime reports whether a vehicle is already committed to a
// shift that overlaps the given window. excludeShiftId lets a shift that is being
// edited ignore its own row.
func VehicleHasShiftDuringTime(orgCtx *gin.Context, vehicleId, startTime, endTime, excludeShiftId string) (hasShift bool, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	var count int64

	query := DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Shift{}).
		Where("vehicle_id = ?", vehicleId).
		Where("is_active = ?", true).
		Where("start_datetime < ? AND estimated_end_datetime > ?", endTime, startTime)

	if excludeShiftId != "" {
		query = query.Where("id <> ?", excludeShiftId)
	}

	err = query.Count(&count).Error

	return count > 0, err
}

// DriverHasShiftDuringTime reports whether a driver is already driving a shift that
// overlaps the given window.
func DriverHasShiftDuringTime(orgCtx *gin.Context, driverId, startTime, endTime, excludeShiftId string) (hasShift bool, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	var count int64

	query := DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Shift{}).
		Where("driver_id = ?", driverId).
		Where("is_active = ?", true).
		Where("start_datetime < ? AND estimated_end_datetime > ?", endTime, startTime)

	if excludeShiftId != "" {
		query = query.Where("id <> ?", excludeShiftId)
	}

	err = query.Count(&count).Error

	return count > 0, err
}

// VehicleHasRideDuringTime reports whether a vehicle is already committed to a
// carpool ride that overlaps the given window.
func VehicleHasRideDuringTime(orgCtx *gin.Context, vehicleId, startTime, endTime string) (hasRide bool, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	var count int64

	err = DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Ride{}).
		Where("vehicle_id = ?", vehicleId).
		Where("is_active = ?", true).
		Where("start_datetime < ? AND estimated_end_datetime > ?", endTime, startTime).
		Count(&count).Error

	return count > 0, err
}

// DriverHasRideDuringTime reports whether a driver is already driving a carpool
// ride that overlaps the given window.
func DriverHasRideDuringTime(orgCtx *gin.Context, driverId, startTime, endTime string) (hasRide bool, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	var count int64

	err = DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Ride{}).
		Where("driver_id = ?", driverId).
		Where("is_active = ?", true).
		Where("start_datetime < ? AND estimated_end_datetime > ?", endTime, startTime).
		Count(&count).Error

	return count > 0, err
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
