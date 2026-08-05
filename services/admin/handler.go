package admin

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

// create a session for admin
func LoginAdminHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in LoginAdminHandler", sessionId)

	var request AdminLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: constants.General_Error,
		})
		return
	}

	logger.LogDebug2("Request received in LoginAdminHandler", sessionId, request)

	err := ValidateAdminLogin(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	adminSessionId, admin, err := LoginAdmin(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "login error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	adminLoginResp := loginAdminResp(adminSessionId, admin)

	logger.LogInfo("Response received in LoginAdminHandler", sessionId)
	logger.LogDebug2("Response received in LoginAdminHandler", sessionId, adminLoginResp)

	ctx.JSON(http.StatusOK, adminLoginResp)
}

func GetDriverDetailsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetDriverDetailsHandler", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	err = ValidatePage(page)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	driverDetails, totalRows, err := GetDrivers(ctx, sessionId, page)
	if err != nil {
		logger.LogError(sessionId, "get driver details error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	if totalRows == 0 {
		logger.LogError(sessionId, "no drivers found error")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusNoContent,
			Message: fmt.Sprintf(constants.Not_Found, "Driver"),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	driverDetailsResp := driverDetailsResp(driverDetails, totalRows)

	logger.LogInfo("Response returned from GetDriverDetailsHandler", sessionId)
	logger.LogDebug2("Response returned from GetDriverDetailsHandler", sessionId, driverDetailsResp)

	ctx.JSON(http.StatusOK, driverDetailsResp)
}

func GetVehiclesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetVehiclesHandler", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: VehicleDetailsResponse{
				Vehicles: []Vehicle{},
			},
		})
		return
	}

	err = ValidatePage(page)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: VehicleDetailsResponse{
				Vehicles: []Vehicle{},
			},
		})
		return
	}

	vehicles, totalRows, err := GetVehicles(ctx, sessionId, page)
	if err != nil {
		logger.LogError(sessionId, "get vechile details error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: VehicleDetailsResponse{
				Vehicles: []Vehicle{},
			},
		})
		return
	}

	if totalRows == 0 {
		logger.LogError(sessionId, "no vehicles found error")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusNoContent,
			Message: fmt.Sprintf(constants.Not_Found, "Vehicles"),
			Data: VehicleDetailsResponse{
				Vehicles: []Vehicle{},
			},
		})
		return
	}

	vehicleDetailsResp := vechileDetailsResp(vehicles, totalRows)

	logger.LogInfo("Response returned from GetVehiclesHandler", sessionId)
	logger.LogDebug2("Response returned from GetVehiclesHandler", sessionId, vehicleDetailsResp)

	ctx.JSON(http.StatusOK, vehicleDetailsResp)
}

func GetRidesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetRidesHandler", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: RideDetailsResponse{
				Rides: []RideDetail{},
			},
		})
		return
	}

	err = ValidatePage(page)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: RideDetailsResponse{
				Rides: []RideDetail{},
			},
		})
		return
	}

	rides, totalRows, err := GetRides(ctx, sessionId, page)
	if err != nil {
		logger.LogError(sessionId, "get ride details error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: RideDetailsResponse{
				Rides: []RideDetail{},
			},
		})
		return
	}

	if totalRows == 0 {
		logger.LogError(sessionId, "no rides found error")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusNoContent,
			Message: fmt.Sprintf(constants.Not_Found, "Rides"),
			Data: RideDetailsResponse{
				Rides: []RideDetail{},
			},
		})
		return
	}

	rideDetailsResp := rideDetailsResp(rides, totalRows)

	logger.LogInfo("Response returned from GetRidesHandler", sessionId)
	logger.LogDebug2("Response returned from GetRidesHandler", sessionId, rideDetailsResp)

	ctx.JSON(http.StatusOK, rideDetailsResp)
}

func DeleteDriverHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeleteDriverHandler", sessionId)

	err := DeleteDriver(ctx, sessionId)
	if err != nil {
		logger.LogError(sessionId, err)
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	logger.LogInfo("Response returned from DeleteDriverHandler", sessionId)
	logger.LogDebug2("Response returned from DeleteDriverHandler", sessionId, driverDetailsResp)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(constants.Success))
}

func AdminBroadcastHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in AdminBroadcastHandler", sessionId)

	var request AdminBroadcastRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: constants.General_Error,
		})
		return
	}

	logger.LogDebug2("Request received in AdminBroadcastHandler", sessionId, request)

	err := ValidateBoardcastRequest(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = CreateAdminBroadcastRequest(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "broadcast error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from AdminBroadcastHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(constants.Success))
}

func AnnouncementHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in AnnouncementHandler", sessionId)

	var request AnnouncementRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: constants.General_Error,
		})
		return
	}

	logger.LogDebug2("Request received in AnnouncementHandler", sessionId, request)

	err := ValidateAnnouncementRequest(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = CreateAnnouncement(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "announcement creation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from AnnouncementHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(constants.Success))
}

func GetApprochRequestsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetApprochRequestsHandler", sessionId)

	approchType := ctx.Query(constants.Type_Key)
	page := ctx.Query(constants.Page_Key)

	err := ValidateApprochInfoReq(approchType, page)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	approches, totalRows, err := GetApprochRequest(ctx, sessionId, approchType, utils.ToInt(page))
	if err != nil {
		logger.LogError(sessionId, err)
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	resp := createApprochInfoResp(approches, totalRows)

	logger.LogInfo("Response returned from GetApprochRequestsHandler", sessionId)
	logger.LogDebug2("Response returned from GetApprochRequestsHandler", sessionId, resp)

	ctx.JSON(http.StatusOK, resp)
}
