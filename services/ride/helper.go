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
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func mapRideData(
	request RideCreationRequest,
	parentId,
	driverId,
	vehicleId string,
) postgress.Ride {

	var parentRideId string

	if !utils.IsStringEmpty(parentId) {
		parentRideId = parentId
	}

	// normalize route points
	normalizedRoutePoints := make([]string, 0, len(request.RoutePoints))
	normalizedRoutePoints = append(normalizedRoutePoints, strings.ToLower(request.StartLocation))
	for _, loc := range request.RoutePoints {
		loc = strings.ToLower(strings.TrimSpace(loc))
		if loc != "" {
			normalizedRoutePoints = append(normalizedRoutePoints, loc)
		}
	}
	normalizedRoutePoints = append(normalizedRoutePoints, strings.ToLower(request.EndLocation))

	return postgress.Ride{
		ID:                   database.GenerateUUID(),
		DriverID:             driverId,
		VehicleID:            vehicleId,
		StartDatetime:        request.StartDatetime,
		EstimatedEndDatetime: request.EstimatedEndDatetime,
		NumberOfSeats:        request.NumberOfSeats,
		SeatsTaken:           0,

		StartLocation: strings.ToLower(strings.TrimSpace(request.StartLocation)),
		EndLocation:   strings.ToLower(strings.TrimSpace(request.EndLocation)),

		RoutePoints: pq.StringArray(normalizedRoutePoints),

		Fare:         request.Fare,
		Code:         utils.GenerateOTP(),
		RouteDetails: request.RouteDetails,

		ParentRideId: parentRideId,
		IsActive:     true,
	}
}

func mapRideToRideTemplateData(request RideCreationRequest, rideId, driverId, vehicleId string) postgress.RideTemplate {
	return postgress.RideTemplate{
		ID:                   database.GenerateUUID(),
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
	}
}

func VehicleHasRideDuringTime(
	vehicleID string,
	startTime string,
	endTime string,
) (bool, error) {

	var count int64

	err := database.DatabaseConn.Postgres.Model(&postgress.Ride{}).
		Where("vehicle_id = ?", vehicleID).
		Where("is_active = ?", true).
		Where(`
			start_datetime < ?
			AND estimated_end_datetime > ?
		`, endTime, startTime).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
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
			rides.route_points AS route_points,
			rides.fare,
			rides.vehicle_id,
			rides.code,
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

func getFilteredRides(
	orgCtx *gin.Context,
	page int,
	startTime,
	endTime,
	searchLoc string,
) (
	rides []postgress.RideDetails,
	totalRows int64,
	err error,
) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	// Build date query targeting the indexed ride_searches table
	dateCondition := ""
	if startTime != "" && endTime != "" {
		dateCondition = "ride_searches.start_datetime >= ? AND rides.estimated_end_datetime <= ?"
	} else if startTime != "" {
		dateCondition = "ride_searches.start_datetime >= ?"
		endTime = ""
	} else if endTime != "" {
		dateCondition = "rides.estimated_end_datetime <= ?"
		startTime = ""
	}

	if startTime == "" && endTime == "" {
		now := time.Now()
		weekLater := now.AddDate(0, 0, 7)

		startTime = now.Format("2006-01-02 15:04:05")
		endTime = weekLater.Format("2006-01-02 15:04:05")

		dateCondition = "ride_searches.start_datetime >= ? AND ride_searches.start_datetime <= ?"
	}

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	// Base query targeting our fast replica table first
	query := database.DatabaseConn.Postgres.
		WithContext(ctx).
		Table("ride_searches").
		Select(`
			ride_searches.ride_id AS id,
			rides.driver_id,
			drivers.driver_name,
			drivers.driver_mobile,
			drivers.rating,
			vehicles.vehicle_number,
			vehicles.vehicle_info,
			ride_searches.start_datetime,
			rides.estimated_end_datetime,
			rides.number_of_seats,
			rides.seats_taken,
			ride_searches.start_location,
			ride_searches.end_location,
			ride_searches.route_points AS route_points,
			rides.vehicle_id,
			rides.fare,
			rides.route_details,
			rides.parent_ride_id,
			ride_searches.is_active,
			rides.created_at,
			rides.updated_at
		`).
		// Lazy-join metadata ONLY on rows that pass search filters
		Joins("JOIN rides ON ride_searches.ride_id = rides.id").
		Joins("JOIN drivers ON rides.driver_id = drivers.id").
		Joins("JOIN vehicles ON rides.vehicle_id = vehicles.id").
		Where("ride_searches.is_active = ?", true)

	// High-performance location match utilizing Trigram and GIN indexes
	if strings.TrimSpace(searchLoc) != "" {
		cleanSearch := strings.TrimSpace(searchLoc)
		normalizedSearch := strings.ToLower(cleanSearch)

		// Evaluates Trigram index on strings and GIN index on route arrays concurrently
		query = query.Where(
			"ride_searches.start_location ILIKE ? OR ride_searches.end_location ILIKE ? OR ride_searches.route_points @> ?",
			"%"+cleanSearch+"%",
			"%"+cleanSearch+"%",
			pq.Array([]string{normalizedSearch}),
		)
	}

	// Apply structured dates
	if dateCondition != "" {
		if startTime != "" && endTime != "" {
			query = query.Where(dateCondition, startTime, endTime)
		} else if startTime != "" {
			query = query.Where(dateCondition, startTime)
		} else if endTime != "" {
			query = query.Where(dateCondition, endTime)
		}
	}

	// Clone the database session cleanly to count matching entries
	countQuery := query.Session(&gorm.Session{})

	// Add pagination and sort order optimization
	query = query.
		Limit(pageSize).
		Offset(offset).
		Order("ride_searches.start_datetime ASC")

	// Execute execution loop
	err = query.Find(&rides).Error
	if err != nil {
		return rides, 0, err
	}

	// Execute row tally
	err = countQuery.Count(&totalRows).Error

	return rides, totalRows, err
}

