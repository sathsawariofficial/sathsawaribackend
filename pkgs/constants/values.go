package constants

import "regexp"

// regex
var (
	Lowercase_Regex   = regexp.MustCompile(`[a-z]`)
	Uppercase_Regex   = regexp.MustCompile(`[A-Z]`)
	Digit_Regex       = regexp.MustCompile(`\d`)
	Mobile_Regex      = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)
	SpecialChar_Regex = regexp.MustCompile(`[\W_]`)
	Email_Regex       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]{1,64}@[a-zA-Z0-9.-]{1,253}\.[a-zA-Z]{2,63}$`)
)

const (
	TERMS_ENGLISH = "terms_en.html"
	TERMS_URDU    = "terms_ur.html"
	DELETE_PAGE   = "delete_account.html"
)

const (
	LANG_ENGLISH = "en"
	LANG_URDU    = "ur"
)

const (
	NOTIFICATION_TYPE_RIDE_CREATED   = "ride_created"
	NOTIFICATION_TYPE_SMS_TO_SERVICE = "sms_to_service"
)

// notification titles
const (
	NOTIFICATION_TITLE_RIDE_CREATION  = "Ride Created"
	NOTIFICATION_TITLE_SMS_TO_SERVICE = "SathSawari sent an OTP"
)

// notification message
const (
	NOTIFICATION_MESSAGE_RIDE_CREATION  = "You have created a ride"
	NOTIFICATION_MESSAGE_SMS_TO_SERVICE = "SathSawari\nYour verifcation code is %s\nDon't share it with anyone."
)
