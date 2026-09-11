package shift

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ownerError(sessionId string, err error) error {
	if isNotFound(err) {
		err = errors.New(constants.Not_Service_Owner)
	} else {
		logger.LogError(sessionId, "failed to read the service error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
	}
	logger.LogError(sessionId, err)

	return err
}

func shiftReadError(sessionId string, err error) error {
	if isNotFound(err) {
		err = errors.New(constants.Shift_Not_Found)
	} else {
		logger.LogError(sessionId, "failed to read the shift error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shift")
	}
	logger.LogError(sessionId, err)

	return err
}

// CreateShift builds a recurring shift for the owner's service. The driver has to be
// the owner or an approved driver, the vehicle approved with a valid seat count, the
// passengers approved members who fit in the vehicle, and none of them may already be
// on an overlapping shift in any service or, for the driver and vehicle, on a ride.
func CreateShift(ctx *gin.Context, sessionId, ownerId string, request CreateShiftRequest) (resp CreateShiftResponse, err error) {
	logger.LogInfo("Request received in CreateShift", sessionId)

	result, err := createShift(ctx, sessionId, ownerId, request)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Creation_Failed, "shift"))
		return
	}

	notifyShiftCreated(ctx, sessionId, ownerId, result)

	resp = CreateShiftResponse{
		ShiftId: result.Shift.ID,
		Seats:   seatInfo(result.Shift.SeatCapacity, result.Shift.OccupiedSeats),
	}

	logger.LogInfo("Response returned from CreateShift", sessionId)
	logger.LogDebug2("Response returned from CreateShift", sessionId, resp)

	return
}

// UpdateShift edits a shift and returns it as it now stands.
func UpdateShift(ctx *gin.Context, sessionId, ownerId string, request UpdateShiftRequest) (resp ShiftDetailResponse, err error) {
	logger.LogInfo("Request received in UpdateShift", sessionId)

	before, after, err := updateShift(ctx, sessionId, ownerId, request)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Update_Failed, "shift"))
		return
	}

	notifyShiftUpdated(ctx, sessionId, ownerId, before, after)

	resp, err = GetShift(ctx, sessionId, ownerId, after.Shift.ID)

	logger.LogInfo("Response returned from UpdateShift", sessionId)

	return
}

func UpdateShiftPassengers(ctx *gin.Context, sessionId, ownerId string, request UpdateShiftPassengersRequest) (resp UpdateShiftPassengersResponse, err error) {
	logger.LogInfo("Request received in UpdateShiftPassengers", sessionId)

	change, err := updateShiftPassengers(ctx, sessionId, ownerId, request)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Update_Failed, "shift passengers"))
		return
	}

	notifyPassengersChanged(ctx, sessionId, ownerId, change)

	resp = passengerChangeResp(change)

	logger.LogInfo("Response returned from UpdateShiftPassengers", sessionId)

	return
}

// DeleteShift archives and deletes a shift and tells its driver and passengers.
func DeleteShift(ctx *gin.Context, sessionId, ownerId, shiftId string) (err error) {
	logger.LogInfo("Request received in DeleteShift", sessionId)

	result, err := deleteShift(ctx, sessionId, ownerId, shiftId)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.DELETE_Failed, "shift"))
		return
	}

	notifyShiftDeleted(ctx, sessionId, ownerId, result)

	logger.LogInfo("Response returned from DeleteShift", sessionId)

	return
}

// GetShift is a shift as its owner, its driver or the owner of its vehicle sees it,
// passengers' numbers included.
func GetShift(ctx *gin.Context, sessionId, userId, shiftId string) (resp ShiftDetailResponse, err error) {
	logger.LogInfo("Request received in GetShift", sessionId)

	details, locations, passengers, err := loadShiftView(ctx, shiftId)
	if err != nil {
		err = shiftReadError(sessionId, err)
		return
	}

	if details.OwnerDriverID != userId && details.DriverID != userId && details.VehicleOwnerID != userId {
		err = errors.New(constants.Operation_Not_Permitted)
		logger.LogError(sessionId, err)
		return
	}

	resp = shiftDetailResp(details, locations, passengers, "", database.BusinessNow())

	logger.LogInfo("Response returned from GetShift", sessionId)

	return
}

