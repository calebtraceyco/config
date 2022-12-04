package config_yaml

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
)

type ServiceConfig struct {
	Name                         yaml.Node `yaml:"Name"`
	URL                          yaml.Node `yaml:"URL"`
	ApiKeyEnvironmentVariable    yaml.Node `yaml:"ApiKeyEnvironmentVariable"`
	PublicKeyEnvironmentVariable yaml.Node `yaml:"PublicKeyEnvironmentVariable"`
	ComponentConfigOverrides     ComponentConfigs
	Endpoints                    EndpointMap

	mergedComponentConfigs ComponentConfigs

	Client *http.Client `json:"-"`
}

type Endpoint struct {
	Name string
	Path string
}

type EndpointMap map[string]*Endpoint

type ServiceConfigMap map[string]*ServiceConfig

func (s *ServiceConfig) MergedComponentConfigs() ComponentConfigs {
	return s.mergedComponentConfigs
}

func (sm *ServiceConfigMap) UnmarshalYAML(node *yaml.Node) error {
	*sm = ServiceConfigMap{}
	var services []ServiceConfig

	if decodeErr := node.Decode(&services); decodeErr != nil {
		return fmt.Errorf("decode error: %v", decodeErr.Error())
	}

	for _, service := range services {
		var serviceKey string
		serviceCopy := service
		if serviceErr := service.Name.Decode(&serviceKey); serviceErr != nil {
			return fmt.Errorf("decode error: %v", serviceErr.Error())
		}
		(*sm)[serviceKey] = &serviceCopy
	}
	return nil
}
