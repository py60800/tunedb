// AbcImport
package zdb

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/py60800/abc2xml"
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
	message  string
}

func NewAbcImporter() *AbcImporter {
	return &AbcImporter{}
}
func (imp *AbcImporter) BuiltInAbc2xml(abc string) error {
	log.Println("Builtin Abc2xml")
	parser := abc2xml.Abc2xmlNew()
	xml, err := parser.Run(abc)
	util.WarnOnError(err)
	if err != nil {
		return err
	}
	imp.title = parser.GetTitle()
	imp.base = NiceName(imp.title)
	imp.kind = parser.GetRythm()
	msg := parser.Warnings()
	imp.message = ""
	pfx := ""
	for _, s := range msg {
		imp.message += pfx + s
		pfx = "\n"
	}

	imp.xmlFile = path.Join("tmp", imp.base+".xml")
	err = os.WriteFile(imp.xmlFile, []byte(xml), 0666)
	util.WarnOnError(err)
	return err
}
func (imp *AbcImporter) ExternalAbc2xml(abc string) error {
	log.Println("")
	txt := strings.Split(abc, "\n")
	buffer := make([]byte, 0)
	imp.kind = ""
	buffer = append(buffer, []byte("X:1\n")...)
	isAbc := false
	for i, line := range txt {
		switch {
		case strings.TrimSpace(line) == "":
			// Ignore blank line
			continue
		case strings.HasPrefix(line, "X:"):
			isAbc = true
		case strings.HasPrefix(line, "R:"):
			isAbc = true
			imp.kind = strings.TrimSpace(strings.TrimPrefix(line, "R:"))
		case strings.HasPrefix(line, "T:"):
			isAbc = true
			if imp.title == "" { // Keeps the firt title
				imp.title = strings.TrimSpace(strings.TrimPrefix(line, "T:"))
				imp.base = NiceName(imp.title)
			}
		}
		buffer = append(buffer, []byte(txt[i])...)
		buffer = append(buffer, byte('\n'))
		log.Println("AbcImport Start:", strings.ReplaceAll(util.STruncate(abc, 80), "\n", "."))
		imp.abc = abc
		imp.initDone = false
	}
	if imp.title == "" {
		imp.title = "Unknown"
		imp.base = imp.title
	}
	if !isAbc {
		return fmt.Errorf("Dubious Abc")
	}
	abcFile := path.Join("tmp", imp.base+".abc")
	f, err := os.Create(abcFile)
	if err != nil {
		util.WarnOnError(err)
		return err
	}
	f.Write(buffer)
	f.Close()
	_, err = Abc2XmlPy(abcFile, "./tmp")
	return err

}

func (imp *AbcImporter) Start(abc string) (string, error) {
	if imp.abc == abc && imp.initDone {
		return "Sequence error", nil
	}

	log.Println("AbcImport Start:", strings.ReplaceAll(util.STruncate(abc, 80), "\n", "."))
	imp.abc = abc
	imp.initDone = false
	imp.message = ""
	var err error
	if mode, ok := util.H.Env["Abc2XmlProg"]; !ok || mode == "Builtin" {
		err = imp.BuiltInAbc2xml(abc)
	} else {
		err = imp.ExternalAbc2xml(abc)
	}
	imp.initDone = err == nil
	return imp.message, err
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
		"XmlFile": imp.xmlFile, "MsczFile": msczFile}); cmd != nil {
		cmd.Run()
	}
	if _, ok := util.GetModificationDate(msczFile); ok {
		tuneDB.MsczTuneSave(msczFile, Kind, time.Now())
		return nil
	}
	return fmt.Errorf("Importation Failed")
}
