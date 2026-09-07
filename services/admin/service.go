package admin

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func LoginAdmin(ctx *gin.Context, sessionId string, request AdminLoginRequest) (token string, admin postgress.Admin, err error) {
	logger.LogInfo("Request returned from LoginAdmin", sessionId)

	admin, err = getAdmin(ctx, request.Username)
	if err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Login_Failed)
		return
	}
	if utils.IsStringEmpty(admin.ID) {
		logger.LogError(sessionId, "admin not found error")
		err = fmt.Errorf(constants.General_Unknown, "admin")
		return
	}

	token, err = loginTokenCreation(sessionId, request, admin)
	if err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Login_Failed)
		return
	}

	logger.LogInfo("Response returned from LoginAdmin", sessionId)

	return
}

func GetDrivers(ctx *gin.Context, sessionId string, page int) (driverDetails []postgress.Driver, totalRows int64, err error) {
	logger.LogInfo("Request received in GetDrivers", sessionId)

	driverDetails, totalRows, err = getAllDriversWithVehicles(ctx, page)
	if err != nil {
		logger.LogError(sessionId, " get driver details error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetDrivers", sessionId)
	logger.LogDebug2("Response returned from GetDrivers", sessionId, fmt.Sprintf("driver details: %v, total rows: %v", driverDetails, totalRows))

	return
}

func GetVehicles(ctx *gin.Context, sessionId string, page int) (vehicles []postgress.Vehicle, totalRows int64, err error) {
	logger.LogInfo("Request received in GetVehicles", sessionId)

	vehicles, totalRows, err = getAllVehicles(ctx, page)
	if err != nil {
		logger.LogError(sessionId, " get driver details error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetVehicles", sessionId)
	logger.LogDebug2("Response returned from GetVehicles", sessionId, fmt.Sprintf("vehicles: %v, total rows: %v", vehicles, totalRows))

	return
}

func GetRides(ctx *gin.Context, sessionId string, page int) (rides []postgress.RideDetails, totalRows int64, err error) {
	logger.LogInfo("Request received in GetRides", sessionId)

	rides, totalRows, err = getAllRides(ctx, page)
	if err != nil {
		logger.LogError(sessionId, " get ride details error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetRides", sessionId)
	logger.LogDebug2("Response returned from GetRides", sessionId, fmt.Sprintf("rides: %v, total rows: %v", rides, totalRows))

	return
}

func DeleteDriver(ctx *gin.Context, sessionId string) (err error) {
	logger.LogInfo("Request received in DeleteDriver", sessionId)

	driverId := ctx.Query(constants.User_KEY)
	adminId := ctx.GetString(constants.User_KEY)
	logger.LogDebug("Driver to be deleted by admin", sessionId, driverId)

	driver, err := database.GetDriverById(ctx, driverId)
	if err != nil {
		logger.LogError(sessionId, "error failed to get driver: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "find driver")
		return
	}
	if utils.IsStringEmpty(driver.ID) {
		err = fmt.Errorf(constants.Failed_To_Do_Job, "find driver")
		logger.LogError(sessionId, err)
		return
	}

	if strings.EqualFold(driver.Status, constants.Status_InActive) {
		err = fmt.Errorf(constants.Failed_To_Do_Job, "driver, as it is already deleted")
		logger.LogError(sessionId, "error failed to delete driver: "+err.Error())
		return
	}

	err = database.DeleteDriver(ctx, driver, adminId)
	if err != nil {
		logger.LogError(sessionId, "error failed to delete driver status: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "finish operation")
		return
	}

	logger.LogInfo("Response returned from DeleteDriver", sessionId)

	return
}

func CreateAdminBroadcastRequest(ctx *gin.Context, sessionId string, request AdminBroadcastRequest) (err error) {
	logger.LogInfo("Request received in CreateAdminBroadcastRequest", sessionId)

	err = createAdminBroadcast(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, " create broadcast request: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from CreateAdminBroadcastRequest", sessionId)

	return
}

func CreateAnnouncement(ctx *gin.Context, sessionId string, request AnnouncementRequest) (err error) {
	logger.LogInfo("Request received in CreateAnnouncement", sessionId)

	err = createAnnouncement(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, " create broadcast request: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from CreateAnnouncement", sessionId)

	return
}

func GetApprochRequest(ctx *gin.Context, sessionId, approchType string, page int) (approches []postgress.ApprochInfo, totalRows int64, err error) {
	logger.LogInfo("Request received in GetApprochRequest", sessionId)
	logger.LogDebug("Request received in GetApprochRequest", sessionId, fmt.Sprintf("approch type: %s, page: %d", approchType, page))

	approches, totalRows, err = getApprochRequests(ctx, sessionId, approchType, page)
	if err != nil {
		logger.LogError(sessionId, " create broadcast request: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetApprochRequest", sessionId)
	logger.LogDebug("Response returned from GetApprochRequest", sessionId, totalRows)
	logger.LogDebug("Response returned from GetApprochRequest", sessionId, approches)

	return
}

func CreateRole(ctx *gin.Context, sessionId string, request AdminRoleRequest) (roleId string, err error) {
	logger.LogInfo("Request received in CreateRole", sessionId)

	count, err := countRolesByName(ctx, request.Name, "")
	if err != nil {
		logger.LogError(sessionId, "failed to look for an existing role error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}
	if count > 0 {
		err = fmt.Errorf(constants.Invalid_Data, "role name, it is already taken")
		logger.LogError(sessionId, err)
		return
	}

	role, err := createRole(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to create role error: "+err.Error())
		err = fmt.Errorf(constants.Creation_Failed, "role")
		return
	}
	roleId = role.ID

	logger.LogInfo("Response returned from CreateRole", sessionId)
	logger.LogDebug2("Response returned from CreateRole", sessionId, roleId)

	return
}

func GetRoles(ctx *gin.Context, sessionId string, page int) (roles []postgress.Role, permissions map[string][]adminRolePermissionRow, totalRows int64, err error) {
	logger.LogInfo("Request received in GetRoles", sessionId)

	roles, permissions, totalRows, err = getRolesPage(ctx, sessionId, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get roles error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the roles")
		return
	}

	logger.LogInfo("Response returned from GetRoles", sessionId)
	logger.LogDebug2("Response returned from GetRoles", sessionId, totalRows)

	return
}

// UpdateRole edits a role. A system role may have its description changed but never
// its name, the group permission checks look those roles up by name so a rename
// would quietly cut every owner and sub manager off from their group.
func UpdateRole(ctx *gin.Context, sessionId, roleId string, request AdminRoleUpdateRequest) (err error) {
	logger.LogInfo("Request received in UpdateRole", sessionId)

	role, err := getRoleById(ctx, roleId)
	if err != nil {
		logger.LogError(sessionId, "failed to get role error: "+err.Error())
		err = errors.New(constants.Role_Not_Found)
		return
	}

	updates := map[string]interface{}{}

	if request.Name != nil && !strings.EqualFold(*request.Name, role.Name) {
		if role.IsSystem {
			err = errors.New(constants.Operation_Not_Permitted)
			logger.LogError(sessionId, "a system role cannot be renamed error: "+err.Error())
			return
		}

		count, e := countRolesByName(ctx, *request.Name, roleId)
		if e != nil {
			logger.LogError(sessionId, "failed to look for an existing role error: "+e.Error())
			err = errors.New(constants.Unknown_Error)
			return
		}
		if count > 0 {
			err = fmt.Errorf(constants.Invalid_Data, "role name, it is already taken")
			logger.LogError(sessionId, err)
			return
		}

		updates["name"] = *request.Name
	}

	if request.Description != nil {
		updates["description"] = *request.Description
	}

	if len(updates) == 0 {
		logger.LogInfo("Response returned from UpdateRole", sessionId)
		return
	}

	if err = updateRole(ctx, sessionId, roleId, updates); err != nil {
		logger.LogError(sessionId, "failed to update role error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "role")
		return
	}

	logger.LogInfo("Response returned from UpdateRole", sessionId)

	return
}

func DeleteRole(ctx *gin.Context, sessionId, adminId, roleId string) (err error) {
	logger.LogInfo("Request received in DeleteRole", sessionId)

	role, err := getRoleById(ctx, roleId)
	if err != nil {
		logger.LogError(sessionId, "failed to get role error: "+err.Error())
		err = errors.New(constants.Role_Not_Found)
		return
	}

	if role.IsSystem {
		err = errors.New(constants.Not_System_Role)
		logger.LogError(sessionId, err)
		return
	}

	count, err := countMembersUsingRole(ctx, roleId)
	if err != nil {
		logger.LogError(sessionId, "failed to count the members holding the role error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}
	if count > 0 {
		err = errors.New(constants.Operation_Not_Permitted)
		logger.LogError(sessionId, fmt.Sprintf("role is still held by %d member(s) error: %s", count, err.Error()))
		return
	}

	if err = deleteRole(ctx, sessionId, adminId, role); err != nil {
		logger.LogError(sessionId, "failed to delete role error: "+err.Error())
		err = fmt.Errorf(constants.DELETE_Failed, "role")
		return
	}

	logger.LogInfo("Response returned from DeleteRole", sessionId)

	return
}

func CreatePermission(ctx *gin.Context, sessionId string, request AdminPermissionRequest) (permissionId string, err error) {
	logger.LogInfo("Request received in CreatePermission", sessionId)

	count, err := countPermissionsByCode(ctx, request.Code)
	if err != nil {
		logger.LogError(sessionId, "failed to look for an existing permission error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}
	if count > 0 {
		err = fmt.Errorf(constants.Invalid_Data, "permission code, it is already taken")
		logger.LogError(sessionId, err)
		return
	}

	permission, err := createPermission(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to create permission error: "+err.Error())
		err = fmt.Errorf(constants.Creation_Failed, "permission")
		return
	}
	permissionId = permission.ID

	logger.LogInfo("Response returned from CreatePermission", sessionId)
	logger.LogDebug2("Response returned from CreatePermission", sessionId, permissionId)

	return
}

func GetPermissions(ctx *gin.Context, sessionId string, page int) (permissions []postgress.Permission, totalRows int64, err error) {
	logger.LogInfo("Request received in GetPermissions", sessionId)

	permissions, totalRows, err = getPermissionsPage(ctx, sessionId, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get permissions error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the permissions")
		return
	}

	logger.LogInfo("Response returned from GetPermissions", sessionId)
	logger.LogDebug2("Response returned from GetPermissions", sessionId, totalRows)

	return
}

// SetRolePermissions replaces the whole permission set of a role in one call. It is
// deliberately allowed on system roles too, remapping what a role may do without a
// code change is the entire point of keeping permissions in the database.
func SetRolePermissions(ctx *gin.Context, sessionId string, request AdminRolePermissionsRequest) (err error) {
	logger.LogInfo("Request received in SetRolePermissions", sessionId)

	if _, err = getRoleById(ctx, request.RoleId); err != nil {
		logger.LogError(sessionId, "failed to get role error: "+err.Error())
		err = errors.New(constants.Role_Not_Found)
		return
	}

	if len(request.PermissionIds) > 0 {
		count, e := countPermissionsByIds(ctx, request.PermissionIds)
		if e != nil {
			logger.LogError(sessionId, "failed to validate the permissions error: "+e.Error())
			err = errors.New(constants.Unknown_Error)
			return
		}

		if int(count) != len(request.PermissionIds) {
			err = errors.New(constants.Permission_Not_Found)
			logger.LogError(sessionId, "one or more permissions do not exist error: "+err.Error())
			return
		}
	}

	if err = replaceRolePermissions(ctx, sessionId, request.RoleId, request.PermissionIds); err != nil {
		logger.LogError(sessionId, "failed to set the role permissions error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "role permissions")
		return
	}

	logger.LogInfo("Response returned from SetRolePermissions", sessionId)

	return
}

// GetPlatformOverview is the admin's first screen: how much of everything exists and
// how much of it is live.
func GetPlatformOverview(ctx *gin.Context, sessionId string) (overview postgress.PlatformOverview, err error) {
	logger.LogInfo("Request received in GetPlatformOverview", sessionId)

	overview, err = getPlatformOverview(ctx, sessionId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the overview error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the overview")
		return
	}

	logger.LogInfo("Response returned from GetPlatformOverview", sessionId)

	return
}

func GetPassengers(ctx *gin.Context, sessionId, search, status string, page int) (passengers []postgress.Passenger, totalRows int64, err error) {
	logger.LogInfo("Request received in GetPassengers", sessionId)

	passengers, totalRows, err = getAllPassengers(ctx, sessionId, search, status, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the passengers error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the passengers")
		return
	}

	logger.LogInfo("Response returned from GetPassengers", sessionId)

	return
}

// GetPassengerProfile explains one account: who they ride with and what they asked
// for, which is what a complaint or a support call actually needs.
func GetPassengerProfile(ctx *gin.Context, sessionId, passengerId string) (
	passenger postgress.Passenger,
	groups []postgress.GroupPassengerDetails,
	preferences []postgress.PassengerLocationPreference,
	err error,
) {
	logger.LogInfo("Request received in GetPassengerProfile", sessionId)

	passenger, err = getPassengerById(ctx, passengerId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the passenger error: "+err.Error())
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	groups, preferences, err = getPassengerContext(ctx, passengerId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the passenger context error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the passenger")
		return
	}

	logger.LogInfo("Response returned from GetPassengerProfile", sessionId)

	return
}

// UpdatePassengerStatus is the softer moderation tool: an account switched off
// cannot log in or be seated, but nothing it was part of is destroyed.
func UpdatePassengerStatus(ctx *gin.Context, sessionId, adminId, passengerId, status string) (err error) {
	logger.LogInfo("Request received in UpdatePassengerStatus", sessionId)

	if _, err = getPassengerById(ctx, passengerId); err != nil {
		logger.LogError(sessionId, "failed to get the passenger error: "+err.Error())
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	if err = updatePassengerStatus(ctx, sessionId, passengerId, status, adminId); err != nil {
		logger.LogError(sessionId, "failed to update the passenger error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "passenger")
		return
	}

	logger.LogInfo("Response returned from UpdatePassengerStatus", sessionId)

	return
}

// DeletePassenger removes an account for good. The shared helper frees the seats it
// held on upcoming shifts and ends its fleet memberships, so nothing anywhere is
// left pointing at somebody who no longer exists.
func DeletePassenger(ctx *gin.Context, sessionId, adminId, passengerId string) (err error) {
	logger.LogInfo("Request received in DeletePassenger", sessionId)

	passenger, err := getPassengerById(ctx, passengerId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the passenger error: "+err.Error())
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	if err = database.DeletePassenger(ctx, passenger, adminId); err != nil {
		logger.LogError(sessionId, "failed to delete the passenger error: "+err.Error())
		err = fmt.Errorf(constants.DELETE_Failed, "passenger")
		return
	}

	logger.LogInfo("Response returned from DeletePassenger", sessionId)

	return
}

func UpdateDriverStatus(ctx *gin.Context, sessionId, adminId, driverId, status string) (err error) {
	logger.LogInfo("Request received in UpdateDriverStatus", sessionId)

	driver, err := database.GetDriverById(ctx, driverId)
	if err != nil || utils.IsStringEmpty(driver.ID) {
		logger.LogError(sessionId, "failed to get the driver error")
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	if err = updateDriverStatus(ctx, sessionId, driverId, status, adminId); err != nil {
		logger.LogError(sessionId, "failed to update the driver error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "driver")
		return
	}

	logger.LogInfo("Response returned from UpdateDriverStatus", sessionId)

	return
}

func GetGroups(ctx *gin.Context, sessionId, search, status string, page int) (groups []postgress.AdminGroupOverview, totalRows int64, err error) {
	logger.LogInfo("Request received in GetGroups", sessionId)

	groups, totalRows, err = getAllGroups(ctx, sessionId, search, status, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the groups error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the groups")
		return
	}

	logger.LogInfo("Response returned from GetGroups", sessionId)

	return
}

// GetGroupDetails is the admin's read of a fleet, unfiltered by status so somebody
// looking into a complaint can see who was turned away as well as who was let in.
func GetGroupDetails(ctx *gin.Context, sessionId, groupId string) (
	group postgress.Group,
	overview postgress.AdminGroupOverview,
	members []postgress.GroupMemberDetails,
	vehicles []postgress.GroupVehicleDetails,
	passengers []postgress.GroupPassengerDetails,
	err error,
) {
	logger.LogInfo("Request received in GetGroupDetails", sessionId)

	group, err = getGroupById(ctx, groupId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the group error: "+err.Error())
		err = errors.New(constants.Group_Not_Found)
		return
	}

	owner, e := database.GetDriverById(ctx, group.OwnerDriverID)
	if e != nil {
		logger.LogError(sessionId, e)
	}

	members, vehicles, passengers, err = getGroupRosters(ctx, sessionId, groupId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the rosters error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the group")
		return
	}

	overview = postgress.AdminGroupOverview{
		ID:            group.ID,
		Name:          group.Name,
		Description:   group.Description,
		Status:        group.Status,
		OwnerDriverID: group.OwnerDriverID,
		OwnerName:     owner.DriverName,
		OwnerMobile:   owner.DriverMobile,
		CreatedAt:     group.CreatedAt,
	}

	for _, member := range members {
		if member.Status == constants.Membership_Status_Approved {
			overview.MemberCount++
		}
	}
	for _, vehicle := range vehicles {
		if vehicle.Status == constants.Membership_Status_Approved {
			overview.VehicleCount++
		}
	}
	for _, passenger := range passengers {
		if passenger.Status == constants.Membership_Status_Approved {
			overview.PassengerCount++
		}
	}

	logger.LogInfo("Response returned from GetGroupDetails", sessionId)

	return
}

// UpdateGroupStatus lets an admin shut a fleet down without deleting anything, which
// is the right hammer for a group that is misbehaving. Switching it off stops it
// being found, joined or built on, while its history stays intact.
func UpdateGroupStatus(ctx *gin.Context, sessionId, groupId, status string) (err error) {
	logger.LogInfo("Request received in UpdateGroupStatus", sessionId)

	if _, err = getGroupById(ctx, groupId); err != nil {
		logger.LogError(sessionId, "failed to get the group error: "+err.Error())
		err = errors.New(constants.Group_Not_Found)
		return
	}

	if err = updateGroupStatus(ctx, sessionId, groupId, status); err != nil {
		logger.LogError(sessionId, "failed to update the group error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "group")
		return
	}

	logger.LogInfo("Response returned from UpdateGroupStatus", sessionId)

	return
}

func GetShifts(ctx *gin.Context, sessionId, groupId, driverId, direction, startTime, endTime, status string, page int) (shifts []postgress.ShiftDetails, totalRows int64, err error) {
	logger.LogInfo("Request received in GetShifts", sessionId)

	shifts, totalRows, err = getAllShifts(ctx, sessionId, groupId, driverId, direction, startTime, endTime, status, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the shifts error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shifts")
		return
	}

	logger.LogInfo("Response returned from GetShifts", sessionId)

	return
}

// GetShiftDetails is the whole trip: who is driving, who is aboard, and the route in
// the order it is driven.
func GetShiftDetails(ctx *gin.Context, sessionId, shiftId string) (
	shift postgress.ShiftDetails,
	stops []postgress.ShiftStop,
	seats []postgress.ShiftSeatDetails,
	err error,
) {
	logger.LogInfo("Request received in GetShiftDetails", sessionId)

	shift, err = getShiftDetailsById(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the shift error: "+err.Error())
		err = errors.New(constants.Shift_Not_Found)
		return
	}

	stops, seats, err = getShiftRoute(ctx, shiftId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the route error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shift")
		return
	}

	logger.LogInfo("Response returned from GetShiftDetails", sessionId)

	return
}
