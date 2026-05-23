package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func IsStringEmptyWithKey(targetString string, keyName string, result *string) bool {
	isEmpty := IsStringEmpty(targetString)
	if isEmpty {
		keyNamesList := strings.Split(keyName, ".")
		listLen := len(keyNamesList)

		if listLen > 0 {
			*result = keyNamesList[listLen-1]
		} else {
			*result = keyName
		}
	}

	return isEmpty
}

/*
It checks if string is empty by checking its length
*/
func IsStringEmpty(val string) bool {
	var isValid bool

	if len(val) == 0 {
		isValid = true
	}

	return isValid
}

func GetCurrentTime() time.Time {
	return time.Now().UTC()
}

/*
it return the non empty value, by giving priority to option 1
*/
func ChoiseMaker(option1, option2 string) string {
	result := option1

	if IsStringEmpty(option1) {
		result = option2
	}

	return result
}

func IsValidPassword(password string) bool {
	passLen := len(password)

	// Check length
	if passLen < constants.Password_Min_Len || passLen > constants.Password_Max_Len {
		return false
	}

	// Check for at least one lowercase letter
	if !constants.Lowercase_Regex.MatchString(password) {
		return false
	}

	// Check for at least one uppercase letter
	if !constants.Uppercase_Regex.MatchString(password) {
		return false
	}

	// Check for at least one digit
	if !constants.Digit_Regex.MatchString(password) {
		return false
	}

	// Check for at least one special character
	if !constants.SpecialChar_Regex.MatchString(password) {
		return false
	}

	// If all checks passed, return true
	return true
}

func PKValidation(pk string) bool {
	if IsStringEmpty(pk) || len(pk) > constants.Postgress_PK_Size {
		return false
	}

	return true
}

func CreateBase64Encoding(value string) string {
	// Encode to base64
	base64EncodedValue := base64.StdEncoding.EncodeToString([]byte(value))
	return base64EncodedValue
}

func IsValidEmail(email string) bool {
	if len(email) < 6 || len(email) > 254 {
		return false
	}
	return constants.Email_Regex.MatchString(email)
}

// GenerateOTP returns a 6-digit random number as a string
func GenerateOTP() string {
	otp := rand.Intn(900000) + 100000 // 100000-999999
	return fmt.Sprintf("%06d", otp)
}

func ConvertStrToTime(dateTime string) (time.Time, error) {
	return time.Parse(constants.DateTimeLayout, dateTime)
}

func HandleMobileNumberInQuery(mobile_number string) string {
	mobileNumber := strings.TrimSpace(mobile_number)
	if !strings.HasPrefix(mobileNumber, "+") {
		mobileNumber = "+" + mobileNumber
	}

	return mobileNumber
}

func InSlice(arr []int, target int) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

func GeneralSuccessResp(replyMessage string) APIResponse {
	return APIResponse{
		Code:    http.StatusOK,
		Message: ChoiseMaker(replyMessage, constants.Success),
	}
}

func GeneralSocketResp(sessionId string, code int, replyMessage string) []byte {
	response := APIResponse{
		Code:    code,
		Message: replyMessage,
	}

	bResp, err := json.Marshal(response)
	if err != nil {
		logger.LogError(sessionId, err)
	}

	return bResp
}

func SendNotification(orgCtx *gin.Context, sessionId, notificationType, driverId, title, message string, data map[string]string) {
	logger.LogDebug("Request received  in SendNotification", sessionId, fmt.Sprintf("notificationType: %s, driverId: %s, title: %s, message: %s", notificationType, driverId, title, message))

	var fcm string

	switch notificationType {
	case constants.NOTIFICATION_TYPE_RIDE_CREATED:
		driverFCM, err := database.GetDriverFCM(orgCtx, driverId)
		if err != nil {
			logger.LogError(sessionId, err)
			return
		}
		fcm = driverFCM.FCM
	case constants.NOTIFICATION_TYPE_SMS_TO_SERVICE:
		smsFCM, err := database.GetSMSFCM(orgCtx)
		if err != nil {
			logger.LogError(sessionId, err)
			return
		}
		fcm = smsFCM.FCM
	default:
		logger.LogWarning(sessionId, fmt.Sprintf("unhandled notification type: %s", notificationType))
		return
	}

	redis.SendNotification(database.DatabaseConn.RedisConn, redis.NotificationRequest{
		Token:            fcm,
		Title:            title,
		Message:          message,
		UserType:         constants.User_Driver,
		UserId:           driverId,
		NotificationType: notificationType,
		Data:             data,
	})
}

func GetPageNumber(ctx *gin.Context) (page int, err error) {
	pageStr := ctx.DefaultQuery(constants.Page_Key, "1")
	page, err = strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	return
}

func ValidateId(id string) error {
	if !PKValidation(id) {
		return fmt.Errorf(constants.Invalid_Data, "id")
	}

	return nil
}

func CalculatePagesize(totalRows int64) (pageSize int) {
	size := int(math.Ceil(float64(totalRows) / float64(configuration.ConfigurationData.PageSize))) // Output: 2

	if size == 0 {
		pageSize = 1
	} else {
		pageSize = size
	}

	return
}

func RandomId() string {
	return xid.New().String()
}

func GetSocketUpgrader() websocket.Upgrader {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all connections by default (not recommended for production)
			return true
		},
	}

	Upgrade = upgrader
	return upgrader
}

func VerifyOTPOperations(operation string) (isValid bool) {
	switch operation {
	case constants.ACTIVATE_DRIVER_OPERATION,
		constants.ACTIVATE_VEHICLE_OPERATION,
		constants.UPDATE_PASSWORD_OPERATION,
		constants.FORGOT_PASSWORD_OPERATION:
		return true
	default:
		return false
	}
}