// GetPassengerShift is a shift as one of its passengers sees it.
func GetPassengerShift(ctx *gin.Context, sessionId, passengerId, shiftId string) (resp ShiftDetailResponse, err error) {
	logger.LogInfo("Request received in GetPassengerShift", sessionId)

	details, locations, passengers, err := loadShiftView(ctx, shiftId)
	if err != nil {
		err = shiftReadError(sessionId, err)
		return
	}

	onShift := false
	for _, passenger := range passengers {
		if passenger.PassengerID == passengerId {
			onShift = true
		}
	}

	if !onShift {
		err = errors.New(constants.Not_Shift_Passenger)
		logger.LogError(sessionId, err)
		return
	}

	resp = shiftDetailResp(details, locations, passengers, passengerId, database.BusinessNow())

	logger.LogInfo("Response returned from GetPassengerShift", sessionId)

	return
}

func GetServiceShifts(ctx *gin.Context, sessionId, ownerId string, filter shiftListFilter, page int) (resp ShiftsResponse, err error) {
	logger.LogInfo("Request received in GetServiceShifts", sessionId)

	service, err := database.GetActiveServiceByOwner(ctx, ownerId)
	if err != nil {
		err = ownerError(sessionId, err)
		return
	}

	rows, totalRows, err := listShifts(ctx, func(query *gorm.DB) *gorm.DB {
		return query.Where("shifts.service_id = ?", service.ID)
	}, filter, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the shifts error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shifts")
		return
	}

	resp = shiftsPage(rows, totalRows, database.BusinessNow())

	logger.LogInfo("Response returned from GetServiceShifts", sessionId)

	return
}

// GetDriverShifts lists the shifts a driver drives and the shifts their vehicles are
// on, in any service. A member who joined with vehicles only sees their vehicles' shifts.
func GetDriverShifts(ctx *gin.Context, sessionId, driverId string, filter shiftListFilter, page int) (resp ShiftsResponse, err error) {
	logger.LogInfo("Request received in GetDriverShifts", sessionId)

	rows, totalRows, err := listShifts(ctx, func(query *gorm.DB) *gorm.DB {
		return query.Where("(shifts.driver_id = ? OR vehicles.driver_id = ?)", driverId, driverId)
	}, filter, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the shifts error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shifts")
		return
	}

	resp = shiftsPage(rows, totalRows, database.BusinessNow())

	logger.LogInfo("Response returned from GetDriverShifts", sessionId)

	return
}

func GetPassengerShifts(ctx *gin.Context, sessionId, passengerId string, filter shiftListFilter, page int) (resp ShiftsResponse, err error) {
	logger.LogInfo("Request received in GetPassengerShifts", sessionId)

	rows, totalRows, err := listShifts(ctx, func(query *gorm.DB) *gorm.DB {
		return query.Where(`EXISTS (
			SELECT 1 FROM shift_passengers
			WHERE shift_passengers.shift_id = shifts.id
			  AND shift_passengers.passenger_id = ?
			  AND shift_passengers.status = ?
		)`, passengerId, constants.Shift_Passenger_Active)
	}, filter, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the shifts error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shifts")
		return
	}

	resp = shiftsPage(rows, totalRows, database.BusinessNow())

	logger.LogInfo("Response returned from GetPassengerShifts", sessionId)

	return
}

