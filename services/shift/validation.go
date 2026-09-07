package shift

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"time"
)

func ValidateShiftId(shiftId string) error {
	if !utils.PKValidation(shiftId) {
		return fmt.Errorf(constants.Invalid_Data, "shift id")
	}

	return nil
}

func ValidateGroupId(groupId string) error {
	if !utils.PKValidation(groupId) {
		return fmt.Errorf(constants.Invalid_Data, "group id")
	}

	return nil
}

func ValidateTemplateId(templateId string) error {
	if !utils.PKValidation(templateId) {
		return fmt.Errorf(constants.Invalid_Data, "shift template id")
	}

	return nil
}

// ValidateCreateShift checks the whole payload before a single row is written: the
// route has to be in order, the seats have to be sane, and nobody may be seated
// twice on the same trip.
func ValidateCreateShift(sessionId string, request *CreateShiftRequest) error {
	logger.LogInfo("Request received in ValidateCreateShift", sessionId)

	if err := ValidateGroupId(request.GroupId); err != nil {
		return err
	}

	if !utils.PKValidation(request.VehicleId) {
		return fmt.Errorf(constants.Invalid_Data, "vehicle id")
	}

	if !utils.PKValidation(request.DriverId) {
		return fmt.Errorf(constants.Invalid_Data, "driver id")
	}

	if request.Direction != constants.Shift_Direction_Pickup && request.Direction != constants.Shift_Direction_Drop {
		return fmt.Errorf(constants.Invalid_Data, "direction")
	}

	var errMessage string
	if utils.IsStringEmptyWithKey(request.StartDatetime, "Start date", &errMessage) ||
		utils.IsStringEmptyWithKey(request.EstimatedEndDatetime, "Estimated end date", &errMessage) ||
		utils.IsStringEmptyWithKey(request.StartLocation, "Start location", &errMessage) ||
		utils.IsStringEmptyWithKey(request.EndLocation, "End location", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	if err := validateShiftWindow(request.StartDatetime, request.EstimatedEndDatetime); err != nil {
		return err
	}

	if len(request.RouteDetails) > constants.RouteDetails_Max_Len {
		return fmt.Errorf("length of the route details should not be more than %v characters", constants.RouteDetails_Max_Len)
	}

	if len(request.Stops) == 0 {
		return fmt.Errorf(constants.Missing_Data, "At least one stop")
	}

	if len(request.Stops) > constants.Shift_Stops_Max_Len {
		return fmt.Errorf("a shift cannot have more than %v stops", constants.Shift_Stops_Max_Len)
	}

	seenSeats := map[int]bool{}
	seenPassengers := map[string]bool{}

	for _, stop := range request.Stops {
		if utils.IsStringEmpty(stop.Location) {
			return fmt.Errorf(constants.Missing_Data, "Stop location")
		}

		locationLen := len(stop.Location)
		if locationLen < constants.Location_Min_Len || locationLen > constants.Location_Max_Len {
			return fmt.Errorf("length of a stop location should be between %v and %v characters", constants.Location_Min_Len, constants.Location_Max_Len)
		}

		for _, seat := range stop.Seats {
			if err := validateSeat(seat.SeatNumber, seat.Gender, seat.PassengerId, seenSeats, seenPassengers); err != nil {
				return err
			}
		}
	}

	for _, day := range request.DaysOfWeek {
		if day < constants.Day_Of_Week_Min_Value || day > constants.Day_Of_Week_Max_Value {
			return fmt.Errorf(constants.Invalid_Data, "day of week")
		}
	}

	logger.LogInfo("Response returned from ValidateCreateShift", sessionId)

	return nil
}

func ValidateUpdateShiftSeats(request *UpdateShiftSeatsRequest) error {
	if err := ValidateShiftId(request.ShiftId); err != nil {
		return err
	}

	if len(request.Seats) == 0 {
		return fmt.Errorf(constants.Missing_Data, "Seats")
	}

	if len(request.Seats) > constants.Number_Of_Seats_Max_Value {
		return fmt.Errorf("no more than %v seats can be changed in one call", constants.Number_Of_Seats_Max_Value)
	}

	seenSeats := map[int]bool{}
	seenPassengers := map[string]bool{}

	for _, seat := range request.Seats {
		if err := validateSeat(seat.SeatNumber, seat.Gender, seat.PassengerId, seenSeats, seenPassengers); err != nil {
			return err
		}

		if seat.StopSequence < 0 {
			return fmt.Errorf(constants.Invalid_Data, "stop sequence")
		}

		// somebody has to be waiting somewhere, a seated passenger without a stop
		// would never be picked up
		if !utils.IsStringEmpty(seat.PassengerId) && seat.StopSequence == 0 {
			return fmt.Errorf(constants.Missing_Data, "Stop sequence for the seated passenger")
		}
	}

	return nil
}

// validateSeat applies the rules one seat has to satisfy no matter which api it
// arrives through: a real seat number, a valid gender, and a passenger who is not
// already sitting somewhere else on the same trip.
func validateSeat(seatNumber int, gender, passengerId string, seenSeats map[int]bool, seenPassengers map[string]bool) error {
	if seatNumber < constants.Number_Of_Seats_Min_Value || seatNumber > constants.Number_Of_Seats_Max_Value {
		return fmt.Errorf(constants.Invalid_Data, "seat number")
	}

	if seenSeats[seatNumber] {
		return fmt.Errorf("seat %d is listed more than once", seatNumber)
	}
	seenSeats[seatNumber] = true

	seatGender := utils.Gender(gender)
	if !utils.IsStringEmpty(gender) && !seatGender.IsValid() {
		return fmt.Errorf(constants.Invalid_Data, "seat gender")
	}

	if utils.IsStringEmpty(passengerId) {
		return nil
	}

	if !utils.PKValidation(passengerId) {
		return fmt.Errorf(constants.Invalid_Data, "passenger id")
	}

	// a seat holding somebody must say which gender it is kept for, that is what
	// stops the seat being handed to the other gender later
	if utils.IsStringEmpty(gender) {
		return fmt.Errorf(constants.Missing_Data, "Seat gender")
	}

	if seenPassengers[passengerId] {
		return fmt.Errorf("a passenger cannot be seated twice on the same shift")
	}
	seenPassengers[passengerId] = true

	return nil
}

func validateShiftWindow(startDatetime, estimatedEndDatetime string) error {
	start, err := time.ParseInLocation(constants.DateTimeLayout, startDatetime, time.Local)
	if err != nil {
		return fmt.Errorf(constants.Invalid_Data, "start date")
	}

	end, err := time.ParseInLocation(constants.DateTimeLayout, estimatedEndDatetime, time.Local)
	if err != nil {
		return fmt.Errorf(constants.Invalid_Data, "estimated end date")
	}

	if !end.After(start) {
		return fmt.Errorf(constants.Invalid_Data, "estimated end date, it has to be after the start")
	}

	return nil
}

// ValidateRescheduleShift checks a reschedule. Every field is optional except the
// shift itself, but a time window has to be sent whole or not at all, half a window
// cannot be judged against the clash rules.
func ValidateRescheduleShift(request *RescheduleShiftRequest) error {
	if err := ValidateShiftId(request.ShiftId); err != nil {
		return err
	}

	startGiven := !utils.IsStringEmpty(request.StartDatetime)
	endGiven := !utils.IsStringEmpty(request.EstimatedEndDatetime)

	if startGiven != endGiven {
		return fmt.Errorf(constants.Missing_Data, "Both the start and the estimated end date")
	}

	if startGiven {
		if err := validateShiftWindow(request.StartDatetime, request.EstimatedEndDatetime); err != nil {
			return err
		}
	}

	if len(request.RouteDetails) > constants.RouteDetails_Max_Len {
		return fmt.Errorf("length of the route details should not be more than %v characters", constants.RouteDetails_Max_Len)
	}

	if len(request.StopTimes) > constants.Shift_Stops_Max_Len {
		return fmt.Errorf("a shift cannot have more than %v stops", constants.Shift_Stops_Max_Len)
	}

	for _, stop := range request.StopTimes {
		if stop.SequenceNumber < 1 {
			return fmt.Errorf(constants.Invalid_Data, "stop sequence")
		}

		if utils.IsStringEmpty(stop.ScheduledTime) {
			return fmt.Errorf(constants.Missing_Data, "Scheduled time")
		}
	}

	return nil
}
