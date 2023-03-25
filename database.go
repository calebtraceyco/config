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
	Name                    string        `yaml:"Name"`
	Database                string        `yaml:"Database"`
	Host                    string        `yaml:"Host"`
	Port                    int           `yaml:"Port"`
	Server                  string        `yaml:"Server"`
	Username                string        `yaml:"Username"`
	Password                string        `yaml:"Password"`
	AuthRequired            bool          `yaml:"AuthRequired"`
	AuthEnvironmentVariable string        `yaml:"AuthEnvironmentVariable"`
	RawConnectionString     string        `yaml:"RawConnectionString"`
	Scheme                  string        `yaml:"Scheme"`
	MaxConnections          int           `yaml:"MaxConnections"`
	MaxIdleConnections      int           `yaml:"MaxIdleConnections"`
	DB                      *sql.DB       `yaml:"-"`
	Pool                    *pgxpool.Pool `yaml:"-"`
	componentConfigs        ComponentConfigs
}

type DatabaseConfigMap map[string]*DatabaseConfig

func (dbc *DatabaseConfig) DbComponentConfigs() ComponentConfigs {
	return dbc.componentConfigs
}

func (dbc *DatabaseConfig) DatabaseService() (pool *pgxpool.Pool, errs []error) {
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

		var poolErr error
		if pool, poolErr = pgxpool.NewWithConfig(ctx, cfg); poolErr != nil {
			log.Errorf("DatabaseService: failed to establish connection pool: %v", poolErr)
			return nil, []error{poolErr}
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
