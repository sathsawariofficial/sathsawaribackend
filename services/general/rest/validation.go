package general

import (
	"fmt"
	"rideshare/pkgs/constants"
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
