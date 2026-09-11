package pickdrop

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
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// a request still waiting and a membership already granted are both open, a driver,
// a passenger or a vehicle may hold at most one of either at any time
var openStatuses = []string{constants.Membership_Status_Pending, constants.Membership_Status_Approved}

func withTimeout(orgCtx *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
}

func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func pageBounds(page int) (limit, offset int) {
	limit = configuration.ConfigurationData.PageSize
	offset = (page - 1) * limit
	return
}

func forUpdate(tx *gorm.DB) *gorm.DB {
	return tx.Clauses(clause.Locking{Strength: "UPDATE"})
}

func driverNotice(driverId, notificationType, title, message string) notice {
	return notice{UserId: driverId, UserType: constants.User_Driver, Type: notificationType, Title: title, Message: message}
}

func passengerNotice(passengerId, notificationType, title, message string) notice {
	return notice{UserId: passengerId, UserType: constants.User_Passenger, Type: notificationType, Title: title, Message: message}
}

func sendNotices(ctx *gin.Context, sessionId, serviceId string, notices []notice) {
	data := map[string]string{constants.NOTIFICATION_KEY_SERVICE_ID: serviceId}

	for _, item := range notices {
		utils.SendUserNotification(ctx, sessionId, item.Type, item.UserId, item.UserType, item.Title, item.Message, data)
	}
}

// requireOwnedService is the gate of every owner api: the caller has to run an active
// service, and only the one they run is ever touched.
func requireOwnedService(ctx *gin.Context, sessionId, driverId string) (service postgress.PickDropService, err error) {
	service, err = database.GetActiveServiceByOwner(ctx, driverId)
	if err == nil {
		return
	}

	if isNotFound(err) {
		err = errors.New(constants.Not_Service_Owner)
	} else {
		logger.LogError(sessionId, "failed to read the service error: "+err.Error())
		err = errors.New(constants.Unknown_Error)
	}
	logger.LogError(sessionId, err)

	return
}

// lockActiveDriver takes the driver's row for the rest of the transaction. Every
// change to who a driver belongs to goes through this, so two of them for the same
// driver run one after the other and the one membership rule cannot be raced.
func lockActiveDriver(tx *gorm.DB, driverId string) (driver postgress.Driver, err error) {
	err = forUpdate(tx).
		Where("id = ?", driverId).
		Where("status = ?", constants.Status_Active).
		First(&driver).Error
	if isNotFound(err) {
		err = utils.Refuse(constants.Driver_Not_Found)
	}

	return
}

func lockActivePassenger(tx *gorm.DB, passengerId string) (passenger postgress.Passenger, err error) {
	err = forUpdate(tx).
		Where("id = ?", passengerId).
		Where("status = ?", constants.Status_Active).
		First(&passenger).Error
	if isNotFound(err) {
		err = utils.Refuse(constants.Passenger_Not_Found)
	}

	return
}

func lockOwnedService(tx *gorm.DB, ownerId string) (service postgress.PickDropService, err error) {
	err = forUpdate(tx).
		Where("owner_driver_id = ?", ownerId).
		Where("status = ?", constants.Status_Active).
		First(&service).Error
	if isNotFound(err) {
		err = utils.Refuse(constants.Not_Service_Owner)
	}

	return
}

func getActiveService(tx *gorm.DB, serviceId string) (service postgress.PickDropService, err error) {
	err = tx.Where("id = ?", serviceId).
		Where("status = ?", constants.Status_Active).
		First(&service).Error
	if isNotFound(err) {
		err = utils.Refuse(constants.Service_Not_Found)
	}

	return
}

// checkDriverIsFree enforces that a driver belongs to one service at most, as its
// owner or as a member, and that one open request is the most they can hold.
func checkDriverIsFree(tx *gorm.DB, driverId string) error {
	var owned int64
	if err := tx.Model(&postgress.PickDropService{}).
		Where("owner_driver_id = ?", driverId).
		Where("status = ?", constants.Status_Active).
		Count(&owned).Error; err != nil {
		return err
	}

	if owned > 0 {
		return utils.Refuse(constants.Already_Owns_Service)
	}

	var open postgress.PickDropDriver
	err := tx.Where("driver_id = ?", driverId).Where("status IN ?", openStatuses).First(&open).Error
	if isNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if open.Status == constants.Membership_Status_Pending {
		return utils.Refuse(constants.Pending_Service_Request)
	}

	return utils.Refuse(constants.Already_In_Service)
}

func checkPassengerIsFree(tx *gorm.DB, passengerId string) error {
	var open postgress.PickDropPassenger
	err := tx.Where("passenger_id = ?", passengerId).Where("status IN ?", openStatuses).First(&open).Error
	if isNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if open.Status == constants.Membership_Status_Pending {
		return utils.Refuse(constants.Pending_Service_Request)
	}

	return utils.Refuse(constants.Already_In_Service)
}

// loadOfferableVehicles loads, in one query, the vehicles a driver really owns and
// has not already offered to a service. Seats are not required to offer a vehicle,
// only to put it on a shift.
func loadOfferableVehicles(tx *gorm.DB, driverId string, vehicleIds []string) (vehicles []postgress.Vehicle, err error) {
	if len(vehicleIds) == 0 {
		return
	}

	if err = tx.Where("id IN ?", vehicleIds).
		Where("driver_id = ?", driverId).
		Where("status = ?", constants.Status_Active).
		Find(&vehicles).Error; err != nil {
		return
	}

	if len(vehicles) != len(vehicleIds) {
		err = utils.Refuse(constants.Vehicle_Not_Found)
		return
	}

	var offered []postgress.PickDropVehicle
	if err = tx.Where("vehicle_id IN ?", vehicleIds).Where("status IN ?", openStatuses).Find(&offered).Error; err != nil {
		return
	}

	for _, offer := range offered {
		for _, vehicle := range vehicles {
			if vehicle.ID == offer.VehicleID {
				err = utils.Refuse(fmt.Sprintf(constants.Vehicle_Already_In_Service, vehicle.VehicleNumber))
				return
			}
		}
	}

	return
}

func insertVehicleOffers(tx *gorm.DB, serviceId, driverId string, vehicles []postgress.Vehicle, status, decidedBy string) error {
	if len(vehicles) == 0 {
		return nil
	}

	now := time.Now()
	rows := make([]postgress.PickDropVehicle, 0, len(vehicles))

	for _, vehicle := range vehicles {
		row := postgress.PickDropVehicle{
			ID:          database.GenerateUUID(),
			ServiceID:   serviceId,
			VehicleID:   vehicle.ID,
			DriverID:    driverId,
			Status:      status,
			RequestedAt: now,
		}

		if status == constants.Membership_Status_Approved {
			row.DecidedBy = decidedBy
			row.DecidedAt = &now
		}

		rows = append(rows, row)
	}

	return tx.Create(&rows).Error
}

// joinLabel is how a driver's way of joining reads in a notification.
func joinLabel(joinType string) string {
	if joinType == constants.Join_Type_Vehicles {
		return "vehicle provider, with vehicles only"
	}

	return constants.Request_Type_Driver
}

func vehicleNumber(tx *gorm.DB, vehicleId string) string {
	var vehicle postgress.Vehicle
	if err := tx.Select("vehicle_number").Where("id = ?", vehicleId).Take(&vehicle).Error; err != nil {
		return ""
	}

	return vehicle.VehicleNumber
}

// createService enables Pick & Drop for a driver: the service, with the driver as its
// one owner, and the vehicles they chose, already approved since they are the owner's.
func createService(orgCtx *gin.Context, sessionId, driverId string, request EnableServiceRequest) (serviceId string, err error) {
	logger.LogInfo("Request received in createService", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, e := lockActiveDriver(tx, driverId); e != nil {
			return e
		}

		if e := checkDriverIsFree(tx, driverId); e != nil {
			return e
		}

		vehicles, e := loadOfferableVehicles(tx, driverId, request.VehicleIds)
		if e != nil {
			return e
		}

		service := postgress.PickDropService{
			ID:            database.GenerateUUID(),
			Name:          request.Name,
			Description:   request.Description,
			OwnerDriverID: driverId,
			Status:        constants.Status_Active,
		}

		if e = tx.Create(&service).Error; e != nil {
			return e
		}
		serviceId = service.ID

		return insertVehicleOffers(tx, service.ID, driverId, vehicles, constants.Membership_Status_Approved, driverId)
	})

	logger.LogInfo("Response returned from createService", sessionId)

	return
}

