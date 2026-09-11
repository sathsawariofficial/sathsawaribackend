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
	"gorm.io/gorm/clause"
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

// rideSlots lays out every ride a request puts on the calendar: the ride itself first,
// then for a recurring request each repeat in date order, exactly as the series is
// written.
func rideSlots(request RideCreationRequest) (slots []database.RideSlot, err error) {
	start, err := time.ParseInLocation(constants.DateTimeLayout, request.StartDatetime, constants.Business_Location)
	if err != nil {
		return
	}

	end, err := time.ParseInLocation(constants.DateTimeLayout, request.EstimatedEndDatetime, constants.Business_Location)
	if err != nil {
		return
	}

	slots = append(slots, database.RideSlot{Start: start, End: end})

	if !request.IsRecurring {
		return
	}

	frequency := findFrequency(request.Frequency, request.Period)

	if request.Period == WEEKLY {
		// the i-th day after the start, numbered 1 (Monday) to 7 (Sunday), gets a ride
		// only when it is one of the requested days
		dayOfWeek := int(start.Weekday()) + 1
		for i := 1; i <= frequency; i++ {
			day := ((dayOfWeek - 1) % 7) + 1
			dayOfWeek++

			if !utils.InSlice(request.DaysOfWeek, day) {
				continue
			}

			childStart, childEnd := shiftDailyDates(start, end, i)
			slots = append(slots, database.RideSlot{Start: childStart, End: childEnd})
		}

		return
	}

	for i := 1; i <= frequency; i++ {
		childStart, childEnd := shiftDailyDates(start, end, i)
		if request.Period == MONTHLY {
			childStart, childEnd = shiftMonthlyDates(start, end, i)
		}

		slots = append(slots, database.RideSlot{Start: childStart, End: childEnd})
	}

	return
}

// seriesOverlap finds the first date a series runs into itself. A ride lasting longer
// than the gap to its next repeat would need the driver and the vehicle twice at once.
func seriesOverlap(slots []database.RideSlot) (string, bool) {
	for i := 1; i < len(slots); i++ {
		if slots[i].Start.Before(slots[i-1].End) {
			return slots[i].Start.Format(constants.Date_Layout), true
		}
	}

	return "", false
}

// createRides writes a ride and its whole series in one transaction. The driver and
// the vehicle are locked under the same keys shifts use, then every date is checked
// against the driver's rides on any vehicle, the vehicle's rides and the shift trips
// of either, so nothing can take the time between the check and the write, and a
// series is created whole or not at all.
func createRides(orgCtx *gin.Context, sessionId string, request RideCreationRequest, vehicleId string, slots []database.RideSlot) (rideId string, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := database.LockShiftResources(tx, []string{request.EXTDriverId}, []string{vehicleId}, nil); e != nil {
			return e
		}

		clash, e := database.FindRideScheduleClash(tx, request.EXTDriverId, vehicleId, slots, "")
		if e != nil {
			return e
		}

		if clash != nil {
			return utils.Refuse(clash.RideMessage())
		}

		rides := make([]postgress.Ride, 0, len(slots))
		for _, slot := range slots {
			rideRequest := request
			rideRequest.StartDatetime = slot.Start.Format(constants.DateTimeLayout)
			rideRequest.EstimatedEndDatetime = slot.End.Format(constants.DateTimeLayout)

			parentId := ""
			if len(rides) > 0 {
				parentId = rides[0].ID
			}

			rides = append(rides, mapRideData(rideRequest, parentId, request.EXTDriverId, vehicleId))
		}

		if e = tx.Create(&rides).Error; e != nil {
			return e
		}

		rideId = rides[0].ID

		return nil
	})

	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Creation_Failed, "ride"))
	}

	return
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

// getRideSeriesByRootId returns the root ride (if it still exists) plus all
// of its remaining children, scoped to the owning driver, row-locked so a
// concurrent booking on any of these rides is blocked until the caller's
// transaction commits or rolls back. Must be called with tx already inside
// an open transaction on the same rows it is about to delete.
func getRideSeriesByRootId(tx *gorm.DB, driverId, rootId string) (rides []postgress.Ride, err error) {
	err = tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("driver_id = ?", driverId).
		Where("(id = ? OR parent_ride_id = ?)", rootId, rootId).
		Where("is_active = ?", true).
		Order("start_datetime ASC").
		Find(&rides).Error
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
			rides.parent_ride_id,
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
		Where("rides.id = ? AND rides.is_active = ?", rideId, true).
		First(&ride).Error
	if err != nil {
		logger.LogError(sessionId, err)
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

// checkRideDeletionGuards evaluates the two ride-cancellation safety rules
// against a single ride row: it must not already have a seat booked, and it
// must not start within Ride_Cancel_Min_Hours_Before_Start hours from now.
// `now` is passed in rather than computed internally so a whole batch of
// rides can be judged against one consistent timestamp snapshot.
// Returns blocked=false when the ride is safe to cancel/delete.
func checkRideDeletionGuards(ride postgress.Ride, now time.Time) (blocked bool, reasonCode string, message string) {
	if ride.SeatsTaken > 0 {
		return true, Ride_Cancel_Skip_Reason_Booked,
			fmt.Sprintf(Ride_Cancel_Skip_Message_Booked, ride.SeatsTaken)
	}

	startDate, err := time.ParseInLocation(constants.DateTimeLayout, ride.StartDatetime, constants.Business_Location)
	if err != nil {
		// fail closed: never delete a ride whose start time we can't verify
		return true, Ride_Cancel_Skip_Reason_Invalid, Ride_Cancel_Skip_Message_Invalid
	}

	cutoff := now.Add(time.Duration(constants.Ride_Cancel_Min_Hours_Before_Start) * time.Hour)
	if startDate.Before(cutoff) {
		return true, Ride_Cancel_Skip_Reason_Imminent,
			fmt.Sprintf(Ride_Cancel_Skip_Message_Imminent, constants.Ride_Cancel_Min_Hours_Before_Start)
	}

	return false, "", ""
}
