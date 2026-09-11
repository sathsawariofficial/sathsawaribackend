package constants

const (
	DEFAULT_SESSION = "default_session"
	WROKER_SESSION  = "worker_session"
)

// Socket Request Type
const (
	SR_PLACE_SEARCH = "PLACE_SEARCH"
)

const (
	DEFAULT_SHORT_CODE_LEN = 8
	DEFAULT_APP_HASH       = "aqfUly2KTch"
	URL_SHORTNER_SEED      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	APP_BASE_URL           = "https://sathsawari.com/"
	RIDE_BASE_URL          = APP_BASE_URL + "r/"
	RIDE_REQUEST_BASE_URL  = APP_BASE_URL + "rq/"
)

// place alerts
const (
	// settings read per page while looking up who follows a place
	DEFAULT_PLACE_ALERT_PAGE_SIZE = 500
	// pushes sent to firebase at once while a place alert fans out
	DEFAULT_PLACE_ALERT_SENDERS = 8
)
