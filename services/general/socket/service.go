package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
)

func LocationSearch(ctx context.Context, sessionId, requestId string, reqType string, requestBytes []byte) any {
	logger.LogInfo("Request received in LocationSearch", sessionId)

	var request LocationSearchRequest
	err := json.Unmarshal(requestBytes, &request)
	if err != nil {
		logger.LogError(sessionId, err)
		return utils.GeneralSocketResp(sessionId, http.StatusBadRequest, fmt.Sprintf(constants.Invalid_Data, "Request"))
	}

	locations, err := utils.SearchLocations(request.Place, request.UserLat, request.UserLng)
	if err != nil {
		logger.LogError(sessionId, err)

		locErr := database.SaveMissingLocation(ctx, postgress.MissingLocations{
			DeviceId: request.DeviceId,
			Place:    request.Place,
			UserLat:  request.UserLat,
			UserLng:  request.UserLng,
		})
		if locErr != nil {
			logger.LogError(sessionId, locErr)
		}

		return utils.GeneralSocketResp(sessionId, http.StatusBadRequest, fmt.Sprintf(constants.Failed_To_Do_Job, "get location"))
	}

	if len(locations) == 0 {
		locErr := database.SaveMissingLocation(ctx, postgress.MissingLocations{
			DeviceId: request.DeviceId,
			Place:    request.Place,
			UserLat:  request.UserLat,
			UserLng:  request.UserLng,
		})
		if locErr != nil {
			logger.LogError(sessionId, locErr)
		}
	}

	resp := locationSearchResponse(locations)

	logger.LogInfo("Response returned from LocationSearch", sessionId)
	logger.LogDebug("Response returned from LocationSearch", sessionId, resp)

	return resp
}
