package config

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/shanth1/gotools/consts"
	gotoolslog "github.com/shanth1/gotools/log"
)

type Config struct {
	Env consts.Env `yaml:"-" env:"APP_ENV" validate:"required,oneof=local dev stage prod"`

	Log LoggerConfig `mapstructure:"logger" yaml:"logger" validate:"required"`

	HTTP    HTTP    `mapstructure:"http" yaml:"http" validate:"required"`
	Storage Storage `mapstructure:"storage" yaml:"storage" validate:"required"`
	Image   Image   `mapstructure:"image" yaml:"image" validate:"required"`
}

// LoggerConfig оборачивает gotoolslog.Config и добавляет env-теги
type LoggerConfig struct {
	App          string `mapstructure:"app" yaml:"app" env:"LOGGER_APP"`
	Service      string `mapstructure:"service" yaml:"service" env:"LOGGER_SERVICE"`
	Level        string `mapstructure:"level" yaml:"level" env:"LOGGER_LEVEL"`
	UDPAddress   string `mapstructure:"udp_address" yaml:"udp_address" env:"LOGGER_UDP"`
	EnableCaller bool   `mapstructure:"enable_caller" yaml:"enable_caller" env:"LOGGER_ENABLE_CALLER"`
	consts.Role
	Console bool `mapstructure:"console" yaml:"console" env:"LOGGER_CONSOLE"`
}

// ToGotoolsConfig конвертирует LoggerConfig в gotoolslog.Config
func (l LoggerConfig) ToGotoolsConfig() gotoolslog.Config {
	return gotoolslog.Config{
		App:          l.App,
		Service:      l.Service,
		Level:        l.Level,
		UDPAddress:   l.UDPAddress,
		EnableCaller: l.EnableCaller,
		Console:      l.Console,
	}
}

type HTTP struct {
	Address         string        `mapstructure:"address" yaml:"address" env:"HTTP_ADDR" validate:"required,hostname_port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" yaml:"read_timeout" validate:"min=100ms"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" yaml:"write_timeout" validate:"min=100ms"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout" validate:"min=1s"`
	MaxUploadSizeMB int64         `mapstructure:"max_upload_size_mb" yaml:"max_upload_size_mb" validate:"min=1"`
}

func (h HTTP) MaxUploadSizeBytes() int64 {
	return h.MaxUploadSizeMB * 1024 * 1024
}

type Storage struct {
	Path             string `mapstructure:"path" yaml:"path" validate:"required"`
	CompressedSubdir string `mapstructure:"compressed_subdir" yaml:"compressed_subdir" validate:"required"`
	TmpSubdir        string `mapstructure:"tmp_subdir" yaml:"tmp_subdir"`
}

type Image struct {
	DefaultFormat  string   `mapstructure:"default_format" yaml:"default_format" validate:"required,oneof=jpeg png webp"`
	DefaultQuality int      `mapstructure:"default_quality" yaml:"default_quality" validate:"min=1,max=100"`
	MaxWidth       int      `mapstructure:"max_width" yaml:"max_width" validate:"min=100"`
	MaxHeight      int      `mapstructure:"max_height" yaml:"max_height" validate:"min=100"`
	AllowFormats   []string `mapstructure:"allow_formats" yaml:"allow_formats" validate:"required"`
}

func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
