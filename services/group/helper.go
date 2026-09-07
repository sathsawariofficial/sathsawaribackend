package group

import (
	"context"
	"errors"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func withTimeout(orgCtx *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
}

// onConflictGroupVehicle turns a re-offered vehicle into an update of the existing
// row, the (group_id, vehicle_id) pair is unique so a plain insert would fail for
// a vehicle that had been offered and turned down before.
func onConflictGroupVehicle() clause.OnConflict {
	return clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "vehicle_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"driver_id", "status", "decided_by", "decided_at", "updated_at"}),
	}
}

// getRoleIdByName resolves a seeded role to its id. Roles are looked up by their
// name rather than a hardcoded id so an admin can reshape what the role may do
// without any code change.
func getRoleIdByName(orgCtx *gin.Context, name string) (roleId string, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	var role postgress.Role
	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("name = ?", name).
		First(&role).Error

	return role.ID, err
}

// getOwnedVehicles loads every vehicle of the given ids that really belongs to the
// driver, in one query rather than one lookup per vehicle.
func getOwnedVehicles(orgCtx *gin.Context, driverId string, vehicleIds []string) (vehicles []postgress.Vehicle, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("id IN ?", vehicleIds).
		Where("driver_id = ?", driverId).
		Where("status = ?", constants.Status_Active).
		Find(&vehicles).Error

	return
}

// checkVehiclesUsable makes sure every vehicle asked for belongs to the driver and
// carries a seat count, a vehicle of unknown size cannot be seated safely.
func checkVehiclesUsable(orgCtx *gin.Context, driverId string, vehicleIds []string) (vehicles []postgress.Vehicle, err error) {
	vehicles, err = getOwnedVehicles(orgCtx, driverId, vehicleIds)
	if err != nil {
		return
	}

	if len(vehicles) != len(vehicleIds) {
		err = fmt.Errorf(constants.Not_Found, "vehicle")
		return
	}

	for _, vehicle := range vehicles {
		if vehicle.NumberOfSeats <= 0 {
			err = fmt.Errorf("%s: %s", constants.Vehicle_Seats_Missing, vehicle.VehicleNumber)
			return
		}
	}

	return
}

func getGroupMember(orgCtx *gin.Context, groupId, driverId string) (member postgress.GroupMember, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("group_id = ?", groupId).
		Where("driver_id = ?", driverId).
		First(&member).Error

	return
}

func getGroupPassenger(orgCtx *gin.Context, groupId, passengerId string) (groupPassenger postgress.GroupPassenger, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("group_id = ?", groupId).
		Where("passenger_id = ?", passengerId).
		First(&groupPassenger).Error

	return
}

// createGroupWithVehicles builds the whole fleet in one transaction: the group, the
// owner's own membership carrying the owner role, and every vehicle it starts with.
func createGroupWithVehicles(orgCtx *gin.Context, sessionId, driverId, ownerRoleId string, request CreateGroupRequest) (groupId string, err error) {
	logger.LogInfo("Request received in createGroupWithVehicles", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()
	if tx.Error != nil {
		return groupId, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	group := postgress.Group{
		ID:            database.GenerateUUID(),
		Name:          request.Name,
		Description:   request.Description,
		OwnerDriverID: driverId,
		Status:        constants.Status_Active,
	}

	if err = tx.Create(&group).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}
	groupId = group.ID

	member := postgress.GroupMember{
		ID:        database.GenerateUUID(),
		GroupID:   group.ID,
		DriverID:  driverId,
		RoleID:    ownerRoleId,
		JoinType:  constants.Join_Type_Both,
		Status:    constants.Membership_Status_Approved,
		DecidedBy: driverId,
		DecidedAt: &now,
	}

	if err = tx.Create(&member).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	groupVehicles := make([]postgress.GroupVehicle, 0, len(request.VehicleIds))
	for _, vehicleId := range request.VehicleIds {
		groupVehicles = append(groupVehicles, postgress.GroupVehicle{
			ID:        database.GenerateUUID(),
			GroupID:   group.ID,
			VehicleID: vehicleId,
			DriverID:  driverId,
			Status:    constants.Membership_Status_Approved,
			DecidedBy: driverId,
			DecidedAt: &now,
		})
	}

	if err = tx.Create(&groupVehicles).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return
	}

	err = tx.Commit().Error

	logger.LogInfo("Response returned from createGroupWithVehicles", sessionId)

	return
}

