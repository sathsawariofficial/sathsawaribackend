package general

import "time"

type GetNotificationsResponse struct {
	TotalPages    int                   `json:"totalPages"`
	Notifications []NotificationRequest `json:"notifications"`
}

type NotificationRequest struct {
	ID        string    `json:"id"`
	UserId    string    `json:"userId"`
	UserType  int       `json:"userType"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type SMSFCMRequest struct {
	FCM string `json:"fcm"`
	App string `json:"app"`
}
