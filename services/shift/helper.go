package shift

import (
	"context"
	"errors"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func withTimeout(orgCtx *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
}

// requirePermission is the gate every managing action goes through. It asks the
// database what the caller's role may do rather than testing a role name, so a role
// an admin invents later works here with no code change.
func requirePermission(ctx *gin.Context, sessionId, groupId, driverId, permission string) error {
	hasPermission, err := database.HasGroupPermission(ctx, groupId, driverId, permission)
	if err != nil {
		logger.LogError(sessionId, "failed to check the permission error: "+err.Error())
		return errors.New(constants.Unknown_Error)
	}

	if !hasPermission {
		err = errors.New(constants.Operation_Not_Permitted)
		logger.LogError(sessionId, fmt.Sprintf("driver lacks %s error: %s", permission, err.Error()))
		return err
	}

	return nil
}

// getApprovedGroupVehicle loads a vehicle only if that vehicle has actually been
// let into the fleet, a shift can never be built on an outside vehicle.
func getApprovedGroupVehicle(orgCtx *gin.Context, groupId, vehicleId string) (vehicle postgress.Vehicle, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Table("vehicles").
		Select("vehicles.*").
		Joins("JOIN group_vehicles ON group_vehicles.vehicle_id = vehicles.id").
		Where("group_vehicles.group_id = ?", groupId).
		Where("group_vehicles.vehicle_id = ?", vehicleId).
		Where("group_vehicles.status = ?", constants.Membership_Status_Approved).
		Where("vehicles.status = ?", constants.Status_Active).
		First(&vehicle).Error

	return
}

// getEligiblePassengers loads, in one query, the passengers of the given ids that
// are actually approved members of the fleet. Anybody missing from the result is
// simply not allowed on this shift.
func getEligiblePassengers(orgCtx *gin.Context, groupId string, passengerIds []string) (passengers map[string]postgress.Passenger, err error) {
	passengers = map[string]postgress.Passenger{}

	if len(passengerIds) == 0 {
		return
	}

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	var rows []postgress.Passenger
	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Table("passengers").
		Select("passengers.*").
		Joins("JOIN group_passengers ON group_passengers.passenger_id = passengers.id").
		Where("group_passengers.group_id = ?", groupId).
		Where("group_passengers.status = ?", constants.Membership_Status_Approved).
		Where("passengers.id IN ?", passengerIds).
		Where("passengers.status = ?", constants.Status_Active).
		Find(&rows).Error

	for _, row := range rows {
		passengers[row.ID] = row
	}

	return
}

// checkClashes enforces the rules that keep one vehicle and one driver in one place
// at a time. Both the fleet's own shifts and the carpool rides are checked, since a
// driver is the same person in both parts of the product.
func checkClashes(ctx *gin.Context, sessionId, vehicleId, driverId, startDatetime, endDatetime, excludeShiftId string) error {
	hasShift, err := database.VehicleHasShiftDuringTime(ctx, vehicleId, startDatetime, endDatetime, excludeShiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to check the vehicle shifts error: "+err.Error())
		return errors.New(constants.Unknown_Error)
	}
	if hasShift {
		err = errors.New(constants.Vehicle_Busy)
		logger.LogError(sessionId, err)
		return err
	}

	hasRide, err := database.VehicleHasRideDuringTime(ctx, vehicleId, startDatetime, endDatetime)
	if err != nil {
		logger.LogError(sessionId, "failed to check the vehicle rides error: "+err.Error())
		return errors.New(constants.Unknown_Error)
	}
	if hasRide {
		err = errors.New(constants.Vehicle_Busy)
		logger.LogError(sessionId, err)
		return err
	}

	hasDriverShift, err := database.DriverHasShiftDuringTime(ctx, driverId, startDatetime, endDatetime, excludeShiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to check the driver shifts error: "+err.Error())
		return errors.New(constants.Unknown_Error)
	}
	if hasDriverShift {
		err = errors.New(constants.Driver_Busy)
		logger.LogError(sessionId, err)
		return err
	}

	hasDriverRide, err := database.DriverHasRideDuringTime(ctx, driverId, startDatetime, endDatetime)
	if err != nil {
		logger.LogError(sessionId, "failed to check the driver rides error: "+err.Error())
		return errors.New(constants.Unknown_Error)
	}
	if hasDriverRide {
		err = errors.New(constants.Driver_Busy)
		logger.LogError(sessionId, err)
		return err
	}

	return nil
}

// checkPassengersFree refuses to seat somebody who is already aboard another shift
// at that hour. A passenger cannot be in two vehicles at once any more than a
// driver or a vehicle can, which the clash rules already covered for those two.
func checkPassengersFree(ctx *gin.Context, sessionId string, passengerIds []string, startDatetime, endDatetime, excludeShiftId string, names map[string]postgress.Passenger) error {
	busy, err := database.PassengersBusyDuringTime(ctx, passengerIds, startDatetime, endDatetime, excludeShiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to check the passenger shifts error: "+err.Error())
		return errors.New(constants.Unknown_Error)
	}

	for passengerId, clashStart := range busy {
		name := passengerId
		if passenger, ok := names[passengerId]; ok {
			name = passenger.PassengerName
		}

		err = fmt.Errorf(constants.Passenger_Busy, name, clashStart)
		logger.LogError(sessionId, err)
		return err
	}

	return nil
}

// flattenSeatPlan pulls the seats out of the nested stop payload and remembers
// which stop each one belongs to.
func flattenSeatPlan(stops []StopRequest) (plan []seatPlan) {
	for stopIndex, stop := range stops {
		for _, seat := range stop.Seats {
			plan = append(plan, seatPlan{
				SeatNumber:  seat.SeatNumber,
				Gender:      seat.Gender,
				PassengerId: seat.PassengerId,
				StopIndex:   stopIndex,
			})
		}
	}

	return
}

// checkSeatsAgainstVehicle makes sure the plan fits the vehicle it is being put in,
// and that every seated passenger matches the gender the seat is kept for.
func checkSeatsAgainstVehicle(plan []seatPlan, vehicle postgress.Vehicle, passengers map[string]postgress.Passenger) error {
	for _, seat := range plan {
		if seat.SeatNumber > vehicle.NumberOfSeats {
			return fmt.Errorf("seat %d does not exist, %s has %d seat(s)", seat.SeatNumber, vehicle.VehicleNumber, vehicle.NumberOfSeats)
		}

		if utils.IsStringEmpty(seat.PassengerId) {
			continue
		}

		passenger, ok := passengers[seat.PassengerId]
		if !ok {
			return errors.New(constants.Not_Group_Member)
		}

		if !strings.EqualFold(passenger.Gender, seat.Gender) {
			return fmt.Errorf(constants.Seat_Gender_Mismatch, seat.SeatNumber, seat.Gender)
		}
	}

	return nil
}

// createShiftWithStopsAndSeats writes the whole trip in one transaction: the shift,
// its route in order, and a row for every seat of the vehicle so an empty seat is
// just as real as a taken one. A template is stamped out of the same payload when
// the manager asked for one.
func createShiftWithStopsAndSeats(
	orgCtx *gin.Context,
	sessionId, createdByDriverId string,
	vehicle postgress.Vehicle,
	plan []seatPlan,
	request CreateShiftRequest,
) (shiftId, templateId string, err error) {
	logger.LogInfo("Request received in createShiftWithStopsAndSeats", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()
	if tx.Error != nil {
		return shiftId, templateId, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	seatsTaken := 0
	for _, seat := range plan {
		if !utils.IsStringEmpty(seat.PassengerId) {
			seatsTaken++
		}
	}

	shift := postgress.Shift{
		ID:                   database.GenerateUUID(),
		GroupID:              request.GroupId,
		VehicleID:            request.VehicleId,
		DriverID:             request.DriverId,
		Direction:            request.Direction,
		StartDatetime:        request.StartDatetime,
		EstimatedEndDatetime: request.EstimatedEndDatetime,
		StartLocation:        strings.TrimSpace(request.StartLocation),
		EndLocation:          strings.TrimSpace(request.EndLocation),
		NumberOfSeats:        vehicle.NumberOfSeats,
		SeatsTaken:           seatsTaken,
		RouteDetails:         request.RouteDetails,
		CreatedByDriverID:    createdByDriverId,
		IsActive:             true,
	}

	if err = tx.Create(&shift).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}
	shiftId = shift.ID

	// the stops keep the order the manager arranged them in, that order is the route
	stopIds := make([]string, 0, len(request.Stops))
	stops := make([]postgress.ShiftStop, 0, len(request.Stops))
	for index, stop := range request.Stops {
		stopId := database.GenerateUUID()
		stopIds = append(stopIds, stopId)

		stops = append(stops, postgress.ShiftStop{
			ID:             stopId,
			ShiftID:        shift.ID,
			SequenceNumber: index + 1,
			Location:       strings.TrimSpace(stop.Location),
			Lat:            stop.Lat,
			Lng:            stop.Lng,
			ScheduledTime:  stop.ScheduledTime,
		})
	}

	if err = tx.Create(&stops).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	assignedSeats := map[int]seatPlan{}
	for _, seat := range plan {
		assignedSeats[seat.SeatNumber] = seat
	}

	// every physical seat of the vehicle gets a row, the free ones included, so the
	// manager can see and fill the gaps later
	seats := make([]postgress.ShiftSeat, 0, vehicle.NumberOfSeats)
	for seatNumber := 1; seatNumber <= vehicle.NumberOfSeats; seatNumber++ {
		seat := postgress.ShiftSeat{
			ID:         database.GenerateUUID(),
			ShiftID:    shift.ID,
			SeatNumber: seatNumber,
			Status:     constants.Seat_Status_Empty,
		}

		if planned, ok := assignedSeats[seatNumber]; ok {
			seat.Gender = planned.Gender
			if !utils.IsStringEmpty(planned.PassengerId) {
				seat.PassengerID = planned.PassengerId
				seat.StopID = stopIds[planned.StopIndex]
				seat.Status = constants.Seat_Status_Assigned
			}
		}

		seats = append(seats, seat)
	}

	if err = tx.Create(&seats).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	if request.MakeTemplate {
		templateId, err = createTemplateInTx(tx, sessionId, shift, request, plan)
		if err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}

		if err = tx.Model(&postgress.Shift{}).Where("id = ?", shift.ID).Update("template_id", templateId).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}
	}

	err = tx.Commit().Error

	logger.LogInfo("Response returned from createShiftWithStopsAndSeats", sessionId)

	return
}

