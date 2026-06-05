package general

import (
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func getNotificationsResp(notifications []postgress.NotificationRequest, totalRows int64) utils.APIResponse {
	var userNotifications []NotificationRequest
	for _, notification := range notifications {
		userNotifications = append(userNotifications, NotificationRequest{
			ID:        notification.ID,
			UserId:    notification.UserId,
			UserType:  notification.UserType,
			Title:     notification.Title,
			Message:   notification.Message,
			CreatedAt: notification.CreatedAt,
		})
	}

	userNotificationResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: &GetNotificationsResponse{
			TotalPages:    utils.CalculatePagesize(totalRows),
			Notifications: userNotifications,
		},
	}

	return userNotificationResp
}

func createApprochResp(approchId string) utils.APIResponse {
	rideResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Request_Received_Successfully,
		Data:    approchId,
	}

	return rideResp
}
