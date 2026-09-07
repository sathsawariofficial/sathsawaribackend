package admin

import (
	"context"
	"errors"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getAdmin(orgCtx *gin.Context, mobile string) (admin postgress.Admin, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where(`username = ?`, mobile).Find(&admin).Error
	return
}

func loginTokenCreation(sessionId string, request AdminLoginRequest, admin postgress.Admin) (token string, err error) {
	if err = utils.ComparePassword(admin.Password, request.Password); err != nil {
		logger.LogError(sessionId, "invalid password error")
		err = errors.New(constants.Invalid_Password)
		return
	}

	excryptedAdminId, err := utils.EncryptAES(sessionId, admin.ID)
	if err != nil {
		logger.LogError(sessionId, "failed to set session error: "+err.Error())
		err = errors.New(constants.Login_Failed)
		return
	}

	err = redis.DeleteRedisValue(database.DatabaseConn.RedisConn, excryptedAdminId)
	if err != nil {
		logger.LogError(sessionId, "trying to delete session before new login error: "+err.Error())
	}

	token, err = utils.CreateJWT(sessionId, map[string]string{
		"adminId":   excryptedAdminId,
		"tokenType": constants.ADMIN_TOKEN,
	})
	if err != nil {
		logger.LogError(sessionId, "failed to created jwt token error: "+err.Error())
		err = errors.New(constants.Login_Failed)
		return
	}

	err = redis.SetRedisValueTTL(database.DatabaseConn.RedisConn, excryptedAdminId, token, time.Duration(configuration.ConfigurationData.Auth.ExpirationTime)*time.Second)
	if err != nil {
		logger.LogError(sessionId, "session deleted error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, "login the admin")
		return
	}

	return
}

func getAllRides(orgCtx *gin.Context, page int) (rides []postgress.RideDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Table("rides").
		Select(`
			rides.id,
			rides.driver_id,
			drivers.driver_name,
			drivers.driver_mobile,
			vehicles.vehicle_number,
			vehicles.vehicle_info,
			rides.start_datetime,
			rides.estimated_end_datetime,
			rides.number_of_seats,
			rides.seats_taken,
			rides.start_location,
			rides.end_location,
			rides.fare,
			rides.route_details,
			rides.is_active,
			rides.created_at,
			rides.updated_at
		`).
		Joins("JOIN drivers ON rides.driver_id = drivers.id").
		Joins("JOIN vehicles ON rides.vehicle_id = vehicles.id").
		Limit(pageSize).
		Offset(offset)

	err = query.Order("created_at desc").Find(&rides).Error
	err = query.Count(&totalRows).Error

	return
}

func getAllVehicles(orgCtx *gin.Context, page int) (vehicles []postgress.Vehicle, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).
		Table("vehicles").
		Select(`
			id,
			driver_id,
			vehicle_number,
			vehicle_info,
			number_of_seats,
			has_ac,
			has_heating,
			status,
			created_at,
			updated_at
		`)

	// counted on a clean clone, the old count ran against the limited query and so
	// could never report more than one page
	countQuery := query.Session(&gorm.Session{})
	if err = countQuery.Count(&totalRows).Error; err != nil {
		return
	}

	err = query.
		Order("created_at desc").
		Limit(pageSize).
		Offset(offset).
		Find(&vehicles).Error

	return
}

func getAllDriversWithVehicles(orgCtx *gin.Context, page int) (drivers []postgress.Driver, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	// Step 1: Count active drivers
	err = db.Model(&postgress.Driver{}).
		Where("status = ?", constants.Status_Active).
		Count(&totalRows).Error
	if err != nil {
		return
	}

	// Step 2: Fetch drivers with vehicles
	err = db.
		Preload("Vehicles").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&drivers).Error

	if err != nil {
		return
	}

	return
}

func createAdminBroadcastRequest(req AdminBroadcastRequest) postgress.BroadcastNotificationRequests {
	return postgress.BroadcastNotificationRequests{
		ID:               database.GenerateUUID(),
		UserType:         req.UserType,
		Title:            req.Title,
		Message:          req.Message,
		NotificationType: req.NotificationType,
	}
}

