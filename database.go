package config_yaml

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
)

type DatabaseConfig struct {
	Name                    yaml.Node `yaml:"Name"`
	Database                yaml.Node `yaml:"Database"`
	Server                  yaml.Node `yaml:"Server"`
	Username                yaml.Node `yaml:"Username"`
	Password                yaml.Node `yaml:"Password"`
	AuthRequired            yaml.Node `yaml:"AuthRequired"`
	AuthEnvironmentVariable yaml.Node `yaml:"AuthEnvironmentVariable"`
	RawConnectionString     yaml.Node `yaml:"RawConnectionString"`
	Scheme                  yaml.Node `yaml:"Scheme"`
	MaxConnections          yaml.Node `yaml:"MaxConnections"`
	MaxIdleConnections      yaml.Node `yaml:"MaxIdleConnections"`
	DB                      *sql.DB   `yaml:"-" json:"-"`
	componentConfigs        ComponentConfigs
}

type DatabaseConfigMap map[string]*DatabaseConfig

func (dbc *DatabaseConfig) DbComponentConfigs() ComponentConfigs {
	return dbc.componentConfigs
}

func (dbc *DatabaseConfig) DatabaseService() (db *sql.DB, err error) {
	dbc.mapAuthentication()
	dbc.validate()

	connectionStr := ""

	switch dbc.Scheme.Value {
	case Postgres:
		u := &url.URL{
			Scheme: dbc.Scheme.Value,
			User:   url.UserPassword(dbc.Username.Value, dbc.Password.Value),
			Host:   dbc.Server.Value,
		}
		connectionStr = u.String()
	default:
		connectionStr = dbc.RawConnectionString.Value
	}

	if db, err = sql.Open(dbc.Scheme.Value, connectionStr); err != nil {
		log.Errorf("DatabaseService: failed postgres connection: %s; \nerror: %v", connectionStr, err)
		return nil, err
	}

	if dbc.MaxConnections.Value != "" {
		db.SetMaxOpenConns(toInt(dbc.MaxConnections.Value))
	}
	if dbc.MaxIdleConnections.Value != "" {
		db.SetMaxIdleConns(toInt(dbc.MaxIdleConnections.Value))
	}

	if pingErr := db.Ping(); pingErr != nil {
		return nil, fmt.Errorf("unable to ping database; err: %v", pingErr.Error())
	}

	return db, nil
}

func (dbc *DatabaseConfig) mapAuthentication() {
	if dbc.Password.Value == "" && dbc.AuthRequired.Value == "true" && dbc.AuthEnvironmentVariable.Value != "" {
		dbc.Password = yaml.Node{Value: os.Getenv(dbc.AuthEnvironmentVariable.Value)}
		if dbc.RawConnectionString.Value != "" {
			dbc.RawConnectionString.Value = fmt.Sprintf(dbc.RawConnectionString.Value, dbc.Password.Value)
		}
	}
}

func (dbc *DatabaseConfig) validate() {
	if dbc.AuthEnvironmentVariable.Value == "" || dbc.Server.Value == "" || dbc.Username.Value == "" || dbc.Database.Value == "" {
		// TODO possibly return this as an error
		log.Errorf("Missing DB config fields for %v", dbc)
	}
}

func (dm *DatabaseConfigMap) UnmarshalYAML(node *yaml.Node) error {
	*dm = DatabaseConfigMap{}
	var databases []DatabaseConfig

	if decodeErr := node.Decode(&databases); decodeErr != nil {
		return fmt.Errorf("decode error: %v", decodeErr.Error())
	}

	for _, db := range databases {
		var databaseKey string
		dbCopy := db

		if databaseErr := db.Name.Decode(&databaseKey); databaseErr != nil {
			return fmt.Errorf("decode error: %v", databaseErr.Error())
		}
		(*dm)[databaseKey] = &dbCopy
	}
	return nil
}

const Postgres = "postgres"
