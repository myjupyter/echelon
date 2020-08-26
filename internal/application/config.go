package application

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
)

const (
	configFileRegexp = `\w+\.(json|toml|yaml|yml)`
	configPathRegexp = `(.+)\/`
)

type Config struct {
	dataBaseAddr string
	dataBaseUser string
	dataBasePass string

	appAddr string

	logLevel string

	configurator *viper.Viper
}

func NewConfig(configs string) *Config {
	re, err := regexp.Compile(configFileRegexp)
	if err != nil {
		log.Fatal(err)
	}

	r, err := regexp.Compile(configPathRegexp)
	if err != nil {
		log.Fatal(err)
	}

	fileWithExt := re.FindString(configs)
	configPath := r.FindString(configs)

	if len(fileWithExt) == 0 || len(configPath) == 0 {
		log.Fatal("Wrong path to config file")
	}

	splited := strings.Split(fileWithExt, ".")

	configurator := viper.New()
	configurator.SetConfigName(splited[0])
	configurator.SetConfigType(splited[1])
	configurator.AddConfigPath(configPath)

	return &Config{configurator: configurator}
}

func (c *Config) loadConfig() error {
	if err := c.configurator.ReadInConfig(); err != nil {
		return err
	}

	c.dataBaseAddr = c.configurator.GetString("database.host") + ":" + c.configurator.GetString("database.port")
	c.dataBaseUser = c.configurator.GetString("database.login")
	c.dataBasePass = c.configurator.GetString("database.password")
	c.appAddr = c.configurator.GetString("application.host") + ":" + c.configurator.GetString("application.port")
	c.logLevel = c.configurator.GetString("log_level")

	return nil
}