func getQueryParams(ctx *gin.Context) (startTime, endTime, searchLoc, startLoc, endLoc string) {
	startTime = ctx.DefaultQuery(constants.Start_Time_Key, "")
	endTime = ctx.DefaultQuery(constants.Extimated_End_Time_Key, "")
	searchLoc = ctx.DefaultQuery(constants.Search_Loc_Key, "")
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
		ID:           database.GenerateUUID(),
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

func getRide(
	orgCtx *gin.Context,
	sessionId,
	rideId string,
) (
	ride postgress.RideDetails,
	childRides []postgress.RideDetails,
	err error,
) {
	logger.LogInfo("Request received in getRide", sessionId)
	logger.LogDebug("Request received in getRide", sessionId, rideId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	selectClause := `
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
		rides.route_points,
		rides.fare,
		rides.vehicle_id,
		rides.code,
		rides.route_details,
		rides.parent_ride_id AS parent_id,
		rides.is_active,
		rides.created_at,
		rides.updated_at
	`

	err = db.
		Table("rides").
		Select(selectClause).
		Joins("JOIN drivers ON rides.driver_id = drivers.id").
		Joins("JOIN vehicles ON rides.vehicle_id = vehicles.id").
		Where("rides.id = ? AND ride.is_active = ?", rideId, true).
		First(&ride).Error
	if err != nil {
		return
	}

	err = db.
		Table("rides").
		Select(selectClause).
		Joins("JOIN drivers ON rides.driver_id = drivers.id").
		Joins("JOIN vehicles ON rides.vehicle_id = vehicles.id").
		Where("rides.parent_ride_id = ?", rideId).
		Order("rides.start_datetime ASC").
		Find(&childRides).Error

	logger.LogInfo("Response returned from getRide", sessionId)

	return
}

func getRideTemplates(orgCtx *gin.Context, sessionId, driverId string) (rideTemplates []postgress.RideTemplate, err error) {
	logger.LogInfo("Request recevied in getRideTemplates", sessionId)
	logger.LogDebug("Request recevied in getRideTemplates", sessionId, driverId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.
		WithContext(ctx).
		Joins("JOIN vehicles ON vehicles.id = ride_templates.vehicle_id").
		Where("ride_templates.driver_id = ?", driverId).
		Where("vehicles.status = ?", constants.Status_Active).
		Preload("Vehicle").
		Find(&rideTemplates).Error

	logger.LogInfo("Response returned from getRideTemplates", sessionId)
	return
}

func deleteRideTemplate(orgCtx *gin.Context, sessionId, rideTemplateId string) (err error) {
	logger.LogInfo("Request received in deleteRideTemplate", sessionId)
	logger.LogDebug("Request received in deleteRideTemplate", sessionId, rideTemplateId)

	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("id = ?", rideTemplateId).
		Delete(&postgress.RideTemplate{}).Error

	logger.LogInfo("Response returned from deleteRideTemplate", sessionId)
	return
}
