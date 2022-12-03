package config_yaml

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/http"
	"strconv"
	"time"
)

type ClientConfig struct {
	Timeout            yaml.Node `yaml:"Timeout"`
	IdleConnTimeout    yaml.Node `yaml:"IdleConnTimeout"`
	MaxIdleConsPerHost yaml.Node `yaml:"MaxIdleConsPerHost"`
	MaxConsPerHost     yaml.Node `yaml:"MaxConsPerHost"`
}

type ClientConfigFunc func(ClientConfig) *http.Client

func (c ClientConfig) Client() *http.Client {
	client, errs := createHTTPClient(c)
	if errs != nil && len(errs) > 0 {
		for _, err := range errs {
			log.Panic(err.Error())
		}
	}
	return client
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
