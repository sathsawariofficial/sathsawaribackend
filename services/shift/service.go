package shift

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateShift builds one directional trip of one vehicle on one day. Everything is
// checked before a single row is written: who is allowed to build it, whether the
// vehicle and the driver belong to the fleet, whether the seats fit the vehicle and
// the people in them, and whether either the vehicle or the driver is already
// promised somewhere else at that hour, on another shift or on a carpool ride.
func CreateShift(ctx *gin.Context, sessionId, createdByDriverId string, request CreateShiftRequest) (shiftId, templateId string, err error) {
	logger.LogInfo("Request received in CreateShift", sessionId)

	if _, err = database.GetGroupById(ctx, request.GroupId); err != nil {
		logger.LogError(sessionId, "failed to get group error: "+err.Error())
		err = errors.New(constants.Group_Not_Found)
		return
	}

	if err = requirePermission(ctx, sessionId, request.GroupId, createdByDriverId, constants.PERMISSION_SHIFT_CREATE); err != nil {
		return
	}

	vehicle, err := getApprovedGroupVehicle(ctx, request.GroupId, request.VehicleId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the group vehicle error: "+err.Error())
		err = errors.New(constants.Vehicle_Not_Found)
		return
	}

	if vehicle.NumberOfSeats <= 0 {
		err = errors.New(constants.Vehicle_Seats_Missing)
		logger.LogError(sessionId, err)
		return
	}

	isMember, err := database.IsApprovedGroupMember(ctx, request.GroupId, request.DriverId)
	if err != nil {
		logger.LogError(sessionId, "failed to check the driver membership error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}
	if !isMember {
		err = errors.New(constants.Not_Group_Member)
		logger.LogError(sessionId, "the assigned driver is not in the group error: "+err.Error())
		return
	}

	plan := flattenSeatPlan(request.Stops)

	passengers, err := getEligiblePassengers(ctx, request.GroupId, passengerIdsFromPlan(plan))
	if err != nil {
		logger.LogError(sessionId, "failed to load the passengers error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	if err = checkSeatsAgainstVehicle(plan, vehicle, passengers); err != nil {
		logger.LogError(sessionId, "seat check error: "+err.Error())
		return
	}

	if err = checkClashes(ctx, sessionId, request.VehicleId, request.DriverId, request.StartDatetime, request.EstimatedEndDatetime, ""); err != nil {
		return
	}

	shiftId, templateId, err = createShiftWithStopsAndSeats(ctx, sessionId, createdByDriverId, vehicle, plan, request)
	if err != nil {
		logger.LogError(sessionId, "failed to create shift error: "+err.Error())
		err = fmt.Errorf(constants.Creation_Failed, "shift")
		return
	}

	notifyShift(ctx, sessionId, shiftId, constants.NOTIFICATION_TYPE_SHIFT_CREATED, constants.NOTIFICATION_TITLE_SHIFT_CREATED)

	logger.LogInfo("Response returned from CreateShift", sessionId)
	logger.LogDebug2("Response returned from CreateShift", sessionId, shiftId)

	return
}

// UpdateShiftSeats re-seats an existing trip in one call. The same rules as at
// creation apply, a seat kept for one gender never takes the other and nobody is
// seated twice.
func UpdateShiftSeats(ctx *gin.Context, sessionId, driverId string, request UpdateShiftSeatsRequest) (err error) {
	logger.LogInfo("Request received in UpdateShiftSeats", sessionId)

	shift, err := getShiftById(ctx, request.ShiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to get shift error: "+err.Error())
		err = errors.New(constants.Shift_Not_Found)
		return
	}

	if err = requirePermission(ctx, sessionId, shift.GroupID, driverId, constants.PERMISSION_SHIFT_ASSIGN_SEATS); err != nil {
		return
	}

	stops, err := getShiftStops(ctx, shift.ID)
	if err != nil {
		logger.LogError(sessionId, "failed to get the stops error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	stopsBySequence := map[int]string{}
	for _, stop := range stops {
		stopsBySequence[stop.SequenceNumber] = stop.ID
	}

	existingSeats, err := getShiftSeats(ctx, shift.ID)
	if err != nil {
		logger.LogError(sessionId, "failed to get the seats error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	passengerIds := []string{}
	for _, seat := range request.Seats {
		if !utils.IsStringEmpty(seat.PassengerId) {
			passengerIds = append(passengerIds, seat.PassengerId)
		}

		if seat.SeatNumber > shift.NumberOfSeats {
			err = fmt.Errorf("seat %d does not exist on this shift, it has %d seat(s)", seat.SeatNumber, shift.NumberOfSeats)
			logger.LogError(sessionId, err)
			return
		}

		if !utils.IsStringEmpty(seat.PassengerId) {
			if _, ok := stopsBySequence[seat.StopSequence]; !ok {
				err = fmt.Errorf(constants.Invalid_Data, "stop sequence")
				logger.LogError(sessionId, err)
				return
			}
		}
	}

	passengers, err := getEligiblePassengers(ctx, shift.GroupID, passengerIds)
	if err != nil {
		logger.LogError(sessionId, "failed to load the passengers error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	// judge the whole shift as it will look after the change, not just the seats in
	// the payload, otherwise somebody could be seated twice across two calls
	finalSeats, err := mergeSeatState(existingSeats, request.Seats)
	if err != nil {
		logger.LogError(sessionId, "seat merge error: "+err.Error())
		return
	}

	for _, seat := range request.Seats {
		if utils.IsStringEmpty(seat.PassengerId) {
			continue
		}

		passenger, ok := passengers[seat.PassengerId]
		if !ok {
			err = errors.New(constants.Not_Group_Member)
			logger.LogError(sessionId, "a passenger is not in the group error: "+err.Error())
			return
		}

		// the seat's gender and the person sitting in it always have to agree. The
		// manager may re-declare that gender in the same call, which is how a male
		// rider is swapped for a female one in one go, what can never happen is a
		// passenger ending up in a seat kept for the other gender.
		if !strings.EqualFold(passenger.Gender, seat.Gender) {
			err = fmt.Errorf(constants.Seat_Gender_Mismatch, seat.SeatNumber, seat.Gender)
			logger.LogError(sessionId, err)
			return
		}
	}

	seatsTaken := 0
	for _, passengerId := range finalSeats {
		if !utils.IsStringEmpty(passengerId) {
			seatsTaken++
		}
	}

	if err = applySeatUpdates(ctx, sessionId, shift.ID, request.Seats, stopsBySequence, seatsTaken); err != nil {
		logger.LogError(sessionId, "failed to update the seats error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "seats")
		return
	}

	// everybody who was on this trip before and everybody on it now needs to be told
	notifyShiftSeatChange(ctx, sessionId, shift.ID, existingSeats)

	logger.LogInfo("Response returned from UpdateShiftSeats", sessionId)

	return
}

// GetShift returns one trip with its route in order and its seats. Anybody inside
// the fleet may look, and so may a passenger who holds a seat on it.
func GetShift(ctx *gin.Context, sessionId, userId, shiftId string) (
	shift postgress.ShiftDetails,
	stops []postgress.ShiftStop,
	seats []postgress.ShiftSeatDetails,
	err error,
) {
	logger.LogInfo("Request received in GetShift", sessionId)

	shift, err = getShiftDetails(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to get shift error: "+err.Error())
		err = errors.New(constants.Shift_Not_Found)
		return
	}

	isMember, err := database.IsApprovedGroupMember(ctx, shift.GroupID, userId)
	if err != nil {
		logger.LogError(sessionId, "failed to check the membership error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	if !isMember {
		var isSeated bool
		isSeated, err = isSeatedOnShift(ctx, shiftId, userId)
		if err != nil {
			logger.LogError(sessionId, "failed to check the seat error: "+err.Error())
			err = errors.New(constants.Unknown_Error)
			return
		}

		if !isSeated {
			err = errors.New(constants.Operation_Not_Permitted)
			logger.LogError(sessionId, err)
			return
		}
	}

	if stops, err = getShiftStops(ctx, shiftId); err != nil {
		logger.LogError(sessionId, "failed to get the stops error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shift")
		return
	}

	if seats, err = getShiftSeats(ctx, shiftId); err != nil {
		logger.LogError(sessionId, "failed to get the seats error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shift")
		return
	}

	logger.LogInfo("Response returned from GetShift", sessionId)

	return
}

// GetGroupShifts is the roster a manager works from.
func GetGroupShifts(ctx *gin.Context, sessionId, driverId, groupId, filterDriverId, direction, startTime, endTime, status string, page int) (
	shifts []postgress.ShiftDetails,
	totalRows int64,
	err error,
) {
	logger.LogInfo("Request received in GetGroupShifts", sessionId)

	isMember, err := database.IsApprovedGroupMember(ctx, groupId, driverId)
	if err != nil {
		logger.LogError(sessionId, "failed to check the membership error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}
	if !isMember {
		err = errors.New(constants.Not_Group_Member)
		logger.LogError(sessionId, err)
		return
	}

	shifts, totalRows, err = getShiftsByGroup(ctx, groupId, filterDriverId, direction, startTime, endTime, status, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the shifts error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shifts")
		return
	}

	logger.LogInfo("Response returned from GetGroupShifts", sessionId)

	return
}

// GetMyShifts answers "what am I on" for whoever is asking, the driver behind the
// wheel or the passenger holding a seat.
func GetMyShifts(ctx *gin.Context, sessionId, userId, direction, startTime, endTime string, page int) (
	shifts []postgress.ShiftDetails,
	totalRows int64,
	err error,
) {
	logger.LogInfo("Request received in GetMyShifts", sessionId)

	shifts, totalRows, err = getShiftsForUser(ctx, userId, direction, startTime, endTime, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the shifts error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shifts")
		return
	}

	logger.LogInfo("Response returned from GetMyShifts", sessionId)

	return
}

// CancelShift calls a trip off and tells everybody who was counting on it. It
// refuses this close to departure, the same way a carpool ride does.
func CancelShift(ctx *gin.Context, sessionId, driverId, shiftId string) (err error) {
	logger.LogInfo("Request received in CancelShift", sessionId)

	shift, err := getShiftById(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to get shift error: "+err.Error())
		err = errors.New(constants.Shift_Not_Found)
		return
	}

	if err = requirePermission(ctx, sessionId, shift.GroupID, driverId, constants.PERMISSION_SHIFT_CANCEL); err != nil {
		return
	}

	if blocked, message := checkShiftCancellationGuard(shift, time.Now()); blocked {
		err = errors.New(message)
		logger.LogError(sessionId, "shift cancellation blocked: "+message)
		return
	}

	// the seats have to be read before the shift goes inactive, they are who we owe
	// the message to
	seats, err := getShiftSeats(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the seats error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	if err = cancelShift(ctx, sessionId, shiftId); err != nil {
		logger.LogError(sessionId, "failed to cancel shift error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "cancel the shift")
		return
	}

	notifyShiftCancelled(ctx, sessionId, shiftId, seats)

	logger.LogInfo("Response returned from CancelShift", sessionId)

	return
}

// GetShiftTemplates hands back the saved shapes of past shifts. The app prefills the
// create form with one of these and posts it back through CreateShift, which is how
// the ride templates already work, so there is no separate build from template api.
func GetShiftTemplates(ctx *gin.Context, sessionId, driverId, groupId string) (templates []postgress.ShiftTemplate, err error) {
	logger.LogInfo("Request received in GetShiftTemplates", sessionId)

	isMember, err := database.IsApprovedGroupMember(ctx, groupId, driverId)
	if err != nil {
		logger.LogError(sessionId, "failed to check the membership error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}
	if !isMember {
		err = errors.New(constants.Not_Group_Member)
		logger.LogError(sessionId, err)
		return
	}

	templates, err = getShiftTemplates(ctx, groupId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the templates error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the templates")
		return
	}

	logger.LogInfo("Response returned from GetShiftTemplates", sessionId)

	return
}

func DeleteShiftTemplate(ctx *gin.Context, sessionId, driverId, templateId string) (err error) {
	logger.LogInfo("Request received in DeleteShiftTemplate", sessionId)

	template, err := getShiftTemplateById(ctx, templateId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the template error: "+err.Error())
		err = fmt.Errorf(constants.Not_Found, "shift template")
		return
	}

	if err = requirePermission(ctx, sessionId, template.GroupID, driverId, constants.PERMISSION_SHIFT_MANAGE_TEMPLATES); err != nil {
		return
	}

	if err = deleteShiftTemplate(ctx, sessionId, templateId); err != nil {
		logger.LogError(sessionId, "failed to delete the template error: "+err.Error())
		err = fmt.Errorf(constants.DELETE_Failed, "shift template")
		return
	}

	logger.LogInfo("Response returned from DeleteShiftTemplate", sessionId)

	return
}

// RescheduleShift moves a shift that is already built. Without it, pushing a
// departure back fifteen minutes would mean cancelling and rebuilding, which throws
// away the seat plan and is refused outright inside the two hour window.
//
// The vehicle and the driver stay put on purpose: swapping either can invalidate
// every seat, so that remains a cancel and rebuild. Moving the window re-runs the
// full clash check, ignoring this shift's own row, and everybody aboard is told.
func RescheduleShift(ctx *gin.Context, sessionId, driverId string, request RescheduleShiftRequest) (err error) {
	logger.LogInfo("Request received in RescheduleShift", sessionId)

	shift, err := getShiftById(ctx, request.ShiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to get shift error: "+err.Error())
		err = errors.New(constants.Shift_Not_Found)
		return
	}

	if err = requirePermission(ctx, sessionId, shift.GroupID, driverId, constants.PERMISSION_SHIFT_CREATE); err != nil {
		return
	}

	updates := map[string]interface{}{}

	if !utils.IsStringEmpty(request.StartDatetime) {
		// the shift's own row is excluded, otherwise it would always clash with itself
		if err = checkClashes(ctx, sessionId, shift.VehicleID, shift.DriverID,
			request.StartDatetime, request.EstimatedEndDatetime, shift.ID); err != nil {
			return
		}

		updates["start_datetime"] = request.StartDatetime
		updates["estimated_end_datetime"] = request.EstimatedEndDatetime
	}

	if !utils.IsStringEmpty(request.StartLocation) {
		updates["start_location"] = strings.TrimSpace(request.StartLocation)
	}

	if !utils.IsStringEmpty(request.EndLocation) {
		updates["end_location"] = strings.TrimSpace(request.EndLocation)
	}

	if !utils.IsStringEmpty(request.RouteDetails) {
		updates["route_details"] = request.RouteDetails
	}

	if len(updates) == 0 && len(request.StopTimes) == 0 {
		logger.LogInfo("Response returned from RescheduleShift", sessionId)
		return
	}

	if err = applyShiftReschedule(ctx, sessionId, shift.ID, updates, request.StopTimes); err != nil {
		logger.LogError(sessionId, "failed to reschedule the shift error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "shift")
		return
	}

	notifyShift(ctx, sessionId, shift.ID, constants.NOTIFICATION_TYPE_SHIFT_UPDATED, constants.NOTIFICATION_TITLE_SHIFT_UPDATED)

	logger.LogInfo("Response returned from RescheduleShift", sessionId)

	return
}
