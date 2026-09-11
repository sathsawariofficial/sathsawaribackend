package pickdrop

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/utils"
	"strings"
)

func ValidateId(id, key string) error {
	if !utils.PKValidation(id) {
		return fmt.Errorf(constants.Invalid_Data, key)
	}

	return nil
}

func validateIdList(ids []string, key string) error {
	if len(ids) > constants.Bulk_Request_Max_Len {
		return fmt.Errorf("no more than %v entries can be sent in one call", constants.Bulk_Request_Max_Len)
	}

	seen := map[string]bool{}
	for _, id := range ids {
		if !utils.PKValidation(id) {
			return fmt.Errorf(constants.Invalid_Data, key)
		}

		if seen[id] {
			return fmt.Errorf(constants.Invalid_Data, key+", it is repeated")
		}
		seen[id] = true
	}

	return nil
}

func ValidateEnableService(request *EnableServiceRequest) error {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)

	var errMessage string
	if utils.IsStringEmptyWithKey(request.Name, "Name", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	nameLen := len(request.Name)
	if nameLen < constants.Service_Name_Min_Len || nameLen > constants.Service_Name_Max_Len {
		return fmt.Errorf("length of the name should be between %v and %v characters", constants.Service_Name_Min_Len, constants.Service_Name_Max_Len)
	}

	if len(request.Description) > constants.Description_Max_Len {
		return fmt.Errorf("length of the description should not be more than %v characters", constants.Description_Max_Len)
	}

	// adding a vehicle is optional, a service may start with none
	return validateIdList(request.VehicleIds, "vehicle id")
}

func ValidateVehicleIds(request *VehicleIdsRequest) error {
	if len(request.VehicleIds) == 0 {
		return fmt.Errorf(constants.Missing_Data, "Vehicle ids")
	}

	return validateIdList(request.VehicleIds, "vehicle id")
}

func ValidateDriverJoin(request *DriverJoinRequest) error {
	if err := ValidateId(request.ServiceId, "service id"); err != nil {
		return err
	}

	request.JoinType = strings.TrimSpace(request.JoinType)
	switch request.JoinType {
	case "":
		request.JoinType = constants.Join_Type_Driver
	case constants.Join_Type_Driver, constants.Join_Type_Vehicles:
	default:
		return fmt.Errorf(constants.Invalid_Data, "join type, use driver or vehicles")
	}

	// a driver may ask to join without offering any vehicle, but somebody who joins
	// only through their vehicles has to bring at least one
	if request.JoinType == constants.Join_Type_Vehicles && len(request.VehicleIds) == 0 {
		return errors.New(constants.Vehicles_Join_Needs_Vehicle)
	}

	return validateIdList(request.VehicleIds, "vehicle id")
}

func ValidatePassengerJoin(request *PassengerJoinRequest) error {
	return ValidateId(request.ServiceId, "service id")
}

func ValidateRequestType(requestType string) error {
	switch requestType {
	case constants.Request_Type_Driver, constants.Request_Type_Vehicle, constants.Request_Type_Passenger:
		return nil
	}

	return fmt.Errorf(constants.Invalid_Data, "type, use driver, vehicle or passenger")
}

// ValidateRequestStatus accepts any membership status, all, or nothing, which lists
// the requests still waiting.
func ValidateRequestStatus(status string) error {
	switch status {
	case "",
		constants.Membership_Status_All,
		constants.Membership_Status_Pending,
		constants.Membership_Status_Approved,
		constants.Membership_Status_Rejected,
		constants.Membership_Status_Withdrawn,
		constants.Membership_Status_Left,
		constants.Membership_Status_Removed,
		constants.Membership_Status_Closed:
		return nil
	}

	return fmt.Errorf(constants.Invalid_Data, "status")
}

func ValidateDecideRequests(request *DecideRequestsRequest) error {
	if len(request.Decisions) == 0 {
		return fmt.Errorf(constants.Missing_Data, "Decisions")
	}

	if len(request.Decisions) > constants.Bulk_Request_Max_Len {
		return fmt.Errorf("no more than %v decisions can be sent in one call", constants.Bulk_Request_Max_Len)
	}

	seen := map[string]bool{}
	for _, decision := range request.Decisions {
		if err := ValidateRequestType(decision.Type); err != nil {
			return err
		}

		if err := ValidateId(decision.RequestId, "request id"); err != nil {
			return err
		}

		switch decision.Action {
		case constants.Request_Action_Approve, constants.Request_Action_Reject, constants.Request_Action_Remove:
		default:
			return fmt.Errorf(constants.Invalid_Data, "action, use approve, reject or remove")
		}

		key := decision.Type + ":" + decision.RequestId
		if seen[key] {
			return fmt.Errorf(constants.Invalid_Data, "decisions, a request is listed twice")
		}
		seen[key] = true
	}

	return nil
}

