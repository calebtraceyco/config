package config

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strconv"
)

type builder struct {
	config     *Config
	configPath string
}

func (b *builder) newConfig(configPath string) (*Config, []error) {
	var errs []error
	if file, loadErrs := b.loadConfig(configPath); loadErrs != nil {
		return nil, loadErrs
	} else {
		if readErr := b.Read(file); readErr != nil {
			return nil, []error{fmt.Errorf("newConfig: failed to read file: %v; error: %w", file.Name(), readErr)}
		}
	}

	for _, service := range b.config.Services {
		service.setClient(service.mergedComponents().Client)
	}

	var collErr error
	var dbErrs []error
	// initialize the Collector for each crawler
	for _, crawler := range b.config.Crawlers {
		if crawler.Collector, collErr = crawler.collector(); collErr != nil {
			errs = appendAndLog(fmt.Errorf("newConfig: failed to build crawler; error: %w", collErr), errs)
		}
	}
	for _, database := range b.config.Databases {
		if database.Pool, dbErrs = database.DatabaseService(); dbErrs != nil {
			errs = appendAndLog(fmt.Errorf("newConfig: %v", dbErrs), errs)
		}
	}

	return b.config, errs
}

func (b *builder) loadConfig(configPath string) (*os.File, []error) {
	if file, loadErr := b.Load(configPath); loadErr != nil {
		return nil, []error{fmt.Errorf("newConfig: %w", loadErr)}
	} else {
		return file, nil
	}
}

func (b *builder) Load(path string) (file *os.File, err error) {
	log.Tracef("Loading config: %v", path)
	b.configPath = path
	if file, err = os.Open(path); err != nil {
		return nil, fmt.Errorf("Load: failed to open config file %v; %w", path, err)
	}
	return file, err
}

func (b *builder) Read(configData io.Reader) (err error) {
	if b.config, err = initialConfig(configData); err != nil {
		return err
	}

	if mergeErr := mergeServiceComponentConfigs(b.config); mergeErr != nil {
		return fmt.Errorf("Read: failed to merge component configs, error: %w", mergeErr)
	}
	return nil
}

func initialConfig(data io.Reader) (*Config, error) {
	buf := new(bytes.Buffer)

	if _, buffErr := io.Copy(buf, data); buffErr != nil {
		return nil, fmt.Errorf("initialConfig: failed to read config data; err: %w", buffErr)
	}

	config := new(Config)

	if err := yaml.Unmarshal(buf.Bytes(), &config); err != nil {
		return nil, fmt.Errorf("initialConfig: failed unmarshalling config data; err: %w", err)
	}
	config.Hash = fmt.Sprintf("%x", md5.Sum(buf.Bytes()))

	return config, nil
}

func mergeServiceComponentConfigs(c *Config) error {
	for i, service := range c.Services {
		if mergeErr := mergeConfigs(&service.ComponentConfigOverrides, &c.ComponentConfigs, &service.mergedComponentConfigs); mergeErr != nil {
			return fmt.Errorf("mergeServiceComponentConfigs: failed to merging component config: %v; error %w", i, mergeErr)
		}
	}
	return nil
}

func mergeConfigs(override *ComponentConfigs, defaultC *ComponentConfigs, mergedC *ComponentConfigs) error {
	if mergedC == nil {
		return fmt.Errorf("mergeConfigs: nil pointer passed for merged components")
	}
	if err := copier.Copy(mergedC, override); err != nil && override != nil {
		return fmt.Errorf("mergeConfigs: failed to copy config overrides; error: %w", err)
	}
	if err := mergo.Merge(mergedC, defaultC); err != nil && defaultC != nil {
		return fmt.Errorf("mergeConfigs: failed to copy config defaults; error: %w", err)
	}
	return nil
}

func toInt(str string) int {
	res := 0
	var err error
	if str != "" {
		if res, err = strconv.Atoi(str); err != nil {
			log.Errorf("toInt: failed to convert '%s' to int; error: %v", str, err)
		}
	}

	return res
}
