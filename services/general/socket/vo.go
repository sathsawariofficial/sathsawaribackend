package socket

import (
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/utils"
)

func locationSearchResponse(locations []utils.Location) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data:    locations,
	}
}
