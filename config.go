package config_yaml

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
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
	log.Infoln(configPath)

	config, errs := buildFromPath(&builder{}, configPath)

	if len(errs) > 0 || config == nil {
		for _, err := range errs {
			log.Panicf("Config error: %v\n", err.Error())
		}
		if config == nil {
			log.Panicln("Config file not found")
		}
		log.Panicln("Exiting: Failed to load the config file")
	}
	return config
}

func buildFromPath(b configBuilder, configPath string) (*Config, []error) {
	var errs []error
	var dbErr, collErr, err error
	// load the yaml config file
	file, err := b.Load(configPath)
	if err != nil {
		return nil, []error{err}
	}

	defer func(configFile *os.File) {
		if closeErr := configFile.Close(); err != nil {
			log.Errorln(closeErr.Error())
		}
	}(file)
	// read the file and build the initial config
	if err = b.Read(file); err != nil {
		return nil, []error{err}
	}

	clientFn := b.ClientFn()
	// initialize clients and merge override configurations
	for _, service := range b.Config().Services {
		service.Client = clientFn(service.MergedComponentConfigs().Client)
	}
	// initialize the Collector for each crawler
	for _, crawler := range b.Config().Crawlers {
		crawler.Collector, collErr = crawler.CrawlerCollector()
		if collErr != nil {
			log.Error(collErr.Error())
			errs = append(errs, collErr)
		}
	}
	// initialize each database connection
	for _, database := range b.Config().Databases {
		database.DB, dbErr = database.DatabaseService()
		if dbErr != nil {
			log.Error(dbErr.Error())
			errs = append(errs, dbErr)
		}
	}

	return b.Config(), errs
}

// Database returns an initialized database configuration by name
func (c *Config) Database(name string) (*DatabaseConfig, error) {
	if database, ok := c.Databases[name]; ok {
		return database, nil
	}
	// return error if the database not found in config
	return nil, fmt.Errorf("database config: %v not found", name)

}

// Service returns an initialized service configuration by name
func (c *Config) Service(name string) (*ServiceConfig, error) {
	if service, ok := c.Services[name]; ok {
		return service, nil
	}
	// return error if the service not found in config
	return nil, fmt.Errorf("service config: %v not found", name)
}

// Crawler returns an initialized crawler configuration by name
func (c *Config) Crawler(name string) (*CrawlerConfig, error) {
	if crawler, ok := c.Crawlers[name]; ok {
		return crawler, nil
	}
	// return error if the crawler not found in config
	return nil, fmt.Errorf("crawler config: %v not found", name)
}
