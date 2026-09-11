package shift

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func withTimeout(orgCtx *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
}

func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func pageBounds(page int) (limit, offset int) {
	limit = configuration.ConfigurationData.PageSize
	offset = (page - 1) * limit
	return
}

func forUpdate(tx *gorm.DB) *gorm.DB {
	return tx.Clauses(clause.Locking{Strength: "UPDATE"})
}

func driverNotice(driverId, notificationType, title, message string) notice {
	return notice{UserId: driverId, UserType: constants.User_Driver, Type: notificationType, Title: title, Message: message}
}

func passengerNotice(passengerId, notificationType, title, message string) notice {
	return notice{UserId: passengerId, UserType: constants.User_Passenger, Type: notificationType, Title: title, Message: message}
}

// vehicleOwnerNotices tells the driver who brought a vehicle what happens to it on a
// shift, unless they are the one driving it or the one making the change.
func vehicleOwnerNotices(vehicle postgress.Vehicle, driverId, actorId, notificationType, title, message string) []notice {
	if utils.IsStringEmpty(vehicle.DriverId) || vehicle.DriverId == driverId || vehicle.DriverId == actorId {
		return nil
	}

	return []notice{driverNotice(vehicle.DriverId, notificationType, title, message)}
}

func shiftData(shiftId, serviceId string) map[string]string {
	return map[string]string{
		constants.NOTIFICATION_KEY_SHIFT_ID:   shiftId,
		constants.NOTIFICATION_KEY_SERVICE_ID: serviceId,
	}
}

func sendNotices(ctx *gin.Context, sessionId string, data map[string]string, notices []notice) {
	for _, item := range notices {
		utils.SendUserNotification(ctx, sessionId, item.Type, item.UserId, item.UserType, item.Title, item.Message, data)
	}
}

func displayName(name string) string {
	if utils.IsStringEmpty(strings.TrimSpace(name)) {
		return "unnamed shift"
	}

	return name
}

var weekdayNames = []string{"", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

// scheduleLabel describes when a shift runs for a person, "Mon, Wed, Fri from
// 2026-09-14 to 2026-12-31".
func scheduleLabel(days pq.Int64Array, startDate, endDate string) string {
	sorted := database.IntsFromArray(days)
	sort.Ints(sorted)

	names := make([]string, 0, len(sorted))
	for _, day := range sorted {
		if day >= constants.Day_Of_Week_Min_Value && day <= constants.Day_Of_Week_Max_Value {
			names = append(names, weekdayNames[day])
		}
	}

	label := strings.Join(names, ", ") + " from " + startDate
	if !utils.IsStringEmpty(endDate) {
		label += " to " + endDate
	}

	return label
}

func windowOf(shift postgress.Shift) database.ShiftWindow {
	return database.ShiftWindow{
		DaysOfWeek: database.IntsFromArray(shift.DaysOfWeek),
		StartDate:  shift.StartDate,
		EndDate:    shift.EndDate,
		StartTime:  shift.StartTime,
		EndTime:    shift.EndTime,
	}
}

func detailsWindow(details postgress.ShiftDetails) database.ShiftWindow {
	return database.ShiftWindow{
		DaysOfWeek: database.IntsFromArray(details.DaysOfWeek),
		StartDate:  details.StartDate,
		EndDate:    details.EndDate,
		StartTime:  details.StartTime,
		EndTime:    details.EndTime,
	}
}

func sameDays(a, b pq.Int64Array) bool {
	if len(a) != len(b) {
		return false
	}

	set := map[int64]bool{}
	for _, day := range a {
		set[day] = true
	}

	for _, day := range b {
		if !set[day] {
			return false
		}
	}

	return true
}

// routeJSON freezes a route in the exact shape occurrences are written with, so a
// route rewritten by an edit reads back the same as one written by the scheduler.
func routeJSON(locations []postgress.ShiftLocation) string {
	stops := make([]ShiftLocationDetail, 0, len(locations))
	for _, location := range locations {
		stops = append(stops, ShiftLocationDetail{
			Sequence: location.Sequence,
			Location: location.Location,
			Lat:      location.Lat,
			Lng:      location.Lng,
			Time:     location.Time,
		})
	}

	encoded, err := json.Marshal(stops)
	if err != nil {
		return "[]"
	}

	return string(encoded)
}

func locationBySequence(locations []postgress.ShiftLocation, sequence int) postgress.ShiftLocation {
	for _, location := range locations {
		if location.Sequence == sequence {
			return location
		}
	}

	return postgress.ShiftLocation{}
}

func buildLocations(shiftId string, route []utils.RouteLocation) []postgress.ShiftLocation {
	locations := make([]postgress.ShiftLocation, 0, len(route))
	for index, location := range route {
		locations = append(locations, postgress.ShiftLocation{
			ID:       database.GenerateUUID(),
			ShiftID:  shiftId,
			Sequence: index + 1,
			Location: location.Location,
			Lat:      location.Lat,
			Lng:      location.Lng,
			Time:     location.Time,
		})
	}

	return locations
}

func passengerNames(tx *gorm.DB, passengerIds []string) (names map[string]string, err error) {
	names = map[string]string{}

	if len(passengerIds) == 0 {
		return
	}

	var rows []postgress.Passenger
	if err = tx.Select("id, passenger_name").Where("id IN ?", passengerIds).Find(&rows).Error; err != nil {
		return
	}

	for _, row := range rows {
		names[row.ID] = row.PassengerName
	}

	return
}

func getOwnedActiveService(tx *gorm.DB, ownerId string) (service postgress.PickDropService, err error) {
	err = tx.Where("owner_driver_id = ?", ownerId).
		Where("status = ?", constants.Status_Active).
		First(&service).Error
	if isNotFound(err) {
		err = utils.Refuse(constants.Not_Service_Owner)
	}

	return
}

// lockOwnedShift takes a shift of the caller's own service for the rest of the
// transaction. Only the owner ever changes a shift, and only one change at a time.
func lockOwnedShift(tx *gorm.DB, ownerId, shiftId string, mustBeActive bool) (shift postgress.Shift, service postgress.PickDropService, err error) {
	if service, err = getOwnedActiveService(tx, ownerId); err != nil {
		return
	}

	err = forUpdate(tx).Where("id = ?", shiftId).Where("service_id = ?", service.ID).First(&shift).Error
	if isNotFound(err) {
		err = utils.Refuse(constants.Shift_Not_Found)
		return
	}
	if err != nil {
		return
	}

	if mustBeActive && shift.Status != constants.Shift_Status_Active {
		err = utils.Refuse(fmt.Sprintf(constants.Shift_Not_Active, shift.Status))
	}

	return
}

// checkDriverEligible allows the owner themselves or an approved driver of the
// service behind the wheel, nobody else. A member who joined with vehicles only
// belongs to the service but never drives its shifts.
func checkDriverEligible(tx *gorm.DB, service postgress.PickDropService, driverId string) (driver postgress.Driver, err error) {
	err = tx.Where("id = ?", driverId).Where("status = ?", constants.Status_Active).First(&driver).Error
	if isNotFound(err) {
		err = utils.Refuse(constants.Driver_Not_In_Service)
		return
	}
	if err != nil || driverId == service.OwnerDriverID {
		return
	}

	var membership postgress.PickDropDriver
	err = tx.Where("service_id = ?", service.ID).
		Where("driver_id = ?", driverId).
		Where("status = ?", constants.Membership_Status_Approved).
		First(&membership).Error
	if isNotFound(err) {
		err = utils.Refuse(constants.Driver_Not_In_Service)
		return
	}
	if err != nil {
		return
	}

	if membership.JoinType == constants.Join_Type_Vehicles {
		err = utils.Refuse(fmt.Sprintf(constants.Driver_Vehicles_Only, driver.DriverName))
	}

	return
}

// checkVehicleEligible allows a vehicle approved for the service that still exists and
// has a valid seat count. A vehicle of unknown size can never be put on a shift.
func checkVehicleEligible(tx *gorm.DB, serviceId, vehicleId string) (vehicle postgress.Vehicle, err error) {
	var approvedVehicles int64
	if err = tx.Model(&postgress.PickDropVehicle{}).
		Where("service_id = ?", serviceId).
		Where("vehicle_id = ?", vehicleId).
		Where("status = ?", constants.Membership_Status_Approved).
		Count(&approvedVehicles).Error; err != nil {
		return
	}

	if approvedVehicles == 0 {
		err = utils.Refuse(constants.Vehicle_Not_In_Service)
		return
	}

	err = tx.Where("id = ?", vehicleId).Where("status = ?", constants.Status_Active).First(&vehicle).Error
	if isNotFound(err) {
		err = utils.Refuse(constants.Vehicle_Not_In_Service)
		return
	}
	if err != nil {
		return
	}

	if vehicle.NumberOfSeats < constants.Number_Of_Seats_Min_Value {
		err = utils.Refuse(fmt.Sprintf(constants.Vehicle_Seats_Missing, vehicle.VehicleNumber))
	}

	return
}

// loadEligiblePassengers loads, in one query, the passengers that are approved members
// of the service with an active account, and refuses the first one that is not.
func loadEligiblePassengers(tx *gorm.DB, serviceId string, passengerIds []string) (passengers map[string]postgress.Passenger, err error) {
	passengers = map[string]postgress.Passenger{}

	if len(passengerIds) == 0 {
		return
	}

	var rows []postgress.Passenger
	if err = tx.Table("passengers").
		Select("passengers.*").
		Joins("JOIN pick_drop_passengers ON pick_drop_passengers.passenger_id = passengers.id AND pick_drop_passengers.service_id = ? AND pick_drop_passengers.status = ?",
			serviceId, constants.Membership_Status_Approved).
		Where("passengers.id IN ?", passengerIds).
		Where("passengers.status = ?", constants.Status_Active).
		Find(&rows).Error; err != nil {
		return
	}

	for _, row := range rows {
		passengers[row.ID] = row
	}

	for _, passengerId := range passengerIds {
		if _, found := passengers[passengerId]; !found {
			err = utils.Refuse(fmt.Sprintf(constants.Passenger_Not_In_Service, "Passenger "+passengerId))
			return
		}
	}

	return
}

// checkClashes refuses the first driver, vehicle or passenger already committed to an
// overlapping shift in any service, or, for a driver or a vehicle, to a ride.
func checkClashes(tx *gorm.DB, window database.ShiftWindow, driverIds, vehicleIds, passengerIds []string, excludeShiftId string) error {
	shiftClashes, err := database.FindShiftClashes(tx, window, driverIds, vehicleIds, passengerIds, excludeShiftId)
	if err != nil {
		return err
	}

	for _, ids := range [][]string{driverIds, vehicleIds, passengerIds} {
		for _, id := range ids {
			if clash, found := shiftClashes[id]; found {
				return utils.Refuse(clash.Message())
			}
		}
	}

	rideClashes, err := database.FindRideClashes(tx, window, driverIds, vehicleIds)
	if err != nil {
		return err
	}

	for _, ids := range [][]string{driverIds, vehicleIds} {
		for _, id := range ids {
			if clash, found := rideClashes[id]; found {
				return utils.Refuse(clash.Message())
			}
		}
	}

	return nil
}

// createShift writes a whole shift in one transaction. Every driver, vehicle and
// passenger involved is locked first, so the clash checks that follow cannot be raced
// by another assignment of the same people, and the seat count is checked against the
// vehicle and again by the database constraint.
func createShift(orgCtx *gin.Context, sessionId, ownerId string, request CreateShiftRequest) (result shiftContext, err error) {
	logger.LogInfo("Request received in createShift", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		service, e := getOwnedActiveService(tx, ownerId)
		if e != nil {
			return e
		}

		passengerIds := make([]string, 0, len(request.Passengers))
		for _, passenger := range request.Passengers {
			passengerIds = append(passengerIds, passenger.PassengerId)
		}

		if e = database.LockShiftResources(tx, []string{request.DriverId}, []string{request.VehicleId}, passengerIds); e != nil {
			return e
		}

		driver, e := checkDriverEligible(tx, service, request.DriverId)
		if e != nil {
			return e
		}

		vehicle, e := checkVehicleEligible(tx, service.ID, request.VehicleId)
		if e != nil {
			return e
		}

		passengers, e := loadEligiblePassengers(tx, service.ID, passengerIds)
		if e != nil {
			return e
		}

		if len(passengerIds) > vehicle.NumberOfSeats {
			return utils.Refuse(fmt.Sprintf(constants.Vehicle_Capacity_Exceeded, vehicle.VehicleNumber, vehicle.NumberOfSeats, len(passengerIds)))
		}

		window := database.ShiftWindow{
			DaysOfWeek: request.DaysOfWeek,
			StartDate:  request.StartDate,
			EndDate:    request.EndDate,
			StartTime:  request.Locations[0].Time,
			EndTime:    request.Locations[len(request.Locations)-1].Time,
		}

		if e = checkClashes(tx, window, []string{request.DriverId}, []string{request.VehicleId}, passengerIds, ""); e != nil {
			return e
		}

		shift := postgress.Shift{
			ID:            database.GenerateUUID(),
			ServiceID:     service.ID,
			Name:          vehicle.VehicleNumber,
			DriverID:      request.DriverId,
			VehicleID:     request.VehicleId,
			DaysOfWeek:    database.IntArray(request.DaysOfWeek),
			StartDate:     request.StartDate,
			EndDate:       request.EndDate,
			StartTime:     window.StartTime,
			EndTime:       window.EndTime,
			SeatCapacity:  vehicle.NumberOfSeats,
			OccupiedSeats: len(passengerIds),
			Status:        constants.Shift_Status_Active,
			CreatedBy:     ownerId,
		}

		if e = tx.Create(&shift).Error; e != nil {
			return e
		}

		locations := buildLocations(shift.ID, request.Locations)
		if e = tx.Create(&locations).Error; e != nil {
			return e
		}

		now := time.Now()
		rows := make([]postgress.ShiftPassenger, 0, len(request.Passengers))
		names := map[string]string{}

		for _, passenger := range request.Passengers {
			rows = append(rows, postgress.ShiftPassenger{
				ID:               database.GenerateUUID(),
				ShiftID:          shift.ID,
				ServiceID:        service.ID,
				PassengerID:      passenger.PassengerId,
				LocationSequence: passenger.LocationSequence,
				Status:           constants.Shift_Passenger_Active,
				AddedBy:          ownerId,
				AddedAt:          now,
			})
			names[passenger.PassengerId] = passengers[passenger.PassengerId].PassengerName
		}

		if len(rows) > 0 {
			if e = tx.Create(&rows).Error; e != nil {
				return e
			}
		}

		today := database.BusinessToday()
		if e = database.MaterializeOccurrences(tx, today, database.AddDays(today, constants.Occurrence_Horizon_Days), shift.ID); e != nil {
			return e
		}

		result = shiftContext{
			Shift:      shift,
			Service:    service,
			Driver:     driver,
			Vehicle:    vehicle,
			Locations:  locations,
			Passengers: rows,
			Names:      names,
		}

		return nil
	})

	logger.LogInfo("Response returned from createShift", sessionId)

	return
}

