package monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"rideshare/pkgs/configuration"
)

func CheckAlerts(
	stats ResourceStats,
) {
	cfg := configuration.ConfigurationData.Alert

	if stats.CPUPercent >= cfg.MaxCPUPercent {

		SendDiscordAlert(
			cfg.DiscordWebhookURL,
			fmt.Sprintf(
				"🚨 High CPU Usage %.2f%%",
				stats.CPUPercent,
			),
		)
	}

	if float64(stats.MemPercent) >= cfg.MaxMemPercent {

		SendDiscordAlert(
			cfg.DiscordWebhookURL,
			fmt.Sprintf(
				"🚨 High Process Memory Usage %.2f%%",
				stats.MemPercent,
			),
		)
	}

	if stats.FDPercent >= cfg.MaxFDPercent {

		SendDiscordAlert(
			cfg.DiscordWebhookURL,
			fmt.Sprintf(
				"🚨 High FD Usage %.2f%%",
				stats.FDPercent,
			),
		)
	}
}

func SendDiscordAlert(
	webhookURL string,
	message string,
) error {

	payload := map[string]string{
		"content": message,
	}

	body, _ := json.Marshal(payload)

	_, err := http.Post(
		webhookURL,
		"application/json",
		bytes.NewBuffer(body),
	)

	return err
}
