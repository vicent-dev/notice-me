package main

import (
	"embed"
	"gopkg.in/yaml.v2"
	"log"
)

//go:embed config.yaml
var f embed.FS

const (
	configFileName = "config.yaml"
)

type config struct {
	db struct {
		user string `yaml:"user"`
		pwd  string `yaml:"pwd"`
		port string `yaml:"port"`
		host string `yaml:"host"`
		name string `yaml:"name"`
	} `yaml:"db"`
	rabbit struct {
		user string `yaml:"user"`
		pwd  string `yaml:"pwd"`
		port string `yaml:"port"`
		host string `yaml:"host"`
	} `yaml:"rabbit"`
}

func loadConfig() *config {
	c := &config{}

	cFile := getConfigFile()
	err := yaml.Unmarshal(cFile, c)

	if err != nil {
		log.Fatalln(err)
	}

	return c
}

func getConfigFile() []byte {
	bs, _ := f.ReadFile(configFileName)
	return bs
}
