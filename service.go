package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/http"
)

type ServiceConfig struct {
	Name                         string `yaml:"Name"`
	URL                          string `yaml:"URL"`
	ApiKeyEnvironmentVariable    string `yaml:"ApiKeyEnvironmentVariable"`
	PublicKeyEnvironmentVariable string `yaml:"PublicKeyEnvironmentVariable"`
	ComponentConfigOverrides     ComponentConfigs
	Endpoints                    EndpointMap

	mergedComponentConfigs ComponentConfigs

	Client *http.Client `json:"-"`
}

func (s *ServiceConfig) setClient(cc ClientConfig) {
	s.Client = httpClient(cc)
}

type Endpoint struct {
	Name string
	Path string
}

type EndpointMap map[string]*Endpoint

type ServiceConfigMap map[string]*ServiceConfig

func (s *ServiceConfig) mergedComponents() ComponentConfigs {
	return s.mergedComponentConfigs
}

func (m *ServiceConfigMap) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.SequenceNode {
		return fmt.Errorf("UnmarshalYAML: expected a sequence, got %v", value.Kind)
	}

	configs := make(ServiceConfigMap, len(*m))
	for _, item := range value.Content {
		config := new(ServiceConfig)
		if err := item.Decode(&config); err != nil {
			log.Errorf("UnmarshalYAML - decode error: %v", err)
			return err
		}

		configs[config.Name] = config
	}

	*m = configs
	return nil
}
