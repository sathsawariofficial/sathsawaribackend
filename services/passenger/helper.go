package passenger

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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func getRideRequestFromQuery(orgCtx *gin.Context, sessionId string) GetRideRequest {
	logger.LogInfo("Request received in getRideRequestFromQuery", sessionId)

	request := GetRideRequest{
		StartDatetime:        orgCtx.Query("startDatetime"),
		EstimatedEndDatetime: orgCtx.Query("estimatedEndDatetime"),
		StartLocation:        orgCtx.Query("startLocation"),
		EndLocation:          orgCtx.Query("endLocation"),
	}

	logger.LogInfo("Response returned from getRideRequestFromQuery", sessionId)
	logger.LogDebug("Response returned from getRideRequestFromQuery", sessionId, request)

	return request
}

func bookRide(orgCtx *gin.Context, sessionId string, request BookSeatRequest) (string, error) {
	var bookingId string

	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()

	result := tx.Exec(`
		UPDATE rides
		SET seats_taken = seats_taken + ?
		WHERE id = ?
		  AND code = ?
		  AND is_active = true
		  AND seats_taken + ? <= number_of_seats
	`,
		request.Seats,
		request.RideId,
		request.Code,
		request.Seats,
	)

	if result.Error != nil {
		tx.Rollback()
		return bookingId, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return bookingId, fmt.Errorf("invalid ride code, ride not found, or insufficient seats available")
	}

	bookingId = fmt.Sprintf("SS-%s", database.GenerateUUID())
	if err := tx.Create(&postgress.RideBooking{
		ID:           database.GenerateUUID(),
		BookingID:    bookingId,
		RideID:       request.RideId,
		MobileNumber: request.MobileNumber,
		Name:         request.Name,
		Seats:        request.Seats,
	}).Error; err != nil {
		logger.LogError(sessionId, err)
		tx.Rollback()
		return bookingId, fmt.Errorf(constants.Registeration_Failed, "booking")
	}

	return bookingId, tx.Commit().Error
}

func mapRideRequest(request RideRequest) postgress.RideRequest {
	return postgress.RideRequest{
		ID:                   database.GenerateUUID(),
		StartDatetime:        request.StartDatetime,
		EstimatedEndDatetime: request.EstimatedEndDatetime,
		NumberOfSeats:        request.NumberOfSeats,
		StartLocation:        request.StartLocation,
		EndLocation:          request.EndLocation,
		RouteDetails:         request.RouteDetails,
		ContactNumber:        request.ContactNumber,
		IsActive:             true,
	}
}