// updateShift edits a shift and re-runs exactly the checks the change calls for: a new
// driver or vehicle has to be eligible and free, a new vehicle has to seat everybody
// already on the shift, and a new schedule is checked for the driver, the vehicle and
// every passenger. Trips already made keep what they froze, only trips still ahead
// follow the edit.
func updateShift(orgCtx *gin.Context, sessionId, ownerId string, request UpdateShiftRequest) (before, after shiftContext, err error) {
	logger.LogInfo("Request received in updateShift", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		shift, service, e := lockOwnedShift(tx, ownerId, request.ShiftId, true)
		if e != nil {
			return e
		}

		var locations []postgress.ShiftLocation
		if e = tx.Where("shift_id = ?", shift.ID).Order("sequence ASC").Find(&locations).Error; e != nil {
			return e
		}

		var passengers []postgress.ShiftPassenger
		if e = tx.Where("shift_id = ?", shift.ID).Where("status = ?", constants.Shift_Passenger_Active).Find(&passengers).Error; e != nil {
			return e
		}

		updated := shift

		if request.DriverId != nil {
			updated.DriverID = *request.DriverId
		}

		if request.VehicleId != nil {
			updated.VehicleID = *request.VehicleId
		}

		if request.DaysOfWeek != nil {
			updated.DaysOfWeek = database.IntArray(request.DaysOfWeek)
		}

		// a start date left as it was is fine even once it lies behind us, a running
		// shift keeps its history, a new one cannot be moved into the past
		if request.StartDate != nil && *request.StartDate != shift.StartDate {
			if *request.StartDate < database.BusinessToday() {
				return utils.Refuse(constants.Shift_In_The_Past)
			}
			updated.StartDate = *request.StartDate
		}

		if request.EndDate != nil {
			updated.EndDate = *request.EndDate
		}

		if updated.EndDate != "" && updated.EndDate < updated.StartDate {
			return utils.Refuse(fmt.Sprintf(constants.Invalid_Data, "end date, it cannot be before the start date"))
		}

		newLocations := locations
		if request.Locations != nil {
			waiting := map[int]int{}
			lowest := 0
			for _, passenger := range passengers {
				if passenger.LocationSequence > len(request.Locations) {
					waiting[passenger.LocationSequence]++
					if lowest == 0 || passenger.LocationSequence < lowest {
						lowest = passenger.LocationSequence
					}
				}
			}

			if lowest > 0 {
				return utils.Refuse(fmt.Sprintf(constants.Passenger_Stop_Missing, waiting[lowest], lowest))
			}

			newLocations = buildLocations(shift.ID, request.Locations)
			updated.StartTime = newLocations[0].Time
			updated.EndTime = newLocations[len(newLocations)-1].Time
		}

		driverChanged := updated.DriverID != shift.DriverID
		vehicleChanged := updated.VehicleID != shift.VehicleID
		scheduleChanged := !sameDays(updated.DaysOfWeek, shift.DaysOfWeek) ||
			updated.StartDate != shift.StartDate ||
			updated.EndDate != shift.EndDate ||
			updated.StartTime != shift.StartTime ||
			updated.EndTime != shift.EndTime

		passengerIds := make([]string, 0, len(passengers))
		for _, passenger := range passengers {
			passengerIds = append(passengerIds, passenger.PassengerID)
		}

		if e = database.LockShiftResources(tx, []string{updated.DriverID}, []string{updated.VehicleID}, passengerIds); e != nil {
			return e
		}

		var driver postgress.Driver
		if driverChanged {
			if driver, e = checkDriverEligible(tx, service, updated.DriverID); e != nil {
				return e
			}
		} else if e = tx.Where("id = ?", updated.DriverID).Take(&driver).Error; e != nil && !isNotFound(e) {
			return e
		}

		var vehicle postgress.Vehicle
		if vehicleChanged {
			if vehicle, e = checkVehicleEligible(tx, service.ID, updated.VehicleID); e != nil {
				return e
			}

			if vehicle.NumberOfSeats < len(passengers) {
				return utils.Refuse(fmt.Sprintf(constants.Vehicle_Capacity_Exceeded, vehicle.VehicleNumber, vehicle.NumberOfSeats, len(passengers)))
			}

			updated.SeatCapacity = vehicle.NumberOfSeats
		} else if e = tx.Where("id = ?", updated.VehicleID).Take(&vehicle).Error; e != nil && !isNotFound(e) {
			return e
		}

		// a shift is named after its vehicle
		if !utils.IsStringEmpty(vehicle.VehicleNumber) {
			updated.Name = vehicle.VehicleNumber
		}

		// the vehicle taken off the shift, so its owner can be told
		previousVehicle := vehicle
		if vehicleChanged {
			previousVehicle = postgress.Vehicle{}
			if e = tx.Where("id = ?", shift.VehicleID).Take(&previousVehicle).Error; e != nil && !isNotFound(e) {
				return e
			}
		}

		var clashDrivers, clashVehicles, clashPassengers []string
		if driverChanged || scheduleChanged {
			clashDrivers = []string{updated.DriverID}
		}
		if vehicleChanged || scheduleChanged {
			clashVehicles = []string{updated.VehicleID}
		}
		if scheduleChanged {
			clashPassengers = passengerIds
		}

		if e = checkClashes(tx, windowOf(updated), clashDrivers, clashVehicles, clashPassengers, shift.ID); e != nil {
			return e
		}

		if e = tx.Model(&postgress.Shift{}).Where("id = ?", shift.ID).Updates(map[string]interface{}{
			"name":          updated.Name,
			"driver_id":     updated.DriverID,
			"vehicle_id":    updated.VehicleID,
			"days_of_week":  updated.DaysOfWeek,
			"start_date":    updated.StartDate,
			"end_date":      updated.EndDate,
			"start_time":    updated.StartTime,
			"end_time":      updated.EndTime,
			"seat_capacity": updated.SeatCapacity,
		}).Error; e != nil {
			return e
		}

		if request.Locations != nil {
			if e = tx.Where("shift_id = ?", shift.ID).Delete(&postgress.ShiftLocation{}).Error; e != nil {
				return e
			}

			if e = tx.Create(&newLocations).Error; e != nil {
				return e
			}
		}

		if e = syncUpcomingOccurrences(tx, updated, newLocations, service, driver, vehicle); e != nil {
			return e
		}

		names, e := passengerNames(tx, passengerIds)
		if e != nil {
			return e
		}

		before = shiftContext{Shift: shift, Service: service, Vehicle: previousVehicle, Locations: locations, Passengers: passengers, Names: names}
		after = shiftContext{Shift: updated, Service: service, Driver: driver, Vehicle: vehicle, Locations: newLocations, Passengers: passengers, Names: names}

		return nil
	})

	logger.LogInfo("Response returned from updateShift", sessionId)

	return
}

