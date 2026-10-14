package config

import (
	"log"
	"notice-me-server/static"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
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
		Host    string   `yaml:"host"`
		Port    string   `yaml:"port"`
		Cors    []string `yaml:"cors"`
		Env     string   `yaml:"env"`
		TlsKey  string   `yaml:"tlsKey"`
		TlsCert string   `yaml:"tlsCert"`
	} `yaml:"server"`
	Db struct {
		User string `yaml:"user"`
		Pwd  string `yaml:"pwd"`
		Port string `yaml:"port"`
		Host string `yaml:"host"`
		Name string `yaml:"name"`
	} `yaml:"db"`
	Rabbit struct {
		User           string        `yaml:"user"`
		Pwd            string        `yaml:"pwd"`
		Port           string        `yaml:"port"`
		Host           string        `yaml:"host"`
		Queues         []QueueConfig `yaml:"queues"`
		ConsumersCount int           `yaml:"consumersCount"`
	} `yaml:"rabbit"`
}

var envRef = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-(.*?))?\}`)

func expandEnv(in string) string {
	return envRef.ReplaceAllStringFunc(in, func(ref string) string {
		match := envRef.FindStringSubmatch(ref)
		if len(match) < 3 {
			return ref
		}
		if value := os.Getenv(match[1]); value != "" {
			return value
		}
		return match[2]
	})
}

func (c *Config) withDefaults() {
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	if c.Server.Env == "" {
		c.Server.Env = "development"
	}
	if c.Db.Port == "" {
		c.Db.Port = "3306"
	}
	if c.Rabbit.Port == "" {
		c.Rabbit.Port = "5672"
	}
}

func LoadConfig() *Config {
	_ = godotenv.Load() // non-fatal

	cFile := static.GetConfigFile()
	expanded := expandEnv(string(cFile))

	c := &Config{}
	if err := yaml.Unmarshal([]byte(expanded), c); err != nil {
		log.Fatalln(err)
	}

	c.withDefaults()
	return c
}