func getFilterAndPaginateRideRequests(
	orgCtx *gin.Context,
	page int,
	startTime,
	endTime,
	startLoc,
	endLoc string,
) (rides []postgress.RideRequest, totalPages int, err error) {
	// 1. Setup pagination variables
	pageSize := configuration.ConfigurationData.PageSize
	if pageSize <= 0 {
		pageSize = 10 // Fallback default
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// 2. Manage context with timeout
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	// 3. Build the base query
	query := database.DatabaseConn.Postgres.WithContext(ctx).Model(&postgress.RideRequest{}).Where("is_active = ?", true)

	// 4. Apply conditional location filters (ILIKE)
	if startLoc = strings.TrimSpace(startLoc); startLoc != "" {
		query = query.Where("start_location ILIKE ?", "%"+startLoc+"%")
	}
	if endLoc = strings.TrimSpace(endLoc); endLoc != "" {
		query = query.Where("end_location ILIKE ?", "%"+endLoc+"%")
	}

	// 5. Apply conditional date/time filters
	startTime = strings.TrimSpace(startTime)
	endTime = strings.TrimSpace(endTime)

	if startTime != "" && endTime != "" {
		query = query.Where("start_datetime >= ? AND estimated_end_datetime <= ?", startTime, endTime)
	} else if startTime != "" {
		query = query.Where("start_datetime >= ?", startTime)
	} else if endTime != "" {
		query = query.Where("estimated_end_datetime <= ?", endTime)
	}

	// 6. Get total count using a session clone to avoid polluting the execution query
	var totalRows int64
	err = query.Session(&gorm.Session{}).Count(&totalRows).Error
	if err != nil {
		return nil, 0, err
	}

	// 7. Execute paginated query
	err = query.
		Order("start_datetime ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&rides).
		Error
	if err != nil {
		return nil, 0, err
	}

	// 8. Calculate total pages
	totalPages = int(math.Ceil(float64(totalRows) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	return rides, totalPages, nil
}

func getRideRequestByID(orgCtx *gin.Context, rideRequestID string) (postgress.RideRequest, error) {
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	var rideRequest postgress.RideRequest

	err := database.DatabaseConn.Postgres.
		WithContext(ctx).
		Where("is_active = ?", true).
		First(&rideRequest, "id = ?", rideRequestID).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rideRequest, fmt.Errorf("ride request not found")
		}
		return rideRequest, err
	}

	return rideRequest, nil
}

func mapPassengerData(request PassengerRegistrationRequest) postgress.Passenger {
	return postgress.Passenger{
		ID:              database.GenerateUUID(),
		PassengerName:   request.Name,
		PassengerMobile: request.MobileNumber,
		Password:        request.EXTPassword,
		Gender:          request.Gender.String(),
		Status:          constants.Status_PendingApproval,
	}
}

func mapPassengerFCMData(userId, fcm string) postgress.UserFCM {
	return postgress.UserFCM{
		ID:     database.GenerateUUID(),
		UserId: userId,
		FCM:    fcm,
	}
}

func savePassengerInfo(orgCtx *gin.Context, sessionId string, request PassengerRegistrationRequest) (passengerId string, err error) {
	logger.LogInfo("Request received in savePassengerInfo", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	excryptedPassword, err := utils.HashPassword(sessionId, request.Password)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Invalid_Data, "password")
		return
	}
	request.EXTPassword = excryptedPassword

	passenger := mapPassengerData(request)
	if err = database.DatabaseConn.Postgres.WithContext(ctx).Create(&passenger).Error; err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Registeration_Failed, "passenger")
		return
	}
	passengerId = passenger.ID

	logger.LogInfo("Response returned from savePassengerInfo", sessionId)

	return
}

func updatePassengerFCM(orgCtx *gin.Context, id, fcm string) error {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	return database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.UserFCM{}).
		Where("user_id = ?", id).
		Update("fcm", fcm).Error
}

// savePassengerSchedule writes the whole weekly form in one statement. The unique
// index on (passenger_id, day_of_week, direction) turns a resend of the same leg
// into an update instead of a duplicate row.
func savePassengerSchedule(orgCtx *gin.Context, sessionId, passengerId string, preferences []SchedulePreference) (err error) {
	logger.LogInfo("Request received in savePassengerSchedule", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	rows := make([]postgress.PassengerLocationPreference, 0, len(preferences))
	for _, preference := range preferences {
		rows = append(rows, postgress.PassengerLocationPreference{
			ID:            database.GenerateUUID(),
			PassengerID:   passengerId,
			DayOfWeek:     preference.DayOfWeek,
			Direction:     preference.Direction,
			IsEnabled:     preference.IsEnabled,
			Location:      strings.TrimSpace(preference.Location),
			Lat:           preference.Lat,
			Lng:           preference.Lng,
			ScheduledTime: preference.ScheduledTime,
		})
	}

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "passenger_id"}, {Name: "day_of_week"}, {Name: "direction"}},
			DoUpdates: clause.AssignmentColumns([]string{"is_enabled", "location", "lat", "lng", "scheduled_time", "updated_at"}),
		}).
		Create(&rows).Error

	logger.LogInfo("Response returned from savePassengerSchedule", sessionId)

	return
}

func getPassengerSchedule(orgCtx *gin.Context, passengerId string) (preferences []postgress.PassengerLocationPreference, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Where("passenger_id = ?", passengerId).
		Order("day_of_week ASC, direction ASC").
		Find(&preferences).Error

	return
}
