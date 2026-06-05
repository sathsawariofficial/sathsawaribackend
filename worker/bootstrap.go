package worker

func StartWorkers() {
	// it closes rides whose endtime has reached
	go CloseActiveRidesScheduler()

	// send notifications
	go ProcessNotifications()

	// start stats monitor
	go StartResourceMonitor()

	// broad cast notifications
	go ProcessBroadcastNotifications()
}
