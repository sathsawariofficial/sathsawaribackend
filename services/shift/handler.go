package shift

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

// the owner builds a recurring shift for their service
func CreateShiftHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in CreateShiftHandler", sessionId)

	ownerId := ctx.GetString(constants.User_KEY)

	var request CreateShiftRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Creation_Failed, "shift"),
		})
		return
	}

	logger.LogDebug2("Request received in CreateShiftHandler", sessionId, request)

	if err := ValidateCreateShift(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := CreateShift(ctx, sessionId, ownerId, request)
	if err != nil {
		logger.LogError(sessionId, "create shift error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from CreateShiftHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(fmt.Sprintf(constants.Created_Successfully, "Shift"), resp))
}

// the owner changes the driver, vehicle, schedule or route of a shift
func UpdateShiftHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateShiftHandler", sessionId)

	ownerId := ctx.GetString(constants.User_KEY)

	var request UpdateShiftRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "shift"),
		})
		return
	}

	logger.LogDebug2("Request received in UpdateShiftHandler", sessionId, request)

	if err := ValidateUpdateShift(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := UpdateShift(ctx, sessionId, ownerId, request)
	if err != nil {
		logger.LogError(sessionId, "update shift error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from UpdateShiftHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(fmt.Sprintf(constants.Updated_Successfully, "Shift"), resp))
}

// the owner deletes a shift, it is archived first
func DeleteShiftHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeleteShiftHandler", sessionId)

	ownerId := ctx.GetString(constants.User_KEY)
	shiftId := ctx.Query(constants.Shift_Key)

	if err := ValidateId(shiftId, "shift id"); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := DeleteShift(ctx, sessionId, ownerId, shiftId); err != nil {
		logger.LogError(sessionId, "delete shift error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DeleteShiftHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Shift deleted")))
}

