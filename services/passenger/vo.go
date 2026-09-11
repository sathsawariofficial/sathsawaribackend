package passenger

import (
	"fmt"
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

func registerPassengerResp(passengerId, otp string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Registered_Successfully, "Passenger"),
		Data: PassengerRegistrationResponse{
			PassengerId: passengerId,
			OTP:         otp,
		},
	}
}

func loginPassengerResp(passengerSessionId, otp string, passenger postgress.Passenger) utils.APIResponse {
	message := fmt.Sprintf(constants.Loggedin_Successfully, "Passenger")
	if utils.IsStringEmpty(passengerSessionId) {
		message = constants.SENT_OTP_Successfully
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: message,
		Data: PassengerLoginResponse{
			Token: passengerSessionId,
			OTP:   otp,
			Passenger: PassengerInfo{
				ID:              passenger.ID,
				PassengerMobile: passenger.PassengerMobile,
				PassengerName:   passenger.PassengerName,
				Gender:          passenger.Gender,
				IsVerified:      passenger.Status == constants.Status_Active,
			},
		},
	}
}

func passengerProfileResp(passenger postgress.Passenger) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: PassengerProfileInfoResponse{
			PassengerDetails: PassengerInfo{
				ID:              passenger.ID,
				PassengerMobile: passenger.PassengerMobile,
				PassengerName:   passenger.PassengerName,
				Gender:          passenger.Gender,
				IsVerified:      passenger.Status == constants.Status_Active,
			},
		},
	}
}

func passengerOTPResp(message, otp string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: message,
		Data: PassengerOTPResponse{
			OTP: otp,
		},
	}
}

// availabilityResp always answers with the whole week, a day never filled in reads as
// not required.
func availabilityResp(serviceId string, days []postgress.PassengerAvailability, locations []postgress.PassengerAvailabilityLocation) utils.APIResponse {
	byDay := map[int][]utils.RouteLocation{}
	for _, location := range locations {
		byDay[location.DayOfWeek] = append(byDay[location.DayOfWeek], utils.RouteLocation{
			Location: location.Location,
			Lat:      location.Lat,
			Lng:      location.Lng,
			Time:     location.Time,
		})
	}

	required := map[int]bool{}
	for _, day := range days {
		required[day.DayOfWeek] = day.IsRequired
	}

	week := make([]AvailabilityDay, 0, constants.Day_Of_Week_Max_Value)
	for day := constants.Day_Of_Week_Min_Value; day <= constants.Day_Of_Week_Max_Value; day++ {
		dayLocations := byDay[day]
		if dayLocations == nil {
			dayLocations = []utils.RouteLocation{}
		}

		week = append(week, AvailabilityDay{
			DayOfWeek:  day,
			IsRequired: required[day],
			Locations:  dayLocations,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AvailabilityResponse{
			ServiceId: serviceId,
			Days:      week,
		},
	}
}
