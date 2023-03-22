package config_yaml

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type configFlag int

const (
	Unset configFlag = iota
	True
	False
)

type Config struct {
	AppName yaml.Node `yaml:"AppName"`
	Env     yaml.Node `yaml:"Env"`
	Port    yaml.Node `yaml:"Port"`

	ComponentConfigs ComponentConfigs  `yaml:"ComponentConfigs"`
	Databases        DatabaseConfigMap `yaml:"Databases"`
	Services         ServiceConfigMap  `yaml:"Services"`
	Crawlers         CrawlConfigMap    `yaml:"Crawlers"`

	Hash string `yaml:"Hash"`
}

type ComponentConfigs struct {
	//TODO add logging
	Client ClientConfig
}

func New(configPath string) *Config {
	log.Infof("config file: %s", configPath)

	config, errs := new(builder).newConfig(configPath)
	if len(errs) > 0 || config == nil {
		for _, err := range errs {
			log.Panicf("configuration error: %v\n", err.Error())
		}
		if config == nil {
			log.Panicln("configuration file not found")
		}
		log.Panicln("Exiting: Failed to load the config file")
	}
	return config
}

// Database returns an initialized database configuration by name
func (c *Config) Database(name string) (*DatabaseConfig, error) {
	if database, ok := c.Databases[name]; ok {
		return database, nil
	}
	// return error if the database not found in config
	return nil, fmt.Errorf("Database: %w", errors.New(fmt.Sprintf("%s not found", name)))
}

// Service returns an initialized service configuration by name
func (c *Config) Service(name string) (*ServiceConfig, error) {
	if service, ok := c.Services[name]; ok {
		return service, nil
	}
	// return error if the service not found in config
	return nil, fmt.Errorf("Service: %w", errors.New(fmt.Sprintf("%s not found", name)))
}

// Crawler returns an initialized crawler configuration by name
func (c *Config) Crawler(name string) (*Scraper, error) {
	if crawler, ok := c.Crawlers[name]; ok {
		return crawler, nil
	}
	// return error if the crawler not found in config
	return nil, fmt.Errorf("Crawler: %w", errors.New(fmt.Sprintf("%s not found", name)))
}

func appendAndLog(err error, errs []error) []error {
	log.Error(err)
	return append(errs, err)
}
