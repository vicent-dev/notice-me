package static

import (
	"embed"
)

//go:embed config.yaml
var f embed.FS

const (
	configFileName = "config.yaml"
)

func GetConfigFile() []byte {
	bs, _ := f.ReadFile(configFileName)
	return bs
}
