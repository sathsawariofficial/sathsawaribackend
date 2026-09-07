package group

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/utils"
)

func ValidateCreateGroup(request *CreateGroupRequest) error {
	var errMessage string
	if utils.IsStringEmptyWithKey(request.Name, "Name", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	nameLen := len(request.Name)
	if nameLen < constants.Group_Name_Min_Len || nameLen > constants.Group_Name_Max_Len {
		return fmt.Errorf("length of the group name should be between %v and %v characters", constants.Group_Name_Min_Len, constants.Group_Name_Max_Len)
	}

	if len(request.Description) > constants.Description_Max_Len {
		return fmt.Errorf("length of the description should not be more than %v characters", constants.Description_Max_Len)
	}

	// a group with no vehicle cannot run a single shift, so one is required up front
	if len(request.VehicleIds) == 0 {
		return fmt.Errorf(constants.Missing_Data, "At least one vehicle")
	}

	return validateIdList(request.VehicleIds, "vehicle id")
}

func ValidateGroupId(groupId string) error {
	if !utils.PKValidation(groupId) {
		return fmt.Errorf(constants.Invalid_Data, "group id")
	}

	return nil
}

func ValidateDriverJoinRequest(request *DriverJoinRequest) error {
	if err := ValidateGroupId(request.GroupId); err != nil {
		return err
	}

	switch request.JoinType {
	case constants.Join_Type_Driver_Only:
		// somebody offering only to drive brings no vehicle of their own
		if len(request.VehicleIds) > 0 {
			return fmt.Errorf(constants.Invalid_Data, "vehicles, a driver only request carries none")
		}
	case constants.Join_Type_Vehicle_Only, constants.Join_Type_Both:
		if len(request.VehicleIds) == 0 {
			return fmt.Errorf(constants.Missing_Data, "At least one vehicle")
		}
	default:
		return fmt.Errorf(constants.Invalid_Data, "join type")
	}

	return validateIdList(request.VehicleIds, "vehicle id")
}

func ValidatePassengerJoinRequest(request *PassengerJoinRequest) error {
	return ValidateGroupId(request.GroupId)
}

func ValidateDecideMembership(request *DecideMembershipRequest) error {
	if err := ValidateGroupId(request.GroupId); err != nil {
		return err
	}

	if len(request.Decisions) == 0 {
		return fmt.Errorf(constants.Missing_Data, "Decisions")
	}

	if len(request.Decisions) > constants.Bulk_Request_Max_Len {
		return fmt.Errorf("no more than %v decisions can be sent in one call", constants.Bulk_Request_Max_Len)
	}

	for _, decision := range request.Decisions {
		if !utils.PKValidation(decision.MemberId) {
			return fmt.Errorf(constants.Invalid_Data, "member id")
		}

		switch decision.MemberType {
		case constants.Member_Type_Driver, constants.Member_Type_Vehicle, constants.Member_Type_Passenger:
		default:
			return fmt.Errorf(constants.Invalid_Data, "member type")
		}

		switch decision.Action {
		case constants.Membership_Action_Approve, constants.Membership_Action_Reject, constants.Membership_Action_Remove:
		default:
			return fmt.Errorf(constants.Invalid_Data, "action")
		}
	}

	return nil
}

func ValidateSetSubManagers(request *SetSubManagersRequest) error {
	if err := ValidateGroupId(request.GroupId); err != nil {
		return err
	}

	if len(request.Promote) == 0 && len(request.Demote) == 0 {
		return fmt.Errorf(constants.Missing_Data, "Drivers to promote or demote")
	}

	if len(request.Promote)+len(request.Demote) > constants.Bulk_Request_Max_Len {
		return fmt.Errorf("no more than %v drivers can be changed in one call", constants.Bulk_Request_Max_Len)
	}

	if err := validateIdList(request.Promote, "driver id"); err != nil {
		return err
	}

	return validateIdList(request.Demote, "driver id")
}

func validateIdList(ids []string, key string) error {
	if len(ids) > constants.Bulk_Request_Max_Len {
		return fmt.Errorf("no more than %v entries can be sent in one call", constants.Bulk_Request_Max_Len)
	}

	seen := map[string]bool{}
	for _, id := range ids {
		if !utils.PKValidation(id) {
			return fmt.Errorf(constants.Invalid_Data, key)
		}

		if seen[id] {
			return fmt.Errorf(constants.Invalid_Data, key+", it is repeated")
		}
		seen[id] = true
	}

	return nil
}

// ValidateScheduleFilters checks the optional narrowing a manager can put on the
// travel forms, an empty filter simply means every day or every direction.
func ValidateScheduleFilters(direction string, dayOfWeek int) error {
	if !utils.IsStringEmpty(direction) &&
		direction != constants.Shift_Direction_Pickup &&
		direction != constants.Shift_Direction_Drop {
		return fmt.Errorf(constants.Invalid_Data, "direction")
	}

	if dayOfWeek != 0 && (dayOfWeek < constants.Day_Of_Week_Min_Value || dayOfWeek > constants.Day_Of_Week_Max_Value) {
		return fmt.Errorf(constants.Invalid_Data, "day of week")
	}

	return nil
}
