package config_yaml

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
	"strconv"
	"time"
)

type DatabaseConfig struct {
	Name                    yaml.Node     `yaml:"Name"`
	Database                yaml.Node     `yaml:"Database"`
	Host                    yaml.Node     `yaml:"Host"`
	Port                    yaml.Node     `yaml:"Port"`
	Server                  yaml.Node     `yaml:"Server"`
	Username                yaml.Node     `yaml:"Username"`
	Password                yaml.Node     `yaml:"Password"`
	AuthRequired            yaml.Node     `yaml:"AuthRequired"`
	AuthEnvironmentVariable yaml.Node     `yaml:"AuthEnvironmentVariable"`
	RawConnectionString     yaml.Node     `yaml:"RawConnectionString"`
	Scheme                  yaml.Node     `yaml:"Scheme"`
	MaxConnections          yaml.Node     `yaml:"MaxConnections"`
	MaxIdleConnections      yaml.Node     `yaml:"MaxIdleConnections"`
	DB                      *sql.DB       `yaml:"-" json:"-"`
	Pool                    *pgxpool.Pool `yaml:"-"`
	componentConfigs        ComponentConfigs
}

type DatabaseConfigMap map[string]*DatabaseConfig

func (dbc *DatabaseConfig) DbComponentConfigs() ComponentConfigs {
	return dbc.componentConfigs
}

func (dbc *DatabaseConfig) DatabaseService() (pool *pgxpool.Pool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

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

	if cfg, cfgErr := pgxpool.ParseConfig(connectionStr); cfgErr != nil {
		log.Errorf("DatabaseService: postgres connection failed: %s; \nerror: %v", connectionStr, err)
		return nil, err
	} else {
		var poolErr error
		pool, poolErr = pgxpool.NewWithConfig(ctx, cfg)
		defer pool.Close()
		if poolErr != nil {
			log.Errorf("DatabaseService: failed to establish connection pool: %v", poolErr)
			return nil, poolErr
		}
	}

	if dbc.MaxConnections.Value != "" {
		if parsed, parseErr := strconv.ParseInt(dbc.MaxConnections.Value, 10, 32); parseErr != nil {
			log.Errorf("DatabaseService: failed to parse Max Connections value: %v", parsed)
			return nil, parseErr
		} else {
			pool.Config().MaxConns = int32(parsed)
		}
	}
	//if dbc.MaxIdleConnections.Value != "" {
	//	db.SetMaxIdleConns(toInt(dbc.MaxIdleConnections.Value))
	//}

	if pingErr := pool.Ping(ctx); pingErr != nil {
		return nil, fmt.Errorf("unable to ping database; err: %v", pingErr.Error())
	}

	return pool, nil

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
		log.Errorf("hey dummy! you're missing DB config fields...")
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