func getServiceDetails(db *gorm.DB, serviceId string) (details postgress.PickDropServiceDetails, err error) {
	err = db.Table("pick_drop_services").
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
		Take(&details).Error

	return
}

// getServiceCounts counts everything an owner's dashboard shows in one round trip.
func getServiceCounts(db *gorm.DB, serviceId string) (counts serviceCountsRow, err error) {
	err = db.Raw(`
		SELECT
			(SELECT COUNT(*) FROM pick_drop_drivers
				WHERE service_id = @service AND status = @approved AND join_type = @driverJoin) AS approved_drivers,
			(SELECT COUNT(*) FROM pick_drop_drivers
				WHERE service_id = @service AND status = @approved AND join_type = @vehiclesJoin) AS approved_vehicle_owners,
			(SELECT COUNT(*) FROM pick_drop_vehicles
				JOIN vehicles ON vehicles.id = pick_drop_vehicles.vehicle_id AND vehicles.status = @active
				WHERE pick_drop_vehicles.service_id = @service AND pick_drop_vehicles.status = @approved) AS approved_vehicles,
			(SELECT COUNT(*) FROM pick_drop_passengers
				WHERE service_id = @service AND status = @approved) AS approved_passengers,
			(SELECT COUNT(*) FROM pick_drop_drivers WHERE service_id = @service AND status = @pending)
				+ (SELECT COUNT(*) FROM pick_drop_vehicles WHERE service_id = @service AND status = @pending)
				+ (SELECT COUNT(*) FROM pick_drop_passengers WHERE service_id = @service AND status = @pending) AS pending_requests,
			(SELECT COUNT(*) FROM shifts
				WHERE service_id = @service AND status = @active) AS active_shifts,
			(SELECT COUNT(*) FROM pick_drop_advertisements
				WHERE service_id = @service) AS advertisements
	`, map[string]interface{}{
		"service":      serviceId,
		"approved":     constants.Membership_Status_Approved,
		"pending":      constants.Membership_Status_Pending,
		"active":       constants.Status_Active,
		"driverJoin":   constants.Join_Type_Driver,
		"vehiclesJoin": constants.Join_Type_Vehicles,
	}).Scan(&counts).Error

	return
}

// countApprovedVehicles is the number the advertisement limit is worked out from:
// vehicles approved for the service that still exist.
func countApprovedVehicles(tx *gorm.DB, serviceId string) (count int64, err error) {
	err = tx.Table("pick_drop_vehicles").
		Joins("JOIN vehicles ON vehicles.id = pick_drop_vehicles.vehicle_id AND vehicles.status = ?", constants.Status_Active).
		Where("pick_drop_vehicles.service_id = ?", serviceId).
		Where("pick_drop_vehicles.status = ?", constants.Membership_Status_Approved).
		Count(&count).Error

	return
}

func vehicleJoins(db *gorm.DB) *gorm.DB {
	return db.Table("pick_drop_vehicles").
		Joins("LEFT JOIN vehicles ON vehicles.id = pick_drop_vehicles.vehicle_id").
		Joins("LEFT JOIN drivers ON drivers.id = pick_drop_vehicles.driver_id")
}

const vehicleColumns = `
	pick_drop_vehicles.id,
	pick_drop_vehicles.service_id,
	pick_drop_vehicles.vehicle_id,
	COALESCE(vehicles.vehicle_number, '') AS vehicle_number,
	COALESCE(vehicles.vehicle_info, '') AS vehicle_info,
	COALESCE(vehicles.number_of_seats, 0) AS number_of_seats,
	COALESCE(vehicles.has_ac, false) AS has_ac,
	COALESCE(vehicles.has_heating, false) AS has_heating,
	pick_drop_vehicles.driver_id,
	COALESCE(drivers.driver_name, '') AS driver_name,
	COALESCE(drivers.driver_mobile, '') AS driver_mobile,
	CASE
		WHEN pick_drop_vehicles.driver_id = (
			SELECT pick_drop_services.owner_driver_id FROM pick_drop_services
			WHERE pick_drop_services.id = pick_drop_vehicles.service_id
		) THEN '` + constants.Service_Role_Owner + `'
		ELSE COALESCE((
			SELECT pick_drop_drivers.join_type FROM pick_drop_drivers
			WHERE pick_drop_drivers.service_id = pick_drop_vehicles.service_id
			  AND pick_drop_drivers.driver_id = pick_drop_vehicles.driver_id
			ORDER BY pick_drop_drivers.requested_at DESC
			LIMIT 1
		), '` + constants.Join_Type_Driver + `')
	END AS owner_join_type,
	pick_drop_vehicles.status,
	pick_drop_vehicles.requested_at,
	pick_drop_vehicles.decided_at,
	pick_drop_vehicles.ended_at,
	(SELECT COUNT(*) FROM shifts WHERE shifts.vehicle_id = pick_drop_vehicles.vehicle_id AND shifts.status = ?) AS active_shifts`

func getDriverVehicleOffers(db *gorm.DB, serviceId string, driverIds []string) (vehicles []postgress.PickDropVehicleDetails, err error) {
	if len(driverIds) == 0 {
		return
	}

	err = vehicleJoins(db).
		Select(vehicleColumns, constants.Shift_Status_Active).
		Where("pick_drop_vehicles.service_id = ?", serviceId).
		Where("pick_drop_vehicles.driver_id IN ?", driverIds).
		Order("pick_drop_vehicles.requested_at DESC").
		Find(&vehicles).Error

	return
}

// getMyService works out what Pick & Drop is to a driver: the service they own, the
// one they belong to or are waiting on, or nothing.
func getMyService(orgCtx *gin.Context, driverId string) (data myServiceData, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)
	data.Role = constants.Service_Role_None

	var owned postgress.PickDropService
	err = db.Where("owner_driver_id = ?", driverId).Where("status = ?", constants.Status_Active).First(&owned).Error
	if err == nil {
		var details postgress.PickDropServiceDetails
		if details, err = getServiceDetails(db, owned.ID); err != nil {
			return
		}

		var counts serviceCountsRow
		if counts, err = getServiceCounts(db, owned.ID); err != nil {
			return
		}

		data.Role = constants.Service_Role_Owner
		data.Service = &details
		data.Counts = &counts
		return
	}

	if !isNotFound(err) {
		return
	}

	var membership postgress.PickDropDriver
	err = db.Where("driver_id = ?", driverId).Where("status IN ?", openStatuses).First(&membership).Error
	if isNotFound(err) {
		err = nil
		return
	}
	if err != nil {
		return
	}

	var details postgress.PickDropServiceDetails
	if details, err = getServiceDetails(db, membership.ServiceID); err != nil {
		return
	}

	data.Role = constants.Service_Role_Driver
	data.Service = &details
	data.Membership = &membership
	data.Vehicles, err = getDriverVehicleOffers(db, membership.ServiceID, []string{driverId})

	return
}

