// HomeContext
package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/leemcloughlin/logfile"
	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zdb"
)

var wHeader = "Default"

var ConfigBase = path.Join(".", "context")
var tuneDB = "TuneDb"

func MakeHomeContext(baseDir string) {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Errorf("No Home directory %v", err))
		}
		baseDir = path.Join(homeDir, "Music", tuneDB)
		_, err = os.Stat(baseDir)
		if err != nil {
			err = os.MkdirAll(baseDir, 0777)
			if err != nil {
				panic(fmt.Errorf("Can't create working directory: %v", err))
			}
		}
	} else {
		os.MkdirAll(baseDir, 0777) // Create New if required
	}
	wHeader = path.Base(baseDir)
	err := os.Chdir(baseDir)
	if err != nil {
		panic(fmt.Errorf("Failed to select home directory : %v", err))
	}
	util.CheckPid()

	os.Mkdir("log", 0777)

	// Set log file
	lf := logfile.Defaults

	lf.FileName = path.Join("log", fmt.Sprintf("tunedb-%s.log", time.Now().Format("060102-150405")))
	if fl, err := logfile.New(&lf); err == nil {
		log.SetOutput(fl)
	}
	log.Printf("TuneDB start %s %s\n", time.Now().Format("06/01/02-15:04:05"), baseDir)
	os.Mkdir("tmp", 0777)
	zdb.CreateDefaultFiles(ConfigBase)

}
