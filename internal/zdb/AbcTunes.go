// AbcTunes
package zdb

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/py60800/tunedb/internal/util"

	"gorm.io/gorm"
)

type AbcTune struct {
	File  string
	X     int
	Hash  string
	Title string
	Kind  string
	Key   string
	Raw   string
	Xml   string
	Svg   string
}

func Hash(data string) string {
	hash := md5.Sum([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])

}
func AbcParseFile(file string) []AbcTune {
	tunes := make([]AbcTune, 0)
	readFile, err := os.Open(file)
	util.WarnOnError(err)
	found := make(map[int]int)
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	inTune := false

	var tune AbcTune

	finishTune := func() {
		if inTune {
			if _, exists := found[tune.X]; !exists {
				tune.Hash = Hash(tune.Raw)
				tunes = append(tunes, tune)
				found[tune.X] = 0
			} else {
				log.Printf("X Duplicate: %s X:%d\n", file, tune.X)
			}
		}
		inTune = false
		tune = AbcTune{File: file}

	}
	for fileScanner.Scan() {
		line := fileScanner.Text()
		line = strings.TrimSpace(fileScanner.Text())
		if line == "" {
			finishTune()
		}
		isX := false
		if len(line) > 1 && line[1] == ':' {
			if unicode.IsLetter(rune(line[0])) {
				inTune = true
				lineData := strings.TrimSpace(line[2:])
				switch unicode.ToUpper(rune(line[0])) {
				case 'X':
					isX = true
					sp := strings.Split(lineData, " \t") // Get rid of any garbage
					if len(sp) > 0 {
						tune.X, _ = strconv.Atoi(sp[0])
					}
				case 'T':
					tune.Title = lineData
				case 'K':
					tune.Key = lineData
				case 'R':
					tune.Kind = lineData
				}
			}
		}
		if inTune && !isX {
			tune.Raw = tune.Raw + "\n" + line
		}
	}
	finishTune()

	readFile.Close()
	return tunes
}

func MakeTempFile(tmpDir string, t *AbcTune, base string) string {
	tmpFile := path.Join(tmpDir, fmt.Sprintf("%s-%03d.abc", base, t.X))
	file, err := os.Create(tmpFile)
	util.PanicOnError(err)
	fmt.Fprintf(file, "X:1")
	fmt.Fprintln(file, t.Raw)
	file.Close()
	return tmpFile
}

func Abc2Xml(abcFile string, dir string) (string, error) {
	cmd := util.H.MkCmd("Abc2Xml", map[string]string{
		"AbcFile": abcFile,
		"XmlDir":  dir,
	})
	if cmd == nil {
		return "", fmt.Errorf("Command missing")
	}
	err := cmd.Run()
	util.WarnOnError(err)

	base := strings.TrimSuffix(path.Base(abcFile), ".abc")
	result := path.Join(dir, base+".xml")

	_, err = os.Stat(result)
	if err != nil {
		util.WarnOnError(fmt.Errorf("Xml file not found: %v %v\n", result, err))
		return "", err
	}
	return result, nil
}
func Abc2Svg(abcFile string, dir string) string {
	base := strings.TrimSuffix(path.Base(abcFile), ".abc")
	expectedResult := path.Join(dir, base+".svg")

	cmd := util.H.MkCmd("Abc2Img", map[string]string{
		"AbcFile": abcFile,
		"ImgFile": expectedResult,
	})
	if cmd == nil {
		return ""
	}
	err := cmd.Run()
	if err != nil {
		util.WarnOnError(fmt.Errorf("Abc2Svg: %v %v %v\n", abcFile, err, cmd.Args))
		return ""
	}
	result := ""
	if _, err = os.Stat(expectedResult); err == nil {
		result = expectedResult
	} else {
		// X:1 forced by design
		altName := path.Join(dir, fmt.Sprintf("%s%03d.svg", base, 1))
		if _, err = os.Stat(altName); err == nil {
			result = altName
		}
	}

	if result == "" {
		log.Println("No SVG for ", abcFile)
	}
	return result
}

var previous map[string]int

func AbcTuneStore(db *TuneDB, tune *AbcTune, TmpDir, XmlDir, ImgDir string) {
	var d DTune
	res := db.cnx.Where("abc_hash = ?", tune.Hash).First(&d)
	if res.Error != gorm.ErrRecordNotFound {
		log.Println("Abc Duplicate:", tune.Title)
		return // Already in DB
	}
	// Check if in database
	base := strings.TrimSuffix(path.Base(tune.File), ".abc")
	var theTune DTune
	res = db.cnx.Where("file = ? and x_ref = ? and file_type = ?", tune.File, tune.X, FileTypeAbc).First(&theTune)

	if res.Error == gorm.ErrRecordNotFound {
		log.Printf("AbcTuneStore: Record Not found\n")
		tmpFile := MakeTempFile(TmpDir, tune, base)
		var err error
		if tune.Xml, err = Abc2Xml(tmpFile, XmlDir); err != nil {
			util.WarnOnError(err)
		}
		tune.Svg = Abc2Svg(tmpFile, ImgDir)

		theTune = DTune{
			File:     tune.File,
			Date:     time.Now(),
			FileType: FileTypeAbc,
			XRef:     tune.X,
			AbcHash:  tune.Hash,
			Xml:      tune.Xml,
			Img:      tune.Svg,
			Title:    tune.Title,
			Kind:     CleanTuneKind(tune.Kind),
			NiceName: NiceName(tune.Title),
		}
		log.Println("AbcTuneStore Create:", theTune.File, theTune.XRef, theTune.FileType)
		db.cnx.Create(&theTune)
	} else {
		log.Println("AbcTuneStore : Tune found", theTune)
		if theTune.AbcHash == tune.Hash {
			// Already in Database
			log.Println("AbcTuneStore No change")
			return
		}
		tmpFile := MakeTempFile(TmpDir, tune, base)
		var err error
		if theTune.Xml, err = Abc2Xml(tmpFile, XmlDir); err != nil {
			util.WarnOnError(err)
		}
		theTune.Img = Abc2Svg(tmpFile, ImgDir)
		theTune.AbcHash = tune.Hash
		theTune.Title = tune.Title
		theTune.Kind = CleanTuneKind(tune.Kind)

		db.cnx.Save(&theTune)

	}
}
func AbcParse(db *TuneDB, dir string, recurse bool) {
	files := GetFileListR(dir, ".abc", recurse)
	XmlDir := path.Join(dir, "xml")
	ImgDir := path.Join(dir, "img")
	TmpDir := path.Join(dir, "tmp")
	os.Mkdir(XmlDir, 0777)
	os.Mkdir(ImgDir, 0777)
	os.Mkdir(TmpDir, 0777)
	for _, f := range files {
		abcTunes := AbcParseFile(f)
		for _, abc := range abcTunes {
			AbcTuneStore(db, &abc, TmpDir, XmlDir, ImgDir)
		}
	}
}
func AbcDBUpdate(db *TuneDB) {
	dirs := db.SourceRepositoryGetAll()
	previous = make(map[string]int)
	for _, d := range dirs {
		if d.Type == "Abc" {
			AbcParse(db, d.Location, d.Recurse)
		}
	}
}