// disableService switches a service off. It refuses while shifts are still active,
// people are counting on those. Everybody still in or waiting at the door is let go,
// the advertisements are archived, and nothing else is deleted.
func disableService(orgCtx *gin.Context, sessionId, ownerId string) (service postgress.PickDropService, notices []notice, err error) {
	logger.LogInfo("Request received in disableService", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e error
		if service, e = lockOwnedService(tx, ownerId); e != nil {
			return e
		}

		var activeShifts int64
		if e = tx.Model(&postgress.Shift{}).
			Where("service_id = ?", service.ID).
			Where("status = ?", constants.Shift_Status_Active).
			Count(&activeShifts).Error; e != nil {
			return e
		}

		if activeShifts > 0 {
			return utils.Refuse(fmt.Sprintf(constants.Service_Has_Active_Shifts, activeShifts))
		}

		var drivers []postgress.PickDropDriver
		if e = tx.Where("service_id = ?", service.ID).Where("status IN ?", openStatuses).Find(&drivers).Error; e != nil {
			return e
		}

		var passengers []postgress.PickDropPassenger
		if e = tx.Where("service_id = ?", service.ID).Where("status IN ?", openStatuses).Find(&passengers).Error; e != nil {
			return e
		}

		now := time.Now()

		for _, model := range []interface{}{&postgress.PickDropDriver{}, &postgress.PickDropVehicle{}, &postgress.PickDropPassenger{}} {
			if e = tx.Model(model).
				Where("service_id = ?", service.ID).
				Where("status IN ?", openStatuses).
				Updates(map[string]interface{}{
					"status":   constants.Membership_Status_Closed,
					"ended_by": ownerId,
					"ended_at": &now,
				}).Error; e != nil {
				return e
			}
		}

		var ads []postgress.PickDropAdvertisement
		if e = tx.Where("service_id = ?", service.ID).Find(&ads).Error; e != nil {
			return e
		}

		if e = archiveAdvertisements(tx, ads, ownerId); e != nil {
			return e
		}

		if e = tx.Model(&postgress.PickDropService{}).
			Where("id = ?", service.ID).
			Updates(map[string]interface{}{
				"status":      constants.Status_InActive,
				"disabled_at": &now,
			}).Error; e != nil {
			return e
		}

		message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_CLOSED, service.Name)
		for _, driver := range drivers {
			notices = append(notices, driverNotice(driver.DriverID, constants.NOTIFICATION_TYPE_PICKDROP_UPDATE, constants.NOTIFICATION_TITLE_PICKDROP_UPDATE, message))
		}
		for _, passenger := range passengers {
			notices = append(notices, passengerNotice(passenger.PassengerID, constants.NOTIFICATION_TYPE_PICKDROP_UPDATE, constants.NOTIFICATION_TITLE_PICKDROP_UPDATE, message))
		}

		return nil
	})

	logger.LogInfo("Response returned from disableService", sessionId)

	return
}

// archiveAdvertisements copies advertisements and their locations into the archive
// and hard deletes the originals, the pattern rides already follow.
func archiveAdvertisements(tx *gorm.DB, ads []postgress.PickDropAdvertisement, deletedBy string) error {
	if len(ads) == 0 {
		return nil
	}

	adIds := make([]string, 0, len(ads))
	archived := make([]postgress.DELPickDropAdvertisement, 0, len(ads))

	for _, ad := range ads {
		adIds = append(adIds, ad.ID)
		archived = append(archived, postgress.DELPickDropAdvertisement{
			ID:            ad.ID,
			ServiceID:     ad.ServiceID,
			Title:         ad.Title,
			Description:   ad.Description,
			Fare:          ad.Fare,
			DaysOfWeek:    ad.DaysOfWeek,
			StartTime:     ad.StartTime,
			EndTime:       ad.EndTime,
			StartLocation: ad.StartLocation,
			EndLocation:   ad.EndLocation,
			RoutePoints:   ad.RoutePoints,
			CreatedBy:     ad.CreatedBy,
			DeletedBy:     deletedBy,
			CreatedAt:     ad.CreatedAt,
			UpdatedAt:     time.Now(),
		})
	}

	var locations []postgress.PickDropAdvertisementLocation
	if err := tx.Where("advertisement_id IN ?", adIds).Find(&locations).Error; err != nil {
		return err
	}

	if err := tx.Create(&archived).Error; err != nil {
		return err
	}

	if len(locations) > 0 {
		archivedLocations := make([]postgress.DELPickDropAdvertisementLocation, 0, len(locations))
		for _, location := range locations {
			archivedLocations = append(archivedLocations, postgress.DELPickDropAdvertisementLocation{
				ID:              location.ID,
				AdvertisementID: location.AdvertisementID,
				Sequence:        location.Sequence,
				Location:        location.Location,
				Lat:             location.Lat,
				Lng:             location.Lng,
				Time:            location.Time,
				CreatedAt:       location.CreatedAt,
			})
		}

		if err := tx.Create(&archivedLocations).Error; err != nil {
			return err
		}
	}

	if err := tx.Where("advertisement_id IN ?", adIds).Delete(&postgress.PickDropAdvertisementLocation{}).Error; err != nil {
		return err
	}

	return tx.Where("id IN ?", adIds).Delete(&postgress.PickDropAdvertisement{}).Error
}

// searchServices shows the services somebody could ask to join, with their size and
// the caller's own standing in each, counted as subselects rather than a query per row.
func searchServices(orgCtx *gin.Context, userId, search string, page int) (services []postgress.PickDropServiceDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	base := database.DatabaseConn.Postgres.WithContext(ctx).
		Table("pick_drop_services").
		Joins("LEFT JOIN drivers ON drivers.id = pick_drop_services.owner_driver_id").
		Where("pick_drop_services.status = ?", constants.Status_Active)

	if !utils.IsStringEmpty(search) {
		base = base.Where("pick_drop_services.name ILIKE ?", "%"+strings.TrimSpace(search)+"%")
	}

	if err = base.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	// one endpoint serves both kinds of token, so the caller's id is looked for as the
	// owner, as a driver member and as a passenger member
	err = base.
		Select(`
			pick_drop_services.id,
			pick_drop_services.name,
			pick_drop_services.description,
			pick_drop_services.owner_driver_id,
			COALESCE(drivers.driver_name, '') AS owner_name,
			pick_drop_services.status,
			(SELECT COUNT(*) FROM pick_drop_drivers d WHERE d.service_id = pick_drop_services.id AND d.status = ?) AS driver_count,
			(SELECT COUNT(*) FROM pick_drop_vehicles v WHERE v.service_id = pick_drop_services.id AND v.status = ?) AS vehicle_count,
			(SELECT COUNT(*) FROM pick_drop_passengers p WHERE p.service_id = pick_drop_services.id AND p.status = ?) AS passenger_count,
			COALESCE(
				CASE WHEN pick_drop_services.owner_driver_id = ? THEN 'owner' END,
				(SELECT d.status FROM pick_drop_drivers d WHERE d.service_id = pick_drop_services.id AND d.driver_id = ? AND d.status IN ? LIMIT 1),
				(SELECT p.status FROM pick_drop_passengers p WHERE p.service_id = pick_drop_services.id AND p.passenger_id = ? AND p.status IN ? LIMIT 1),
				''
			) AS my_status,
			pick_drop_services.created_at
		`,
			constants.Membership_Status_Approved,
			constants.Membership_Status_Approved,
			constants.Membership_Status_Approved,
			userId,
			userId, openStatuses,
			userId, openStatuses,
		).
		Order("pick_drop_services.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&services).Error

	return
}

func addOwnerVehicles(orgCtx *gin.Context, sessionId, ownerId string, vehicleIds []string) (err error) {
	logger.LogInfo("Request received in addOwnerVehicles", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		service, e := lockOwnedService(tx, ownerId)
		if e != nil {
			return e
		}

		vehicles, e := loadOfferableVehicles(tx, ownerId, vehicleIds)
		if e != nil {
			return e
		}

		return insertVehicleOffers(tx, service.ID, ownerId, vehicles, constants.Membership_Status_Approved, ownerId)
	})

	logger.LogInfo("Response returned from addOwnerVehicles", sessionId)

	return
}

// removeServiceVehicle takes a vehicle out of the service. It is refused while the
// vehicle is still on an active shift, a shift without its vehicle strands people.
func removeServiceVehicle(orgCtx *gin.Context, sessionId, ownerId, vehicleId string) (service postgress.PickDropService, notices []notice, err error) {
	logger.LogInfo("Request received in removeServiceVehicle", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e error
		if service, e = lockOwnedService(tx, ownerId); e != nil {
			return e
		}

		var row postgress.PickDropVehicle
		if e = forUpdate(tx).
			Where("service_id = ?", service.ID).
			Where("vehicle_id = ?", vehicleId).
			Where("status = ?", constants.Membership_Status_Approved).
			First(&row).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Vehicle_Not_In_Service)
			}
			return e
		}

		_, vehicleShifts, e := database.CountActiveShiftAssignments(tx, "", []string{vehicleId})
		if e != nil {
			return e
		}

		if vehicleShifts > 0 {
			return utils.Refuse(fmt.Sprintf(constants.On_Active_Shifts, "This vehicle is", vehicleShifts))
		}

		now := time.Now()
		if e = tx.Model(&postgress.PickDropVehicle{}).
			Where("id = ?", row.ID).
			Updates(map[string]interface{}{
				"status":   constants.Membership_Status_Removed,
				"ended_by": ownerId,
				"ended_at": &now,
			}).Error; e != nil {
			return e
		}

		if row.DriverID != ownerId {
			notices = append(notices, driverNotice(row.DriverID, constants.NOTIFICATION_TYPE_PICKDROP_DECISION, constants.NOTIFICATION_TITLE_PICKDROP_DECISION,
				fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_REMOVED, vehicleNumber(tx, vehicleId), service.Name)))
		}

		return nil
	})

	logger.LogInfo("Response returned from removeServiceVehicle", sessionId)

	return
}