func createAnnouncementRequest(req AnnouncementRequest) postgress.AnnouncementRequests {
	return postgress.AnnouncementRequests{
		ID:      database.GenerateUUID(),
		Title:   req.Title,
		Message: req.Message,
		Type:    string(req.Type),
	}
}

func createAdminBroadcast(orgCtx *gin.Context, sessionId string, request AdminBroadcastRequest) error {
	logger.LogInfo("Request received in createAdminBroadcast", sessionId)

	broadcastReq := createAdminBroadcastRequest(request)
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)
	err := db.Create(&broadcastReq).Error
	if err != nil {
		logger.LogError(sessionId, err)
		return err
	}

	logger.LogInfo("Response returned from createAdminBroadcast", sessionId)
	return nil
}

func createAnnouncement(orgCtx *gin.Context, sessionId string, request AnnouncementRequest) error {
	logger.LogInfo("Request received in createAnnouncement", sessionId)

	broadcastReq := createAnnouncementRequest(request)
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)
	err := db.Create(&broadcastReq).Error
	if err != nil {
		logger.LogError(sessionId, err)
		return err
	}

	logger.LogInfo("Response returned from createAnnouncement", sessionId)
	return nil
}

func getApprochRequests(orgCtx *gin.Context, sessionId, approchType string, page int) (approchs []postgress.ApprochInfo, totalRows int64, err error) {
	logger.LogInfo("Request received in getApprochRequests", sessionId)
	logger.LogDebug("Request received in getApprochRequests", sessionId, fmt.Sprintf("approch type: %s, page: %d", approchType, page))

	pageSize := configuration.ConfigurationData.PageSize
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	// Base query (shared for count + fetch)
	baseQuery := db.Model(&postgress.ApprochInfo{})

	if !utils.IsStringEmpty(approchType) {
		baseQuery = baseQuery.Where("type = ?", approchType)
	}

	// Step 1: Count total rows
	err = baseQuery.Count(&totalRows).Error
	if err != nil {
		logger.LogError(sessionId, err)
		return
	}

	// Step 2: Fetch paginated data
	query := baseQuery.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset)

	err = query.Find(&approchs).Error
	if err != nil {
		logger.LogError(sessionId, err)
		return
	}

	logger.LogInfo("Response returned from getApprochRequests", sessionId)
	logger.LogDebug("Response returned from getApprochRequests", sessionId, approchs)

	return
}

func createAccouncement(orgCtx *gin.Context, sessionId string, request AdminBroadcastRequest) error {
	logger.LogInfo("Request received in createAccouncement", sessionId)

	broadcastReq := createAdminBroadcastRequest(request)
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)
	err := db.Create(&broadcastReq).Error
	if err != nil {
		logger.LogError(sessionId, err)
		return err
	}

	logger.LogInfo("Response returned from createAccouncement", sessionId)
	return nil
}

func createRole(orgCtx *gin.Context, sessionId string, request AdminRoleRequest) (role postgress.Role, err error) {
	logger.LogInfo("Request received in createRole", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	role = postgress.Role{
		ID:          database.GenerateUUID(),
		Name:        request.Name,
		Description: request.Description,
		IsSystem:    false,
	}

	err = database.DatabaseConn.Postgres.WithContext(ctx).Create(&role).Error

	logger.LogInfo("Response returned from createRole", sessionId)

	return
}

func getRoleById(orgCtx *gin.Context, roleId string) (role postgress.Role, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where(`id = ?`, roleId).First(&role).Error
	return
}

func countRolesByName(orgCtx *gin.Context, name, excludeRoleId string) (count int64, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Role{}).
		Where("name = ?", name)

	if !utils.IsStringEmpty(excludeRoleId) {
		query = query.Where("id <> ?", excludeRoleId)
	}

	err = query.Count(&count).Error
	return
}

