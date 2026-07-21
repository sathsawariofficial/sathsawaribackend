package passenger

import (
	"context"
	"fmt"
	"math"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

func bookRide(orgCtx *gin.Context, sessionId string, request BookSeatRequest) error {
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
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("invalid ride code, ride not found, or insufficient seats available")
	}

	if err := tx.Create(&postgress.RideBooking{
		ID:           database.GenerateUUID(),
		RideID:       request.RideId,
		MobileNumber: request.MobileNumber,
		Name:         request.Name,
		Seats:        request.Seats,
	}).Error; err != nil {
		logger.LogError(sessionId, err)
		tx.Rollback()
		return fmt.Errorf(constants.Registeration_Failed, "booking")
	}

	return tx.Commit().Error
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
