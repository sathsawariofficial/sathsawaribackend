package middleware

import (
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
)

func getDriverStatusById(driverId string) (string, error) {
	var driver postgress.Driver

	err := database.DatabaseConn.Postgres.
		Select("status").
		Where("id = ?", driverId).
		First(&driver).Error
	if err != nil {
		return "", err
	}

	return driver.Status, nil
}