// syncUpcomingOccurrences brings the trips still ahead in line with an edited shift.
// A trip on a day the shift no longer runs is dropped, the rest take the new times,
// driver, vehicle and route, and trips that have started are never touched.
func syncUpcomingOccurrences(tx *gorm.DB, shift postgress.Shift, locations []postgress.ShiftLocation, service postgress.PickDropService, driver postgress.Driver, vehicle postgress.Vehicle) error {
	now := database.BusinessNowMinute()

	var stale []string
	if err := tx.Model(&postgress.ShiftOccurrence{}).
		Where("shift_id = ?", shift.ID).
		Where("starts_at > ?", now).
		Where(`NOT (
			EXTRACT(ISODOW FROM occurrence_date::date)::int = ANY(?::integer[])
			AND occurrence_date >= ?
			AND (? = '' OR occurrence_date <= ?)
		)`, shift.DaysOfWeek, shift.StartDate, shift.EndDate, shift.EndDate).
		Pluck("id", &stale).Error; err != nil {
		return err
	}

	if len(stale) > 0 {
		if err := tx.Where("occurrence_id IN ?", stale).Delete(&postgress.ShiftAttendance{}).Error; err != nil {
			return err
		}

		if err := tx.Where("id IN ?", stale).Delete(&postgress.ShiftOccurrence{}).Error; err != nil {
			return err
		}
	}

	if err := tx.Exec(`
		UPDATE shift_occurrences
		SET start_time = ?, end_time = ?,
			starts_at = occurrence_date || ' ' || ?, ends_at = occurrence_date || ' ' || ?,
			driver_id = ?, vehicle_id = ?,
			shift_name = ?, service_name = ?, driver_name = ?, driver_mobile = ?, vehicle_number = ?,
			route = ?, updated_at = ?
		WHERE shift_id = ? AND starts_at > ?
	`,
		shift.StartTime, shift.EndTime,
		shift.StartTime, shift.EndTime,
		shift.DriverID, shift.VehicleID,
		shift.Name, service.Name, driver.DriverName, driver.DriverMobile, vehicle.VehicleNumber,
		routeJSON(locations), time.Now(),
		shift.ID, now,
	).Error; err != nil {
		return err
	}

	if err := tx.Exec(`
		UPDATE shift_attendances
		SET location = shift_locations.location, location_time = shift_locations.time, updated_at = ?
		FROM shift_occurrences, shift_locations
		WHERE shift_occurrences.id = shift_attendances.occurrence_id
		  AND shift_occurrences.shift_id = ?
		  AND shift_occurrences.starts_at > ?
		  AND shift_locations.shift_id = shift_occurrences.shift_id
		  AND shift_locations.sequence = shift_attendances.location_sequence
	`, time.Now(), shift.ID, now).Error; err != nil {
		return err
	}

	today := database.BusinessToday()

	return database.MaterializeOccurrences(tx, today, database.AddDays(today, constants.Occurrence_Horizon_Days), shift.ID)
}

