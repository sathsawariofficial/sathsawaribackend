package passenger

import (
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func bookSeatResponse(msg, uuid string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: msg,
		Data: BookSeatResponse{
			BookingId: uuid,
		},
	}
}

func filteredRideRequestsResp(requests []postgress.RideRequest, totalRows int) utils.APIResponse {
	var rideRequests []RideRequestDetails
	for _, request := range requests {
		openUrl := utils.CreateOpenRideLink(constants.LIKE_TYPE_RIDE_REQUEST_URL, utils.GenerateShortCode(request.ID))

		rideRequests = append(rideRequests, RideRequestDetails{
			ID:                   request.ID,
			ContactNumber:        request.ContactNumber,
			StartDatetime:        request.StartDatetime,
			EstimatedEndDatetime: request.EstimatedEndDatetime,
			NumberOfSeats:        request.NumberOfSeats,
			StartLocation:        request.StartLocation,
			EndLocation:          request.EndLocation,
			RouteDetails:         request.RouteDetails,
			OpenURL:              openUrl,
		})
	}

	rideDetailsResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: RidesRequestsDetailsResponse{
			TotalPages: utils.CalculatePagesize(int64(totalRows)),
			Rides:      rideRequests,
		},
	}

	return rideDetailsResp
}

func rideRequestDetailsResp(rideRequest postgress.RideRequest) utils.APIResponse {
	openUrl := utils.CreateOpenRideLink(constants.LIKE_TYPE_RIDE_REQUEST_URL, utils.GenerateShortCode(rideRequest.ID))

	rideDetailsResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: RideRequestDetails{
			ID:                   rideRequest.ID,
			ContactNumber:        rideRequest.ContactNumber,
			StartDatetime:        rideRequest.StartDatetime,
			EstimatedEndDatetime: rideRequest.EstimatedEndDatetime,
			NumberOfSeats:        rideRequest.NumberOfSeats,
			StartLocation:        rideRequest.StartLocation,
			EndLocation:          rideRequest.EndLocation,
			RouteDetails:         rideRequest.RouteDetails,
			OpenURL:              openUrl,
		},
	}

	return rideDetailsResp
}

func rideRequestResponse(msg, requestId, openURL string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: msg,
		Data: RideRequestResponse{
			OpenURL:   openURL,
			RequestId: requestId,
		},
	}
}
