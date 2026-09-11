package postgress

import (
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPortgress() (db *gorm.DB, err error) {
	postgressConf := configuration.ConfigurationData.Database.Postgres

	dsn := fmt.Sprintf(
		"host=%v user=%v password=%v dbname=%v port=%v sslmode=disable",
		postgressConf.Host,
		postgressConf.User,
		postgressConf.Password,
		postgressConf.Name,
		postgressConf.Port,
	)

	fmt.Println("Configuration", dsn)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	// Get underlying sql.DB for pool config (VERY IMPORTANT in v2)
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// Connection pool tuning (highly recommended)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// the first group and shift design never reached production and its shifts table
	// has a different shape, so it has to go before the new one is migrated over it
	if err = dropLegacyFleetTables(db); err != nil {
		panic("failed to drop the legacy fleet tables: " + err.Error())
	}

	// Auto migrate models
	err = db.AutoMigrate(
		&Admin{},
		&Driver{},
		&Vehicle{},
		&Ride{},
		&PassengerRating{},
		&ApprochInfo{},
		&RideTemplate{},
		&UserFCM{},
		&NotificationRequest{},
		&SMSFCM{},
		&MissingLocations{},
		&DriverDevice{},
		&BroadcastNotificationRequests{},
		&RideBooking{},
		&RideRequest{},
		&AnnouncementRequests{},
		&PlaceNotificationSetting{},

		&Passenger{},
		&PassengerAvailability{},
		&PassengerAvailabilityLocation{},

		&PickDropService{},
		&PickDropDriver{},
		&PickDropVehicle{},
		&PickDropPassenger{},
		&PickDropAdvertisement{},
		&PickDropAdvertisementLocation{},

		&ShiftRequest{},
		&ShiftRequestLocation{},
		&Shift{},
		&ShiftLocation{},
		&ShiftPassenger{},
		&ShiftOccurrence{},
		&ShiftAttendance{},
		&ShiftDriverUpdate{},

		&DELVehicle{},
		&DELRide{},
		&DELDriver{},
		&DELPassenger{},
		&DELPassengerAvailability{},
		&DELPassengerAvailabilityLocation{},
		&DELPickDropAdvertisement{},
		&DELPickDropAdvertisementLocation{},
		&DELShiftRequest{},
		&DELShiftRequestLocation{},
		&DELShift{},
		&DELShiftLocation{},
	)

	if err != nil {
		panic(err)
	}

	if db == nil {
		logger.LogFatal(constants.DEFAULT_SESSION, "failed to connect to postgres")
	}

	// --- START: CQRS SEARCH OPTIMIZATION MIGRATION ---
	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Enable the Trigram extension for text matching
		if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm;`).Error; err != nil {
			return err
		}

		// 2. Create the dedicated, lightweight search table
		if err := tx.Exec(`
            CREATE TABLE IF NOT EXISTS ride_searches (
                ride_id TEXT PRIMARY KEY,
                start_location TEXT NOT NULL,
                end_location TEXT NOT NULL,
                route_points TEXT[] NOT NULL,
                start_datetime TEXT NOT NULL,
                available_seats INT NOT NULL,
                is_active BOOLEAN NOT NULL
            );
        `).Error; err != nil {
			return err
		}

		// 3. Create GIN and Trigram indexes for fast searches
		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_search_start_loc ON ride_searches USING gin (start_location gin_trgm_ops);`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_search_end_loc ON ride_searches USING gin (end_location gin_trgm_ops);`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_search_route_points ON ride_searches USING gin (route_points);`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_search_composite_filter ON ride_searches (is_active, available_seats, start_datetime);`).Error; err != nil {
			return err
		}

		// 4. Create the Sync Function (Handles Insert/Update and calculates remaining seats)
		if err := tx.Exec(`
            CREATE OR REPLACE FUNCTION sync_rides_to_search_replica()
            RETURNS TRIGGER AS $$
            BEGIN
                IF NEW.is_active = false THEN
                    DELETE FROM ride_searches WHERE ride_id = NEW.id;
                    RETURN NEW;
                END IF;

                INSERT INTO ride_searches (ride_id, start_location, end_location, route_points, start_datetime, available_seats, is_active)
                VALUES (
                    NEW.id,
                    NEW.start_location,
                    NEW.end_location,
                    NEW.route_points,
                    NEW.start_datetime,
                    (NEW.number_of_seats - NEW.seats_taken), -- Compute actual available seats
                    NEW.is_active
                )
                ON CONFLICT (ride_id) DO UPDATE SET
                    start_location = EXCLUDED.start_location,
                    end_location = EXCLUDED.end_location,
                    route_points = EXCLUDED.route_points,
                    start_datetime = EXCLUDED.start_datetime,
                    available_seats = EXCLUDED.available_seats,
                    is_active = EXCLUDED.is_active;
                RETURN NEW;
            END;
            $$ LANGUAGE plpgsql;
        `).Error; err != nil {
			return err
		}

		// 5. Create the Deletion Sync Function
		if err := tx.Exec(`
            CREATE OR REPLACE FUNCTION sync_delete_rides_to_search_replica()
            RETURNS TRIGGER AS $$
            BEGIN
                DELETE FROM ride_searches WHERE ride_id = OLD.id;
                RETURN OLD;
            END;
            $$ LANGUAGE plpgsql;
        `).Error; err != nil {
			return err
		}

		// 6. Bind the Sync Trigger (Drops existing first to prevent duplication errors)
		if err := tx.Exec(`DROP TRIGGER IF EXISTS trg_sync_rides ON rides;`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
            CREATE TRIGGER trg_sync_rides
            AFTER INSERT OR UPDATE ON rides
            FOR EACH ROW
            EXECUTE FUNCTION sync_rides_to_search_replica();
        `).Error; err != nil {
			return err
		}

		// 7. Bind the Delete Trigger
		if err := tx.Exec(`DROP TRIGGER IF EXISTS trg_sync_delete_rides ON rides;`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
            CREATE TRIGGER trg_sync_delete_rides
            AFTER DELETE ON rides
            FOR EACH ROW
            EXECUTE FUNCTION sync_delete_rides_to_search_replica();
        `).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
            INSERT INTO ride_searches (ride_id, start_location, end_location, route_points, start_datetime, available_seats, is_active)
            SELECT
                id,
                start_location,
                end_location,
                route_points,
                start_datetime,
                (number_of_seats - seats_taken),
                is_active
            FROM rides
            WHERE is_active = true
            ON CONFLICT (ride_id) DO UPDATE SET
                start_location = EXCLUDED.start_location,
                end_location = EXCLUDED.end_location,
                route_points = EXCLUDED.route_points,
                start_datetime = EXCLUDED.start_datetime,
                available_seats = EXCLUDED.available_seats,
                is_active = EXCLUDED.is_active;
        `).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
            DELETE FROM ride_searches
            WHERE NOT EXISTS (
                SELECT 1
                FROM rides
                WHERE rides.id = ride_searches.ride_id
                  AND rides.is_active = true
            );
        `).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		panic("failed to complete migration optimizations: " + err.Error())
	}
	// --- END: CQRS SEARCH OPTIMIZATION MIGRATION ---

	db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_rides_route_points
		ON rides
		USING GIN(route_points)
	`)

	if err = createPickDropConstraints(db); err != nil {
		panic("failed to create the pick & drop constraints: " + err.Error())
	}

	// a place alert looks settings up by any one of their places
	db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_place_notification_settings_places
		ON place_notification_settings
		USING GIN(places)
	`)

	return
}

// dropLegacyFleetTables removes the first group and shift design. It only ever runs
// against a database that still has that design, recognised by the old shifts table
// carrying a group_id column, so on production, which never had it, and on every
// boot after the first, it does nothing.
func dropLegacyFleetTables(db *gorm.DB) error {
	var legacyColumns int64
	if err := db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = 'shifts'
		  AND column_name = 'group_id'
	`).Scan(&legacyColumns).Error; err != nil {
		return err
	}

	if legacyColumns == 0 {
		return nil
	}

	logger.LogWarning(constants.DEFAULT_SESSION, "dropping the legacy group and shift tables, they are replaced by pick & drop")

	return db.Exec(`
		DROP TABLE IF EXISTS
			shift_template_seats,
			shift_template_stops,
			shift_templates,
			shift_seats,
			shift_stops,
			shifts,
			group_passengers,
			group_vehicles,
			group_members,
			groups,
			role_permissions,
			permissions,
			roles,
			del_roles,
			passenger_location_preferences,
			del_passenger_location_preferences
	`).Error
}