func requestJoinAsDriver(orgCtx *gin.Context, sessionId, driverId string, request DriverJoinRequest) (service postgress.PickDropService, requestId string, notices []notice, err error) {
	logger.LogInfo("Request received in requestJoinAsDriver", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		driver, e := lockActiveDriver(tx, driverId)
		if e != nil {
			return e
		}

		if service, e = getActiveService(tx, request.ServiceId); e != nil {
			return e
		}

		if e = checkDriverIsFree(tx, driverId); e != nil {
			return e
		}

		vehicles, e := loadOfferableVehicles(tx, driverId, request.VehicleIds)
		if e != nil {
			return e
		}

		membership := postgress.PickDropDriver{
			ID:          database.GenerateUUID(),
			ServiceID:   service.ID,
			DriverID:    driverId,
			JoinType:    request.JoinType,
			Status:      constants.Membership_Status_Pending,
			RequestedAt: time.Now(),
		}

		if e = tx.Create(&membership).Error; e != nil {
			return e
		}
		requestId = membership.ID

		if e = insertVehicleOffers(tx, service.ID, driverId, vehicles, constants.Membership_Status_Pending, ""); e != nil {
			return e
		}

		notices = append(notices, driverNotice(service.OwnerDriverID, constants.NOTIFICATION_TYPE_PICKDROP_REQUEST, constants.NOTIFICATION_TITLE_PICKDROP_REQUEST,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_JOIN_REQUEST, driver.DriverName, service.Name, joinLabel(request.JoinType))))

		return nil
	})

	logger.LogInfo("Response returned from requestJoinAsDriver", sessionId)

	return
}

// offerVehicles lets a driver who belongs to a service, or is waiting to, offer more
// vehicles later. Each still needs the owner's approval.
func offerVehicles(orgCtx *gin.Context, sessionId, driverId string, vehicleIds []string) (service postgress.PickDropService, notices []notice, err error) {
	logger.LogInfo("Request received in offerVehicles", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		driver, e := lockActiveDriver(tx, driverId)
		if e != nil {
			return e
		}

		var membership postgress.PickDropDriver
		if e = tx.Where("driver_id = ?", driverId).Where("status IN ?", openStatuses).First(&membership).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Not_Service_Member)
			}
			return e
		}

		if service, e = getActiveService(tx, membership.ServiceID); e != nil {
			return e
		}

		vehicles, e := loadOfferableVehicles(tx, driverId, vehicleIds)
		if e != nil {
			return e
		}

		if e = insertVehicleOffers(tx, service.ID, driverId, vehicles, constants.Membership_Status_Pending, ""); e != nil {
			return e
		}

		notices = append(notices, driverNotice(service.OwnerDriverID, constants.NOTIFICATION_TYPE_PICKDROP_REQUEST, constants.NOTIFICATION_TITLE_PICKDROP_REQUEST,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_OFFER, driver.DriverName, len(vehicles), service.Name)))

		return nil
	})

	logger.LogInfo("Response returned from offerVehicles", sessionId)

	return
}

// withdrawVehicle takes back a vehicle offer, or an approved vehicle that no active
// shift depends on.
func withdrawVehicle(orgCtx *gin.Context, sessionId, driverId, vehicleId string) (service postgress.PickDropService, notices []notice, err error) {
	logger.LogInfo("Request received in withdrawVehicle", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		driver, e := lockActiveDriver(tx, driverId)
		if e != nil {
			return e
		}

		var row postgress.PickDropVehicle
		if e = forUpdate(tx).
			Where("vehicle_id = ?", vehicleId).
			Where("driver_id = ?", driverId).
			Where("status IN ?", openStatuses).
			First(&row).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Vehicle_Not_In_Service)
			}
			return e
		}

		// a driver who joined with vehicles only belongs to the service through them, so
		// they keep at least one and leave the service to take the last one out
		var membership postgress.PickDropDriver
		if e = tx.Where("service_id = ?", row.ServiceID).
			Where("driver_id = ?", driverId).
			Where("status IN ?", openStatuses).
			First(&membership).Error; e != nil && !isNotFound(e) {
			return e
		}

		if membership.JoinType == constants.Join_Type_Vehicles {
			var others int64
			if e = tx.Model(&postgress.PickDropVehicle{}).
				Where("service_id = ?", row.ServiceID).
				Where("driver_id = ?", driverId).
				Where("status IN ?", openStatuses).
				Where("id <> ?", row.ID).
				Count(&others).Error; e != nil {
				return e
			}

			if others == 0 {
				return utils.Refuse(constants.Last_Vehicle_Of_Vehicles_Member)
			}
		}

		status := constants.Membership_Status_Withdrawn
		if row.Status == constants.Membership_Status_Approved {
			_, vehicleShifts, e := database.CountActiveShiftAssignments(tx, "", []string{vehicleId})
			if e != nil {
				return e
			}

			if vehicleShifts > 0 {
				return utils.Refuse(fmt.Sprintf(constants.On_Active_Shifts, "This vehicle is", vehicleShifts))
			}

			status = constants.Membership_Status_Left
		}

		now := time.Now()
		if e = tx.Model(&postgress.PickDropVehicle{}).
			Where("id = ?", row.ID).
			Updates(map[string]interface{}{
				"status":   status,
				"ended_by": driverId,
				"ended_at": &now,
			}).Error; e != nil {
			return e
		}

		if e = tx.Where("id = ?", row.ServiceID).First(&service).Error; e != nil {
			return e
		}

		if status == constants.Membership_Status_Left && service.OwnerDriverID != driverId {
			notices = append(notices, driverNotice(service.OwnerDriverID, constants.NOTIFICATION_TYPE_PICKDROP_UPDATE, constants.NOTIFICATION_TITLE_PICKDROP_UPDATE,
				fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_LEFT, driver.DriverName, vehicleNumber(tx, vehicleId), service.Name)))
		}

		return nil
	})

	logger.LogInfo("Response returned from withdrawVehicle", sessionId)

	return
}

