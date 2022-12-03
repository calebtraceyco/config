package config_yaml

import (
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
	DisableCompression yaml.Node `yaml:"DisableCompression"`
	InsecureSkipVerify yaml.Node `yaml:"InsecureSkipVerify"`
}

func createHTTPClient(config ClientConfig) *http.Client {
	timeout, _ := strconv.Atoi(config.Timeout.Value)
	idleConnTimeout, _ := strconv.Atoi(config.IdleConnTimeout.Value)
	maxIdleConnsPerHost, _ := strconv.Atoi(config.MaxIdleConsPerHost.Value)
	maxConnsPerHost, _ := strconv.Atoi(config.MaxConsPerHost.Value)

	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			IdleConnTimeout:     time.Duration(idleConnTimeout) * time.Second,
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			MaxConnsPerHost:     maxConnsPerHost,
			DisableCompression:  false,
		},
	}
}
