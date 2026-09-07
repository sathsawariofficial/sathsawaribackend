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

// creates a role an admin can hand out inside a group
func CreateRoleHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in CreateRoleHandler", sessionId)

	var request AdminRoleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Creation_Failed, "role"),
		})
		return
	}

	logger.LogDebug2("Request received in CreateRoleHandler", sessionId, request)

	if err := ValidateRoleRequest(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	roleId, err := CreateRole(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "create role error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	roleResp := createdRoleResp(roleId)

	logger.LogInfo("Response returned from CreateRoleHandler", sessionId)

	ctx.JSON(http.StatusOK, roleResp)
}

// lists every role with the permissions it currently holds
func GetRolesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetRolesHandler", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	roles, permissions, totalRows, err := GetRoles(ctx, sessionId, page)
	if err != nil {
		logger.LogError(sessionId, "get roles error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rolesResp := rolesResp(roles, permissions, totalRows)

	logger.LogInfo("Response returned from GetRolesHandler", sessionId)

	ctx.JSON(http.StatusOK, rolesResp)
}

// updates the name or the description of a role
func UpdateRoleHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateRoleHandler", sessionId)

	roleId := ctx.Query(constants.Role_Key)

	var request AdminRoleUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "role"),
		})
		return
	}

	logger.LogDebug2("Request received in UpdateRoleHandler", sessionId, request)

	if err := ValidateRoleUpdateRequest(roleId, &request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := UpdateRole(ctx, sessionId, roleId, request); err != nil {
		logger.LogError(sessionId, "update role error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from UpdateRoleHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Updated_Successfully, "Role")))
}

// deletes a role that is neither a system role nor still held by a member
func DeleteRoleHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeleteRoleHandler", sessionId)

	roleId := ctx.Query(constants.Role_Key)

	if err := utils.ValidateId(roleId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "role id"),
		})
		return
	}

	if err := DeleteRole(ctx, sessionId, roleId); err != nil {
		logger.LogError(sessionId, "delete role error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DeleteRoleHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Deleted_Successfully, "Role")))
}

// creates a permission that can then be mapped onto any role
func CreatePermissionHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in CreatePermissionHandler", sessionId)

	var request AdminPermissionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Creation_Failed, "permission"),
		})
		return
	}

	logger.LogDebug2("Request received in CreatePermissionHandler", sessionId, request)

	if err := ValidatePermissionRequest(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	permissionId, err := CreatePermission(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "create permission error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	permissionResp := createdPermissionResp(permissionId)

	logger.LogInfo("Response returned from CreatePermissionHandler", sessionId)

	ctx.JSON(http.StatusOK, permissionResp)
}

// lists every permission the system knows about
func GetPermissionsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPermissionsHandler", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	permissions, totalRows, err := GetPermissions(ctx, sessionId, page)
	if err != nil {
		logger.LogError(sessionId, "get permissions error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	permissionsResp := permissionsResp(permissions, totalRows)

	logger.LogInfo("Response returned from GetPermissionsHandler", sessionId)

	ctx.JSON(http.StatusOK, permissionsResp)
}

// replaces the whole permission set of a role in one call
func SetRolePermissionsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SetRolePermissionsHandler", sessionId)

	var request AdminRolePermissionsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "role permissions"),
		})
		return
	}

	logger.LogDebug2("Request received in SetRolePermissionsHandler", sessionId, request)

	if err := ValidateRolePermissionsRequest(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := SetRolePermissions(ctx, sessionId, request); err != nil {
		logger.LogError(sessionId, "set role permissions error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from SetRolePermissionsHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Updated_Successfully, "Role permissions")))
}
