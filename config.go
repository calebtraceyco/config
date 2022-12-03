package config_yaml

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	AppName          yaml.Node         `yaml:"AppName"`
	Env              yaml.Node         `yaml:"Env"`
	Port             yaml.Node         `yaml:"Port"`
	ComponentConfigs ComponentConfigs  `yaml:"ComponentConfigs"`
	DatabaseConfigs  DatabaseConfigMap `yaml:"DatabaseConfigs"`
	ServiceConfigs   ServiceConfigMap  `yaml:"ServiceConfigs"`
	CrawlerConfigs   CrawlConfigMap    `yaml:"CrawlerConfigs"`
	Hash             string            `yaml:"Hash"`
}

type ComponentConfigs struct {
	//TODO add logging
	Client ClientConfig
}

func NewFromFile(configPath string) *Config {
	log.Infoln(configPath)
	conf, confErrs := newFromFile(&builder{}, configPath)

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

func newFromFile(b configBuilder, configPath string) (*Config, []error) {
	var errs []error
	var dbErr, collErr, err error

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

	if err = b.Read(configFile); err != nil {
		return nil, []error{err}

	}

	for _, cConfig := range b.Get().CrawlerConfigs {
		cConfig.Collector, collErr = cConfig.CrawlerService()
		if collErr != nil {
			log.Error(collErr.Error())
			errs = append(errs, collErr)
		}
	}

	for _, dbConfig := range b.Get().DatabaseConfigs {
		dbConfig.DB, dbErr = dbConfig.DatabaseService()
		if dbErr != nil {
			log.Error(dbErr.Error())
			errs = append(errs, dbErr)
		}
	}

	return b.Get(), errs
}

func (c *Config) DatabaseConfig(name string) (*DatabaseConfig, error) {

	if database, ok := c.DatabaseConfigs[name]; ok {
		return database, nil
	}
	return nil, fmt.Errorf("database config: %v not found", name)

}

func (c *Config) ServiceConfig(name string) (*ServiceConfig, error) {

	if service, ok := c.ServiceConfigs[name]; ok {
		return service, nil
	}
	return nil, fmt.Errorf("service config: %v not found", name)
}

func (c *Config) CrawlConfig(name string) (*CrawlConfig, error) {

	if crawler, ok := c.CrawlerConfigs[name]; ok {
		return crawler, nil
	}
	return nil, fmt.Errorf("crawler config: %v not found", name)
}
