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

		&DELVehicle{},
		&DELRide{},
		&DELDriver{},
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

	return
}
