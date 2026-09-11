package admin

import (
	"context"
	"errors"
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
	"gorm.io/gorm"
)

func getAdmin(orgCtx *gin.Context, mobile string) (admin postgress.Admin, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where(`username = ?`, mobile).Find(&admin).Error
	return
}

func loginTokenCreation(sessionId string, request AdminLoginRequest, admin postgress.Admin) (token string, err error) {
	if err = utils.ComparePassword(admin.Password, request.Password); err != nil {
		logger.LogError(sessionId, "invalid password error")
		err = errors.New(constants.Invalid_Password)
		return
	}

	excryptedAdminId, err := utils.EncryptAES(sessionId, admin.ID)
	if err != nil {
		logger.LogError(sessionId, "failed to set session error: "+err.Error())
		err = errors.New(constants.Login_Failed)
		return
	}

	err = redis.DeleteRedisValue(database.DatabaseConn.RedisConn, excryptedAdminId)
	if err != nil {
		logger.LogError(sessionId, "trying to delete session before new login error: "+err.Error())
	}

	token, err = utils.CreateJWT(sessionId, map[string]string{
		"adminId":   excryptedAdminId,
		"tokenType": constants.ADMIN_TOKEN,
	})
	if err != nil {
		logger.LogError(sessionId, "failed to created jwt token error: "+err.Error())
		err = errors.New(constants.Login_Failed)
		return
	}

	err = redis.SetRedisValueTTL(database.DatabaseConn.RedisConn, excryptedAdminId, token, time.Duration(configuration.ConfigurationData.Auth.ExpirationTime)*time.Second)
	if err != nil {
		logger.LogError(sessionId, "session deleted error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, "login the admin")
		return
	}

	return
}

func getAllRides(orgCtx *gin.Context, page int) (rides []postgress.RideDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Table("rides").
		Select(`
			rides.id,
			rides.driver_id,
			drivers.driver_name,
			drivers.driver_mobile,
			vehicles.vehicle_number,
			vehicles.vehicle_info,
			rides.start_datetime,
			rides.estimated_end_datetime,
			rides.number_of_seats,
			rides.seats_taken,
			rides.start_location,
			rides.end_location,
			rides.fare,
			rides.route_details,
			rides.is_active,
			rides.created_at,
			rides.updated_at
		`).
		Joins("JOIN drivers ON rides.driver_id = drivers.id").
		Joins("JOIN vehicles ON rides.vehicle_id = vehicles.id")

	// counted before Limit/Offset are applied, on a clean clone, or this would never
	// report more than one page
	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Order("created_at desc").Limit(pageSize).Offset(offset).Find(&rides).Error

	return
}

func getAllVehicles(orgCtx *gin.Context, page int) (vehicles []postgress.Vehicle, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).
		Table("vehicles").
		Select(`
			id,
			driver_id,
			vehicle_number,
			vehicle_info,
			number_of_seats,
			has_ac,
			has_heating,
			status,
			created_at,
			updated_at
		`)

	// counted on a clean clone, the old count ran against the limited query and so
	// could never report more than one page
	countQuery := query.Session(&gorm.Session{})
	if err = countQuery.Count(&totalRows).Error; err != nil {
		return
	}

	err = query.
		Order("created_at desc").
		Limit(pageSize).
		Offset(offset).
		Find(&vehicles).Error

	return
}

func getAllDriversWithVehicles(orgCtx *gin.Context, page int) (drivers []postgress.Driver, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	// Step 1: Count active drivers
	err = db.Model(&postgress.Driver{}).
		Where("status = ?", constants.Status_Active).
		Count(&totalRows).Error
	if err != nil {
		return
	}

	// Step 2: Fetch drivers with vehicles
	err = db.
		Preload("Vehicles").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&drivers).Error

	if err != nil {
		return
	}

	return
}

func createAdminBroadcastRequest(req AdminBroadcastRequest) postgress.BroadcastNotificationRequests {
	return postgress.BroadcastNotificationRequests{
		ID:               database.GenerateUUID(),
		UserType:         req.UserType,
		Title:            req.Title,
		Message:          req.Message,
		NotificationType: req.NotificationType,
	}
}