// ParseAvailabilityFilter reads the optional schedule an owner can send to the
// available lists. It is all or nothing: a schedule without days or times cannot be
// checked for clashes.
func ParseAvailabilityFilter(days, startDate, endDate, startTime, endTime, excludeShiftId string) (filter AvailabilityFilter, err error) {
	if days == "" && startDate == "" && endDate == "" && startTime == "" && endTime == "" && excludeShiftId == "" {
		return
	}

	if utils.IsStringEmpty(days) || utils.IsStringEmpty(startDate) || utils.IsStringEmpty(startTime) || utils.IsStringEmpty(endTime) {
		return filter, fmt.Errorf(constants.Missing_Data, "days_of_week, start_date, start_time and end_time, they have to be sent together")
	}

	if filter.DaysOfWeek, err = utils.ParseDaysOfWeekQuery(days); err != nil {
		return
	}

	if _, err = utils.ParseBusinessDate(startDate, "start date"); err != nil {
		return
	}
	filter.StartDate = strings.TrimSpace(startDate)

	if !utils.IsStringEmpty(endDate) {
		if _, err = utils.ParseBusinessDate(endDate, "end date"); err != nil {
			return
		}

		filter.EndDate = strings.TrimSpace(endDate)
		if filter.EndDate < filter.StartDate {
			return filter, fmt.Errorf(constants.Invalid_Data, "end date, it cannot be before the start date")
		}
	}

	if filter.StartTime, err = utils.NormalizeClock(startTime); err != nil {
		return filter, fmt.Errorf(constants.Invalid_Data, "start time, use HH:MM")
	}

	if filter.EndTime, err = utils.NormalizeClock(endTime); err != nil {
		return filter, fmt.Errorf(constants.Invalid_Data, "end time, use HH:MM")
	}

	if filter.EndTime <= filter.StartTime {
		return filter, fmt.Errorf(constants.Invalid_Data, "end time, it has to be after the start time")
	}

	if !utils.IsStringEmpty(excludeShiftId) {
		if err = ValidateId(excludeShiftId, "exclude shift id"); err != nil {
			return
		}
		filter.ExcludeShiftId = excludeShiftId
	}

	return
}

// ValidateAdvertisement checks an advertisement and cleans it in place: trimmed text,
// times as HH:MM and every location inside the advertised hours.
func ValidateAdvertisement(request *AdvertisementRequest) (err error) {
	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)

	var errMessage string
	if utils.IsStringEmptyWithKey(request.Title, "Title", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	titleLen := len(request.Title)
	if titleLen < constants.Title_Min_Len || titleLen > constants.Title_Max_Len {
		return fmt.Errorf("length of the title should be between %v and %v characters", constants.Title_Min_Len, constants.Title_Max_Len)
	}

	if len(request.Description) > constants.Description_Max_Len {
		return fmt.Errorf("length of the description should not be more than %v characters", constants.Description_Max_Len)
	}

	if request.Fare < constants.Fare_Min_Len || request.Fare > constants.Fare_Max_Len {
		return fmt.Errorf("value of the fare should be between %v and %v", constants.Fare_Min_Len, constants.Fare_Max_Len)
	}

	if err = utils.ValidateDaysOfWeek(request.DaysOfWeek); err != nil {
		return
	}

	if request.StartTime, err = utils.NormalizeClock(request.StartTime); err != nil {
		return fmt.Errorf(constants.Invalid_Data, "start time, use HH:MM")
	}

	if request.EndTime, err = utils.NormalizeClock(request.EndTime); err != nil {
		return fmt.Errorf(constants.Invalid_Data, "end time, use HH:MM")
	}

	if request.EndTime <= request.StartTime {
		return fmt.Errorf(constants.Invalid_Data, "end time, it has to be after the start time")
	}

	if request.Locations, err = utils.ValidateRouteLocations(request.Locations, constants.Advertisement_Locations_Min_Len, constants.Advertisement_Locations_Max_Len); err != nil {
		return
	}

	first := request.Locations[0].Time
	last := request.Locations[len(request.Locations)-1].Time
	if first < request.StartTime || last > request.EndTime {
		return fmt.Errorf("every location has to be reached between the start time %s and the end time %s", request.StartTime, request.EndTime)
	}

	return nil
}

func ParseAdvertisementSearch(search, dayOfWeek, startTime, endTime string) (filter AdvertisementSearch, err error) {
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