// GetShiftHistory lists every shift the owner's service ever ran, deleted ones included.
func GetShiftHistory(ctx *gin.Context, sessionId, ownerId string, page int) (resp ShiftHistoryResponse, err error) {
	logger.LogInfo("Request received in GetShiftHistory", sessionId)

	service, err := database.GetActiveServiceByOwner(ctx, ownerId)
	if err != nil {
		err = ownerError(sessionId, err)
		return
	}

	rows, totalRows, err := getShiftHistory(ctx, service.ID, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the shift history error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shift history")
		return
	}

	resp = shiftHistoryPage(rows, totalRows)

	logger.LogInfo("Response returned from GetShiftHistory", sessionId)

	return
}

// GetPassengerHistory lists who travelled on the owner's shifts, when they joined the
// service, when they were put on and taken off each shift and how often they travelled.
func GetPassengerHistory(ctx *gin.Context, sessionId, ownerId, passengerId string, page int) (resp PassengerHistoryResponse, err error) {
	logger.LogInfo("Request received in GetPassengerHistory", sessionId)

	service, err := database.GetActiveServiceByOwner(ctx, ownerId)
	if err != nil {
		err = ownerError(sessionId, err)
		return
	}

	rows, totalRows, err := getPassengerHistory(ctx, service.ID, passengerId, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the passenger history error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the passenger history")
		return
	}

	resp = passengerHistoryPage(rows, totalRows)

	logger.LogInfo("Response returned from GetPassengerHistory", sessionId)

	return
}

// GetOccurrences lists trips with their attendance counted. An owner sees every trip
// of their service, a driver the trips they drove.
func GetOccurrences(ctx *gin.Context, sessionId, userId, shiftId string, page int) (resp OccurrencesResponse, err error) {
	logger.LogInfo("Request received in GetOccurrences", sessionId)

	scopeColumn, scopeValue := "driver_id", userId

	service, e := database.GetActiveServiceByOwner(ctx, userId)
	if e == nil {
		scopeColumn, scopeValue = "service_id", service.ID
	} else if !isNotFound(e) {
		logger.LogError(sessionId, "failed to read the service error: "+e.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	rows, totalRows, err := getOccurrences(ctx, scopeColumn, scopeValue, shiftId, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the trips error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the trips")
		return
	}

	resp = occurrencesPage(rows, totalRows)

	logger.LogInfo("Response returned from GetOccurrences", sessionId)

	return
}

func GetAttendance(ctx *gin.Context, sessionId, userId, shiftId, date string) (resp AttendanceResponse, err error) {
	logger.LogInfo("Request received in GetAttendance", sessionId)

	if resp, err = attendanceView(ctx, userId, shiftId, date, false); err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "get the attendance"))
		return
	}

	logger.LogInfo("Response returned from GetAttendance", sessionId)

	return
}

// GetPassengerAttendance lets the passengers of a trip see who else travels on it and
// who will be absent.
func GetPassengerAttendance(ctx *gin.Context, sessionId, passengerId, shiftId, date string) (resp AttendanceResponse, err error) {
	logger.LogInfo("Request received in GetPassengerAttendance", sessionId)

	if resp, err = attendanceView(ctx, passengerId, shiftId, date, true); err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "get the attendance"))
		return
	}

	logger.LogInfo("Response returned from GetPassengerAttendance", sessionId)

	return
}

// MarkAttendance marks the passenger absent, or present again, for one trip and tells
// the driver of that trip and the owner.
func MarkAttendance(ctx *gin.Context, sessionId, passengerId string, request MarkAttendanceRequest) (resp MarkAttendanceResponse, err error) {
	logger.LogInfo("Request received in MarkAttendance", sessionId)

	change, err := markAttendance(ctx, sessionId, passengerId, request)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Update_Failed, "attendance"))
		return
	}

	notifyAttendance(ctx, sessionId, change)

	resp = MarkAttendanceResponse{
		ShiftId: request.ShiftId,
		Date:    request.Date,
		Status:  request.Status,
		Changed: change.Changed,
	}

	logger.LogInfo("Response returned from MarkAttendance", sessionId)

	return
}

