package group

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func createGroupResp(groupId string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Created_Successfully, "Group"),
		Data: CreateGroupResponse{
			GroupId: groupId,
		},
	}
}

func mapGroupDetails(group postgress.Group, driverId string) GroupDetails {
	return GroupDetails{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		OwnerId:     group.OwnerDriverID,
		Status:      group.Status,
		IsOwner:     group.OwnerDriverID == driverId,
		CreatedAt:   group.CreatedAt,
	}
}

func groupsResp(groups []postgress.Group, driverId string, totalRows int64) utils.APIResponse {
	groupDetails := []GroupDetails{}

	for _, group := range groups {
		groupDetails = append(groupDetails, mapGroupDetails(group, driverId))
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: GroupsResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Groups:     groupDetails,
		},
	}
}

func mapMemberDetails(members []postgress.GroupMemberDetails) []MemberDetails {
	memberDetails := []MemberDetails{}

	for _, member := range members {
		memberDetails = append(memberDetails, MemberDetails{
			ID:           member.ID,
			DriverId:     member.DriverID,
			DriverName:   member.DriverName,
			DriverMobile: member.DriverMobile,
			Rating:       member.Rating,
			RoleName:     member.RoleName,
			JoinType:     member.JoinType,
			Status:       member.Status,
			CreatedAt:    member.CreatedAt,
		})
	}

	return memberDetails
}

func mapVehicleDetails(vehicles []postgress.GroupVehicleDetails) []VehicleDetails {
	vehicleDetails := []VehicleDetails{}

	for _, vehicle := range vehicles {
		vehicleDetails = append(vehicleDetails, VehicleDetails{
			ID:            vehicle.ID,
			VehicleId:     vehicle.VehicleID,
			VehicleNumber: vehicle.VehicleNumber,
			VehicleInfo:   vehicle.VehicleInfo,
			NumberOfSeats: vehicle.NumberOfSeats,
			HasAC:         vehicle.HasAC,
			HasHeating:    vehicle.HasHeating,
			DriverId:      vehicle.DriverID,
			DriverName:    vehicle.DriverName,
			DriverMobile:  vehicle.DriverMobile,
			Status:        vehicle.Status,
			CreatedAt:     vehicle.CreatedAt,
		})
	}

	return vehicleDetails
}

func mapPassengerDetails(passengers []postgress.GroupPassengerDetails) []PassengerDetails {
	passengerDetails := []PassengerDetails{}

	for _, passenger := range passengers {
		passengerDetails = append(passengerDetails, PassengerDetails{
			ID:              passenger.ID,
			PassengerId:     passenger.PassengerID,
			PassengerName:   passenger.PassengerName,
			PassengerMobile: passenger.PassengerMobile,
			Gender:          passenger.Gender,
			Status:          passenger.Status,
			CreatedAt:       passenger.CreatedAt,
		})
	}

	return passengerDetails
}

func groupDetailsResp(
	group postgress.Group,
	driverId string,
	members []postgress.GroupMemberDetails,
	vehicles []postgress.GroupVehicleDetails,
	passengers []postgress.GroupPassengerDetails,
) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: GroupDetailsResponse{
			Group:      mapGroupDetails(group, driverId),
			Members:    mapMemberDetails(members),
			Vehicles:   mapVehicleDetails(vehicles),
			Passengers: mapPassengerDetails(passengers),
		},
	}
}

func pendingRequestsResp(
	members []postgress.GroupMemberDetails,
	vehicles []postgress.GroupVehicleDetails,
	passengers []postgress.GroupPassengerDetails,
) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: PendingRequestsResponse{
			Members:    mapMemberDetails(members),
			Vehicles:   mapVehicleDetails(vehicles),
			Passengers: mapPassengerDetails(passengers),
		},
	}
}

func decideMembershipResp(applied, skipped []DecisionResult) utils.APIResponse {
	if applied == nil {
		applied = []DecisionResult{}
	}
	if skipped == nil {
		skipped = []DecisionResult{}
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Updated_Successfully, "Requests"),
		Data: DecideMembershipResponse{
			Applied: applied,
			Skipped: skipped,
		},
	}
}

func setSubManagersResp(applied, skipped []SubManagerResult) utils.APIResponse {
	if applied == nil {
		applied = []SubManagerResult{}
	}
	if skipped == nil {
		skipped = []SubManagerResult{}
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Updated_Successfully, "Sub managers"),
		Data: SetSubManagersResponse{
			Applied: applied,
			Skipped: skipped,
		},
	}
}

func passengerSchedulesResp(schedules []postgress.GroupPassengerScheduleDetails, totalRows int64) utils.APIResponse {
	entries := []PassengerScheduleEntry{}

	for _, schedule := range schedules {
		entries = append(entries, PassengerScheduleEntry{
			PassengerId:     schedule.PassengerID,
			PassengerName:   schedule.PassengerName,
			PassengerMobile: schedule.PassengerMobile,
			Gender:          schedule.Gender,
			DayOfWeek:       schedule.DayOfWeek,
			Direction:       schedule.Direction,
			IsEnabled:       schedule.IsEnabled,
			Location:        schedule.Location,
			Lat:             schedule.Lat,
			Lng:             schedule.Lng,
			ScheduledTime:   schedule.ScheduledTime,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: PassengerSchedulesResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Schedules:  entries,
		},
	}
}

func groupSearchResp(groups []postgress.GroupSearchDetails, totalRows int64) utils.APIResponse {
	summaries := []GroupSummary{}

	for _, group := range groups {
		summaries = append(summaries, GroupSummary{
			ID:             group.ID,
			Name:           group.Name,
			Description:    group.Description,
			OwnerId:        group.OwnerDriverID,
			OwnerName:      group.OwnerName,
			VehicleCount:   group.VehicleCount,
			PassengerCount: group.PassengerCount,
			MyStatus:       group.MyStatus,
			CreatedAt:      group.CreatedAt,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: GroupSearchResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Groups:     summaries,
		},
	}
}
