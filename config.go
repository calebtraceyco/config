package config_yaml

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	//"log"
	"net/http"
	"os"
)

type Config struct {
	AppName          yaml.Node         `yaml:"AppName"`
	Env              yaml.Node         `yaml:"Env"`
	Port             yaml.Node         `yaml:"Port"`
	ComponentConfigs ComponentConfigs  `yaml:"ComponentConfigs"`
	DatabaseConfigs  DatabaseConfigMap `yaml:"DatabaseConfigs"`
	ServiceConfigs   ServiceConfigMap  `yaml:"ServiceConfigs"`
	Hash             string            `yaml:"Hash"`
}

type ClientConfig struct {
	Timeout            yaml.Node `yaml:"Timeout"`
	IdleConnTimeout    yaml.Node `yaml:"IdleConnTimeout"`
	MaxIdleConsPerHost yaml.Node `yaml:"MaxIdleConsPerHost"`
	MaxConsPerHost     yaml.Node `yaml:"MaxConsPerHost"`
}

type ComponentConfigs struct {
	//TODO add logging
	Client ClientConfig
}

type ClientConfigFunc func(ClientConfig) *http.Client

type DbConfigBuildFn func(cfg *DatabaseConfig, appName string) (*sql.DB, error)

func NewFromFile(configPath string) *Config {
	log.Infoln(configPath)
	conf, confErrs := newFromFile(&builder{}, InitDbService, configPath)

	if len(confErrs) > 0 || conf == nil {
		for _, err := range confErrs {
			log.Panicf("Config error: %v\n", err.Error())
		}
		if conf == nil {
			log.Panicln("Config file not found")
		}
		log.Panicln("Exiting: Failed to load the config file")
	}
	return conf
}

func newFromFile(b configBuilder, dbBuilderFn DbConfigBuildFn, configPath string) (*Config, []error) {
	var errs []error
	var dbErr error
	var err error

	configFile, err := b.Load(configPath)
	if err != nil {
		return nil, []error{err}
	}

	defer func(configFile *os.File) {
		closeErr := configFile.Close()
		if closeErr != nil {
			log.Errorln(closeErr.Error())
		}
	}(configFile)

	err = b.Read(configFile)
	if err != nil {
		return nil, []error{err}
	}

	//clientFn := b.ClientInit()

	//for _, svcConfig := range b.Get().ServiceConfigs {
	//	svcConfig.HTTPClient = clientFn(svcConfig.SvcComponentConfigs().Client)
	//}

	for _, dbConfig := range b.Get().DatabaseConfigs {
		dbConfig.DB, dbErr = dbBuilderFn(dbConfig, b.Get().AppName.Value)
		if dbErr != nil {
			log.Error(err.Error())
			errs = append(errs, dbErr)
		}
	}

	return b.Get(), errs
}

func (c *Config) GetDatabaseConfig(name string) (*DatabaseConfig, error) {
	var database *DatabaseConfig
	database, ok := c.DatabaseConfigs[name]

	if !ok {
		err := fmt.Errorf("database config: %v not found", name)
		return nil, err
	}

	return database, nil
}

func (c *Config) GetServiceConfig(name string) (*ServiceConfig, error) {
	var service *ServiceConfig
	service, ok := c.ServiceConfigs[name]

	if !ok {
		err := fmt.Errorf("service config: %v not found", name)
		return nil, err
	}

	return service, nil
}
