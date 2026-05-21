// main.go
package main

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/middleware"
	"rideshare/services/admin"
	"rideshare/services/driver"
	general_rest "rideshare/services/general/rest"
	"rideshare/services/general/socket"
	"rideshare/services/ride"
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
	router.Use(middleware.Logger())

	router.GET("/", general_rest.GetHomePage)
	router.GET("/terms", general_rest.GetTermsAndConditions)
	router.GET("/privacy-policy", general_rest.GetPrivacyPolicy)

	v1 := router.Group("/api/v1")
	// open protection
	v1.Use(middleware.Authentication())
	{
		public := v1.Group("")
		{
			public.POST("/otp/resend", driver.ResendOTPHandler)
			public.POST("/otp/verify", driver.VerifyOTPHandler)
			public.GET("/account/delete", general_rest.GetDeletePage)

			// TODO: this is a temp open api it will be moved to admin later
			public.POST("/sms/partner", general_rest.SaveSMSFCMHandler)

			adminPublic := public.Group("/admin")
			{
				adminPublic.POST("/login", admin.LoginAdminHandler)
			}

			driverPublic := public.Group("/driver")
			{
				driverPublic.POST("/register", driver.RegisterDriverHandler)
				driverPublic.POST("/pin", driver.SetDriverPinHandler)
				driverPublic.POST("/login", driver.LoginDriverHandler)
				driverPublic.GET("/password/forgot", driver.ForgotPasswordHandler)
				driverPublic.POST("/rate", driver.RateDriverHandler)
			}

			ridePublic := public.Group("/ride")
			{
				ridePublic.GET("/filtered", ride.GetFilteredRidesHandler)
				ridePublic.POST("/approach", ride.CreateApprochHandler)
			}

			/*
				// NOTE: this feature is not needed atm
				passengerPublic := public.Group("/passenger")
				{
					passengerPublic.POST("/seat/book", passenger.BookSeatDemandHandler)
				}
			*/
		}

		protected := v1.Group("")
		protected.Use(middleware.Authentication())
		{
			protected.POST("/otp/send", driver.SendOTPHandler)

			adminProtected := protected.Group("/admin")
			{
				adminProtected.POST("/rides", admin.GetRidesHandler)
				adminProtected.POST("/veicles", admin.GetVehiclesHandler)
				adminProtected.POST("/driver", admin.GetDriverDetailsHandler)
			}

			driverProtected := protected.Group("/driver")
			{
				driverProtected.GET("/rides", ride.DriverRideHandler)
				driverProtected.PATCH("/ride/update", ride.UpdateRideHandler)
				driverProtected.GET("/logout", driver.LogoutDriverHandler)
				driverProtected.GET("/info", driver.DriverProfileInfoHandler)
				driverProtected.PATCH("/status", driver.UpdateProfileStatusHandler)
				driverProtected.DELETE("/delete", driver.DeleteDriverProfileHandler)
				driverProtected.POST("/password/reset", driver.ChangePasswordHandler)
				/*
					// NOTE: this feature is not needed atm
						driverProtected.POST("/book/ride", ride.BookSeatByDriverHandler)
						driverProtected.GET("/seats/booked", ride.GetBookSeatsHandler)
						driverProtected.DELETE("/seat/booked", ride.UpdateBookSeatHandler)
				*/
			}

			vehicleProtected := protected.Group("/vehicle")
			{
				vehicleProtected.GET("/", driver.GetVehicleHandler)
				vehicleProtected.PATCH("/update", driver.UpdateVehicleHandler)
				vehicleProtected.POST("/register", driver.RegisterVehicleHandler)
			}

			rideProtected := protected.Group("/ride")
			{
				rideProtected.POST("/create", ride.CreateRideHandler)
				rideProtected.GET("/templates", ride.GetRideTemplatesHandler)
				rideProtected.DELETE("/template", ride.DeleteRideTemplatesHandler)
			}

			userProtected := protected.Group("/user")
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
