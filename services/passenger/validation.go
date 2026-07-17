package passenger

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"time"
)

func ValidateBookSeat(sessionId string, request BookSeatRequest) error {
	if !utils.PKValidation(request.RideId) {
		return fmt.Errorf(constants.Invalid_Data, "ride id")
	}

	nameLen := len(request.Name)
	if !(nameLen >= constants.Name_Min_Len && nameLen <= constants.Name_Max_Len) {
		return fmt.Errorf("length of the name should be between %v and %v characters", constants.Name_Min_Len, constants.Name_Max_Len)
	}

	if err := utils.IsValidMobileNumber(request.MobileNumber); err != nil {
		return err
	}

	if request.Seats == 0 {
		return fmt.Errorf(constants.Missing_Data, "seats")
	}

	var errMessage string
	if utils.IsStringEmptyWithKey(request.Code, "Code", &errMessage) {
		logger.LogError(sessionId, errMessage)
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	return nil
}

func ValidateRideRequest(sessionId string, request RideRequest) error {
	logger.LogInfo("Request received in ValidateRideRequest", sessionId)

	var errMessage string
	if utils.IsStringEmptyWithKey(request.StartDatetime, "StartDatetime", &errMessage) ||
		utils.IsStringEmptyWithKey(request.EstimatedEndDatetime, "EstimatedEndDatetime", &errMessage) ||
		utils.IsStringEmptyWithKey(request.StartLocation, "StartLocation", &errMessage) ||
		utils.IsStringEmptyWithKey(request.EndLocation, "EndLocation", &errMessage) ||
		utils.IsStringEmptyWithKey(request.RouteDetails, "RouteDetails", &errMessage) {
		logger.LogError(sessionId, errMessage)
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	// Length validations
	startDateLen := len(request.StartDatetime)
	estimatedEndDatetimeLen := len(request.EstimatedEndDatetime)
	startLocationLen := len(request.StartLocation)
	endLocationLen := len(request.EndLocation)
	routeDetailsLen := len(request.RouteDetails)

	if request.NumberOfSeats < constants.Number_Of_Seats_Min_Value || request.NumberOfSeats > constants.Number_Of_Seats_Max_Value {
		return fmt.Errorf("value of the number of seats should be between %v and %v", constants.Number_Of_Seats_Min_Value, constants.Number_Of_Seats_Max_Value)
	}
	if startDateLen < constants.General_Min_Len || startDateLen > constants.General_Max_Len {
		return fmt.Errorf("length of the start date should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
	}
	if estimatedEndDatetimeLen < constants.General_Min_Len || estimatedEndDatetimeLen > constants.General_Max_Len {
		return fmt.Errorf("length of the estimated end date should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
	}
	if startLocationLen < constants.General_Min_Len || startLocationLen > constants.General_Max_Len {
		return fmt.Errorf("length of the start location should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
	}
	if endLocationLen < constants.General_Min_Len || endLocationLen > constants.General_Max_Len {
		return fmt.Errorf("length of the end location should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
	}
	if routeDetailsLen < constants.RouteDetails_Min_Len || routeDetailsLen > constants.RouteDetails_Max_Len {
		return fmt.Errorf("length of the route details should be between %v and %v characters", constants.RouteDetails_Min_Len, constants.RouteDetails_Max_Len)
	}

	// Parse with local timezone (important for consistency)
	startDate, err := time.ParseInLocation(constants.DateTimeLayout, request.StartDatetime, time.Local)
	if err != nil {
		return fmt.Errorf("invalid start date")
	}

	endDate, err := time.ParseInLocation(constants.DateTimeLayout, request.EstimatedEndDatetime, time.Local)
	if err != nil {
		return fmt.Errorf("invalid estimated end date")
	}

	now := time.Now()

	now = now.Truncate(time.Second)
	startDate = startDate.Truncate(time.Second)
	endDate = endDate.Truncate(time.Second)
	maxAllowedStartDate := now.AddDate(0, 0, constants.Max_Start_Date_Gap)
	maxAllowedEndDate := startDate.AddDate(0, 0, constants.Max_End_Date_Gap)

	if startDate.Before(now) {
		return fmt.Errorf("start date cannot be in the past")
	}
	if endDate.Before(now) {
		return fmt.Errorf("estimated end date cannot be in the past")
	}
	if endDate.Before(startDate) {
		return fmt.Errorf("estimated end date cannot be before start date")
	}
	if startDate.After(maxAllowedStartDate) {
		return fmt.Errorf("startDate cannot be more than %d days from now", constants.Max_Start_Date_Gap)
	}
	if endDate.After(maxAllowedEndDate) {
		return fmt.Errorf("endDate cannot be more than %d hours from start date", constants.Max_End_Date_Gap*24)
	}

	return nil
}

func ValidateGetRideRequest(sessionId string, request GetRideRequest) error {
	logger.LogInfo("Request received in ValidateGetRideRequest", sessionId)

	if !utils.IsStringEmpty(request.StartDatetime) {
		if len(request.StartDatetime) < constants.General_Min_Len || len(request.StartDatetime) > constants.General_Max_Len {
			return fmt.Errorf("length of the start time should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
		}
	}
	if !utils.IsStringEmpty(request.EstimatedEndDatetime) {
		if len(request.EstimatedEndDatetime) < constants.General_Min_Len || len(request.EstimatedEndDatetime) > constants.General_Max_Len {
			return fmt.Errorf("length of the end time should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
		}
	}
	if !utils.IsStringEmpty(request.StartLocation) {
		if len(request.StartLocation) < constants.General_Min_Len || len(request.StartLocation) > constants.General_Max_Len {
			return fmt.Errorf("length of the start location should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
		}
	}
	if !utils.IsStringEmpty(request.EndLocation) {
		if len(request.EndLocation) < constants.General_Min_Len || len(request.EndLocation) > constants.General_Max_Len {
			return fmt.Errorf("length of the end location should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
		}
	}

	return nil
}
