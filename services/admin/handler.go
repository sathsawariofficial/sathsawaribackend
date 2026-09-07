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

	adminId := ctx.GetString(constants.User_KEY)
	roleId := ctx.Query(constants.Role_Key)

	if err := utils.ValidateId(roleId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "role id"),
		})
		return
	}

	if err := DeleteRole(ctx, sessionId, adminId, roleId); err != nil {
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

// the whole platform counted in one call, the admin's first screen
func GetOverviewHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetOverviewHandler", sessionId)

	overview, err := GetPlatformOverview(ctx, sessionId)
	if err != nil {
		logger.LogError(sessionId, "get overview error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	overviewResp := platformOverviewResp(overview)

	logger.LogInfo("Response returned from GetOverviewHandler", sessionId)

	ctx.JSON(http.StatusOK, overviewResp)
}

// lists passenger accounts, searchable by name or number
func GetPassengersHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPassengersHandler", sessionId)

	search := ctx.DefaultQuery(constants.Search_Loc_Key, "")
	status := ctx.DefaultQuery(constants.Status_Key, "")

	if err := ValidateSearchFilters(status); err != nil {
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

	passengers, totalRows, err := GetPassengers(ctx, sessionId, search, status, page)
	if err != nil {
		logger.LogError(sessionId, "get passengers error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	passengersResp := passengersResp(passengers, totalRows)

	logger.LogInfo("Response returned from GetPassengersHandler", sessionId)

	ctx.JSON(http.StatusOK, passengersResp)
}

// one passenger with the fleets they ride with and the travel form they filled in
func GetPassengerProfileHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPassengerProfileHandler", sessionId)

	passengerId := ctx.Query(constants.Passenger_Key)

	if err := utils.ValidateId(passengerId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "passenger id"),
		})
		return
	}

	passenger, groups, preferences, err := GetPassengerProfile(ctx, sessionId, passengerId)
	if err != nil {
		logger.LogError(sessionId, "get passenger error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	profileResp := passengerProfileResp(passenger, groups, preferences)

	logger.LogInfo("Response returned from GetPassengerProfileHandler", sessionId)

	ctx.JSON(http.StatusOK, profileResp)
}

// switches a passenger account on or off without destroying anything
func UpdatePassengerStatusHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdatePassengerStatusHandler", sessionId)

	adminId := ctx.GetString(constants.User_KEY)
	passengerId := ctx.Query(constants.Passenger_Key)

	var request AdminStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "passenger"),
		})
		return
	}

	if err := ValidateAccountStatus(passengerId, &request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := UpdatePassengerStatus(ctx, sessionId, adminId, passengerId, request.Status); err != nil {
		logger.LogError(sessionId, "update passenger status error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from UpdatePassengerStatusHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Updated_Successfully, "Passenger")))
}

// removes a passenger account and everything still pointing at it
func DeletePassengerHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeletePassengerHandler", sessionId)

	adminId := ctx.GetString(constants.User_KEY)
	passengerId := ctx.Query(constants.Passenger_Key)

	if err := utils.ValidateId(passengerId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "passenger id"),
		})
		return
	}

	if err := DeletePassenger(ctx, sessionId, adminId, passengerId); err != nil {
		logger.LogError(sessionId, "delete passenger error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DeletePassengerHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Deleted_Successfully, "Passenger")))
}

// switches a driver account on or off, the softer alternative to deleting them
func UpdateDriverStatusHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateDriverStatusHandler", sessionId)

	adminId := ctx.GetString(constants.User_KEY)
	driverId := ctx.Query(constants.Driver_Key)

	var request AdminStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "driver"),
		})
		return
	}

	if err := ValidateAccountStatus(driverId, &request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := UpdateDriverStatus(ctx, sessionId, adminId, driverId, request.Status); err != nil {
		logger.LogError(sessionId, "update driver status error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from UpdateDriverStatusHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Updated_Successfully, "Driver")))
}

// every fleet on the platform with the size of each
func GetGroupsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetGroupsHandler", sessionId)

	search := ctx.DefaultQuery(constants.Search_Loc_Key, "")
	status := ctx.DefaultQuery(constants.Status_Key, "")

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	groups, totalRows, err := GetGroups(ctx, sessionId, search, status, page)
	if err != nil {
		logger.LogError(sessionId, "get groups error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	groupsResp := adminGroupsResp(groups, totalRows)

	logger.LogInfo("Response returned from GetGroupsHandler", sessionId)

	ctx.JSON(http.StatusOK, groupsResp)
}

// one fleet with its full rosters, including everybody who was turned away
func GetGroupDetailsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetGroupDetailsHandler", sessionId)

	groupId := ctx.Query(constants.Group_Key)

	if err := utils.ValidateId(groupId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "group id"),
		})
		return
	}

	_, overview, members, vehicles, passengers, err := GetGroupDetails(ctx, sessionId, groupId)
	if err != nil {
		logger.LogError(sessionId, "get group error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	detailsResp := adminGroupDetailsResp(overview, members, vehicles, passengers)

	logger.LogInfo("Response returned from GetGroupDetailsHandler", sessionId)

	ctx.JSON(http.StatusOK, detailsResp)
}

// shuts a fleet down, or brings it back, without deleting its history
func UpdateGroupStatusHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateGroupStatusHandler", sessionId)

	groupId := ctx.Query(constants.Group_Key)

	var request AdminStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "group"),
		})
		return
	}

	if err := ValidateAccountStatus(groupId, &request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := UpdateGroupStatus(ctx, sessionId, groupId, request.Status); err != nil {
		logger.LogError(sessionId, "update group status error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from UpdateGroupStatusHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Updated_Successfully, "Group")))
}

// every shift on the platform, filterable
func GetShiftsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetShiftsHandler", sessionId)

	groupId := ctx.DefaultQuery(constants.Group_Key, "")
	driverId := ctx.DefaultQuery(constants.Driver_Key, "")
	direction := ctx.DefaultQuery(constants.Direction_Key, "")
	startTime := ctx.DefaultQuery(constants.Start_Time_Key, "")
	endTime := ctx.DefaultQuery(constants.Extimated_End_Time_Key, "")
	status := ctx.DefaultQuery(constants.Status_Key, "")

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	shifts, totalRows, err := GetShifts(ctx, sessionId, groupId, driverId, direction, startTime, endTime, status, page)
	if err != nil {
		logger.LogError(sessionId, "get shifts error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	shiftsResp := adminShiftsResp(shifts, totalRows)

	logger.LogInfo("Response returned from GetShiftsHandler", sessionId)

	ctx.JSON(http.StatusOK, shiftsResp)
}

// one shift with its route in order and everybody aboard
func GetShiftDetailsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetShiftDetailsHandler", sessionId)

	shiftId := ctx.Query(constants.Shift_Key)

	if err := utils.ValidateId(shiftId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "shift id"),
		})
		return
	}

	shift, stops, seats, err := GetShiftDetails(ctx, sessionId, shiftId)
	if err != nil {
		logger.LogError(sessionId, "get shift error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	detailsResp := adminShiftDetailsResp(shift, stops, seats)

	logger.LogInfo("Response returned from GetShiftDetailsHandler", sessionId)

	ctx.JSON(http.StatusOK, detailsResp)
}