// updateShiftPassengers applies additions, moves and removals as one change. The seat
// check is made against the count once everything lands, so a removal frees the seat
// an addition in the same call needs. A removed passenger's row is kept as history,
// and the passenger who takes their seat gets a row, and attendance, of their own.
func updateShiftPassengers(orgCtx *gin.Context, sessionId, ownerId string, request UpdateShiftPassengersRequest) (change passengerChange, err error) {
	logger.LogInfo("Request received in updateShiftPassengers", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		shift, service, e := lockOwnedShift(tx, ownerId, request.ShiftId, true)
		if e != nil {
			return e
		}

		var locations []postgress.ShiftLocation
		if e = tx.Where("shift_id = ?", shift.ID).Order("sequence ASC").Find(&locations).Error; e != nil {
			return e
		}

		var current []postgress.ShiftPassenger
		if e = tx.Where("shift_id = ?", shift.ID).Where("status = ?", constants.Shift_Passenger_Active).Find(&current).Error; e != nil {
			return e
		}

		onShift := map[string]postgress.ShiftPassenger{}
		for _, row := range current {
			onShift[row.PassengerID] = row
		}

		for _, passengerId := range request.Remove {
			if _, found := onShift[passengerId]; !found {
				return utils.Refuse(fmt.Sprintf(constants.Passenger_Not_On_Shift, passengerId))
			}
		}

		for _, move := range request.Move {
			if _, found := onShift[move.PassengerId]; !found {
				return utils.Refuse(fmt.Sprintf(constants.Passenger_Not_On_Shift, move.PassengerId))
			}

			if move.LocationSequence > len(locations) {
				return utils.Refuse(fmt.Sprintf(constants.Invalid_Data, fmt.Sprintf("location sequence of passenger %s, the route has %d stop(s)", move.PassengerId, len(locations))))
			}
		}

		addIds := make([]string, 0, len(request.Add))
		for _, add := range request.Add {
			if _, found := onShift[add.PassengerId]; found {
				return utils.Refuse(fmt.Sprintf(constants.Passenger_Already_On, "Passenger "+add.PassengerId))
			}

			if add.LocationSequence > len(locations) {
				return utils.Refuse(fmt.Sprintf(constants.Invalid_Data, fmt.Sprintf("location sequence of passenger %s, the route has %d stop(s)", add.PassengerId, len(locations))))
			}

			addIds = append(addIds, add.PassengerId)
		}

		if e = database.LockShiftResources(tx, nil, nil, addIds); e != nil {
			return e
		}

		passengers, e := loadEligiblePassengers(tx, service.ID, addIds)
		if e != nil {
			return e
		}

		finalCount := len(current) - len(request.Remove) + len(request.Add)
		if finalCount > shift.SeatCapacity {
			return utils.Refuse(fmt.Sprintf(constants.Vehicle_Capacity_Exceeded, vehicleNumberOf(tx, shift.VehicleID), shift.SeatCapacity, finalCount))
		}

		if len(addIds) > 0 {
			if e = checkClashes(tx, windowOf(shift), nil, nil, addIds, shift.ID); e != nil {
				return e
			}
		}

		now := time.Now()
		nowMinute := database.BusinessNowMinute()

		if len(request.Remove) > 0 {
			if e = tx.Model(&postgress.ShiftPassenger{}).
				Where("shift_id = ?", shift.ID).
				Where("status = ?", constants.Shift_Passenger_Active).
				Where("passenger_id IN ?", request.Remove).
				Updates(map[string]interface{}{
					"status":         constants.Shift_Passenger_Removed,
					"removed_by":     ownerId,
					"removed_at":     &now,
					"removal_reason": constants.Removal_Reason_Owner,
				}).Error; e != nil {
				return e
			}

			if e = database.DeleteUpcomingAttendance(tx, []string{shift.ID}, request.Remove); e != nil {
				return e
			}

			for _, passengerId := range request.Remove {
				change.Removed = append(change.Removed, onShift[passengerId])
			}
		}

		for _, move := range request.Move {
			stop := locationBySequence(locations, move.LocationSequence)
			row := onShift[move.PassengerId]

			if e = tx.Model(&postgress.ShiftPassenger{}).Where("id = ?", row.ID).Update("location_sequence", move.LocationSequence).Error; e != nil {
				return e
			}

			if e = tx.Exec(`
				UPDATE shift_attendances
				SET location_sequence = ?, location = ?, location_time = ?, updated_at = ?
				WHERE shift_id = ?
				  AND passenger_id = ?
				  AND occurrence_id IN (SELECT id FROM shift_occurrences WHERE shift_id = ? AND starts_at > ?)
			`, move.LocationSequence, stop.Location, stop.Time, now, shift.ID, move.PassengerId, shift.ID, nowMinute).Error; e != nil {
				return e
			}

			row.LocationSequence = move.LocationSequence
			change.Moved = append(change.Moved, row)
		}

		if len(request.Add) > 0 {
			rows := make([]postgress.ShiftPassenger, 0, len(request.Add))
			for _, add := range request.Add {
				rows = append(rows, postgress.ShiftPassenger{
					ID:               database.GenerateUUID(),
					ShiftID:          shift.ID,
					ServiceID:        service.ID,
					PassengerID:      add.PassengerId,
					LocationSequence: add.LocationSequence,
					Status:           constants.Shift_Passenger_Active,
					AddedBy:          ownerId,
					AddedAt:          now,
				})
			}

			if e = tx.Create(&rows).Error; e != nil {
				return e
			}
			change.Added = rows

			// trips already written ahead pick the new passengers up straight away
			if e = database.FillOccurrenceAttendance(tx, database.BusinessToday(), "9999-12-31", shift.ID); e != nil {
				return e
			}
		}

		if e = database.SyncOccupiedSeats(tx, []string{shift.ID}); e != nil {
			return e
		}

		if e = tx.Where("id = ?", shift.ID).First(&shift).Error; e != nil {
			return e
		}

		lookup := append(append([]string{}, request.Remove...), addIds...)
		for _, move := range request.Move {
			lookup = append(lookup, move.PassengerId)
		}

		names, e := passengerNames(tx, lookup)
		if e != nil {
			return e
		}

		for passengerId, passenger := range passengers {
			names[passengerId] = passenger.PassengerName
		}

		var driver postgress.Driver
		if e = tx.Where("id = ?", shift.DriverID).Take(&driver).Error; e != nil && !isNotFound(e) {
			return e
		}

		var vehicle postgress.Vehicle
		if e = tx.Where("id = ?", shift.VehicleID).Take(&vehicle).Error; e != nil && !isNotFound(e) {
			return e
		}

		change.shiftContext = shiftContext{
			Shift:     shift,
			Service:   service,
			Driver:    driver,
			Vehicle:   vehicle,
			Locations: locations,
			Names:     names,
		}

		return nil
	})

	logger.LogInfo("Response returned from updateShiftPassengers", sessionId)

	return
}

func vehicleNumberOf(tx *gorm.DB, vehicleId string) string {
	var vehicle postgress.Vehicle
	if err := tx.Select("vehicle_number").Where("id = ?", vehicleId).Take(&vehicle).Error; err != nil {
		return ""
	}

	return vehicle.VehicleNumber
}

// deleteShift archives a shift with its route and hard deletes it, the pattern rides
// follow. Trips still ahead go with it. Trips already made, their attendance and the
// shift passenger rows stay, the rows marked removed, which is what keeps travel
// history and passenger history true after the shift is gone.
func deleteShift(orgCtx *gin.Context, sessionId, ownerId, shiftId string) (result shiftContext, err error) {
	logger.LogInfo("Request received in deleteShift", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		shift, service, e := lockOwnedShift(tx, ownerId, shiftId, false)
		if e != nil {
			return e
		}

		var locations []postgress.ShiftLocation
		if e = tx.Where("shift_id = ?", shift.ID).Order("sequence ASC").Find(&locations).Error; e != nil {
			return e
		}

		var passengers []postgress.ShiftPassenger
		if e = tx.Where("shift_id = ?", shift.ID).Where("status = ?", constants.Shift_Passenger_Active).Find(&passengers).Error; e != nil {
			return e
		}

		now := time.Now()

		if e = tx.Create(&postgress.DELShift{
			ID:            shift.ID,
			ServiceID:     shift.ServiceID,
			Name:          shift.Name,
			DriverID:      shift.DriverID,
			VehicleID:     shift.VehicleID,
			DaysOfWeek:    shift.DaysOfWeek,
			StartDate:     shift.StartDate,
			EndDate:       shift.EndDate,
			StartTime:     shift.StartTime,
			EndTime:       shift.EndTime,
			SeatCapacity:  shift.SeatCapacity,
			OccupiedSeats: shift.OccupiedSeats,
			Status:        shift.Status,
			CreatedBy:     shift.CreatedBy,
			DeletedBy:     ownerId,
			DeletedAt:     now,
			CreatedAt:     shift.CreatedAt,
			UpdatedAt:     now,
		}).Error; e != nil {
			return e
		}

		if len(locations) > 0 {
			archived := make([]postgress.DELShiftLocation, 0, len(locations))
			for _, location := range locations {
				archived = append(archived, postgress.DELShiftLocation{
					ID:        location.ID,
					ShiftID:   location.ShiftID,
					Sequence:  location.Sequence,
					Location:  location.Location,
					Lat:       location.Lat,
					Lng:       location.Lng,
					Time:      location.Time,
					CreatedAt: location.CreatedAt,
				})
			}

			if e = tx.Create(&archived).Error; e != nil {
				return e
			}
		}

		var upcoming []string
		if e = tx.Model(&postgress.ShiftOccurrence{}).
			Where("shift_id = ?", shift.ID).
			Where("starts_at > ?", database.BusinessNowMinute()).
			Pluck("id", &upcoming).Error; e != nil {
			return e
		}

		if len(upcoming) > 0 {
			if e = tx.Where("occurrence_id IN ?", upcoming).Delete(&postgress.ShiftAttendance{}).Error; e != nil {
				return e
			}

			if e = tx.Where("id IN ?", upcoming).Delete(&postgress.ShiftOccurrence{}).Error; e != nil {
				return e
			}
		}

		if e = tx.Model(&postgress.ShiftPassenger{}).
			Where("shift_id = ?", shift.ID).
			Where("status = ?", constants.Shift_Passenger_Active).
			Updates(map[string]interface{}{
				"status":         constants.Shift_Passenger_Removed,
				"removed_by":     ownerId,
				"removed_at":     &now,
				"removal_reason": constants.Removal_Reason_Shift_Deleted,
			}).Error; e != nil {
			return e
		}

		if e = tx.Where("shift_id = ?", shift.ID).Delete(&postgress.ShiftLocation{}).Error; e != nil {
			return e
		}

		if e = tx.Where("id = ?", shift.ID).Delete(&postgress.Shift{}).Error; e != nil {
			return e
		}

		var vehicle postgress.Vehicle
		if e = tx.Where("id = ?", shift.VehicleID).Take(&vehicle).Error; e != nil && !isNotFound(e) {
			return e
		}

		result = shiftContext{Shift: shift, Service: service, Vehicle: vehicle, Locations: locations, Passengers: passengers}

		return nil
	})

	logger.LogInfo("Response returned from deleteShift", sessionId)

	return
}

func shiftJoins(db *gorm.DB) *gorm.DB {
	return db.Table("shifts").
		Joins("LEFT JOIN pick_drop_services ON pick_drop_services.id = shifts.service_id").
		Joins("LEFT JOIN drivers ON drivers.id = shifts.driver_id").
		Joins("LEFT JOIN vehicles ON vehicles.id = shifts.vehicle_id").
		Joins("LEFT JOIN drivers AS vehicle_owners ON vehicle_owners.id = vehicles.driver_id")
}

const shiftColumns = `
	shifts.id,
	shifts.name,
	shifts.service_id,
	COALESCE(pick_drop_services.name, '') AS service_name,
	COALESCE(pick_drop_services.owner_driver_id, '') AS owner_driver_id,
	shifts.driver_id,
	COALESCE(drivers.driver_name, '') AS driver_name,
	COALESCE(drivers.driver_mobile, '') AS driver_mobile,
	shifts.vehicle_id,
	COALESCE(vehicles.vehicle_number, '') AS vehicle_number,
	COALESCE(vehicles.vehicle_info, '') AS vehicle_info,
	COALESCE(vehicles.driver_id, '') AS vehicle_owner_id,
	COALESCE(vehicle_owners.driver_name, '') AS vehicle_owner_name,
	shifts.days_of_week,
	shifts.start_date,
	shifts.end_date,
	shifts.start_time,
	shifts.end_time,
	shifts.seat_capacity,
	shifts.occupied_seats,
	shifts.status,
	shifts.created_at,
	shifts.updated_at`

