package shift

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/utils"
	"strings"
)

func ValidateId(id, key string) error {
	if !utils.PKValidation(id) {
		return fmt.Errorf(constants.Invalid_Data, key)
	}

	return nil
}

func validateDate(value *string, key string) error {
	*value = strings.TrimSpace(*value)

	if utils.IsStringEmpty(*value) {
		return fmt.Errorf(constants.Missing_Data, key)
	}

	_, err := utils.ParseBusinessDate(*value, strings.ToLower(key))
	return err
}

// validateEndDate allows no end date at all, a shift may run until it is changed.
func validateEndDate(value *string, startDate string) error {
	*value = strings.TrimSpace(*value)

	if utils.IsStringEmpty(*value) {
		return nil
	}

	if _, err := utils.ParseBusinessDate(*value, "end date"); err != nil {
		return err
	}

	if *value < database.BusinessToday() {
		return errors.New(constants.End_Date_In_The_Past)
	}

	if startDate != "" && *value < startDate {
		return fmt.Errorf(constants.Invalid_Data, "end date, it cannot be before the start date")
	}

	return nil
}

// validatePassengerInputs checks passengers placed on stops. stops is the length of
// the route when it is known, otherwise only the lower bound can be checked here.
func validatePassengerInputs(passengers []ShiftPassengerInput, stops int, seen map[string]bool) error {
	for _, passenger := range passengers {
		if err := ValidateId(passenger.PassengerId, "passenger id"); err != nil {
			return err
		}

		if seen[passenger.PassengerId] {
			return fmt.Errorf(constants.Invalid_Data, "passengers, a passenger is listed twice")
		}
		seen[passenger.PassengerId] = true

		if passenger.LocationSequence < 1 || (stops > 0 && passenger.LocationSequence > stops) {
			return fmt.Errorf(constants.Invalid_Data, fmt.Sprintf("location sequence of passenger %s", passenger.PassengerId))
		}
	}

	return nil
}

func ValidateCreateShift(request *CreateShiftRequest) (err error) {
	if err = ValidateId(request.DriverId, "driver id"); err != nil {
		return
	}

	if err = ValidateId(request.VehicleId, "vehicle id"); err != nil {
		return
	}

	if err = utils.ValidateDaysOfWeek(request.DaysOfWeek); err != nil {
		return
	}

	if err = validateDate(&request.StartDate, "Start date"); err != nil {
		return
	}

	if request.StartDate < database.BusinessToday() {
		return errors.New(constants.Shift_In_The_Past)
	}

	if err = validateEndDate(&request.EndDate, request.StartDate); err != nil {
		return
	}

	if request.Locations, err = utils.ValidateRouteLocations(request.Locations, constants.Shift_Locations_Min_Len, constants.Shift_Locations_Max_Len); err != nil {
		return
	}

	if len(request.Passengers) > constants.Number_Of_Seats_Max_Value {
		return fmt.Errorf("no more than %v passengers can be sent in one call", constants.Number_Of_Seats_Max_Value)
	}

	return validatePassengerInputs(request.Passengers, len(request.Locations), map[string]bool{})
}

// ValidateUpdateShift checks the shape of an edit. Whether a new start date is in
// the past is judged by the service, which knows whether it changed at all.
func ValidateUpdateShift(request *UpdateShiftRequest) (err error) {
	if err = ValidateId(request.ShiftId, "shift id"); err != nil {
		return
	}

	if request.DriverId == nil && request.VehicleId == nil && request.DaysOfWeek == nil &&
		request.StartDate == nil && request.EndDate == nil && request.Locations == nil {
		return fmt.Errorf(constants.Missing_Data, "At least one change")
	}

	if request.DriverId != nil {
		if err = ValidateId(*request.DriverId, "driver id"); err != nil {
			return
		}
	}

	if request.VehicleId != nil {
		if err = ValidateId(*request.VehicleId, "vehicle id"); err != nil {
			return
		}
	}

	if request.DaysOfWeek != nil {
		if err = utils.ValidateDaysOfWeek(request.DaysOfWeek); err != nil {
			return
		}
	}

	startDate := ""
	if request.StartDate != nil {
		if err = validateDate(request.StartDate, "Start date"); err != nil {
			return
		}
		startDate = *request.StartDate
	}

	if request.EndDate != nil {
		if err = validateEndDate(request.EndDate, startDate); err != nil {
			return
		}
	}

	if request.Locations != nil {
		if request.Locations, err = utils.ValidateRouteLocations(request.Locations, constants.Shift_Locations_Min_Len, constants.Shift_Locations_Max_Len); err != nil {
			return
		}
	}

	return nil
}

func ValidateUpdateShiftPassengers(request *UpdateShiftPassengersRequest) error {
	if err := ValidateId(request.ShiftId, "shift id"); err != nil {
		return err
	}

	total := len(request.Add) + len(request.Move) + len(request.Remove)
	if total == 0 {
		return fmt.Errorf(constants.Missing_Data, "Passengers to add, move or remove")
	}

	if total > constants.Number_Of_Seats_Max_Value {
		return fmt.Errorf("no more than %v passengers can be changed in one call", constants.Number_Of_Seats_Max_Value)
	}

	seen := map[string]bool{}

	if err := validatePassengerInputs(request.Add, 0, seen); err != nil {
		return err
	}

	if err := validatePassengerInputs(request.Move, 0, seen); err != nil {
		return err
	}

	for _, passengerId := range request.Remove {
		if err := ValidateId(passengerId, "passenger id"); err != nil {
			return err
		}

		if seen[passengerId] {
			return fmt.Errorf(constants.Invalid_Data, "passengers, a passenger is listed twice")
		}
		seen[passengerId] = true
	}

	return nil
}

