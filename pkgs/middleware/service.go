package middleware

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/monitoring"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		defer func() {
			if err := recover(); err != nil {

				stack := string(debug.Stack())

				message := fmt.Sprintf(
					"[PANIC] method=%s path=%s err=%v\n%s",
					c.Request.Method,
					c.Request.URL.Path,
					err,
					stack,
				)
				logger.LogError(Middleware_Session, message)

				monitoring.SendDiscordAlert(configuration.ConfigurationData.Alert.DiscordWebhookURL, message)

				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					gin.H{
						"message": "internal server error",
					},
				)
			}
		}()

		c.Next()
	}
}

func StatsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()

		c.Next()

		duration := time.Since(start)

		statusCode := c.Writer.Status()

		key := c.Request.Method + ":" + c.FullPath()

		monitoring.StatsStoreInstance.Mu.Lock()
		defer monitoring.StatsStoreInstance.Mu.Unlock()

		stat, exists := monitoring.StatsStoreInstance.Stats[key]

		if !exists {
			stat = &monitoring.EndpointStats{
				Method:          c.Request.Method,
				Path:            c.FullPath(),
				StatusCodes:     make(map[int]uint64),
				FailureMessages: make(map[string]uint64),
			}

			monitoring.StatsStoreInstance.Stats[key] = stat
		}

		stat.TotalRequests++
		stat.StatusCodes[statusCode]++
		stat.TotalLatency += duration

		if statusCode >= 200 && statusCode < 400 {
			stat.SuccessCount++
		} else {
			stat.FailureCount++

			if failureMsg, exists := c.Get("failure_message"); exists {
				if msg, ok := failureMsg.(string); ok {
					stat.FailureMessages[msg]++
				}
			}
		}
	}
}
