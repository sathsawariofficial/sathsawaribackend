package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/redis"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the Authorization header
		authHeader := ctx.GetHeader("Authorization")
		sessionId := Middleware_Session
		logger.LogInfo("Request received in Authentication [middleware]", sessionId)

		// Check if the header is present and starts with "Bearer "
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			err := fmt.Errorf("authorization header is missing or invalid")
			logger.LogError(sessionId, "login session id: "+err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.APIResponse{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			})
			return
		}

		// Extract the jwt token (which follows "Bearer ")
		token := strings.TrimPrefix(authHeader, "Bearer ")

		if token == "" {
			err := fmt.Errorf("token is missing in the authorization header")
			logger.LogError(sessionId, "login token not found: "+err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.APIResponse{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			})
			return
		}

		encryptedDriverId, err := utils.VerifyJWT(sessionId, token)
		if err != nil {
			err = errors.New(constants.Invalid_Session)
			logger.LogError(sessionId, "failed to encrypt driver id from token: "+err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.APIResponse{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			})
			return
		}

		if !utils.IsStringEmpty(encryptedDriverId) {
			_, err = redis.GetRedisValue(database.DatabaseConn.RedisConn, encryptedDriverId)
			if err != nil {
				logger.LogError(sessionId, "session not found error: "+err.Error())
				err = fmt.Errorf(constants.Unable_To_Do_Job, "verify user")
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.APIResponse{
					Code:    http.StatusUnauthorized,
					Message: err.Error(),
				})
				return
			}

			driverId, err := utils.DecryptAES(sessionId, encryptedDriverId)
			if err != nil {
				err = errors.New(constants.Invalid_Session)
				logger.LogError(sessionId, "failed to decrypt driver id: "+err.Error())
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.APIResponse{
					Code:    http.StatusUnauthorized,
					Message: err.Error(),
				})
				return
			}

			driverStatus, err := getDriverStatusById(driverId)
			if err != nil {
				logger.LogError(sessionId, "get driver error: "+err.Error())
				err = errors.New(constants.Driver_Not_Found)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.APIResponse{
					Code:    http.StatusUnauthorized,
					Message: err.Error(),
				})
				return
			}

			if driverStatus != constants.Status_Active {
				logger.LogError(sessionId, "invalid driver status error: "+driverStatus)
				err = errors.New(constants.Operation_Not_Permitted)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.APIResponse{
					Code:    http.StatusUnauthorized,
					Message: err.Error(),
				})
				return
			}

			logger.LogInfo("forwarding request to handler", sessionId)

			ctx.Set(constants.Sessoin_KEY, token)
			ctx.Set(constants.Driver_KEY, driverId)
			ctx.Set(constants.Encrypted_Driver_KEY, encryptedDriverId)
		}

		ctx.Next()

		logger.LogInfo("Response returned from Authentication [middleware]", sessionId)
	}
}

func SocketAuth(oldCtx *context.Context, sessionId, token string) []byte {
	logger.LogInfo("Request received in SocketAuth", sessionId)

	if token == "" {
		err := fmt.Errorf("token is missing in the authorization header")
		logger.LogError(sessionId, "login token not found: "+err.Error())
		return utils.GeneralSocketResp(sessionId, http.StatusUnauthorized, err.Error())
	}

	encryptedDriverId, err := utils.VerifyJWT(sessionId, token)
	if err != nil {
		err = errors.New(constants.Invalid_Session)
		logger.LogError(sessionId, "failed to encrypt driver id from token: "+err.Error())
		return utils.GeneralSocketResp(sessionId, http.StatusUnauthorized, constants.Operation_Not_Permitted)
	}

	// Driver token handling
	if utils.IsStringEmpty(encryptedDriverId) {
		_, err = redis.GetRedisValue(database.DatabaseConn.RedisConn, encryptedDriverId)
		if err != nil {
			logger.LogError(sessionId, "session not found error: "+err.Error())
			err = fmt.Errorf(constants.Unable_To_Do_Job, "verify user")
			return utils.GeneralSocketResp(sessionId, http.StatusUnauthorized, err.Error())
		}

		driverId, err := utils.DecryptAES(sessionId, encryptedDriverId)
		if err != nil {
			err = errors.New(constants.Invalid_Session)
			logger.LogError(sessionId, "failed to decrypt driver id: "+err.Error())
			return utils.GeneralSocketResp(sessionId, http.StatusUnauthorized, err.Error())
		}

		driverStatus, err := getDriverStatusById(driverId)
		if err != nil {
			logger.LogError(sessionId, "get driver error: "+err.Error())
			err = errors.New(constants.Driver_Not_Found)
			return utils.GeneralSocketResp(sessionId, http.StatusUnauthorized, err.Error())
		}

		if driverStatus != constants.Status_Active {
			logger.LogError(sessionId, "invalid driver status error: "+driverStatus)
			err = errors.New(constants.Operation_Not_Permitted)
			return utils.GeneralSocketResp(sessionId, http.StatusUnauthorized, err.Error())
		}

		logger.LogInfo("forwarding request to handler", sessionId)

		ctx := context.WithValue(*oldCtx, constants.Sessoin_KEY, token)
		ctx = context.WithValue(ctx, constants.Driver_KEY, driverId)
		ctx = context.WithValue(ctx, constants.Encrypted_Driver_KEY, encryptedDriverId)
		oldCtx = &ctx
	}

	logger.LogInfo("Response returned from SocketAuth", sessionId)

	return nil
}
