package ride

import (
	"context"
	"errors"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateRide(ctx *gin.Context, sessionId string, request RideCreationRequest) (rideId string, err error) {
	logger.LogInfo("Request received in CreateRide", sessionId)

	vehicle, err := getVehicleById(ctx, request.VehicleId)
	if err != nil {
		logger.LogError(sessionId, "failed to get vehicle error: "+err.Error())
		err = errors.New(constants.Vehicle_Not_Found)
		return
	}
	if utils.IsStringEmpty(vehicle.ID) {
		logger.LogError(sessionId, "failed to get vehicle error")
		err = errors.New(constants.Vehicle_Not_Found)
		return
	}

	hasRide, err := VehicleHasRideDuringTime(request.VehicleId, request.StartDatetime, request.EstimatedEndDatetime)
	if err != nil {
		logger.LogError(sessionId, "failed to determine if vehicle already has a ride error: "+err.Error())
		err = errors.New("Vehicle might already have a ride scheduled for this duration")
		return
	}
	if hasRide {
		err = errors.New("Vehicle already has a ride scheduled for this duration")
		logger.LogError(sessionId, err)
		return
	}

	// Save the ride in the database
	ride := mapRideData(request, "", request.EXTDriverId, vehicle.ID)
	if err = database.DatabaseConn.Postgres.Create(&ride).Error; err != nil {
		logger.LogError(sessionId, "failed to create ride error: "+err.Error())
		err = fmt.Errorf(constants.Creation_Failed, "ride")
		return
	}
	rideId = ride.ID

	if request.IsRecurring {
		logger.LogDebug("recurring ride", sessionId, fmt.Sprintf("frequency: %d, period: %d", request.Frequency, request.Period))

		err = handleRecurring(ctx, rideId, rideId, request.VehicleId, request)
		if err != nil {
			logger.LogError(sessionId, "failed to create recurring rides ride error: "+err.Error())
			err = fmt.Errorf(constants.Creation_Failed, "recurring rides")
			return
		}
	}

	if request.MakeTemplate {
		logger.LogInfo("make template", sessionId)
		template := mapRideToRideTemplateData(request, rideId, request.EXTDriverId, vehicle.ID)
		logger.LogDebug("template", sessionId, template)

		if err = database.DatabaseConn.Postgres.Create(&template).Error; err != nil {
			logger.LogWarning(sessionId, "failed to create ride template error: "+err.Error())
			err = nil
			return
		}
	}

	utils.SendNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_RIDE_CREATED, request.EXTDriverId, constants.NOTIFICATION_TITLE_RIDE_CREATION, constants.NOTIFICATION_MESSAGE_RIDE_CREATION, nil)

	logger.LogInfo("Response returned from CreateRide", sessionId)
	logger.LogDebug2("Response returned from CreateRide", sessionId, rideId)

	return
}

func handleRecurring(ctx *gin.Context, sessionId, rideId, vehicleId string, request RideCreationRequest) (err error) {
	logger.LogInfo("Request received in handleRecurring", sessionId)

	startDatetime, err := utils.ConvertStrToTime(request.StartDatetime)
	if err != nil {
		logger.LogError(sessionId, err)
		return err
	}

	endDatetime, err := utils.ConvertStrToTime(request.EstimatedEndDatetime)
	if err != nil {
		logger.LogError(sessionId, err)
		return err
	}

	frequency := findFrequency(request.Frequency, request.Period)

	logger.LogDebug("frequency iterations", sessionId, fmt.Sprintf("frequency: %d, period: %d, total iterations: %d", request.Frequency, request.Period, frequency))

	if request.Period == WEEKLY {
		// as we want to start adding recurring payments from day after startdate
		dayOfWeek := int(startDatetime.Weekday()) + 1
		shift := 1
		for i := 1; i <= frequency; i++ {
			rideRequest := request

			// keep day between 1-7 [only create ride on days user has requested]
			day := ((dayOfWeek - 1) % 7) + 1
			// [day of the week based on start date]
			dayOfWeek++
			if !utils.InSlice(request.DaysOfWeek, day) {
				// [number of days to be added since the start date]
				shift++
				continue
			}

			currentStart, currentEnd := shiftDailyDates(startDatetime, endDatetime, shift)

			rideRequest.StartDatetime = currentStart.Format(constants.DateTimeLayout)
			rideRequest.EstimatedEndDatetime = currentEnd.Format(constants.DateTimeLayout)
			// [number of days to be added since the start date]
			shift++

			ride := mapRideData(rideRequest, rideId, rideRequest.EXTDriverId, vehicleId)

			if err = database.DatabaseConn.Postgres.Create(&ride).Error; err != nil {
				logger.LogError(sessionId, "failed to create ride error: "+err.Error())
				return fmt.Errorf(constants.Creation_Failed, "ride")
			}
		}
	} else {
		for i := 1; i <= frequency; i++ {
			rideRequest := request

			step := i
			// add i-th day from start date
			currentStart, currentEnd := shiftDailyDates(startDatetime, endDatetime, step)

			if request.Period == MONTHLY {
				currentStart, currentEnd = shiftMonthlyDates(startDatetime, endDatetime, step)
			}

			rideRequest.StartDatetime = currentStart.Format(constants.DateTimeLayout)
			rideRequest.EstimatedEndDatetime = currentEnd.Format(constants.DateTimeLayout)

			ride := mapRideData(
				rideRequest,
				rideId,
				rideRequest.EXTDriverId,
				vehicleId,
			)

			if err = database.DatabaseConn.Postgres.Create(&ride).Error; err != nil {
				logger.LogError(sessionId, "failed to create ride error: "+err.Error())
				return fmt.Errorf(constants.Creation_Failed, "ride")
			}
		}
	}

	logger.LogInfo("Response returned from handleRecurring", sessionId)

	return
}