func createAnnouncementRequest(req AnnouncementRequest) postgress.AnnouncementRequests {
	return postgress.AnnouncementRequests{
		ID:      database.GenerateUUID(),
		Title:   req.Title,
		Message: req.Message,
		Type:    string(req.Type),
	}
}

func createAdminBroadcast(orgCtx *gin.Context, sessionId string, request AdminBroadcastRequest) error {
	logger.LogInfo("Request received in createAdminBroadcast", sessionId)

	broadcastReq := createAdminBroadcastRequest(request)
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)
	err := db.Create(&broadcastReq).Error
	if err != nil {
		logger.LogError(sessionId, err)
		return err
	}

	logger.LogInfo("Response returned from createAdminBroadcast", sessionId)
	return nil
}

func createAnnouncement(orgCtx *gin.Context, sessionId string, request AnnouncementRequest) error {
	logger.LogInfo("Request received in createAnnouncement", sessionId)

	broadcastReq := createAnnouncementRequest(request)
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)
	err := db.Create(&broadcastReq).Error
	if err != nil {
		logger.LogError(sessionId, err)
		return err
	}

	logger.LogInfo("Response returned from createAnnouncement", sessionId)
	return nil
}

func getApprochRequests(orgCtx *gin.Context, sessionId, approchType string, page int) (approchs []postgress.ApprochInfo, totalRows int64, err error) {
	logger.LogInfo("Request received in getApprochRequests", sessionId)
	logger.LogDebug("Request received in getApprochRequests", sessionId, fmt.Sprintf("approch type: %s, page: %d", approchType, page))

	pageSize := configuration.ConfigurationData.PageSize
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	// Base query (shared for count + fetch)
	baseQuery := db.Model(&postgress.ApprochInfo{})

	if !utils.IsStringEmpty(approchType) {
		baseQuery = baseQuery.Where("type = ?", approchType)
	}

	// Step 1: Count total rows
	err = baseQuery.Count(&totalRows).Error
	if err != nil {
		logger.LogError(sessionId, err)
		return
	}

	// Step 2: Fetch paginated data
	query := baseQuery.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset)

	err = query.Find(&approchs).Error
	if err != nil {
		logger.LogError(sessionId, err)
		return
	}

	logger.LogInfo("Response returned from getApprochRequests", sessionId)
	logger.LogDebug("Response returned from getApprochRequests", sessionId, approchs)

	return
}

func createAccouncement(orgCtx *gin.Context, sessionId string, request AdminBroadcastRequest) error {
	logger.LogInfo("Request received in createAccouncement", sessionId)

	broadcastReq := createAdminBroadcastRequest(request)
	ctx, cancel := context.WithTimeout(
		orgCtx,
		time.Duration(configuration.ConfigurationData.Timeout)*time.Second,
	)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)
	err := db.Create(&broadcastReq).Error
	if err != nil {
		logger.LogError(sessionId, err)
		return err
	}

	logger.LogInfo("Response returned from createAccouncement", sessionId)
	return nil
}

