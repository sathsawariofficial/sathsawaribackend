package logger

import (
	"encoding/json"
	"fmt"
	"reflect"
	"rideshare/pkgs/configuration"
	"runtime"
	"strings"
	"time"
)

/*
DEBUG --> may print either just string ('preparing public data') or json object, make sure json objects printed
this case do not exceed 50 json objects, you can do this by making sure you use this only in places where you persume
json printed is smalled in size

For DEBUG levels, make sure they printed inside code, not start or end of a function.
*/
func LogDebug(msg string, session string, val interface{}) {
	if configuration.ConfigurationData.Tracing.FileModeEnable {
		_, file, line, _ := runtime.Caller(1)
		logMsg := formatLogMsg(file, session, msg, line, val)
		ZapLogger.Debug(logMsg)
		fmt.Println(logMsg)
	}
	//formatLogMsg(DEBUG, file, session, msg, line), zap.Any("data", val)
}

/*
DEBUG_2 --> may print both string and or json object of any size
*/
func LogDebug2(msg string, session string, val interface{}) {
	if configuration.ConfigurationData.Tracing.FileModeEnable {
		_, file, line, _ := runtime.Caller(1)
		logMsg := formatLogMsg(file, session, msg, line, val)
		ZapLogger.Debug(logMsg)
		fmt.Println(logMsg)
	}
}

// TODO: we are not printing any data using val so it is just there to keep backward compatibility becasue alot of logs
// have empty val parameter

/*
INFO --> must always print just string (like, "file entered" or "function xyz called" or similiar)
*/
func LogInfo(msg string, session string, val ...any) {
	if configuration.ConfigurationData.Tracing.FileModeEnable {
		_, file, line, _ := runtime.Caller(1)
		logMsg := formatLogMsg(file, session, msg, line, val)
		ZapLogger.Info(logMsg)
		fmt.Println(logMsg)
	}
}

/*
WARNING --> prints exact warning in string format, this is used when non breaking exception may happen because of user input
*/
func LogWarning(session string, val interface{}) {
	if configuration.ConfigurationData.Tracing.FileModeEnable {
		_, file, line, _ := runtime.Caller(1)
		logMsg := formatLogMsg(file, session, "", line, val)
		ZapLogger.Warn(logMsg)
		fmt.Println(logMsg)
	}
}

/*
ERROR --> prints exact warning in string format, this is used when breaking exception may happen because of user input
*/

func LogError(session string, val interface{}) {
	if configuration.ConfigurationData.Tracing.FileModeEnable {
		_, file, line, _ := runtime.Caller(1)

		var data string
		err, ok := val.(error)
		if ok {
			data = err.Error()
		} else {
			data = fmt.Sprintf("%v", val)
		}

		logMsg := formatLogMsg(file, session, "", line, data)

		ZapLogger.Error(logMsg)
		fmt.Println(logMsg)
	}
}

func LogFatal(session string, val interface{}) {
	if configuration.ConfigurationData.Tracing.FileModeEnable {
		_, file, line, _ := runtime.Caller(1)
		logMsg := formatLogMsg(file, session, "", line, val)
		ZapLogger.Fatal(logMsg)
		fmt.Println(logMsg)
	}
}

func LogPanic(session string, val interface{}) {
	if configuration.ConfigurationData.Tracing.FileModeEnable {
		_, file, line, _ := runtime.Caller(1)
		logMsg := formatLogMsg(file, session, "", line, val)
		ZapLogger.Panic(logMsg)
		fmt.Println(logMsg)
	}
}

// creates short path
func formatPath(path string) string {
	formattedPath := path
	splitedString := strings.Split(path, "/")
	pathLen := len(splitedString)

	if pathLen > 2 {
		formattedPath = fmt.Sprintf("%v", strings.Join(splitedString[pathLen-2:], "/"))
	}

	return formattedPath
}

func formatLogMsg(path, sessionId, msg string, line int, val any) string {
	filePath := formatPath(path)

	curTime := time.Now()
	milliseconds := curTime.Nanosecond() / 1e6 // Convert nanoseconds to milliseconds
	curDateStr := fmt.Sprintf("%s:%03d", curTime.Format(TIME_FORMAT), milliseconds)

	fileDetail := fmt.Sprintf("%s:%d", filePath, line)
	var valStr string

	// Enhanced nil check
	if val != nil && !isNil(val) {
		valStr = toString(val)
	}

	var formattedLog string
	if val != nil && !isNil(val) {
		formattedLog = fmt.Sprintf("{%s}  =>  %s:::  %s::%s   %s\n%s",
			"RIDE SHARE",
			curDateStr,
			fileDetail,
			sessionId,
			msg,
			valStr)
	} else {
		formattedLog = fmt.Sprintf("{%s}  =>  %s:::  %s::%s   %s",
			"RIDE SHARE",
			curDateStr,
			fileDetail,
			sessionId,
			msg)
	}

	return formattedLog
}

// Helper function to check nil values for interface{}
func isNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	// Check for pointers, interfaces, maps, slices, and channels
	return v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface ||
		v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Chan && v.IsNil()
}

func toString(value interface{}) string {
	// Handle nil
	if value == nil {
		return "null"
	}

	// Handle basic types
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	}

	// Handle arrays, slices, maps, and structs
	vType := reflect.TypeOf(value)
	switch vType.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		// Use JSON encoding for a consistent string representation
		bytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("error marshaling value: %v", err)
		}
		return string(bytes)
	}

	// Fallback to fmt.Sprintf for unknown types
	return fmt.Sprintf("%v", value)
}
