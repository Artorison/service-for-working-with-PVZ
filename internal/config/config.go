package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DB  DataBaseCfg `yaml:"database"`
	App AppCfg      `yaml:"server"`
}

type DataBaseCfg struct {
	Host     string `yaml:"host" env:"DB_HOST" env-default:"db"`
	Port     string `yaml:"port" env:"DB_PORT" env-default:"5455"`
	User     string `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-default:"123456"`
	DBName   string `yaml:"dbname" env:"DB_NAME" env-default:"postgres"`
}

type AppCfg struct {
	Port    string `yaml:"port"`
	Address string `yaml:"address"`
}

func LoadConfig(configPath string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) GetPort() string {
	return c.App.Port
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("http://%s:%s/", c.App.Address, c.App.Port)
}

func (dbCfg *DataBaseCfg) GetDsn() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.DBName,
	)
}
