package worker

func StartWorkers() {
	// it closes rides whose endtime has reached
	go CloseActiveRidesScheduler()

	// it closes ride requests where endtime has reached
	go CloseActiveRideRequestsScheduler()

	// it writes shift trips ahead, retires finished ones and sends trip reminders
	go ShiftScheduler()

	// send notifications
	go ProcessNotifications()

	// start stats monitor
	go StartResourceMonitor()

	// broad cast notifications
	go ProcessBroadcastNotifications()
}
