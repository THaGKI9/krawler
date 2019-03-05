package krawler

import (
	"io"

	"github.com/thagki9/krawler/constant"
	"gopkg.in/yaml.v2"
)

// RawConfig defines the structure of a YAML config file.
type RawConfig struct {
	Logger  RawLoggerConfig  `yaml:"logger"`
	Request RawRequestConfig `yaml:"request"`
}

// RawLoggerConfig defines the structure of LoggerConfig
type RawLoggerConfig struct {
	Level    string `yaml:"level"`
	Console  bool   `yaml:"console"`
	FilePath string `yaml:"filepath"`
}

// RawRequestConfig defines the structure of RequestConfig
type RawRequestConfig struct {
	UserAgent      string `yaml:"userAgent"`
	Timeout        int    `yaml:"timeout"`
	MaxRetryTimes  int    `yaml:"maxRetryTimes"`
	FollowRedirect bool   `yaml:"followRedirect"`
	Concurrency    int    `yaml:"concurrency"`
}

// DumpYAML will dump raw config into YAML
func (c *RawConfig) DumpYAML(writer io.Writer) error {
	content, err := yaml.Marshal(c)
	if err != nil {
		return nil
	}

	_, err = writer.Write(content)
	if err != nil {
		return nil
	}

	return nil
}

// DefaultRawConfig defines the default value of RawConfig.
var DefaultRawConfig = RawConfig{
	Logger: RawLoggerConfig{
		Level:    "debug",
		Console:  true,
		FilePath: "",
	},
	Request: RawRequestConfig{
		Concurrency:    5,
		FollowRedirect: true,
		MaxRetryTimes:  3,
		Timeout:        5000,
		UserAgent:      "krawler/" + constant.KrawlerVersion,
	},
}
