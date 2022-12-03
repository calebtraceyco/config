package config_yaml

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
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

	clientFn := b.ClientFn()

	for _, service := range b.Config().Services {
		service.Client = clientFn(service.MergedComponentConfigs().Client)
	}

	for _, crawler := range b.Config().Crawlers {
		crawler.Collector, collErr = crawler.CrawlerService()
		if collErr != nil {
			log.Error(collErr.Error())
			errs = append(errs, collErr)
		}
	}

	for _, database := range b.Config().Databases {
		database.DB, dbErr = database.DatabaseService()
		if dbErr != nil {
			log.Error(dbErr.Error())
			errs = append(errs, dbErr)
		}
	}
	return b.Config(), errs
}

func (c *Config) Database(name string) (*DatabaseConfig, error) {
	if database, ok := c.Databases[name]; ok {
		return database, nil
	}
	// return error if the database not found in config
	return nil, fmt.Errorf("database config: %v not found", name)

}

func (c *Config) Service(name string) (*ServiceConfig, error) {
	if service, ok := c.Services[name]; ok {
		return service, nil
	}
	// return error if the service not found in config
	return nil, fmt.Errorf("service config: %v not found", name)
}

func (c *Config) Crawler(name string) (*CrawlerConfig, error) {
	if crawler, ok := c.Crawlers[name]; ok {
		return crawler, nil
	}
	// return error if the crawler not found in config
	return nil, fmt.Errorf("crawler config: %v not found", name)
}
