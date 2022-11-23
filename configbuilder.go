package config_yaml

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"strconv"

	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

type configBuilder interface {
	ClientInit() ClientConfigFunc
	Load(string) (*os.File, error)
	Read(io.Reader) error
	Get() *Config
	Path() string
}

type builder struct {
	config     *Config
	configPath string
}

func (b *builder) Get() *Config {
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
	config, errs := initialConfig(configData)
	if errs != nil {
		return errs
	}

	b.config = config
	return nil
}

func initialConfig(data io.Reader) (*Config, error) {
	buf := new(bytes.Buffer)
	_, buffErr := io.Copy(buf, data)
	if buffErr != nil {
		return nil, fmt.Errorf("error reading config data; err: %v", buffErr.Error())
	}
	c := &Config{}
	err := yaml.Unmarshal(buf.Bytes(), &c)

	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("error unmarshalling config data; err: %v", err.Error())
	}
	c.Hash = fmt.Sprintf("%x", md5.Sum(buf.Bytes()))

	return c, nil
}

func (b *builder) ClientInit() ClientConfigFunc {
	buildClientFn := func(config ClientConfig) *http.Client {
		client, errs := createHTTPClient(config)
		if errs != nil && len(errs) > 0 {
			for _, err := range errs {
				log.Panic(err.Error())
			}
		}
		return client
	}
	return buildClientFn
}

func createHTTPClient(config ClientConfig) (*http.Client, []error) {
	var errs []error
	timeout, err := strconv.Atoi(config.Timeout.Value)
	if err != nil {
		errs = append(errs, err)
	}
	idleConnTimeout, err := strconv.Atoi(config.IdleConnTimeout.Value)
	if err != nil {
		errs = append(errs, err)
	}
	maxIdleConnsPerHost, err := strconv.Atoi(config.MaxIdleConsPerHost.Value)
	if err != nil {
		errs = append(errs, err)
	}
	maxConnsPerHost, err := strconv.Atoi(config.MaxConsPerHost.Value)
	if err != nil {
		errs = append(errs, err)
	}

	if errs != nil && len(errs) > 0 {
		return nil, errs
	}

	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			IdleConnTimeout:     time.Duration(idleConnTimeout) * time.Second,
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			MaxConnsPerHost:     maxConnsPerHost,
			DisableCompression:  false,
		},
	}, nil
}