func getShiftDetails(db *gorm.DB, shiftId string) (details postgress.ShiftDetails, err error) {
	err = shiftJoins(db).Select(shiftColumns).Where("shifts.id = ?", shiftId).Take(&details).Error
	return
}

func getShiftRoute(db *gorm.DB, shiftId string) (locations []postgress.ShiftLocation, err error) {
	err = db.Where("shift_id = ?", shiftId).Order("sequence ASC").Find(&locations).Error
	return
}

func getActiveShiftPassengers(db *gorm.DB, shiftId string) (passengers []postgress.ShiftPassengerDetails, err error) {
	err = db.Table("shift_passengers").
		Select(`
			shift_passengers.id,
			shift_passengers.shift_id,
			shift_passengers.passenger_id,
			COALESCE(passengers.passenger_name, '') AS passenger_name,
			COALESCE(passengers.passenger_mobile, '') AS passenger_mobile,
			shift_passengers.location_sequence,
			shift_passengers.status,
			shift_passengers.added_at
		`).
		Joins("LEFT JOIN passengers ON passengers.id = shift_passengers.passenger_id").
		Where("shift_passengers.shift_id = ?", shiftId).
		Where("shift_passengers.status = ?", constants.Shift_Passenger_Active).
		Order("shift_passengers.location_sequence ASC, passenger_name ASC").
		Find(&passengers).Error

	return
}

// loadShiftView reads a shift with its route and passengers in three queries.
func loadShiftView(orgCtx *gin.Context, shiftId string) (
	details postgress.ShiftDetails,
	locations []postgress.ShiftLocation,
	passengers []postgress.ShiftPassengerDetails,
	err error,
) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	if details, err = getShiftDetails(db, shiftId); err != nil {
		return
	}

	if locations, err = getShiftRoute(db, shiftId); err != nil {
		return
	}

	passengers, err = getActiveShiftPassengers(db, shiftId)

	return
}

func applyShiftListFilter(query *gorm.DB, filter shiftListFilter) *gorm.DB {
	if filter.Status != constants.Shift_Status_All {
		query = query.Where("shifts.status = ?", filter.Status)
	}

	if filter.DayOfWeek > 0 {
		query = query.Where("? = ANY(shifts.days_of_week)", filter.DayOfWeek)
	}

	if !utils.IsStringEmpty(filter.Search) {
		term := "%" + filter.Search + "%"
		query = query.Where(`(
			shifts.name ILIKE ?
			OR vehicles.vehicle_number ILIKE ?
			OR EXISTS (SELECT 1 FROM shift_locations WHERE shift_locations.shift_id = shifts.id AND shift_locations.location ILIKE ?)
		)`, term, term, term)
	}

	// the same time window as the ride search: trips starting at or after the start
	// time and over by the end time
	if !utils.IsStringEmpty(filter.StartTime) {
		query = query.Where("shifts.start_time >= ?", filter.StartTime)
	}

	if !utils.IsStringEmpty(filter.EndTime) {
		query = query.Where("shifts.end_time <= ?", filter.EndTime)
	}

	return query
}

