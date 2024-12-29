package config

import (
	"errors"
	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name    string
		Version string
		Author  string
	}
	Service struct {
		Http struct {
			Host    string
			Port    int
			Prefork bool
		}
		Redis struct {
			Addr string
			DB   int
		}
	}
	Logger struct {
		Level        string
		ReportCaller bool `mapstructure:"report_caller"`
		Format       string
		PrettyPrint  bool `mapstructure:"pretty_print"`
	}
	Workerpool struct {
		Size int
	}
}

var configCache *Config

func LoadConfig(filename string) (*Config, error) {
	cfg := viper.New()
	cfg.SetConfigFile(filename)
	cfg.AutomaticEnv()

	if err := cfg.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := cfg.Unmarshal(&configCache); err != nil {
		return nil, err
	}

	cfg.WatchConfig()

	return configCache, nil
}

func GetConfig() (*Config, error) {
	if configCache == nil {
		return nil, errors.New("no config file loaded")
	}

	return configCache, nil
}
