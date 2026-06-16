package passenger

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

func BookSeatHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in BookSeatHandler", sessionId)

	var request BookSeatRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Unable_To_Do_Job, "book ride"),
		})
		return
	}

	logger.LogDebug2("Response received in BookSeatHandler", sessionId, request)

	err := ValidateBookSeat(sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = BookSeat(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "failed to book seat error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bookingResp := utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Booked the seat"))

	logger.LogInfo("Response returned from GetBookSeatsHandler", sessionId)
	logger.LogDebug2("Response returned from GetBookSeatsHandler", sessionId, bookingResp)

	ctx.JSON(http.StatusOK, bookingResp)
}
