package ride

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"
	"time"
)

func ValidateRideCreation(sessionId, driverId string, request *RideCreationRequest) error {
	logger.LogInfo("Request received in ValidateRideCreation", sessionId)
	logger.LogDebug("Request received in ValidateRideCreation", sessionId, "driver id:"+driverId)

	if !utils.PKValidation(driverId) {
		return fmt.Errorf(constants.Invalid_Data, "driver id")
	}

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
	routePointsLen := len(request.RoutePoints)

	if request.NumberOfSeats < constants.Number_Of_Seats_Min_Value || request.NumberOfSeats > constants.Number_Of_Seats_Max_Value {
		return fmt.Errorf("value of the number of seats should be between %v and %v", constants.Number_Of_Seats_Min_Value, constants.Number_Of_Seats_Max_Value)
	}
	if request.Fare < constants.Fare_Min_Len || request.Fare > constants.Fare_Max_Len {
		return fmt.Errorf("value of the fare should be between %v and %v", constants.Fare_Min_Len, constants.Fare_Max_Len)
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
	if routePointsLen != 0 && routePointsLen > constants.RoutePoints_Max_Len {
		return fmt.Errorf("route points should be not be more then %v", constants.RoutePoints_Max_Len)
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

	if request.IsRecurring {
		if request.Frequency == 0 {
			return fmt.Errorf(constants.Missing_Data, "frequency")
		}
		if request.Period == 0 {
			return fmt.Errorf(constants.Missing_Data, "period")
		}
		lenDOW := len(request.DaysOfWeek)
		if request.Period == WEEKLY && lenDOW == 0 {
			return fmt.Errorf(constants.Missing_Data, "days of week")
		}
		// we do not expect days of week for daily and monthly rides
		if request.Period != WEEKLY && lenDOW > 0 {
			return fmt.Errorf(constants.Invalid_Data, "days of week")
		}
		if lenDOW > 7 {
			return errors.New("There cannot be more than 7 days")
		}
		if request.Period == DAILY {
			if request.Frequency > constants.Max_Daily_Frequency {
				return fmt.Errorf("Daily frequency cannot be more than %d", constants.Max_Daily_Frequency)
			}
		} else if request.Period == WEEKLY {
			if request.Frequency > constants.Max_Weekly_Frequency {
				return fmt.Errorf("Weekly frequency cannot be more than %d", constants.Max_Weekly_Frequency)
			}
		} else if request.Period == MONTHLY {
			if request.Frequency > constants.Max_Monthly_Frequency {
				return fmt.Errorf("Monthly frequency cannot be more than %d in a week", constants.Max_Monthly_Frequency)
			}
		} else {
			return fmt.Errorf(constants.Invalid_Data, "period")
		}
		for _, day := range request.DaysOfWeek {
			if day != MONDAY &&
				day != TUESDAY &&
				day != WEDNESDAY &&
				day != THRUSDAY &&
				day != FRIDAY &&
				day != SATURDAY &&
				day != SUNDAY {
				return errors.New("Week days are invalid")
			}
		}
	}

	return nil
}

func ValidateFilteredRides(page int, startTime, endTime, searchLoc string) error {
	if page != 0 {
		if page < constants.Page_Min_Value || page > constants.Page_Max_Value {
			return fmt.Errorf("value of the page number should be between %v and %v", constants.Page_Min_Value, constants.Page_Max_Value)
		}
	}
	if !utils.IsStringEmpty(startTime) {
		if len(startTime) < constants.General_Min_Len || len(startTime) > constants.General_Max_Len {
			return fmt.Errorf("length of the start time should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
		}
	}
	if !utils.IsStringEmpty(endTime) {
		if len(endTime) < constants.General_Min_Len || len(endTime) > constants.General_Max_Len {
			return fmt.Errorf("length of the end time should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
		}
	}
	if !utils.IsStringEmpty(searchLoc) {
		if len(searchLoc) < constants.General_Min_Len || len(searchLoc) > constants.General_Max_Len {
			return fmt.Errorf("length of the start location should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
		}
	}

	return nil
}

func ValidateDriverRide(driverId, status string) error {
	if !utils.PKValidation(driverId) {
		return fmt.Errorf(constants.Invalid_Data, "driver id")
	}

	if !(strings.EqualFold(status, constants.Ride_Status_Active) ||
		strings.EqualFold(status, constants.Ride_Status_InActive) ||
		strings.EqualFold(status, constants.Ride_Status_All)) {
		return fmt.Errorf(constants.Invalid_Data, "ride status")
	}

	return nil
}

func ValidateUpdateRide(rideId string, request UpdateRideRequest) error {
	if !utils.PKValidation(rideId) {
		return fmt.Errorf(constants.Invalid_Data, "ride id")
	}

	if request.Status != nil && !utils.IsStringEmpty(*request.Status) && !(strings.EqualFold(*request.Status, constants.Ride_Status_Active) ||
		strings.EqualFold(*request.Status, constants.Ride_Status_InActive)) {
		return fmt.Errorf(constants.Invalid_Data, "ride status")
	}

	return nil
}

func ValidateBookRide(sessionId string, request *BookSeatRequest) error {
	if !utils.PKValidation(request.RideId) {
		return fmt.Errorf(constants.Invalid_Data, "ride id")
	}

	nameLen := len(request.Name)

	if nameLen < constants.Name_Min_Len || nameLen > constants.Name_Max_Len {
		return fmt.Errorf("value of the name should be between %v and %v", constants.Name_Min_Len, constants.Name_Max_Len)
	}

	return nil
}
