package config

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"time"
)

type DatabaseConfig struct {
	Name                    string           `yaml:"Name" json:"Name,omitempty"`
	Database                string           `yaml:"Database" json:"Database,omitempty"`
	Host                    string           `yaml:"Host" json:"Host,omitempty"`
	Port                    int              `yaml:"Port" json:"Port,omitempty"`
	Server                  string           `yaml:"Server" json:"Server,omitempty"`
	Username                string           `yaml:"Username" json:"Username,omitempty"`
	Password                string           `yaml:"Password" json:"Password,omitempty"`
	AuthRequired            bool             `yaml:"AuthRequired" json:"AuthRequired,omitempty"`
	AuthEnvironmentVariable string           `yaml:"AuthEnvironmentVariable" json:"AuthEnvironmentVariable,omitempty"`
	RawConnectionString     string           `yaml:"RawConnectionString" json:"RawConnectionString,omitempty"`
	Scheme                  string           `yaml:"Scheme" json:"Scheme,omitempty"`
	MaxConnections          int              `yaml:"MaxConnections" json:"MaxConnections,omitempty"`
	MaxIdleConnections      int              `yaml:"MaxIdleConnections" json:"MaxIdleConnections,omitempty"`
	DB                      *sql.DB          `yaml:"-" json:"DB,omitempty"`
	Pool                    *pgxpool.Pool    `yaml:"-" json:"Pool,omitempty"`
	ComponentConfigs        ComponentConfigs `json:"ComponentConfigs"`
}

type DatabaseConfigMap map[string]*DatabaseConfig

func (dbc *DatabaseConfig) DbComponentConfigs() ComponentConfigs {
	return dbc.ComponentConfigs
}

func (dbc *DatabaseConfig) DatabaseService() (pool *pgxpool.Pool, errs []error) {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	dbc.mapAuthentication()
	if errs = dbc.validate(); errs != nil {
		return pool, errs
	}

	connectionStr := ""

	switch dbc.Scheme {
	case Postgres:
		u := &url.URL{
			Scheme: dbc.Scheme,
			User:   url.UserPassword(dbc.Username, dbc.Password),
			Host:   dbc.Server,
		}
		connectionStr = u.String()
	default:
		connectionStr = dbc.RawConnectionString
	}

	if cfg, cfgErr := pgxpool.ParseConfig(connectionStr); cfgErr != nil {
		log.Errorf("DatabaseService: postgres connection failed: %s; \nerror: %v", connectionStr, cfgErr)
		return nil, []error{cfgErr}

	} else {

		if pool, err = pgxpool.NewWithConfig(ctx, cfg); err != nil {
			log.Errorf("DatabaseService: failed to establish connection pool: %v", err)
			return nil, []error{err}
		}
		if pingErr := pool.Ping(ctx); pingErr != nil {
			return nil, []error{fmt.Errorf("DatabaseService: unable to ping database; err: %v", pingErr)}
		}
	}
	log.Tracef("Database connection successful: '%s'\n", dbc.Database)

	return pool, nil
}

func (dbc *DatabaseConfig) mapAuthentication() {
	if dbc.Password == "" && dbc.AuthRequired && dbc.AuthEnvironmentVariable != "" {
		dbc.Password = os.Getenv(dbc.AuthEnvironmentVariable)
		if dbc.RawConnectionString == "" {
			dbc.RawConnectionString = fmt.Sprintf(dbc.RawConnectionString, dbc.Password, dbc.Database)
		}
	}
}

func (dbc *DatabaseConfig) validate() (errs []error) {
	switch {
	case dbc.AuthEnvironmentVariable == "":
		errs = append(errs, validationError("AuthEnvironmentVariable", "DatabaseConfig"))
	case dbc.Server == "":
		errs = append(errs, validationError("Server", "DatabaseConfig"))
	case dbc.Username == "":
		errs = append(errs, validationError("Username", "DatabaseConfig"))
	case dbc.Database == "":
		errs = append(errs, validationError("Database", "DatabaseConfig"))
	}

	return errs
}

func validationError(field, component string) error {
	return fmt.Errorf("component: %s - '%s' is a required field", component, field)
}

func (m *DatabaseConfigMap) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.SequenceNode {
		return fmt.Errorf("UnmarshalYAML: expected a sequence, got %v", value.Kind)
	}

	configs := make(DatabaseConfigMap, len(*m))
	for _, item := range value.Content {
		config := new(DatabaseConfig)
		if err := item.Decode(&config); err != nil {
			log.Errorf("UnmarshalYAML - decode error: %v", err)
			return err
		}

		configs[config.Name] = config
	}

	*m = configs
	return nil
}

const Postgres = "postgres"
