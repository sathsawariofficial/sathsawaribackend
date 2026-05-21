package ride

import (
	"context"
	"errors"
	"fmt"
	"math"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func mapRideData(request RideCreationRequest, parentId, driverId, vehicleId string) postgress.Ride {
	var parentRideId string

	if !utils.IsStringEmpty(parentId) {
		parentRideId = parentId
	}

	return postgress.Ride{
		ID:                   utils.GenerateUUID(),
		DriverID:             driverId,
		VehicleID:            vehicleId,
		StartDatetime:        request.StartDatetime,
		EstimatedEndDatetime: request.EstimatedEndDatetime,
		NumberOfSeats:        request.NumberOfSeats,
		SeatsTaken:           0,
		StartLocation:        request.StartLocation,
		EndLocation:          request.EndLocation,
		Fare:                 request.Fare,
		RouteDetails:         request.RouteDetails,
		ParentRideId:         parentRideId,
		IsActive:             true,
	}
}

func mapRideToRideTemplateData(request RideCreationRequest, rideId, driverId, vehicleId string) postgress.RideTemplate {
	return postgress.RideTemplate{
		ID:                   utils.GenerateUUID(),
		RideID:               rideId,
		DriverID:             driverId,
		VehicleID:            vehicleId,
		StartDatetime:        request.StartDatetime,
		EstimatedEndDatetime: request.EstimatedEndDatetime,
		NumberOfSeats:        request.NumberOfSeats,
		SeatsTaken:           0,
		StartLocation:        request.StartLocation,
		EndLocation:          request.EndLocation,
		Fare:                 request.Fare,
		RouteDetails:         request.RouteDetails,
		IsActive:             true,
	}
}

func mapContactData(request ApprochRequest) postgress.ApprochInfo {
	return postgress.ApprochInfo{
		ID:      utils.GenerateUUID(),
		Name:    request.Name,
		Number:  request.Number,
		Email:   request.Email,
		Message: request.Message,
	}
}

func getVehicleById(orgCtx *gin.Context, vehicleId string) (driver postgress.Vehicle, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where(`id = ?`, vehicleId).Where(`status = ?`, constants.Status_Active).Find(&driver).Error
	return
}

func getRideById(orgCtx *gin.Context, rideId string) (ride postgress.Ride, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where(`id = ?`, rideId).Where(`is_active = ?`, true).Find(&ride).Error
	return
}

func getAllActiveRides(orgCtx *gin.Context, page int) (rides []postgress.RideDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Table("rides").
		Select(`
			rides.id,
			rides.driver_id,
			drivers.driver_name,
			drivers.driver_mobile,
			vehicles.vehicle_number,
			vehicles.vehicle_info,
			rides.start_datetime,
			rides.estimated_end_datetime,
			rides.number_of_seats,
			rides.seats_taken,
			rides.start_location,
			rides.end_location,
			rides.fare,
			rides.route_details,
			rides.is_active,
			rides.created_at,
			rides.updated_at
		`).
		Joins("JOIN drivers ON rides.driver_id = drivers.id").
		Joins("JOIN vehicles ON rides.vehicle_id = vehicles.id").
		Where("rides.estimated_end_datetime >= ?", utils.GetCurrentTime()).
		Where(`rides.is_active = ?`, true).
		Limit(pageSize).
		Offset(offset)

	err = query.Order("created_at desc").Find(&rides).Error
	err = query.Count(&totalRows).Error

	return
}

func getAllRidesByDriver(orgCtx *gin.Context, page int, driverId, startTime, endTime, startLoc, endLoc, status string) (rides []postgress.RideDetails, pages int, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	// Only apply date filters if both startTime and endTime are provided
	dateCondition := ""
	if startTime != "" && endTime != "" {
		dateCondition = "rides.start_datetime >= ? AND rides.estimated_end_datetime <= ?"
	} else if startTime != "" {
		dateCondition = "rides.start_datetime >= ?"
		endTime = "" // Don't bind the endTime in the query
	} else if endTime != "" {
		dateCondition = "rides.estimated_end_datetime <= ?"
		startTime = "" // Don't bind the startTime in the query
	}

	// Default each location parameter to a wildcard if it's empty
	if startLoc == "" {
		startLoc = "%"
	}
	if endLoc == "" {
		endLoc = "%"
	}

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Table("rides").
		Select(`
			rides.id,
			rides.driver_id,
			drivers.driver_name,
			drivers.driver_mobile,
			drivers.rating,
			vehicles.vehicle_number,
			vehicles.vehicle_info,
			rides.start_datetime,
			rides.estimated_end_datetime,
			rides.number_of_seats,
			rides.seats_taken,
			rides.start_location,
			rides.end_location,
			rides.fare,
			rides.route_details,
			rides.is_active,
			rides.created_at,
			rides.updated_at
		`).
		Joins("JOIN drivers ON rides.driver_id = drivers.id").
		Joins("JOIN vehicles ON rides.vehicle_id = vehicles.id").
		Where("rides.start_location ILIKE ? AND rides.end_location ILIKE ?", "%"+startLoc+"%", "%"+endLoc+"%").
		Where("drivers.id = ?", driverId)

	if strings.EqualFold(status, constants.Ride_Status_Active) {
		query = query.Where(`rides.is_active = ?`, true)
	} else if strings.EqualFold(status, constants.Ride_Status_InActive) {
		query = query.Where(`rides.is_active = ?`, false)
	}

	// Add the date condition to the query if present
	if dateCondition != "" {
		if startTime != "" && endTime != "" {
			query = query.Where(dateCondition, startTime, endTime)
		} else if startTime != "" {
			query = query.Where(dateCondition, startTime)
		} else if endTime != "" {
			query = query.Where(dateCondition, endTime)
		}
	}

	countQuery := query
	query = query.Order("rides.start_datetime ASC").Limit(pageSize).Offset(offset)

	// Select the fields you need from the joined tables (Ride, Driver, and Vehicle)
	var totalRows int64
	err = query.Find(&rides).Error
	err = countQuery.Count(&totalRows).Error
	pages = int(math.Ceil(float64(totalRows) / float64(pageSize)))

	return
}

func getFilteredRides(orgCtx *gin.Context, page int, startTime, endTime, startLoc, endLoc string) (rides []postgress.RideDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	// Only apply date filters if both startTime and endTime are provided
	dateCondition := ""
	if startTime != "" && endTime != "" {
		dateCondition = "rides.start_datetime >= ? AND rides.estimated_end_datetime <= ?"
	} else if startTime != "" {
		dateCondition = "rides.start_datetime >= ?"
		endTime = "" // Don't bind the endTime in the query
	} else if endTime != "" {
		dateCondition = "rides.estimated_end_datetime <= ?"
		startTime = "" // Don't bind the startTime in the query
	}

	if startTime == "" && endTime == "" {
		now := time.Now()
		weekLater := now.AddDate(0, 0, 7)

		startTime = now.Format("2006-01-02 15:04:05")
		endTime = weekLater.Format("2006-01-02 15:04:05")

		dateCondition = "rides.start_datetime >= ? AND rides.start_datetime <= ?"
	}

	// Default each location parameter to a wildcard if it's empty
	if startLoc == "" {
		startLoc = "%"
	}
	if endLoc == "" {
		endLoc = "%"
	}

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Table("rides").
		Select(`
		rides.id,
		rides.driver_id,
		drivers.driver_name,
		drivers.driver_mobile,
		drivers.rating,
		vehicles.vehicle_number,
		vehicles.vehicle_info,
		rides.start_datetime,
		rides.estimated_end_datetime,
		rides.number_of_seats,
		rides.seats_taken,
		rides.start_location,
		rides.end_location,
		rides.fare,
		rides.route_details,
		rides.is_active,
		rides.created_at,
		rides.updated_at
	`).
		Joins("JOIN drivers ON rides.driver_id = drivers.id").
		Joins("JOIN vehicles ON rides.vehicle_id = vehicles.id").
		Where("rides.start_location ILIKE ? AND rides.end_location ILIKE ?", "%"+startLoc+"%", "%"+endLoc+"%").
		Where("rides.is_active = ?", true)

	// Add the date condition to the query if present
	if dateCondition != "" {
		if startTime != "" && endTime != "" {
			query = query.Where(dateCondition, startTime, endTime)
		} else if startTime != "" {
			query = query.Where(dateCondition, startTime)
		} else if endTime != "" {
			query = query.Where(dateCondition, endTime)
		}
	}

	countQuery := query
	query = query.Limit(pageSize).Offset(offset)

	// Select the fields you need from the joined tables (Ride, Driver, and Vehicle)
	err = query.Order("rides.start_datetime ASC").Find(&rides).Error
	err = countQuery.Count(&totalRows).Error

	return
}

func getQueryParams(ctx *gin.Context) (startTime, endTime, startLoc, endLoc string) {
	startTime = ctx.DefaultQuery(constants.Start_Time_Key, "")
	endTime = ctx.DefaultQuery(constants.Extimated_End_Time_Key, "")
	startLoc = ctx.DefaultQuery(constants.Start_Loc_Key, "")
	endLoc = ctx.DefaultQuery(constants.End_Loc_Key, "")

	return
}

func shiftDailyDates(baseStart, baseEnd time.Time, i int) (time.Time, time.Time) {
	return baseStart.AddDate(0, 0, i), baseEnd.AddDate(0, 0, i)
}

func shiftMonthlyDates(baseStart, baseEnd time.Time, i int) (time.Time, time.Time) {
	return baseStart.AddDate(0, i, 0), baseEnd.AddDate(0, i, 0)
}

func findFrequency(frequency, period int) int {
	if period == WEEKLY {
		frequency = 7 * frequency
	}

	return frequency
}

func updateTakenSeats(orgCtx *gin.Context, rideId string, seats_taken int) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	return database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Ride{}).
		Where("id = ?", rideId).
		Update("seats_taken", seats_taken).
		Error
}

