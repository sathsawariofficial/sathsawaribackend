package shift

import (
	"encoding/json"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
	"time"
)

func successResp(message string, data any) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	}
}

// seatInfo is worked out from the counter the database keeps in step, and never shows
// a negative remainder.
func seatInfo(total, occupied int) SeatInfo {
	remaining := total - occupied
	if remaining < 0 {
		remaining = 0
	}

	return SeatInfo{
		TotalSeats:     total,
		OccupiedSeats:  occupied,
		RemainingSeats: remaining,
	}
}

func mapShiftSummary(details postgress.ShiftDetails, now time.Time) ShiftSummary {
	next := ""
	if details.Status == constants.Shift_Status_Active {
		next = database.NextOccurrence(detailsWindow(details), now)
	}

	return ShiftSummary{
		ID:               details.ID,
		Name:             details.Name,
		ServiceId:        details.ServiceID,
		ServiceName:      details.ServiceName,
		DriverId:         details.DriverID,
		DriverName:       details.DriverName,
		DriverMobile:     details.DriverMobile,
		VehicleId:        details.VehicleID,
		VehicleNumber:    details.VehicleNumber,
		VehicleInfo:      details.VehicleInfo,
		VehicleOwnerId:   details.VehicleOwnerID,
		VehicleOwnerName: details.VehicleOwnerName,
		DaysOfWeek:       database.IntsFromArray(details.DaysOfWeek),
		StartDate:        details.StartDate,
		EndDate:          details.EndDate,
		StartTime:        details.StartTime,
		EndTime:          details.EndTime,
		Seats:            seatInfo(details.SeatCapacity, details.OccupiedSeats),
		Status:           details.Status,
		NextOccurrence:   next,
		CreatedAt:        details.CreatedAt,
	}
}

func mapRoute(locations []postgress.ShiftLocation) []ShiftLocationDetail {
	route := []ShiftLocationDetail{}

	for _, location := range locations {
		route = append(route, ShiftLocationDetail{
			Sequence: location.Sequence,
			Location: location.Location,
			Lat:      location.Lat,
			Lng:      location.Lng,
			Time:     location.Time,
		})
	}

	return route
}

// shiftDetailResp builds a shift view. When a passenger is reading it the other
// passengers are shown by name and stop only, and their own stop is picked out.
func shiftDetailResp(details postgress.ShiftDetails, locations []postgress.ShiftLocation, passengers []postgress.ShiftPassengerDetails, viewerPassengerId string, now time.Time) ShiftDetailResponse {
	resp := ShiftDetailResponse{
		Shift:      mapShiftSummary(details, now),
		Route:      mapRoute(locations),
		Passengers: []ShiftPassengerDetail{},
	}

	for _, passenger := range passengers {
		stop := locationBySequence(locations, passenger.LocationSequence)
		addedAt := passenger.AddedAt

		entry := ShiftPassengerDetail{
			PassengerId:      passenger.PassengerID,
			PassengerName:    passenger.PassengerName,
			PassengerMobile:  passenger.PassengerMobile,
			LocationSequence: passenger.LocationSequence,
			Location:         stop.Location,
			Time:             stop.Time,
			AddedAt:          &addedAt,
		}

		if !utils.IsStringEmpty(viewerPassengerId) {
			entry.PassengerMobile = ""
			entry.AddedAt = nil

			if passenger.PassengerID == viewerPassengerId {
				mine := entry
				resp.MyStop = &mine
			}
		}

		resp.Passengers = append(resp.Passengers, entry)
	}

	return resp
}

func shiftsPage(rows []postgress.ShiftDetails, totalRows int64, now time.Time) ShiftsResponse {
	shifts := []ShiftSummary{}
	for _, row := range rows {
		shifts = append(shifts, mapShiftSummary(row, now))
	}

	return ShiftsResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		Shifts:     shifts,
	}
}

func shiftHistoryPage(rows []postgress.ShiftHistoryDetails, totalRows int64) ShiftHistoryResponse {
	shifts := []ShiftHistoryItem{}

	for _, row := range rows {
		shifts = append(shifts, ShiftHistoryItem{
			ID:            row.ID,
			Name:          row.Name,
			DriverName:    row.DriverName,
			VehicleNumber: row.VehicleNumber,
			DaysOfWeek:    database.IntsFromArray(row.DaysOfWeek),
			StartDate:     row.StartDate,
			EndDate:       row.EndDate,
			StartTime:     row.StartTime,
			EndTime:       row.EndTime,
			Status:        row.Status,
			CreatedAt:     row.CreatedAt,
			DeletedAt:     row.DeletedAt,
		})
	}

	return ShiftHistoryResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		Shifts:     shifts,
	}
}

func passengerHistoryPage(rows []postgress.ShiftParticipationDetails, totalRows int64) PassengerHistoryResponse {
	history := []PassengerHistoryItem{}

	for _, row := range rows {
		history = append(history, PassengerHistoryItem{
			ShiftPassengerId: row.ID,
			ShiftId:          row.ShiftID,
			ShiftName:        row.ShiftName,
			ShiftStatus:      row.ShiftStatus,
			PassengerId:      row.PassengerID,
			PassengerName:    row.PassengerName,
			PassengerMobile:  row.PassengerMobile,
			LocationSequence: row.LocationSequence,
			Status:           row.Status,
			RemovalReason:    row.RemovalReason,
			JoinedServiceAt:  row.JoinedServiceAt,
			AddedAt:          row.AddedAt,
			RemovedAt:        row.RemovedAt,
			TripsPresent:     row.TripsPresent,
			TripsAbsent:      row.TripsAbsent,
		})
	}

	return PassengerHistoryResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		History:    history,
	}
}