// getGroupsByDriver returns every group the driver owns or has been let into, in
// one query, paginated.
func getGroupsByDriver(orgCtx *gin.Context, driverId string, page int) (groups []postgress.Group, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).
		Table("groups").
		Where("groups.status = ?", constants.Status_Active).
		Where(`
			groups.owner_driver_id = ?
			OR EXISTS (
				SELECT 1 FROM group_members
				WHERE group_members.group_id = groups.id
				  AND group_members.driver_id = ?
				  AND group_members.status = ?
			)
		`, driverId, driverId, constants.Membership_Status_Approved)

	if err = query.Count(&totalRows).Error; err != nil {
		return
	}

	err = query.
		Order("groups.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&groups).Error

	return
}

// getGroupRoster loads the three membership lists of a group with one joined query
// each. statuses narrows it down, an empty list means every status.
func getGroupRoster(orgCtx *gin.Context, sessionId, groupId string, statuses []string) (
	members []postgress.GroupMemberDetails,
	vehicles []postgress.GroupVehicleDetails,
	passengers []postgress.GroupPassengerDetails,
	err error,
) {
	logger.LogInfo("Request received in getGroupRoster", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	memberQuery := db.Table("group_members").
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
		Where("group_members.group_id = ?", groupId)

	vehicleQuery := db.Table("group_vehicles").
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
		Where("group_vehicles.group_id = ?", groupId)

	passengerQuery := db.Table("group_passengers").
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
		Where("group_passengers.group_id = ?", groupId)

	if len(statuses) > 0 {
		memberQuery = memberQuery.Where("group_members.status IN ?", statuses)
		vehicleQuery = vehicleQuery.Where("group_vehicles.status IN ?", statuses)
		passengerQuery = passengerQuery.Where("group_passengers.status IN ?", statuses)
	}

	if err = memberQuery.Order("group_members.created_at ASC").Find(&members).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	if err = vehicleQuery.Order("group_vehicles.created_at ASC").Find(&vehicles).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	if err = passengerQuery.Order("group_passengers.created_at ASC").Find(&passengers).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	logger.LogInfo("Response returned from getGroupRoster", sessionId)

	return
}

// saveDriverJoinRequest records a driver asking to join, together with the vehicles
// they are offering. The membership row is unique per (group, driver), so a driver
// who was turned down before or who left is moved back to pending rather than
// inserted a second time.
func saveDriverJoinRequest(orgCtx *gin.Context, sessionId, driverId string, existing postgress.GroupMember, request DriverJoinRequest) (err error) {
	logger.LogInfo("Request received in saveDriverJoinRequest", sessionId)

	ctx, cancel := withTimeout(orgCtx)
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

	if existing.ID != "" {
		if err = tx.Model(&postgress.GroupMember{}).
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"join_type":  request.JoinType,
				"status":     constants.Membership_Status_Pending,
				"role_id":    "",
				"decided_by": "",
				"decided_at": nil,
			}).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}
	} else {
		member := postgress.GroupMember{
			ID:       database.GenerateUUID(),
			GroupID:  request.GroupId,
			DriverID: driverId,
			JoinType: request.JoinType,
			Status:   constants.Membership_Status_Pending,
		}

		if err = tx.Create(&member).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}
	}

	if len(request.VehicleIds) > 0 {
		groupVehicles := make([]postgress.GroupVehicle, 0, len(request.VehicleIds))
		for _, vehicleId := range request.VehicleIds {
			groupVehicles = append(groupVehicles, postgress.GroupVehicle{
				ID:        database.GenerateUUID(),
				GroupID:   request.GroupId,
				VehicleID: vehicleId,
				DriverID:  driverId,
				Status:    constants.Membership_Status_Pending,
			})
		}

		// a vehicle that was offered before comes back to pending instead of
		// colliding with the unique (group, vehicle) index
		if err = tx.Clauses(onConflictGroupVehicle()).Create(&groupVehicles).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}
	}

	err = tx.Commit().Error

	logger.LogInfo("Response returned from saveDriverJoinRequest", sessionId)

	return
}

