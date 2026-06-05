package general

import (
	"fmt"
	"net/http"
	"os"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

func GetNotificationsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetNotificationsHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	err := utils.ValidateId(driverId)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)

	notifications, totalPages, err := GetNotifications(ctx, sessionId, driverId, page)
	if err != nil {
		logger.LogError(sessionId, "driver rides error: "+err.Error())
		if err.Error() == constants.Ride_Not_Found {
			ctx.JSON(http.StatusBadRequest, utils.APIResponse{
				Code:    http.StatusNoContent,
				Message: err.Error(),
				Data: GetNotificationsResponse{
					Notifications: []NotificationRequest{},
				},
			})
		} else {
			ctx.JSON(http.StatusBadRequest, utils.APIResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
				Data: GetNotificationsResponse{
					Notifications: []NotificationRequest{},
				},
			})
		}
		return
	}

	userNotificationResp := getNotificationsResp(notifications, totalPages)

	logger.LogInfo("Response returned from GetNotificationsHandler", sessionId)
	logger.LogDebug2("Response returned from GetNotificationsHandler", sessionId, userNotificationResp)

	ctx.JSON(http.StatusOK, userNotificationResp)
}

func GetHomePage(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetHomePage", sessionId)

	basePath := configuration.ConfigurationData.General.DocsPath
	path := basePath + constants.HOME_PAGE

	logger.LogDebug("Reading terms from path", sessionId, path)

	data, err := os.ReadFile(path)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	logger.LogInfo("Response returned from GetHomePage", sessionId)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func GetTermsAndConditions(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetTermsAndConditions", sessionId)

	basePath := configuration.ConfigurationData.General.DocsPath

	lang := ctx.Query(constants.Language_Key)
	logger.LogDebug("Terms language", sessionId, lang)

	subPath := constants.TERMS_ENGLISH
	if lang == constants.LANG_URDU {
		subPath = constants.TERMS_URDU
	}

	path := basePath + subPath

	logger.LogDebug("Reading terms from path", sessionId, path)

	data, err := os.ReadFile(path)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	logger.LogInfo("Response returned from GetTermsAndConditions", sessionId)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func GetPrivacyPolicy(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPrivacyPolicy", sessionId)

	basePath := configuration.ConfigurationData.General.DocsPath
	path := basePath + constants.PRIVACY_POLICY

	logger.LogDebug("Reading terms from path", sessionId, path)

	data, err := os.ReadFile(path)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	logger.LogInfo("Response returned from GetPrivacyPolicy", sessionId)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func GetDeletePage(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetDeletePage", sessionId)

	basePath := configuration.ConfigurationData.General.DocsPath
	path := basePath + constants.DELETE_PAGE

	logger.LogDebug("Reading delete account from path", sessionId, path)

	data, err := os.ReadFile(path)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	logger.LogInfo("Response returned from GetDeletePage", sessionId)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func SaveSMSFCMHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SaveSMSFCMHandler", sessionId)

	var request SMSFCMRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Registeration_Failed, "driver"),
		})
		return
	}

	logger.LogDebug2("Response received in SaveSMSFCMHandler", sessionId, request)

	err := ValidateSMSFCM(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = SaveSMSFCM(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	logger.LogInfo("Response returned from SaveSMSFCMHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(constants.Success))
}

// create approch
func CreateApprochHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in CreateApprochHandler", sessionId)

	var request ApprochRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Creation_Failed, "ride"),
		})
		return
	}

	logger.LogDebug2("Response received in CreateApprochHandler", sessionId, request)

	err := ValidateApproch(sessionId, &request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	approchId, err := SaveApprochDetails(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "ride creation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	approchResp := createApprochResp(approchId)

	logger.LogInfo("Response returned from CreateApprochHandler", sessionId)
	logger.LogDebug2("Response returned from CreateApprochHandler", sessionId, approchResp)

	ctx.JSON(http.StatusOK, approchResp)
}
