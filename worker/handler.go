package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/monitoring"
	"rideshare/pkgs/utils"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

func CloseActiveRidesScheduler() {
	sessionId := constants.WROKER_SESSION
	logger.LogInfo("Starting CloseActiveRidesScheduler", sessionId)

	defer func() {
		if r := recover(); r != nil {
			logger.LogError(sessionId, fmt.Errorf("panic recovered in scheduler: %v", r))
		}
	}()

	ticker := time.NewTicker(
		time.Duration(configuration.ConfigurationData.General.Tickers.RideCloseScheduler) * time.Second,
	)
	defer ticker.Stop()

	// run once immediately on start
	closeActiveRides()

	for {
		select {
		case <-ticker.C:
			closeActiveRides()
		}
	}
}

func ProcessNotifications() {
	defer func() {
		if r := recover(); r != nil {
			logger.LogError(constants.WROKER_SESSION, fmt.Errorf("panic recovered: %v", r))
		}
	}()

	redis.ProcessNotifications(database.DatabaseConn.RedisConn, func(data redis.NotificationRequest) {
		var err error
		if data.NotificationType != constants.NOTIFICATION_TYPE_SMS_TO_SERVICE {
			_ = saveNotification(data)
		}
		err = utils.SendPush(data.Token, data.Title, data.Message, data.NotificationType, data.Data)
		if err != nil {
			logger.LogError(constants.WROKER_SESSION, err)
		}
		logger.LogError(constants.WROKER_SESSION, err)
	})
}

func StartResourceMonitor() {
	defer func() {
		if r := recover(); r != nil {
			logger.LogError(constants.WROKER_SESSION, fmt.Errorf("panic recovered: %v", r))
		}
	}()

	cfg := configuration.ConfigurationData.Alert

	pid := int32(os.Getpid())

	proc, err := process.NewProcess(pid)
	if err != nil {
		logger.LogError(constants.WROKER_SESSION, "failed to create process handle: "+err.Error())
		return
	}

	logFilePath := cfg.ResourceLogPath
	if logFilePath == "" {
		logger.LogError(constants.WROKER_SESSION, "resource log path is empty in config")
		return
	}

	apiStatsFilePath := cfg.APIStatsPath
	if logFilePath == "" {
		logger.LogError(constants.WROKER_SESSION, "api stats path is empty in config")
		return
	}

	// resource usage stats
	resourceFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.LogError(constants.WROKER_SESSION, "failed to open file: "+err.Error())
		return
	}
	defer resourceFile.Close()

	resourceEncoder := json.NewEncoder(resourceFile)

	// api stats
	apisStatsFile, err := os.OpenFile(apiStatsFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.LogError(constants.WROKER_SESSION, "failed to open file: "+err.Error())
		return
	}
	defer apisStatsFile.Close()

	apisStatsEncoder := json.NewEncoder(apisStatsFile)

	// convert minutes → duration
	checkInterval := time.Duration(cfg.CheckInterval) * time.Minute
	logWriteInterval := time.Duration(cfg.LogWriteInterval) * time.Minute

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	lastWrite := time.Now()

	for range ticker.C {

		cpuPercent, err := proc.CPUPercent()
		if err != nil {
			logger.LogError(constants.WROKER_SESSION, "cpu percent error: "+err.Error())
			continue
		}

		memPercent, err := proc.MemoryPercent()
		if err != nil {
			logger.LogError(constants.WROKER_SESSION, "memory percent error: "+err.Error())
			continue
		}

		memInfo, err := proc.MemoryInfo()
		if err != nil {
			logger.LogError(constants.WROKER_SESSION, "memory info error: "+err.Error())
			continue
		}

		memMB := float64(memInfo.RSS) / 1024 / 1024

		fdCount, err := proc.NumFDs()
		if err != nil {
			logger.LogError(constants.WROKER_SESSION, "fd count error: "+err.Error())
			continue
		}

		fdLimit := 10000.0
		fdPercent := (float64(fdCount) / fdLimit) * 100

		stats := monitoring.ResourceStats{
			CPUPercent: cpuPercent,
			MemPercent: memPercent,
			MemMB:      memMB,
			FDPercent:  fdPercent,
			FDCount:    int(fdCount),
		}

		monitoring.CheckAlerts(stats)

		entry := monitoring.ResourceLogEntry{
			Timestamp:  time.Now(),
			CPUPercent: cpuPercent,
			MemPercent: memPercent,
			MemMB:      memMB,
			FDCount:    int(fdCount),
			FDPercent:  fdPercent,
		}

		// interval-based writing
		if time.Since(lastWrite) >= logWriteInterval {
			err := resourceEncoder.Encode(entry)
			if err != nil {
				logger.LogError(constants.WROKER_SESSION, "failed to write resource log: "+err.Error())
			}

			err = apisStatsEncoder.Encode(monitoring.StatsStoreInstance.Stats)
			if err != nil {
				logger.LogError(constants.WROKER_SESSION, "failed to write api stats: "+err.Error())
			} else {
				lastWrite = time.Now()
			}
		}
	}
}

func ProcessBroadcastNotifications() {
	sessionId := constants.WROKER_SESSION
	logger.LogInfo("Starting ProcessBroadcastNotifications scheduler", sessionId)

	defer func() {
		if r := recover(); r != nil {
			logger.LogError(sessionId, fmt.Errorf("panic recovered: %v", r))
		}
	}()

	ticker := time.NewTicker(
		time.Duration(configuration.ConfigurationData.General.Tickers.BroadcastScheduler) * time.Second,
	)
	defer ticker.Stop()

	ginCtx := utils.GetWorkerGinContext()
	const batchSize = 50

	// run immediately on start
	processBroadcastBatch(ginCtx, sessionId, batchSize)

	for {
		select {
		case <-ticker.C:
			processBroadcastBatch(ginCtx, sessionId, batchSize)
		}
	}
}