// getRolesPage reads one page of roles together with their permissions in exactly
// two queries, the page of roles and then a single joined query for every
// permission of every role on that page, which is then stitched in memory. Asking
// per role would be an N+1.
func getRolesPage(orgCtx *gin.Context, sessionId string, page int) (roles []postgress.Role, permissions map[string][]adminRolePermissionRow, totalRows int64, err error) {
	logger.LogInfo("Request received in getRolesPage", sessionId)

	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	if err = db.Model(&postgress.Role{}).Count(&totalRows).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	if err = db.Model(&postgress.Role{}).
		Order("created_at ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&roles).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	permissions = map[string][]adminRolePermissionRow{}
	if len(roles) == 0 {
		return
	}

	roleIds := make([]string, 0, len(roles))
	for _, role := range roles {
		roleIds = append(roleIds, role.ID)
	}

	var rows []adminRolePermissionRow
	if err = db.Table("role_permissions").
		Select(`
			role_permissions.role_id,
			permissions.id,
			permissions.code,
			permissions.description,
			permissions.is_system,
			permissions.created_at
		`).
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id IN ?", roleIds).
		Order("permissions.code ASC").
		Find(&rows).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	for _, row := range rows {
		permissions[row.RoleID] = append(permissions[row.RoleID], row)
	}

	logger.LogInfo("Response returned from getRolesPage", sessionId)

	return
}

func updateRole(orgCtx *gin.Context, sessionId, roleId string, updates map[string]interface{}) (err error) {
	logger.LogInfo("Request received in updateRole", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Role{}).
		Where("id = ?", roleId).
		Updates(updates).Error

	logger.LogInfo("Response returned from updateRole", sessionId)

	return
}

// countMembersUsingRole reports how many group memberships still point at a role,
// a role that is still handed out to somebody must not disappear under them.
func countMembersUsingRole(orgCtx *gin.Context, roleId string) (count int64, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.GroupMember{}).
		Where("role_id = ?", roleId).
		Count(&count).Error

	return
}

