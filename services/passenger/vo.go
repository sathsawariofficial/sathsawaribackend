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

func filteredRidesResp(rides []postgress.RideRequest, totalRows int) utils.APIResponse {
	var rideRequests []RideRequestDetails
	for _, ride := range rides {
		rideRequests = append(rideRequests, RideRequestDetails{
			ID:                   ride.ID,
			ContactNumber:        ride.ContactNumber,
			StartDatetime:        ride.StartDatetime,
			EstimatedEndDatetime: ride.EstimatedEndDatetime,
			NumberOfSeats:        ride.NumberOfSeats,
			StartLocation:        ride.StartLocation,
			EndLocation:          ride.EndLocation,
			RouteDetails:         ride.RouteDetails,
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
