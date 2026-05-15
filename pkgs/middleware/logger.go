package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"rideshare/pkgs/logger"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Generate a unique session ID for the request
		session := xid.New().String()
		ctx.Set(Middleware_Session, session)

		// Wrap the ResponseWriter to capture response body
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw

		// Log the request details
		logger.LogDebug("Received Request [logging middleware] ", session, LogBody{
			IP:          ctx.ClientIP(),
			URL:         ctx.Request.URL,
			Host:        ctx.Request.Host,
			Method:      ctx.Request.Method,
			Header:      ctx.Request.Header,
			Query:       ctx.Request.URL.Query(),     // Add query parameters
			RequestBody: getGinRequest(session, ctx), // Capture the request body
		})

		// Proceed to the next handler
		ctx.Next()

		// Log the response details
		logger.LogDebug("Response Returned [logging middleware] ", session, LogBody{
			IP:           ctx.ClientIP(),
			URL:          ctx.Request.URL,
			Host:         ctx.Request.Host,
			Method:       ctx.Request.Method,
			ResponseCode: ctx.Writer.Status(),
			Header:       ctx.Writer.Header(),
			Body:         blw.body.String(),
		})
	}
}

type LogBody struct {
	IP           string      `json:"ip"`
	URL          *url.URL    `json:"url"`
	Host         string      `json:"host"`
	Method       string      `json:"method"`
	Header       http.Header `json:"header"`
	Query        url.Values  `json:"query,omitempty"`       // Added field for query parameters
	RequestBody  string      `json:"requestBody,omitempty"` // Added request body
	ResponseCode int         `json:"responseCode,omitempty"`
	Body         any         `json:"body"`
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func getGinRequest(sessionId string, ctx *gin.Context) string {
	// Read the request body
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logger.LogError(sessionId, fmt.Sprintf("failed to read request body: %v", err.Error()))
		return ""
	}

	// Repopulate the request body so it can be read again
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Log the request body
	logger.LogDebug2("Request body", sessionId, string(bodyBytes))

	return string(bodyBytes)
}
