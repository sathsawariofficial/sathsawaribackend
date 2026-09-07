// main.go
package main

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	httpcall "rideshare/pkgs/externalCall/http"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/middleware"
	"rideshare/pkgs/monitoring"
	"rideshare/services/admin"
	"rideshare/services/driver"
	general "rideshare/services/general/rest"
	general_rest "rideshare/services/general/rest"
	"rideshare/services/general/socket"
	"rideshare/services/group"
	"rideshare/services/passenger"
	"rideshare/services/ride"
	"rideshare/services/shift"
	"rideshare/worker"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	router := gin.Default()

	// CORS
	config := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router.Use(cors.New(config))
	router.Use(middleware.Logger(), middleware.RecoveryMiddleware(), middleware.StatsMiddleware(), middleware.HealthMiddleware())

	setup()

	router.GET("/", general_rest.GetHomePage)
	router.GET("/terms", general_rest.GetTermsAndConditions)
	router.GET("/privacy-policy", general_rest.GetPrivacyPolicy)
	router.GET("/r/:id", ride.GetOpenRideHandler)
	router.GET("/rq/:id", ride.GetOpenRideHandler)

	v1 := router.Group("/api/v1")
	{
		public := v1.Group("")
		public.Use(middleware.Authentication(constants.OPEN_TOKEN))
		{
			public.POST("/otp/resend", general.ResendOTPHandler)
			public.POST("/otp/verify", general.VerifyOTPHandler)
			public.GET("/account/delete", general_rest.GetDeletePage)
			public.POST("/approach", general.CreateApprochHandler)
			public.GET("/announcements", general_rest.GetAnnouncementsHandler)

			// TODO: this is a temp open api it will be moved to admin later
			public.POST("/sms/partner", general_rest.SaveSMSFCMHandler)

			adminPublic := public.Group("/admin")
			{
				adminPublic.POST("/login", admin.LoginAdminHandler)
			}

			driverPublic := public.Group("/driver")
			{
				driverPublic.POST("/register", driver.RegisterDriverHandler)
				driverPublic.POST("/login", driver.LoginDriverHandler)
				driverPublic.GET("/password/forgot", driver.ForgotPasswordHandler)
				driverPublic.GET("/pin/forgot", driver.ForgotPinHandler)
				driverPublic.POST("/rate", driver.RateDriverHandler)
			}

			ridePublic := public.Group("/ride")
			{
				ridePublic.GET("/filtered", ride.GetFilteredRidesHandler)
				ridePublic.GET("", ride.GetRideHandler)
				ridePublic.GET("/request", passenger.GetRideRequestHandler)
			}

			passengerPublic := public.Group("/passenger")
			{
				passengerPublic.POST("/seat/book", passenger.BookSeatHandler)
				passengerPublic.POST("/ride/request", passenger.RideRequestHandler)
				passengerPublic.POST("/register", passenger.RegisterPassengerHandler)
				passengerPublic.POST("/login", passenger.LoginPassengerHandler)
				passengerPublic.GET("/password/forgot", passenger.ForgotPasswordHandler)
			}
		}

		protected := v1.Group("")
		{
			adminProtected := protected.Group("/admin")
			adminProtected.Use(middleware.Authentication(constants.ADMIN_TOKEN))
			{
				adminProtected.GET("/rides", admin.GetRidesHandler)
				adminProtected.GET("/vehicles", admin.GetVehiclesHandler)
				adminProtected.GET("/drivers", admin.GetDriverDetailsHandler)
				adminProtected.DELETE("/driver", admin.DeleteDriverHandler)
				adminProtected.POST("/broadcast", admin.AdminBroadcastHandler)
				adminProtected.GET("/approch", admin.GetApprochRequestsHandler)
				adminProtected.POST("/announcement", admin.AnnouncementHandler)

				// role and permission management, an admin can add a new role or
				// remap what a role may do without a code change
				adminProtected.POST("/role", admin.CreateRoleHandler)
				adminProtected.GET("/roles", admin.GetRolesHandler)
				adminProtected.PATCH("/role", admin.UpdateRoleHandler)
				adminProtected.DELETE("/role", admin.DeleteRoleHandler)
				adminProtected.POST("/permission", admin.CreatePermissionHandler)
				adminProtected.GET("/permissions", admin.GetPermissionsHandler)
				adminProtected.PUT("/role/permissions", admin.SetRolePermissionsHandler)
			}

			driverProtected := protected.Group("/driver")
			driverProtected.Use(middleware.Authentication(constants.DRIVER_TOKEN))
			{
				driverProtected.POST("/pin", driver.SetDriverPinHandler)
				driverProtected.GET("/rides", ride.DriverRideHandler)
				driverProtected.PATCH("/ride/update", ride.UpdateRideHandler)
				driverProtected.GET("/logout", driver.LogoutDriverHandler)
				driverProtected.GET("/info", driver.DriverProfileInfoHandler)
				driverProtected.PATCH("/status", driver.UpdateProfileStatusHandler)
				driverProtected.DELETE("/delete", driver.DeleteDriverProfileHandler)
				driverProtected.POST("/password/reset", driver.ChangePasswordHandler)
				driverProtected.POST("/pin/reset", driver.ChangePinHandler)
				driverProtected.POST("/seat/book", driver.BookSeatHandler)
				driverProtected.GET("/bookings", driver.GetBookSeatHandler)
				driverProtected.GET("/ride/requests", passenger.GetRideRequestsHandler)
				driverProtected.GET("/booking/reserve", driver.ReserveSeatHandler)

				/*
					// NOTE: this feature is not needed atm
						driverProtected.POST("/book/ride", ride.BookSeatByDriverHandler)
						driverProtected.GET("/seats/booked", ride.GetBookSeatsHandler)
						driverProtected.DELETE("/seat/booked", ride.UpdateBookSeatHandler)
				*/
			}

			vehicleProtected := protected.Group("/vehicle")
			vehicleProtected.Use(middleware.Authentication(constants.DRIVER_TOKEN))
			{
				vehicleProtected.GET("/", driver.GetVehicleHandler)
				vehicleProtected.PATCH("/update", driver.UpdateVehicleHandler)
				vehicleProtected.POST("/register", driver.RegisterVehicleHandler)
			}

			rideProtected := protected.Group("/ride")
			rideProtected.Use(middleware.Authentication(constants.DRIVER_TOKEN))
			{
				rideProtected.POST("/create", ride.CreateRideHandler)
				rideProtected.GET("/templates", ride.GetRideTemplatesHandler)
				rideProtected.DELETE("/template", ride.DeleteRideTemplatesHandler)
				rideProtected.DELETE("/series", ride.CancelRideSeriesHandler)
			}

			// the fleet a driver owns or belongs to
			groupProtected := protected.Group("/group")
			groupProtected.Use(middleware.Authentication(constants.DRIVER_TOKEN))
			{
				groupProtected.POST("", group.CreateGroupHandler)
				groupProtected.GET("/mine", group.GetMyGroupsHandler)
				groupProtected.GET("", group.GetGroupDetailsHandler)
				groupProtected.POST("/request/driver", group.RequestJoinGroupAsDriverHandler)
				groupProtected.GET("/requests", group.GetGroupRequestsHandler)
				groupProtected.PATCH("/requests", group.DecideGroupRequestsHandler)
				groupProtected.PATCH("/submanagers", group.SetGroupSubManagersHandler)
				groupProtected.DELETE("", group.DeleteGroupHandler)

				// the travel forms the manager reads while building a shift
				groupProtected.GET("/passengers/schedules", group.GetGroupPassengerSchedulesHandler)
			}

			// building and running the shifts of a fleet
			shiftProtected := protected.Group("/shift")
			shiftProtected.Use(middleware.Authentication(constants.DRIVER_TOKEN))
			{
				shiftProtected.POST("", shift.CreateShiftHandler)
				shiftProtected.PUT("/seats", shift.UpdateShiftSeatsHandler)
				shiftProtected.GET("/detail", shift.GetShiftHandler)
				shiftProtected.GET("", shift.GetGroupShiftsHandler)
				shiftProtected.GET("/mine", shift.GetMyShiftsHandler)
				shiftProtected.DELETE("", shift.CancelShiftHandler)
				shiftProtected.GET("/templates", shift.GetShiftTemplatesHandler)
				shiftProtected.DELETE("/template", shift.DeleteShiftTemplateHandler)
			}

			// everything a signed in passenger can reach
			passengerProtected := protected.Group("/passenger")
			passengerProtected.Use(middleware.Authentication(constants.PASSENGER_TOKEN))
			{
				passengerProtected.GET("/info", passenger.PassengerProfileInfoHandler)
				passengerProtected.GET("/logout", passenger.LogoutPassengerHandler)
				passengerProtected.POST("/password/reset", passenger.ChangePasswordHandler)
				passengerProtected.DELETE("/delete", passenger.DeletePassengerProfileHandler)
				passengerProtected.PUT("/schedule", passenger.SetPassengerScheduleHandler)
				passengerProtected.GET("/schedule", passenger.GetPassengerScheduleHandler)

				// a passenger asks to join a fleet and follows the shifts they are on
				passengerProtected.POST("/group/request", group.RequestJoinGroupAsPassengerHandler)
				passengerProtected.GET("/shifts", shift.GetMyShiftsHandler)
				passengerProtected.GET("/shift/detail", shift.GetShiftHandler)
				passengerProtected.GET("/notifications", general_rest.GetNotificationsHandler)
			}

			userProtected := protected.Group("/user")
			userProtected.Use(middleware.Authentication(constants.DRIVER_TOKEN))
			{
				userProtected.GET("/notifications", general_rest.GetNotificationsHandler)
			}

			protected.GET("/session", func(c *gin.Context) {
				c.Status(200)
			})
		}
	}

	worker.StartWorkers()

	go startSocketServer()

	port := configuration.ConfigurationData.RestPort
	logger.LogInfo("starting HTTPS server on port:"+port, constants.DEFAULT_SESSION)

	if strings.EqualFold(configuration.ConfigurationData.Envirnment, constants.Production) {
		logger.LogDebug("envirnment", constants.DEFAULT_SESSION, configuration.ConfigurationData.Envirnment)

		// TLS server config
		server := &http.Server{
			Addr:    ":" + port,
			Handler: router,
		}

		// Start HTTPS server
		err := server.ListenAndServeTLS(
			"/etc/letsencrypt/live/sathsawari.com/fullchain.pem",
			"/etc/letsencrypt/live/sathsawari.com/privkey.pem",
		)

		if err != nil {
			logger.LogError("server failed: "+err.Error(), constants.DEFAULT_SESSION)
		}
	} else {
		router.Run(":" + configuration.ConfigurationData.RestPort)
	}
}

func startSocketServer() {
	http.HandleFunc("/ws", socket.StatsSocketHandler)

	logger.LogInfo("Socket server on port"+configuration.ConfigurationData.SocketPort, constants.DEFAULT_SESSION)
	go http.ListenAndServe(fmt.Sprintf(":%s", configuration.ConfigurationData.SocketPort), nil)
}

func setup() {
	monitoring.NewStatsStore()
	httpcall.NewClient()
}
