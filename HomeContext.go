// HomeContext
package main

import (
	"os"
	"path"

	"github.com/py60800/tunedb/internal/zdb"
)

var wHeader = "Default"

var ConfigBase = path.Join(".", "context")

func MakeHomeContext(baseDir string) {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		baseDir = path.Join(homeDir, "Music", "tunedb")
		_, err = os.Stat(baseDir)
		if err != nil {
			err = os.MkdirAll(baseDir, 0777)
			if err != nil {
				panic(err)
			}
		}
	} else {
		os.MkdirAll(baseDir, 0777) // Create New if required

	}
	wHeader = path.Base(baseDir)
	err := os.Chdir(baseDir)
	if err != nil {
		panic(err)
	}

	os.Mkdir("tmp", 0777)
	zdb.CreateDefaultFiles(ConfigBase)

}
