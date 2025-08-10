// Helpers
package zdb

import (
	"fmt"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/py60800/tunedb/internal/svgtab"
	"github.com/py60800/tunedb/internal/util"

	"github.com/mozillazg/go-unidecode"
	"gorm.io/gorm"

	"os/exec"
)

var RequiredHelpers = []string{"Mus4", "Python"}

func CheckHelpers() (string, error) {
	res := ""

	for _, h := range RequiredHelpers {
		exe := util.H.Get(h)
		if _, err := exec.LookPath(exe); err != nil {
			res += fmt.Sprintf(" Not found : %v\n", exe)
		}
	}
	if res != "" {
		res += "Please fix required element in \"config.yml\""
		return res, fmt.Errorf("Missing required helper")
	}
	return "", nil
}

func (f *MP3File) ToString() string {
	return util.STruncate(fmt.Sprintf("%s : %s", f.Artist, f.Title), 70)
}
func MuseEdit(file string) {
	cmd := util.H.MkCmd("MusEdit", map[string]string{"File": file})
	if cmd != nil {
		cmd.Start()
	}
}
func LaunchAudacity(file string) {
	if file != "" {
		cmd := util.H.MkCmd("Audacity", map[string]string{"Mp3File": file})
		if cmd != nil {
			cmd.Start()
		}
	}
}

func (t *DTune) MsczRefreshImg() {
	start := time.Now().Add(-2 * time.Second)
	base := strings.TrimSuffix(path.Base(t.File), ".mscz")
	imgFile := path.Join(path.Dir(t.Img), base+".svg")
	cmd := util.H.MkCmd("Mscz2Img", map[string]string{"ImgFile": imgFile, "File": t.File})
	if cmd == nil {
		return
	}
	if err := cmd.Run(); err != nil {
		fmt.Println("Refresh Img:", err)
	}

	date, ok := GetModificationDate(imgFile)
	if !ok || date.Before(start) {
		altImg := strings.TrimSuffix(imgFile, ".svg") + "-1.svg"
		date, ok = GetModificationDate(altImg)
		if ok {
			t.Img = altImg
		}

	} else {
		t.Img = imgFile
	}
	svgtab.SvgEnhance(t.Img)
}
func (t *DTune) MsczRefreshXml() {
	start := time.Now().Add(-2 * time.Second)
	cmd := util.H.MkCmd("Mscz2Xml", map[string]string{"File": t.File, "XmlFile": t.Xml})
	if cmd == nil {
		return
	}
	if err := cmd.Run(); err != nil {
		fmt.Println("MsczXml Failed:", err)
	}
	date, ok := GetModificationDate(t.Xml)
	if !ok || date.Before(start) {
		fmt.Println("RefreshXml Failed")
	}

}

func NiceName(s string) string {
	lt := [...][2]string{
		[2]string{`-[0-9]$`, ``},
		[2]string{`_s( |_)`, `'s `},
		[2]string{`_s$`, `'s`},
		[2]string{`_`, ` `},
		[2]string{` +`, ` `},
		[2]string{`(^The) (.*)`, `$2, $1`},
		[2]string{`(^An) (.*)`, `$2, $1`},
		[2]string{`(^A) (.*)`, `$2, $1`},
		[2]string{`  The`, `, The`},
	}
	t := []byte(s)
	for _, r := range lt {
		re := regexp.MustCompile(r[0])
		t = re.ReplaceAll(t, []byte(r[1]))
	}
	return strings.ToLower(unidecode.Unidecode(string(t)))

}
func warnOnDbError(cnx *gorm.DB) {
	if cnx.Error != nil {
		pc, file, ln, ok := runtime.Caller(1)
		if ok {
			caller := "??"
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				caller = fn.Name()
			}
			fmt.Printf("caller:%v file:%v line:%v error:%v\n", caller, path.Base(file), ln, cnx.Error)
		}
	}
}
