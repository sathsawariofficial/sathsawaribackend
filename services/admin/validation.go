package admin

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/utils"
	"strings"
)

func ValidateAdminLogin(request *AdminLoginRequest) error {
	var errMessage string
	if utils.IsStringEmptyWithKey(request.Username, "Username", &errMessage) ||
		utils.IsStringEmptyWithKey(request.Password, "Password", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	usernameLen := len(request.Username)
	passwordLen := len(request.Password)

	if !(usernameLen >= constants.Username_Min_Len && usernameLen <= constants.Username_Max_Len) {
		return fmt.Errorf("length of the username should be between %v and %v characters", constants.MobileNumber_Min_Len, constants.MobileNumber_Max_Len)
	}
	if !(passwordLen >= constants.Long_Password_Min_Len && passwordLen <= constants.Long_Password_Max_Len) {
		return fmt.Errorf("length of the password should be between %v and %v characters", constants.Long_Password_Min_Len, constants.Long_Password_Max_Len)
	}

	return nil
}

func ValidatePage(page int) error {
	if page != 0 {
		if page < constants.Page_Min_Value || page > constants.Page_Max_Value {
			return fmt.Errorf("value of the page number should be between %v and %v", constants.Page_Min_Value, constants.Page_Max_Value)
		}
	}

	return nil
}

func ValidateApprochInfoReq(approchType, page string) error {
	if utils.IsStringEmpty(approchType) && !constants.NewApprochType(approchType) {
		return fmt.Errorf(constants.Invalid_Data, "approch type")
	}

	return ValidatePage(utils.ToInt(page))
}

func ValidateBoardcastRequest(request *AdminBroadcastRequest) error {
	var errMessage string
	if utils.IsStringEmptyWithKey(request.Title, "Title", &errMessage) ||
		utils.IsStringEmptyWithKey(request.Message, "Message", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	if !constants.NewNotificationType(request.NotificationType) {
		return fmt.Errorf(constants.Invalid_Data, "notification type")
	}

	if !constants.NewUserType(request.UserType) {
		return fmt.Errorf(constants.Invalid_Data, "user type")
	}

	return nil
}

func ValidateAnnouncementRequest(request *AnnouncementRequest) error {
	var errMessage string
	if utils.IsStringEmptyWithKey(request.Title, "Title", &errMessage) ||
		utils.IsStringEmptyWithKey(request.Message, "Message", &errMessage) ||
		utils.IsStringEmptyWithKey(request.Message, "Type", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	if !constants.NewAnnouncementType(request.Type) {
		return fmt.Errorf(constants.Invalid_Data, "notification type")
	}

	return nil
}

// ValidateAccountStatus checks a moderation status change. Only the two states an
// admin can put an account into are accepted, pending belongs to the otp flow and
// is never something an admin sets by hand.
func ValidateAccountStatus(id string, request *AdminStatusRequest) error {
	if err := utils.ValidateId(id); err != nil {
		return err
	}

	if request.Status != constants.Status_Active && request.Status != constants.Status_InActive {
		return fmt.Errorf(constants.Invalid_Data, "status")
	}

	return nil
}

func ValidateSearchFilters(status string) error {
	if !utils.IsStringEmpty(status) &&
		status != constants.Status_Active &&
		status != constants.Status_InActive &&
		status != constants.Status_PendingApproval {
		return fmt.Errorf(constants.Invalid_Data, "status")
	}

	return nil
}

// ParseAdminShiftFilter reads the same search shape the owner's own shift and shift
// request lists use, so an admin filtering the whole platform gets the identical
// day-of-week, place and time-window rules.
func ParseAdminShiftFilter(status, dayOfWeek, search, startTime, endTime string) (filter adminShiftFilter, err error) {
	if !utils.IsStringEmpty(startTime) {
		if filter.StartTime, err = utils.NormalizeClock(startTime); err != nil {
			return filter, fmt.Errorf(constants.Invalid_Data, "start time, use HH:MM")
		}
	}

	if !utils.IsStringEmpty(endTime) {
		if filter.EndTime, err = utils.NormalizeClock(endTime); err != nil {
			return filter, fmt.Errorf(constants.Invalid_Data, "end time, use HH:MM")
		}
	}

	switch status {
	case "":
		filter.Status = constants.Shift_Status_All
	case constants.Shift_Status_Active, constants.Shift_Status_Completed, constants.Shift_Status_All:
		filter.Status = status
	default:
		return filter, fmt.Errorf(constants.Invalid_Data, "status, use active, completed or all")
	}

	if !utils.IsStringEmpty(dayOfWeek) {
		filter.DayOfWeek = utils.ToInt(dayOfWeek)
		if filter.DayOfWeek < constants.Day_Of_Week_Min_Value || filter.DayOfWeek > constants.Day_Of_Week_Max_Value {
			return filter, fmt.Errorf(constants.Invalid_Data, "day of week, use 1 for Monday to 7 for Sunday")
		}
	}

	filter.Search = strings.TrimSpace(search)
	if len(filter.Search) > constants.General_Max_Len {
		return filter, fmt.Errorf("length of the search should not be more than %v characters", constants.General_Max_Len)
	}

	return filter, nil
}
