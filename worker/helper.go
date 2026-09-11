package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func closeActiveRides() {
	sessionId := constants.WROKER_SESSION
	logger.LogInfo("Request received in CloseActiveRides", sessionId)

	defer func() {
		if r := recover(); r != nil {
			logger.LogError(sessionId, fmt.Errorf("panic recovered: %v", r))
		}
	}()

	rides, err := getAllActiveRides()
	if err != nil {
		logger.LogError(sessionId, "failed to get ride error: "+err.Error())
		return
	}

	if len(rides) == 0 {
		logger.LogWarning(sessionId, "no ride found")
		return
	}

	pkt := time.FixedZone("PKT", 5*60*60)
	now := time.Now().In(pkt)

	for _, ride := range rides {
		endTime, err := time.ParseInLocation(
			"2006-01-02 15:04:05",
			strings.TrimSpace(ride.EstimatedEndDatetime),
			pkt,
		)
		if err != nil {
			logger.LogError(sessionId, fmt.Errorf("invalid datetime for ride %s: %v", ride.ID, err))
			continue
		}

		if !endTime.After(now) {
			logger.LogDebug2("closing ride", sessionId, ride.ID)

			if err := database.DatabaseConn.Postgres.
				Model(&postgress.Ride{}).
				Where("id = ?", ride.ID).
				Update("is_active", false).Error; err != nil {
				logger.LogError(sessionId, err)
			}

			err = redis.DeleteRedisValue(database.DatabaseConn.RedisConn, utils.GenerateShortCode((ride.ID)))
			if err != nil {
				logger.LogError(sessionId, err)
			}
		}
	}
}

func closeActiveRideRequests() {
	sessionId := constants.WROKER_SESSION
	logger.LogInfo("Request received in CloseActiveRideRequests", sessionId)

	defer func() {
		if r := recover(); r != nil {
			logger.LogError(sessionId, fmt.Errorf("panic recovered: %v", r))
		}
	}()

	rideRequests, err := getAllActiveRideRequest()
	if err != nil {
		logger.LogError(sessionId, "failed to get ride request error: "+err.Error())
		return
	}

	if len(rideRequests) == 0 {
		logger.LogWarning(sessionId, "no ride request found")
		return
	}

	pkt := time.FixedZone("PKT", 5*60*60)
	now := time.Now().In(pkt)

	for _, request := range rideRequests {
		endTime, err := time.ParseInLocation(
			"2006-01-02 15:04:05",
			strings.TrimSpace(request.EstimatedEndDatetime),
			pkt,
		)
		if err != nil {
			logger.LogError(sessionId, fmt.Errorf("invalid datetime for ride request %s: %v", request.ID, err))
			continue
		}

		if !endTime.After(now) {
			logger.LogDebug2("closing ride request", sessionId, request.ID)

			if err := database.DatabaseConn.Postgres.
				Model(&postgress.RideRequest{}).
				Where("id = ?", request.ID).
				Update("is_active", false).Error; err != nil {
				logger.LogError(sessionId, err)
			}
		}
	}
}

func getAllActiveRides() (rides []postgress.Ride, err error) {
	err = database.DatabaseConn.Postgres.
		Where(`is_active = ?`, true).
		Find(&rides).Error

	return
}

func getAllActiveRideRequest() (requests []postgress.RideRequest, err error) {
	err = database.DatabaseConn.Postgres.
		Where(`is_active = ?`, true).
		Find(&requests).Error

	return
}

func getAllRecurringRides() (rides []postgress.Ride, err error) {
	err = database.DatabaseConn.Postgres.
		Table("rides r").
		Where("r.is_active = ?", true).
		Where("r.is_recurring = ?", true).

		// 4-hour window condition
		Where(`
			r.start_datetime::timestamp BETWEEN NOW() AND NOW() + INTERVAL '4 hours'
		`).

		// Avoid duplicates for today
		Where(`
			NOT EXISTS (
				SELECT 1 FROM rides r2
				WHERE r2.driver_id = r.driver_id
				AND DATE(r2.start_datetime::timestamp) = CURRENT_DATE
			)
		`).
		Find(&rides).Error

	return
}

func saveNotification(notification redis.NotificationRequest) (err error) {
	notifyReq := &postgress.NotificationRequest{
		ID:               database.GenerateUUID(),
		UserId:           notification.UserId,
		UserType:         notification.UserType,
		Title:            notification.Title,
		Message:          notification.Message,
		NotificationType: notification.NotificationType,
	}

	if notification.Data != nil {
		var bVal []byte
		bVal, err = json.Marshal(notification.Data)
		if err != nil {
			logger.LogError(notifyReq.ID, err)
			return
		}

		notifyReq.Data = string(bVal)
	}

	err = database.DatabaseConn.Postgres.Create(notifyReq).Error
	if err != nil {
		logger.LogError(constants.WROKER_SESSION, fmt.Sprintf("Error: %s, Data: %v", err.Error(), notification))
	}
	return
}

