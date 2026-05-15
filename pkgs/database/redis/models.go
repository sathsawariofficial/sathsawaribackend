package redis

type NotificationRequest struct {
	NotificationType string
	Token            string
	Title            string
	Message          string
	UserType         int
	UserId           string
	Data             map[string]string
}
