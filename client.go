package config_yaml

import (
	"gopkg.in/yaml.v3"
	"net/http"
	"time"
)

type ClientConfig struct {
	Timeout            yaml.Node `yaml:"Timeout"`
	IdleConnTimeout    yaml.Node `yaml:"IdleConnTimeout"`
	MaxIdleConsPerHost yaml.Node `yaml:"MaxIdleConsPerHost"`
	MaxConsPerHost     yaml.Node `yaml:"MaxConsPerHost"`
	DisableCompression configFlag
	InsecureSkipVerify configFlag
}

func createHTTPClient(cc ClientConfig) *http.Client {
	disableCompression := false

	if cc.DisableCompression == True {
		disableCompression = true
	}

	return &http.Client{
		Timeout: time.Duration(toInt(cc.Timeout.Value)) * time.Second,
		Transport: &http.Transport{
			IdleConnTimeout:     time.Duration(toInt(cc.IdleConnTimeout.Value)) * time.Second,
			MaxIdleConnsPerHost: toInt(cc.MaxIdleConsPerHost.Value),
			MaxConnsPerHost:     toInt(cc.MaxConsPerHost.Value),
			DisableCompression:  disableCompression,
		},
	}
}