// the owner adds, moves and removes passengers of a shift in one call
func UpdateShiftPassengersHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateShiftPassengersHandler", sessionId)

	ownerId := ctx.GetString(constants.User_KEY)

	var request UpdateShiftPassengersRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "shift passengers"),
		})
		return
	}

	logger.LogDebug2("Request received in UpdateShiftPassengersHandler", sessionId, request)

	if err := ValidateUpdateShiftPassengers(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := UpdateShiftPassengers(ctx, sessionId, ownerId, request)
	if err != nil {
		logger.LogError(sessionId, "update passengers error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from UpdateShiftPassengersHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(fmt.Sprintf(constants.Updated_Successfully, "Shift passengers"), resp))
}

// one shift as its owner or its driver sees it
func GetShiftHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetShiftHandler", sessionId)

	userId := ctx.GetString(constants.User_KEY)
	shiftId := ctx.Query(constants.Shift_Key)

	if err := ValidateId(shiftId, "shift id"); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := GetShift(ctx, sessionId, userId, shiftId)
	if err != nil {
		logger.LogError(sessionId, "get shift error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetShiftHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

func readShiftListFilter(ctx *gin.Context) (shiftListFilter, error) {
	return ParseShiftListFilter(
		ctx.DefaultQuery(constants.Status_Key, ""),
		ctx.DefaultQuery(constants.Day_Of_Week_Key, ""),
		ctx.DefaultQuery(constants.Search_Loc_Key, ""),
		ctx.DefaultQuery(constants.Start_Time_Key, ""),
		ctx.DefaultQuery(constants.End_Time_Key, ""),
	)
}

// the shifts of the owner's service
func GetServiceShiftsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetServiceShiftsHandler", sessionId)

	ownerId := ctx.GetString(constants.User_KEY)

	filter, err := readShiftListFilter(ctx)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetServiceShifts(ctx, sessionId, ownerId, filter, page)
	if err != nil {
		logger.LogError(sessionId, "get shifts error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetServiceShiftsHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// the shifts the calling driver is assigned to
func GetDriverShiftsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetDriverShiftsHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	filter, err := readShiftListFilter(ctx)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetDriverShifts(ctx, sessionId, driverId, filter, page)
	if err != nil {
		logger.LogError(sessionId, "get shifts error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetDriverShiftsHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// every shift the owner's service ever ran, deleted ones included
func GetShiftHistoryHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetShiftHistoryHandler", sessionId)

	ownerId := ctx.GetString(constants.User_KEY)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetShiftHistory(ctx, sessionId, ownerId, page)
	if err != nil {
		logger.LogError(sessionId, "get shift history error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetShiftHistoryHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// who travelled on the owner's shifts, optionally one passenger
func GetPassengerHistoryHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPassengerHistoryHandler", sessionId)

	ownerId := ctx.GetString(constants.User_KEY)
	passengerId := ctx.DefaultQuery(constants.Passenger_Key, "")

	if !utils.IsStringEmpty(passengerId) {
		if err := ValidateId(passengerId, "passenger id"); err != nil {
			logger.LogError(sessionId, "validation error: "+err.Error())
			ctx.JSON(http.StatusBadRequest, utils.APIResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetPassengerHistory(ctx, sessionId, ownerId, passengerId, page)
	if err != nil {
		logger.LogError(sessionId, "get passenger history error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetPassengerHistoryHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// the trips of the owner's service, or the trips the calling driver drove
func GetOccurrencesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetOccurrencesHandler", sessionId)

	userId := ctx.GetString(constants.User_KEY)
	shiftId := ctx.DefaultQuery(constants.Shift_Key, "")

	if !utils.IsStringEmpty(shiftId) {
		if err := ValidateId(shiftId, "shift id"); err != nil {
			logger.LogError(sessionId, "validation error: "+err.Error())
			ctx.JSON(http.StatusBadRequest, utils.APIResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetOccurrences(ctx, sessionId, userId, shiftId, page)
	if err != nil {
		logger.LogError(sessionId, "get trips error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetOccurrencesHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// who travels on one trip, for the owner and the driver
func GetAttendanceHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetAttendanceHandler", sessionId)

	userId := ctx.GetString(constants.User_KEY)
	shiftId := ctx.Query(constants.Shift_Key)
	date := ctx.Query(constants.Date_Key)

	if err := ValidateAttendanceQuery(shiftId, date); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := GetAttendance(ctx, sessionId, userId, shiftId, date)
	if err != nil {
		logger.LogError(sessionId, "get attendance error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetAttendanceHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// the driver of a shift tells today's passengers where they are
func SendDriverLocationHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SendDriverLocationHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request DriverLocationRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Failed_To_Do_Job, "send the update"),
		})
		return
	}

	if err := ValidateDriverLocation(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := SendDriverLocation(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "send location error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from SendDriverLocationHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(fmt.Sprintf(constants.Success_Info, "Update sent"), resp))
}

// the owner searches passengers' shift requests
func SearchShiftRequestsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SearchShiftRequestsHandler", sessionId)

	ownerId := ctx.GetString(constants.User_KEY)

	filter, err := ParseShiftRequestSearch(
		ctx.DefaultQuery(constants.Search_Loc_Key, ""),
		ctx.DefaultQuery(constants.Day_Of_Week_Key, ""),
		ctx.DefaultQuery(constants.Start_Time_Key, ""),
		ctx.DefaultQuery(constants.End_Time_Key, ""),
	)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := SearchShiftRequests(ctx, sessionId, ownerId, filter, page)
	if err != nil {
		logger.LogError(sessionId, "search shift requests error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from SearchShiftRequestsHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// a passenger puts out a recurring requirement for owners to find
func CreateShiftRequestHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in CreateShiftRequestHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	var request ShiftRequestInput
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Creation_Failed, "shift request"),
		})
		return
	}

	logger.LogDebug2("Request received in CreateShiftRequestHandler", sessionId, request)

	if err := ValidateShiftRequest(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	requestId, err := AddShiftRequest(ctx, sessionId, passengerId, request)
	if err != nil {
		logger.LogError(sessionId, "create shift request error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from CreateShiftRequestHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(fmt.Sprintf(constants.Created_Successfully, "Shift request"), ShiftRequestResponse{RequestId: requestId}))
}

// the calling passenger's own shift requests
func GetMyShiftRequestsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetMyShiftRequestsHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	filter, err := ParseShiftRequestSearch(
		ctx.DefaultQuery(constants.Search_Loc_Key, ""),
		ctx.DefaultQuery(constants.Day_Of_Week_Key, ""),
		ctx.DefaultQuery(constants.Start_Time_Key, ""),
		ctx.DefaultQuery(constants.End_Time_Key, ""),
	)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := GetMyShiftRequests(ctx, sessionId, passengerId, filter, page)
	if err != nil {
		logger.LogError(sessionId, "get shift requests error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetMyShiftRequestsHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// a passenger deletes one of their shift requests, it is archived first
func DeleteShiftRequestHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeleteShiftRequestHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)
	requestId := ctx.Query(constants.Request_Key)

	if err := ValidateId(requestId, "request id"); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := DeleteShiftRequest(ctx, sessionId, passengerId, requestId); err != nil {
		logger.LogError(sessionId, "delete shift request error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DeleteShiftRequestHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Shift request deleted")))
}

// the shifts the calling passenger is on
func GetPassengerShiftsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPassengerShiftsHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	filter, err := readShiftListFilter(ctx)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetPassengerShifts(ctx, sessionId, passengerId, filter, page)
	if err != nil {
		logger.LogError(sessionId, "get shifts error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetPassengerShiftsHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// one shift as a passenger on it sees it
func GetPassengerShiftHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPassengerShiftHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)
	shiftId := ctx.Query(constants.Shift_Key)

	if err := ValidateId(shiftId, "shift id"); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := GetPassengerShift(ctx, sessionId, passengerId, shiftId)
	if err != nil {
		logger.LogError(sessionId, "get shift error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetPassengerShiftHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// a passenger marks themselves absent, or present again, for one trip
func MarkAttendanceHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in MarkAttendanceHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	var request MarkAttendanceRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "attendance"),
		})
		return
	}

	if err := ValidateMarkAttendance(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := MarkAttendance(ctx, sessionId, passengerId, request)
	if err != nil {
		logger.LogError(sessionId, "mark attendance error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from MarkAttendanceHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(fmt.Sprintf(constants.Updated_Successfully, "Attendance"), resp))
}

// who travels on one trip, for a passenger on it
func GetPassengerAttendanceHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPassengerAttendanceHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)
	shiftId := ctx.Query(constants.Shift_Key)
	date := ctx.Query(constants.Date_Key)

	if err := ValidateAttendanceQuery(shiftId, date); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := GetPassengerAttendance(ctx, sessionId, passengerId, shiftId, date)
	if err != nil {
		logger.LogError(sessionId, "get attendance error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetPassengerAttendanceHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}

// the trips the calling passenger was on
func GetTravelHistoryHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetTravelHistoryHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetTravelHistory(ctx, sessionId, passengerId, page)
	if err != nil {
		logger.LogError(sessionId, "get travel history error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetTravelHistoryHandler", sessionId)

	ctx.JSON(http.StatusOK, successResp(constants.Success, resp))
}
