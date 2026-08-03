package ride

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"time"
)

func createRideResp(rideId, openURL string) utils.APIResponse {
	rideResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Created_Successfully, "Ride"),
		Data: CreateRideResponse{
			Id:      rideId,
			OpenURL: openURL,
		},
	}

	return rideResp
}

func getDriverRidesResp(rides []postgress.RideDetails, totalPages int, driverId string) utils.APIResponse {
	var ridesDetils []RideDetails
	for _, ride := range rides {
		openURL := utils.CreateOpenRideLink(constants.LIKE_TYPE_RIDE_URL, utils.GenerateShortCode(ride.ID))

		ridesDetils = append(ridesDetils, RideDetails{
			ID:                   ride.ID,
			DriverID:             ride.DriverID,
			DriverName:           ride.DriverName,
			DriverMobile:         ride.DriverMobile,
			Rating:               ride.Rating,
			VehicleId:            ride.VehicleId,
			VehicleNumber:        ride.VehicleNumber,
			VehicleInfo:          ride.VehicleInfo,
			StartDatetime:        ride.StartDatetime,
			EstimatedEndDatetime: ride.EstimatedEndDatetime,
			NumberOfSeats:        ride.NumberOfSeats,
			SeatsTaken:           ride.SeatsTaken,
			StartLocation:        ride.StartLocation,
			EndLocation:          ride.EndLocation,
			RoutePoints:          ride.RoutePoints,
			Code:                 ride.Code,
			Fare:                 ride.Fare,
			RouteDetails:         ride.RouteDetails,
			IsActive:             ride.IsActive,
			OpenURL:              openURL,
		})
	}

	driverRideResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: &DriverRideResponse{
			TotalPages: totalPages,
			DriverId:   driverId,
			Rides:      ridesDetils,
		},
	}

	return driverRideResp
}

func filteredRidesResp(rides []postgress.RideDetails, totalRows int64) utils.APIResponse {
	var ridesDetils []RideDetails
	for _, ride := range rides {
		openURL := utils.CreateOpenRideLink(constants.LIKE_TYPE_RIDE_URL, utils.GenerateShortCode(ride.ID))
		ridesDetils = append(ridesDetils, RideDetails{
			ID:                   ride.ID,
			DriverID:             ride.DriverID,
			DriverName:           ride.DriverName,
			DriverMobile:         ride.DriverMobile,
			Rating:               ride.Rating,
			VehicleNumber:        ride.VehicleNumber,
			VehicleInfo:          ride.VehicleInfo,
			StartDatetime:        ride.StartDatetime,
			EstimatedEndDatetime: ride.EstimatedEndDatetime,
			NumberOfSeats:        ride.NumberOfSeats,
			SeatsTaken:           ride.SeatsTaken,
			StartLocation:        ride.StartLocation,
			EndLocation:          ride.EndLocation,
			RoutePoints:          ride.RoutePoints,
			VehicleId:            ride.VehicleId,
			Fare:                 ride.Fare,
			RouteDetails:         ride.RouteDetails,
			ParentRideId:         ride.ParentRideId,
			IsActive:             ride.IsActive,
			Code:                 ride.Code,
			OpenURL:              openURL,
		})
	}

	rideDetailsResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: RidesDetailsResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Rides:      ridesDetils,
		},
	}

	return rideDetailsResp
}

func updateRideResp() utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
	}
}

func getBookedSeatsResp(bookedSeats []postgress.RidePassenger) utils.APIResponse {
	var seats []BookSeats

	for _, seat := range bookedSeats {
		seats = append(seats, BookSeats{
			RideId:       seat.RideId,
			PassengerId:  seat.PassengerId,
			Name:         seat.Name,
			MobileNumber: seat.MobileNumber,
		})
	}

	rideResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data:    seats,
	}

	return rideResp
}

