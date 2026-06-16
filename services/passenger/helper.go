package passenger

import (
	"fmt"
	"rideshare/pkgs/database"
)

func IncrementSeatsTaken(rideID, code string) error {
	result := database.DatabaseConn.Postgres.Exec(`
		UPDATE rides
		SET seats_taken = seats_taken + 1
		WHERE id = ?
		  AND code = ?
		  AND is_active = true
		  AND seats_taken < number_of_seats
	`, rideID, code)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("invalid ride code, ride not found, or no seats available")
	}

	return nil
}
