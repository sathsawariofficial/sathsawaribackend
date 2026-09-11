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
	"rideshare/services/passenger"
	"rideshare/services/pickdrop"
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

				// a passenger has no account here, the places they follow belong to their device
				passengerPublic.GET("/notification/settings", passenger.GetNotificationSettingsHandler)
				passengerPublic.PUT("/notification/settings", passenger.SaveNotificationSettingsHandler)
			}

			// Pick & Drop advertisements, searchable by anybody looking for a service
			pickDropPublic := public.Group("/pickdrop")
			{
				pickDropPublic.GET("/advertisements/search", pickdrop.SearchAdvertisementsHandler)
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

				// the whole platform at a glance
				adminProtected.GET("/overview", admin.GetOverviewHandler)

				// moderation: switch an account off rather than destroying it
				adminProtected.PATCH("/driver/status", admin.UpdateDriverStatusHandler)

				// passenger accounts
				adminProtected.GET("/passengers", admin.GetPassengersHandler)
				adminProtected.GET("/passenger", admin.GetPassengerProfileHandler)
				adminProtected.PATCH("/passenger/status", admin.UpdatePassengerStatusHandler)
				adminProtected.DELETE("/passenger", admin.DeletePassengerHandler)

				// Pick & Drop is still run by each service's own owner, the admin
				// console only ever looks
				adminProtected.GET("/pickdrop/services", admin.GetPickDropServicesHandler)
				adminProtected.GET("/pickdrop/service", admin.GetPickDropServiceDetailHandler)
				adminProtected.GET("/pickdrop/advertisements", admin.GetAdvertisementsHandler)
				adminProtected.GET("/shifts", admin.GetShiftsHandler)
				adminProtected.GET("/shift/requests", admin.GetShiftRequestsHandler)
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

				// the places a driver follows to hear about ride requests
				driverProtected.GET("/notification/settings", driver.GetNotificationSettingsHandler)
				driverProtected.PUT("/notification/settings", driver.SaveNotificationSettingsHandler)

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

			// the Pick & Drop service a driver owns, or joins as a driver
			pickDropProtected := protected.Group("/pickdrop")
			pickDropProtected.Use(middleware.Authentication(constants.DRIVER_TOKEN))
			{
				pickDropProtected.POST("", pickdrop.EnableServiceHandler)
				pickDropProtected.GET("", pickdrop.GetMyServiceHandler)
				pickDropProtected.DELETE("", pickdrop.DisableServiceHandler)
				pickDropProtected.GET("/search", pickdrop.SearchServicesHandler)

				// the owner's own vehicles, and the vehicles a joining driver offers
				pickDropProtected.POST("/vehicles", pickdrop.AddServiceVehiclesHandler)
				pickDropProtected.DELETE("/vehicle", pickdrop.RemoveServiceVehicleHandler)
				pickDropProtected.POST("/vehicles/offer", pickdrop.OfferVehiclesHandler)
				pickDropProtected.DELETE("/vehicle/offer", pickdrop.WithdrawVehicleHandler)

				// joining and leaving, and the owner deciding who gets in
				pickDropProtected.POST("/request", pickdrop.RequestJoinAsDriverHandler)
				pickDropProtected.DELETE("/leave", pickdrop.LeaveServiceAsDriverHandler)
				pickDropProtected.GET("/requests", pickdrop.GetJoinRequestsHandler)
				pickDropProtected.GET("/request", pickdrop.GetJoinRequestHandler)
				pickDropProtected.PATCH("/requests", pickdrop.DecideJoinRequestsHandler)

				// what the owner builds shifts from
				pickDropProtected.GET("/drivers/available", pickdrop.GetAvailableDriversHandler)
				pickDropProtected.GET("/vehicles/available", pickdrop.GetAvailableVehiclesHandler)
				pickDropProtected.GET("/passengers/available", pickdrop.GetAvailablePassengersHandler)

				// the owner's advertisements
				pickDropProtected.POST("/advertisement", pickdrop.CreateAdvertisementHandler)
				pickDropProtected.GET("/advertisements", pickdrop.GetMyAdvertisementsHandler)
				pickDropProtected.DELETE("/advertisement", pickdrop.DeleteAdvertisementHandler)
			}

			// the recurring shifts of a service, built by its owner and driven by
			// the drivers assigned to them
			shiftProtected := protected.Group("/shift")
			shiftProtected.Use(middleware.Authentication(constants.DRIVER_TOKEN))
			{
				shiftProtected.POST("", shift.CreateShiftHandler)
				shiftProtected.PATCH("", shift.UpdateShiftHandler)
				shiftProtected.DELETE("", shift.DeleteShiftHandler)
				shiftProtected.PUT("/passengers", shift.UpdateShiftPassengersHandler)
				shiftProtected.GET("", shift.GetServiceShiftsHandler)
				shiftProtected.GET("/detail", shift.GetShiftHandler)
				shiftProtected.GET("/mine", shift.GetDriverShiftsHandler)
				shiftProtected.GET("/history", shift.GetShiftHistoryHandler)
				shiftProtected.GET("/passengers/history", shift.GetPassengerHistoryHandler)
				shiftProtected.GET("/occurrences", shift.GetOccurrencesHandler)
				shiftProtected.GET("/attendance", shift.GetAttendanceHandler)
				shiftProtected.POST("/location", shift.SendDriverLocationHandler)
				shiftProtected.GET("/requests", shift.SearchShiftRequestsHandler)
			}

			// everything a signed in passenger can reach
			passengerProtected := protected.Group("/passenger")
			passengerProtected.Use(middleware.Authentication(constants.PASSENGER_TOKEN))
			{
				passengerProtected.GET("/info", passenger.PassengerProfileInfoHandler)
				passengerProtected.GET("/logout", passenger.LogoutPassengerHandler)
				passengerProtected.POST("/password/reset", passenger.ChangePasswordHandler)
				passengerProtected.DELETE("/delete", passenger.DeletePassengerProfileHandler)
				// a passenger joins one Pick & Drop service and tells it when they travel
				passengerProtected.GET("/pickdrop/search", pickdrop.SearchServicesHandler)
				passengerProtected.GET("/pickdrop", pickdrop.GetPassengerMembershipHandler)
				passengerProtected.POST("/pickdrop/request", pickdrop.RequestJoinAsPassengerHandler)
				passengerProtected.DELETE("/pickdrop/leave", pickdrop.LeaveServiceAsPassengerHandler)
				passengerProtected.PUT("/availability", passenger.SetAvailabilityHandler)
				passengerProtected.GET("/availability", passenger.GetAvailabilityHandler)

				// the shifts they are on, their attendance and their travel history
				passengerProtected.GET("/shifts", shift.GetPassengerShiftsHandler)
				passengerProtected.GET("/shift/detail", shift.GetPassengerShiftHandler)
				passengerProtected.GET("/shift/attendance", shift.GetPassengerAttendanceHandler)
				passengerProtected.PATCH("/shift/attendance", shift.MarkAttendanceHandler)
				passengerProtected.GET("/shift/history", shift.GetTravelHistoryHandler)

				// the requirements they put out for owners to find
				passengerProtected.POST("/shift/request", shift.CreateShiftRequestHandler)
				passengerProtected.GET("/shift/requests", shift.GetMyShiftRequestsHandler)
				passengerProtected.DELETE("/shift/request", shift.DeleteShiftRequestHandler)
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
