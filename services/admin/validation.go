package admin

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/utils"
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
