// AbcImport
package zdb

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zixml"
)

func guessTarget(kind string) (string, string) {
	kind = strings.ReplaceAll(strings.ToLower(kind), " ", "")
	repo := tuneDB.SourceRepositoryGetAll()
	if len(repo) == 0 {
		return "", ""
	} // should not happen
	defaultR := repo[0].Location
	for _, r := range repo {
		if r.Type != "Mscz" {
			continue
		}
		rk := strings.ReplaceAll(strings.ToLower(r.DefaultKind), " ", "")
		if rk == kind {
			return r.Location, r.DefaultKind
		}
		if rk == "misc" {
			defaultR = r.Location
		}
	}
	return defaultR, "-"
}

type AbcImporter struct {
	initDone bool
	abc      string // cache
	title    string
	base     string
	index    string
	kind     string
	xmlFile  string
}

func NewAbcImporter() *AbcImporter {
	return &AbcImporter{}
}
func (imp *AbcImporter) Start(abc string) error {
	if imp.abc == abc && imp.initDone {
		return nil
	}

	log.Println("AbcImport Start:", strings.ReplaceAll(util.STruncate(abc, 80), "\n", "."))
	imp.abc = abc
	imp.initDone = false

	txt := strings.Split(abc, "\n")
	buffer := make([]byte, 0)
	imp.kind = ""
	buffer = append(buffer, []byte("X:1\n")...)
	isAbc := false
	for i, line := range txt {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "X:") {
			isAbc = true
			continue
		}
		if strings.HasPrefix(line, "R:") {
			isAbc = true
			imp.kind = strings.TrimSpace(strings.TrimPrefix(line, "R:"))
		}

		if strings.HasPrefix(line, "T:") {
			isAbc = true
			if imp.title == "" { // Keeps the firt title
				imp.title = strings.TrimSpace(strings.TrimPrefix(line, "T:"))
				imp.base = NiceName(imp.title)
			}
		}
		buffer = append(buffer, []byte(txt[i])...)
		buffer = append(buffer, byte('\n'))
	}
	if imp.title == "" {
		imp.title = "Unknown"
		imp.base = imp.title
	}
	if !isAbc {
		return fmt.Errorf("Dubious Abc")
	}
	log.Println("Importer start:", imp.title)
	abcFile := path.Join("tmp", imp.base+".abc")
	f, err := os.Create(abcFile)
	if err != nil {
		util.WarnOnError(err)
		return err
	}
	f.Write(buffer)
	f.Close()
	imp.xmlFile, err = Abc2Xml(abcFile, "./tmp")
	if err == nil {
		imp.initDone = true
	}
	return err
}

func (imp *AbcImporter) CheckDuplicates() (string, bool) {
	if !imp.initDone {
		return "Sequence error", true
	}

	imp.index = zixml.ComputeIndexForFile(imp.xmlFile)
	duplicates := tuneDB.GetDuplicates(imp.index)

	if len(duplicates) > 0 {
		warning := "Potential duplicates:\n"
		duplicates = util.Truncate(duplicates, 10)
		for i, d := range duplicates {
			warning += fmt.Sprintf("%d - %s\n", i, d)
		}
		return warning, true
	} else {
		return "", false
	}
}
func (imp *AbcImporter) MuseImport() error {
	if !imp.initDone {
		return fmt.Errorf("Sequence Error")
	}
	MuseEdit(imp.xmlFile)
	return nil
}
func (imp *AbcImporter) DirectImport() error {
	if !imp.initDone {
		return fmt.Errorf("Sequence Error")
	}
	targetRep, Kind := guessTarget(imp.kind)
	msczFile := UniqFileName(path.Join(targetRep, imp.base+".mscz"))
	if cmd := util.H.MkCmd("Xml2Mscz", map[string]string{
		"XmlFile":  imp.xmlFile,
		"MsczFile": msczFile,
	}); cmd != nil {
		cmd.Run()
	}
	if _, ok := util.GetModificationDate(msczFile); ok {
		tuneDB.MsczTuneSave(msczFile, Kind, time.Now())
		return nil
	}
	return fmt.Errorf("Importation Failed")
}
