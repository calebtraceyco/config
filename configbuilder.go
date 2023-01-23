package config_yaml

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"strconv"
)

type builder struct {
	config     *Config
	configPath string
}

func (b *builder) NewConfig(configPath string) (config *Config, errs []error) {
	file, loadErr := b.Load(configPath)
	if loadErr != nil {
		return nil, []error{fmt.Errorf("NewConfig: %w", loadErr)}
	}
	defer func(configFile *os.File) {
		if closeErr := configFile.Close(); closeErr != nil {
			log.Error(fmt.Errorf("NewConfig: failed to close file: %v; error: %w", file.Name(), closeErr))
		}
	}(file)

	if readErr := b.Read(file); readErr != nil {
		return nil, []error{fmt.Errorf("NewConfig: failed to read file: %v; error: %w", file.Name(), readErr)}
	}

	//clientFn := b.ClientFn()

	for _, service := range b.Config().Services {
		service.SetClient(service.MergedComponentConfigs().Client)
		//service.Client = clientFn(service.MergedComponentConfigs().Client)
	}
	var collErr, dbErr error
	// initialize the Collector for each crawler
	for _, crawler := range b.Config().Crawlers {
		crawler.Collector, collErr = crawler.CrawlerCollector()
		if collErr != nil {
			collErr = fmt.Errorf("NewConfig: failed to build crawler; error: %w", collErr)
			log.Error(collErr)
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

	return b.config, errs
}

type ClientConfigFromFn func(ClientConfig) *http.Client

func (b *builder) ClientFn() ClientConfigFromFn {
	buildClientFn := func(cc ClientConfig) *http.Client {
		return createHTTPClient(cc)
	}
	return buildClientFn
}

func (b *builder) Config() *Config {
	return b.config
}

func (b *builder) Path() string {
	return b.configPath
}

func (b *builder) Load(path string) (*os.File, error) {
	log.Tracef("Loading config: %v", path)
	b.configPath = path

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Load: failed to open config file %v; %w", path, err)
	}

	return file, err
}

func (b *builder) Read(configData io.Reader) (err error) {
	b.config, err = initialConfig(configData)
	if err != nil {
		return err
	}

	if mergeErr := mergeServiceComponentConfigs(b.config); mergeErr != nil {
		return fmt.Errorf("error merging component configs, err: %w", mergeErr)
	}
	return nil
}

func initialConfig(data io.Reader) (*Config, error) {
	buf := new(bytes.Buffer)

	if _, buffErr := io.Copy(buf, data); buffErr != nil {
		return nil, fmt.Errorf("error reading config data; err: %v", buffErr.Error())
	}

	config := new(Config)

	if err := yaml.Unmarshal(buf.Bytes(), &config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config data; err: %v", err.Error())
	}
	config.Hash = fmt.Sprintf("%x", md5.Sum(buf.Bytes()))

	return config, nil
}

func mergeServiceComponentConfigs(c *Config) error {
	componentConfigs := c.ComponentConfigs
	for i, service := range c.Services {
		mergeErr := mergeConfigs(&service.ComponentConfigOverrides, &componentConfigs, &service.mergedComponentConfigs)
		if mergeErr != nil {
			return fmt.Errorf("mergeServiceComponentConfigs: failed to merging component config: %v; error %w", i, mergeErr)
		}
	}
	return nil
}

func mergeConfigs(override *ComponentConfigs, defaultC *ComponentConfigs, mergedC *ComponentConfigs) error {
	if mergedC == nil {
		return errors.New("mergeConfigs: nil pointer passed for merged components")
	}

	if override != nil {
		if err := copier.Copy(mergedC, override); err != nil {
			return fmt.Errorf("mergeConfigs: failed to copy config overrides; error: %w", err)
		}
	}

	if defaultC != nil {
		if err := mergo.Merge(mergedC, defaultC); err != nil {
			return fmt.Errorf("mergeConfigs: failed to copy config defaults; error: %w", err)
		}
	}
	return nil
}

func toInt(str string) int {
	res, err := strconv.Atoi(str)
	if err != nil {
		log.Errorf("toInt: failed to convert '%s' to int; error: %v", str, err)
	}
	return res
}