// listShifts pages through the shifts one scope may see: a service, a driver, or a
// passenger.
func listShifts(orgCtx *gin.Context, scope func(*gorm.DB) *gorm.DB, filter shiftListFilter, page int) (shifts []postgress.ShiftDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	query := applyShiftListFilter(scope(shiftJoins(database.DatabaseConn.Postgres.WithContext(ctx))), filter)

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Select(shiftColumns).
		Order("shifts.start_time ASC, shifts.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&shifts).Error

	return
}

// getShiftHistory lists every shift the service ever had, live, completed and deleted,
// the deleted ones read back from their archive.
func getShiftHistory(orgCtx *gin.Context, serviceId string, page int) (shifts []postgress.ShiftHistoryDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	params := map[string]interface{}{
		"service": serviceId,
		"deleted": constants.Shift_Status_Deleted,
		"limit":   limit,
		"offset":  offset,
	}

	history := `
		SELECT shifts.id, shifts.name,
			COALESCE(drivers.driver_name, '') AS driver_name,
			COALESCE(vehicles.vehicle_number, '') AS vehicle_number,
			shifts.days_of_week, shifts.start_date, shifts.end_date, shifts.start_time, shifts.end_time,
			shifts.status, shifts.created_at, NULL::timestamptz AS deleted_at
		FROM shifts
		LEFT JOIN drivers ON drivers.id = shifts.driver_id
		LEFT JOIN vehicles ON vehicles.id = shifts.vehicle_id
		WHERE shifts.service_id = @service
		UNION ALL
		SELECT del_shifts.id, del_shifts.name,
			COALESCE(drivers.driver_name, del_drivers.driver_name, ''),
			COALESCE(vehicles.vehicle_number, ''),
			del_shifts.days_of_week, del_shifts.start_date, del_shifts.end_date, del_shifts.start_time, del_shifts.end_time,
			@deleted, del_shifts.created_at, del_shifts.deleted_at
		FROM del_shifts
		LEFT JOIN drivers ON drivers.id = del_shifts.driver_id
		LEFT JOIN del_drivers ON del_drivers.id = del_shifts.driver_id
		LEFT JOIN vehicles ON vehicles.id = del_shifts.vehicle_id
		WHERE del_shifts.service_id = @service`

	if err = db.Raw("SELECT COUNT(*) FROM ("+history+") history", params).Scan(&totalRows).Error; err != nil {
		return
	}

	err = db.Raw("SELECT * FROM ("+history+") history ORDER BY history.created_at DESC LIMIT @limit OFFSET @offset", params).Scan(&shifts).Error

	return
}

// getPassengerHistory lists who travelled on the service's shifts. It is read from the
// shift passenger rows, which are never deleted, and from the membership rows, which
// are never reused, so a passenger who left, a passenger who was replaced and a shift
// that was deleted all stay on it. The joined date is the approval of the membership
// the passenger held when they were put on the shift.
func getPassengerHistory(orgCtx *gin.Context, serviceId, passengerId string, page int) (history []postgress.ShiftParticipationDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	params := map[string]interface{}{
		"service":   serviceId,
		"passenger": passengerId,
		"deleted":   constants.Shift_Status_Deleted,
		"present":   constants.Attendance_Present,
		"absent":    constants.Attendance_Absent,
		"now":       database.BusinessNowMinute(),
		"members":   []string{constants.Membership_Status_Approved, constants.Membership_Status_Left, constants.Membership_Status_Removed, constants.Membership_Status_Closed},
		"limit":     limit,
		"offset":    offset,
	}

	where := "WHERE shift_passengers.service_id = @service"
	if !utils.IsStringEmpty(passengerId) {
		where += " AND shift_passengers.passenger_id = @passenger"
	}

	if err = db.Raw("SELECT COUNT(*) FROM shift_passengers "+where, params).Scan(&totalRows).Error; err != nil {
		return
	}

	err = db.Raw(`
		SELECT
			shift_passengers.id,
			shift_passengers.shift_id,
			COALESCE(shifts.name, del_shifts.name, '') AS shift_name,
			CASE WHEN shifts.id IS NOT NULL THEN shifts.status WHEN del_shifts.id IS NOT NULL THEN @deleted ELSE '' END AS shift_status,
			shift_passengers.passenger_id,
			COALESCE(passengers.passenger_name, del_passengers.passenger_name, '') AS passenger_name,
			COALESCE(passengers.passenger_mobile, '') AS passenger_mobile,
			shift_passengers.location_sequence,
			shift_passengers.status,
			shift_passengers.removal_reason,
			(
				SELECT MAX(pick_drop_passengers.decided_at) FROM pick_drop_passengers
				WHERE pick_drop_passengers.service_id = shift_passengers.service_id
				  AND pick_drop_passengers.passenger_id = shift_passengers.passenger_id
				  AND pick_drop_passengers.status IN @members
				  AND pick_drop_passengers.decided_at <= shift_passengers.added_at
			) AS joined_service_at,
			shift_passengers.added_at,
			shift_passengers.removed_at,
			(
				SELECT COUNT(*) FROM shift_attendances
				JOIN shift_occurrences ON shift_occurrences.id = shift_attendances.occurrence_id
				WHERE shift_attendances.shift_id = shift_passengers.shift_id
				  AND shift_attendances.passenger_id = shift_passengers.passenger_id
				  AND shift_attendances.status = @present
				  AND shift_occurrences.starts_at <= @now
			) AS trips_present,
			(
				SELECT COUNT(*) FROM shift_attendances
				JOIN shift_occurrences ON shift_occurrences.id = shift_attendances.occurrence_id
				WHERE shift_attendances.shift_id = shift_passengers.shift_id
				  AND shift_attendances.passenger_id = shift_passengers.passenger_id
				  AND shift_attendances.status = @absent
				  AND shift_occurrences.starts_at <= @now
			) AS trips_absent
		FROM shift_passengers
		LEFT JOIN shifts ON shifts.id = shift_passengers.shift_id
		LEFT JOIN del_shifts ON del_shifts.id = shift_passengers.shift_id
		LEFT JOIN passengers ON passengers.id = shift_passengers.passenger_id
		LEFT JOIN del_passengers ON del_passengers.id = shift_passengers.passenger_id
		`+where+`
		ORDER BY shift_passengers.added_at DESC
		LIMIT @limit OFFSET @offset`, params).Scan(&history).Error

	return
}

// getOccurrences lists trips with their attendance counted, scoped to a service or to
// the driver who drove them.
func getOccurrences(orgCtx *gin.Context, scopeColumn, scopeValue, shiftId string, page int) (occurrences []postgress.ShiftOccurrenceDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).
		Table("shift_occurrences").
		Where("shift_occurrences."+scopeColumn+" = ?", scopeValue)

	if !utils.IsStringEmpty(shiftId) {
		query = query.Where("shift_occurrences.shift_id = ?", shiftId)
	}

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Select(`
			shift_occurrences.id,
			shift_occurrences.shift_id,
			shift_occurrences.shift_name,
			shift_occurrences.service_name,
			shift_occurrences.occurrence_date,
			shift_occurrences.start_time,
			shift_occurrences.end_time,
			shift_occurrences.driver_name,
			shift_occurrences.vehicle_number,
			shift_occurrences.status,
			(SELECT COUNT(*) FROM shift_attendances WHERE shift_attendances.occurrence_id = shift_occurrences.id AND shift_attendances.status = ?) AS present_count,
			(SELECT COUNT(*) FROM shift_attendances WHERE shift_attendances.occurrence_id = shift_occurrences.id AND shift_attendances.status = ?) AS absent_count
		`, constants.Attendance_Present, constants.Attendance_Absent).
		Order("shift_occurrences.starts_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&occurrences).Error

	return
}

func getOccurrence(db *gorm.DB, shiftId, date string) (occurrence postgress.ShiftOccurrence, found bool, err error) {
	err = db.Where("shift_id = ?", shiftId).Where("occurrence_date = ?", date).First(&occurrence).Error
	if isNotFound(err) {
		return occurrence, false, nil
	}

	return occurrence, err == nil, err
}

func getOccurrenceAttendance(db *gorm.DB, occurrenceId string) (attendance []postgress.ShiftAttendanceDetails, err error) {
	err = db.Table("shift_attendances").
		Select(`
			shift_attendances.passenger_id,
			COALESCE(passengers.passenger_name, del_passengers.passenger_name, '') AS passenger_name,
			COALESCE(passengers.passenger_mobile, '') AS passenger_mobile,
			shift_attendances.location_sequence,
			shift_attendances.location,
			shift_attendances.location_time,
			shift_attendances.status,
			shift_attendances.marked_at
		`).
		Joins("LEFT JOIN passengers ON passengers.id = shift_attendances.passenger_id").
		Joins("LEFT JOIN del_passengers ON del_passengers.id = shift_attendances.passenger_id").
		Where("shift_attendances.occurrence_id = ?", occurrenceId).
		Order("shift_attendances.location_sequence ASC, passenger_name ASC").
		Find(&attendance).Error

	return
}

func getServiceOwner(db *gorm.DB, serviceId string) (ownerId string, err error) {
	var service postgress.PickDropService
	if err = db.Select("owner_driver_id").Where("id = ?", serviceId).Take(&service).Error; err != nil {
		return
	}

	return service.OwnerDriverID, nil
}

// attendanceView reads who travels on one trip and whether the caller may see it. The
// owner, the driver of the day and the passengers of that trip may, and a passenger
// sees the others' names and attendance but not their numbers. A trip the scheduler has
// not written yet is projected from the shift as it stands, everybody present.
func attendanceView(orgCtx *gin.Context, userId, shiftId, date string, asPassenger bool) (resp AttendanceResponse, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	live := true
	details, err := getShiftDetails(db, shiftId)
	if err != nil {
		if !isNotFound(err) {
			return
		}
		live, err = false, nil
	}

	occurrence, found, err := getOccurrence(db, shiftId, date)
	if err != nil {
		return
	}

	if !live && !found {
		err = utils.Refuse(constants.Shift_Not_Found)
		return
	}

	var entries []postgress.ShiftAttendanceDetails

	if found {
		if entries, err = getOccurrenceAttendance(db, occurrence.ID); err != nil {
			return
		}

		resp = AttendanceResponse{
			ShiftId:   shiftId,
			ShiftName: occurrence.ShiftName,
			Date:      date,
			StartTime: occurrence.StartTime,
			EndTime:   occurrence.EndTime,
			Status:    occurrence.Status,
		}
	} else {
		if details.Status != constants.Shift_Status_Active || !database.OccursOn(detailsWindow(details), date) || date < database.BusinessToday() {
			err = utils.Refuse(fmt.Sprintf(constants.Not_An_Occurrence, date))
			return
		}

		passengers, e := getActiveShiftPassengers(db, shiftId)
		if e != nil {
			err = e
			return
		}

		locations, e := getShiftRoute(db, shiftId)
		if e != nil {
			err = e
			return
		}

		for _, passenger := range passengers {
			stop := locationBySequence(locations, passenger.LocationSequence)
			entries = append(entries, postgress.ShiftAttendanceDetails{
				PassengerID:      passenger.PassengerID,
				PassengerName:    passenger.PassengerName,
				PassengerMobile:  passenger.PassengerMobile,
				LocationSequence: passenger.LocationSequence,
				Location:         stop.Location,
				LocationTime:     stop.Time,
				Status:           constants.Attendance_Present,
			})
		}

		resp = AttendanceResponse{
			ShiftId:   shiftId,
			ShiftName: details.Name,
			Date:      date,
			StartTime: details.StartTime,
			EndTime:   details.EndTime,
			Status:    constants.Occurrence_Status_Scheduled,
		}
	}

	allowed := false

	if asPassenger {
		for _, entry := range entries {
			if entry.PassengerID == userId {
				allowed = true
			}
		}
	} else {
		if live && (details.OwnerDriverID == userId || details.DriverID == userId) {
			allowed = true
		}

		if found && occurrence.DriverID == userId {
			allowed = true
		}

		if found && !allowed {
			if ownerId, e := getServiceOwner(db, occurrence.ServiceID); e == nil && ownerId == userId {
				allowed = true
			}
		}
	}

	if !allowed {
		err = utils.Refuse(constants.Operation_Not_Permitted)
		return
	}

	resp.Attendance = attendanceEntries(entries, userId, asPassenger)

	return
}

// markAttendance records a passenger's absence, or their return, for one trip. The
// trip is written ahead on the spot if the scheduler has not reached it yet, so an
// absence can be marked days in advance. Only a trip that has not started can change.
func markAttendance(orgCtx *gin.Context, sessionId, passengerId string, request MarkAttendanceRequest) (change attendanceChange, err error) {
	logger.LogInfo("Request received in markAttendance", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var shift postgress.Shift
		if e := tx.Where("id = ?", request.ShiftId).Where("status = ?", constants.Shift_Status_Active).First(&shift).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Shift_Not_Found)
			}
			return e
		}

		var seat postgress.ShiftPassenger
		if e := tx.Where("shift_id = ?", shift.ID).
			Where("passenger_id = ?", passengerId).
			Where("status = ?", constants.Shift_Passenger_Active).
			First(&seat).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Not_Shift_Passenger)
			}
			return e
		}

		if !database.OccursOn(windowOf(shift), request.Date) {
			return utils.Refuse(fmt.Sprintf(constants.Not_An_Occurrence, request.Date))
		}

		if request.Date+" "+shift.StartTime <= database.BusinessNowMinute() {
			return utils.Refuse(constants.Occurrence_Started)
		}

		if e := database.MaterializeOccurrences(tx, request.Date, request.Date, shift.ID); e != nil {
			return e
		}

		if e := tx.Where("shift_id = ?", shift.ID).Where("occurrence_date = ?", request.Date).First(&change.Occurrence).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(fmt.Sprintf(constants.Not_An_Occurrence, request.Date))
			}
			return e
		}

		if e := forUpdate(tx).
			Where("occurrence_id = ?", change.Occurrence.ID).
			Where("passenger_id = ?", passengerId).
			First(&change.Attendance).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Not_Shift_Passenger)
			}
			return e
		}

		if change.Attendance.Status == request.Status {
			return nil
		}

		now := time.Now()
		if e := tx.Model(&postgress.ShiftAttendance{}).Where("id = ?", change.Attendance.ID).Updates(map[string]interface{}{
			"status":    request.Status,
			"marked_at": &now,
		}).Error; e != nil {
			return e
		}

		change.Changed = true
		change.Attendance.Status = request.Status

		ownerId, e := getServiceOwner(tx, shift.ServiceID)
		if e != nil {
			return e
		}
		change.OwnerId = ownerId

		names, e := passengerNames(tx, []string{passengerId})
		if e != nil {
			return e
		}
		change.PassengerName = names[passengerId]

		return nil
	})

	logger.LogInfo("Response returned from markAttendance", sessionId)

	return
}

