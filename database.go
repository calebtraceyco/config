package config_yaml

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"strings"
)

type DatabaseConfig struct {
	Name                yaml.Node `yaml:"Name"`
	Database            yaml.Node `yaml:"Database"`
	Server              yaml.Node `yaml:"Server"`
	Username            yaml.Node `yaml:"Username"`
	PasswordEnvVariable yaml.Node `yaml:"PasswordEnvVariable"`
	Scheme              yaml.Node `yaml:"Scheme"`
	MaxConnections      yaml.Node `yaml:"MaxConnections"`
	MaxIdleConnections  yaml.Node `yaml:"MaxIdleConnections"`
	DB                  *sql.DB
	componentConfigs    ComponentConfigs
}

func (dbc *DatabaseConfig) SetDatabase(db *sql.DB) {
	dbc.DB = db
}

type DatabaseConfigMap map[string]*DatabaseConfig

func (dbc *DatabaseConfig) DbComponentConfigs() ComponentConfigs {
	return dbc.componentConfigs
}

func (dbc *DatabaseConfig) DatabaseService() (*sql.DB, error) {

	if dbc.PasswordEnvVariable.Value == "" || dbc.Server.Value == "" || dbc.Username.Value == "" || dbc.Database.Value == "" {
		log.Errorf("Missing DB config feilds for %v", dbc)
	}

	u := &url.URL{
		Scheme:   dbc.Scheme.Value,
		User:     url.UserPassword(dbc.Username.Value, os.Getenv(dbc.PasswordEnvVariable.Value)),
		Host:     dbc.Server.Value,
		RawQuery: url.Values{}.Encode(),
	}

	db, err := sql.Open(
		dbc.Scheme.Value,
		strings.Join(
			[]string{u.String(), "/", dbc.Database.Value, "?sslmode=disable"}, ""),
	)

	if err != nil {
		err = fmt.Errorf("DatabaseService: failed to open postgres connection; error: %w", err)
		log.Error(err)
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
