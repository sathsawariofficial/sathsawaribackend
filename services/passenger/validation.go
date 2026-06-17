package passenger

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
)

func ValidateBookSeat(sessionId string, request BookSeatRequest) error {
	if !utils.PKValidation(request.RideId) {
		return fmt.Errorf(constants.Invalid_Data, "ride id")
	}

	nameLen := len(request.Name)
	if !(nameLen >= constants.Name_Min_Len && nameLen <= constants.Name_Max_Len) {
		return fmt.Errorf("length of the name should be between %v and %v characters", constants.Name_Min_Len, constants.Name_Max_Len)
	}

	if err := utils.IsValidMobileNumber(request.MobileNumber); err != nil {
		return err
	}

	if request.Seats == 0 {
		return fmt.Errorf(constants.Missing_Data, "seats")
	}

	var errMessage string
	if utils.IsStringEmptyWithKey(request.Code, "Code", &errMessage) {
		logger.LogError(sessionId, errMessage)
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	return nil
}