// createTemplateInTx saves the shape of the shift so it can be pulled back into the
// create form later. The template seats point at their stop by sequence rather than
// by id, so the template stays usable for a brand new shift.
func createTemplateInTx(tx *gorm.DB, sessionId string, shift postgress.Shift, request CreateShiftRequest, plan []seatPlan) (templateId string, err error) {
	daysOfWeek := make(pq.Int64Array, 0, len(request.DaysOfWeek))
	for _, day := range request.DaysOfWeek {
		daysOfWeek = append(daysOfWeek, int64(day))
	}

	template := postgress.ShiftTemplate{
		ID:                database.GenerateUUID(),
		ShiftID:           shift.ID,
		GroupID:           shift.GroupID,
		Name:              request.TemplateName,
		VehicleID:         shift.VehicleID,
		DriverID:          shift.DriverID,
		Direction:         shift.Direction,
		StartDatetime:     shift.StartDatetime,
		EstimatedEndTime:  shift.EstimatedEndDatetime,
		StartLocation:     shift.StartLocation,
		EndLocation:       shift.EndLocation,
		NumberOfSeats:     shift.NumberOfSeats,
		RouteDetails:      shift.RouteDetails,
		DaysOfWeek:        daysOfWeek,
		CreatedByDriverID: shift.CreatedByDriverID,
	}

	if err = tx.Create(&template).Error; err != nil {
		return
	}
	templateId = template.ID

	templateStops := make([]postgress.ShiftTemplateStop, 0, len(request.Stops))
	for index, stop := range request.Stops {
		templateStops = append(templateStops, postgress.ShiftTemplateStop{
			ID:              database.GenerateUUID(),
			ShiftTemplateID: template.ID,
			SequenceNumber:  index + 1,
			Location:        strings.TrimSpace(stop.Location),
			Lat:             stop.Lat,
			Lng:             stop.Lng,
			ScheduledTime:   stop.ScheduledTime,
		})
	}

	if err = tx.Create(&templateStops).Error; err != nil {
		return
	}

	if len(plan) == 0 {
		return
	}

	templateSeats := make([]postgress.ShiftTemplateSeat, 0, len(plan))
	for _, seat := range plan {
		templateSeats = append(templateSeats, postgress.ShiftTemplateSeat{
			ID:              database.GenerateUUID(),
			ShiftTemplateID: template.ID,
			SeatNumber:      seat.SeatNumber,
			Gender:          seat.Gender,
			PassengerID:     seat.PassengerId,
			StopSequence:    seat.StopIndex + 1,
		})
	}

	err = tx.Create(&templateSeats).Error

	return
}

