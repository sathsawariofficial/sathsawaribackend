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
	"time"

	"github.com/gin-gonic/gin"
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
			status,
			created_at,
			updated_at
		`).
		Limit(pageSize).
		Offset(offset)

	countQuery := query

	err = query.Order("created_at desc").Find(&vehicles).Error
	if err != nil {
		return
	}

	err = countQuery.Count(&totalRows).Error
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

func deleteRole(orgCtx *gin.Context, sessionId, roleId string) (err error) {
	logger.LogInfo("Request received in deleteRole", sessionId)

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