func ValidateAttendanceQuery(shiftId, date string) error {
	if err := ValidateId(shiftId, "shift id"); err != nil {
		return err
	}

	return validateDate(&date, "Date")
}

func ValidateMarkAttendance(request *MarkAttendanceRequest) error {
	if err := ValidateId(request.ShiftId, "shift id"); err != nil {
		return err
	}

	if err := validateDate(&request.Date, "Date"); err != nil {
		return err
	}

	if request.Status != constants.Attendance_Absent && request.Status != constants.Attendance_Present {
		return fmt.Errorf(constants.Invalid_Data, "status, use absent or present")
	}

	return nil
}

func ValidateDriverLocation(request *DriverLocationRequest) error {
	if err := ValidateId(request.ShiftId, "shift id"); err != nil {
		return err
	}

	request.Message = strings.TrimSpace(request.Message)
	request.Location = strings.TrimSpace(request.Location)

	if utils.IsStringEmpty(request.Message) && utils.IsStringEmpty(request.Location) {
		return fmt.Errorf(constants.Missing_Data, "Message or location")
	}

	if len(request.Message) > constants.Message_Max_Len {
		return fmt.Errorf("length of the message should not be more than %v characters", constants.Message_Max_Len)
	}

	if len(request.Location) > constants.Location_Max_Len {
		return fmt.Errorf("length of the location should not be more than %v characters", constants.Location_Max_Len)
	}

	if request.Lat < -90 || request.Lat > 90 || request.Lng < -180 || request.Lng > 180 {
		return fmt.Errorf(constants.Invalid_Data, "coordinates")
	}

	return nil
}

func ValidateShiftRequest(request *ShiftRequestInput) (err error) {
	request.ServiceId = strings.TrimSpace(request.ServiceId)
	request.Note = strings.TrimSpace(request.Note)

	if !utils.IsStringEmpty(request.ServiceId) {
		if err = ValidateId(request.ServiceId, "service id"); err != nil {
			return
		}
	}

	if err = utils.ValidateDaysOfWeek(request.DaysOfWeek); err != nil {
		return
	}

	if err = utils.IsValidMobileNumber(request.ContactNumber); err != nil {
		return
	}

	if len(request.Note) > constants.Description_Max_Len {
		return fmt.Errorf("length of the note should not be more than %v characters", constants.Description_Max_Len)
	}

	request.Locations, err = utils.ValidateRouteLocations(request.Locations, constants.Shift_Request_Locations_Min_Len, constants.Shift_Request_Locations_Max_Len)

	return
}

// ParseShiftListFilter reads the optional narrowing of a shift list, the active
// shifts are listed when no status is asked for.
func ParseShiftListFilter(status, dayOfWeek, search, startTime, endTime string) (filter shiftListFilter, err error) {
	if !utils.IsStringEmpty(startTime) {
		if filter.StartTime, err = utils.NormalizeClock(startTime); err != nil {
			return filter, fmt.Errorf(constants.Invalid_Data, "start time, use HH:MM")
		}
	}

	if !utils.IsStringEmpty(endTime) {
		if filter.EndTime, err = utils.NormalizeClock(endTime); err != nil {
			return filter, fmt.Errorf(constants.Invalid_Data, "end time, use HH:MM")
		}
	}

	switch status {
	case "":
		filter.Status = constants.Shift_Status_Active
	case constants.Shift_Status_Active, constants.Shift_Status_Completed, constants.Shift_Status_All:
		filter.Status = status
	default:
		return filter, fmt.Errorf(constants.Invalid_Data, "status, use active, completed or all")
	}

	if !utils.IsStringEmpty(dayOfWeek) {
		filter.DayOfWeek = utils.ToInt(dayOfWeek)
		if filter.DayOfWeek < constants.Day_Of_Week_Min_Value || filter.DayOfWeek > constants.Day_Of_Week_Max_Value {
			return filter, fmt.Errorf(constants.Invalid_Data, "day of week, use 1 for Monday to 7 for Sunday")
		}
	}

	filter.Search = strings.TrimSpace(search)
	if len(filter.Search) > constants.General_Max_Len {
		return filter, fmt.Errorf("length of the search should not be more than %v characters", constants.General_Max_Len)
	}

	return filter, nil
}

func ParseShiftRequestSearch(search, dayOfWeek, startTime, endTime string) (filter shiftRequestSearch, err error) {
	filter.Search = strings.TrimSpace(search)
	if len(filter.Search) > constants.General_Max_Len {
		return filter, fmt.Errorf("length of the search should not be more than %v characters", constants.General_Max_Len)
	}

	if !utils.IsStringEmpty(dayOfWeek) {
		filter.DayOfWeek = utils.ToInt(dayOfWeek)
		if filter.DayOfWeek < constants.Day_Of_Week_Min_Value || filter.DayOfWeek > constants.Day_Of_Week_Max_Value {
			return filter, fmt.Errorf(constants.Invalid_Data, "day of week, use 1 for Monday to 7 for Sunday")
		}
	}

	if !utils.IsStringEmpty(startTime) {
		if filter.StartTime, err = utils.NormalizeClock(startTime); err != nil {
			return filter, fmt.Errorf(constants.Invalid_Data, "start time, use HH:MM")
		}
	}

	if !utils.IsStringEmpty(endTime) {
		if filter.EndTime, err = utils.NormalizeClock(endTime); err != nil {
			return filter, fmt.Errorf(constants.Invalid_Data, "end time, use HH:MM")
		}
	}

	return filter, nil
}
