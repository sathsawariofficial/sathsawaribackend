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

func ValidateRoleRequest(request *AdminRoleRequest) error {
	var errMessage string
	if utils.IsStringEmptyWithKey(request.Name, "Name", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	nameLen := len(request.Name)
	if nameLen < constants.Role_Name_Min_Len || nameLen > constants.Role_Name_Max_Len {
		return fmt.Errorf("length of the role name should be between %v and %v characters", constants.Role_Name_Min_Len, constants.Role_Name_Max_Len)
	}

	if len(request.Description) > constants.Description_Max_Len {
		return fmt.Errorf("length of the description should not be more than %v characters", constants.Description_Max_Len)
	}

	return nil
}

func ValidateRoleUpdateRequest(roleId string, request *AdminRoleUpdateRequest) error {
	if err := utils.ValidateId(roleId); err != nil {
		return fmt.Errorf(constants.Invalid_Data, "role id")
	}

	if request.Name == nil && request.Description == nil {
		return fmt.Errorf(constants.Missing_Data, "Name or description")
	}

	if request.Name != nil {
		nameLen := len(*request.Name)
		if nameLen < constants.Role_Name_Min_Len || nameLen > constants.Role_Name_Max_Len {
			return fmt.Errorf("length of the role name should be between %v and %v characters", constants.Role_Name_Min_Len, constants.Role_Name_Max_Len)
		}
	}

	if request.Description != nil && len(*request.Description) > constants.Description_Max_Len {
		return fmt.Errorf("length of the description should not be more than %v characters", constants.Description_Max_Len)
	}

	return nil
}

func ValidatePermissionRequest(request *AdminPermissionRequest) error {
	var errMessage string
	if utils.IsStringEmptyWithKey(request.Code, "Code", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	codeLen := len(request.Code)
	if codeLen < constants.Permission_Code_Min_Len || codeLen > constants.Permission_Code_Max_Len {
		return fmt.Errorf("length of the permission code should be between %v and %v characters", constants.Permission_Code_Min_Len, constants.Permission_Code_Max_Len)
	}

	if len(request.Description) > constants.Description_Max_Len {
		return fmt.Errorf("length of the description should not be more than %v characters", constants.Description_Max_Len)
	}

	return nil
}

func ValidateRolePermissionsRequest(request *AdminRolePermissionsRequest) error {
	if err := utils.ValidateId(request.RoleId); err != nil {
		return fmt.Errorf(constants.Invalid_Data, "role id")
	}

	if len(request.PermissionIds) > constants.Bulk_Request_Max_Len {
		return fmt.Errorf("a role cannot be given more than %v permissions in one call", constants.Bulk_Request_Max_Len)
	}

	// an empty list is allowed on purpose, it strips a role back to no permissions
	for _, permissionId := range request.PermissionIds {
		if err := utils.ValidateId(permissionId); err != nil {
			return fmt.Errorf(constants.Invalid_Data, "permission id")
		}
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
