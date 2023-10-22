package config

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Server struct {
		Host string `yaml:"host" envconfig:"SERVER_HOST"`
		Port string `yaml:"port" envconfig:"SERVER_PORT"`
	} `yaml:"server"`

	Database struct {
		DbHost     string `yaml:"host" envconfig:"DB_HOST"`
		DbPort     string `yaml:"port" envconfig:"DB_PORT"`
		DbUsername string `yaml:"username" envconfig:"DB_USERNAME"`
		DbPassword string `yaml:"password" envconfig:"DB_PASSWORD"`
		DbName     string `yaml:"name" envconfig:"DB_NAME"`
	} `yaml:"database"`
}

type DatabaseConfig struct {
	DbHost     string
	DbPort     string
	DbUsername string
	DbPassword string
	DbName     string
}

func (cfg *Config) GetDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		DbHost:     cfg.Database.DbHost,
		DbPort:     cfg.Database.DbPort,
		DbUsername: cfg.Database.DbUsername,
		DbPassword: cfg.Database.DbPassword,
		DbName:     cfg.Database.DbName,
	}
}

func New(configFileName string) *Config {
	var config Config
	file, err := os.Open(configFileName)
	if err != nil {
		os.Exit(2)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		os.Exit(2)
	}

	// read env variables
	readEnv(&config)
	return &config
}

func readEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		os.Exit(2)
	}
}
