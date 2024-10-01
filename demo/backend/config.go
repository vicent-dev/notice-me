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
	Db struct {
		User string `yaml:"user"`
		Pwd  string `yaml:"pwd"`
		Port string `yaml:"port"`
		Host string `yaml:"host"`
		Name string `yaml:"name"`
	} `yaml:"db"`
	Rabbit struct {
		User string `yaml:"user"`
		Pwd  string `yaml:"pwd"`
		Port string `yaml:"port"`
		Host string `yaml:"host"`
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