func getShiftById(orgCtx *gin.Context, shiftId string) (shift postgress.Shift, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("id = ?", shiftId).
		Where("is_active = ?", true).
		First(&shift).Error

	return
}

const shiftSelectClause = `
	shifts.id,
	shifts.group_id,
	groups.name AS group_name,
	shifts.vehicle_id,
	vehicles.vehicle_number,
	vehicles.vehicle_info,
	vehicles.has_ac,
	vehicles.has_heating,
	shifts.driver_id,
	drivers.driver_name,
	drivers.driver_mobile,
	drivers.rating,
	shifts.direction,
	shifts.start_datetime,
	shifts.estimated_end_datetime,
	shifts.start_location,
	shifts.end_location,
	shifts.number_of_seats,
	shifts.seats_taken,
	shifts.route_details,
	shifts.created_by_driver_id,
	creators.driver_name AS created_by_name,
	creators.driver_mobile AS created_by_mobile,
	shifts.is_active,
	shifts.created_at,
	shifts.updated_at
`

// shiftDetailsQuery joins in both the driver of the day and the manager who built
// the shift, so every read already carries the two numbers a rider might need.
func shiftDetailsQuery(db *gorm.DB) *gorm.DB {
	return db.Table("shifts").
		Select(shiftSelectClause).
		Joins("JOIN groups ON groups.id = shifts.group_id").
		Joins("JOIN vehicles ON vehicles.id = shifts.vehicle_id").
		Joins("JOIN drivers ON drivers.id = shifts.driver_id").
		Joins("JOIN drivers AS creators ON creators.id = shifts.created_by_driver_id")
}

