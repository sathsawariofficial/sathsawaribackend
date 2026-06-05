package general

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
)

func ValidateSMSFCM(request *SMSFCMRequest) error {
	if utils.IsStringEmpty(request.FCM) {
		return fmt.Errorf(constants.Invalid_Data, "fcm")
	}

	if utils.IsStringEmpty(request.App) {
		return fmt.Errorf(constants.Invalid_Data, "app")
	}

	return nil
}

func ValidateApproch(sessionId string, request *ApprochRequest) error {
	logger.LogInfo("Request received in ValidateApproch", sessionId)

	var errMessage string
	if utils.IsStringEmptyWithKey(request.Name, "Name", &errMessage) ||
		utils.IsStringEmptyWithKey(request.Number, "Number", &errMessage) ||
		utils.IsStringEmptyWithKey(request.Email, "Email", &errMessage) ||
		utils.IsStringEmptyWithKey(request.Message, "Message", &errMessage) {
		logger.LogError(sessionId, errMessage)
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	// Length validations
	nameLen := len(request.Name)
	numberLen := len(request.Number)
	emailLen := len(request.Email)
	messageLen := len(request.Message)

	if nameLen < constants.Name_Min_Len || nameLen > constants.Name_Max_Len {
		return fmt.Errorf("value of the name should be between %v and %v", constants.Name_Min_Len, constants.Name_Max_Len)
	}
	if numberLen < constants.MobileNumber_Min_Len || numberLen > constants.MobileNumber_Max_Len {
		return fmt.Errorf("value of the mobilenumber should be between %v and %v", constants.MobileNumber_Min_Len, constants.MobileNumber_Max_Len)
	}
	if emailLen < constants.Email_Min_Len || emailLen > constants.Email_Max_Len {
		return fmt.Errorf("length of the email should be between %v and %v characters", constants.Email_Min_Len, constants.Email_Max_Len)
	}
	if messageLen < constants.Message_Min_Len || messageLen > constants.Message_Max_Len {
		return fmt.Errorf("length of the message should be between %v and %v characters", constants.Message_Min_Len, constants.Message_Max_Len)
	}

	if !utils.IsValidEmail(request.Email) {
		return fmt.Errorf(constants.Invalid_Data, "Email")
	}

	if constants.NewApprochType(request.Type) {
		return fmt.Errorf(constants.Invalid_Data, "Approch type")
	}

	return nil
}
