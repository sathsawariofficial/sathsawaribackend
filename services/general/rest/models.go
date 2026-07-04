package general

import "time"

type GetNotificationsResponse struct {
	TotalPages    int                   `json:"totalPages"`
	Notifications []NotificationRequest `json:"notifications"`
}

type NotificationRequest struct {
	ID        string         `json:"id"`
	UserId    string         `json:"userId"`
	UserType  int            `json:"userType"`
	Title     string         `json:"title"`
	Message   string         `json:"message"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"createdAt"`
}

type SMSFCMRequest struct {
	FCM     string `json:"fcm"`
	App     string `json:"app"`
	AppHash string `json:"appHash"`
}

type ApprochRequest struct {
	Name    string `json:"name"`
	Number  string `json:"number"`
	Email   string `json:"email"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type VerifyOTPRequest struct {
	MobileNumber string `json:"mobile"`
	OTP          string `json:"otp"`
	Password     string `json:"password"`
	Pin          string `json:"pin"`
	Operation    string `json:"operation"`
}

type ForgotPasswordResponse struct {
	OTP string `json:"tempOTP"`
}
