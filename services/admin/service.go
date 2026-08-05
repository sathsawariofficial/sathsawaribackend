package admin

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"

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

func GetDrivers(ctx *gin.Context, sessionId string, page int) (driverDetails []postgress.Driver, totalRows int64, err error) {
	logger.LogInfo("Request received in GetDrivers", sessionId)

	driverDetails, totalRows, err = getAllDriversWithVehicles(ctx, page)
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

	vehicles, totalRows, err = getAllVehicles(ctx, page)
	if err != nil {
		logger.LogError(sessionId, " get driver details error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetVehicles", sessionId)
	logger.LogDebug2("Response returned from GetVehicles", sessionId, fmt.Sprintf("vehicles: %v, total rows: %v", vehicles, totalRows))

	return
}

func GetRides(ctx *gin.Context, sessionId string, page int) (rides []postgress.RideDetails, totalRows int64, err error) {
	logger.LogInfo("Request received in GetRides", sessionId)

	rides, totalRows, err = getAllRides(ctx, page)
	if err != nil {
		logger.LogError(sessionId, " get ride details error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetRides", sessionId)
	logger.LogDebug2("Response returned from GetRides", sessionId, fmt.Sprintf("rides: %v, total rows: %v", rides, totalRows))

	return
}

func DeleteDriver(ctx *gin.Context, sessionId string) (err error) {
	logger.LogInfo("Request received in DeleteDriver", sessionId)

	driverId := ctx.Query(constants.User_KEY)
	adminId := ctx.GetString(constants.User_KEY)
	logger.LogDebug("Driver to be deleted by admin", sessionId, driverId)

	driver, err := database.GetDriverById(ctx, driverId)
	if err != nil {
		logger.LogError(sessionId, "error failed to get driver: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "find driver")
		return
	}
	if utils.IsStringEmpty(driver.ID) {
		err = fmt.Errorf(constants.Failed_To_Do_Job, "find driver")
		logger.LogError(sessionId, err)
		return
	}

	if strings.EqualFold(driver.Status, constants.Status_InActive) {
		err = fmt.Errorf(constants.Failed_To_Do_Job, "driver, as it is already deleted")
		logger.LogError(sessionId, "error failed to delete driver: "+err.Error())
		return
	}

	err = database.DeleteDriver(ctx, driver, adminId)
	if err != nil {
		logger.LogError(sessionId, "error failed to delete driver status: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "finish operation")
		return
	}

	logger.LogInfo("Response returned from DeleteDriver", sessionId)

	return
}

func CreateAdminBroadcastRequest(ctx *gin.Context, sessionId string, request AdminBroadcastRequest) (err error) {
	logger.LogInfo("Request received in CreateAdminBroadcastRequest", sessionId)

	err = createAdminBroadcast(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, " create broadcast request: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from CreateAdminBroadcastRequest", sessionId)

	return
}

func CreateAnnouncement(ctx *gin.Context, sessionId string, request AnnouncementRequest) (err error) {
	logger.LogInfo("Request received in CreateAnnouncement", sessionId)

	err = createAnnouncement(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, " create broadcast request: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from CreateAnnouncement", sessionId)

	return
}

func GetApprochRequest(ctx *gin.Context, sessionId, approchType string, page int) (approches []postgress.ApprochInfo, totalRows int64, err error) {
	logger.LogInfo("Request received in GetApprochRequest", sessionId)
	logger.LogDebug("Request received in GetApprochRequest", sessionId, fmt.Sprintf("approch type: %s, page: %d", approchType, page))

	approches, totalRows, err = getApprochRequests(ctx, sessionId, approchType, page)
	if err != nil {
		logger.LogError(sessionId, " create broadcast request: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetApprochRequest", sessionId)
	logger.LogDebug("Response returned from GetApprochRequest", sessionId, totalRows)
	logger.LogDebug("Response returned from GetApprochRequest", sessionId, approches)

	return
}
