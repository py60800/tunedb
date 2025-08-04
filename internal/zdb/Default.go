package zdb

import (
	_ "embed"
	"os"
	"path"
)

//go:embed svg.mss
var defaultSvgStyle []byte

//go:embed abc2xml.py
var defaultAbc2Xml []byte

/*
//go:embed config.yml
var defaultConfiguration []byte
*/

func createDefault(file string, content []byte) {
	if _, ok := GetModificationDate(file); !ok {
		f, _ := os.Create(file)
		f.Write(content)
		f.Close()
	}
}

func CreateDefaultFiles(defaultBase string) {
	os.Mkdir(defaultBase, 0777)
	createDefault(path.Join(defaultBase, "svg.mss"), defaultSvgStyle)
	createDefault(path.Join(defaultBase, "abc2xml.py"), defaultAbc2Xml)
	createDefault(path.Join(defaultBase, "config.yml"), defaultConfiguration)
}
