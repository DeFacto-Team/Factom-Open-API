package config

import (
	"flag"
	"github.com/go-yaml/yaml"
	"github.com/jinzhu/configor"
	"io/ioutil"
)

// App config struct
type Config struct {
	LogLevel  int `default:"4"`
	GzipLevel int `default:"-1"`
	API       struct {
		HTTPPort int  `default:"8080"`
		Logging  bool `default:"false"`
	}
	Store struct {
		Host     string `required:"true"`
		Port     int    `required:"true"`
		User     string `required:"true"`
		Password string `required:"true"`
		DBName   string `required:"true"`
	}
	Factom struct {
		URL       string `required:"true"`
		User      string `default:"",required:"true"`
		Password  string `default:"",required:"true"`
		EsAddress string `required:"true"`
	}
}

// Create config from configFile
func NewConfig(configFile string) (*Config, error) {

	config := &Config{}

	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		return nil, err
	}

	flag.IntVar(&config.API.HTTPPort, "port", config.API.HTTPPort, "Open API port")
	flag.BoolVar(&config.API.Logging, "logging", config.API.Logging, "Enable logging")

	flag.IntVar(&config.LogLevel, "loglevel", config.LogLevel, "Log level (4 - info, 5 - debug, 6 - debug+db)")
	flag.IntVar(&config.GzipLevel, "gziplevel", config.GzipLevel, "Gzip level")

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
