package general

import (
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func mapSMSFcmRequest(request SMSFCMRequest) *postgress.SMSFCM {
	return &postgress.SMSFCM{
		ID:      utils.GenerateUUID(),
		App:     request.App,
		FCM:     request.FCM,
		APPHash: utils.ChoiseMaker(request.AppHash, constants.DEFAULT_APP_HASH),
	}
}
