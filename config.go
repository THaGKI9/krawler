package krawler

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config defines a struct that will be shared among all components.
// It contains information to instruct the behavior of the crawler
// along with some internal stuffs like logger, etc.
type Config struct {
	Logger *log.Logger

	RequestUserAgent      string
	RequestTimeout        time.Duration
	RequestMaxRetryTimes  int
	RequestFollowRedirect bool
	RequestConcurrency    int
}

// NewConfig loads default configuration.
func NewConfig() *Config {
	config, _ := NewConfigFromRawConfig(&DefaultRawConfig)
	return config
}

// NewConfigFromRawConfig creates config from a RawConfig and do validation.
func NewConfigFromRawConfig(rawConfig *RawConfig) (*Config, error) {
	logFormatter := &log.TextFormatter{FullTimestamp: true}

	logger := log.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(logFormatter)

	logLevel, err := log.ParseLevel(strings.ToUpper(rawConfig.Logger.Level))
	if err != nil {
		logger.Warnf("%v is invalid for logger level, set to default value %v", rawConfig.Logger.Level, DefaultRawConfig.Logger.Level)
		defaultLogLevel, err := log.ParseLevel(DefaultRawConfig.Logger.Level)
		if err != nil {
			panic(err)
		}
		logLevel = defaultLogLevel
	}
	logger.SetLevel(logLevel)

	if !rawConfig.Logger.Console {
		logger.SetOutput(ioutil.Discard)
	}

	if !rawConfig.Logger.Console && rawConfig.Logger.FilePath == "" {
		logger.SetLevel(log.TraceLevel)
	}

	if rawConfig.Logger.FilePath != "" {
		fileHook := lfshook.LfsHook{}
		fileHook.SetFormatter(logFormatter)
		fileHook.SetDefaultPath(rawConfig.Logger.FilePath)
		logger.AddHook(&fileHook)
	}

	if rawConfig.Request.Timeout <= 0 {
		logger.Warnf("%v is invalid for request timeout configuration, set to default value %v", rawConfig.Request.Timeout, DefaultRawConfig.Request.Timeout)
		rawConfig.Request.Timeout = DefaultRawConfig.Request.Timeout
	}

	if rawConfig.Request.Concurrency <= 0 {
		logger.Warnf("%v is invalid for request timeout configuration, set to default value %v", rawConfig.Request.Concurrency, DefaultRawConfig.Request.Concurrency)
		rawConfig.Request.Concurrency = DefaultRawConfig.Request.Concurrency
	}

	config := new(Config)
	config.Logger = logger
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
	defer func() {
		if newErr := file.Close(); newErr != nil {
			err = newErr
			config = nil
		}
	}()

	if err != nil {
		return nil, err
	}

	config, err = LoadConfigFromFile(file)
	return
}
