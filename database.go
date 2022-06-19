package config_yaml

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Name             yaml.Node `yaml:"Name"`
	Database         yaml.Node `yaml:"Database"`
	Server           yaml.Node `yaml:"Server"`
	Port             yaml.Node `yaml:"Port"`
	Username         yaml.Node `yaml:"Username"`
	Password         yaml.Node `yaml:"Password"`
	Scheme           yaml.Node `yaml:"Scheme"`
	MongoClient      *mongo.Client
	componentConfigs ComponentConfigs
}

type DatabaseConfigMap map[string]*DatabaseConfig

func (s *DatabaseConfig) DbComponentConfigs() ComponentConfigs {
	return s.componentConfigs
}

func InitDbService(dbc *DatabaseConfig) (*mongo.Client, []error) {
	var errs []error

	if dbc.Password.Value == "" {
		errs = append(errs, MissingField("Password"))
		return nil, errs
	}
	connectionString := ""
	switch dbc.Scheme.Value {
	case "mongo":
		{
			connectionString = "mongodb://" + dbc.Username.Value + ":" + dbc.Password.Value + "@" + dbc.Server.Value + ":" + dbc.Port.Value
		}
	}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatalf("unable to open connection with database; err: %v", err.Error())
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalln(err)
	}

	return client, nil
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
