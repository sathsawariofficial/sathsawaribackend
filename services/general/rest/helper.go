package general

import (
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func mapSMSFcmRequest(request SMSFCMRequest) *postgress.SMSFCM {
	return &postgress.SMSFCM{
		ID:  utils.GenerateUUID(),
		App: request.App,
		FCM: request.FCM,
	}
}