func getShiftDetails(orgCtx *gin.Context, shiftId string) (shift postgress.ShiftDetails, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = shiftDetailsQuery(database.DatabaseConn.Postgres.WithContext(ctx)).
		Where("shifts.id = ?", shiftId).
		First(&shift).Error

	return
}

func getShiftStops(orgCtx *gin.Context, shiftId string) (stops []postgress.ShiftStop, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("shift_id = ?", shiftId).
		Order("sequence_number ASC").
		Find(&stops).Error

	return
}

// getShiftSeats reads the seats with the passenger and the stop already joined on,
// so a shift is shown in one query rather than one per seat.
func getShiftSeats(orgCtx *gin.Context, shiftId string) (seats []postgress.ShiftSeatDetails, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Table("shift_seats").
		Select(`
			shift_seats.id,
			shift_seats.shift_id,
			shift_seats.seat_number,
			shift_seats.gender,
			shift_seats.status,
			shift_seats.passenger_id,
			passengers.passenger_name,
			passengers.passenger_mobile,
			shift_seats.stop_id,
			shift_stops.sequence_number,
			shift_stops.location,
			shift_stops.lat,
			shift_stops.lng,
			shift_stops.scheduled_time
		`).
		Joins("LEFT JOIN passengers ON passengers.id = shift_seats.passenger_id").
		Joins("LEFT JOIN shift_stops ON shift_stops.id = shift_seats.stop_id").
		Where("shift_seats.shift_id = ?", shiftId).
		Order("shift_seats.seat_number ASC").
		Find(&seats).Error

	return
}