// SendDriverLocation lets the driver of a shift tell today's passengers where they are.
func SendDriverLocation(ctx *gin.Context, sessionId, driverId string, request DriverLocationRequest) (resp DriverLocationResponse, err error) {
	logger.LogInfo("Request received in SendDriverLocation", sessionId)

	result, err := sendDriverUpdate(ctx, sessionId, driverId, request)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "send the update"))
		return
	}

	notifyDriverUpdate(ctx, sessionId, request, result)

	resp = DriverLocationResponse{
		UpdateId:   result.Update.ID,
		Date:       result.Occurrence.OccurrenceDate,
		Recipients: len(result.Recipients),
	}

	logger.LogInfo("Response returned from SendDriverLocation", sessionId)

	return
}

func AddShiftRequest(ctx *gin.Context, sessionId, passengerId string, request ShiftRequestInput) (requestId string, err error) {
	logger.LogInfo("Request received in AddShiftRequest", sessionId)

	if requestId, err = createShiftRequest(ctx, sessionId, passengerId, request); err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Creation_Failed, "shift request"))
		return
	}

	logger.LogInfo("Response returned from AddShiftRequest", sessionId)
	logger.LogDebug2("Response returned from AddShiftRequest", sessionId, requestId)

	return
}

func GetMyShiftRequests(ctx *gin.Context, sessionId, passengerId string, filter shiftRequestSearch, page int) (resp ShiftRequestsResponse, err error) {
	logger.LogInfo("Request received in GetMyShiftRequests", sessionId)

	rows, locations, totalRows, err := getShiftRequests(ctx, func(query *gorm.DB) *gorm.DB {
		return query.Where("shift_requests.passenger_id = ?", passengerId)
	}, filter, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the shift requests error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shift requests")
		return
	}

	resp = shiftRequestsPage(rows, locations, totalRows)

	logger.LogInfo("Response returned from GetMyShiftRequests", sessionId)

	return
}

func DeleteShiftRequest(ctx *gin.Context, sessionId, passengerId, requestId string) (err error) {
	logger.LogInfo("Request received in DeleteShiftRequest", sessionId)

	if err = deleteShiftRequest(ctx, sessionId, passengerId, requestId); err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.DELETE_Failed, "shift request"))
		return
	}

	logger.LogInfo("Response returned from DeleteShiftRequest", sessionId)

	return
}

// SearchShiftRequests lets an owner find passengers' requirements. A request addressed
// to another service is never shown to them.
func SearchShiftRequests(ctx *gin.Context, sessionId, ownerId string, filter shiftRequestSearch, page int) (resp ShiftRequestsResponse, err error) {
	logger.LogInfo("Request received in SearchShiftRequests", sessionId)

	service, err := database.GetActiveServiceByOwner(ctx, ownerId)
	if err != nil {
		err = ownerError(sessionId, err)
		return
	}

	rows, locations, totalRows, err := getShiftRequests(ctx, func(query *gorm.DB) *gorm.DB {
		return query.Where("(COALESCE(shift_requests.service_id, '') = '' OR shift_requests.service_id = ?)", service.ID)
	}, filter, page)
	if err != nil {
		logger.LogError(sessionId, "failed to search the shift requests error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "search the shift requests")
		return
	}

	resp = shiftRequestsPage(rows, locations, totalRows)

	logger.LogInfo("Response returned from SearchShiftRequests", sessionId)

	return
}

// GetTravelHistory lists the trips a passenger was on, with the driver, vehicle and
// route of that day and whether they travelled.
func GetTravelHistory(ctx *gin.Context, sessionId, passengerId string, page int) (resp TravelHistoryResponse, err error) {
	logger.LogInfo("Request received in GetTravelHistory", sessionId)

	rows, totalRows, err := getTravelHistory(ctx, passengerId, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the travel history error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the travel history")
		return
	}

	resp = travelHistoryPage(rows, totalRows)

	logger.LogInfo("Response returned from GetTravelHistory", sessionId)

	return
}