func GetFilteredRides(ctx *gin.Context, sessionId string, page int, startTime, endTime, searchLoc string) (rides []postgress.RideDetails, totalRows int64, err error) {
	logger.LogInfo("Request received in GetFilteredRides", sessionId)

	rides, totalRows, err = getFilteredRides(ctx, page, startTime, endTime, searchLoc)
	if err != nil {
		logger.LogError(sessionId, " get filtered rides error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
		return
	}

	logger.LogInfo("Response returned from GetFilteredRides", sessionId)
	logger.LogDebug2("Response returned from GetFilteredRides", sessionId, fmt.Sprintf("rides: %v, total rows: %v", rides, totalRows))

	return
}

// get driver's ride based on status
func DriverRide(ctx *gin.Context, sessionId string, ride_status, driverId, startTime, endTime, startLoc, endLoc string) (totalPages int, rides []postgress.RideDetails, err error) {
	logger.LogInfo("Request received in DriverRide", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		return
	}

	rides, totalPages, err = getAllRidesByDriver(ctx, page, driverId, startTime, endTime, startLoc, endLoc, ride_status)
	if err != nil {
		err = errors.New(constants.Unknown_Error)
		return
	}

	if len(rides) == 0 {
		logger.LogInfo(sessionId, "no rides found")
		err = errors.New(constants.Ride_Not_Found)
		return
	}

	logger.LogInfo("Response returned from DriverRide", sessionId)
	logger.LogDebug2("Response returned from DriverRide", sessionId, fmt.Sprintf("rides: %v, total rows: %v", rides, totalPages))

	return
}

func UpdateRide(orgCtx *gin.Context, sessionId, rideId string, request UpdateRideRequest) error {
	logger.LogInfo("Request received in UpdateRide", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	tx := database.DatabaseConn.Postgres.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ride postgress.Ride
	if err := tx.Where("id = ?", rideId).First(&ride).Error; err != nil {
		tx.Rollback()
		logger.LogError(sessionId, err)
		return errors.New(constants.Ride_Not_Found)
	}

	// Archive and delete if ride is being deactivated
	if request.Status != nil && strings.EqualFold(*request.Status, constants.Ride_Status_InActive) {
		delRide := postgress.DELRide{
			ID:                   ride.ID,
			DriverID:             ride.DriverID,
			VehicleID:            ride.VehicleID,
			StartDatetime:        ride.StartDatetime,
			EstimatedEndDatetime: ride.EstimatedEndDatetime,
			NumberOfSeats:        ride.NumberOfSeats,
			SeatsTaken:           ride.SeatsTaken,
			StartLocation:        ride.StartLocation,
			EndLocation:          ride.EndLocation,
			RoutePoints:          ride.RoutePoints,
			Fare:                 ride.Fare,
			RouteDetails:         ride.RouteDetails,
			IsActive:             false,
			ParentRideId:         ride.ParentRideId,
			Code:                 ride.Code,
			CreatedAt:            ride.CreatedAt,
			UpdatedAt:            time.Now(),
		}

		if err := tx.Create(&delRide).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return errors.New(constants.General_Error)
		}

		if err := tx.Delete(&postgress.Ride{}, "id = ?", ride.ID).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return errors.New(constants.General_Error)
		}

		if err := tx.Commit().Error; err != nil {
			logger.LogError(sessionId, err)
			return errors.New(constants.Unknown_Error)
		}

		logger.LogInfo("Response returned from UpdateRide", sessionId)
		return nil
	}

	// Normal updates
	updates := map[string]interface{}{}

	if request.Status != nil {
		updates["is_active"] = strings.EqualFold(*request.Status, constants.Ride_Status_Active)
	}

	if request.NumberOfSeats > 0 {
		updates["number_of_seats"] = request.NumberOfSeats
	}

	if len(updates) > 0 {
		if err := tx.Model(&ride).Updates(updates).Error; err != nil {
			tx.Rollback()
			logger.LogError(sessionId, err)
			return errors.New(constants.General_Error)
		}
	}

	if err := tx.Commit().Error; err != nil {
		logger.LogError(sessionId, err)
		return errors.New(constants.Unknown_Error)
	}

	logger.LogInfo("Response returned from UpdateRide", sessionId)
	return nil
}

