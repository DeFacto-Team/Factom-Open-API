package config

import (
	"github.com/go-yaml/yaml"
	"github.com/jinzhu/configor"
	"github.com/mcuadros/go-defaults"
	"io/ioutil"
)

// App config struct
type Config struct {
	Admin struct {
		User     string `default:""`
		Password string `default:""`
	}
	API struct {
		HTTPPort int  `required:"true" default:"8081"`
		Logging  bool `required:"true" default:"true"`
		LogLevel int  `required:"true" default:"4"`
	}
	Store struct {
		Host     string `required:"true" default:"foa-db"`
		Port     int    `required:"true" default:"5432"`
		User     string `required:"true" default:"postgres"`
		Password string `required:"true" default:"postgres"`
		DBName   string `required:"true" default:"postgres"`
	}
	Factom struct {
		URL       string `default:"https://api.factomd.net"`
		User      string `default:""`
		Password  string `default:""`
		EsAddress string `default:""`
	}
}

// Create config from configFile
func NewConfig(configFile string) (*Config, error) {

	config := new(Config)
	defaults.SetDefaults(config)

	configBytes, err := ioutil.ReadFile(configFile)
	if err == nil {
		err = yaml.Unmarshal(configBytes, &config)
		if err != nil {
			return nil, err
		}
	}

	if err := configor.Load(config); err != nil {
		return nil, err
	}
	return config, nil
}
