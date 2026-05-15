package worker

import (
	"rideshare/pkgs/configuration"
	"time"
)

func StartWorkers() {
	// it closes rides whose endtime has reached
	go func() {
		for {
			CloseActiveRides()
			<-time.After(time.Duration(configuration.ConfigurationData.RideCloseScheduler) * time.Second)
		}
	}()

	// send notifications
	go ProcessNotifications()
}