// getShiftsByGroup is the roster view a manager works from.
func getShiftsByGroup(orgCtx *gin.Context, groupId, driverId, direction, startTime, endTime, status string, page int) (shifts []postgress.ShiftDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	query := shiftDetailsQuery(database.DatabaseConn.Postgres.WithContext(ctx)).
		Where("shifts.group_id = ?", groupId)

	if !utils.IsStringEmpty(driverId) {
		query = query.Where("shifts.driver_id = ?", driverId)
	}

	if !utils.IsStringEmpty(direction) {
		query = query.Where("shifts.direction = ?", direction)
	}

	if !utils.IsStringEmpty(startTime) {
		query = query.Where("shifts.start_datetime >= ?", startTime)
	}

	if !utils.IsStringEmpty(endTime) {
		query = query.Where("shifts.estimated_end_datetime <= ?", endTime)
	}

	if strings.EqualFold(status, constants.Ride_Status_Active) {
		query = query.Where("shifts.is_active = ?", true)
	} else if strings.EqualFold(status, constants.Ride_Status_InActive) {
		query = query.Where("shifts.is_active = ?", false)
	}

	countQuery := query.Session(&gorm.Session{})

	if err = countQuery.Count(&totalRows).Error; err != nil {
		return
	}

	err = query.
		Order("shifts.start_datetime ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&shifts).Error

	return
}

// getShiftsForUser answers "what am I on" for both sides of a shift in one query,
// the driver behind the wheel and the passenger holding a seat.
func getShiftsForUser(orgCtx *gin.Context, userId, direction, startTime, endTime string, page int) (shifts []postgress.ShiftDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	query := shiftDetailsQuery(database.DatabaseConn.Postgres.WithContext(ctx)).
		Where("shifts.is_active = ?", true).
		Where(`
			shifts.driver_id = ?
			OR EXISTS (
				SELECT 1 FROM shift_seats
				WHERE shift_seats.shift_id = shifts.id
				  AND shift_seats.passenger_id = ?
				  AND shift_seats.status = ?
			)
		`, userId, userId, constants.Seat_Status_Assigned)

	if !utils.IsStringEmpty(direction) {
		query = query.Where("shifts.direction = ?", direction)
	}

	if !utils.IsStringEmpty(startTime) {
		query = query.Where("shifts.start_datetime >= ?", startTime)
	}

	if !utils.IsStringEmpty(endTime) {
		query = query.Where("shifts.estimated_end_datetime <= ?", endTime)
	}

	countQuery := query.Session(&gorm.Session{})

	if err = countQuery.Count(&totalRows).Error; err != nil {
		return
	}

	err = query.
		Order("shifts.start_datetime ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&shifts).Error

	return
}

func isSeatedOnShift(orgCtx *gin.Context, shiftId, passengerId string) (isSeated bool, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	var count int64
	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.ShiftSeat{}).
		Where("shift_id = ?", shiftId).
		Where("passenger_id = ?", passengerId).
		Count(&count).Error

	return count > 0, err
}

// applySeatUpdates re-seats a shift in one transaction and keeps the taken count in
// step with what actually changed.
func applySeatUpdates(orgCtx *gin.Context, sessionId, shiftId string, updates []SeatUpdate, stopsBySequence map[int]string, seatsTaken int) (err error) {
	logger.LogInfo("Request received in applySeatUpdates", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, seat := range updates {
		values := map[string]interface{}{
			"gender":       seat.Gender,
			"passenger_id": seat.PassengerId,
			"status":       constants.Seat_Status_Empty,
			"stop_id":      "",
		}

		if !utils.IsStringEmpty(seat.PassengerId) {
			values["status"] = constants.Seat_Status_Assigned
			values["stop_id"] = stopsBySequence[seat.StopSequence]
		}

		if err = tx.Model(&postgress.ShiftSeat{}).
			Where("shift_id = ?", shiftId).
			Where("seat_number = ?", seat.SeatNumber).
			Updates(values).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}
	}

	if err = tx.Model(&postgress.Shift{}).
		Where("id = ?", shiftId).
		Update("seats_taken", seatsTaken).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	err = tx.Commit().Error

	logger.LogInfo("Response returned from applySeatUpdates", sessionId)

	return
}

func cancelShift(orgCtx *gin.Context, sessionId, shiftId string) (err error) {
	logger.LogInfo("Request received in cancelShift", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Shift{}).
		Where("id = ?", shiftId).
		Update("is_active", false).Error

	logger.LogInfo("Response returned from cancelShift", sessionId)

	return
}

// checkShiftCancellationGuard refuses a cancellation that lands too close to the
// departure, mirroring the guard the carpool rides already use. A start time we
// cannot read fails closed.
func checkShiftCancellationGuard(shift postgress.Shift, now time.Time) (blocked bool, message string) {
	startDate, err := time.ParseInLocation(constants.DateTimeLayout, shift.StartDatetime, time.Local)
	if err != nil {
		return true, fmt.Sprintf(constants.Invalid_Data, "shift start time")
	}

	cutoff := now.Add(time.Duration(constants.Shift_Cancel_Min_Hours_Before_Start) * time.Hour)
	if startDate.Before(cutoff) {
		return true, fmt.Sprintf("this shift starts in less than %d hour(s) and cannot be cancelled", constants.Shift_Cancel_Min_Hours_Before_Start)
	}

	return false, ""
}

func getShiftTemplates(orgCtx *gin.Context, groupId string) (templates []postgress.ShiftTemplate, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("group_id = ?", groupId).
		Preload("Vehicle").
		Preload("Stops", func(db *gorm.DB) *gorm.DB {
			return db.Order("shift_template_stops.sequence_number ASC")
		}).
		Preload("Seats", func(db *gorm.DB) *gorm.DB {
			return db.Order("shift_template_seats.seat_number ASC")
		}).
		Order("created_at DESC").
		Find(&templates).Error

	return
}

func getShiftTemplateById(orgCtx *gin.Context, templateId string) (template postgress.ShiftTemplate, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("id = ?", templateId).
		First(&template).Error

	return
}

func deleteShiftTemplate(orgCtx *gin.Context, sessionId, templateId string) (err error) {
	logger.LogInfo("Request received in deleteShiftTemplate", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Where("shift_template_id = ?", templateId).Delete(&postgress.ShiftTemplateSeat{}).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	if err = tx.Where("shift_template_id = ?", templateId).Delete(&postgress.ShiftTemplateStop{}).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	if err = tx.Where("id = ?", templateId).Delete(&postgress.ShiftTemplate{}).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	err = tx.Commit().Error

	logger.LogInfo("Response returned from deleteShiftTemplate", sessionId)

	return
}

func passengerIdsFromPlan(plan []seatPlan) (passengerIds []string) {
	for _, seat := range plan {
		if !utils.IsStringEmpty(seat.PassengerId) {
			passengerIds = append(passengerIds, seat.PassengerId)
		}
	}

	return
}

// mergeSeatState works out how the whole shift will look once the update lands, and
// refuses a change that would leave one passenger holding two seats. Judging only
// the seats in the payload would miss somebody who is already sitting elsewhere on
// the same trip.
func mergeSeatState(existing []postgress.ShiftSeatDetails, updates []SeatUpdate) (final map[int]string, err error) {
	final = map[int]string{}

	for _, seat := range existing {
		final[seat.SeatNumber] = seat.PassengerID
	}

	for _, seat := range updates {
		final[seat.SeatNumber] = seat.PassengerId
	}

	seen := map[string]int{}
	for seatNumber, passengerId := range final {
		if utils.IsStringEmpty(passengerId) {
			continue
		}

		if other, ok := seen[passengerId]; ok {
			return nil, fmt.Errorf("a passenger cannot hold seat %d and seat %d on the same shift", other, seatNumber)
		}
		seen[passengerId] = seatNumber
	}

	return
}

// notifyShift tells the driver and everybody seated what their trip looks like.
// Both the driver's number and the number of the manager who built the shift travel
// with the message, so anybody with a problem on the day knows who to call.
func notifyShift(ctx *gin.Context, sessionId, shiftId, notificationType, title string) {
	shift, err := getShiftDetails(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to load the shift for the notification error: "+err.Error())
		return
	}

	seats, err := getShiftSeats(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to load the seats for the notification error: "+err.Error())
		return
	}

	stops, err := getShiftStops(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to load the stops for the notification error: "+err.Error())
		return
	}

	data := map[string]string{
		constants.NOTIFICATION_KEY_SHIFT_ID: shift.ID,
		constants.NOTIFICATION_KEY_GROUP_ID: shift.GroupID,
	}

	driverMessage := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_DRIVER,
		shift.Direction,
		shift.StartDatetime,
		shift.VehicleNumber,
		shift.SeatsTaken,
		len(stops),
		shift.StartLocation,
		shift.CreatedByName,
		shift.CreatedByMobile,
	)

	utils.SendUserNotification(ctx, sessionId, notificationType, shift.DriverID, constants.User_Driver, title, driverMessage, data)

	for _, seat := range seats {
		if utils.IsStringEmpty(seat.PassengerID) {
			continue
		}

		utils.SendUserNotification(ctx, sessionId, notificationType, seat.PassengerID, constants.User_Passenger, title,
			passengerShiftMessage(shift, seat), data)
	}
}

func passengerShiftMessage(shift postgress.ShiftDetails, seat postgress.ShiftSeatDetails) string {
	return fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_PASSENGER,
		shift.Direction,
		shift.StartDatetime,
		shift.VehicleNumber,
		seat.Location,
		seat.ScheduledTime,
		shift.DriverName,
		shift.DriverMobile,
		shift.CreatedByName,
		shift.CreatedByMobile,
	)
}

