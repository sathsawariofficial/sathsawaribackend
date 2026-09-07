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

// builds one directional trip of one vehicle on one day
func CreateShiftHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in CreateShiftHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

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

	if err := ValidateCreateShift(sessionId, &request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	shiftId, templateId, err := CreateShift(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "create shift error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	shiftResp := createShiftResp(shiftId, templateId)

	logger.LogInfo("Response returned from CreateShiftHandler", sessionId)

	ctx.JSON(http.StatusOK, shiftResp)
}

// re-seats an existing trip in one call
func UpdateShiftSeatsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateShiftSeatsHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request UpdateShiftSeatsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "seats"),
		})
		return
	}

	logger.LogDebug2("Request received in UpdateShiftSeatsHandler", sessionId, request)

	if err := ValidateUpdateShiftSeats(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := UpdateShiftSeats(ctx, sessionId, driverId, request); err != nil {
		logger.LogError(sessionId, "update seats error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from UpdateShiftSeatsHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Updated_Successfully, "Seats")))
}

// returns one trip with its route in order and its seats
func GetShiftHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetShiftHandler", sessionId)

	userId := ctx.GetString(constants.User_KEY)
	shiftId := ctx.Query(constants.Shift_Key)

	if err := ValidateShiftId(shiftId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	shift, stops, seats, err := GetShift(ctx, sessionId, userId, shiftId)
	if err != nil {
		logger.LogError(sessionId, "get shift error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	detailsResp := shiftDetailsResp(shift, stops, seats)

	logger.LogInfo("Response returned from GetShiftHandler", sessionId)

	ctx.JSON(http.StatusOK, detailsResp)
}

// the roster of a fleet, this is the view a manager works from
func GetGroupShiftsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetGroupShiftsHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	groupId := ctx.Query(constants.Group_Key)

	if err := ValidateGroupId(groupId); err != nil {
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

	filterDriverId, direction, startTime, endTime, status := getShiftQueryParams(ctx)

	shifts, totalRows, err := GetGroupShifts(ctx, sessionId, driverId, groupId, filterDriverId, direction, startTime, endTime, status, page)
	if err != nil {
		logger.LogError(sessionId, "get shifts error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	shiftsResp := shiftsResp(shifts, totalRows)

	logger.LogInfo("Response returned from GetGroupShiftsHandler", sessionId)

	ctx.JSON(http.StatusOK, shiftsResp)
}

// the trips the caller is on, whether they drive them or hold a seat on them
func GetMyShiftsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetMyShiftsHandler", sessionId)

	userId := ctx.GetString(constants.User_KEY)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	_, direction, startTime, endTime, _ := getShiftQueryParams(ctx)

	shifts, totalRows, err := GetMyShifts(ctx, sessionId, userId, direction, startTime, endTime, page)
	if err != nil {
		logger.LogError(sessionId, "get shifts error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	shiftsResp := shiftsResp(shifts, totalRows)

	logger.LogInfo("Response returned from GetMyShiftsHandler", sessionId)

	ctx.JSON(http.StatusOK, shiftsResp)
}

// calls a trip off and tells everybody who was counting on it
func CancelShiftHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in CancelShiftHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	shiftId := ctx.Query(constants.Shift_Key)

	if err := ValidateShiftId(shiftId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := CancelShift(ctx, sessionId, driverId, shiftId); err != nil {
		logger.LogError(sessionId, "cancel shift error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from CancelShiftHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Deleted_Successfully, "Shift")))
}

// the saved shapes of past shifts, the app prefills the create form with one
func GetShiftTemplatesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetShiftTemplatesHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	groupId := ctx.Query(constants.Group_Key)

	if err := ValidateGroupId(groupId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	templates, err := GetShiftTemplates(ctx, sessionId, driverId, groupId)
	if err != nil {
		logger.LogError(sessionId, "get templates error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	templatesResp := shiftTemplatesResp(templates)

	logger.LogInfo("Response returned from GetShiftTemplatesHandler", sessionId)

	ctx.JSON(http.StatusOK, templatesResp)
}

func DeleteShiftTemplateHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeleteShiftTemplateHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	templateId := ctx.Query(constants.Shift_Template_Key)

	if err := ValidateTemplateId(templateId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := DeleteShiftTemplate(ctx, sessionId, driverId, templateId); err != nil {
		logger.LogError(sessionId, "delete template error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DeleteShiftTemplateHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Deleted_Successfully, "Shift template")))
}

func getShiftQueryParams(ctx *gin.Context) (driverId, direction, startTime, endTime, status string) {
	driverId = ctx.DefaultQuery(constants.Driver_Key, "")
	direction = ctx.DefaultQuery(constants.Direction_Key, "")
	startTime = ctx.DefaultQuery(constants.Start_Time_Key, "")
	endTime = ctx.DefaultQuery(constants.Extimated_End_Time_Key, "")
	status = ctx.DefaultQuery(constants.Status_Key, "")

	return
}

// moves a shift in time, or rewrites its route text, without losing the seat plan
func RescheduleShiftHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RescheduleShiftHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request RescheduleShiftRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "shift"),
		})
		return
	}

	logger.LogDebug2("Request received in RescheduleShiftHandler", sessionId, request)

	if err := ValidateRescheduleShift(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := RescheduleShift(ctx, sessionId, driverId, request); err != nil {
		logger.LogError(sessionId, "reschedule shift error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from RescheduleShiftHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Updated_Successfully, "Shift")))
}
