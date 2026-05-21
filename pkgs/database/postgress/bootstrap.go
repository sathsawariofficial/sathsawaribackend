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
		&DriverFCM{},
		&NotificationRequest{},
		&SMSFCM{},
		&MissingLocations{},
		&DriverDevice{},
	)

	if err != nil {
		panic(err)
	}

	if db == nil {
		logger.LogFatal(constants.DEFAULT_SESSION, "failed to connect to postgres")
	}

	return
}
