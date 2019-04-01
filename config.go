package krawler

import (
	"io/ioutil"
	"time"

	"github.com/rifflock/lfshook"

	log "github.com/sirupsen/logrus"
	"github.com/thagki9/krawler/constant"
)

// Config defines the structure of a YAML config file.
type Config struct {
	Logger  LoggerConfig
	Request RequestConfig
}

// LoggerConfig defines the structure of LoggerConfig
type LoggerConfig struct {
	Level    log.Level
	Console  bool
	FilePath string
}

// RequestConfig defines the structure of RequestConfig
type RequestConfig struct {
	UserAgent      string
	Timeout        time.Duration
	MaxRetryTimes  int
	FollowRedirect bool
	Concurrency    int
}

func GetDefaultConfig() *Config {
	return &Config{
		Logger: LoggerConfig{
			Level:    log.DebugLevel,
			Console:  true,
			FilePath: "",
		},
		Request: RequestConfig{
			Concurrency:    5,
			FollowRedirect: true,
			MaxRetryTimes:  3,
			Timeout:        time.Second * 5,
			UserAgent:      "krawler/" + constant.KrawlerVersion,
		},
	}

}

// defaultConfig defines the default value of Config.
var defaultConfig = GetDefaultConfig()

// checkConfig check and fix the config if necessary
func (config *Config) checkConfig() {
	log.SetLevel(config.Logger.Level)

	if !config.Logger.Console {
		log.SetOutput(ioutil.Discard)
	}

	if !config.Logger.Console && config.Logger.FilePath == "" {
		log.SetLevel(log.TraceLevel)
	}

	if config.Logger.FilePath != "" {
		fileHook := lfshook.LfsHook{}
		fileHook.SetFormatter(log.StandardLogger().Formatter)
		fileHook.SetDefaultPath(config.Logger.FilePath)
		log.AddHook(&fileHook)
	}

	if config.Request.Timeout <= 0 {
		log.Warnf("%v is invalid for request timeout configuration, set to default value %v", config.Request.Timeout, defaultConfig.Request.Timeout)
		config.Request.Timeout = defaultConfig.Request.Timeout
	}

	if config.Request.Concurrency <= 0 {
		log.Warnf("%v is invalid for request timeout configuration, set to default value %v", config.Request.Concurrency, defaultConfig.Request.Concurrency)
		config.Request.Concurrency = defaultConfig.Request.Concurrency
	}
}