// sendDriverUpdate stores what a driver tells the passengers of today's trip. Only the
// driver of the shift may send it and it only reaches passengers travelling today, an
// absent passenger or somebody not on the shift never gets it.
func sendDriverUpdate(orgCtx *gin.Context, sessionId, driverId string, request DriverLocationRequest) (result driverUpdateResult, err error) {
	logger.LogInfo("Request received in sendDriverUpdate", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var shift postgress.Shift
		if e := tx.Where("id = ?", request.ShiftId).Where("status = ?", constants.Shift_Status_Active).First(&shift).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Shift_Not_Found)
			}
			return e
		}

		if shift.DriverID != driverId {
			return utils.Refuse(constants.Not_Shift_Driver)
		}

		today := database.BusinessToday()
		if !database.OccursOn(windowOf(shift), today) {
			return utils.Refuse(fmt.Sprintf(constants.Not_An_Occurrence, today))
		}

		if e := database.MaterializeOccurrences(tx, today, today, shift.ID); e != nil {
			return e
		}

		if e := tx.Where("shift_id = ?", shift.ID).Where("occurrence_date = ?", today).First(&result.Occurrence).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(fmt.Sprintf(constants.Not_An_Occurrence, today))
			}
			return e
		}

		if e := tx.Model(&postgress.ShiftAttendance{}).
			Where("occurrence_id = ?", result.Occurrence.ID).
			Where("status = ?", constants.Attendance_Present).
			Pluck("passenger_id", &result.Recipients).Error; e != nil {
			return e
		}

		result.Update = postgress.ShiftDriverUpdate{
			ID:           database.GenerateUUID(),
			ShiftID:      shift.ID,
			OccurrenceID: result.Occurrence.ID,
			DriverID:     driverId,
			Message:      request.Message,
			Location:     request.Location,
			Lat:          request.Lat,
			Lng:          request.Lng,
			Recipients:   len(result.Recipients),
		}

		return tx.Create(&result.Update).Error
	})

	logger.LogInfo("Response returned from sendDriverUpdate", sessionId)

	return
}

func createShiftRequest(orgCtx *gin.Context, sessionId, passengerId string, request ShiftRequestInput) (requestId string, err error) {
	logger.LogInfo("Request received in createShiftRequest", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var passenger postgress.Passenger
		if e := tx.Where("id = ?", passengerId).Where("status = ?", constants.Status_Active).First(&passenger).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Passenger_Not_Found)
			}
			return e
		}

		if !utils.IsStringEmpty(request.ServiceId) {
			var service postgress.PickDropService
			if e := tx.Where("id = ?", request.ServiceId).Where("status = ?", constants.Status_Active).First(&service).Error; e != nil {
				if isNotFound(e) {
					return utils.Refuse(constants.Service_Not_Found)
				}
				return e
			}
		}

		points := utils.RoutePoints(request.Locations)

		row := postgress.ShiftRequest{
			ID:            database.GenerateUUID(),
			PassengerID:   passengerId,
			ServiceID:     request.ServiceId,
			ContactNumber: request.ContactNumber,
			Note:          request.Note,
			DaysOfWeek:    database.IntArray(request.DaysOfWeek),
			StartTime:     request.Locations[0].Time,
			EndTime:       request.Locations[len(request.Locations)-1].Time,
			StartLocation: points[0],
			EndLocation:   points[len(points)-1],
			RoutePoints:   pq.StringArray(points),
		}

		if e := tx.Create(&row).Error; e != nil {
			return e
		}
		requestId = row.ID

		locations := make([]postgress.ShiftRequestLocation, 0, len(request.Locations))
		for index, location := range request.Locations {
			locations = append(locations, postgress.ShiftRequestLocation{
				ID:             database.GenerateUUID(),
				ShiftRequestID: row.ID,
				Sequence:       index + 1,
				Location:       location.Location,
				Lat:            location.Lat,
				Lng:            location.Lng,
				Time:           location.Time,
			})
		}

		return tx.Create(&locations).Error
	})

	logger.LogInfo("Response returned from createShiftRequest", sessionId)

	return
}

// getShiftRequests is both a passenger's own list and the owners' search, matching
// places along the whole requested route the way the ride request search does.
func getShiftRequests(orgCtx *gin.Context, scope func(*gorm.DB) *gorm.DB, filter shiftRequestSearch, page int) (
	requests []postgress.ShiftRequestDetails,
	locations []postgress.ShiftRequestLocation,
	totalRows int64,
	err error,
) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	query := scope(db.Table("shift_requests").
		Joins("LEFT JOIN passengers ON passengers.id = shift_requests.passenger_id").
		Joins("LEFT JOIN pick_drop_services ON pick_drop_services.id = shift_requests.service_id"))

	if !utils.IsStringEmpty(filter.Search) {
		term := "%" + filter.Search + "%"
		query = query.Where(`(
			shift_requests.start_location ILIKE ?
			OR shift_requests.end_location ILIKE ?
			OR EXISTS (SELECT 1 FROM unnest(shift_requests.route_points) AS point WHERE point ILIKE ?)
		)`, term, term, term)
	}

	if filter.DayOfWeek > 0 {
		query = query.Where("? = ANY(shift_requests.days_of_week)", filter.DayOfWeek)
	}

	if !utils.IsStringEmpty(filter.StartTime) {
		query = query.Where("shift_requests.start_time >= ?", filter.StartTime)
	}

	if !utils.IsStringEmpty(filter.EndTime) {
		query = query.Where("shift_requests.end_time <= ?", filter.EndTime)
	}

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	if err = query.Select(`
			shift_requests.id,
			shift_requests.passenger_id,
			COALESCE(passengers.passenger_name, '') AS passenger_name,
			COALESCE(shift_requests.service_id, '') AS service_id,
			COALESCE(pick_drop_services.name, '') AS service_name,
			shift_requests.contact_number,
			shift_requests.note,
			shift_requests.days_of_week,
			shift_requests.start_time,
			shift_requests.end_time,
			shift_requests.created_at
		`).
		Order("shift_requests.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&requests).Error; err != nil {
		return
	}

	if len(requests) == 0 {
		return
	}

	requestIds := make([]string, 0, len(requests))
	for _, request := range requests {
		requestIds = append(requestIds, request.ID)
	}

	err = db.Where("shift_request_id IN ?", requestIds).
		Order("shift_request_id ASC, sequence ASC").
		Find(&locations).Error

	return
}

func deleteShiftRequest(orgCtx *gin.Context, sessionId, passengerId, requestId string) (err error) {
	logger.LogInfo("Request received in deleteShiftRequest", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row postgress.ShiftRequest
		if e := forUpdate(tx).Where("id = ?", requestId).Where("passenger_id = ?", passengerId).First(&row).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Request_Not_Found)
			}
			return e
		}

		return database.ArchiveShiftRequests(tx, []postgress.ShiftRequest{row}, []string{row.ID}, passengerId)
	})

	logger.LogInfo("Response returned from deleteShiftRequest", sessionId)

	return
}

// getTravelHistory lists the trips a passenger was on, newest first, read entirely from
// the frozen occurrences so it stays true after the shift changes or is deleted.
func getTravelHistory(orgCtx *gin.Context, passengerId string, page int) (trips []postgress.TravelHistoryDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).
		Table("shift_attendances").
		Joins("JOIN shift_occurrences ON shift_occurrences.id = shift_attendances.occurrence_id").
		Where("shift_attendances.passenger_id = ?", passengerId).
		Where("shift_occurrences.starts_at <= ?", database.BusinessNowMinute())

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Select(`
			shift_occurrences.id AS occurrence_id,
			shift_occurrences.shift_id,
			shift_occurrences.shift_name,
			shift_occurrences.service_name,
			shift_occurrences.occurrence_date,
			shift_occurrences.start_time,
			shift_occurrences.end_time,
			shift_occurrences.driver_name,
			shift_occurrences.driver_mobile,
			shift_occurrences.vehicle_number,
			shift_occurrences.route,
			shift_attendances.location_sequence,
			shift_attendances.location,
			shift_attendances.location_time,
			shift_attendances.status
		`).
		Order("shift_occurrences.starts_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&trips).Error

	return
}

func notifyShiftCreated(ctx *gin.Context, sessionId, actorId string, result shiftContext) {
	name := displayName(result.Shift.Name)
	schedule := scheduleLabel(result.Shift.DaysOfWeek, result.Shift.StartDate, result.Shift.EndDate)

	var notices []notice

	if result.Shift.DriverID != actorId {
		notices = append(notices, driverNotice(result.Shift.DriverID, constants.NOTIFICATION_TYPE_SHIFT_ASSIGNED, constants.NOTIFICATION_TITLE_SHIFT_ASSIGNED,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_DRIVER_ASSIGNED, name, result.Service.Name, schedule,
				result.Shift.StartTime, result.Shift.EndTime, result.Vehicle.VehicleNumber)))
	}

	for _, passenger := range result.Passengers {
		stop := locationBySequence(result.Locations, passenger.LocationSequence)
		notices = append(notices, passengerNotice(passenger.PassengerID, constants.NOTIFICATION_TYPE_SHIFT_CREATED, constants.NOTIFICATION_TITLE_SHIFT_CREATED,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_PASSENGER_ADDED, name, result.Service.Name, schedule,
				stop.Location, stop.Time, result.Driver.DriverName, result.Vehicle.VehicleNumber)))
	}

	notices = append(notices, vehicleOwnerNotices(result.Vehicle, result.Shift.DriverID, actorId,
		constants.NOTIFICATION_TYPE_SHIFT_ASSIGNED, constants.NOTIFICATION_TITLE_SHIFT_ASSIGNED,
		fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_VEHICLE_ASSIGNED, result.Vehicle.VehicleNumber, name, result.Service.Name, schedule,
			result.Shift.StartTime, result.Shift.EndTime, result.Driver.DriverName))...)

	sendNotices(ctx, sessionId, shiftData(result.Shift.ID, result.Service.ID), notices)
}

