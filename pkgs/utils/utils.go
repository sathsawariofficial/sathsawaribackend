package utils

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
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
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)

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
	SendUserNotification(orgCtx, sessionId, notificationType, driverId, constants.User_Driver, title, message, data)
}

// SendUserNotification is SendNotification with the type of the receiving user
// spelled out, a shift notification has to reach passengers as well as drivers.
func SendUserNotification(orgCtx *gin.Context, sessionId, notificationType, userId string, userType int, title, message string, data map[string]string) {
	logger.LogDebug("Request received  in SendNotification", sessionId, fmt.Sprintf("notificationType: %s, userId: %s, title: %s, message: %s", notificationType, userId, title, message))

	var fcm string

	lookupUserFCM := func() bool {
		userFCM, err := database.GetDriverFCM(orgCtx, userId)
		if err != nil {
			logger.LogError(sessionId, err)
			return false
		}
		fcm = userFCM.FCM
		return true
	}

	switch notificationType {
	case constants.NOTIFICATION_TYPE_RIDE_CREATED,
		constants.NOTIFICATION_TYPE_PIN_CREATED,
		constants.NOTIFICATION_TYPE_INFORMATION,
		constants.NOTIFICATION_TYPE_MARKETING,
		constants.NOTIFICATION_TITLE_RIDE_BOOKED:
		if !lookupUserFCM() {
			return
		}
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
		if !constants.IsPickDropNotificationType(notificationType) {
			logger.LogWarning(sessionId, fmt.Sprintf("unhandled notification type: %s", notificationType))
			return
		}

		if !lookupUserFCM() {
			return
		}
	}

	redis.SendNotification(database.DatabaseConn.RedisConn, redis.NotificationRequest{
		Token:            fcm,
		Title:            title,
		Message:          message,
		UserType:         userType,
		UserId:           userId,
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
		constants.ACTIVATE_PASSENGER_OPERATION,
		constants.PASSENGER_UPDATE_PASSWORD_OPERATION,
		constants.PASSENGER_FORGOT_PASSWORD_OPERATION,
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

func GetPassenger(orgCtx *gin.Context, mobile string) (passenger postgress.Passenger, err error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(orgCtx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	err = database.DatabaseConn.Postgres.WithContext(ctx).Where(`passenger_mobile = ?`, mobile).Find(&passenger).Error

	return
}

func GenerateShortCode(id string) string {
	hash := sha256.Sum256([]byte(id))

	// First 8 bytes -> uint64
	n := binary.BigEndian.Uint64(hash[:constants.DEFAULT_SHORT_CODE_LEN])

	return encodeBase62(n)
}

func encodeBase62(n uint64) string {
	if n == 0 {
		return "0"
	}

	var out []byte

	for n > 0 {
		out = append(out, constants.URL_SHORTNER_SEED[n%62])
		n /= 62
	}

	// reverse
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}

	return string(out)
}

func CreateOpenRideLink(urlType, shortCode string) string {
	switch urlType {
	case constants.LIKE_TYPE_RIDE_REQUEST_URL:
		return constants.RIDE_REQUEST_BASE_URL + shortCode
	case constants.LIKE_TYPE_RIDE_URL:
		return constants.RIDE_BASE_URL + shortCode
	default:
		return constants.RIDE_BASE_URL + shortCode
	}
}

func IsUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// NormalizeClock accepts a wall clock time as HH:MM or HH:MM:SS and hands it back as
// HH:MM, the form every schedule is stored and compared in.
func NormalizeClock(value string) (string, error) {
	value = strings.TrimSpace(value)

	for _, layout := range []string{constants.Clock_Layout, "15:04:05"} {
		if clock, err := time.Parse(layout, value); err == nil {
			return clock.Format(constants.Clock_Layout), nil
		}
	}

	return "", fmt.Errorf(constants.Invalid_Data, "time, use HH:MM")
}

// ParseBusinessDate reads a YYYY-MM-DD date on the business wall clock.
func ParseBusinessDate(value, key string) (time.Time, error) {
	day, err := time.ParseInLocation(constants.Date_Layout, strings.TrimSpace(value), constants.Business_Location)
	if err != nil {
		return day, fmt.Errorf(constants.Invalid_Data, key+", use YYYY-MM-DD")
	}

	return day, nil
}

// ValidateDaysOfWeek checks a set of weekdays, Monday 1 to Sunday 7, each at most once.
func ValidateDaysOfWeek(days []int) error {
	if len(days) == 0 {
		return fmt.Errorf(constants.Missing_Data, "Days of week")
	}

	seen := map[int]bool{}
	for _, day := range days {
		if day < constants.Day_Of_Week_Min_Value || day > constants.Day_Of_Week_Max_Value {
			return fmt.Errorf(constants.Invalid_Data, "day of week, use 1 for Monday to 7 for Sunday")
		}

		if seen[day] {
			return fmt.Errorf(constants.Invalid_Data, "days of week, a day is repeated")
		}
		seen[day] = true
	}

	return nil
}

// ParseDaysOfWeekQuery reads days of week sent as a comma separated query value.
func ParseDaysOfWeekQuery(value string) (days []int, err error) {
	if IsStringEmpty(strings.TrimSpace(value)) {
		return
	}

	for _, part := range strings.Split(value, ",") {
		day, e := strconv.Atoi(strings.TrimSpace(part))
		if e != nil {
			return nil, fmt.Errorf(constants.Invalid_Data, "days of week")
		}
		days = append(days, day)
	}

	return days, ValidateDaysOfWeek(days)
}

// ValidateRouteLocations checks an ordered route and hands it back cleaned: names
// trimmed and every time as HH:MM. Each place has to be reached after the one before
// it, a route is driven in order.
func ValidateRouteLocations(locations []RouteLocation, minLen, maxLen int) ([]RouteLocation, error) {
	if len(locations) < minLen {
		return nil, fmt.Errorf("at least %d location(s) are required", minLen)
	}

	if len(locations) > maxLen {
		return nil, fmt.Errorf("no more than %d locations can be sent", maxLen)
	}

	cleaned := make([]RouteLocation, 0, len(locations))
	previous := ""

	for index, location := range locations {
		name := strings.TrimSpace(location.Location)
		if IsStringEmpty(name) {
			return nil, fmt.Errorf(constants.Missing_Data, fmt.Sprintf("Location %d", index+1))
		}

		if len(name) > constants.Location_Max_Len {
			return nil, fmt.Errorf("length of location %d should not be more than %v characters", index+1, constants.Location_Max_Len)
		}

		if location.Lat < -90 || location.Lat > 90 || location.Lng < -180 || location.Lng > 180 {
			return nil, fmt.Errorf(constants.Invalid_Data, fmt.Sprintf("coordinates of location %d", index+1))
		}

		clock, err := NormalizeClock(location.Time)
		if err != nil {
			return nil, fmt.Errorf(constants.Invalid_Data, fmt.Sprintf("time of location %d, use HH:MM", index+1))
		}

		if previous != "" && clock <= previous {
			return nil, fmt.Errorf("location %d has to be reached after location %d", index+1, index)
		}
		previous = clock

		cleaned = append(cleaned, RouteLocation{
			Location: name,
			Lat:      location.Lat,
			Lng:      location.Lng,
			Time:     clock,
		})
	}

	return cleaned, nil
}

func Refuse(message string) error {
	return Refusal{Message: message}
}

// ClientError turns whatever came out of a transaction into what the caller may see:
// a refusal keeps its own message, anything else is logged and hidden behind the
// fallback.
func ClientError(sessionId string, err error, fallback string) error {
	logger.LogError(sessionId, err)

	var refusal Refusal
	if errors.As(err, &refusal) {
		return errors.New(refusal.Message)
	}

	return errors.New(fallback)
}

// RoutePoints lower cases every place of a route in order, the same normalisation
// ride route points get, so the same array search finds them.
func RoutePoints(locations []RouteLocation) []string {
	points := make([]string, 0, len(locations))
	for _, location := range locations {
		points = append(points, strings.ToLower(strings.TrimSpace(location.Location)))
	}

	return points
}

// NormalizePlace is the form a place name is compared in: trimmed, inner runs of spaces
// collapsed and lower cased, so "Saddar", " saddar " and "SADDAR" are the same place.
func NormalizePlace(place string) string {
	return strings.ToLower(strings.Join(strings.Fields(place), " "))
}

// NormalizePlaces puts the places a user follows in the form they are matched in, dropping
// blanks and repeats, and refuses a list that is too long or a name no location could have.
func NormalizePlaces(places []string) ([]string, error) {
	normalized := make([]string, 0, len(places))

	for _, place := range places {
		if place = NormalizePlace(place); place == "" || slices.Contains(normalized, place) {
			continue
		}
		if len(place) > constants.General_Max_Len {
			return nil, fmt.Errorf("length of a place should be between %v and %v characters", constants.General_Min_Len, constants.General_Max_Len)
		}

		normalized = append(normalized, place)
	}

	if len(normalized) > constants.Notification_Places_Max_Count {
		return nil, fmt.Errorf("there cannot be more than %d places", constants.Notification_Places_Max_Count)
	}

	return normalized, nil
}

// DisplayDateTime turns a stored ride time into the short form a notification shows, and
// hands the stored text back unchanged if it does not parse.
func DisplayDateTime(dateTime string) string {
	parsed, err := ConvertStrToTime(dateTime)
	if err != nil {
		return dateTime
	}

	return parsed.Format(constants.DisplayDateTimeLayout)
}

// NotifyPlaceSubscribers tells every user of a type who follows one of the places that a
// ride, or a ride request, taking it in has just been put out, with the link that opens
// it. Run it in its own goroutine once the ride or request is saved.
//
// Place alerts are never kept in the notifications table. They are pushed to firebase
// from here, a few at a time, rather than queued on the notification channel: that
// channel has a single subscriber that also relays OTPs, and a place followed by
// thousands would hold every OTP up behind it.
func NotifyPlaceSubscribers(sessionId string, userType int, notificationType string, places []string, title, message, openURL string, data map[string]string) {
	logger.LogInfo("Request received in NotifyPlaceSubscribers", sessionId)

	defer func() {
		if r := recover(); r != nil {
			logger.LogError(sessionId, fmt.Errorf("panic recovered in NotifyPlaceSubscribers: %v", r))
		}
	}()

	matched := make([]string, 0, len(places))
	for _, place := range places {
		if place = NormalizePlace(place); place != "" && !slices.Contains(matched, place) {
			matched = append(matched, place)
		}
	}
	if len(matched) == 0 {
		return
	}

	payload := make(map[string]string, len(data)+5)
	for key, value := range data {
		payload[key] = value
	}
	payload[constants.NOTIFICATION_KEY_TYPE] = notificationType
	payload[constants.NOTIFICATION_KEY_TITLE] = title
	payload[constants.NOTIFICATION_KEY_BODY] = message
	payload[constants.NOTIFICATION_KEY_OPEN_URL] = openURL
	payload[constants.NOTIFICATION_KEY_ACTION] = constants.NOTIFICATION_ACTION_OPEN_URL

	var (
		wg      sync.WaitGroup
		failed  atomic.Int64
		senders = make(chan struct{}, constants.DEFAULT_PLACE_ALERT_SENDERS)
		// a device reached through more than one setting is still told once
		notified = make(map[string]bool)
	)

	afterId := ""
	for {
		recipients, err := database.GetPlaceAlertRecipients(context.Background(), userType, matched, afterId, constants.DEFAULT_PLACE_ALERT_PAGE_SIZE)
		if err != nil {
			logger.LogError(sessionId, "failed to get place alert recipients error: "+err.Error())
			break
		}

		for _, recipient := range recipients {
			if notified[recipient.FCM] {
				continue
			}
			notified[recipient.FCM] = true

			wg.Add(1)
			senders <- struct{}{}
			go func(token string) {
				defer func() {
					if r := recover(); r != nil {
						failed.Add(1)
						logger.LogError(sessionId, fmt.Errorf("panic recovered sending a place alert: %v", r))
					}
					<-senders
					wg.Done()
				}()

				if err := SendLinkPush(token, title, message, payload); err != nil {
					failed.Add(1)
					logger.LogDebug("place alert not delivered", sessionId, err.Error())
				}
			}(recipient.FCM)
		}

		if len(recipients) < constants.DEFAULT_PLACE_ALERT_PAGE_SIZE {
			break
		}
		afterId = recipients[len(recipients)-1].SettingId
	}

	wg.Wait()

	logger.LogInfo(fmt.Sprintf("place alert %s sent to %d device(s), %d failed", notificationType, len(notified), failed.Load()), sessionId)
}
