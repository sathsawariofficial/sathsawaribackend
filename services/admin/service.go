package admin

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
)

func LoginAdmin(ctx *gin.Context, sessionId string, request AdminLoginRequest) (token string, admin postgress.Admin, err error) {
	logger.LogInfo("Request returned from LoginAdmin", sessionId)

	admin, err = getAdmin(ctx, request.Username)
	if err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Login_Failed)
		return
	}
	if utils.IsStringEmpty(admin.ID) {
		logger.LogError(sessionId, "admin not found error")
		err = fmt.Errorf(constants.General_Unknown, "admin")
		return
	}

	token, err = loginTokenCreation(sessionId, request, admin)
	if err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Login_Failed)
		return
	}

	logger.LogInfo("Response returned from LoginAdmin", sessionId)

	return
}

func GetDrivers(ctx *gin.Context, sessionId string, page int) (driverDetails []postgress.DriverWithVehicle, totalRows int64, err error) {
	logger.LogInfo("Request received in GetDrivers", sessionId)

	driverDetails, totalRows, err = getAllActiveDriversWithVehicles(ctx, page)
	if err != nil {
		logger.LogError(sessionId, " get driver details error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetDrivers", sessionId)
	logger.LogDebug2("Response returned from GetDrivers", sessionId, fmt.Sprintf("driver details: %v, total rows: %v", driverDetails, totalRows))

	return
}

func GetVehicles(ctx *gin.Context, sessionId string, page int) (vehicles []postgress.Vehicle, totalRows int64, err error) {
	logger.LogInfo("Request received in GetVehicles", sessionId)

	vehicles, totalRows, err = getAllActiveVehicles(ctx, page)
	if err != nil {
		logger.LogError(sessionId, " get driver details error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetVehicles", sessionId)
	logger.LogDebug2("Response returned from GetVehicles", sessionId, fmt.Sprintf("vehicles: %v, total rows: %v", vehicles, totalRows))

	return
}