func BookSeatByDriver(ctx *gin.Context, sessionId string, request BookSeatRequest) (err error) {
	logger.LogInfo("Request returned from BookSeatByDriver", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	bookedRide, err := bookSeatByDriver(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, err)
		return
	}

	logger.LogInfo("Response returned from BookSeatByDriver", sessionId)
	logger.LogDebug2("Response returned from BookSeatByDriver", sessionId, fmt.Sprintf("booked ride Id: %v", bookedRide.ID))

	return
}

func GetBookedSeats(ctx *gin.Context, sessionId, rideId string) (bookSeats []postgress.RidePassenger, err error) {
	logger.LogInfo("Request returned from GetBookedSeats", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	ride, err := getRideById(ctx, rideId)
	if err != nil {
		logger.LogError(sessionId, err)
		err = errors.New(constants.Ride_Not_Found)
		return
	}

	if ride.DriverID != driverId {
		err = errors.New(constants.Operation_Not_Permitted)
		logger.LogError(sessionId, "driver is not the owner of this ride error: "+err.Error())
		return
	}

	bookSeats, err = getBookedSeatsByRideId(ctx, rideId)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the booked seats")
		return
	}

	logger.LogInfo("Response returned from GetBookedSeats", sessionId)
	logger.LogDebug2("Response returned from GetBookedSeats", sessionId, fmt.Sprintf("booked seats: %v", bookSeats))

	return
}

func UpdateBookedSeat(ctx *gin.Context, sessionId, ridePassengerId string) (err error) {
	logger.LogInfo("Request returned from UpdateBookedSeat", sessionId)

	err = updateStatusBookedSeatsById(ctx, ridePassengerId)
	if err != nil {
		logger.LogError(sessionId, "failed to cancel booked seat error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "cancel the seat")

		return
	}

	logger.LogInfo("Response returned from UpdateBookedSeat", sessionId)

	return
}

func GetRide(ctx *gin.Context, sessionId, vehicleId string) (ride postgress.Ride, err error) {
	logger.LogInfo("Request returned from GetRide", sessionId)
	logger.LogDebug("Request returned from GetRide", sessionId, vehicleId)

	ride, err = getRide(ctx, sessionId, vehicleId)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get template")
		return
	}

	logger.LogInfo("Response returned from GetRide", sessionId)
	logger.LogDebug("Response returned from GetRide", sessionId, ride)

	return
}

func GetRideTemplates(ctx *gin.Context, sessionId string) (rideTemplates []postgress.RideTemplate, err error) {
	logger.LogInfo("Request returned from GetRideTemplates", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	logger.LogDebug("Request returned from GetRideTemplates", sessionId, driverId)

	rideTemplates, err = getRideTemplates(ctx, sessionId, driverId)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get template")
		return
	}

	logger.LogInfo("Response returned from GetRideTemplates", sessionId)
	logger.LogDebug("Response returned from GetRideTemplates", sessionId, rideTemplates)

	return
}

func DeleteRideTemplate(ctx *gin.Context, sessionId, rideTemplateId string) (err error) {
	logger.LogInfo("Request returned from DeleteRideTemplate", sessionId)

	err = deleteRideTemplate(ctx, sessionId, rideTemplateId)
	if err != nil {
		logger.LogError(sessionId, err)
		err = fmt.Errorf(constants.Failed_To_Do_Job, "delete template")
		return
	}

	logger.LogInfo("Response returned from DeleteRideTemplate", sessionId)

	return
}
