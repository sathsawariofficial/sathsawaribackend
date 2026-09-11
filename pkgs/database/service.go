package database

import (
	"context"
	"errors"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
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

// checkDriverPickDropTies stops a driver being deleted out from under Pick & Drop. A
// service with a deleted owner would have nobody able to run it, and a shift whose
// driver or vehicle no longer exists would strand its passengers, so both are
// refused and the caller is told what to sort out first.
func checkDriverPickDropTies(ctx context.Context, driverId string) error {
	db := DatabaseConn.Postgres.WithContext(ctx)

	var ownedServices int64
	if err := db.Model(&postgress.PickDropService{}).
		Where("owner_driver_id = ?", driverId).
		Where("status = ?", constants.Status_Active).
		Count(&ownedServices).Error; err != nil {
		return err
	}

	if ownedServices > 0 {
		return errors.New("this driver owns an active Pick & Drop service, disable it first")
	}

	var vehicleIds []string
	if err := db.Model(&postgress.Vehicle{}).Where("driver_id = ?", driverId).Pluck("id", &vehicleIds).Error; err != nil {
		return err
	}

	driverShifts, vehicleShifts, err := CountActiveShiftAssignments(db, driverId, vehicleIds)
	if err != nil {
		return err
	}

	if driverShifts > 0 {
		return fmt.Errorf(constants.On_Active_Shifts, "This driver is", driverShifts)
	}

	if vehicleShifts > 0 {
		return fmt.Errorf(constants.On_Active_Shifts, "A vehicle of this driver is", vehicleShifts)
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

	if err := checkDriverPickDropTies(ctx, driver.ID); err != nil {
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

	////////// LEAVE PICK & DROP //////////
	// the driver is on their way out, so their membership and the vehicles they
	// offered stop counting towards the service they belonged to
	if err := EndOpenMemberships(tx, &postgress.PickDropDriver{}, "driver_id", driver.ID, updateById); err != nil {
		tx.Rollback()
		return err
	}

	if err := EndOpenMemberships(tx, &postgress.PickDropVehicle{}, "driver_id", driver.ID, updateById); err != nil {
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
// attached to: it comes off its active shifts with the seats given back, its Pick &
// Drop membership ends, and its weekly availability and shift requests are archived.
// The mobile number is prefixed on the archive row so it is free to register again.
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

	////////// COME OFF EVERY ACTIVE SHIFT //////////
	if _, err := RemovePassengerFromActiveShifts(tx, passenger.ID, "", constants.Removal_Reason_Account_Deleted, updateById); err != nil {
		tx.Rollback()
		return err
	}

	////////// LEAVE PICK & DROP //////////
	if err := EndOpenMemberships(tx, &postgress.PickDropPassenger{}, "passenger_id", passenger.ID, updateById); err != nil {
		tx.Rollback()
		return err
	}

	////////// ARCHIVE THE WEEKLY AVAILABILITY //////////
	if err := archivePassengerAvailability(tx, passenger.ID, updateById); err != nil {
		tx.Rollback()
		return err
	}

	////////// ARCHIVE THE SHIFT REQUESTS //////////
	if err := archivePassengerShiftRequests(tx, passenger.ID, updateById); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&postgress.Passenger{}, "id = ?", passenger.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func archivePassengerAvailability(tx *gorm.DB, passengerId, deletedBy string) error {
	var days []postgress.PassengerAvailability
	if err := tx.Where("passenger_id = ?", passengerId).Find(&days).Error; err != nil {
		return err
	}

	var locations []postgress.PassengerAvailabilityLocation
	if err := tx.Where("passenger_id = ?", passengerId).Find(&locations).Error; err != nil {
		return err
	}

	if len(days) > 0 {
		archived := make([]postgress.DELPassengerAvailability, 0, len(days))
		for _, day := range days {
			archived = append(archived, postgress.DELPassengerAvailability{
				ID:          day.ID,
				PassengerID: day.PassengerID,
				DayOfWeek:   day.DayOfWeek,
				IsRequired:  day.IsRequired,
				DeletedBy:   deletedBy,
				CreatedAt:   day.CreatedAt,
				UpdatedAt:   time.Now(),
			})
		}

		if err := tx.Create(&archived).Error; err != nil {
			return err
		}
	}

	if len(locations) > 0 {
		archived := make([]postgress.DELPassengerAvailabilityLocation, 0, len(locations))
		for _, location := range locations {
			archived = append(archived, postgress.DELPassengerAvailabilityLocation{
				ID:          location.ID,
				PassengerID: location.PassengerID,
				DayOfWeek:   location.DayOfWeek,
				Sequence:    location.Sequence,
				Location:    location.Location,
				Lat:         location.Lat,
				Lng:         location.Lng,
				Time:        location.Time,
				CreatedAt:   location.CreatedAt,
			})
		}

		if err := tx.Create(&archived).Error; err != nil {
			return err
		}
	}

	if err := tx.Where("passenger_id = ?", passengerId).Delete(&postgress.PassengerAvailabilityLocation{}).Error; err != nil {
		return err
	}

	return tx.Where("passenger_id = ?", passengerId).Delete(&postgress.PassengerAvailability{}).Error
}

func archivePassengerShiftRequests(tx *gorm.DB, passengerId, deletedBy string) error {
	var requests []postgress.ShiftRequest
	if err := tx.Where("passenger_id = ?", passengerId).Find(&requests).Error; err != nil {
		return err
	}

	if len(requests) == 0 {
		return nil
	}

	requestIds := make([]string, 0, len(requests))
	for _, request := range requests {
		requestIds = append(requestIds, request.ID)
	}

	return ArchiveShiftRequests(tx, requests, requestIds, deletedBy)
}

// ArchiveShiftRequests copies shift requests and their locations into the archive and
// hard deletes the originals, the same pattern rides follow.
func ArchiveShiftRequests(tx *gorm.DB, requests []postgress.ShiftRequest, requestIds []string, deletedBy string) error {
	var locations []postgress.ShiftRequestLocation
	if err := tx.Where("shift_request_id IN ?", requestIds).Find(&locations).Error; err != nil {
		return err
	}

	archived := make([]postgress.DELShiftRequest, 0, len(requests))
	for _, request := range requests {
		archived = append(archived, postgress.DELShiftRequest{
			ID:            request.ID,
			PassengerID:   request.PassengerID,
			ServiceID:     request.ServiceID,
			ContactNumber: request.ContactNumber,
			Note:          request.Note,
			DaysOfWeek:    request.DaysOfWeek,
			StartTime:     request.StartTime,
			EndTime:       request.EndTime,
			StartLocation: request.StartLocation,
			EndLocation:   request.EndLocation,
			RoutePoints:   request.RoutePoints,
			DeletedBy:     deletedBy,
			CreatedAt:     request.CreatedAt,
			UpdatedAt:     time.Now(),
		})
	}

	if err := tx.Create(&archived).Error; err != nil {
		return err
	}

	if len(locations) > 0 {
		archivedLocations := make([]postgress.DELShiftRequestLocation, 0, len(locations))
		for _, location := range locations {
			archivedLocations = append(archivedLocations, postgress.DELShiftRequestLocation{
				ID:             location.ID,
				ShiftRequestID: location.ShiftRequestID,
				Sequence:       location.Sequence,
				Location:       location.Location,
				Lat:            location.Lat,
				Lng:            location.Lng,
				Time:           location.Time,
				CreatedAt:      location.CreatedAt,
			})
		}

		if err := tx.Create(&archivedLocations).Error; err != nil {
			return err
		}
	}

	if err := tx.Where("shift_request_id IN ?", requestIds).Delete(&postgress.ShiftRequestLocation{}).Error; err != nil {
		return err
	}

	return tx.Where("id IN ?", requestIds).Delete(&postgress.ShiftRequest{}).Error
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

////////////////////////////// PICK & DROP //////////////////////////////

// GetActiveServiceByOwner returns the Pick & Drop service a driver runs. A driver can
// run at most one, the database refuses a second.
func GetActiveServiceByOwner(orgCtx *gin.Context, driverId string) (service postgress.PickDropService, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).
		Where("owner_driver_id = ?", driverId).
		Where("status = ?", constants.Status_Active).
		First(&service).Error
	return
}

func GetActiveServiceById(orgCtx *gin.Context, serviceId string) (service postgress.PickDropService, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).
		Where("id = ?", serviceId).
		Where("status = ?", constants.Status_Active).
		First(&service).Error
	return
}

// GetApprovedPassengerMembership returns the membership a passenger currently holds.
// A passenger belongs to at most one service, so there is never more than one.
func GetApprovedPassengerMembership(orgCtx *gin.Context, passengerId string) (membership postgress.PickDropPassenger, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = DatabaseConn.Postgres.WithContext(ctx).
		Where("passenger_id = ?", passengerId).
		Where("status = ?", constants.Membership_Status_Approved).
		First(&membership).Error
	return
}

// EndOpenMemberships closes every open row of one person or vehicle: a request still
// waiting is withdrawn and an approved one has left. It is what walking away from a
// service looks like, whichever table the row lives in.
func EndOpenMemberships(tx *gorm.DB, model interface{}, column, id, endedBy string) error {
	now := time.Now()

	if err := tx.Model(model).
		Where(column+" = ?", id).
		Where("status = ?", constants.Membership_Status_Pending).
		Updates(map[string]interface{}{
			"status":   constants.Membership_Status_Withdrawn,
			"ended_by": endedBy,
			"ended_at": &now,
		}).Error; err != nil {
		return err
	}

	return tx.Model(model).
		Where(column+" = ?", id).
		Where("status = ?", constants.Membership_Status_Approved).
		Updates(map[string]interface{}{
			"status":   constants.Membership_Status_Left,
			"ended_by": endedBy,
			"ended_at": &now,
		}).Error
}

// CountActiveShiftAssignments reports how many active shifts a driver drives and how
// many run on any of the given vehicles. A driver or a vehicle cannot walk away from
// a shift people are counting on, so leaving is refused while either is above zero.
func CountActiveShiftAssignments(db *gorm.DB, driverId string, vehicleIds []string) (driverShifts, vehicleShifts int64, err error) {
	if driverId != "" {
		if err = db.Model(&postgress.Shift{}).
			Where("driver_id = ?", driverId).
			Where("status = ?", constants.Shift_Status_Active).
			Count(&driverShifts).Error; err != nil {
			return
		}
	}

	if len(vehicleIds) > 0 {
		err = db.Model(&postgress.Shift{}).
			Where("vehicle_id IN ?", vehicleIds).
			Where("status = ?", constants.Shift_Status_Active).
			Count(&vehicleShifts).Error
	}

	return
}

// RemovePassengerFromActiveShifts takes a passenger off every active shift, or only
// the shifts of one service when serviceId is given. The shift passenger rows are
// marked removed rather than deleted, so the participation stays on record, the
// attendance of trips that have not started yet is dropped since those trips will
// never include them, and the seats are given back. Attendance of trips already
// travelled is left exactly as it is.
func RemovePassengerFromActiveShifts(tx *gorm.DB, passengerId, serviceId, reason, removedBy string) (shiftIds []string, err error) {
	query := tx.Model(&postgress.ShiftPassenger{}).
		Joins("JOIN shifts ON shifts.id = shift_passengers.shift_id AND shifts.status = ?", constants.Shift_Status_Active).
		Where("shift_passengers.passenger_id = ?", passengerId).
		Where("shift_passengers.status = ?", constants.Shift_Passenger_Active)

	if serviceId != "" {
		query = query.Where("shift_passengers.service_id = ?", serviceId)
	}

	if err = query.Pluck("shift_passengers.shift_id", &shiftIds).Error; err != nil || len(shiftIds) == 0 {
		return
	}

	// taking the shift rows first serialises this with an owner adding somebody to
	// the same shift, so the seat count recomputed below always sees both changes
	if err = LockShifts(tx, shiftIds); err != nil {
		return
	}

	now := time.Now()

	if err = tx.Model(&postgress.ShiftPassenger{}).
		Where("passenger_id = ?", passengerId).
		Where("status = ?", constants.Shift_Passenger_Active).
		Where("shift_id IN ?", shiftIds).
		Updates(map[string]interface{}{
			"status":         constants.Shift_Passenger_Removed,
			"removed_by":     removedBy,
			"removed_at":     &now,
			"removal_reason": reason,
		}).Error; err != nil {
		return
	}

	if err = DeleteUpcomingAttendance(tx, shiftIds, []string{passengerId}); err != nil {
		return
	}

	err = SyncOccupiedSeats(tx, shiftIds)

	return
}

// LockShifts takes the rows of the given shifts for update, always in id order so two
// transactions locking overlapping sets can never deadlock each other.
func LockShifts(tx *gorm.DB, shiftIds []string) error {
	var locked []string

	return tx.Model(&postgress.Shift{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id IN ?", shiftIds).
		Order("id").
		Pluck("id", &locked).Error
}

// DeleteUpcomingAttendance drops the attendance rows of trips that have not started
// yet. It is only ever used when somebody comes off a shift, a trip still ahead of
// us will simply not include them.
func DeleteUpcomingAttendance(tx *gorm.DB, shiftIds, passengerIds []string) error {
	return tx.Exec(`
		DELETE FROM shift_attendances
		WHERE passenger_id IN ?
		  AND occurrence_id IN (
			SELECT id FROM shift_occurrences
			WHERE shift_id IN ?
			  AND starts_at > ?
		  )
	`, passengerIds, shiftIds, BusinessNowMinute()).Error
}

// SyncOccupiedSeats recomputes the taken seats of the given shifts from their active
// passengers. The CHECK constraint on shifts turns an overbooked result into an error.
func SyncOccupiedSeats(tx *gorm.DB, shiftIds []string) error {
	return tx.Exec(`
		UPDATE shifts
		SET occupied_seats = (
			SELECT COUNT(*) FROM shift_passengers
			WHERE shift_passengers.shift_id = shifts.id
			  AND shift_passengers.status = ?
		),
		updated_at = ?
		WHERE shifts.id IN ?
	`, constants.Shift_Passenger_Active, time.Now(), shiftIds).Error
}

////////////////////////////// SCHEDULE TIME //////////////////////////////

// BusinessNow is the current time on the business wall clock.
func BusinessNow() time.Time {
	return time.Now().In(constants.Business_Location)
}

func BusinessToday() string {
	return BusinessNow().Format(constants.Date_Layout)
}

func BusinessNowMinute() string {
	return BusinessNow().Format(constants.Minute_Datetime_Layout)
}

// AddDays moves a YYYY-MM-DD date by a number of days.
func AddDays(date string, days int) string {
	day, err := time.ParseInLocation(constants.Date_Layout, date, constants.Business_Location)
	if err != nil {
		return date
	}

	return day.AddDate(0, 0, days).Format(constants.Date_Layout)
}

func IntArray(values []int) pq.Int64Array {
	array := make(pq.Int64Array, 0, len(values))
	for _, value := range values {
		array = append(array, int64(value))
	}

	return array
}

func IntsFromArray(values pq.Int64Array) []int {
	ints := make([]int, 0, len(values))
	for _, value := range values {
		ints = append(ints, int(value))
	}

	return ints
}

// ISOWeekday numbers the days the way the whole product does, Monday 1 to Sunday 7.
func ISOWeekday(day time.Time) int {
	weekday := int(day.Weekday())
	if weekday == 0 {
		return 7
	}

	return weekday
}

// WeekdayDate labels a date for a person, "Wednesday 2026-09-16".
func WeekdayDate(date string) string {
	day, err := time.ParseInLocation(constants.Date_Layout, date, constants.Business_Location)
	if err != nil {
		return date
	}

	return day.Weekday().String() + " " + date
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}

// OccursOn reports whether a shift runs on a given date.
func OccursOn(window ShiftWindow, date string) bool {
	if date < window.StartDate {
		return false
	}

	if window.EndDate != "" && date > window.EndDate {
		return false
	}

	day, err := time.ParseInLocation(constants.Date_Layout, date, constants.Business_Location)
	if err != nil {
		return false
	}

	return containsInt(window.DaysOfWeek, ISOWeekday(day))
}

// FirstSharedDate finds the first date both windows actually run on. Sharing a
// weekday is not enough: a Monday to Friday shift ending on the 10th and a Wednesday
// shift starting on the 20th never meet. Any seven consecutive days contain every
// weekday, so looking at most seven days into the overlap of the two date ranges
// settles it exactly.
func FirstSharedDate(a, b ShiftWindow) (string, bool) {
	start := a.StartDate
	if b.StartDate > start {
		start = b.StartDate
	}

	var end string
	switch {
	case a.EndDate == "":
		end = b.EndDate
	case b.EndDate == "":
		end = a.EndDate
	case a.EndDate < b.EndDate:
		end = a.EndDate
	default:
		end = b.EndDate
	}

	first, err := time.ParseInLocation(constants.Date_Layout, start, constants.Business_Location)
	if err != nil {
		return "", false
	}

	for offset := 0; offset < 7; offset++ {
		day := first.AddDate(0, 0, offset)
		date := day.Format(constants.Date_Layout)

		if end != "" && date > end {
			break
		}

		weekday := ISOWeekday(day)
		if containsInt(a.DaysOfWeek, weekday) && containsInt(b.DaysOfWeek, weekday) {
			return date, true
		}
	}

	return "", false
}

// NextOccurrence is the next date a shift runs that has not started yet, or empty
// when it will never run again.
func NextOccurrence(window ShiftWindow, now time.Time) string {
	today := now.Format(constants.Date_Layout)
	clock := now.Format(constants.Clock_Layout)

	start := today
	if window.StartDate > start {
		start = window.StartDate
	}

	first, err := time.ParseInLocation(constants.Date_Layout, start, constants.Business_Location)
	if err != nil {
		return ""
	}

	// eight days, today may already be under way
	for offset := 0; offset < 8; offset++ {
		date := first.AddDate(0, 0, offset).Format(constants.Date_Layout)

		if !OccursOn(window, date) {
			if window.EndDate != "" && date > window.EndDate {
				return ""
			}
			continue
		}

		if date == today && window.StartTime <= clock {
			continue
		}

		return date
	}

	return ""
}

// UpcomingWindow drops the part of a window already behind us. A shift that has run
// for months is judged only on the trips still to come, so a clash that only existed
// last month can never stop it being edited today.
func UpcomingWindow(window ShiftWindow) ShiftWindow {
	if today := BusinessToday(); window.StartDate < today {
		window.StartDate = today
	}

	return window
}

// rideOverlapsWindow walks every calendar day a ride touches and checks it against the
// shift's trip on that day, if the shift runs then. It returns the date of the first
// trip the ride runs into.
func rideOverlapsWindow(window ShiftWindow, start, end time.Time) (string, bool) {
	start, end = start.In(constants.Business_Location), end.In(constants.Business_Location)
	day := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, constants.Business_Location)

	for !day.After(end) {
		date := day.Format(constants.Date_Layout)

		if OccursOn(window, date) {
			tripStart, err := time.ParseInLocation(constants.Minute_Datetime_Layout, date+" "+window.StartTime, constants.Business_Location)
			if err != nil {
				return "", false
			}

			tripEnd, err := time.ParseInLocation(constants.Minute_Datetime_Layout, date+" "+window.EndTime, constants.Business_Location)
			if err != nil {
				return "", false
			}

			if start.Before(tripEnd) && end.After(tripStart) {
				return date, true
			}
		}

		day = day.AddDate(0, 0, 1)
	}

	return "", false
}

////////////////////////////// CLASHES //////////////////////////////

type shiftClashRow struct {
	ID        string
	Name      string
	SubjectID string
	Subject   string
	// without the type gorm cannot parse the field and silently leaves it empty, and a
	// shift with no days never shares a date with anything, so every clash is missed
	DaysOfWeek pq.Int64Array `gorm:"type:integer[]"`
	StartDate  string
	EndDate    string
	StartTime  string
	EndTime    string
}

const shiftClashColumns = "shifts.id, shifts.name, shifts.days_of_week, shifts.start_date, shifts.end_date, shifts.start_time, shifts.end_time"

// shiftOverlapQuery narrows the active shifts down to the ones that could possibly
// clash with a window: their clock times overlap, they share a weekday and their date
// ranges touch. Touching times do not overlap, a shift ending at 09:00 leaves the
// driver free for one starting at 09:00. The exact shared date is settled afterwards.
func shiftOverlapQuery(db *gorm.DB, window ShiftWindow, excludeShiftId string) *gorm.DB {
	query := db.Table("shifts").
		Where("shifts.status = ?", constants.Shift_Status_Active).
		Where("shifts.start_time < ? AND shifts.end_time > ?", window.EndTime, window.StartTime).
		Where("shifts.days_of_week && ?::integer[]", IntArray(window.DaysOfWeek)).
		Where("(shifts.end_date = '' OR shifts.end_date >= ?)", window.StartDate)

	if window.EndDate != "" {
		query = query.Where("shifts.start_date <= ?", window.EndDate)
	}

	if excludeShiftId != "" {
		query = query.Where("shifts.id <> ?", excludeShiftId)
	}

	return query
}

// FindShiftClashes looks for any driver, vehicle or passenger already committed to an
// overlapping shift, in any Pick & Drop service. It costs one query per kind however
// many ids are asked about, and returns the clash of every id that has one.
func FindShiftClashes(db *gorm.DB, window ShiftWindow, driverIds, vehicleIds, passengerIds []string, excludeShiftId string) (clashes map[string]ShiftClash, err error) {
	clashes = map[string]ShiftClash{}
	window = UpcomingWindow(window)

	collect := func(kind string, rows []shiftClashRow) {
		for _, row := range rows {
			if _, known := clashes[row.SubjectID]; known {
				continue
			}

			other := ShiftWindow{
				DaysOfWeek: IntsFromArray(row.DaysOfWeek),
				StartDate:  row.StartDate,
				EndDate:    row.EndDate,
				StartTime:  row.StartTime,
				EndTime:    row.EndTime,
			}

			date, shared := FirstSharedDate(window, other)
			if !shared {
				continue
			}

			clashes[row.SubjectID] = ShiftClash{
				Kind:      kind,
				SubjectID: row.SubjectID,
				Subject:   row.Subject,
				ShiftName: row.Name,
				Date:      WeekdayDate(date),
				StartTime: row.StartTime,
				EndTime:   row.EndTime,
			}
		}
	}

	if len(driverIds) > 0 {
		var rows []shiftClashRow
		if err = shiftOverlapQuery(db, window, excludeShiftId).
			Select(shiftClashColumns+", shifts.driver_id AS subject_id, 'Driver ' || COALESCE(drivers.driver_name, '') AS subject").
			Joins("LEFT JOIN drivers ON drivers.id = shifts.driver_id").
			Where("shifts.driver_id IN ?", driverIds).
			Find(&rows).Error; err != nil {
			return
		}
		collect(Clash_Kind_Driver, rows)
	}

	if len(vehicleIds) > 0 {
		var rows []shiftClashRow
		if err = shiftOverlapQuery(db, window, excludeShiftId).
			Select(shiftClashColumns+", shifts.vehicle_id AS subject_id, 'Vehicle ' || COALESCE(vehicles.vehicle_number, '') AS subject").
			Joins("LEFT JOIN vehicles ON vehicles.id = shifts.vehicle_id").
			Where("shifts.vehicle_id IN ?", vehicleIds).
			Find(&rows).Error; err != nil {
			return
		}
		collect(Clash_Kind_Vehicle, rows)
	}

	if len(passengerIds) > 0 {
		var rows []shiftClashRow
		if err = shiftOverlapQuery(db, window, excludeShiftId).
			Select(shiftClashColumns+", shift_passengers.passenger_id AS subject_id, 'Passenger ' || COALESCE(passengers.passenger_name, '') AS subject").
			Joins("JOIN shift_passengers ON shift_passengers.shift_id = shifts.id AND shift_passengers.status = ?", constants.Shift_Passenger_Active).
			Joins("LEFT JOIN passengers ON passengers.id = shift_passengers.passenger_id").
			Where("shift_passengers.passenger_id IN ?", passengerIds).
			Find(&rows).Error; err != nil {
			return
		}
		collect(Clash_Kind_Passenger, rows)
	}

	return
}

// FindRideClashes looks for ride share rides of the given drivers or vehicles that
// overlap any trip of the window. A driver and a vehicle are the same things in both
// halves of the product, they cannot be on a ride and a shift at once.
func FindRideClashes(db *gorm.DB, window ShiftWindow, driverIds, vehicleIds []string) (clashes map[string]ShiftClash, err error) {
	clashes = map[string]ShiftClash{}

	if len(driverIds) == 0 && len(vehicleIds) == 0 {
		return
	}

	window = UpcomingWindow(window)

	// a ride already over is history, whatever date the window starts on
	from := window.StartDate + " 00:00:00"
	if now := BusinessNow().Format(constants.DateTimeLayout); now > from {
		from = now
	}

	query := db.Table("rides").
		Select("id, driver_id, vehicle_id, start_datetime, estimated_end_datetime").
		Where("is_active = ?", true).
		Where("(driver_id IN ? OR vehicle_id IN ?)", driverIds, vehicleIds).
		Where("estimated_end_datetime > ?", from)

	if window.EndDate != "" {
		query = query.Where("start_datetime <= ?", window.EndDate+" 23:59:59")
	}

	var rows []rideClashRow
	if err = query.Find(&rows).Error; err != nil {
		return
	}

	drivers := map[string]bool{}
	for _, id := range driverIds {
		drivers[id] = true
	}

	vehicles := map[string]bool{}
	for _, id := range vehicleIds {
		vehicles[id] = true
	}

	for _, ride := range rows {
		start, e := time.ParseInLocation(constants.DateTimeLayout, ride.StartDatetime, constants.Business_Location)
		if e != nil {
			continue
		}

		end, e := time.ParseInLocation(constants.DateTimeLayout, ride.EstimatedEndDatetime, constants.Business_Location)
		if e != nil {
			continue
		}

		if _, overlaps := rideOverlapsWindow(window, start, end); !overlaps {
			continue
		}

		clash := ShiftClash{
			IsRide:    true,
			StartTime: start.Format(constants.Minute_Datetime_Layout),
			EndTime:   end.Format(constants.Minute_Datetime_Layout),
		}

		if _, known := clashes[ride.DriverID]; drivers[ride.DriverID] && !known {
			clash.Kind, clash.SubjectID, clash.Subject = Clash_Kind_Driver, ride.DriverID, "The driver"
			clashes[ride.DriverID] = clash
		}

		if _, known := clashes[ride.VehicleID]; vehicles[ride.VehicleID] && !known {
			clash.Kind, clash.SubjectID, clash.Subject = Clash_Kind_Vehicle, ride.VehicleID, "The vehicle"
			clashes[ride.VehicleID] = clash
		}
	}

	return
}

type rideClashRow struct {
	ID                   string
	DriverID             string
	VehicleID            string
	StartDatetime        string
	EstimatedEndDatetime string
}

// FindRideScheduleClash is the ride share side of the same rule: whether any slot a
// ride request lays down overlaps an active ride of the driver on any vehicle, an
// active ride of the vehicle whoever drives it, or a trip of an active shift either
// of them is on. It costs two queries however long the series is, and returns the
// clash of the earliest slot. Run it in the transaction that writes the rides, after
// LockShiftResources has locked the driver and the vehicle.
func FindRideScheduleClash(tx *gorm.DB, driverId, vehicleId string, slots []RideSlot, excludeRideId string) (clash *ShiftClash, err error) {
	if len(slots) == 0 {
		return
	}

	first, last := slots[0].Start, slots[0].End
	for _, slot := range slots {
		if slot.Start.Before(first) {
			first = slot.Start
		}
		if slot.End.After(last) {
			last = slot.End
		}
	}
	first, last = first.In(constants.Business_Location), last.In(constants.Business_Location)

	query := tx.Table("rides").
		Select("id, driver_id, vehicle_id, start_datetime, estimated_end_datetime").
		Where("is_active = ?", true).
		Where("(driver_id = ? OR vehicle_id = ?)", driverId, vehicleId).
		Where("start_datetime < ? AND estimated_end_datetime > ?", last.Format(constants.DateTimeLayout), first.Format(constants.DateTimeLayout)).
		Order("start_datetime ASC")

	if excludeRideId != "" {
		query = query.Where("id <> ?", excludeRideId)
	}

	var rows []rideClashRow
	if err = query.Find(&rows).Error; err != nil {
		return
	}

	type booked struct {
		row        rideClashRow
		start, end time.Time
	}

	rides := make([]booked, 0, len(rows))
	for _, row := range rows {
		start, e := time.ParseInLocation(constants.DateTimeLayout, row.StartDatetime, constants.Business_Location)
		if e != nil {
			continue
		}

		end, e := time.ParseInLocation(constants.DateTimeLayout, row.EstimatedEndDatetime, constants.Business_Location)
		if e != nil {
			continue
		}

		rides = append(rides, booked{row: row, start: start, end: end})
	}

	var shifts []postgress.Shift
	if err = tx.
		Where("status = ?", constants.Shift_Status_Active).
		Where("(driver_id = ? OR vehicle_id = ?)", driverId, vehicleId).
		Where("(end_date = '' OR end_date >= ?)", first.Format(constants.Date_Layout)).
		Where("start_date <= ?", last.Format(constants.Date_Layout)).
		Order("start_time ASC").
		Find(&shifts).Error; err != nil {
		return
	}

	// the driver is named when it is them, the vehicle only when somebody else drives it
	subject := func(isDriver bool) (string, string, string) {
		if isDriver {
			return Clash_Kind_Driver, driverId, "The driver"
		}

		return Clash_Kind_Vehicle, vehicleId, "The vehicle"
	}

	for _, slot := range slots {
		for _, ride := range rides {
			if !slot.Start.Before(ride.end) || !slot.End.After(ride.start) {
				continue
			}

			found := ShiftClash{
				IsRide:    true,
				StartTime: ride.start.Format(constants.Minute_Datetime_Layout),
				EndTime:   ride.end.Format(constants.Minute_Datetime_Layout),
			}
			found.Kind, found.SubjectID, found.Subject = subject(ride.row.DriverID == driverId)

			return &found, nil
		}

		for _, shift := range shifts {
			window := ShiftWindow{
				DaysOfWeek: IntsFromArray(shift.DaysOfWeek),
				StartDate:  shift.StartDate,
				EndDate:    shift.EndDate,
				StartTime:  shift.StartTime,
				EndTime:    shift.EndTime,
			}

			date, overlaps := rideOverlapsWindow(window, slot.Start, slot.End)
			if !overlaps {
				continue
			}

			found := ShiftClash{
				ShiftName: shift.Name,
				Date:      WeekdayDate(date),
				StartTime: shift.StartTime,
				EndTime:   shift.EndTime,
			}
			found.Kind, found.SubjectID, found.Subject = subject(shift.DriverID == driverId)

			return &found, nil
		}
	}

	return nil, nil
}

// LockShiftResources takes a transaction scoped advisory lock on every driver, vehicle
// and passenger about to be put on a shift or a ride. The clash checks read what is
// already committed, so without this two owners assigning the same driver at the same
// moment, or a driver creating a ride while their owner puts them on a shift, could
// both pass and both write. Keys are taken in sorted order so two assignments sharing
// resources can never deadlock, and the locks end with the transaction.
func LockShiftResources(tx *gorm.DB, driverIds, vehicleIds, passengerIds []string) error {
	seen := map[string]bool{}
	keys := []string{}

	add := func(kind string, ids []string) {
		for _, id := range ids {
			if id == "" {
				continue
			}

			key := "shift:" + kind + ":" + id
			if !seen[key] {
				seen[key] = true
				keys = append(keys, key)
			}
		}
	}

	add(Clash_Kind_Driver, driverIds)
	add(Clash_Kind_Vehicle, vehicleIds)
	add(Clash_Kind_Passenger, passengerIds)

	sort.Strings(keys)

	for _, key := range keys {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", key).Error; err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////// OCCURRENCES //////////////////////////////

// CreateOccurrences writes the trips active shifts make between two dates, one row a
// day, freezing the driver, vehicle, names and route as they stand. It is written in
// one statement however many shifts there are, never invents a trip from before the
// shift existed, and is safe to run again and again: a day already written is left
// untouched.
func CreateOccurrences(db *gorm.DB, fromDate, toDate, shiftId string) error {
	sql := `
		INSERT INTO shift_occurrences (
			id, shift_id, service_id, occurrence_date, start_time, end_time, starts_at, ends_at,
			driver_id, vehicle_id, shift_name, service_name, driver_name, driver_mobile, vehicle_number,
			route, status, created_at, updated_at
		)
		SELECT
			gen_random_uuid()::text, s.id, s.service_id, days.day, s.start_time, s.end_time,
			days.day || ' ' || s.start_time, days.day || ' ' || s.end_time,
			s.driver_id, s.vehicle_id, s.name,
			COALESCE(pick_drop_services.name, ''), COALESCE(drivers.driver_name, ''),
			COALESCE(drivers.driver_mobile, ''), COALESCE(vehicles.vehicle_number, ''),
			COALESCE((
				SELECT json_agg(json_build_object(
					'sequence', l.sequence, 'location', l.location, 'lat', l.lat, 'lng', l.lng, 'time', l.time
				) ORDER BY l.sequence)
				FROM shift_locations l WHERE l.shift_id = s.id
			)::text, '[]'),
			?, now(), now()
		FROM shifts s
		CROSS JOIN (
			SELECT to_char(g, 'YYYY-MM-DD') AS day, EXTRACT(ISODOW FROM g)::int AS weekday
			FROM generate_series(?::date, ?::date, interval '1 day') AS g
		) days
		LEFT JOIN pick_drop_services ON pick_drop_services.id = s.service_id
		LEFT JOIN drivers ON drivers.id = s.driver_id
		LEFT JOIN vehicles ON vehicles.id = s.vehicle_id
		WHERE s.status = ?
		  AND days.weekday = ANY(s.days_of_week)
		  AND days.day >= s.start_date
		  AND (s.end_date = '' OR days.day <= s.end_date)
		  AND days.day || ' ' || s.start_time >= to_char(s.created_at AT TIME ZONE ?, 'YYYY-MM-DD HH24:MI')`

	args := []interface{}{
		constants.Occurrence_Status_Scheduled,
		fromDate,
		toDate,
		constants.Shift_Status_Active,
		constants.Business_Location_Name,
	}

	if shiftId != "" {
		sql += " AND s.id = ?"
		args = append(args, shiftId)
	}

	sql += " ON CONFLICT (shift_id, occurrence_date) DO NOTHING"

	return db.Exec(sql, args...).Error
}

// FillOccurrenceAttendance gives every passenger who was on a shift when a trip was
// due to start their present by default attendance row for that trip. Who was on it
// is read from when they were added and removed, never from who is on it now, so a
// replacement passenger never lands on a trip that was made before them. Like
// CreateOccurrences it is idempotent.
func FillOccurrenceAttendance(db *gorm.DB, fromDate, toDate, shiftId string) error {
	sql := `
		INSERT INTO shift_attendances (
			id, occurrence_id, shift_id, passenger_id, occurrence_date,
			location_sequence, location, location_time, status, created_at, updated_at
		)
		SELECT
			gen_random_uuid()::text, o.id, o.shift_id, sp.passenger_id, o.occurrence_date,
			sp.location_sequence, COALESCE(l.location, ''), COALESCE(l.time, ''), ?, now(), now()
		FROM shift_occurrences o
		JOIN shift_passengers sp ON sp.shift_id = o.shift_id
		LEFT JOIN shift_locations l ON l.shift_id = o.shift_id AND l.sequence = sp.location_sequence
		WHERE o.occurrence_date >= ?
		  AND o.occurrence_date <= ?
		  AND to_char(sp.added_at AT TIME ZONE ?, 'YYYY-MM-DD HH24:MI') <= o.starts_at
		  AND (sp.removed_at IS NULL OR to_char(sp.removed_at AT TIME ZONE ?, 'YYYY-MM-DD HH24:MI') > o.starts_at)`

	args := []interface{}{
		constants.Attendance_Present,
		fromDate,
		toDate,
		constants.Business_Location_Name,
		constants.Business_Location_Name,
	}

	if shiftId != "" {
		sql += " AND o.shift_id = ?"
		args = append(args, shiftId)
	}

	sql += " ON CONFLICT (occurrence_id, passenger_id) DO NOTHING"

	return db.Exec(sql, args...).Error
}

// MaterializeOccurrences writes the trips between two dates and their attendance.
func MaterializeOccurrences(db *gorm.DB, fromDate, toDate, shiftId string) error {
	if err := CreateOccurrences(db, fromDate, toDate, shiftId); err != nil {
		return err
	}

	return FillOccurrenceAttendance(db, fromDate, toDate, shiftId)
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