// getPlatformOverview counts the whole product in a single round trip rather than a
// dozen, since every number on the admin's first screen is just a count.
func getPlatformOverview(orgCtx *gin.Context, sessionId string) (overview postgress.PlatformOverview, err error) {
	logger.LogInfo("Request received in getPlatformOverview", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM drivers)                            AS total_drivers,
			(SELECT COUNT(*) FROM drivers WHERE status = ?)           AS active_drivers,
			(SELECT COUNT(*) FROM passengers)                         AS total_passengers,
			(SELECT COUNT(*) FROM passengers WHERE status = ?)        AS active_passengers,
			(SELECT COUNT(*) FROM vehicles)                           AS total_vehicles,
			(SELECT COUNT(*) FROM vehicles WHERE number_of_seats > 0) AS seated_vehicles,
			(SELECT COUNT(*) FROM rides)                              AS total_rides,
			(SELECT COUNT(*) FROM rides WHERE is_active = true)       AS active_rides,
			(SELECT COUNT(*) FROM pick_drop_services)                 AS total_services,
			(SELECT COUNT(*) FROM pick_drop_services WHERE status = ?) AS active_services,
			(SELECT COUNT(*) FROM shifts)                             AS total_shifts,
			(SELECT COUNT(*) FROM shifts WHERE status = ?)            AS active_shifts,
			(SELECT COUNT(*) FROM pick_drop_advertisements)           AS total_advertisements,
			(SELECT COUNT(*) FROM shift_requests)                     AS open_shift_requests,
			(SELECT COUNT(*) FROM pick_drop_drivers WHERE status = ?)
				+ (SELECT COUNT(*) FROM pick_drop_vehicles WHERE status = ?)
				+ (SELECT COUNT(*) FROM pick_drop_passengers WHERE status = ?) AS pending_join_requests
	`,
		constants.Status_Active,
		constants.Status_Active,
		constants.Status_Active,
		constants.Shift_Status_Active,
		constants.Membership_Status_Pending,
		constants.Membership_Status_Pending,
		constants.Membership_Status_Pending,
	).Scan(&overview).Error

	logger.LogInfo("Response returned from getPlatformOverview", sessionId)

	return
}

// getAllPassengers lists passenger accounts, optionally narrowed by status or by a
// search across the name and the mobile number.
func getAllPassengers(orgCtx *gin.Context, sessionId, search, status string, page int) (passengers []postgress.Passenger, totalRows int64, err error) {
	logger.LogInfo("Request received in getAllPassengers", sessionId)

	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Model(&postgress.Passenger{})

	if !utils.IsStringEmpty(status) {
		query = query.Where("status = ?", status)
	}

	if !utils.IsStringEmpty(search) {
		term := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("passenger_name ILIKE ? OR passenger_mobile ILIKE ?", term, term)
	}

	countQuery := query.Session(&gorm.Session{})
	if err = countQuery.Count(&totalRows).Error; err != nil {
		logger.LogError(sessionId, err)
		return
	}

	err = query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&passengers).Error

	logger.LogInfo("Response returned from getAllPassengers", sessionId)

	return
}

func getPassengerById(orgCtx *gin.Context, passengerId string) (passenger postgress.Passenger, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where("id = ?", passengerId).First(&passenger).Error
	return
}

func updatePassengerStatus(orgCtx *gin.Context, sessionId, passengerId, status, updatedBy string) (err error) {
	logger.LogInfo("Request received in updatePassengerStatus", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Passenger{}).
		Where("id = ?", passengerId).
		Updates(map[string]interface{}{
			"status":    status,
			"update_by": updatedBy,
		}).Error

	logger.LogInfo("Response returned from updatePassengerStatus", sessionId)

	return
}

func updateDriverStatus(orgCtx *gin.Context, sessionId, driverId, status, updatedBy string) (err error) {
	logger.LogInfo("Request received in updateDriverStatus", sessionId)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).
		Model(&postgress.Driver{}).
		Where("id = ?", driverId).
		Updates(map[string]interface{}{
			"status":    status,
			"update_by": updatedBy,
		}).Error

	logger.LogInfo("Response returned from updateDriverStatus", sessionId)

	return
}

////////////////////////////// PICK & DROP OVERSIGHT //////////////////////////////
// Read only. services/admin never imports services/pickdrop or services/shift, so
// these queries are written again here against the shared postgres models rather
// than reusing the owner-scoped ones those packages keep to themselves.

// getAllPickDropServices lists every Pick & Drop service on the platform, active or
// disabled, with the same roster counts an owner sees on their own service.
func getAllPickDropServices(orgCtx *gin.Context, search, status string, page int) (services []postgress.PickDropServiceDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).
		Table("pick_drop_services").
		Joins("LEFT JOIN drivers ON drivers.id = pick_drop_services.owner_driver_id")

	if !utils.IsStringEmpty(status) {
		query = query.Where("pick_drop_services.status = ?", status)
	}

	if !utils.IsStringEmpty(search) {
		term := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("pick_drop_services.name ILIKE ? OR drivers.driver_name ILIKE ? OR drivers.driver_mobile ILIKE ?", term, term, term)
	}

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Select(`
			pick_drop_services.id,
			pick_drop_services.name,
			pick_drop_services.description,
			pick_drop_services.owner_driver_id,
			COALESCE(drivers.driver_name, '') AS owner_name,
			COALESCE(drivers.driver_mobile, '') AS owner_mobile,
			pick_drop_services.status,
			(SELECT COUNT(*) FROM pick_drop_drivers d WHERE d.service_id = pick_drop_services.id AND d.status = ?) AS driver_count,
			(SELECT COUNT(*) FROM pick_drop_vehicles v WHERE v.service_id = pick_drop_services.id AND v.status = ?) AS vehicle_count,
			(SELECT COUNT(*) FROM pick_drop_passengers p WHERE p.service_id = pick_drop_services.id AND p.status = ?) AS passenger_count,
			(SELECT COUNT(*) FROM shifts s WHERE s.service_id = pick_drop_services.id AND s.status = ?) AS shift_count,
			pick_drop_services.created_at
		`,
		constants.Membership_Status_Approved,
		constants.Membership_Status_Approved,
		constants.Membership_Status_Approved,
		constants.Shift_Status_Active,
	).
		Order("pick_drop_services.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&services).Error

	return
}

// pickDropServiceCountsRow is the same roster breakdown an owner's own dashboard
// works out, counted in one round trip.
type pickDropServiceCountsRow struct {
	ApprovedDrivers       int64
	ApprovedVehicleOwners int64
	ApprovedVehicles      int64
	ApprovedPassengers    int64
	PendingRequests       int64
	ActiveShifts          int64
}

// getPickDropServiceAdminDetail reads one service, whoever owns it and whatever its
// status, with its roster counts and every shift it has ever had that is still active.
func getPickDropServiceAdminDetail(orgCtx *gin.Context, serviceId string) (
	service postgress.PickDropServiceDetails,
	counts pickDropServiceCountsRow,
	shifts []postgress.ShiftDetails,
	err error,
) {
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	if err = db.Table("pick_drop_services").
		Select(`
			pick_drop_services.id,
			pick_drop_services.name,
			pick_drop_services.description,
			pick_drop_services.owner_driver_id,
			COALESCE(drivers.driver_name, '') AS owner_name,
			COALESCE(drivers.driver_mobile, '') AS owner_mobile,
			pick_drop_services.status,
			pick_drop_services.created_at
		`).
		Joins("LEFT JOIN drivers ON drivers.id = pick_drop_services.owner_driver_id").
		Where("pick_drop_services.id = ?", serviceId).
		Take(&service).Error; err != nil {
		return
	}

	if err = db.Raw(`
		SELECT
			(SELECT COUNT(*) FROM pick_drop_drivers
				WHERE service_id = @service AND status = @approved AND join_type = @driverJoin) AS approved_drivers,
			(SELECT COUNT(*) FROM pick_drop_drivers
				WHERE service_id = @service AND status = @approved AND join_type = @vehiclesJoin) AS approved_vehicle_owners,
			(SELECT COUNT(*) FROM pick_drop_vehicles
				JOIN vehicles ON vehicles.id = pick_drop_vehicles.vehicle_id AND vehicles.status = @active
				WHERE pick_drop_vehicles.service_id = @service AND pick_drop_vehicles.status = @approved) AS approved_vehicles,
			(SELECT COUNT(*) FROM pick_drop_passengers WHERE service_id = @service AND status = @approved) AS approved_passengers,
			(SELECT COUNT(*) FROM pick_drop_drivers WHERE service_id = @service AND status = @pending)
				+ (SELECT COUNT(*) FROM pick_drop_vehicles WHERE service_id = @service AND status = @pending)
				+ (SELECT COUNT(*) FROM pick_drop_passengers WHERE service_id = @service AND status = @pending) AS pending_requests,
			(SELECT COUNT(*) FROM shifts WHERE service_id = @service AND status = @activeShift) AS active_shifts
	`,
		map[string]interface{}{
			"service":      serviceId,
			"approved":     constants.Membership_Status_Approved,
			"pending":      constants.Membership_Status_Pending,
			"active":       constants.Status_Active,
			"activeShift":  constants.Shift_Status_Active,
			"driverJoin":   constants.Join_Type_Driver,
			"vehiclesJoin": constants.Join_Type_Vehicles,
		},
	).Scan(&counts).Error; err != nil {
		return
	}

	err = shiftAdminJoins(db).
		Where("shifts.service_id = ?", serviceId).
		Select(shiftAdminColumns).
		Order("shifts.status ASC, shifts.start_time ASC").
		Find(&shifts).Error

	return
}

// shiftAdminJoins is the platform-wide equivalent of the shift join services/shift
// keeps to itself: every shift, whichever service or driver, with the service, the
// driver, the vehicle and the vehicle's own owner (who is not always who drives it).
func shiftAdminJoins(db *gorm.DB) *gorm.DB {
	return db.Table("shifts").
		Joins("LEFT JOIN pick_drop_services ON pick_drop_services.id = shifts.service_id").
		Joins("LEFT JOIN drivers ON drivers.id = shifts.driver_id").
		Joins("LEFT JOIN vehicles ON vehicles.id = shifts.vehicle_id").
		Joins("LEFT JOIN drivers AS vehicle_owners ON vehicle_owners.id = vehicles.driver_id")
}

const shiftAdminColumns = `
	shifts.id,
	shifts.name,
	shifts.service_id,
	COALESCE(pick_drop_services.name, '') AS service_name,
	COALESCE(pick_drop_services.owner_driver_id, '') AS owner_driver_id,
	shifts.driver_id,
	COALESCE(drivers.driver_name, '') AS driver_name,
	COALESCE(drivers.driver_mobile, '') AS driver_mobile,
	shifts.vehicle_id,
	COALESCE(vehicles.vehicle_number, '') AS vehicle_number,
	COALESCE(vehicles.vehicle_info, '') AS vehicle_info,
	COALESCE(vehicles.driver_id, '') AS vehicle_owner_id,
	COALESCE(vehicle_owners.driver_name, '') AS vehicle_owner_name,
	shifts.days_of_week,
	shifts.start_date,
	shifts.end_date,
	shifts.start_time,
	shifts.end_time,
	shifts.seat_capacity,
	shifts.occupied_seats,
	shifts.status,
	shifts.created_at,
	shifts.updated_at`

func applyAdminShiftFilter(query *gorm.DB, filter adminShiftFilter) *gorm.DB {
	if filter.Status != constants.Shift_Status_All {
		query = query.Where("shifts.status = ?", filter.Status)
	}

	if filter.DayOfWeek > 0 {
		query = query.Where("? = ANY(shifts.days_of_week)", filter.DayOfWeek)
	}

	if !utils.IsStringEmpty(filter.Search) {
		term := "%" + filter.Search + "%"
		query = query.Where(`(
			shifts.name ILIKE ?
			OR pick_drop_services.name ILIKE ?
			OR drivers.driver_name ILIKE ?
			OR EXISTS (SELECT 1 FROM shift_locations WHERE shift_locations.shift_id = shifts.id AND shift_locations.location ILIKE ?)
		)`, term, term, term, term)
	}

	if !utils.IsStringEmpty(filter.StartTime) {
		query = query.Where("shifts.start_time >= ?", filter.StartTime)
	}

	if !utils.IsStringEmpty(filter.EndTime) {
		query = query.Where("shifts.end_time <= ?", filter.EndTime)
	}

	return query
}

// listShiftsAdmin is every shift on the platform, in any service, searchable the same
// way an owner searches their own.
func listShiftsAdmin(orgCtx *gin.Context, filter adminShiftFilter, page int) (shifts []postgress.ShiftDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := applyAdminShiftFilter(shiftAdminJoins(database.DatabaseConn.Postgres.WithContext(ctx)), filter)

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Select(shiftAdminColumns).
		Order("shifts.start_time ASC, shifts.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&shifts).Error

	return
}

// listShiftRequestsAdmin is every open shift request on the platform, addressed to a
// service or still open to any of them, searchable the same way an owner searches.
func listShiftRequestsAdmin(orgCtx *gin.Context, filter adminShiftFilter, page int) (requests []postgress.ShiftRequestDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Table("shift_requests").
		Joins("LEFT JOIN passengers ON passengers.id = shift_requests.passenger_id").
		Joins("LEFT JOIN pick_drop_services ON pick_drop_services.id = shift_requests.service_id")

	if !utils.IsStringEmpty(filter.Search) {
		term := "%" + filter.Search + "%"
		query = query.Where(`(
			shift_requests.start_location ILIKE ?
			OR shift_requests.end_location ILIKE ?
			OR passengers.passenger_name ILIKE ?
			OR EXISTS (SELECT 1 FROM unnest(shift_requests.route_points) AS point WHERE point ILIKE ?)
		)`, term, term, term, term)
	}

	if filter.DayOfWeek > 0 {
		query = query.Where("? = ANY(shift_requests.days_of_week)", filter.DayOfWeek)
	}

	if !utils.IsStringEmpty(filter.StartTime) {
		query = query.Where("shift_requests.start_time >= ?", filter.StartTime)
	}

	if !utils.IsStringEmpty(filter.EndTime) {
		query = query.Where("shift_requests.end_time <= ?", filter.EndTime)
	}

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Select(`
			shift_requests.id,
			shift_requests.passenger_id,
			COALESCE(passengers.passenger_name, '') AS passenger_name,
			COALESCE(shift_requests.service_id, '') AS service_id,
			COALESCE(pick_drop_services.name, '') AS service_name,
			shift_requests.contact_number,
			shift_requests.note,
			shift_requests.days_of_week,
			shift_requests.start_time,
			shift_requests.end_time,
			shift_requests.created_at
		`).
		Order("shift_requests.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&requests).Error

	return
}

// listAdvertisementsAdmin is every advertisement on the platform, whatever the
// status of the service that posted it.
func listAdvertisementsAdmin(orgCtx *gin.Context, search string, page int) (ads []postgress.AdvertisementDetails, totalRows int64, err error) {
	pageSize := configuration.ConfigurationData.PageSize
	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	query := database.DatabaseConn.Postgres.WithContext(ctx).Table("pick_drop_advertisements").
		Joins("LEFT JOIN pick_drop_services ON pick_drop_services.id = pick_drop_advertisements.service_id").
		Joins("LEFT JOIN drivers ON drivers.id = pick_drop_services.owner_driver_id")

	if !utils.IsStringEmpty(search) {
		term := "%" + strings.TrimSpace(search) + "%"
		query = query.Where(`(
			pick_drop_advertisements.title ILIKE ?
			OR pick_drop_advertisements.start_location ILIKE ?
			OR pick_drop_advertisements.end_location ILIKE ?
			OR pick_drop_services.name ILIKE ?
		)`, term, term, term, term)
	}

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Select(`
			pick_drop_advertisements.id,
			pick_drop_advertisements.service_id,
			COALESCE(pick_drop_services.name, '') AS service_name,
			COALESCE(pick_drop_services.owner_driver_id, '') AS owner_driver_id,
			COALESCE(drivers.driver_name, '') AS owner_name,
			COALESCE(drivers.driver_mobile, '') AS owner_mobile,
			pick_drop_advertisements.title,
			pick_drop_advertisements.description,
			pick_drop_advertisements.fare,
			pick_drop_advertisements.days_of_week,
			pick_drop_advertisements.start_time,
			pick_drop_advertisements.end_time,
			(SELECT COUNT(*) FROM pick_drop_vehicles
				JOIN vehicles ON vehicles.id = pick_drop_vehicles.vehicle_id AND vehicles.status = ?
				WHERE pick_drop_vehicles.service_id = pick_drop_advertisements.service_id
				  AND pick_drop_vehicles.status = ?) AS vehicle_count,
			pick_drop_advertisements.created_at
		`, constants.Status_Active, constants.Membership_Status_Approved).
		Order("pick_drop_advertisements.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&ads).Error

	return
}
