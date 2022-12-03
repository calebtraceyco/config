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
	//HTTPClient                   *http.Client
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

func (scm *ServiceConfigMap) UnmarshalYAML(node *yaml.Node) error {
	*scm = ServiceConfigMap{}
	var services []ServiceConfig

	if decodeErr := node.Decode(&services); decodeErr != nil {
		return fmt.Errorf("decode error: %v", decodeErr.Error())
	}

	for _, service := range services {
		var serviceString string
		serviceCopy := service
		serviceErr := service.Name.Decode(&serviceString)
		if serviceErr != nil {
			return fmt.Errorf("decode error: %v", serviceErr.Error())
		}
		(*scm)[serviceString] = &serviceCopy
	}

	return nil
}
