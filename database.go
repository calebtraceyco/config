package config_yaml

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/url"
	"strconv"
)

type DatabaseConfig struct {
	Name               yaml.Node `yaml:"Name"`
	Database           yaml.Node `yaml:"Database"`
	Server             yaml.Node `yaml:"Server"`
	Username           yaml.Node `yaml:"Username"`
	Password           yaml.Node `yaml:"Password"`
	Scheme             yaml.Node `yaml:"Scheme"`
	MaxConnections     yaml.Node `yaml:"MaxConnections"`
	MaxIdleConnections yaml.Node `yaml:"MaxIdleConnections"`
	DB                 *sql.DB
	componentConfigs   ComponentConfigs
}

type DatabaseConfigMap map[string]*DatabaseConfig

func (s *DatabaseConfig) DbComponentConfigs() ComponentConfigs {
	return s.componentConfigs
}

func InitDbService(dbc *DatabaseConfig, appName string) (*sql.DB, error) {
	if dbc.Password.Value == "" || dbc.Server.Value == "" || dbc.Username.Value == "" || dbc.Database.Value == "" {
		log.Errorf("Missing DB config feilds for %v", dbc)
	}

	query := url.Values{}
	u := &url.URL{
		Scheme:   dbc.Scheme.Value,
		User:     url.UserPassword(dbc.Username.Value, dbc.Password.Value),
		Host:     dbc.Server.Value,
		RawQuery: query.Encode(),
	}
	connectionString := u.String() + "/" + dbc.Database.Value + "?sslmode=disable"

	db, err := sql.Open(dbc.Scheme.Value, connectionString)
	if err != nil {
		log.Errorf("failed to open postgres connection; err: %v", err.Error())
		return nil, fmt.Errorf("cannot open connection to the database")
	}
	if dbc.MaxConnections.Value != "" {
		mc, _ := strconv.Atoi(dbc.MaxConnections.Value)
		db.SetMaxOpenConns(mc)
	}
	if dbc.MaxIdleConnections.Value != "" {
		mic, _ := strconv.Atoi(dbc.MaxIdleConnections.Value)
		db.SetMaxIdleConns(mic)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		return nil, fmt.Errorf("unable to ping database; err: %v", pingErr.Error())
	}

	return db, nil
}

func (dcm *DatabaseConfigMap) UnmarshalYAML(node *yaml.Node) error {
	*dcm = DatabaseConfigMap{}
	var databases []DatabaseConfig

	if decodeErr := node.Decode(&databases); decodeErr != nil {
		return fmt.Errorf("decode error: %v", decodeErr.Error())
	}

	for _, db := range databases {
		var dbString string
		dbCopy := db
		serviceErr := db.Name.Decode(&dbString)
		if serviceErr != nil {
			return fmt.Errorf("decode error: %v", serviceErr.Error())
		}
		(*dcm)[dbString] = &dbCopy
	}

	return nil
}
