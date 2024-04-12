package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	DBConnect string `yaml:"DB_CONNECT"`
	Cancel time.Duration `yaml:"cancel"`
	UserToken string `yaml:"userToken"`
	AdminToken string `yaml:"adminToken"`
	Port string `yaml:"port"`
	Redis string `yaml:"redisConn"`
}

func Load(configPath string) (*Config, error) {

	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}