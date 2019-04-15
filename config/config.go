package config

import (
	"flag"
	"github.com/go-yaml/yaml"
	"github.com/jinzhu/configor"
	"github.com/mcuadros/go-defaults"
	"io/ioutil"
)

// App config struct
type Config struct {
	API struct {
		HTTPPort  int  `required:"true" default:"8081"`
		Logging   bool `required:"true" default:"true"`
		LogLevel  int  `required:"true" default:"4"`
		GzipLevel int  `required:"true" default:"-1"`
	}
	Store struct {
		Host     string `required:"true" default:"foa-db"`
		Port     int    `required:"true" default:"5432"`
		User     string `required:"true" default:"postgres"`
		Password string `required:"true" default:"postgres"`
		DBName   string `required:"true" default:"postgres"`
	}
	Factom struct {
		URL       string `required:"true" default:"https://api.factomd.net"`
		User      string `default:""`
		Password  string `default:""`
		EsAddress string `required:"true" default:""`
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

	flag.IntVar(&config.API.HTTPPort, "port", config.API.HTTPPort, "Open API port")
	flag.BoolVar(&config.API.Logging, "logging", config.API.Logging, "Enable logging")
	flag.IntVar(&config.API.LogLevel, "loglevel", config.API.LogLevel, "Log level (4 - info, 5 - debug, 6 - debug+db)")
	flag.IntVar(&config.API.GzipLevel, "gziplevel", config.API.GzipLevel, "Gzip level")

	flag.StringVar(&config.Store.Host, "dbhost", config.Store.Host, "Postgres DB host")
	flag.IntVar(&config.Store.Port, "dbport", config.Store.Port, "Postgres DB port")
	flag.StringVar(&config.Store.User, "dbuser", config.Store.User, "Postgres DB user")
	flag.StringVar(&config.Store.Password, "dbpass", config.Store.Password, "Postgres DB password")
	flag.StringVar(&config.Store.DBName, "dbname", config.Store.DBName, "Postgres DB name")

	flag.StringVar(&config.Factom.URL, "factomd", config.Factom.URL, "factomd server with port")
	flag.StringVar(&config.Factom.User, "factomduser", config.Factom.User, "factomd user")
	flag.StringVar(&config.Factom.Password, "factomdpass", config.Factom.Password, "factomd password")
	flag.StringVar(&config.Factom.EsAddress, "esaddress", config.Factom.EsAddress, "Es address")

	flag.Parse()

	if err := configor.Load(config); err != nil {
		return nil, err
	}
	return config, nil
}
