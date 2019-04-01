package krawler

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config defines a struct that will be shared among all components.
// It contains information to instruct the behavior of the crawler.
type Config struct {
	RequestUserAgent      string
	RequestTimeout        time.Duration
	RequestMaxRetryTimes  int
	RequestFollowRedirect bool
	RequestConcurrency    int
}

// NewConfigFromRawConfig creates config from a RawConfig and do validation.
func NewConfigFromRawConfig(rawConfig *RawConfig) (*Config, error) {
	logLevel, err := log.ParseLevel(strings.ToUpper(rawConfig.Logger.Level))
	if err != nil {
		log.Warnf("%v is invalid for log level, set to default value %v", rawConfig.Logger.Level, DefaultRawConfig.Logger.Level)
		defaultLogLevel, err := log.ParseLevel(DefaultRawConfig.Logger.Level)
		if err != nil {
			panic(err)
		}
		logLevel = defaultLogLevel
	}
	log.SetLevel(logLevel)

	if !rawConfig.Logger.Console {
		log.SetOutput(ioutil.Discard)
	}

	if !rawConfig.Logger.Console && rawConfig.Logger.FilePath == "" {
		log.SetLevel(log.TraceLevel)
	}

	if rawConfig.Logger.FilePath != "" {
		fileHook := lfshook.LfsHook{}
		fileHook.SetFormatter(log.StandardLogger().Formatter)
		fileHook.SetDefaultPath(rawConfig.Logger.FilePath)
		log.AddHook(&fileHook)
	}

	if rawConfig.Request.Timeout <= 0 {
		log.Warnf("%v is invalid for request timeout configuration, set to default value %v", rawConfig.Request.Timeout, DefaultRawConfig.Request.Timeout)
		rawConfig.Request.Timeout = DefaultRawConfig.Request.Timeout
	}

	if rawConfig.Request.Concurrency <= 0 {
		log.Warnf("%v is invalid for request timeout configuration, set to default value %v", rawConfig.Request.Concurrency, DefaultRawConfig.Request.Concurrency)
		rawConfig.Request.Concurrency = DefaultRawConfig.Request.Concurrency
	}

	config := new(Config)
	config.RequestTimeout = time.Duration(rawConfig.Request.Timeout) * time.Millisecond
	config.RequestFollowRedirect = rawConfig.Request.FollowRedirect
	config.RequestMaxRetryTimes = rawConfig.Request.MaxRetryTimes
	config.RequestConcurrency = rawConfig.Request.Concurrency

	return config, nil
}

// LoadConfigFromFile loads config from opened file.
// The file will not be closed by this function.
func LoadConfigFromFile(file *os.File) (*Config, error) {
	rawConfig := DefaultRawConfig

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(content, &rawConfig)
	if err != nil {
		return nil, err
	}

	return NewConfigFromRawConfig(&rawConfig)
}

// LoadConfigFromPath loads config from file path.
// The file will be closed by this function.
func LoadConfigFromPath(configPath string) (config *Config, err error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("fail to open config file, reason: %v", err)
	}

	defer func() {
		if newErr := file.Close(); newErr != nil {
			err = fmt.Errorf("fail to close config file, reason: %v", newErr)
			config = nil
		}
	}()

	config, err = LoadConfigFromFile(file)
	return
}
