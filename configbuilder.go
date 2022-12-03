package config_yaml

import (
	"bytes"
	"crypto/md5"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
)

type configBuilder interface {
	InitClient() ClientConfigFunc
	Load(string) (*os.File, error)
	Read(io.Reader) error
	Get() *Config
	Path() string
}

type builder struct {
	config     *Config
	configPath string
}

func (b *builder) InitClient() ClientConfigFunc {
	return func(cc ClientConfig) *http.Client {
		client, errs := createHTTPClient(cc)
		if len(errs) > 0 {
			for _, err := range errs {
				panic(err)
			}
		}
		return client
	}
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
