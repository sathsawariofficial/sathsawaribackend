package group

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

// creates a fleet owned by the calling driver
func CreateGroupHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in CreateGroupHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request CreateGroupRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Creation_Failed, "group"),
		})
		return
	}

	logger.LogDebug2("Request received in CreateGroupHandler", sessionId, request)

	if err := ValidateCreateGroup(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	groupId, err := CreateGroup(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "create group error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	groupResp := createGroupResp(groupId)

	logger.LogInfo("Response returned from CreateGroupHandler", sessionId)

	ctx.JSON(http.StatusOK, groupResp)
}

// lists the fleets the calling driver owns or belongs to
func GetMyGroupsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetMyGroupsHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	groups, totalRows, err := GetMyGroups(ctx, sessionId, driverId, page)
	if err != nil {
		logger.LogError(sessionId, "get groups error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	groupsResp := groupsResp(groups, driverId, totalRows)

	logger.LogInfo("Response returned from GetMyGroupsHandler", sessionId)

	ctx.JSON(http.StatusOK, groupsResp)
}

// returns a fleet with its driver, vehicle and passenger rosters
func GetGroupDetailsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetGroupDetailsHandler", sessionId)

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

	group, members, vehicles, passengers, err := GetGroupDetails(ctx, sessionId, driverId, groupId)
	if err != nil {
		logger.LogError(sessionId, "get group details error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	detailsResp := groupDetailsResp(group, driverId, members, vehicles, passengers)

	logger.LogInfo("Response returned from GetGroupDetailsHandler", sessionId)

	ctx.JSON(http.StatusOK, detailsResp)
}

// a driver asks to join a fleet, with or without vehicles of their own
func RequestJoinGroupAsDriverHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RequestJoinGroupAsDriverHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request DriverJoinRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Failed_To_Do_Job, "send the request"),
		})
		return
	}

	logger.LogDebug2("Request received in RequestJoinGroupAsDriverHandler", sessionId, request)

	if err := ValidateDriverJoinRequest(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := RequestJoinAsDriver(ctx, sessionId, driverId, request); err != nil {
		logger.LogError(sessionId, "join request error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from RequestJoinGroupAsDriverHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(constants.Request_Received_Successfully))
}

// a passenger asks to join a fleet
func RequestJoinGroupAsPassengerHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RequestJoinGroupAsPassengerHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	var request PassengerJoinRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Failed_To_Do_Job, "send the request"),
		})
		return
	}

	logger.LogDebug2("Request received in RequestJoinGroupAsPassengerHandler", sessionId, request)

	if err := ValidatePassengerJoinRequest(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := RequestJoinAsPassenger(ctx, sessionId, passengerId, request); err != nil {
		logger.LogError(sessionId, "join request error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from RequestJoinGroupAsPassengerHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(constants.Request_Received_Successfully))
}

// lists everybody still waiting at the door of a fleet
func GetGroupRequestsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetGroupRequestsHandler", sessionId)

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

	members, vehicles, passengers, err := GetPendingRequests(ctx, sessionId, driverId, groupId)
	if err != nil {
		logger.LogError(sessionId, "get requests error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	requestsResp := pendingRequestsResp(members, vehicles, passengers)

	logger.LogInfo("Response returned from GetGroupRequestsHandler", sessionId)

	ctx.JSON(http.StatusOK, requestsResp)
}

// settles a whole batch of join requests and removals in one call
func DecideGroupRequestsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DecideGroupRequestsHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request DecideMembershipRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Failed_To_Do_Job, "apply the decisions"),
		})
		return
	}

	logger.LogDebug2("Request received in DecideGroupRequestsHandler", sessionId, request)

	if err := ValidateDecideMembership(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	applied, skipped, err := DecideMembership(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "decide requests error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	decisionResp := decideMembershipResp(applied, skipped)

	logger.LogInfo("Response returned from DecideGroupRequestsHandler", sessionId)

	ctx.JSON(http.StatusOK, decisionResp)
}

// appoints and stands down sub managers in one call
func SetGroupSubManagersHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SetGroupSubManagersHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request SetSubManagersRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "sub managers"),
		})
		return
	}

	logger.LogDebug2("Request received in SetGroupSubManagersHandler", sessionId, request)

	if err := ValidateSetSubManagers(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	applied, skipped, err := SetSubManagers(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "set sub managers error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	subManagersResp := setSubManagersResp(applied, skipped)

	logger.LogInfo("Response returned from SetGroupSubManagersHandler", sessionId)

	ctx.JSON(http.StatusOK, subManagersResp)
}

// closes a fleet down once nothing is scheduled on it any more
func DeleteGroupHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeleteGroupHandler", sessionId)

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

	if err := DeleteGroup(ctx, sessionId, driverId, groupId); err != nil {
		logger.LogError(sessionId, "delete group error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DeleteGroupHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Deleted_Successfully, "Group")))
}

// the standing travel forms of the fleet's passengers, this is the sheet a manager
// builds a shift from
func GetGroupPassengerSchedulesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetGroupPassengerSchedulesHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	groupId := ctx.Query(constants.Group_Key)
	direction := ctx.DefaultQuery(constants.Direction_Key, "")
	dayOfWeek := utils.ToInt(ctx.DefaultQuery(constants.Day_Of_Week_Key, ""))

	if err := ValidateGroupId(groupId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := ValidateScheduleFilters(direction, dayOfWeek); err != nil {
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

	schedules, totalRows, err := GetPassengerSchedules(ctx, sessionId, driverId, groupId, direction, dayOfWeek, page)
	if err != nil {
		logger.LogError(sessionId, "get passenger schedules error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	schedulesResp := passengerSchedulesResp(schedules, totalRows)

	logger.LogInfo("Response returned from GetGroupPassengerSchedulesHandler", sessionId)

	ctx.JSON(http.StatusOK, schedulesResp)
}
