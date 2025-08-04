// Files
package zdb

import (
	"archive/zip"
	"encoding/xml"

	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/py60800/tunedb/internal/util"
)

type MTag struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",innerxml"`
}
type MScore struct {
	XMLName xml.Name `xml:"Score"`
	MetaTag []MTag   `xml:"metaTag"`
}
type Mscz struct {
	XMLName xml.Name `xml:"museScore"`
	Score   []MScore `xml:"Score"`
}

func MsczGetTitle(file string) string {
	r, err := zip.OpenReader(file)
	if err != nil {
		fmt.Println(file, " ", err)
		return "No title"
	}
	defer r.Close()
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".mscx") {
			rc, err := f.Open()
			defer rc.Close()
			util.WarnOnError(err)
			byteValue, _ := ioutil.ReadAll(rc)
			var score Mscz
			xml.Unmarshal(byteValue, &score)
			if len(score.Score) > 0 {
				for _, t := range score.Score[0].MetaTag {
					if t.Name == "workTitle" {
						return t.Value
					}
				}
			}

		}

	}
	return "No title"
}

type MPartition struct {
	XMLName xml.Name `xml:"score-partwise"`
	Work    MWork    `xml:"work"`
	//	Identification MIdentification `xml:"identification"`
}
type MWork struct {
	Title string `xml:"work-title"`
}

func XmlGetTitle(file string) string {
	xmlFile, err := os.Open(file)
	defer xmlFile.Close()
	util.WarnOnError(err)
	byteValue, _ := ioutil.ReadAll(xmlFile)
	var partition MPartition
	xml.Unmarshal(byteValue, &partition)
	return partition.Work.Title
}

// General purpose
func GetFileList(dir string, extension string) []string {
	files, err := os.ReadDir(dir)
	if err != nil {
		util.WarnOnError(err)
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
func GetFiles(FileBase string, Dir string, SubDir string, Extension string, Suffix string) ([]string, map[string]string) {
	l := GetFileList(path.Join(FileBase, Dir, SubDir), Extension)
	m := make(map[string]string)
	for _, p := range l {
		name := NiceName(strings.TrimSuffix(path.Base(p), Suffix))
		m[name] = p
	}
	return l, m
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
		util.WarnOnError(err)
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

type FileInfo struct {
	Name string
	Date time.Time
}

func GetFileListR2(dir string, extension string, recursive bool) []FileInfo {
	files, err := os.ReadDir(dir)
	if err != nil {
		util.WarnOnError(err)
		return []FileInfo{}
	}
	res := make([]FileInfo, 0, len(files))
	for _, de := range files {
		name := de.Name()
		if de.IsDir() && recursive {
			sub := GetFileListR2(path.Join(dir, name), extension, recursive)
			res = append(res, sub...)
		} else {
			if path.Ext(name) == extension {
				info, _ := de.Info()
				res = append(res, FileInfo{
					Name: path.Join(dir, name),
					Date: info.ModTime(),
				})
			}
		}
	}
	return res
}