func notifyShiftUpdated(ctx *gin.Context, sessionId, actorId string, before, after shiftContext) {
	name := displayName(after.Shift.Name)
	schedule := scheduleLabel(after.Shift.DaysOfWeek, after.Shift.StartDate, after.Shift.EndDate)
	updated := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_UPDATED, name, after.Service.Name, schedule,
		after.Shift.StartTime, after.Shift.EndTime, after.Driver.DriverName, after.Vehicle.VehicleNumber)

	var notices []notice

	if before.Shift.DriverID != after.Shift.DriverID {
		if before.Shift.DriverID != actorId {
			notices = append(notices, driverNotice(before.Shift.DriverID, constants.NOTIFICATION_TYPE_SHIFT_ASSIGNED, constants.NOTIFICATION_TITLE_SHIFT_ASSIGNED,
				fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_DRIVER_UNASSIGNED, name, after.Service.Name)))
		}

		if after.Shift.DriverID != actorId {
			notices = append(notices, driverNotice(after.Shift.DriverID, constants.NOTIFICATION_TYPE_SHIFT_ASSIGNED, constants.NOTIFICATION_TITLE_SHIFT_ASSIGNED,
				fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_DRIVER_ASSIGNED, name, after.Service.Name, schedule,
					after.Shift.StartTime, after.Shift.EndTime, after.Vehicle.VehicleNumber)))
		}
	} else if after.Shift.DriverID != actorId {
		notices = append(notices, driverNotice(after.Shift.DriverID, constants.NOTIFICATION_TYPE_SHIFT_UPDATED, constants.NOTIFICATION_TITLE_SHIFT_UPDATED, updated))
	}

	for _, passenger := range after.Passengers {
		notices = append(notices, passengerNotice(passenger.PassengerID, constants.NOTIFICATION_TYPE_SHIFT_UPDATED, constants.NOTIFICATION_TITLE_SHIFT_UPDATED, updated))
	}

	if before.Vehicle.ID != after.Vehicle.ID {
		notices = append(notices, vehicleOwnerNotices(before.Vehicle, before.Shift.DriverID, actorId,
			constants.NOTIFICATION_TYPE_SHIFT_ASSIGNED, constants.NOTIFICATION_TITLE_SHIFT_ASSIGNED,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_VEHICLE_UNASSIGNED, before.Vehicle.VehicleNumber, displayName(before.Shift.Name), after.Service.Name))...)

		notices = append(notices, vehicleOwnerNotices(after.Vehicle, after.Shift.DriverID, actorId,
			constants.NOTIFICATION_TYPE_SHIFT_ASSIGNED, constants.NOTIFICATION_TITLE_SHIFT_ASSIGNED,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_VEHICLE_ASSIGNED, after.Vehicle.VehicleNumber, name, after.Service.Name, schedule,
				after.Shift.StartTime, after.Shift.EndTime, after.Driver.DriverName))...)
	} else {
		notices = append(notices, vehicleOwnerNotices(after.Vehicle, after.Shift.DriverID, actorId,
			constants.NOTIFICATION_TYPE_SHIFT_UPDATED, constants.NOTIFICATION_TITLE_SHIFT_UPDATED, updated)...)
	}

	sendNotices(ctx, sessionId, shiftData(after.Shift.ID, after.Service.ID), notices)
}

func notifyPassengersChanged(ctx *gin.Context, sessionId, actorId string, change passengerChange) {
	name := displayName(change.Shift.Name)
	schedule := scheduleLabel(change.Shift.DaysOfWeek, change.Shift.StartDate, change.Shift.EndDate)
	notificationType := constants.NOTIFICATION_TYPE_SHIFT_PASSENGER
	title := constants.NOTIFICATION_TITLE_SHIFT_PASSENGER

	var notices []notice

	for _, passenger := range change.Added {
		stop := locationBySequence(change.Locations, passenger.LocationSequence)
		notices = append(notices, passengerNotice(passenger.PassengerID, notificationType, title,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_PASSENGER_ADDED, name, change.Service.Name, schedule,
				stop.Location, stop.Time, change.Driver.DriverName, change.Vehicle.VehicleNumber)))
	}

	for _, passenger := range change.Removed {
		notices = append(notices, passengerNotice(passenger.PassengerID, notificationType, title,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_PASSENGER_REMOVED, name, change.Service.Name)))
	}

	for _, passenger := range change.Moved {
		stop := locationBySequence(change.Locations, passenger.LocationSequence)
		notices = append(notices, passengerNotice(passenger.PassengerID, notificationType, title,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_PASSENGER_MOVED, name, change.Service.Name, stop.Location, stop.Time)))
	}

	if change.Shift.DriverID != actorId {
		notices = append(notices, driverNotice(change.Shift.DriverID, notificationType, title,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_PASSENGERS_CHANGE, name, change.Service.Name,
				change.Shift.OccupiedSeats, change.Shift.SeatCapacity)))
	}

	sendNotices(ctx, sessionId, shiftData(change.Shift.ID, change.Service.ID), notices)
}

func notifyShiftDeleted(ctx *gin.Context, sessionId, actorId string, result shiftContext) {
	message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_DELETED, displayName(result.Shift.Name), result.Service.Name)
	notificationType := constants.NOTIFICATION_TYPE_SHIFT_DELETED
	title := constants.NOTIFICATION_TITLE_SHIFT_DELETED

	var notices []notice

	if result.Shift.DriverID != actorId {
		notices = append(notices, driverNotice(result.Shift.DriverID, notificationType, title, message))
	}

	for _, passenger := range result.Passengers {
		notices = append(notices, passengerNotice(passenger.PassengerID, notificationType, title, message))
	}

	notices = append(notices, vehicleOwnerNotices(result.Vehicle, result.Shift.DriverID, actorId, notificationType, title, message)...)

	sendNotices(ctx, sessionId, shiftData(result.Shift.ID, result.Service.ID), notices)
}

// notifyAttendance tells the driver of that trip and the owner, once each even when
// they are the same person.
func notifyAttendance(ctx *gin.Context, sessionId string, change attendanceChange) {
	if !change.Changed {
		return
	}

	template := constants.NOTIFICATION_MESSAGE_SHIFT_ABSENT
	if change.Attendance.Status == constants.Attendance_Present {
		template = constants.NOTIFICATION_MESSAGE_SHIFT_PRESENT
	}

	message := fmt.Sprintf(template, change.PassengerName, displayName(change.Occurrence.ShiftName),
		database.WeekdayDate(change.Occurrence.OccurrenceDate), change.Attendance.Location, change.Attendance.LocationTime)

	data := shiftData(change.Occurrence.ShiftID, change.Occurrence.ServiceID)
	data[constants.NOTIFICATION_KEY_OCCURRENCE_DATE] = change.Occurrence.OccurrenceDate

	var notices []notice
	seen := map[string]bool{}

	for _, recipient := range []string{change.Occurrence.DriverID, change.OwnerId} {
		if utils.IsStringEmpty(recipient) || seen[recipient] {
			continue
		}
		seen[recipient] = true

		notices = append(notices, driverNotice(recipient, constants.NOTIFICATION_TYPE_SHIFT_ABSENCE, constants.NOTIFICATION_TITLE_SHIFT_ABSENCE, message))
	}

	sendNotices(ctx, sessionId, data, notices)
}

func notifyDriverUpdate(ctx *gin.Context, sessionId string, request DriverLocationRequest, result driverUpdateResult) {
	occurrence := result.Occurrence

	text := request.Message
	if !utils.IsStringEmpty(request.Location) {
		if utils.IsStringEmpty(text) {
			text = "I am at " + request.Location
		} else {
			text += ", at " + request.Location
		}
	}

	message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_DRIVER_UPDATE, occurrence.DriverName, occurrence.ServiceName,
		occurrence.VehicleNumber, displayName(occurrence.ShiftName), text, database.BusinessNow().Format(constants.Minute_Datetime_Layout))

	data := shiftData(occurrence.ShiftID, occurrence.ServiceID)
	data[constants.NOTIFICATION_KEY_OCCURRENCE_DATE] = occurrence.OccurrenceDate
	data[constants.NOTIFICATION_KEY_LOCATION] = request.Location
	data[constants.NOTIFICATION_KEY_LAT] = fmt.Sprintf("%f", request.Lat)
	data[constants.NOTIFICATION_KEY_LNG] = fmt.Sprintf("%f", request.Lng)

	notices := make([]notice, 0, len(result.Recipients))
	for _, passengerId := range result.Recipients {
		notices = append(notices, passengerNotice(passengerId, constants.NOTIFICATION_TYPE_SHIFT_DRIVER_UPDATE, constants.NOTIFICATION_TITLE_SHIFT_DRIVER, message))
	}

	sendNotices(ctx, sessionId, data, notices)
}
