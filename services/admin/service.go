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

// GetPlatformOverview is the admin's first screen: how much of everything exists and
// how much of it is live.
func GetPlatformOverview(ctx *gin.Context, sessionId string) (overview postgress.PlatformOverview, err error) {
	logger.LogInfo("Request received in GetPlatformOverview", sessionId)

	overview, err = getPlatformOverview(ctx, sessionId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the overview error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the overview")
		return
	}

	logger.LogInfo("Response returned from GetPlatformOverview", sessionId)

	return
}

func GetPassengers(ctx *gin.Context, sessionId, search, status string, page int) (passengers []postgress.Passenger, totalRows int64, err error) {
	logger.LogInfo("Request received in GetPassengers", sessionId)

	passengers, totalRows, err = getAllPassengers(ctx, sessionId, search, status, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the passengers error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the passengers")
		return
	}

	logger.LogInfo("Response returned from GetPassengers", sessionId)

	return
}

// GetPassengerProfile is one passenger account. Pick & Drop is run by each service's
// owner and is not managed from the admin, so the profile carries the account only.
func GetPassengerProfile(ctx *gin.Context, sessionId, passengerId string) (passenger postgress.Passenger, err error) {
	logger.LogInfo("Request received in GetPassengerProfile", sessionId)

	passenger, err = getPassengerById(ctx, passengerId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the passenger error: "+err.Error())
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	logger.LogInfo("Response returned from GetPassengerProfile", sessionId)

	return
}

// UpdatePassengerStatus is the softer moderation tool: an account switched off
// cannot log in or be seated, but nothing it was part of is destroyed.
func UpdatePassengerStatus(ctx *gin.Context, sessionId, adminId, passengerId, status string) (err error) {
	logger.LogInfo("Request received in UpdatePassengerStatus", sessionId)

	if _, err = getPassengerById(ctx, passengerId); err != nil {
		logger.LogError(sessionId, "failed to get the passenger error: "+err.Error())
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	if err = updatePassengerStatus(ctx, sessionId, passengerId, status, adminId); err != nil {
		logger.LogError(sessionId, "failed to update the passenger error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "passenger")
		return
	}

	logger.LogInfo("Response returned from UpdatePassengerStatus", sessionId)

	return
}

// DeletePassenger removes an account for good. The shared helper takes it off its
// active shifts, ends its Pick & Drop membership and archives its availability and
// shift requests, so nothing anywhere is left pointing at somebody who no longer exists.
func DeletePassenger(ctx *gin.Context, sessionId, adminId, passengerId string) (err error) {
	logger.LogInfo("Request received in DeletePassenger", sessionId)

	passenger, err := getPassengerById(ctx, passengerId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the passenger error: "+err.Error())
		err = errors.New(constants.Passenger_Not_Found)
		return
	}

	if err = database.DeletePassenger(ctx, passenger, adminId); err != nil {
		logger.LogError(sessionId, "failed to delete the passenger error: "+err.Error())
		err = fmt.Errorf(constants.DELETE_Failed, "passenger")
		return
	}

	logger.LogInfo("Response returned from DeletePassenger", sessionId)

	return
}

func UpdateDriverStatus(ctx *gin.Context, sessionId, adminId, driverId, status string) (err error) {
	logger.LogInfo("Request received in UpdateDriverStatus", sessionId)

	driver, err := database.GetDriverById(ctx, driverId)
	if err != nil || utils.IsStringEmpty(driver.ID) {
		logger.LogError(sessionId, "failed to get the driver error")
		err = errors.New(constants.Driver_Not_Found)
		return
	}

	if err = updateDriverStatus(ctx, sessionId, driverId, status, adminId); err != nil {
		logger.LogError(sessionId, "failed to update the driver error: "+err.Error())
		err = fmt.Errorf(constants.Update_Failed, "driver")
		return
	}

	logger.LogInfo("Response returned from UpdateDriverStatus", sessionId)

	return
}

////////////////////////////// PICK & DROP OVERSIGHT //////////////////////////////

func GetPickDropServices(ctx *gin.Context, sessionId, search, status string, page int) (services []postgress.PickDropServiceDetails, totalRows int64, err error) {
	logger.LogInfo("Request received in GetPickDropServices", sessionId)

	if services, totalRows, err = getAllPickDropServices(ctx, search, status, page); err != nil {
		logger.LogError(sessionId, "failed to get the services error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the services")
		return
	}

	logger.LogInfo("Response returned from GetPickDropServices", sessionId)

	return
}

func GetPickDropServiceDetail(ctx *gin.Context, sessionId, serviceId string) (
	service postgress.PickDropServiceDetails,
	counts pickDropServiceCountsRow,
	shifts []postgress.ShiftDetails,
	err error,
) {
	logger.LogInfo("Request received in GetPickDropServiceDetail", sessionId)

	service, counts, shifts, err = getPickDropServiceAdminDetail(ctx, serviceId)
	if err != nil || utils.IsStringEmpty(service.ID) {
		logger.LogError(sessionId, "failed to get the service error")
		err = errors.New(constants.Service_Not_Found)
		return
	}

	logger.LogInfo("Response returned from GetPickDropServiceDetail", sessionId)

	return
}

func GetShifts(ctx *gin.Context, sessionId string, filter adminShiftFilter, page int) (shifts []postgress.ShiftDetails, totalRows int64, err error) {
	logger.LogInfo("Request received in GetShifts", sessionId)

	if shifts, totalRows, err = listShiftsAdmin(ctx, filter, page); err != nil {
		logger.LogError(sessionId, "failed to get the shifts error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shifts")
		return
	}

	logger.LogInfo("Response returned from GetShifts", sessionId)

	return
}

func GetShiftRequests(ctx *gin.Context, sessionId string, filter adminShiftFilter, page int) (requests []postgress.ShiftRequestDetails, totalRows int64, err error) {
	logger.LogInfo("Request received in GetShiftRequests", sessionId)

	if requests, totalRows, err = listShiftRequestsAdmin(ctx, filter, page); err != nil {
		logger.LogError(sessionId, "failed to get the shift requests error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the shift requests")
		return
	}

	logger.LogInfo("Response returned from GetShiftRequests", sessionId)

	return
}

func GetAdvertisements(ctx *gin.Context, sessionId, search string, page int) (ads []postgress.AdvertisementDetails, totalRows int64, err error) {
	logger.LogInfo("Request received in GetAdvertisements", sessionId)

	if ads, totalRows, err = listAdvertisementsAdmin(ctx, search, page); err != nil {
		logger.LogError(sessionId, "failed to get the advertisements error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the advertisements")
		return
	}

	logger.LogInfo("Response returned from GetAdvertisements", sessionId)

	return
}