func savePassengerJoinRequest(orgCtx *gin.Context, sessionId, passengerId string, existing postgress.GroupPassenger, groupId string) (err error) {
	logger.LogInfo("Request received in savePassengerJoinRequest", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	if existing.ID != "" {
		err = db.Model(&postgress.GroupPassenger{}).
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"status":     constants.Membership_Status_Pending,
				"decided_by": "",
				"decided_at": nil,
			}).Error
	} else {
		groupPassenger := postgress.GroupPassenger{
			ID:          database.GenerateUUID(),
			GroupID:     groupId,
			PassengerID: passengerId,
			Status:      constants.Membership_Status_Pending,
		}

		err = db.Create(&groupPassenger).Error
	}

	if err != nil {
		logger.LogError(sessionId, err)
	}

	logger.LogInfo("Response returned from savePassengerJoinRequest", sessionId)

	return
}

// loadDecisionTargets reads every membership row a bulk decision refers to, one
// query per member type rather than one per decision.
func loadDecisionTargets(orgCtx *gin.Context, groupId string, memberIds map[string][]string) (
	members map[string]postgress.GroupMember,
	vehicles map[string]postgress.GroupVehicle,
	passengers map[string]postgress.GroupPassenger,
	err error,
) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	members = map[string]postgress.GroupMember{}
	vehicles = map[string]postgress.GroupVehicle{}
	passengers = map[string]postgress.GroupPassenger{}

	if ids := memberIds[constants.Member_Type_Driver]; len(ids) > 0 {
		var rows []postgress.GroupMember
		if err = db.Where("group_id = ?", groupId).Where("id IN ?", ids).Find(&rows).Error; err != nil {
			return
		}
		for _, row := range rows {
			members[row.ID] = row
		}
	}

	if ids := memberIds[constants.Member_Type_Vehicle]; len(ids) > 0 {
		var rows []postgress.GroupVehicle
		if err = db.Where("group_id = ?", groupId).Where("id IN ?", ids).Find(&rows).Error; err != nil {
			return
		}
		for _, row := range rows {
			vehicles[row.ID] = row
		}
	}

	if ids := memberIds[constants.Member_Type_Passenger]; len(ids) > 0 {
		var rows []postgress.GroupPassenger
		if err = db.Where("group_id = ?", groupId).Where("id IN ?", ids).Find(&rows).Error; err != nil {
			return
		}
		for _, row := range rows {
			passengers[row.ID] = row
		}
	}

	return
}

// applyMembershipDecisions writes a whole batch of decisions in one transaction,
// with one update statement per (member type, resulting status) bucket instead of
// one statement per decision.
func applyMembershipDecisions(orgCtx *gin.Context, sessionId, deciderId string, buckets map[string]map[string][]string) (err error) {
	logger.LogInfo("Request received in applyMembershipDecisions", sessionId)

	ctx, cancel := withTimeout(orgCtx)
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

	now := time.Now()

	for memberType, statusBuckets := range buckets {
		for status, ids := range statusBuckets {
			if len(ids) == 0 {
				continue
			}

			updates := map[string]interface{}{
				"status":     status,
				"decided_by": deciderId,
				"decided_at": &now,
			}

			var model interface{}
			switch memberType {
			case constants.Member_Type_Driver:
				model = &postgress.GroupMember{}
				// somebody who is no longer in the fleet keeps no role
				if status != constants.Membership_Status_Approved {
					updates["role_id"] = ""
				}
			case constants.Member_Type_Vehicle:
				model = &postgress.GroupVehicle{}
			case constants.Member_Type_Passenger:
				model = &postgress.GroupPassenger{}
			default:
				continue
			}

			if err = tx.Model(model).Where("id IN ?", ids).Updates(updates).Error; err != nil {
				tx.Rollback()
				logger.LogError(sessionId, err)
				return
			}
		}
	}

	err = tx.Commit().Error

	logger.LogInfo("Response returned from applyMembershipDecisions", sessionId)

	return
}