func processBroadcastBatch(ginCtx *gin.Context, sessionId string, batchSize int) {
	var requests []postgress.BroadcastNotificationRequests

	ctx, cancel := context.WithTimeout(
		ginCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	err := database.DatabaseConn.Postgres.
		WithContext(ctx).
		Where("user_type = ? AND processed = false", constants.User_Driver).
		Order("created_at asc").
		Limit(batchSize).
		Find(&requests).Error

	if err != nil {
		logger.LogError(sessionId, "failed to fetch broadcast requests: "+err.Error())
		return
	}

	if len(requests) == 0 {
		return
	}

	for _, req := range requests {
		processBroadcastRequest(ginCtx, sessionId, req)

		// mark as processed (IMPORTANT to avoid reprocessing every tick)
		database.DatabaseConn.Postgres.
			Model(&postgress.BroadcastNotificationRequests{}).
			Where("id = ?", req.ID).
			Update("processed", true)
	}
}

func processBroadcastRequest(ctx *gin.Context, sessionId string, req postgress.BroadcastNotificationRequests) {

	switch req.UserType {

	case constants.User_Driver:
		sendDriverBroadcast(ctx, sessionId, req)

	default:
		logger.LogInfo("unimplemented user type in broadcast: "+fmt.Sprint(req.UserType), sessionId)
	}
}

func sendDriverBroadcast(ginCtx *gin.Context, sessionId string, req postgress.BroadcastNotificationRequests) {
	const driverBatchSize = 50
	offset := 0

	for {

		var drivers []postgress.Driver

		ctx, cancel := context.WithTimeout(
			ginCtx,
			time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
		)

		err := database.DatabaseConn.Postgres.
			WithContext(ctx).
			Limit(driverBatchSize).
			Offset(offset).
			Where(`status = ?`, constants.Status_Active).
			Find(&drivers).Error

		cancel()

		if err != nil {
			logger.LogError(sessionId, "failed to fetch drivers: "+err.Error())
			return
		}

		if len(drivers) == 0 {
			break
		}

		for _, driver := range drivers {
			utils.SendNotification(
				ginCtx,
				sessionId,
				req.NotificationType,
				driver.ID,
				req.Title,
				req.Message,
				map[string]string{
					"driverId":                driver.ID,
					"title":                   req.Title,
					constants.SMS_KEY_MESSAGE: req.Message,
				},
			)
		}

		offset += driverBatchSize
	}
}

// runShiftSchedule is one tick of the shift scheduler. Every step is a set based
// statement, however many shifts there are, and every step is safe to run twice.
func runShiftSchedule(ginCtx *gin.Context, sessionId string) {
	logger.LogInfo("Request received in runShiftSchedule", sessionId)

	defer func() {
		if r := recover(); r != nil {
			logger.LogError(sessionId, fmt.Errorf("panic recovered: %v", r))
		}
	}()

	db := database.DatabaseConn.Postgres
	today := database.BusinessToday()

	// trips are written before finished shifts are retired, and a day back as well as
	// a day ahead, so the last trip of a shift is still recorded even if the worker
	// was down when it ran
	if err := database.MaterializeOccurrences(db,
		database.AddDays(today, -constants.Occurrence_Backfill_Days),
		database.AddDays(today, constants.Occurrence_Horizon_Days), ""); err != nil {
		logger.LogError(sessionId, "failed to write the upcoming trips error: "+err.Error())
	}

	if err := db.Model(&postgress.Shift{}).
		Where("status = ?", constants.Shift_Status_Active).
		Where("end_date <> ''").
		Where("end_date < ?", today).
		Update("status", constants.Shift_Status_Completed).Error; err != nil {
		logger.LogError(sessionId, "failed to retire the finished shifts error: "+err.Error())
	}

	if err := db.Model(&postgress.ShiftOccurrence{}).
		Where("status = ?", constants.Occurrence_Status_Scheduled).
		Where("ends_at <= ?", database.BusinessNowMinute()).
		Update("status", constants.Occurrence_Status_Completed).Error; err != nil {
		logger.LogError(sessionId, "failed to complete the finished trips error: "+err.Error())
	}

	sendShiftReminders(ginCtx, sessionId)

	logger.LogInfo("Response returned from runShiftSchedule", sessionId)
}

// sendShiftReminders tells the driver and every passenger travelling that a trip
// starts within fifteen minutes. Each reminder is claimed by an atomic update of its
// own row before it is sent, the driver's on the occurrence and each passenger's on
// their attendance, so a retried tick or a second worker finds nothing left to claim
// and nobody is ever reminded twice for one trip. An absent passenger is not reminded.
func sendShiftReminders(ginCtx *gin.Context, sessionId string) {
	now := database.BusinessNow()
	nowMinute := now.Format(constants.Minute_Datetime_Layout)
	soon := now.Add(time.Duration(constants.Shift_Reminder_Minutes) * time.Minute).Format(constants.Minute_Datetime_Layout)

	type driverReminder struct {
		ID             string
		ShiftID        string
		ServiceID      string
		DriverID       string
		ShiftName      string
		ServiceName    string
		OccurrenceDate string
		StartTime      string
		VehicleNumber  string
		Route          string
	}

	var drivers []driverReminder
	if err := database.DatabaseConn.Postgres.Raw(`
		UPDATE shift_occurrences
		SET driver_reminder_at = now()
		WHERE status = ?
		  AND starts_at > ?
		  AND starts_at <= ?
		  AND driver_reminder_at IS NULL
		RETURNING id, shift_id, service_id, driver_id, shift_name, service_name, occurrence_date, start_time, vehicle_number, route
	`, constants.Occurrence_Status_Scheduled, nowMinute, soon).Scan(&drivers).Error; err != nil {
		logger.LogError(sessionId, "failed to claim the driver reminders error: "+err.Error())
	}

	type passengerReminder struct {
		PassengerID    string
		Location       string
		LocationTime   string
		ShiftID        string
		ServiceID      string
		ShiftName      string
		ServiceName    string
		OccurrenceDate string
		StartTime      string
		DriverName     string
		VehicleNumber  string
	}

	var passengers []passengerReminder
	if err := database.DatabaseConn.Postgres.Raw(`
		UPDATE shift_attendances
		SET reminder_at = now()
		FROM shift_occurrences
		WHERE shift_occurrences.id = shift_attendances.occurrence_id
		  AND shift_occurrences.status = ?
		  AND shift_occurrences.starts_at > ?
		  AND shift_occurrences.starts_at <= ?
		  AND shift_attendances.status = ?
		  AND shift_attendances.reminder_at IS NULL
		RETURNING shift_attendances.passenger_id, shift_attendances.location, shift_attendances.location_time,
			shift_occurrences.shift_id, shift_occurrences.service_id, shift_occurrences.shift_name,
			shift_occurrences.service_name, shift_occurrences.occurrence_date, shift_occurrences.start_time,
			shift_occurrences.driver_name, shift_occurrences.vehicle_number
	`, constants.Occurrence_Status_Scheduled, nowMinute, soon, constants.Attendance_Present).Scan(&passengers).Error; err != nil {
		logger.LogError(sessionId, "failed to claim the passenger reminders error: "+err.Error())
	}

	shiftName := func(name string) string {
		if strings.TrimSpace(name) == "" {
			return "unnamed shift"
		}
		return name
	}

	reminderData := func(shiftId, serviceId, date string) map[string]string {
		return map[string]string{
			constants.NOTIFICATION_KEY_SHIFT_ID:        shiftId,
			constants.NOTIFICATION_KEY_SERVICE_ID:      serviceId,
			constants.NOTIFICATION_KEY_OCCURRENCE_DATE: date,
		}
	}

	for _, reminder := range drivers {
		var stops []struct {
			Location string `json:"location"`
		}

		firstStop := ""
		if err := json.Unmarshal([]byte(reminder.Route), &stops); err == nil && len(stops) > 0 {
			firstStop = stops[0].Location
		}

		message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_REMINDER_DRIVER,
			shiftName(reminder.ShiftName), reminder.ServiceName, reminder.StartTime, firstStop, reminder.VehicleNumber)

		utils.SendUserNotification(ginCtx, sessionId, constants.NOTIFICATION_TYPE_SHIFT_REMINDER, reminder.DriverID, constants.User_Driver,
			constants.NOTIFICATION_TITLE_SHIFT_REMINDER, message, reminderData(reminder.ShiftID, reminder.ServiceID, reminder.OccurrenceDate))
	}

	for _, reminder := range passengers {
		message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SHIFT_REMINDER_RIDER,
			shiftName(reminder.ShiftName), reminder.ServiceName, reminder.StartTime,
			reminder.Location, reminder.LocationTime, reminder.DriverName, reminder.VehicleNumber)

		utils.SendUserNotification(ginCtx, sessionId, constants.NOTIFICATION_TYPE_SHIFT_REMINDER, reminder.PassengerID, constants.User_Passenger,
			constants.NOTIFICATION_TITLE_SHIFT_REMINDER, message, reminderData(reminder.ShiftID, reminder.ServiceID, reminder.OccurrenceDate))
	}

	if len(drivers)+len(passengers) > 0 {
		logger.LogDebug2("sent shift reminders", sessionId, fmt.Sprintf("drivers: %d, passengers: %d", len(drivers), len(passengers)))
	}
}
