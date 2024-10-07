package static

import (
	"embed"
)

//go:embed config.yaml docs.html
var f embed.FS

const (
	configFileName = "config.yaml"
	docsFileName   = "docs.html"
)

func GetConfigFile() []byte {
	bs, _ := f.ReadFile(configFileName)
	return bs
}

func GetDocsFile() []byte {
	bs, _ := f.ReadFile(docsFileName)
	return bs
}