// getApprovedMembersByDriverIds loads the memberships of a batch of drivers in one
// query, used to check who is even eligible to become a sub manager.
func getApprovedMembersByDriverIds(orgCtx *gin.Context, groupId string, driverIds []string) (members map[string]postgress.GroupMember, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	var rows []postgress.GroupMember
	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("group_id = ?", groupId).
		Where("driver_id IN ?", driverIds).
		Where("status = ?", constants.Membership_Status_Approved).
		Find(&rows).Error

	members = map[string]postgress.GroupMember{}
	for _, row := range rows {
		members[row.DriverID] = row
	}

	return
}

// applySubManagerChanges promotes and demotes in one transaction, one update per
// side rather than one per driver.
func applySubManagerChanges(orgCtx *gin.Context, sessionId, groupId, subManagerRoleId string, promote, demote []string) (err error) {
	logger.LogInfo("Request received in applySubManagerChanges", sessionId)

	ctx, cancel := withTimeout(orgCtx)
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

	if len(promote) > 0 {
		if err = tx.Model(&postgress.GroupMember{}).
			Where("group_id = ?", groupId).
			Where("driver_id IN ?", promote).
			Update("role_id", subManagerRoleId).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}
	}

	if len(demote) > 0 {
		if err = tx.Model(&postgress.GroupMember{}).
			Where("group_id = ?", groupId).
			Where("driver_id IN ?", demote).
			Update("role_id", "").Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return
		}
	}

	err = tx.Commit().Error

	logger.LogInfo("Response returned from applySubManagerChanges", sessionId)

	return
}

// countFutureShifts reports how many shifts of a group are still ahead of us, a
// fleet people are still counting on must not be deleted underneath them.
func countFutureShifts(orgCtx *gin.Context, groupId string) (count int64, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Shift{}).
		Where("group_id = ?", groupId).
		Where("is_active = ?", true).
		Where("start_datetime > ?", time.Now().Format(constants.DateTimeLayout)).
		Count(&count).Error

	return
}

func deactivateGroup(orgCtx *gin.Context, sessionId, groupId string) (err error) {
	logger.LogInfo("Request received in deactivateGroup", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Group{}).
		Where("id = ?", groupId).
		Update("status", constants.Status_InActive).Error

	logger.LogInfo("Response returned from deactivateGroup", sessionId)

	return
}

func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// requirePermission is the single gate every managing action goes through. It asks
// the database what the caller's role may do rather than testing a role name, so a
// brand new role an admin invents works here without a code change.
func requirePermission(ctx *gin.Context, sessionId, groupId, driverId, permission string) error {
	hasPermission, err := database.HasGroupPermission(ctx, groupId, driverId, permission)
	if err != nil {
		logger.LogError(sessionId, "failed to check the permission error: "+err.Error())
		return errors.New(constants.Unknown_Error)
	}

	if !hasPermission {
		err = errors.New(constants.Operation_Not_Permitted)
		logger.LogError(sessionId, fmt.Sprintf("driver lacks %s error: %s", permission, err.Error()))
		return err
	}

	return nil
}

