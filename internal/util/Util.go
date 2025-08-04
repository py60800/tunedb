package util

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type HelperConfig struct {
	Helpers map[string]string
}

var helperConfig HelperConfig

func ReadConfig[T any](fileName string, defaultContent string) T {
	if stat, err := os.Stat(fileName); err != nil || stat.Size() == 0 {
		f, err := os.Create(fileName)
		if err != nil {
			panic(err)
		}
		f.WriteString(defaultContent)
		f.Close()
	}

	// read config
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	var config T
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func (h *Helper) replace(s string, env map[string]string) string {

	rep := make([]string, 0, len(env)*2)
	for k, v := range env {
		rep = append(rep, "${"+k+"}")
		rep = append(rep, v)
	}
	for k, v := range h.currEnv {
		rep = append(rep, "${"+k+"}")
		rep = append(rep, v)
	}
	replacer := strings.NewReplacer(rep...)
	str := s
	for i := 0; ; i++ {
		var b strings.Builder
		replacer.WriteString(&b, str)
		str = b.String()
		if strings.Index(str, "${") < 0 {
			break
		}
		if i >= 10 {
			panic(fmt.Sprintf("Unresolved Symbol loop %v => %v", s, str))
		}
	}
	return str
}

type Helper struct {
	currEnv map[string]string
}

var H *Helper

func HelperInit(context string) {
	helperConfig = ReadConfig[HelperConfig](path.Join(context, "config.yml"), "")
	fmt.Println(helperConfig)
	H = &Helper{currEnv: helperConfig.Helpers}
	H.currEnv["HOME"], _ = os.UserHomeDir()
}
func (h *Helper) Get(name string) string {
	if s, ok := h.currEnv[name]; ok {
		return h.replace(s, map[string]string{})
	}
	panic(fmt.Sprintf("Env lookup failed", name))
	return ""
}
func (h *Helper) MkCmd(cmd string, params map[string]string) *exec.Cmd {
	base, ok := h.currEnv[cmd]
	if !ok {
		panic("Helper MkCmd:" + cmd)
	}
	sp := strings.Fields(base)
	cmdArgs := make([]string, 0)
	for _, item := range sp {
		cmdArgs = append(cmdArgs, h.replace(item, params))
	}
	osCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	fmt.Println("MkCmd:", osCmd.Args)
	return osCmd
}

// MkDir
func MkSubDirs(dir string) (string, string) {
	xmlDir := path.Join(dir, "xml")
	imgDir := path.Join(dir, "img")
	for _, dir := range []string{xmlDir, imgDir} {
		if stat, err := os.Stat(dir); err != nil {
			if err = os.Mkdir(dir, 0777); err != nil {
				panic(fmt.Sprintf("MkDir %v %v", dir, err))
			}
		} else {
			if !stat.IsDir() {
				panic(fmt.Sprintf("%v Not a directory", dir))
			}
		}
	}
	return xmlDir, imgDir
}
func FatalErr(err error) {
	if err != nil {
		var msg string
		pc, file, ln, ok := runtime.Caller(1)
		if ok {
			caller := "??"
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				caller = fn.Name()
			}
			msg = fmt.Sprintf("caller:%v file:%v line:%v error:%v\n", caller, path.Base(file), ln, err)
		} else {
			msg = fmt.Sprint("Unlocalised error:%v", err)
		}
		panic(msg)
	}
}
func CheckErr(err error, txt ...string) bool {
	if err != nil {
		var msg string
		pc, file, ln, ok := runtime.Caller(1)
		if ok {
			caller := "??"
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				caller = fn.Name()
			}
			msg = fmt.Sprintf("caller:%v file:%v line:%v error:%v\n", caller, path.Base(file), ln, err)
		} else {
			msg = fmt.Sprint("Unlocalised error:%v", err)
		}
		for i, t := range txt {
			if i != 0 {
				msg += ", "
			}
			msg += t
		}
		fmt.Print(msg)
		return true
	}
	return false
}

func StrTime(t time.Time) string {
	return t.Format("02/01/06")
}
func Truncate[T any](array []T, mx int) []T {
	if len(array) < mx {
		return array
	}
	return array[:mx]
}

func STruncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