// notifyShiftSeatChange tells the people whose place on the trip actually changed.
// Somebody who kept the same seat is left alone, only the newly seated and the
// people who lost their seat hear about it, along with the driver.
func notifyShiftSeatChange(ctx *gin.Context, sessionId, shiftId string, previousSeats []postgress.ShiftSeatDetails) {
	shift, err := getShiftDetails(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to load the shift for the notification error: "+err.Error())
		return
	}

	currentSeats, err := getShiftSeats(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to load the seats for the notification error: "+err.Error())
		return
	}

	stops, err := getShiftStops(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to load the stops for the notification error: "+err.Error())
		return
	}

	before := map[string]bool{}
	for _, seat := range previousSeats {
		if !utils.IsStringEmpty(seat.PassengerID) {
			before[seat.PassengerID] = true
		}
	}

	data := map[string]string{
		constants.NOTIFICATION_KEY_SHIFT_ID: shift.ID,
		constants.NOTIFICATION_KEY_GROUP_ID: shift.GroupID,
	}

	now := map[string]bool{}
	for _, seat := range currentSeats {
		if utils.IsStringEmpty(seat.PassengerID) {
			continue
		}
		now[seat.PassengerID] = true

		if before[seat.PassengerID] {
			continue
		}

		utils.SendUserNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_SHIFT_UPDATED, seat.PassengerID, constants.User_Passenger,
			constants.NOTIFICATION_TITLE_SHIFT_UPDATED, passengerShiftMessage(shift, seat), data)
	}

	for passengerId := range before {
		if now[passengerId] {
			continue
		}

		message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_CANCELLED,
			shift.Direction,
			shift.StartDatetime,
			shift.VehicleNumber,
			shift.CreatedByName,
			shift.CreatedByMobile,
		)

		utils.SendUserNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_SHIFT_UPDATED, passengerId, constants.User_Passenger,
			constants.NOTIFICATION_TITLE_SHIFT_UPDATED, message, data)
	}

	driverMessage := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_DRIVER,
		shift.Direction,
		shift.StartDatetime,
		shift.VehicleNumber,
		shift.SeatsTaken,
		len(stops),
		shift.StartLocation,
		shift.CreatedByName,
		shift.CreatedByMobile,
	)

	utils.SendUserNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_SHIFT_UPDATED, shift.DriverID, constants.User_Driver,
		constants.NOTIFICATION_TITLE_SHIFT_UPDATED, driverMessage, data)
}

