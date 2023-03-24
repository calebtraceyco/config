package config

import (
	"net/http"
	"time"
)

type ClientConfig struct {
	Timeout            string `yaml:"Timeout"`
	IdleConnTimeout    string `yaml:"IdleConnTimeout"`
	MaxIdleConsPerHost string `yaml:"MaxIdleConsPerHost"`
	MaxConsPerHost     string `yaml:"MaxConsPerHost"`
	DisableCompression configFlag
	InsecureSkipVerify configFlag
}

func httpClient(cc ClientConfig) *http.Client {
	disableCompression := false
	timeout := 15

	if cc.DisableCompression == True {
		disableCompression = true
	}

	if to := toInt(cc.Timeout); to != 0 {
		timeout = to
	}

	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			IdleConnTimeout:     time.Duration(toInt(cc.IdleConnTimeout)) * time.Second,
			MaxIdleConnsPerHost: toInt(cc.MaxIdleConsPerHost),
			MaxConnsPerHost:     toInt(cc.MaxConsPerHost),
			DisableCompression:  disableCompression,
		},
	}
}
