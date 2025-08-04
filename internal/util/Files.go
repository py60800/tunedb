// Files
package util

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"
)

func errorMsg(kind string, err error) string {
	if pc, file, ln, ok := runtime.Caller(2); ok {
		caller := "??"
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			caller = fn.Name()
		}
		return fmt.Sprintf("%v :%v file:%v line:%v error:%v\n", kind, caller, path.Base(file), ln, err)
	} else {
		return fmt.Sprintf("%v : unlocated error:%v\n", kind, ln, err)
	}

}
func PanicOnError(err error) {
	if err != nil {
		panic(fmt.Errorf("%s", errorMsg("Fatal", err)))
	}
}
func WarnOnError(err error) {
	if err != nil {
		fmt.Println(errorMsg("Warn", err))
	}
}

func GetFileList(dir string, extension string) []string {
	files, err := os.ReadDir(dir)
	if err != nil {
		WarnOnError(err)
		return []string{}
	}
	res := make([]string, 0, len(files))
	for _, de := range files {
		e := de.Name()
		if path.Ext(e) == extension {
			res = append(res, path.Join(dir, e))
		}
	}
	return res
}

func GetModificationDate(file string) (time.Time, bool) {
	var t time.Time
	if stat, err := os.Stat(file); err == nil {
		return stat.ModTime(), true
	}
	return t, false

}
func GetFileListR(dir string, extension string, recursive bool) []string {
	files, err := os.ReadDir(dir)
	if err != nil {
		WarnOnError(err)
		return []string{}
	}
	res := make([]string, 0, len(files))
	for _, de := range files {
		name := de.Name()
		if de.IsDir() && recursive {
			sub := GetFileListR(path.Join(dir, name), extension, recursive)
			res = append(res, sub...)
		} else {
			if path.Ext(name) == extension {
				res = append(res, path.Join(dir, name))
			}
		}
	}
	return res
}