// leaveAsDriver withdraws a waiting request or leaves a service. A driver cannot be
// swapped off a shift automatically, so leaving is refused while they, or a vehicle
// they lent, are still on an active shift. The owner cannot leave at all.
func leaveAsDriver(orgCtx *gin.Context, sessionId, driverId string) (service postgress.PickDropService, notices []notice, err error) {
	logger.LogInfo("Request received in leaveAsDriver", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		driver, e := lockActiveDriver(tx, driverId)
		if e != nil {
			return e
		}

		var owned int64
		if e = tx.Model(&postgress.PickDropService{}).
			Where("owner_driver_id = ?", driverId).
			Where("status = ?", constants.Status_Active).
			Count(&owned).Error; e != nil {
			return e
		}

		if owned > 0 {
			return utils.Refuse(constants.Owner_Cannot_Leave)
		}

		var membership postgress.PickDropDriver
		if e = forUpdate(tx).Where("driver_id = ?", driverId).Where("status IN ?", openStatuses).First(&membership).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Not_Service_Member)
			}
			return e
		}

		if e = tx.Where("id = ?", membership.ServiceID).First(&service).Error; e != nil {
			return e
		}

		message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_WITHDRAWN, driver.DriverName, service.Name)

		if membership.Status == constants.Membership_Status_Approved {
			var vehicleIds []string
			if e = tx.Model(&postgress.PickDropVehicle{}).
				Where("service_id = ?", service.ID).
				Where("driver_id = ?", driverId).
				Where("status = ?", constants.Membership_Status_Approved).
				Pluck("vehicle_id", &vehicleIds).Error; e != nil {
				return e
			}

			driverShifts, vehicleShifts, e := database.CountActiveShiftAssignments(tx, driverId, vehicleIds)
			if e != nil {
				return e
			}

			if driverShifts > 0 {
				return utils.Refuse(fmt.Sprintf(constants.On_Active_Shifts, "You are", driverShifts))
			}

			if vehicleShifts > 0 {
				return utils.Refuse(fmt.Sprintf(constants.On_Active_Shifts, "One of your vehicles is", vehicleShifts))
			}

			message = fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_LEFT, driver.DriverName, service.Name)
		}

		if e = database.EndOpenMemberships(tx, &postgress.PickDropDriver{}, "driver_id", driverId, driverId); e != nil {
			return e
		}

		if e = database.EndOpenMemberships(tx, &postgress.PickDropVehicle{}, "driver_id", driverId, driverId); e != nil {
			return e
		}

		notices = append(notices, driverNotice(service.OwnerDriverID, constants.NOTIFICATION_TYPE_PICKDROP_UPDATE, constants.NOTIFICATION_TITLE_PICKDROP_UPDATE, message))

		return nil
	})

	logger.LogInfo("Response returned from leaveAsDriver", sessionId)

	return
}

func requestJoinAsPassenger(orgCtx *gin.Context, sessionId, passengerId string, request PassengerJoinRequest) (service postgress.PickDropService, requestId string, notices []notice, err error) {
	logger.LogInfo("Request received in requestJoinAsPassenger", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		passenger, e := lockActivePassenger(tx, passengerId)
		if e != nil {
			return e
		}

		if service, e = getActiveService(tx, request.ServiceId); e != nil {
			return e
		}

		if e = checkPassengerIsFree(tx, passengerId); e != nil {
			return e
		}

		membership := postgress.PickDropPassenger{
			ID:          database.GenerateUUID(),
			ServiceID:   service.ID,
			PassengerID: passengerId,
			Status:      constants.Membership_Status_Pending,
			RequestedAt: time.Now(),
		}

		if e = tx.Create(&membership).Error; e != nil {
			return e
		}
		requestId = membership.ID

		notices = append(notices, driverNotice(service.OwnerDriverID, constants.NOTIFICATION_TYPE_PICKDROP_REQUEST, constants.NOTIFICATION_TITLE_PICKDROP_REQUEST,
			fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_JOIN_REQUEST, passenger.PassengerName, service.Name, constants.Request_Type_Passenger)))

		return nil
	})

	logger.LogInfo("Response returned from requestJoinAsPassenger", sessionId)

	return
}

func getPassengerMembership(orgCtx *gin.Context, passengerId string) (data passengerMembershipData, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)
	data.Role = constants.Service_Role_None

	var membership postgress.PickDropPassenger
	err = db.Where("passenger_id = ?", passengerId).Where("status IN ?", openStatuses).First(&membership).Error
	if isNotFound(err) {
		err = nil
		return
	}
	if err != nil {
		return
	}

	var details postgress.PickDropServiceDetails
	if details, err = getServiceDetails(db, membership.ServiceID); err != nil {
		return
	}

	data.Role = constants.Service_Role_Passenger
	data.Service = &details
	data.Membership = &membership

	return
}

// leaveAsPassenger withdraws a waiting request or leaves a service. A passenger comes
// off their active shifts automatically and their seats go back to the owner, which
// is how removing a passenger already worked, so leaving is never blocked.
func leaveAsPassenger(orgCtx *gin.Context, sessionId, passengerId string) (service postgress.PickDropService, notices []notice, err error) {
	logger.LogInfo("Request received in leaveAsPassenger", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		passenger, e := lockActivePassenger(tx, passengerId)
		if e != nil {
			return e
		}

		var membership postgress.PickDropPassenger
		if e = forUpdate(tx).Where("passenger_id = ?", passengerId).Where("status IN ?", openStatuses).First(&membership).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Not_Service_Member)
			}
			return e
		}

		if e = tx.Where("id = ?", membership.ServiceID).First(&service).Error; e != nil {
			return e
		}

		message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_WITHDRAWN, passenger.PassengerName, service.Name)

		if membership.Status == constants.Membership_Status_Approved {
			if _, e = database.RemovePassengerFromActiveShifts(tx, passengerId, service.ID, constants.Removal_Reason_Left_Service, passengerId); e != nil {
				return e
			}

			message = fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_LEFT, passenger.PassengerName, service.Name)
		}

		if e = database.EndOpenMemberships(tx, &postgress.PickDropPassenger{}, "passenger_id", passengerId, passengerId); e != nil {
			return e
		}

		notices = append(notices, driverNotice(service.OwnerDriverID, constants.NOTIFICATION_TYPE_PICKDROP_UPDATE, constants.NOTIFICATION_TITLE_PICKDROP_UPDATE, message))

		return nil
	})

	logger.LogInfo("Response returned from leaveAsPassenger", sessionId)

	return
}

// applyStatusFilter narrows a request list to one status. No status means the
// requests still waiting, which is what an owner opens the list for.
func applyStatusFilter(query *gorm.DB, column, status string) *gorm.DB {
	switch status {
	case "":
		return query.Where(column+" = ?", constants.Membership_Status_Pending)
	case constants.Membership_Status_All:
		return query
	default:
		return query.Where(column+" = ?", status)
	}
}

func driverRequestJoins(db *gorm.DB) *gorm.DB {
	return db.Table("pick_drop_drivers").
		Joins("LEFT JOIN drivers ON drivers.id = pick_drop_drivers.driver_id")
}

const driverRequestColumns = `
	pick_drop_drivers.id,
	pick_drop_drivers.service_id,
	pick_drop_drivers.driver_id,
	COALESCE(drivers.driver_name, '') AS driver_name,
	COALESCE(drivers.driver_mobile, '') AS driver_mobile,
	COALESCE(drivers.rating, '') AS rating,
	pick_drop_drivers.join_type,
	pick_drop_drivers.status,
	pick_drop_drivers.requested_at,
	pick_drop_drivers.decided_at,
	pick_drop_drivers.ended_at`

func passengerRequestJoins(db *gorm.DB) *gorm.DB {
	return db.Table("pick_drop_passengers").
		Joins("LEFT JOIN passengers ON passengers.id = pick_drop_passengers.passenger_id")
}

const passengerRequestColumns = `
	pick_drop_passengers.id,
	pick_drop_passengers.service_id,
	pick_drop_passengers.passenger_id,
	COALESCE(passengers.passenger_name, '') AS passenger_name,
	COALESCE(passengers.passenger_mobile, '') AS passenger_mobile,
	COALESCE(passengers.gender, '') AS gender,
	pick_drop_passengers.status,
	pick_drop_passengers.requested_at,
	pick_drop_passengers.decided_at,
	pick_drop_passengers.ended_at,
	(SELECT COUNT(*) FROM shift_passengers
		JOIN shifts ON shifts.id = shift_passengers.shift_id AND shifts.status = ?
		WHERE shift_passengers.passenger_id = pick_drop_passengers.passenger_id
		  AND shift_passengers.status = ?) AS active_shifts`

// getDriverRequests reads one page of driver requests and, in one more query, every
// vehicle those drivers offered to the service.
func getDriverRequests(orgCtx *gin.Context, serviceId, status string, page int) (rows []postgress.PickDropDriverDetails, vehicles []postgress.PickDropVehicleDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	query := applyStatusFilter(driverRequestJoins(db).Where("pick_drop_drivers.service_id = ?", serviceId), "pick_drop_drivers.status", status)

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	if err = query.Select(driverRequestColumns).
		Order("pick_drop_drivers.requested_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		return
	}

	driverIds := make([]string, 0, len(rows))
	for _, row := range rows {
		driverIds = append(driverIds, row.DriverID)
	}

	vehicles, err = getDriverVehicleOffers(db, serviceId, driverIds)

	return
}