func occurrencesPage(rows []postgress.ShiftOccurrenceDetails, totalRows int64) OccurrencesResponse {
	occurrences := []OccurrenceItem{}

	for _, row := range rows {
		occurrences = append(occurrences, OccurrenceItem{
			OccurrenceId:  row.ID,
			ShiftId:       row.ShiftID,
			ShiftName:     row.ShiftName,
			ServiceName:   row.ServiceName,
			Date:          row.OccurrenceDate,
			StartTime:     row.StartTime,
			EndTime:       row.EndTime,
			DriverName:    row.DriverName,
			VehicleNumber: row.VehicleNumber,
			Status:        row.Status,
			PresentCount:  row.PresentCount,
			AbsentCount:   row.AbsentCount,
		})
	}

	return OccurrencesResponse{
		TotalPages:  utils.CalculatePagesize(totalRows),
		Occurrences: occurrences,
	}
}

func attendanceEntries(rows []postgress.ShiftAttendanceDetails, viewerId string, asPassenger bool) []AttendanceEntry {
	entries := []AttendanceEntry{}

	for _, row := range rows {
		entry := AttendanceEntry{
			PassengerId:      row.PassengerID,
			PassengerName:    row.PassengerName,
			PassengerMobile:  row.PassengerMobile,
			LocationSequence: row.LocationSequence,
			Location:         row.Location,
			Time:             row.LocationTime,
			Status:           row.Status,
			MarkedAt:         row.MarkedAt,
		}

		if asPassenger {
			entry.PassengerMobile = ""
			entry.IsMe = row.PassengerID == viewerId
		}

		entries = append(entries, entry)
	}

	return entries
}

func passengerChangeResp(change passengerChange) UpdateShiftPassengersResponse {
	ids := func(rows []postgress.ShiftPassenger) []string {
		passengerIds := []string{}
		for _, row := range rows {
			passengerIds = append(passengerIds, row.PassengerID)
		}
		return passengerIds
	}

	return UpdateShiftPassengersResponse{
		Added:   ids(change.Added),
		Moved:   ids(change.Moved),
		Removed: ids(change.Removed),
		Seats:   seatInfo(change.Shift.SeatCapacity, change.Shift.OccupiedSeats),
	}
}

// parseRoute reads back the route an occurrence froze.
func parseRoute(route string) []ShiftLocationDetail {
	stops := []ShiftLocationDetail{}

	if err := json.Unmarshal([]byte(route), &stops); err != nil || stops == nil {
		return []ShiftLocationDetail{}
	}

	return stops
}

func travelHistoryPage(rows []postgress.TravelHistoryDetails, totalRows int64) TravelHistoryResponse {
	trips := []TravelHistoryItem{}

	for _, row := range rows {
		trips = append(trips, TravelHistoryItem{
			OccurrenceId:  row.OccurrenceID,
			ShiftId:       row.ShiftID,
			ShiftName:     row.ShiftName,
			ServiceName:   row.ServiceName,
			Date:          row.OccurrenceDate,
			StartTime:     row.StartTime,
			EndTime:       row.EndTime,
			DriverName:    row.DriverName,
			DriverMobile:  row.DriverMobile,
			VehicleNumber: row.VehicleNumber,
			Route:         parseRoute(row.Route),
			MyStop: ShiftPassengerDetail{
				LocationSequence: row.LocationSequence,
				Location:         row.Location,
				Time:             row.LocationTime,
			},
			Attendance: row.Status,
		})
	}

	return TravelHistoryResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		Trips:      trips,
	}
}

func shiftRequestsPage(rows []postgress.ShiftRequestDetails, locations []postgress.ShiftRequestLocation, totalRows int64) ShiftRequestsResponse {
	byRequest := map[string][]utils.RouteLocation{}
	for _, location := range locations {
		byRequest[location.ShiftRequestID] = append(byRequest[location.ShiftRequestID], utils.RouteLocation{
			Location: location.Location,
			Lat:      location.Lat,
			Lng:      location.Lng,
			Time:     location.Time,
		})
	}

	requests := []ShiftRequestItem{}
	for _, row := range rows {
		requestLocations := byRequest[row.ID]
		if requestLocations == nil {
			requestLocations = []utils.RouteLocation{}
		}

		requests = append(requests, ShiftRequestItem{
			ID:            row.ID,
			ServiceId:     row.ServiceID,
			ServiceName:   row.ServiceName,
			PassengerName: row.PassengerName,
			ContactNumber: row.ContactNumber,
			Note:          row.Note,
			DaysOfWeek:    database.IntsFromArray(row.DaysOfWeek),
			StartTime:     row.StartTime,
			EndTime:       row.EndTime,
			Locations:     requestLocations,
			CreatedAt:     row.CreatedAt,
		})
	}

	return ShiftRequestsResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		Requests:   requests,
	}
}
