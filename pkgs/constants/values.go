package constants

import (
	"regexp"
	"time"
)

// regex
var (
	Lowercase_Regex   = regexp.MustCompile(`[a-z]`)
	Uppercase_Regex   = regexp.MustCompile(`[A-Z]`)
	Digit_Regex       = regexp.MustCompile(`\d`)
	Mobile_Regex      = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)
	SpecialChar_Regex = regexp.MustCompile(`[\W_]`)
	Email_Regex       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]{1,64}@[a-zA-Z0-9.-]{1,253}\.[a-zA-Z]{2,63}$`)
)

// Business_Location is the wall clock every pick & drop schedule is written in. The
// ride closer already reads ride times in PKT, and a schedule of "08:00 every
// Monday" only means something against one fixed zone, whatever zone the server
// itself happens to run in.
var Business_Location = time.FixedZone("PKT", 5*60*60)

// Business_Location_Name is the same zone as Postgres knows it, for the queries that
// turn a stored timestamp back into wall clock time.
const Business_Location_Name = "Asia/Karachi"

const (
	HOME_PAGE      = "home.html"
	TERMS_ENGLISH  = "terms_en.html"
	TERMS_URDU     = "terms_ur.html"
	DELETE_PAGE    = "delete_account.html"
	PRIVACY_POLICY = "privacy_policy.html"
)

const (
	LANG_ENGLISH = "en"
	LANG_URDU    = "ur"
)

const (
	NOTIFICATION_TYPE_RIDE_CREATED          = "ride_created"
	NOTIFICATION_TYPE_PIN_CREATED           = "pin_created"
	NOTIFICATION_TYPE_SMS_TO_SERVICE        = "sms_to_service"
	NOTIFICATION_TYPE_BACKUP_SMS_TO_SERVICE = "sms_to_backup_service"
	NOTIFICATION_TYPE_INFORMATION           = "information"
	NOTIFICATION_TYPE_MARKETING             = "marketing"

	NOTIFICATION_TYPE_PICKDROP_REQUEST    = "pickdrop_request"
	NOTIFICATION_TYPE_PICKDROP_DECISION   = "pickdrop_decision"
	NOTIFICATION_TYPE_PICKDROP_UPDATE     = "pickdrop_update"
	NOTIFICATION_TYPE_SHIFT_CREATED       = "shift_created"
	NOTIFICATION_TYPE_SHIFT_UPDATED       = "shift_updated"
	NOTIFICATION_TYPE_SHIFT_DELETED       = "shift_deleted"
	NOTIFICATION_TYPE_SHIFT_ASSIGNED      = "shift_assigned"
	NOTIFICATION_TYPE_SHIFT_PASSENGER     = "shift_passenger"
	NOTIFICATION_TYPE_SHIFT_ABSENCE       = "shift_absence"
	NOTIFICATION_TYPE_SHIFT_DRIVER_UPDATE = "shift_driver_update"
	NOTIFICATION_TYPE_SHIFT_REMINDER      = "shift_reminder"
)

const (
	ANNOUNCEMENT_TYPE_GENERAL = "general"
)

func NewAnnouncementType(msg string) bool {
	switch msg {
	case ANNOUNCEMENT_TYPE_GENERAL:
		return true
	default:
		return false
	}
}

// NewNotificationType is what an admin may broadcast. The pick & drop and shift
// types are left out on purpose, those notifications only make sense for the
// people inside one service and are never sent to everybody.
func NewNotificationType(notificationType string) bool {
	switch notificationType {
	case NOTIFICATION_TYPE_RIDE_CREATED,
		NOTIFICATION_TYPE_PIN_CREATED,
		NOTIFICATION_TYPE_SMS_TO_SERVICE,
		NOTIFICATION_TYPE_INFORMATION,
		NOTIFICATION_TYPE_MARKETING:
		return true
	default:
		return false
	}
}

// IsPickDropNotificationType reports whether a notification belongs to the pick &
// drop or shift features, which are delivered to a single user through their fcm.
func IsPickDropNotificationType(notificationType string) bool {
	switch notificationType {
	case NOTIFICATION_TYPE_PICKDROP_REQUEST,
		NOTIFICATION_TYPE_PICKDROP_DECISION,
		NOTIFICATION_TYPE_PICKDROP_UPDATE,
		NOTIFICATION_TYPE_SHIFT_CREATED,
		NOTIFICATION_TYPE_SHIFT_UPDATED,
		NOTIFICATION_TYPE_SHIFT_DELETED,
		NOTIFICATION_TYPE_SHIFT_ASSIGNED,
		NOTIFICATION_TYPE_SHIFT_PASSENGER,
		NOTIFICATION_TYPE_SHIFT_ABSENCE,
		NOTIFICATION_TYPE_SHIFT_DRIVER_UPDATE,
		NOTIFICATION_TYPE_SHIFT_REMINDER:
		return true
	default:
		return false
	}
}

// User Types
const (
	User_Driver    = 1
	User_Passenger = 2
)

func NewUserType(userType int) bool {
	switch userType {
	case User_Driver,
		User_Passenger:
		return true
	default:
		return false
	}
}

const (
	Approch_Complain = "complain"
	Approch_Cantact  = "contact"
)

func NewApprochType(approchType string) bool {
	switch approchType {
	case Approch_Complain,
		Approch_Cantact:
		return true
	default:
		return false
	}
}