func getVehicleRequests(orgCtx *gin.Context, serviceId, status string, page int) (rows []postgress.PickDropVehicleDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	query := applyStatusFilter(vehicleJoins(database.DatabaseConn.Postgres.WithContext(ctx)).
		Where("pick_drop_vehicles.service_id = ?", serviceId), "pick_drop_vehicles.status", status)

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Select(vehicleColumns, constants.Shift_Status_Active).
		Order("pick_drop_vehicles.requested_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error

	return
}

func getPassengerRequests(orgCtx *gin.Context, serviceId, status string, page int) (rows []postgress.PickDropPassengerDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	query := applyStatusFilter(passengerRequestJoins(database.DatabaseConn.Postgres.WithContext(ctx)).
		Where("pick_drop_passengers.service_id = ?", serviceId), "pick_drop_passengers.status", status)

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Select(passengerRequestColumns, constants.Shift_Status_Active, constants.Shift_Passenger_Active).
		Order("pick_drop_passengers.requested_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error

	return
}

func getDriverRequest(orgCtx *gin.Context, serviceId, requestId string) (row postgress.PickDropDriverDetails, vehicles []postgress.PickDropVehicleDetails, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	if err = driverRequestJoins(db).
		Select(driverRequestColumns).
		Where("pick_drop_drivers.service_id = ?", serviceId).
		Where("pick_drop_drivers.id = ?", requestId).
		Take(&row).Error; err != nil {
		return
	}

	vehicles, err = getDriverVehicleOffers(db, serviceId, []string{row.DriverID})

	return
}

func getVehicleRequest(orgCtx *gin.Context, serviceId, requestId string) (row postgress.PickDropVehicleDetails, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = vehicleJoins(database.DatabaseConn.Postgres.WithContext(ctx)).
		Select(vehicleColumns, constants.Shift_Status_Active).
		Where("pick_drop_vehicles.service_id = ?", serviceId).
		Where("pick_drop_vehicles.id = ?", requestId).
		Take(&row).Error

	return
}

func getPassengerRequest(orgCtx *gin.Context, serviceId, requestId string) (
	row postgress.PickDropPassengerDetails,
	days []postgress.PassengerAvailability,
	locations []postgress.PassengerAvailabilityLocation,
	err error,
) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	if err = passengerRequestJoins(db).
		Select(passengerRequestColumns, constants.Shift_Status_Active, constants.Shift_Passenger_Active).
		Where("pick_drop_passengers.service_id = ?", serviceId).
		Where("pick_drop_passengers.id = ?", requestId).
		Take(&row).Error; err != nil {
		return
	}

	days, locations, err = getWeeklyDemand(db, []string{row.PassengerID})

	return
}

// getWeeklyDemand reads the weekly availability of a whole page of passengers in two
// queries, never one per passenger.
func getWeeklyDemand(db *gorm.DB, passengerIds []string) (days []postgress.PassengerAvailability, locations []postgress.PassengerAvailabilityLocation, err error) {
	if len(passengerIds) == 0 {
		return
	}

	if err = db.Where("passenger_id IN ?", passengerIds).
		Order("day_of_week ASC").
		Find(&days).Error; err != nil {
		return
	}

	err = db.Where("passenger_id IN ?", passengerIds).
		Order("day_of_week ASC, sequence ASC").
		Find(&locations).Error

	return
}

// decideDriverRequest applies one decision on a driver request in its own
// transaction, with the request row locked so two owners' devices cannot decide it
// twice.
func decideDriverRequest(orgCtx *gin.Context, service postgress.PickDropService, ownerId, requestId, action string) (notices []notice, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row postgress.PickDropDriver
		if e := forUpdate(tx).Where("id = ?", requestId).Where("service_id = ?", service.ID).First(&row).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Request_Not_Found)
			}
			return e
		}

		now := time.Now()
		title := constants.NOTIFICATION_TITLE_PICKDROP_DECISION
		notificationType := constants.NOTIFICATION_TYPE_PICKDROP_DECISION

		switch action {
		case constants.Request_Action_Approve:
			if row.Status != constants.Membership_Status_Pending {
				return utils.Refuse(fmt.Sprintf(constants.Request_Already_Decided, row.Status))
			}

			if _, e := lockActiveDriver(tx, row.DriverID); e != nil {
				return e
			}

			// joining with vehicles only means nothing without a vehicle still on offer
			if row.JoinType == constants.Join_Type_Vehicles {
				var offered int64
				if e := tx.Model(&postgress.PickDropVehicle{}).
					Where("service_id = ?", service.ID).
					Where("driver_id = ?", row.DriverID).
					Where("status IN ?", openStatuses).
					Count(&offered).Error; e != nil {
					return e
				}

				if offered == 0 {
					return utils.Refuse(constants.Vehicles_Join_Needs_Vehicle)
				}
			}

			if e := tx.Model(&postgress.PickDropDriver{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
				"status":     constants.Membership_Status_Approved,
				"decided_by": ownerId,
				"decided_at": &now,
			}).Error; e != nil {
				return e
			}

			notices = append(notices, driverNotice(row.DriverID, notificationType, title, fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_APPROVED, service.Name)))

		case constants.Request_Action_Reject:
			if row.Status != constants.Membership_Status_Pending {
				return utils.Refuse(fmt.Sprintf(constants.Request_Already_Decided, row.Status))
			}

			for _, model := range []interface{}{&postgress.PickDropDriver{}, &postgress.PickDropVehicle{}} {
				query := tx.Model(model).Where("status = ?", constants.Membership_Status_Pending)
				if _, isDriver := model.(*postgress.PickDropDriver); isDriver {
					query = query.Where("id = ?", row.ID)
				} else {
					query = query.Where("service_id = ?", service.ID).Where("driver_id = ?", row.DriverID)
				}

				if e := query.Updates(map[string]interface{}{
					"status":     constants.Membership_Status_Rejected,
					"decided_by": ownerId,
					"decided_at": &now,
				}).Error; e != nil {
					return e
				}
			}

			notices = append(notices, driverNotice(row.DriverID, notificationType, title, fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_REJECTED, service.Name)))

		case constants.Request_Action_Remove:
			if row.Status != constants.Membership_Status_Approved {
				return utils.Refuse(constants.Only_Approved_Removable)
			}

			var vehicleIds []string
			if e := tx.Model(&postgress.PickDropVehicle{}).
				Where("service_id = ?", service.ID).
				Where("driver_id = ?", row.DriverID).
				Where("status = ?", constants.Membership_Status_Approved).
				Pluck("vehicle_id", &vehicleIds).Error; e != nil {
				return e
			}

			driverShifts, vehicleShifts, e := database.CountActiveShiftAssignments(tx, row.DriverID, vehicleIds)
			if e != nil {
				return e
			}

			if driverShifts > 0 {
				return utils.Refuse(fmt.Sprintf(constants.On_Active_Shifts, "This driver is", driverShifts))
			}

			if vehicleShifts > 0 {
				return utils.Refuse(fmt.Sprintf(constants.On_Active_Shifts, "A vehicle of this driver is", vehicleShifts))
			}

			removal := func() map[string]interface{} {
				return map[string]interface{}{
					"status":   constants.Membership_Status_Removed,
					"ended_by": ownerId,
					"ended_at": &now,
				}
			}

			if e = tx.Model(&postgress.PickDropDriver{}).Where("id = ?", row.ID).Updates(removal()).Error; e != nil {
				return e
			}

			if e = tx.Model(&postgress.PickDropVehicle{}).
				Where("service_id = ?", service.ID).
				Where("driver_id = ?", row.DriverID).
				Where("status IN ?", openStatuses).
				Updates(removal()).Error; e != nil {
				return e
			}

			notices = append(notices, driverNotice(row.DriverID, notificationType, title, fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_REMOVED, service.Name)))
		}

		return nil
	})

	return
}