func deleteRole(orgCtx *gin.Context, sessionId, adminId string, role postgress.Role) (err error) {
	logger.LogInfo("Request received in deleteRole", sessionId)

	roleId := role.ID

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	////////// ARCHIVE THE ROLE //////////
	// who was allowed to do what is worth being able to answer later, so the role is
	// kept with the permission codes it held at this moment, inline, so the record
	// still reads correctly even after those permissions themselves change
	var codes []string
	if err = tx.Table("role_permissions").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleId).
		Pluck("permissions.code", &codes).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	if err = tx.Create(&postgress.DELRole{
		ID:              role.ID,
		Name:            role.Name,
		Description:     role.Description,
		PermissionCodes: codes,
		UpdateBy:        adminId,
		CreatedAt:       role.CreatedAt,
		UpdatedAt:       time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	if err = tx.Where("role_id = ?", roleId).Delete(&postgress.RolePermission{}).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	if err = tx.Where("id = ?", roleId).Delete(&postgress.Role{}).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	err = tx.Commit().Error

	logger.LogInfo("Response returned from deleteRole", sessionId)

	return
}

func createPermission(orgCtx *gin.Context, sessionId string, request AdminPermissionRequest) (permission postgress.Permission, err error) {
	logger.LogInfo("Request received in createPermission", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	permission = postgress.Permission{
		ID:          database.GenerateUUID(),
		Code:        request.Code,
		Description: request.Description,
		IsSystem:    false,
	}

	err = database.DatabaseConn.Postgres.WithContext(ctx).Create(&permission).Error

	logger.LogInfo("Response returned from createPermission", sessionId)

	return
}

func countPermissionsByCode(orgCtx *gin.Context, code string) (count int64, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Permission{}).
		Where("code = ?", code).
		Count(&count).Error

	return
}

func getPermissionsPage(orgCtx *gin.Context, sessionId string, page int) (permissions []postgress.Permission, totalRows int64, err error) {
	logger.LogInfo("Request received in getPermissionsPage", sessionId)

	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	if err = db.Model(&postgress.Permission{}).Count(&totalRows).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	err = db.Model(&postgress.Permission{}).
		Order("code ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&permissions).Error

	logger.LogInfo("Response returned from getPermissionsPage", sessionId)

	return
}

// countPermissionsByIds validates a whole batch of permission ids with one query
// rather than one lookup per id.
func countPermissionsByIds(orgCtx *gin.Context, permissionIds []string) (count int64, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Permission{}).
		Where("id IN ?", permissionIds).
		Count(&count).Error

	return
}

// replaceRolePermissions swaps the whole permission set of a role in one
// transaction, so a role is never seen holding half of its new set.
func replaceRolePermissions(orgCtx *gin.Context, sessionId, roleId string, permissionIds []string) (err error) {
	logger.LogInfo("Request received in replaceRolePermissions", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Where("role_id = ?", roleId).Delete(&postgress.RolePermission{}).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	if len(permissionIds) > 0 {
		rolePermissions := make([]postgress.RolePermission, 0, len(permissionIds))
		for _, permissionId := range permissionIds {
			rolePermissions = append(rolePermissions, postgress.RolePermission{
				ID:           database.GenerateUUID(),
				RoleID:       roleId,
				PermissionID: permissionId,
			})
		}

		if err = tx.Create(&rolePermissions).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}
	}

	err = tx.Commit().Error

	logger.LogInfo("Response returned from replaceRolePermissions", sessionId)

	return
}

// getPlatformOverview counts the whole product in a single round trip rather than a
// dozen, since every number on the admin's first screen is just a count.
func getPlatformOverview(orgCtx *gin.Context, sessionId string) (overview postgress.PlatformOverview, err error) {
	logger.LogInfo("Request received in getPlatformOverview", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	now := time.Now().Format(constants.DateTimeLayout)

	err = database.DatabaseConn.Postgres.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM drivers)                                            AS total_drivers,
			(SELECT COUNT(*) FROM drivers WHERE status = ?)                           AS active_drivers,
			(SELECT COUNT(*) FROM passengers)                                         AS total_passengers,
			(SELECT COUNT(*) FROM passengers WHERE status = ?)                        AS active_passengers,
			(SELECT COUNT(*) FROM vehicles)                                           AS total_vehicles,
			(SELECT COUNT(*) FROM vehicles WHERE number_of_seats > 0)                 AS seated_vehicles,
			(SELECT COUNT(*) FROM groups)                                             AS total_groups,
			(SELECT COUNT(*) FROM groups WHERE status = ?)                            AS active_groups,
			(SELECT COUNT(*) FROM shifts)                                             AS total_shifts,
			(SELECT COUNT(*) FROM shifts WHERE is_active = true AND start_datetime > ?) AS upcoming_shifts,
			(SELECT COUNT(*) FROM rides)                                              AS total_rides,
			(SELECT COUNT(*) FROM rides WHERE is_active = true)                       AS active_rides,
			(
				(SELECT COUNT(*) FROM group_members WHERE status = ?) +
				(SELECT COUNT(*) FROM group_vehicles WHERE status = ?) +
				(SELECT COUNT(*) FROM group_passengers WHERE status = ?)
			)                                                                          AS pending_requests
	`,
		constants.Status_Active,
		constants.Status_Active,
		constants.Status_Active,
		now,
		constants.Membership_Status_Pending,
		constants.Membership_Status_Pending,
		constants.Membership_Status_Pending,
	).Scan(&overview).Error

	logger.LogInfo("Response returned from getPlatformOverview", sessionId)

	return
}

// getAllPassengers lists passenger accounts, optionally narrowed by status or by a
// search across the name and the mobile number.
func getAllPassengers(orgCtx *gin.Context, sessionId, search, status string, page int) (passengers []postgress.Passenger, totalRows int64, err error) {
	logger.LogInfo("Request received in getAllPassengers", sessionId)

	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Model(&postgress.Passenger{})

	if !utils.IsStringEmpty(status) {
		query = query.Where("status = ?", status)
	}

	if !utils.IsStringEmpty(search) {
		term := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("passenger_name ILIKE ? OR passenger_mobile ILIKE ?", term, term)
	}

	countQuery := query.Session(&gorm.Session{})
	if err = countQuery.Count(&totalRows).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	err = query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&passengers).Error

	logger.LogInfo("Response returned from getAllPassengers", sessionId)

	return
}

func getPassengerById(orgCtx *gin.Context, passengerId string) (passenger postgress.Passenger, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where("id = ?", passengerId).First(&passenger).Error
	return
}

// getPassengerContext loads the fleets a passenger belongs to and their standing
// travel form, which together explain what that account is actually doing.
func getPassengerContext(orgCtx *gin.Context, passengerId string) (
	groups []postgress.GroupPassengerDetails,
	preferences []postgress.PassengerLocationPreference,
	err error,
) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	if err = db.Table("group_passengers").
		Select(`
			group_passengers.id,
			group_passengers.group_id,
			group_passengers.passenger_id,
			groups.name AS passenger_name,
			passengers.passenger_mobile,
			passengers.gender,
			group_passengers.status,
			group_passengers.created_at
		`).
		Joins("JOIN groups ON groups.id = group_passengers.group_id").
		Joins("JOIN passengers ON passengers.id = group_passengers.passenger_id").
		Where("group_passengers.passenger_id = ?", passengerId).
		Order("group_passengers.created_at DESC").
		Find(&groups).Error; err != nil {
		return
	}

	err = db.Where("passenger_id = ?", passengerId).
		Order("day_of_week ASC, direction ASC").
		Find(&preferences).Error

	return
}

func updatePassengerStatus(orgCtx *gin.Context, sessionId, passengerId, status, updatedBy string) (err error) {
	logger.LogInfo("Request received in updatePassengerStatus", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Passenger{}).
		Where("id = ?", passengerId).
		Updates(map[string]interface{}{
			"status":    status,
			"update_by": updatedBy,
		}).Error

	logger.LogInfo("Response returned from updatePassengerStatus", sessionId)

	return
}

func updateDriverStatus(orgCtx *gin.Context, sessionId, driverId, status, updatedBy string) (err error) {
	logger.LogInfo("Request received in updateDriverStatus", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Driver{}).
		Where("id = ?", driverId).
		Updates(map[string]interface{}{
			"status":    status,
			"update_by": updatedBy,
		}).Error

	logger.LogInfo("Response returned from updateDriverStatus", sessionId)

	return
}

// getAllGroups lists every fleet on the platform with the size of each, counted with
// subselects so the page costs one query rather than one per fleet.
func getAllGroups(orgCtx *gin.Context, sessionId, search, status string, page int) (groups []postgress.AdminGroupOverview, totalRows int64, err error) {
	logger.LogInfo("Request received in getAllGroups", sessionId)

	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	base := database.DatabaseConn.Postgres.WithContext(ctx).
		Table("groups").
		Joins("JOIN drivers ON drivers.id = groups.owner_driver_id")

	if !utils.IsStringEmpty(status) {
		base = base.Where("groups.status = ?", status)
	}

	if !utils.IsStringEmpty(search) {
		base = base.Where("groups.name ILIKE ?", "%"+strings.TrimSpace(search)+"%")
	}

	countQuery := base.Session(&gorm.Session{})
	if err = countQuery.Count(&totalRows).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	err = base.
		Select(`
			groups.id,
			groups.name,
			groups.description,
			groups.status,
			groups.owner_driver_id,
			drivers.driver_name AS owner_name,
			drivers.driver_mobile AS owner_mobile,
			(SELECT COUNT(*) FROM group_members WHERE group_members.group_id = groups.id AND group_members.status = ?) AS member_count,
			(SELECT COUNT(*) FROM group_vehicles WHERE group_vehicles.group_id = groups.id AND group_vehicles.status = ?) AS vehicle_count,
			(SELECT COUNT(*) FROM group_passengers WHERE group_passengers.group_id = groups.id AND group_passengers.status = ?) AS passenger_count,
			(SELECT COUNT(*) FROM shifts WHERE shifts.group_id = groups.id AND shifts.is_active = true) AS shift_count,
			groups.created_at
		`,
			constants.Membership_Status_Approved,
			constants.Membership_Status_Approved,
			constants.Membership_Status_Approved,
		).
		Order("groups.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&groups).Error

	logger.LogInfo("Response returned from getAllGroups", sessionId)

	return
}

func getGroupById(orgCtx *gin.Context, groupId string) (group postgress.Group, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where("id = ?", groupId).First(&group).Error
	return
}

// getGroupRosters is the admin's read of a fleet's three membership lists. Unlike
// the manager's own view it is not filtered by status, an admin looking into a
// complaint needs to see who was turned away too.
func getGroupRosters(orgCtx *gin.Context, sessionId, groupId string) (
	members []postgress.GroupMemberDetails,
	vehicles []postgress.GroupVehicleDetails,
	passengers []postgress.GroupPassengerDetails,
	err error,
) {
	logger.LogInfo("Request received in getGroupRosters", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	if err = db.Table("group_members").
		Select(`
			group_members.id,
			group_members.group_id,
			group_members.driver_id,
			drivers.driver_name,
			drivers.driver_mobile,
			drivers.rating,
			group_members.role_id,
			roles.name AS role_name,
			group_members.join_type,
			group_members.status,
			group_members.created_at
		`).
		Joins("JOIN drivers ON drivers.id = group_members.driver_id").
		Joins("LEFT JOIN roles ON roles.id = group_members.role_id").
		Where("group_members.group_id = ?", groupId).
		Order("group_members.created_at ASC").
		Find(&members).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	if err = db.Table("group_vehicles").
		Select(`
			group_vehicles.id,
			group_vehicles.group_id,
			group_vehicles.vehicle_id,
			vehicles.vehicle_number,
			vehicles.vehicle_info,
			vehicles.number_of_seats,
			vehicles.has_ac,
			vehicles.has_heating,
			group_vehicles.driver_id,
			drivers.driver_name,
			drivers.driver_mobile,
			group_vehicles.status,
			group_vehicles.created_at
		`).
		Joins("JOIN vehicles ON vehicles.id = group_vehicles.vehicle_id").
		Joins("JOIN drivers ON drivers.id = group_vehicles.driver_id").
		Where("group_vehicles.group_id = ?", groupId).
		Order("group_vehicles.created_at ASC").
		Find(&vehicles).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	err = db.Table("group_passengers").
		Select(`
			group_passengers.id,
			group_passengers.group_id,
			group_passengers.passenger_id,
			passengers.passenger_name,
			passengers.passenger_mobile,
			passengers.gender,
			group_passengers.status,
			group_passengers.created_at
		`).
		Joins("JOIN passengers ON passengers.id = group_passengers.passenger_id").
		Where("group_passengers.group_id = ?", groupId).
		Order("group_passengers.created_at ASC").
		Find(&passengers).Error

	logger.LogInfo("Response returned from getGroupRosters", sessionId)

	return
}

func updateGroupStatus(orgCtx *gin.Context, sessionId, groupId, status string) (err error) {
	logger.LogInfo("Request received in updateGroupStatus", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Group{}).
		Where("id = ?", groupId).
		Update("status", status).Error

	logger.LogInfo("Response returned from updateGroupStatus", sessionId)

	return
}

// getAllShifts is the platform wide shift list, filterable the way an admin chasing
// a complaint would want it.
func getAllShifts(orgCtx *gin.Context, sessionId, groupId, driverId, direction, startTime, endTime, status string, page int) (shifts []postgress.ShiftDetails, totalRows int64, err error) {
	logger.LogInfo("Request received in getAllShifts", sessionId)

	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := adminShiftQuery(database.DatabaseConn.Postgres.WithContext(ctx))

	if !utils.IsStringEmpty(groupId) {
		query = query.Where("shifts.group_id = ?", groupId)
	}

	if !utils.IsStringEmpty(driverId) {
		query = query.Where("shifts.driver_id = ?", driverId)
	}

	if !utils.IsStringEmpty(direction) {
		query = query.Where("shifts.direction = ?", direction)
	}

	if !utils.IsStringEmpty(startTime) {
		query = query.Where("shifts.start_datetime >= ?", startTime)
	}

	if !utils.IsStringEmpty(endTime) {
		query = query.Where("shifts.estimated_end_datetime <= ?", endTime)
	}

	if strings.EqualFold(status, constants.Ride_Status_Active) {
		query = query.Where("shifts.is_active = ?", true)
	} else if strings.EqualFold(status, constants.Ride_Status_InActive) {
		query = query.Where("shifts.is_active = ?", false)
	}

	countQuery := query.Session(&gorm.Session{})
	if err = countQuery.Count(&totalRows).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	err = query.
		Order("shifts.start_datetime DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&shifts).Error

	logger.LogInfo("Response returned from getAllShifts", sessionId)

	return
}

// adminShiftQuery joins in the fleet, the vehicle, the driver of the day and the
// manager who built the shift, which is everything an admin needs to see at once.
func adminShiftQuery(db *gorm.DB) *gorm.DB {
	return db.Table("shifts").
		Select(`
			shifts.id,
			shifts.group_id,
			groups.name AS group_name,
			shifts.vehicle_id,
			vehicles.vehicle_number,
			vehicles.vehicle_info,
			vehicles.has_ac,
			vehicles.has_heating,
			shifts.driver_id,
			drivers.driver_name,
			drivers.driver_mobile,
			drivers.rating,
			shifts.direction,
			shifts.start_datetime,
			shifts.estimated_end_datetime,
			shifts.start_location,
			shifts.end_location,
			shifts.number_of_seats,
			shifts.seats_taken,
			shifts.route_details,
			shifts.created_by_driver_id,
			creators.driver_name AS created_by_name,
			creators.driver_mobile AS created_by_mobile,
			shifts.is_active,
			shifts.created_at,
			shifts.updated_at
		`).
		Joins("JOIN groups ON groups.id = shifts.group_id").
		Joins("JOIN vehicles ON vehicles.id = shifts.vehicle_id").
		Joins("JOIN drivers ON drivers.id = shifts.driver_id").
		Joins("JOIN drivers AS creators ON creators.id = shifts.created_by_driver_id")
}

func getShiftDetailsById(orgCtx *gin.Context, shiftId string) (shift postgress.ShiftDetails, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = adminShiftQuery(database.DatabaseConn.Postgres.WithContext(ctx)).
		Where("shifts.id = ?", shiftId).
		First(&shift).Error

	return
}

// getShiftRoute loads a shift's stops in order and its seats with the passenger and
// stop already joined on, so a whole trip is read in two queries.
func getShiftRoute(orgCtx *gin.Context, shiftId string) (stops []postgress.ShiftStop, seats []postgress.ShiftSeatDetails, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	if err = db.Where("shift_id = ?", shiftId).Order("sequence_number ASC").Find(&stops).Error; err != nil {
		return
	}

	err = db.Table("shift_seats").
		Select(`
			shift_seats.id,
			shift_seats.shift_id,
			shift_seats.seat_number,
			shift_seats.gender,
			shift_seats.status,
			shift_seats.passenger_id,
			passengers.passenger_name,
			passengers.passenger_mobile,
			shift_seats.stop_id,
			shift_stops.sequence_number,
			shift_stops.location,
			shift_stops.lat,
			shift_stops.lng,
			shift_stops.scheduled_time
		`).
		Joins("LEFT JOIN passengers ON passengers.id = shift_seats.passenger_id").
		Joins("LEFT JOIN shift_stops ON shift_stops.id = shift_seats.stop_id").
		Where("shift_seats.shift_id = ?", shiftId).
		Order("shift_seats.seat_number ASC").
		Find(&seats).Error

	return
}
