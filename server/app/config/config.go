package config

import (
	"log"
	"notice-me-server/static"

	"gopkg.in/yaml.v2"
)

type QueueConfig struct {
	Name       string `yaml:"name"`
	Exchange   string `yaml:"exchange"`
	Durable    bool   `yaml:"durable"`
	AutoDelete bool   `yaml:"autoDelete"`
	Exclusive  bool   `yaml:"exclusive"`
	NoWait     bool   `yaml:"noWait"`
}

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	Db struct {
		User string `yaml:"user"`
		Pwd  string `yaml:"pwd"`
		Port string `yaml:"port"`
		Host string `yaml:"host"`
		Name string `yaml:"name"`
	} `yaml:"db"`
	Rabbit struct {
		User    string                 `yaml:"user"`
		Pwd     string                 `yaml:"pwd"`
		Port    string                 `yaml:"port"`
		Host    string                 `yaml:"host"`
		Queues  map[string]QueueConfig `yaml:"queues"`
		Enabled bool                   `yaml:"enabled"`
	} `yaml:"rabbit"`
}

func LoadConfig() *Config {
	c := &Config{}

	cFile := static.GetConfigFile()
	err := yaml.Unmarshal(cFile, c)

	if err != nil {
		log.Fatalln(err)
	}

	return c
}
