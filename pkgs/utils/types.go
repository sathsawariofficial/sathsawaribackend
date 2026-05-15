package utils

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"strconv"
)

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type Gender string

func (g *Gender) IsValid() bool {
	switch g.String() {
	case constants.Gender_Male, constants.Gender_Female:
		return true
	default:
		return false
	}
}

func (g Gender) String() string {
	return string(g)
}

func ToString(val interface{}) string {
	return fmt.Sprintf("%v", val)
}

func ToInt(val string) int {
	var iVal int

	if !IsStringEmpty(val) {
		iVal64, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			logger.LogError(constants.DEFAULT_SESSION, fmt.Sprintf("Failed to convert value %v to int: %v", val, err.Error()))
		}

		iVal = int(iVal64)
	}

	return iVal
}

func ToUInt32(val string) uint32 {
	var iVal uint32

	if !IsStringEmpty(val) {
		iVal64, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			logger.LogError(constants.DEFAULT_SESSION, fmt.Sprintf("Failed to convert value %v to uint32: %v", val, err.Error()))
		}

		iVal = uint32(iVal64)
	}

	return iVal
}

func ToInt32(val string) int32 {
	var iVal int32

	if !IsStringEmpty(val) {
		iVal64, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			logger.LogError(constants.DEFAULT_SESSION, fmt.Sprintf("Failed to convert value %v to int32: %v", val, err.Error()))
		}
		iVal = int32(iVal64)
	}

	return iVal
}

func ToInt64(val string) int64 {
	var iVal int64
	var err error

	if !IsStringEmpty(val) {
		iVal, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			logger.LogError(constants.DEFAULT_SESSION, fmt.Sprintf("Failed to convert value %v to int64: %v", val, err.Error()))
		}
	}

	return iVal
}

func ToFloat64(val string) float64 {
	var iVal float64
	var err error

	if !IsStringEmpty(val) {
		iVal, err = strconv.ParseFloat(val, 64)
		if err != nil {
			logger.LogError(constants.DEFAULT_SESSION, fmt.Sprintf("Failed to convert value %v to float64: %v", val, err.Error()))
		}
	}

	return iVal
}

func MustFloat64(val string) (float64, error) {
	var iVal float64
	var err error

	if !IsStringEmpty(val) {
		iVal, err = strconv.ParseFloat(val, 64)
		if err != nil {
			logger.LogError(constants.DEFAULT_SESSION, fmt.Sprintf("Failed to convert value %v to float64: %v", val, err.Error()))
		}
	}

	return iVal, err
}

type Location struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	District string  `json:"district"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Category string  `json:"category"`
}
