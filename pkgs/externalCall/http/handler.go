package httpcall

import (
	"net/http"
	"rideshare/pkgs/configuration"
	"time"
)

var HttpClient *Client

type Client struct {
	httpClient *http.Client
}

// NewClient creates a reusable HTTP client (DO NOT recreate per request)
func NewClient() {
	HttpClient = &Client{
		httpClient: &http.Client{
			Timeout: time.Duration(configuration.ConfigurationData.Timeout) * time.Second, // global timeout safeguard
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
			},
		},
	}
}
