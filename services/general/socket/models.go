package socket

import "rideshare/pkgs/utils"

type StatsRequestWrapper struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type LocationSearchRequest struct {
	DeviceId string  `json:"device_id"`
	Place    string  `json:"place"`
	UserLat  float64 `json:"lat"`
	UserLng  float64 `json:"long"`
}

type LocationSearchResponse struct {
	Locations []utils.Location `json:"locations"`
}
