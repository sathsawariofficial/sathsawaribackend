package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/middleware"
	"rideshare/pkgs/utils"
	"strings"

	"github.com/gorilla/websocket"
)

func StatsSocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := utils.GetSocketUpgrader()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.LogError("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	remoteAddr := r.RemoteAddr
	logger.LogDebug("WebSocket connection established from", constants.DEFAULT_SESSION, remoteAddr)

	for {
		// Each message can have its own sessionId if needed
		sessionId := utils.RandomId()

		messageType, request, err := conn.ReadMessage()
		if err != nil {

			// Normal client disconnect
			if websocket.IsCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				logger.LogError(sessionId, fmt.Sprintf("client disconnected: %s", err.Error()))
				break
			}

			// Abnormal closure (1006)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				logger.LogError(sessionId, fmt.Sprintf("abnormal websocket close (1006): %s", err.Error()))
				break
			}

			// Real unexpected error (network, panic, broken pipe)
			logger.LogError(sessionId, fmt.Sprintf("read error: %s", err.Error()))
			continue
		}

		logger.LogInfo("----------------------HTTP Request Recieved----------------------", sessionId)
		logger.LogInfo("Received request in StatsSocketHandler (HTTP) ", sessionId)
		logger.LogDebug2("Received request in StatsSocketHandler", sessionId, fmt.Sprintf("Request: %v", string(request)))

		response := StatsSocketMultiplexer(sessionId, request)

		if writeErr := conn.WriteMessage(messageType, []byte(response)); writeErr != nil {
			logger.LogError(sessionId, fmt.Errorf("WebSocket write error: %v", writeErr))
			break // Cannot recover from write failure
		}

		logger.LogInfo("----------------------HTTP Response Returned----------------------", sessionId)
		logger.LogInfo("Returned response from StatsSocketHandler Info (HTTP) ", sessionId)
		logger.LogDebug2("Returned response from StatsSocketHandler", sessionId, fmt.Sprintf("Response: %v", string(response)))
	}

	logger.LogDebug("WebSocket connection closed from", constants.DEFAULT_SESSION, remoteAddr)
}

/*
hanlde stats sockets
*/
func StatsSocketMultiplexer(sessionId string, request []byte) []byte {
	logger.LogInfo("Request recieved in StatsSocketMultiplexer ", sessionId, "")
	logger.LogDebug2("Request received in StatsSocketMultiplexer ", sessionId, request)

	var response []byte

	var wrapperReq StatsRequestWrapper
	err := json.Unmarshal(request, &wrapperReq)
	if err != nil {
		logger.LogError(sessionId, err)
		return utils.GeneralSocketResp(sessionId, http.StatusBadRequest, fmt.Sprintf(constants.Invalid_Data, "request is "))
	}

	ctx := context.Background()
	requestId := utils.RandomId()

	resp := middleware.SocketAuth(&ctx, constants.DEFAULT_SESSION, wrapperReq.Token)
	if resp == nil {
		return resp
	}

	var data any
	switch strings.ToUpper(wrapperReq.Type) {
	case constants.SR_PLACE_SEARCH:
		data = LocationSearch(ctx, sessionId, requestId, wrapperReq.Type, request)
	default:
		return utils.GeneralSocketResp(sessionId, http.StatusBadRequest, fmt.Sprintf(constants.Invalid_Data, "Type"))
	}

	if bVal, ok := isByteSlice(data); ok {
		response = bVal
	} else {
		response, err = json.Marshal(data)
		if err != nil {
			logger.LogError(sessionId, err)
			return utils.GeneralSocketResp(sessionId, http.StatusBadRequest, fmt.Sprintf(constants.Unable_To_Do_Job, constants.Perform_this_operation))
		}
	}

	logger.LogInfo("Response returned from StatsSocketMultiplexer ", sessionId)
	logger.LogDebug2("Response returned from StatsSocketMultiplexer", sessionId, string(response))

	return response
}