// decideVehicleRequest applies one decision on a vehicle offer. A vehicle can only be
// approved once its driver is, the owner decides on the person before their vehicle.
func decideVehicleRequest(orgCtx *gin.Context, service postgress.PickDropService, ownerId, requestId, action string) (notices []notice, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row postgress.PickDropVehicle
		if e := forUpdate(tx).Where("id = ?", requestId).Where("service_id = ?", service.ID).First(&row).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Request_Not_Found)
			}
			return e
		}

		number := vehicleNumber(tx, row.VehicleID)
		now := time.Now()
		title := constants.NOTIFICATION_TITLE_PICKDROP_DECISION
		notificationType := constants.NOTIFICATION_TYPE_PICKDROP_DECISION

		var message string

		switch action {
		case constants.Request_Action_Approve:
			if row.Status != constants.Membership_Status_Pending {
				return utils.Refuse(fmt.Sprintf(constants.Request_Already_Decided, row.Status))
			}

			var vehicle postgress.Vehicle
			if e := tx.Where("id = ?", row.VehicleID).Where("status = ?", constants.Status_Active).First(&vehicle).Error; e != nil {
				if isNotFound(e) {
					return utils.Refuse(constants.Vehicle_Not_Found)
				}
				return e
			}

			var approvedDriver int64
			if e := tx.Model(&postgress.PickDropDriver{}).
				Where("service_id = ?", service.ID).
				Where("driver_id = ?", row.DriverID).
				Where("status = ?", constants.Membership_Status_Approved).
				Count(&approvedDriver).Error; e != nil {
				return e
			}

			if approvedDriver == 0 && row.DriverID != service.OwnerDriverID {
				return utils.Refuse(fmt.Sprintf(constants.Vehicle_Driver_Not_Approved, number))
			}

			if e := tx.Model(&postgress.PickDropVehicle{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
				"status":     constants.Membership_Status_Approved,
				"decided_by": ownerId,
				"decided_at": &now,
			}).Error; e != nil {
				return e
			}

			message = fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_APPROVED, number, service.Name)

		case constants.Request_Action_Reject:
			if row.Status != constants.Membership_Status_Pending {
				return utils.Refuse(fmt.Sprintf(constants.Request_Already_Decided, row.Status))
			}

			if e := tx.Model(&postgress.PickDropVehicle{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
				"status":     constants.Membership_Status_Rejected,
				"decided_by": ownerId,
				"decided_at": &now,
			}).Error; e != nil {
				return e
			}

			message = fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_REJECTED, number, service.Name)

		case constants.Request_Action_Remove:
			if row.Status != constants.Membership_Status_Approved {
				return utils.Refuse(constants.Only_Approved_Removable)
			}

			_, vehicleShifts, e := database.CountActiveShiftAssignments(tx, "", []string{row.VehicleID})
			if e != nil {
				return e
			}

			if vehicleShifts > 0 {
				return utils.Refuse(fmt.Sprintf(constants.On_Active_Shifts, "This vehicle is", vehicleShifts))
			}

			if e = tx.Model(&postgress.PickDropVehicle{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
				"status":   constants.Membership_Status_Removed,
				"ended_by": ownerId,
				"ended_at": &now,
			}).Error; e != nil {
				return e
			}

			message = fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_REMOVED, number, service.Name)
		}

		if row.DriverID != ownerId {
			notices = append(notices, driverNotice(row.DriverID, notificationType, title, message))
		}

		return nil
	})

	return
}

// decidePassengerRequest applies one decision on a passenger request. Removing an
// approved passenger takes them off their active shifts in the same transaction.
func decidePassengerRequest(orgCtx *gin.Context, service postgress.PickDropService, ownerId, requestId, action string) (notices []notice, err error) {
	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row postgress.PickDropPassenger
		if e := forUpdate(tx).Where("id = ?", requestId).Where("service_id = ?", service.ID).First(&row).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Request_Not_Found)
			}
			return e
		}

		now := time.Now()
		title := constants.NOTIFICATION_TITLE_PICKDROP_DECISION
		notificationType := constants.NOTIFICATION_TYPE_PICKDROP_DECISION

		switch action {
		case constants.Request_Action_Approve, constants.Request_Action_Reject:
			if row.Status != constants.Membership_Status_Pending {
				return utils.Refuse(fmt.Sprintf(constants.Request_Already_Decided, row.Status))
			}

			status := constants.Membership_Status_Rejected
			message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_REJECTED, service.Name)

			if action == constants.Request_Action_Approve {
				if _, e := lockActivePassenger(tx, row.PassengerID); e != nil {
					return e
				}

				status = constants.Membership_Status_Approved
				message = fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_APPROVED, service.Name)
			}

			if e := tx.Model(&postgress.PickDropPassenger{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
				"status":     status,
				"decided_by": ownerId,
				"decided_at": &now,
			}).Error; e != nil {
				return e
			}

			notices = append(notices, passengerNotice(row.PassengerID, notificationType, title, message))

		case constants.Request_Action_Remove:
			if row.Status != constants.Membership_Status_Approved {
				return utils.Refuse(constants.Only_Approved_Removable)
			}

			if _, e := database.RemovePassengerFromActiveShifts(tx, row.PassengerID, service.ID, constants.Removal_Reason_Removed_Service, ownerId); e != nil {
				return e
			}

			if e := tx.Model(&postgress.PickDropPassenger{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
				"status":   constants.Membership_Status_Removed,
				"ended_by": ownerId,
				"ended_at": &now,
			}).Error; e != nil {
				return e
			}

			notices = append(notices, passengerNotice(row.PassengerID, notificationType, title, fmt.Sprintf(constants.NOTIFICATION_MESSAGE_PICKDROP_REMOVED, service.Name)))
		}

		return nil
	})

	return
}

// getAvailableDrivers lists who can drive a shift: the owner first, then every
// approved driver, each with how many active shifts they already drive. A member who
// joined with vehicles only is never listed, they do not drive.
func getAvailableDrivers(orgCtx *gin.Context, service postgress.PickDropService, page int) (rows []postgress.AvailableDriverDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	params := map[string]interface{}{
		"owner":      service.OwnerDriverID,
		"service":    service.ID,
		"active":     constants.Status_Active,
		"approved":   constants.Membership_Status_Approved,
		"driverJoin": constants.Join_Type_Driver,
		"limit":      limit,
		"offset":     offset,
	}

	available := `
		SELECT drivers.id AS driver_id, drivers.driver_name, drivers.driver_mobile, COALESCE(drivers.rating, '') AS rating,
			true AS is_owner, pick_drop_services.created_at AS joined_at
		FROM drivers
		JOIN pick_drop_services ON pick_drop_services.id = @service
		WHERE drivers.id = @owner AND drivers.status = @active
		UNION ALL
		SELECT drivers.id, drivers.driver_name, drivers.driver_mobile, COALESCE(drivers.rating, ''),
			false, pick_drop_drivers.decided_at
		FROM pick_drop_drivers
		JOIN drivers ON drivers.id = pick_drop_drivers.driver_id
		WHERE pick_drop_drivers.service_id = @service
		  AND pick_drop_drivers.status = @approved
		  AND pick_drop_drivers.join_type = @driverJoin
		  AND drivers.status = @active`

	if err = db.Raw("SELECT COUNT(*) FROM ("+available+") available", params).Scan(&totalRows).Error; err != nil {
		return
	}

	err = db.Raw(`
		SELECT available.*,
			(SELECT COUNT(*) FROM shifts WHERE shifts.driver_id = available.driver_id AND shifts.status = @active) AS active_shifts
		FROM (`+available+`) available
		ORDER BY available.is_owner DESC, available.driver_name ASC
		LIMIT @limit OFFSET @offset`, params).Scan(&rows).Error

	return
}

func getAvailableVehicles(orgCtx *gin.Context, serviceId string, page int) (rows []postgress.PickDropVehicleDetails, totalRows int64, err error) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	query := vehicleJoins(database.DatabaseConn.Postgres.WithContext(ctx)).
		Where("pick_drop_vehicles.service_id = ?", serviceId).
		Where("pick_drop_vehicles.status = ?", constants.Membership_Status_Approved).
		Where("vehicles.status = ?", constants.Status_Active)

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	err = query.Select(vehicleColumns, constants.Shift_Status_Active).
		Order("vehicles.vehicle_number ASC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error

	return
}

