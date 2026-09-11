package utils

import (
	"context"
	"errors"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
)

var client *messaging.Client

func init() {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		logger.LogError(constants.DEFAULT_SESSION, err)
	}

	client, err = app.Messaging(context.Background())
	if err != nil {
		logger.LogError(constants.DEFAULT_SESSION, err)
	}
}

func SendPush(token, title, body, notificationType string, data map[string]string) error {
	msg := &messaging.Message{
		Token: token,
		Data:  data, // message directed to app
	}

	if notificationType != constants.NOTIFICATION_TYPE_SMS_TO_SERVICE {
		msg.Notification = &messaging.Notification{
			Title: title,
			Body:  body,
		}
	}

	if client == nil {
		err := errors.New("firebase client does not exist")
		logger.LogWarning(constants.WROKER_SESSION, err)
		return err
	}

	_, err := client.Send(context.Background(), msg)
	if err != nil {
		logger.LogError(constants.WROKER_SESSION, err)
		return err
	}

	return nil
}

// SendLinkPush sends a notification that opens a link, the openUrl in its data. iOS draws
// it itself, with the buttons of the OPEN_URL category the app registers. Android gets
// the data alone, at high priority, because a notification the system draws for an app
// in the background cannot carry a button: the app draws it from the title, body and
// openUrl in the data and puts its open button under it.
func SendLinkPush(token, title, body string, data map[string]string) error {
	if client == nil {
		return errors.New("firebase client does not exist")
	}

	msg := &messaging.Message{
		Token: token,
		Data:  data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert:    &messaging.ApsAlert{Title: title, Body: body},
					Sound:    "default",
					Category: constants.NOTIFICATION_ACTION_OPEN_URL,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	_, err := client.Send(ctx, msg)
	return err
}
