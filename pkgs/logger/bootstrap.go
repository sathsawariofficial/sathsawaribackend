package logger

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"rideshare/pkgs/configuration"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var ZapLogger *zap.Logger

type OverloadedEncoder struct {
	zapcore.Encoder
	secretKey []string
}

var once sync.Once

func init() {
	once.Do(func() {
		LoggerInit()
	})
}

func LoggerInit() *zap.Logger {
	configuration.ConfigurationData.Tracing.EncoderConfig = zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	}
	encoder := zapcore.NewConsoleEncoder(configuration.ConfigurationData.Tracing.EncoderConfig)
	zapEncoder := OverloadedEncoder{encoder, []string{}}

	atomicAevel := zap.AtomicLevel{}
	err := atomicAevel.UnmarshalText([]byte(findLevel(configuration.ConfigurationData.Tracing.Level)))
	if err != nil {
		fmt.Println("Error occured while unmarshalling config log level into zap level", err.Error())
		panic(err.Error())
	}

	// file rotation
	fileRotationTime := time.Duration(configuration.ConfigurationData.Tracing.Rotationtime) * time.Minute
	fileRotation := zapcore.AddSync(NewTimeRotationWriter(configuration.ConfigurationData.Tracing.Name+".log", configuration.ConfigurationData.Tracing.Path, fileRotationTime, configuration.ConfigurationData.Tracing.MaxSize))

	// set up zap core configuration
	core := zapcore.NewCore(
		zapEncoder,
		zapcore.NewMultiWriteSyncer(fileRotation),
		atomicAevel,
	)

	ZapLogger = zap.New(core)

	return ZapLogger
}

func findLevel(levels []string) string {
	foundLevel := "debug"

	if len(levels) > 0 {
		foundLevel = levels[0]
		fmt.Println("levels is an array taking first value: ", foundLevel)
	} else {
		fmt.Println("levels are emptys setting to default: ", foundLevel)
	}

	return foundLevel
}

// create a custom EncodeEntry method to filter out secret data
func (m OverloadedEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	filtered := make([]zapcore.Field, 0, len(fields))
	for _, field := range fields {
		res := field

		// to hide the sensitive data
		target := field.Interface
		reflectedType := reflect.TypeOf(target)
		if target != nil {
			if reflectedType.Kind() == reflect.Struct || reflectedType.Kind() == reflect.Ptr {
				targetMap := make(map[string]interface{})

				inrec, _ := json.Marshal(target)
				err := json.Unmarshal(inrec, &targetMap)
				if err == nil && len(targetMap) > 0 {
					for _, key := range m.secretKey {
						delete(targetMap, key)
					}

					res.Interface = targetMap
				}
			}
		}

		filtered = append(filtered, res)
	}

	return m.Encoder.EncodeEntry(entry, filtered)
}
