package group

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
)

// CreateGroup starts a fleet owned by the calling driver. The owner is written in
// as a member holding the owner role, which is what every later permission check
// reads, so the fleet is never left without somebody who can run it.
func CreateGroup(ctx *gin.Context, sessionId, driverId string, request CreateGroupRequest) (groupId string, err error) {
	logger.LogInfo("Request received in CreateGroup", sessionId)

	driver, err := database.GetActiveDriverById(ctx, driverId)
	if err != nil || utils.IsStringEmpty(driver.ID) {
		logger.LogError(sessionId, "get driver error")
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	if _, err = checkVehiclesUsable(ctx, driverId, request.VehicleIds); err != nil {
		logger.LogError(sessionId, "vehicle check error: "+err.Error())
		return
	}

	ownerRoleId, err := getRoleIdByName(ctx, constants.ROLE_GROUP_OWNER)
	if err != nil {
		logger.LogError(sessionId, "failed to get the owner role error: "+err.Error())
		err = errors.New(constants.Role_Not_Found)
		return
	}

	groupId, err = createGroupWithVehicles(ctx, sessionId, driverId, ownerRoleId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to create group error: "+err.Error())
		err = fmt.Errorf(constants.Creation_Failed, "group")
		return
	}

	logger.LogInfo("Response returned from CreateGroup", sessionId)
	logger.LogDebug2("Response returned from CreateGroup", sessionId, groupId)

	return
}

func GetMyGroups(ctx *gin.Context, sessionId, driverId string, page int) (groups []postgress.Group, totalRows int64, err error) {
	logger.LogInfo("Request received in GetMyGroups", sessionId)

	groups, totalRows, err = getGroupsByDriver(ctx, driverId, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get groups error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the groups")
		return
	}

	logger.LogInfo("Response returned from GetMyGroups", sessionId)

	return
}

// GetGroupDetails returns the fleet with its three rosters. Only somebody already
// inside the fleet may look at who else is in it.
func GetGroupDetails(ctx *gin.Context, sessionId, driverId, groupId string) (
	group postgress.Group,
	members []postgress.GroupMemberDetails,
	vehicles []postgress.GroupVehicleDetails,
	passengers []postgress.GroupPassengerDetails,
	err error,
) {
	logger.LogInfo("Request received in GetGroupDetails", sessionId)

	group, err = database.GetGroupById(ctx, groupId)
	if err != nil {
		logger.LogError(sessionId, "failed to get group error: "+err.Error())
		err = errors.New(constants.Group_Not_Found)
		return
	}

	isMember, err := database.IsApprovedGroupMember(ctx, groupId, driverId)
	if err != nil {
		logger.LogError(sessionId, "failed to check the membership error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}
	if !isMember {
		err = errors.New(constants.Not_Group_Member)
		logger.LogError(sessionId, err)
		return
	}

	members, vehicles, passengers, err = getGroupRoster(ctx, sessionId, groupId, nil)
	if err != nil {
		logger.LogError(sessionId, "failed to get the roster error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the group")
		return
	}

	logger.LogInfo("Response returned from GetGroupDetails", sessionId)

	return
}

// RequestJoinAsDriver puts a driver in the queue of a fleet, together with whatever
// vehicles they are offering. Nothing is granted here, the manager still decides.
func RequestJoinAsDriver(ctx *gin.Context, sessionId, driverId string, request DriverJoinRequest) (err error) {
	logger.LogInfo("Request received in RequestJoinAsDriver", sessionId)

	group, err := database.GetGroupById(ctx, request.GroupId)
	if err != nil {
		logger.LogError(sessionId, "failed to get group error: "+err.Error())
		err = errors.New(constants.Group_Not_Found)
		return
	}

	if group.OwnerDriverID == driverId {
		err = errors.New(constants.Already_Group_Member)
		logger.LogError(sessionId, "the owner is already in the group error: "+err.Error())
		return
	}

	if len(request.VehicleIds) > 0 {
		if _, err = checkVehiclesUsable(ctx, driverId, request.VehicleIds); err != nil {
			logger.LogError(sessionId, "vehicle check error: "+err.Error())
			return
		}
	}

	existing, e := getGroupMember(ctx, request.GroupId, driverId)
	if e != nil && !isNotFound(e) {
		logger.LogError(sessionId, "failed to read the membership error: "+e.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	switch existing.Status {
	case constants.Membership_Status_Approved:
		err = errors.New(constants.Already_Group_Member)
		logger.LogError(sessionId, err)
		return
	case constants.Membership_Status_Pending:
		err = errors.New(constants.Request_Already_Exists)
		logger.LogError(sessionId, err)
		return
	}

	if err = saveDriverJoinRequest(ctx, sessionId, driverId, existing, request); err != nil {
		logger.LogError(sessionId, "failed to save the request error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "send the request")
		return
	}

	notifyOwnerOfRequest(ctx, sessionId, group, driverId, constants.User_Driver)

	logger.LogInfo("Response returned from RequestJoinAsDriver", sessionId)

	return
}

// RequestJoinAsPassenger puts a passenger in the queue of a fleet.
func RequestJoinAsPassenger(ctx *gin.Context, sessionId, passengerId string, request PassengerJoinRequest) (err error) {
	logger.LogInfo("Request received in RequestJoinAsPassenger", sessionId)

	group, err := database.GetGroupById(ctx, request.GroupId)
	if err != nil {
		logger.LogError(sessionId, "failed to get group error: "+err.Error())
		err = errors.New(constants.Group_Not_Found)
		return
	}

	passenger, err := database.GetActivePassengerById(ctx, passengerId)
	if err != nil || utils.IsStringEmpty(passenger.ID) {
		logger.LogError(sessionId, "get passenger error")
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	existing, e := getGroupPassenger(ctx, request.GroupId, passengerId)
	if e != nil && !isNotFound(e) {
		logger.LogError(sessionId, "failed to read the membership error: "+e.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	switch existing.Status {
	case constants.Membership_Status_Approved:
		err = errors.New(constants.Already_Group_Member)
		logger.LogError(sessionId, err)
		return
	case constants.Membership_Status_Pending:
		err = errors.New(constants.Request_Already_Exists)
		logger.LogError(sessionId, err)
		return
	}

	if err = savePassengerJoinRequest(ctx, sessionId, passengerId, existing, request.GroupId); err != nil {
		logger.LogError(sessionId, "failed to save the request error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "send the request")
		return
	}

	notifyOwnerOfRequest(ctx, sessionId, group, passengerId, constants.User_Passenger)

	logger.LogInfo("Response returned from RequestJoinAsPassenger", sessionId)

	return
}

func GetPendingRequests(ctx *gin.Context, sessionId, driverId, groupId string) (
	members []postgress.GroupMemberDetails,
	vehicles []postgress.GroupVehicleDetails,
	passengers []postgress.GroupPassengerDetails,
	err error,
) {
	logger.LogInfo("Request received in GetPendingRequests", sessionId)

	if err = requirePermission(ctx, sessionId, groupId, driverId, constants.PERMISSION_GROUP_MANAGE_MEMBERS); err != nil {
		return
	}

	members, vehicles, passengers, err = getGroupRoster(ctx, sessionId, groupId, []string{constants.Membership_Status_Pending})
	if err != nil {
		logger.LogError(sessionId, "failed to get the pending requests error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the requests")
		return
	}

	logger.LogInfo("Response returned from GetPendingRequests", sessionId)

	return
}

// DecideMembership settles a whole batch of join requests at once. Every line is
// judged on its own and the caller is told exactly which ones did not take, the way
// a bulk ride cancellation reports what it skipped.
func DecideMembership(ctx *gin.Context, sessionId, driverId string, request DecideMembershipRequest) (applied, skipped []DecisionResult, err error) {
	logger.LogInfo("Request received in DecideMembership", sessionId)

	group, err := database.GetGroupById(ctx, request.GroupId)
	if err != nil {
		logger.LogError(sessionId, "failed to get group error: "+err.Error())
		err = errors.New(constants.Group_Not_Found)
		return
	}

	if err = requirePermission(ctx, sessionId, request.GroupId, driverId, constants.PERMISSION_GROUP_MANAGE_MEMBERS); err != nil {
		return
	}

	// deciding on a vehicle is a separate permission, a role may be allowed to let
	// people in without being allowed to accept their vehicles
	needsVehiclePermission := false
	memberIds := map[string][]string{}
	for _, decision := range request.Decisions {
		memberIds[decision.MemberType] = append(memberIds[decision.MemberType], decision.MemberId)
		if decision.MemberType == constants.Member_Type_Vehicle {
			needsVehiclePermission = true
		}
	}

	if needsVehiclePermission {
		if err = requirePermission(ctx, sessionId, request.GroupId, driverId, constants.PERMISSION_GROUP_MANAGE_VEHICLES); err != nil {
			return
		}
	}

	members, vehicles, passengers, err := loadDecisionTargets(ctx, request.GroupId, memberIds)
	if err != nil {
		logger.LogError(sessionId, "failed to load the requests error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	buckets := map[string]map[string][]string{}
	notifyDrivers := map[string]string{}
	notifyPassengers := map[string]string{}

	for _, decision := range request.Decisions {
		result := DecisionResult{
			MemberType: decision.MemberType,
			MemberId:   decision.MemberId,
			Action:     decision.Action,
		}

		var currentStatus, subjectId string
		var found bool

		switch decision.MemberType {
		case constants.Member_Type_Driver:
			row, ok := members[decision.MemberId]
			found, currentStatus, subjectId = ok, row.Status, row.DriverID

			// the owner holds the fleet together, they can never be thrown out of it
			if ok && row.DriverID == group.OwnerDriverID {
				result.Reason = constants.Operation_Not_Permitted
				skipped = append(skipped, result)
				continue
			}
		case constants.Member_Type_Vehicle:
			row, ok := vehicles[decision.MemberId]
			found, currentStatus, subjectId = ok, row.Status, row.DriverID
		case constants.Member_Type_Passenger:
			row, ok := passengers[decision.MemberId]
			found, currentStatus, subjectId = ok, row.Status, row.PassengerID
		}

		if !found {
			result.Reason = fmt.Sprintf(constants.Not_Found, "request")
			skipped = append(skipped, result)
			continue
		}

		newStatus, reason := nextMembershipStatus(decision.Action, currentStatus)
		if reason != "" {
			result.Reason = reason
			skipped = append(skipped, result)
			continue
		}

		if buckets[decision.MemberType] == nil {
			buckets[decision.MemberType] = map[string][]string{}
		}
		buckets[decision.MemberType][newStatus] = append(buckets[decision.MemberType][newStatus], decision.MemberId)

		if decision.MemberType == constants.Member_Type_Passenger {
			notifyPassengers[subjectId] = newStatus
		} else {
			notifyDrivers[subjectId] = newStatus
		}

		result.Applied = true
		applied = append(applied, result)
	}

	if len(applied) == 0 {
		logger.LogInfo("Response returned from DecideMembership", sessionId)
		return
	}

	if err = applyMembershipDecisions(ctx, sessionId, driverId, buckets); err != nil {
		logger.LogError(sessionId, "failed to apply the decisions error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "apply the decisions")
		applied = nil
		return
	}

	notifyMembershipDecisions(ctx, sessionId, group, notifyDrivers, notifyPassengers)

	logger.LogInfo("Response returned from DecideMembership", sessionId)
	logger.LogDebug2("Response returned from DecideMembership", sessionId, fmt.Sprintf("applied: %d, skipped: %d", len(applied), len(skipped)))

	return
}

// SetSubManagers appoints and stands down sub managers in one call. A sub manager
// can only be somebody who is already an approved driver of that same fleet, the
// role is a promotion from within and never a way in from outside.
func SetSubManagers(ctx *gin.Context, sessionId, driverId string, request SetSubManagersRequest) (applied, skipped []SubManagerResult, err error) {
	logger.LogInfo("Request received in SetSubManagers", sessionId)

	group, err := database.GetGroupById(ctx, request.GroupId)
	if err != nil {
		logger.LogError(sessionId, "failed to get group error: "+err.Error())
		err = errors.New(constants.Group_Not_Found)
		return
	}

	if err = requirePermission(ctx, sessionId, request.GroupId, driverId, constants.PERMISSION_GROUP_MANAGE_SUBMANAGERS); err != nil {
		return
	}

	subManagerRoleId, err := getRoleIdByName(ctx, constants.ROLE_GROUP_SUBMANAGER)
	if err != nil {
		logger.LogError(sessionId, "failed to get the sub manager role error: "+err.Error())
		err = errors.New(constants.Role_Not_Found)
		return
	}

	lookup := append(append([]string{}, request.Promote...), request.Demote...)
	members, err := getApprovedMembersByDriverIds(ctx, request.GroupId, lookup)
	if err != nil {
		logger.LogError(sessionId, "failed to read the memberships error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	promote := []string{}
	demote := []string{}

	for _, targetId := range request.Promote {
		result := SubManagerResult{DriverId: targetId, Action: constants.Membership_Action_Approve}

		if _, ok := members[targetId]; !ok {
			// this is the rule the fleet owner asked for: no outsider can be handed
			// the sub manager role, they have to be in the fleet first
			result.Reason = constants.Not_Group_Member
			skipped = append(skipped, result)
			continue
		}

		if targetId == group.OwnerDriverID {
			result.Reason = constants.Operation_Not_Permitted
			skipped = append(skipped, result)
			continue
		}

		promote = append(promote, targetId)
		result.Applied = true
		applied = append(applied, result)
	}

	for _, targetId := range request.Demote {
		result := SubManagerResult{DriverId: targetId, Action: constants.Membership_Action_Remove}

		// standing the owner down would leave the fleet with nobody in charge
		if targetId == group.OwnerDriverID {
			result.Reason = constants.Operation_Not_Permitted
			skipped = append(skipped, result)
			continue
		}

		if _, ok := members[targetId]; !ok {
			result.Reason = constants.Not_Group_Member
			skipped = append(skipped, result)
			continue
		}

		demote = append(demote, targetId)
		result.Applied = true
		applied = append(applied, result)
	}

	if len(promote) == 0 && len(demote) == 0 {
		logger.LogInfo("Response returned from SetSubManagers", sessionId)
		return
	}

	if err = applySubManagerChanges(ctx, sessionId, request.GroupId, subManagerRoleId, promote, demote); err != nil {
		logger.LogError(sessionId, "failed to change the sub managers error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "sub managers")
		applied = nil
		return
	}

	notifySubManagerChanges(ctx, sessionId, group, promote, demote)

	logger.LogInfo("Response returned from SetSubManagers", sessionId)

	return
}

// DeleteGroup closes a fleet down. It refuses while shifts are still ahead, since
// passengers and drivers are counting on those.
func DeleteGroup(ctx *gin.Context, sessionId, driverId, groupId string) (err error) {
	logger.LogInfo("Request received in DeleteGroup", sessionId)

	if _, err = database.GetGroupById(ctx, groupId); err != nil {
		logger.LogError(sessionId, "failed to get group error: "+err.Error())
		err = errors.New(constants.Group_Not_Found)
		return
	}

	if err = requirePermission(ctx, sessionId, groupId, driverId, constants.PERMISSION_GROUP_DELETE); err != nil {
		return
	}

	count, err := countFutureShifts(ctx, groupId)
	if err != nil {
		logger.LogError(sessionId, "failed to count the shifts error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}
	if count > 0 {
		err = fmt.Errorf("this group still has %d upcoming shift(s), cancel them before deleting it", count)
		logger.LogError(sessionId, err)
		return
	}

	if err = deactivateGroup(ctx, sessionId, groupId); err != nil {
		logger.LogError(sessionId, "failed to delete group error: "+err.Error())
		err = fmt.Errorf(constants.DELETE_Failed, "group")
		return
	}

	logger.LogInfo("Response returned from DeleteGroup", sessionId)

	return
}
