package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/database/redis"
	httpcall "rideshare/pkgs/externalCall/http"
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

func IsValidMobileNumber(mobile string) error {
	mobileLen := len(mobile)

	if !(mobileLen >= constants.MobileNumber_Min_Len && mobileLen <= constants.MobileNumber_Max_Len) {
		return fmt.Errorf("length of the mobile number should be between %v and %v characters", constants.MobileNumber_Min_Len, constants.MobileNumber_Max_Len)
	}
	if !constants.Mobile_Regex.MatchString(mobile) {
		return fmt.Errorf("invalid mobile number %v", mobile)
	}

	return nil
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

func CleanMobileNumber(mobile_number string) string {
	mobileNumber := strings.TrimPrefix(mobile_number, "+")

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
	case constants.NOTIFICATION_TYPE_RIDE_CREATED,
		constants.NOTIFICATION_TYPE_PIN_CREATED,
		constants.NOTIFICATION_TYPE_INFORMATION,
		constants.NOTIFICATION_TYPE_MARKETING:
		driverFCM, err := database.GetDriverFCM(orgCtx, driverId)
		if err != nil {
			logger.LogError(sessionId, err)
			return
		}
		fcm = driverFCM.FCM
	case constants.NOTIFICATION_TYPE_SMS_TO_SERVICE,
		constants.NOTIFICATION_TYPE_BACKUP_SMS_TO_SERVICE:
		if configuration.ConfigurationData.Integerations.SMS.LocalSMSService ||
			notificationType == constants.NOTIFICATION_TYPE_BACKUP_SMS_TO_SERVICE {
			smsFCM, err := database.GetSMSFCM(orgCtx)
			if err != nil {
				logger.LogError(sessionId, err)
				return
			}
			fcm = smsFCM.FCM
		} else {
			mobileNumber, ok := data[constants.SMS_KEY_MOBILE_NUMBER]
			if !ok {
				err := errors.New("mobile number not found for sms")
				logger.LogError(sessionId, err)
				return
			}
			go SendOTPSMS(orgCtx, sessionId, mobileNumber, message)
			return
		}
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

func SendOTPSMS(orgCtx *gin.Context, sessionId, receiverMobile, message string) {
	logger.LogInfo("Request received in SendOTPSMS", sessionId)
	logger.LogDebug("Request received in SendOTPSMS", sessionId, fmt.Sprintf("receiver mobile: %s, message: %s", receiverMobile, message))

	var err error

	defer func() {
		if err != nil {
			message := fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SMS_TO_SERVICE, message, constants.DEFAULT_APP_HASH)

			// messaging partner
			SendNotification(orgCtx, sessionId, constants.NOTIFICATION_TYPE_BACKUP_SMS_TO_SERVICE, "", constants.NOTIFICATION_TYPE_SMS_TO_SERVICE, message, map[string]string{
				constants.SMS_KEY_MOBILE_NUMBER: receiverMobile,
				constants.SMS_KEY_MESSAGE:       message,
			})

		}
	}()

	request := map[string]string{
		"key":      configuration.ConfigurationData.Integerations.SMS.APIKey,
		"receiver": CleanMobileNumber(receiverMobile),
		"sender":   configuration.ConfigurationData.Integerations.SMS.Mask,
		"otpcode":  message,
	}

	logger.LogDebug("SendOTPSMS", sessionId, request)

	var url string
	url, err = httpcall.AddQueryParams(configuration.ConfigurationData.Integerations.SMS.URL, request)
	if err != nil {
		logger.LogError(sessionId, err)
		return
	}

	resp, err := httpcall.MakeRequest(orgCtx, httpcall.RequestOptions{
		Method:    http.MethodGet,
		Path:      url,
		Body:      nil,
		Headers:   map[string]string{},
		SessionID: sessionId,
	})
	if err != nil {
		logger.LogError(sessionId, err)
		return
	}
	logger.LogDebug("response", sessionId, string(resp))

	var response SMSResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		logger.LogError(sessionId, err)
		return
	}

	logger.LogInfo("Response returned from SendOTPSMS", sessionId)
	logger.LogDebug("Response returned from SendOTPSMS", sessionId, response)
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
		constants.FORGOT_PASSWORD_OPERATION,
		constants.FORGOT_PIN_OPERATION,
		constants.UPDATE_PIN_OPERATION,
		constants.BOOK_RIDE_OPERATION:
		return true
	default:
		return false
	}
}

func GetOTP(ctx *gin.Context, sessionId, mobileNumber string) (otp string, err error) {
	logger.LogInfo("Request received in GetOTP", sessionId)

	otp, err = redis.GetRedisValue(database.DatabaseConn.RedisConn, mobileNumber)
	if err != nil {
		logger.LogError(sessionId, "session deleted error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	logger.LogInfo("Response returned from GetOTP", sessionId)
	logger.LogDebug2("Response returned from GetOTP", sessionId, otp)

	return
}

func SendOTP(ctx *gin.Context, sessionId, mobileNumber, operation string) (otp string, err error) {
	logger.LogInfo("Request received in SendOTP", sessionId)

	otp = GenerateOTP()

	message := otp
	if configuration.ConfigurationData.Integerations.SMS.LocalSMSService {
		message = fmt.Sprintf(constants.NOTIFICATION_MESSAGE_SMS_TO_SERVICE, otp, constants.DEFAULT_APP_HASH)
	}

	// messaging partner
	SendNotification(ctx, sessionId, constants.NOTIFICATION_TYPE_SMS_TO_SERVICE, "", constants.NOTIFICATION_TYPE_SMS_TO_SERVICE, message, map[string]string{
		constants.SMS_KEY_MOBILE_NUMBER: mobileNumber,
		constants.SMS_KEY_MESSAGE:       message,
	})

	key := fmt.Sprintf("%s:%s", mobileNumber, operation)
	err = redis.SetRedisValueTTL(database.DatabaseConn.RedisConn, key, otp, time.Duration(configuration.ConfigurationData.Database.Redis.TTL)*time.Second)
	if err != nil {
		logger.LogError(sessionId, "session deleted error: "+err.Error())
		err = fmt.Errorf(constants.Unable_To_Do_Job, constants.Perform_this_operation)
		return
	}

	logger.LogInfo("Response returned from SendOTP", sessionId)
	logger.LogDebug2("Response returned from SendOTP", sessionId, otp)

	return
}

func GetDriver(orgCtx *gin.Context, mobile string) (driver postgress.Driver, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where(`driver_mobile = ?`, mobile).Find(&driver).Error

	return
}
