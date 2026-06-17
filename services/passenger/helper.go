package passenger

import (
	"context"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"time"

	"github.com/gin-gonic/gin"
)

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
		ID:           utils.GenerateUUID(),
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
