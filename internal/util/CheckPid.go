// CheckPid.go
package util

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
)

var pidFile = "tunedb.pid"

func RemovePidFile() {
	os.Remove(pidFile)
}
func CheckPid() {
	data, _ := os.ReadFile(pidFile)
	prevPid, err := strconv.Atoi(string(data))
	log.Println("Previous pid", prevPid)
	if err == nil {
		p, err := os.FindProcess(prevPid)
		if err == nil {
			if p.Signal(syscall.Signal(0)) == nil {
				// Process exists
				fmt.Println("Database in use !")
				os.Exit(1)
			}
		}
	}

	newPid := strconv.FormatInt(int64(os.Getpid()), 10)
	os.WriteFile(pidFile, []byte(newPid), 0644)
}
