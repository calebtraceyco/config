package config_yaml

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	Name                         yaml.Node `yaml:"Name"`
	URL                          yaml.Node `yaml:"URL"`
	ApiKeyEnvironmentVariable    yaml.Node `yaml:"ApiKeyEnvironmentVariable"`
	PublicKeyEnvironmentVariable yaml.Node `yaml:"PublicKeyEnvironmentVariable"`
	componentConfigs             ComponentConfigs
	//HTTPClient                   *http.Client
}

type ServiceConfigMap map[string]*ServiceConfig

func (s *ServiceConfig) SvcComponentConfigs() ComponentConfigs {
	return s.componentConfigs
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
