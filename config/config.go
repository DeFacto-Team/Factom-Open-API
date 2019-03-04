package config

import (
	"github.com/jinzhu/configor"
)

// App config struct
type Config struct {
	ConfigFile string
	LogLevel   uint32 `default:"4"`
	GzipLevel  int    `default:"-1"`
	Api        struct {
		HttpPort int  `default:"8080"`
		Logging  bool `default:"false"`
	}
	Store struct {
		Host     string `required:"true"`
		Port     int    `required:"true"`
		User     string `required:"true"`
		Password string `required:"true"`
		Dbname   string `required:"true"`
	}
	Factom struct {
		URL       string `required:"true"`
		User      string `default:"", required:"true"`
		Password  string `default:"", required:"true"`
		EsAddress string `required:"true"`
	}
}

// Create config from configFile
func NewConfig(configFile string) (*Config, error) {
	config := &Config{ConfigFile: configFile}
	if err := configor.Load(config, configFile); err != nil {
		return nil, err
	}
	return config, nil
}
