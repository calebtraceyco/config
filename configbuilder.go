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

type configBuilder interface {
	ClientFn() ClientConfigFromFn
	Load(string) (*os.File, error)
	Read(io.Reader) error
	Config() *Config
	Path() string
}

type builder struct {
	config     *Config
	configPath string
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
		return nil, fmt.Errorf("failed to open config file %v; %v", path, err.Error())
	}

	return file, err
}

func (b *builder) Read(configData io.Reader) error {
	config, err := initialConfig(configData)
	if err != nil {
		return err
	}

	if mergeErr := mergeServiceComponentConfigs(config); mergeErr != nil {
		return fmt.Errorf("error merging component configs, err: %w", mergeErr)
	}
	b.config = config
	return nil
}

func initialConfig(data io.Reader) (*Config, error) {
	buf := new(bytes.Buffer)

	if _, buffErr := io.Copy(buf, data); buffErr != nil {
		return nil, fmt.Errorf("error reading config data; err: %v", buffErr.Error())
	}

	config := &Config{}
	if err := yaml.Unmarshal(buf.Bytes(), &config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config data; err: %v", err.Error())
	}
	config.Hash = fmt.Sprintf("%x", md5.Sum(buf.Bytes()))

	return config, nil
}

func mergeServiceComponentConfigs(c *Config) error {
	componentConfigs := c.ComponentConfigs
	for i, service := range c.Services {
		err := mergeConfigs(&service.ComponentConfigOverrides, &componentConfigs, &service.mergedComponentConfigs)
		if err != nil {
			return fmt.Errorf("error merging component config: %v, err %w", i, err)
		}
	}
	return nil
}

func mergeConfigs(override *ComponentConfigs, defaultC *ComponentConfigs, mergedC *ComponentConfigs) error {
	if mergedC == nil {
		return errors.New("nil pointer passed for mergedC")
	}

	if override != nil {
		if err := copier.Copy(mergedC, override); err != nil {
			return err
		}
	}

	if defaultC != nil {
		if err := mergo.Merge(mergedC, defaultC); err != nil {
			return err
		}
	}
	return nil
}

func toInt(str string) int {
	res, err := strconv.Atoi(str)
	if err != nil {
		log.Errorf("strconv error for %s, err: %v", str, err)
	}
	return res
}
