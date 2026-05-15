package configuration

import "go.uber.org/zap/zapcore"

type Configuration struct {
	RestPort           string         `json:"rest_port"`
	SocketPort         string         `json:"socket_port"`
	Envirnment         string         `json:"envirnment"`
	RideCloseScheduler int            `json:"ride_close_scheduler"`
	PageSize           int            `json:"page_size"`
	Timeout            int            `json:"timeout"`
	Database           DatabaseConfig `json:"database"`
	Tracing            Tracing        `json:"tracing"`
	Auth               Auth           `json:"auth"`
	General            General        `json:"general"`
}

type General struct {
	DocsPath string `json:"docs_path"`
}

type Auth struct {
	Secrect        string `json:"secret"`
	ExpirationTime int    `json:"expTime"`
	IV16           string `json:"iv16"`
}

type Tracing struct {
	Level             []string              `json:"level"`
	FileModeEnable    bool                  `json:"fileModeEnable"`
	ConsoleModeEnable bool                  `json:"consoleModeEnable"`
	Rotationtime      int                   `json:"rotationtime"`
	Path              string                `json:"path"`
	Name              string                `json:"name"`
	MaxSize           int                   `json:"maxSize"`
	EncoderConfig     zapcore.EncoderConfig `json:"encoderConfig"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `json:"postgres"`
	Redis    RedisConfig    `json:"redis"`
}

type PostgresConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Port     string `json:"port"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`
	TTL      int    `json:"ttl"`
}
