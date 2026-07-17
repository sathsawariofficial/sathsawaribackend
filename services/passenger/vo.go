package passenger

import (
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func filteredRidesResp(rides []postgress.RideRequest, totalRows int) utils.APIResponse {
	var rideRequests []RideRequestDetails
	for _, ride := range rides {
		rideRequests = append(rideRequests, RideRequestDetails{
			ID:                   ride.ID,
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
		Data: RidesRequestDetailsResponse{
			TotalPages: utils.CalculatePagesize(int64(totalRows)),
			Rides:      rideRequests,
		},
	}

	return rideDetailsResp
}
