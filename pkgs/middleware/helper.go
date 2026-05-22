package middleware

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
)

func checkUserStatusById(userId, tokenType string) error {
	if tokenType == constants.ADMIN_TOKEN {
		var admin postgress.Admin

		err := database.DatabaseConn.Postgres.
			Where("id = ?", userId).
			First(&admin).Error
		if err != nil {
			logger.LogError(Middleware_Session, err)
			err = errors.New(constants.Operation_Not_Permitted)
			return err
		}

		if utils.IsStringEmpty(admin.ID) {
			err = fmt.Errorf("invalid credentails")
			logger.LogError(Middleware_Session, err)
		}

		return nil
	} else {
		var driver postgress.Driver

		err := database.DatabaseConn.Postgres.
			Select("status").
			Where("id = ?", userId).
			First(&driver).Error
		if err != nil {
			logger.LogError(Middleware_Session, err)
			err = errors.New(constants.Operation_Not_Permitted)
			return err
		}

		return nil
	}
}
