package passenger

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
)

func ValidateBookSeatDemand(sessionId string, request BookSeatDemandRequest) error {
	if !utils.PKValidation(request.RideId) {
		return fmt.Errorf(constants.Invalid_Data, "ride id")
	}

	nameLen := len(request.Name)
	numberLen := len(request.Mobilenumber)

	var errMessage string
	if utils.IsStringEmptyWithKey(request.Name, "Name", &errMessage) ||
		utils.IsStringEmptyWithKey(request.Message, "Message", &errMessage) {
		logger.LogError(sessionId, errMessage)
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	if nameLen < constants.Name_Min_Len || nameLen > constants.Name_Max_Len {
		return fmt.Errorf("value of the name should be between %v and %v", constants.Name_Min_Len, constants.Name_Max_Len)
	}

	if numberLen < constants.MobileNumber_Min_Len || numberLen > constants.MobileNumber_Max_Len {
		return fmt.Errorf("value of the mobilenumber should be between %v and %v", constants.MobileNumber_Min_Len, constants.MobileNumber_Max_Len)
	}

	if request.NumberOfSeats == 0 {
		return fmt.Errorf(constants.Invalid_Data, "number of seats")
	}

	return nil
}
