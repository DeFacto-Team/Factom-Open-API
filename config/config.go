package config

import (
	"github.com/go-yaml/yaml"
	"github.com/jinzhu/configor"
	"github.com/mcuadros/go-defaults"
	"io/ioutil"
	"os"
)

// App config struct
type Config struct {
	Admin struct {
		User     string `default:"" json:"adminUser" form:"adminUser" query:"adminUser"`
		Password string `default:"" json:"adminPassword" form:"adminPassword" query:"adminPassword"`
	}
	API struct {
		HTTPPort int  `required:"true" default:"8081" json:"apiHTTPPort" form:"apiHTTPPort" query:"apiHTTPPort"`
		Logging  bool `required:"true" default:"true" json:"apiLogging" form:"apiLogging" query:"apiLogging"`
		LogLevel int  `required:"true" default:"4" json:"apiLogLevel" form:"apiLogLevel" query:"apiLogLevel"`
	}
	Store struct {
		Host     string `required:"true" default:"foa-db" json:"storeHost" form:"storeHost" query:"storeHost"`
		Port     int    `required:"true" default:"5432" json:"storePort" form:"storePort" query:"storePort"`
		User     string `required:"true" default:"postgres" json:"storeUser" form:"storeUser" query:"storeUser"`
		Password string `required:"true" default:"postgres" json:"storePassword" form:"storePassword" query:"storePassword"`
		DBName   string `required:"true" default:"postgres" json:"storeDBName" form:"storeDBName" query:"storeDBName"`
	}
	Factom struct {
		URL       string `default:"https://api.factomd.net" json:"factomURL" form:"factomURL" query:"factomURL"`
		User      string `default:"" json:"factomUser" form:"factomUser" query:"factomUser"`
		Password  string `default:"" json:"factomPassword" form:"factomPassword" query:"factomPassword"`
		EsAddress string `default:"" json:"factomEsAddress" form:"factomEsAddress" query:"factomEsAddress"`
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

func UpdateConfig(configFile string, newConf *Config) error {

	newYaml, err := yaml.Marshal(&newConf)
	if err != nil {
		return err
	}

	// write to file
	f, err := os.Create("/tmp/newconfig")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configFile, newYaml, 0644)
	if err != nil {
		return err
	}

	f.Close()

	return nil

}
