package config

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

func httpClient(cc ClientConfig) *http.Client {
	disableCompression := false
	timeout := 15

	if cc.DisableCompression == True {
		disableCompression = true
	}

	if to := toInt(cc.Timeout.Value); to != 0 {
		timeout = to
	}

	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			IdleConnTimeout:     time.Duration(toInt(cc.IdleConnTimeout.Value)) * time.Second,
			MaxIdleConnsPerHost: toInt(cc.MaxIdleConsPerHost.Value),
			MaxConnsPerHost:     toInt(cc.MaxConsPerHost.Value),
			DisableCompression:  disableCompression,
		},
	}
}