// notifyShiftCancelled tells the driver and everybody who held a seat that the trip
// is off, so nobody is left standing at a stop.
func notifyShiftCancelled(ctx *gin.Context, sessionId, shiftId string, seats []postgress.ShiftSeatDetails) {
	shift, err := getShiftDetails(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to load the shift for the notification error: "+err.Error())
		return
	}

	data := map[string]string{
		constants.NOTIFICATION_KEY_SHIFT_ID: shift.ID,
		constants.NOTIFICATION_KEY_GROUP_ID: shift.GroupID,
	}

	message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_CANCELLED,
		shift.Direction,
		shift.StartDatetime,
		shift.VehicleNumber,
		shift.CreatedByName,
		shift.CreatedByMobile,
	)

	utils.SendUserNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_SHIFT_CANCELLED, shift.DriverID, constants.User_Driver,
		constants.NOTIFICATION_TITLE_SHIFT_CANCELLED, message, data)

	for _, seat := range seats {
		if utils.IsStringEmpty(seat.PassengerID) {
			continue
		}

		utils.SendUserNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_SHIFT_CANCELLED, seat.PassengerID, constants.User_Passenger,
			constants.NOTIFICATION_TITLE_SHIFT_CANCELLED, message, data)
	}
}

// applyShiftReschedule writes the moved window, the route text and any nudged stop
// times in one transaction. The stops are matched by sequence number so their ids,
// and therefore the seats hanging off them, survive untouched.
func applyShiftReschedule(orgCtx *gin.Context, sessionId, shiftId string, updates map[string]interface{}, stopTimes []StopTimeUpdate) (err error) {
	logger.LogInfo("Request received in applyShiftReschedule", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if len(updates) > 0 {
		if err = tx.Model(&postgress.Shift{}).Where("id = ?", shiftId).Updates(updates).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}
	}

	for _, stop := range stopTimes {
		if err = tx.Model(&postgress.ShiftStop{}).
			Where("shift_id = ?", shiftId).
			Where("sequence_number = ?", stop.SequenceNumber).
			Update("scheduled_time", stop.ScheduledTime).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}
	}

	err = tx.Commit().Error

	logger.LogInfo("Response returned from applyShiftReschedule", sessionId)

	return
}