// getAvailablePassengers lists the approved passengers with their weekly demand. A
// day filter keeps only the passengers who need the service on that day.
func getAvailablePassengers(orgCtx *gin.Context, serviceId string, dayOfWeek, page int) (
	rows []postgress.PickDropPassengerDetails,
	days []postgress.PassengerAvailability,
	locations []postgress.PassengerAvailabilityLocation,
	totalRows int64,
	err error,
) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	query := passengerRequestJoins(db).
		Where("pick_drop_passengers.service_id = ?", serviceId).
		Where("pick_drop_passengers.status = ?", constants.Membership_Status_Approved).
		Where("passengers.status = ?", constants.Status_Active)

	if dayOfWeek > 0 {
		query = query.Where(`EXISTS (
			SELECT 1 FROM passenger_availabilities
			WHERE passenger_availabilities.passenger_id = pick_drop_passengers.passenger_id
			  AND passenger_availabilities.day_of_week = ?
			  AND passenger_availabilities.is_required = true
		)`, dayOfWeek)
	}

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	if err = query.Select(passengerRequestColumns, constants.Shift_Status_Active, constants.Shift_Passenger_Active).
		Order("passengers.passenger_name ASC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		return
	}

	passengerIds := make([]string, 0, len(rows))
	for _, row := range rows {
		passengerIds = append(passengerIds, row.PassengerID)
	}

	days, locations, err = getWeeklyDemand(db, passengerIds)

	return
}

// findConflicts checks a whole page of drivers, vehicles or passengers against a
// planned schedule at once: overlapping shifts in any service, and for drivers and
// vehicles overlapping rides too.
func findConflicts(orgCtx *gin.Context, filter AvailabilityFilter, driverIds, vehicleIds, passengerIds []string) (conflicts map[string]database.ShiftClash, err error) {
	conflicts = map[string]database.ShiftClash{}

	if !filter.Enabled() {
		return
	}

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	window := database.ShiftWindow{
		DaysOfWeek: filter.DaysOfWeek,
		StartDate:  filter.StartDate,
		EndDate:    filter.EndDate,
		StartTime:  filter.StartTime,
		EndTime:    filter.EndTime,
	}

	if conflicts, err = database.FindShiftClashes(db, window, driverIds, vehicleIds, passengerIds, filter.ExcludeShiftId); err != nil {
		return
	}

	rides, err := database.FindRideClashes(db, window, driverIds, vehicleIds)
	if err != nil {
		return
	}

	for id, clash := range rides {
		if _, known := conflicts[id]; !known {
			conflicts[id] = clash
		}
	}

	return
}

// createAdvertisement writes an advertisement once the service still has room for
// one. The limit is two per approved vehicle, worked out inside the transaction with
// the service row locked, so two advertisements created at once cannot both squeeze
// into the last place.
func createAdvertisement(orgCtx *gin.Context, sessionId, ownerId string, request AdvertisementRequest) (adId string, err error) {
	logger.LogInfo("Request received in createAdvertisement", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		service, e := lockOwnedService(tx, ownerId)
		if e != nil {
			return e
		}

		approvedVehicles, e := countApprovedVehicles(tx, service.ID)
		if e != nil {
			return e
		}

		limit := approvedVehicles * int64(constants.Advertisements_Per_Vehicle)

		var existing int64
		if e = tx.Model(&postgress.PickDropAdvertisement{}).Where("service_id = ?", service.ID).Count(&existing).Error; e != nil {
			return e
		}

		if existing >= limit {
			return utils.Refuse(fmt.Sprintf(constants.Advertisement_Limit, approvedVehicles, limit))
		}

		points := utils.RoutePoints(request.Locations)

		ad := postgress.PickDropAdvertisement{
			ID:            database.GenerateUUID(),
			ServiceID:     service.ID,
			Title:         request.Title,
			Description:   request.Description,
			Fare:          request.Fare,
			DaysOfWeek:    database.IntArray(request.DaysOfWeek),
			StartTime:     request.StartTime,
			EndTime:       request.EndTime,
			StartLocation: points[0],
			EndLocation:   points[len(points)-1],
			RoutePoints:   pq.StringArray(points),
			CreatedBy:     ownerId,
		}

		if e = tx.Create(&ad).Error; e != nil {
			return e
		}
		adId = ad.ID

		locations := make([]postgress.PickDropAdvertisementLocation, 0, len(request.Locations))
		for index, location := range request.Locations {
			locations = append(locations, postgress.PickDropAdvertisementLocation{
				ID:              database.GenerateUUID(),
				AdvertisementID: ad.ID,
				Sequence:        index + 1,
				Location:        location.Location,
				Lat:             location.Lat,
				Lng:             location.Lng,
				Time:            location.Time,
			})
		}

		return tx.Create(&locations).Error
	})

	logger.LogInfo("Response returned from createAdvertisement", sessionId)

	return
}

// queryAdvertisements is both the owner's list and the public search. Only
// advertisements of active services are ever shown, and the locations of the whole
// page come back in one more query.
func queryAdvertisements(orgCtx *gin.Context, serviceId string, filter AdvertisementSearch, page int) (
	ads []postgress.AdvertisementDetails,
	locations []postgress.PickDropAdvertisementLocation,
	totalRows int64,
	err error,
) {
	limit, offset := pageBounds(page)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	db := database.DatabaseConn.Postgres.WithContext(ctx)

	query := db.Table("pick_drop_advertisements").
		Joins("JOIN pick_drop_services ON pick_drop_services.id = pick_drop_advertisements.service_id AND pick_drop_services.status = ?", constants.Status_Active).
		Joins("LEFT JOIN drivers ON drivers.id = pick_drop_services.owner_driver_id")

	order := "pick_drop_advertisements.start_time ASC, pick_drop_advertisements.created_at DESC"

	if !utils.IsStringEmpty(serviceId) {
		query = query.Where("pick_drop_advertisements.service_id = ?", serviceId)
		order = "pick_drop_advertisements.created_at DESC"
	}

	if !utils.IsStringEmpty(filter.Search) {
		term := "%" + filter.Search + "%"
		query = query.Where(`(
			pick_drop_advertisements.start_location ILIKE ?
			OR pick_drop_advertisements.end_location ILIKE ?
			OR EXISTS (SELECT 1 FROM unnest(pick_drop_advertisements.route_points) AS point WHERE point ILIKE ?)
		)`, term, term, term)
	}

	if filter.DayOfWeek > 0 {
		query = query.Where("? = ANY(pick_drop_advertisements.days_of_week)", filter.DayOfWeek)
	}

	if !utils.IsStringEmpty(filter.StartTime) {
		query = query.Where("pick_drop_advertisements.start_time >= ?", filter.StartTime)
	}

	if !utils.IsStringEmpty(filter.EndTime) {
		query = query.Where("pick_drop_advertisements.end_time <= ?", filter.EndTime)
	}

	if err = query.Session(&gorm.Session{}).Count(&totalRows).Error; err != nil {
		return
	}

	if err = query.Select(`
			pick_drop_advertisements.id,
			pick_drop_advertisements.service_id,
			pick_drop_services.name AS service_name,
			pick_drop_services.owner_driver_id,
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
		Order(order).
		Limit(limit).
		Offset(offset).
		Find(&ads).Error; err != nil {
		return
	}

	if len(ads) == 0 {
		return
	}

	adIds := make([]string, 0, len(ads))
	for _, ad := range ads {
		adIds = append(adIds, ad.ID)
	}

	err = db.Where("advertisement_id IN ?", adIds).
		Order("advertisement_id ASC, sequence ASC").
		Find(&locations).Error

	return
}

func deleteAdvertisement(orgCtx *gin.Context, sessionId, ownerId, adId string) (err error) {
	logger.LogInfo("Request received in deleteAdvertisement", sessionId)

	ctx, cancel := withTimeout(orgCtx)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		service, e := lockOwnedService(tx, ownerId)
		if e != nil {
			return e
		}

		var ad postgress.PickDropAdvertisement
		if e = forUpdate(tx).Where("id = ?", adId).Where("service_id = ?", service.ID).First(&ad).Error; e != nil {
			if isNotFound(e) {
				return utils.Refuse(constants.Advertisement_Not_Found)
			}
			return e
		}

		return archiveAdvertisements(tx, []postgress.PickDropAdvertisement{ad}, ownerId)
	})

	logger.LogInfo("Response returned from deleteAdvertisement", sessionId)

	return
}
