package shift

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func createShiftResp(shiftId, templateId string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Created_Successfully, "Shift"),
		Data: CreateShiftResponse{
			ShiftId:    shiftId,
			TemplateId: templateId,
		},
	}
}

func mapShiftDetails(shift postgress.ShiftDetails) ShiftDetails {
	return ShiftDetails{
		ID:                   shift.ID,
		GroupId:              shift.GroupID,
		GroupName:            shift.GroupName,
		VehicleId:            shift.VehicleID,
		VehicleNumber:        shift.VehicleNumber,
		VehicleInfo:          shift.VehicleInfo,
		HasAC:                shift.HasAC,
		HasHeating:           shift.HasHeating,
		DriverId:             shift.DriverID,
		DriverName:           shift.DriverName,
		DriverMobile:         shift.DriverMobile,
		Rating:               shift.Rating,
		Direction:            shift.Direction,
		StartDatetime:        shift.StartDatetime,
		EstimatedEndDatetime: shift.EstimatedEndDatetime,
		StartLocation:        shift.StartLocation,
		EndLocation:          shift.EndLocation,
		NumberOfSeats:        shift.NumberOfSeats,
		SeatsTaken:           shift.SeatsTaken,
		RouteDetails:         shift.RouteDetails,
		CreatedById:          shift.CreatedByDriverID,
		CreatedByName:        shift.CreatedByName,
		CreatedByMobile:      shift.CreatedByMobile,
		IsActive:             shift.IsActive,
		CreatedAt:            shift.CreatedAt,
	}
}

func shiftDetailsResp(shift postgress.ShiftDetails, stops []postgress.ShiftStop, seats []postgress.ShiftSeatDetails) utils.APIResponse {
	stopDetails := []ShiftStopDetails{}
	for _, stop := range stops {
		stopDetails = append(stopDetails, ShiftStopDetails{
			ID:             stop.ID,
			SequenceNumber: stop.SequenceNumber,
			Location:       stop.Location,
			Lat:            stop.Lat,
			Lng:            stop.Lng,
			ScheduledTime:  stop.ScheduledTime,
		})
	}

	seatDetails := []ShiftSeatDetails{}
	for _, seat := range seats {
		seatDetails = append(seatDetails, ShiftSeatDetails{
			ID:              seat.ID,
			SeatNumber:      seat.SeatNumber,
			Gender:          seat.Gender,
			Status:          seat.Status,
			PassengerId:     seat.PassengerID,
			PassengerName:   seat.PassengerName,
			PassengerMobile: seat.PassengerMobile,
			StopId:          seat.StopID,
			StopSequence:    seat.SequenceNumber,
			Location:        seat.Location,
			ScheduledTime:   seat.ScheduledTime,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: ShiftDetailsResponse{
			Shift: mapShiftDetails(shift),
			Stops: stopDetails,
			Seats: seatDetails,
		},
	}
}

func shiftsResp(shifts []postgress.ShiftDetails, totalRows int64) utils.APIResponse {
	shiftDetails := []ShiftDetails{}
	for _, shift := range shifts {
		shiftDetails = append(shiftDetails, mapShiftDetails(shift))
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: ShiftsResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Shifts:     shiftDetails,
		},
	}
}

func shiftTemplatesResp(templates []postgress.ShiftTemplate) utils.APIResponse {
	templateDetails := []TemplateDetails{}

	for _, template := range templates {
		stops := []ShiftStopDetails{}
		for _, stop := range template.Stops {
			stops = append(stops, ShiftStopDetails{
				ID:             stop.ID,
				SequenceNumber: stop.SequenceNumber,
				Location:       stop.Location,
				Lat:            stop.Lat,
				Lng:            stop.Lng,
				ScheduledTime:  stop.ScheduledTime,
			})
		}

		seats := []TemplateSeatDetails{}
		for _, seat := range template.Seats {
			seats = append(seats, TemplateSeatDetails{
				SeatNumber:   seat.SeatNumber,
				Gender:       seat.Gender,
				PassengerId:  seat.PassengerID,
				StopSequence: seat.StopSequence,
			})
		}

		daysOfWeek := []int{}
		for _, day := range template.DaysOfWeek {
			daysOfWeek = append(daysOfWeek, int(day))
		}

		templateDetails = append(templateDetails, TemplateDetails{
			ID:                   template.ID,
			GroupId:              template.GroupID,
			Name:                 template.Name,
			VehicleId:            template.VehicleID,
			VehicleNumber:        template.Vehicle.VehicleNumber,
			VehicleInfo:          template.Vehicle.VehicleInfo,
			NumberOfSeats:        template.NumberOfSeats,
			DriverId:             template.DriverID,
			Direction:            template.Direction,
			StartDatetime:        template.StartDatetime,
			EstimatedEndDatetime: template.EstimatedEndTime,
			StartLocation:        template.StartLocation,
			EndLocation:          template.EndLocation,
			RouteDetails:         template.RouteDetails,
			DaysOfWeek:           daysOfWeek,
			Stops:                stops,
			Seats:                seats,
			CreatedAt:            template.CreatedAt,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: TemplatesResponse{
			Templates: templateDetails,
		},
	}
}
