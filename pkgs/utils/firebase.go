package utils

import (
	"context"
	"errors"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"

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