func mapBookSeatData(request BookSeatRequest) postgress.RidePassenger {
	return postgress.RidePassenger{
		ID:           utils.GenerateUUID(),
		RideId:       request.RideId,
		PassengerId:  request.PassengerId,
		Name:         request.Name,
		MobileNumber: request.MobileNumber,
		IsActive:     true,
	}
}

func getBookedSeatsByRideId(orgCtx *gin.Context, rideId string) (ridePassenger []postgress.RidePassenger, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where(`id = ?`, rideId).Where(`is_active = ?`, true).Find(&ridePassenger).Error
	return
}

func updateStatusBookedSeatsById(orgCtx *gin.Context, id string) (err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.RidePassenger{}).
		Where(`id = ?`, id).
		Where(`is_active = ?`, true).
		Update("is_active = ?", false).
		Error
	return
}

func bookSeatByDriver(orgCtx *gin.Context, sessionId, driverId string, request BookSeatRequest) (bookedRide postgress.RidePassenger, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()

	ride, err := getRideById(orgCtx, request.RideId)
	if err != nil {
		logger.LogError(sessionId, "failed to get ride error: "+err.Error())
		err = errors.New(constants.Ride_Not_Found)
		tx.Rollback()
		return
	}

	if ride.DriverID != driverId {
		err = errors.New(constants.Operation_Not_Permitted)
		logger.LogError(sessionId, "driver is not the owner of this ride error: "+err.Error())
		tx.Rollback()
		return
	}

	if ride.SeatsTaken+request.NumberOfSeats > ride.NumberOfSeats {
		err = fmt.Errorf(constants.Failed_To_Do_Job, fmt.Sprintf("book the ride due to unavailblity of %d seat(s)", request.NumberOfSeats))
		logger.LogError(sessionId, "failed to book ride error: "+err.Error())
		tx.Rollback()
		return
	}

	for i := 0; i < request.NumberOfSeats; i++ {
		bookedRide = mapBookSeatData(request)
		var cancel context.CancelFunc
		ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
		defer cancel()

		if err = database.DatabaseConn.Postgres.WithContext(ctx).Create(&bookedRide).Error; err != nil {
			logger.LogError(sessionId, "failed to book ride error: "+err.Error())
			err = fmt.Errorf(constants.Failed_To_Do_Job, "book the seat")
			tx.Rollback()

			return
		}
	}

	err = updateTakenSeats(orgCtx, ride.ID, ride.SeatsTaken+1)
	if err != nil {
		logger.LogError(sessionId, "failed to update vehicle error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "book the seat")
		tx.Rollback()

		return
	}

	tx.Commit()
	return
}

func getRideTemplates(orgCtx *gin.Context, sessionId, driverId string) (rideTemplates []postgress.RideTemplate, err error) {
	logger.LogInfo("Request recevied in getRideTemplates", sessionId)
	logger.LogDebug("Request recevied in getRideTemplates", sessionId, driverId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where(`driver_id = ?`, driverId).Where(`is_active = ?`, true).Find(&rideTemplates).Error

	logger.LogInfo("Response returned from getRideTemplates", sessionId)
	return
}

func deleteRideTemplate(orgCtx *gin.Context, sessionId, rideTemplateId string) (err error) {
	logger.LogInfo("Request recevied in deleteRideTemplate", sessionId)
	logger.LogDebug("Request recevied in deleteRideTemplate", sessionId, rideTemplateId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.RideTemplate{}).
		Where(`id = ?`, rideTemplateId).
		Update("is_active", false).Error

	logger.LogInfo("Response returned from deleteRideTemplate", sessionId)
	return
}
