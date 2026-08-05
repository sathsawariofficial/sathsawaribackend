package general

import (
	"encoding/json"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
)

func getNotificationsResp(notifications []postgress.NotificationRequest, totalRows int64) utils.APIResponse {
	var userNotifications []NotificationRequest
	for _, notification := range notifications {
		data := NotificationRequest{
			ID:        notification.ID,
			UserId:    notification.UserId,
			UserType:  notification.UserType,
			Title:     notification.Title,
			Message:   notification.Message,
			CreatedAt: notification.CreatedAt,
		}
		err := json.Unmarshal([]byte(notification.Data), &data.Data)
		if err != nil {
			logger.LogError(notification.ID, err)
		}
		userNotifications = append(userNotifications, data)
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

func getAnnouncementsResp(announcements []postgress.AnnouncementRequests, totalRows int64) utils.APIResponse {
	var userAnnouncements []AnnouncementRequests
	for _, annoucement := range announcements {
		data := AnnouncementRequests{
			ID:        annoucement.ID,
			Title:     annoucement.Title,
			Message:   annoucement.Message,
			CreatedAt: annoucement.CreatedAt,
		}

		userAnnouncements = append(userAnnouncements, data)
	}

	userNotificationResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: &GetAnnouncementsResponse{
			TotalPages:    utils.CalculatePagesize(totalRows),
			Announcements: userAnnouncements,
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

func sendOTPResp(otp string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: ForgotPasswordResponse{
			OTP: otp,
		},
	}
}
