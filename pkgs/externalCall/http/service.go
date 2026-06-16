package httpcall

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/logger"
	"time"
)

func MakeRequest(ctx context.Context, opt RequestOptions) (respBody []byte, err error) {
	logger.LogInfo("Request received in MakeRequest", opt.SessionID)
	logger.LogDebug("Request received in MakeRequest", opt.SessionID, opt)

	var bodyReader io.Reader

	ctx, cancel := context.WithTimeout(ctx, time.Duration(configuration.ConfigurationData.Timeout)*time.Second)
	defer cancel()

	// encode body if provided
	if opt.Body != nil {
		b, err := json.Marshal(opt.Body)
		if err != nil {
			logger.LogError(opt.SessionID, fmt.Errorf("failed to marshal body: %w", err))
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		opt.Method,
		opt.Path,
		bodyReader,
	)
	if err != nil {
		logger.LogError(opt.SessionID, fmt.Errorf("failed to create request: %w", err))
		return
	}

	// default headers
	req.Header.Set("Content-Type", "application/json")

	// session id propagation
	if opt.SessionID != "" {
		req.Header.Set("X-Session-Id", opt.SessionID)
	}

	// custom headers
	for k, v := range opt.Headers {
		req.Header.Set(k, v)
	}

	resp, err := HttpClient.httpClient.Do(req)
	if err != nil {
		logger.LogError(opt.SessionID, err)
		return
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError(opt.SessionID, fmt.Errorf("failed reading response: %w", err))
		return
	}

	logger.LogInfo("Response returned from MakeRequest", opt.SessionID)

	return
}