// nextMembershipStatus works out what one decision should turn a row into, and why
// it cannot be applied when it cannot. Approving or rejecting only makes sense on
// somebody still waiting, removing only on somebody already inside.
func nextMembershipStatus(action, currentStatus string) (newStatus, reason string) {
	switch action {
	case constants.Membership_Action_Approve:
		if currentStatus != constants.Membership_Status_Pending {
			return "", fmt.Sprintf("this request is already %s", currentStatus)
		}
		return constants.Membership_Status_Approved, ""
	case constants.Membership_Action_Reject:
		if currentStatus != constants.Membership_Status_Pending {
			return "", fmt.Sprintf("this request is already %s", currentStatus)
		}
		return constants.Membership_Status_Rejected, ""
	case constants.Membership_Action_Remove:
		if currentStatus != constants.Membership_Status_Approved {
			return "", "only an approved member can be removed"
		}
		return constants.Membership_Status_Removed, ""
	}

	return "", fmt.Sprintf(constants.Invalid_Data, "action")
}

// notifyOwnerOfRequest tells the fleet owner somebody is waiting at the door.
func notifyOwnerOfRequest(ctx *gin.Context, sessionId string, group postgress.Group, requesterId string, userType int) {
	var requesterName string

	if userType == constants.User_Passenger {
		passenger, err := database.GetPassengerById(ctx, requesterId)
		if err != nil {
			logger.LogError(sessionId, err)
			return
		}
		requesterName = passenger.PassengerName
	} else {
		driver, err := database.GetDriverById(ctx, requesterId)
		if err != nil {
			logger.LogError(sessionId, err)
			return
		}
		requesterName = driver.DriverName
	}

	message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_GROUP_JOIN_REQUEST, requesterName, group.Name)

	utils.SendUserNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_GROUP_REQUEST, group.OwnerDriverID, constants.User_Driver,
		constants.NOTIFICATION_TITLE_GROUP_REQUEST, message, map[string]string{
			constants.NOTIFICATION_KEY_GROUP_ID: group.ID,
		})
}

// notifyMembershipDecisions tells everybody in a settled batch what was decided.
func notifyMembershipDecisions(ctx *gin.Context, sessionId string, group postgress.Group, drivers, passengers map[string]string) {
	data := map[string]string{constants.NOTIFICATION_KEY_GROUP_ID: group.ID}

	for driverId, status := range drivers {
		utils.SendUserNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_GROUP_DECISION, driverId, constants.User_Driver,
			constants.NOTIFICATION_TITLE_GROUP_DECISION, membershipDecisionMessage(status, group.Name), data)
	}

	for passengerId, status := range passengers {
		utils.SendUserNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_GROUP_DECISION, passengerId, constants.User_Passenger,
			constants.NOTIFICATION_TITLE_GROUP_DECISION, membershipDecisionMessage(status, group.Name), data)
	}
}

func membershipDecisionMessage(status, groupName string) string {
	switch status {
	case constants.Membership_Status_Approved:
		return fmt.Sprintf(constants.NOTIFICATION_MESSAGE_GROUP_APPROVED, groupName)
	case constants.Membership_Status_Rejected:
		return fmt.Sprintf(constants.NOTIFICATION_MESSAGE_GROUP_REJECTED, groupName)
	default:
		return fmt.Sprintf(constants.NOTIFICATION_MESSAGE_GROUP_REMOVED, groupName)
	}
}

func notifySubManagerChanges(ctx *gin.Context, sessionId string, group postgress.Group, promote, demote []string) {
	data := map[string]string{constants.NOTIFICATION_KEY_GROUP_ID: group.ID}

	for _, driverId := range promote {
		utils.SendUserNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_GROUP_DECISION, driverId, constants.User_Driver,
			constants.NOTIFICATION_TITLE_GROUP_DECISION, fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SUBMANAGER_ADDED, group.Name), data)
	}

	for _, driverId := range demote {
		utils.SendUserNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_GROUP_DECISION, driverId, constants.User_Driver,
			constants.NOTIFICATION_TITLE_GROUP_DECISION, fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SUBMANAGER_REMOVED, group.Name), data)
	}
}