// createPickDropConstraints puts the rules the business depends on into the database
// itself, so they hold even if two requests race past the checks in the code:
// one open membership per driver, passenger and vehicle, one active service per
// owner, one active place per passenger per shift, and never more passengers than
// seats. Everything here is idempotent.
func createPickDropConstraints(db *gorm.DB) error {
	statements := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_pick_drop_services_active_owner
			ON pick_drop_services (owner_driver_id) WHERE status = 'active'`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_pick_drop_drivers_open
			ON pick_drop_drivers (driver_id) WHERE status IN ('pending', 'approved')`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_pick_drop_passengers_open
			ON pick_drop_passengers (passenger_id) WHERE status IN ('pending', 'approved')`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_pick_drop_vehicles_open
			ON pick_drop_vehicles (vehicle_id) WHERE status IN ('pending', 'approved')`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_shift_passengers_active
			ON shift_passengers (shift_id, passenger_id) WHERE status = 'active'`,
		`CREATE INDEX IF NOT EXISTS idx_shifts_days_of_week ON shifts USING gin (days_of_week)`,
		`CREATE INDEX IF NOT EXISTS idx_pick_drop_ads_route_points ON pick_drop_advertisements USING gin (route_points)`,
		`CREATE INDEX IF NOT EXISTS idx_pick_drop_ads_start_loc ON pick_drop_advertisements USING gin (start_location gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_pick_drop_ads_end_loc ON pick_drop_advertisements USING gin (end_location gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_shift_requests_route_points ON shift_requests USING gin (route_points)`,
		`CREATE INDEX IF NOT EXISTS idx_shift_requests_start_loc ON shift_requests USING gin (start_location gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_shift_requests_end_loc ON shift_requests USING gin (end_location gin_trgm_ops)`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_shifts_occupied_seats') THEN
				ALTER TABLE shifts ADD CONSTRAINT chk_shifts_occupied_seats
					CHECK (occupied_seats >= 0 AND occupied_seats <= seat_capacity);
			END IF;
		END $$`,
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