func getRideResp(sessionId string, ride postgress.RideDetails, childRides []postgress.RideDetails) utils.APIResponse {
	var response RideDetailsWithChildrenResponse

	openUrl := utils.CreateOpenRideLink(constants.LIKE_TYPE_RIDE_URL, utils.GenerateShortCode(ride.ID))

	response.Ride = RideDetails{
		ID:                   ride.ID,
		DriverID:             ride.DriverID,
		DriverName:           ride.DriverName,
		DriverMobile:         ride.DriverMobile,
		Rating:               ride.Rating,
		VehicleId:            ride.VehicleId,
		VehicleNumber:        ride.VehicleNumber,
		VehicleInfo:          ride.VehicleInfo,
		RouteDetails:         ride.RouteDetails,
		ParentRideId:         ride.ParentRideId,
		StartDatetime:        ride.StartDatetime,
		EstimatedEndDatetime: ride.EstimatedEndDatetime,
		NumberOfSeats:        ride.NumberOfSeats,
		SeatsTaken:           ride.SeatsTaken,
		StartLocation:        ride.StartLocation,
		EndLocation:          ride.EndLocation,
		RoutePoints:          ride.RoutePoints,
		Fare:                 ride.Fare,
		Code:                 ride.Code,
		OpenURL:              openUrl,
	}

	for _, ride := range childRides {
		openUrl := utils.CreateOpenRideLink(constants.LIKE_TYPE_RIDE_URL, utils.GenerateShortCode(ride.ID))

		response.ChildRides = append(response.ChildRides, RideDetails{
			ID:                   ride.ID,
			DriverID:             ride.DriverID,
			DriverName:           ride.DriverName,
			DriverMobile:         ride.DriverMobile,
			Rating:               ride.Rating,
			VehicleId:            ride.VehicleId,
			VehicleNumber:        ride.VehicleNumber,
			VehicleInfo:          ride.VehicleInfo,
			RouteDetails:         ride.RouteDetails,
			ParentRideId:         ride.ParentRideId,
			StartDatetime:        ride.StartDatetime,
			EstimatedEndDatetime: ride.EstimatedEndDatetime,
			NumberOfSeats:        ride.NumberOfSeats,
			SeatsTaken:           ride.SeatsTaken,
			StartLocation:        ride.StartLocation,
			EndLocation:          ride.EndLocation,
			RoutePoints:          ride.RoutePoints,
			Fare:                 ride.Fare,
			Code:                 ride.Code,
			OpenURL:              openUrl,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data:    response,
	}
}

func getRideTemplatesResp(sessionId string, templates []postgress.RideTemplate) utils.APIResponse {
	var rideTemplates []RideTemplate

	for _, template := range templates {
		startDate, err := time.ParseInLocation(constants.DateTimeLayout, template.StartDatetime, time.Local)
		if err != nil {
			logger.LogError(sessionId, err)
			continue
		}

		endDate, err := time.ParseInLocation(constants.DateTimeLayout, template.EstimatedEndDatetime, time.Local)
		if err != nil {
			logger.LogError(sessionId, err)
			continue
		}

		now := time.Now()

		newStart := time.Date(
			now.Year(), now.Month(), now.Day(),
			startDate.Hour(), startDate.Minute(), startDate.Second(), startDate.Nanosecond(),
			time.Local,
		)

		newEnd := time.Date(
			now.Year(), now.Month(), now.Day(),
			endDate.Hour(), endDate.Minute(), endDate.Second(), endDate.Nanosecond(),
			time.Local,
		)

		rideTemplates = append(rideTemplates, RideTemplate{
			ID:                   template.ID,
			RideID:               template.RideID,
			DriverID:             template.DriverID,
			VehicleID:            template.VehicleID,
			VehicleNumber:        template.Vehicle.VehicleNumber,
			VehicleInfo:          template.Vehicle.VehicleInfo,
			StartDatetime:        newStart.Format(constants.DateTimeLayout),
			EstimatedEndDatetime: newEnd.Format(constants.DateTimeLayout),
			NumberOfSeats:        template.NumberOfSeats,
			SeatsTaken:           template.SeatsTaken,
			StartLocation:        template.StartLocation,
			EndLocation:          template.EndLocation,
			Fare:                 template.Fare,
			RouteDetails:         template.RouteDetails,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data:    rideTemplates,
	}
}
