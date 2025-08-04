// AbcImport
package zdb

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zixml"
)

type AbcImporter struct {
}

func guessTarget(kind string) (string, string) {
	kind = strings.ReplaceAll(strings.ToLower(kind), " ", "")
	repo := tuneDB.SourceRepositoryGetAll()
	if len(repo) == 0 {
		return "", ""
	} // should not happen
	defaultR := repo[0].Location
	for _, r := range repo {

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
func AbcImport(abc string, direct bool) (string, error) {
	fmt.Println("Abc Import:", abc)
	txt := strings.Split(abc, "\n")
	warning := ""
	var base string
	buffer := make([]byte, 0)
	kind := ""
	for i, line := range txt {
		if strings.HasPrefix(line, "X:") {
			txt[i] = "X:1"
		}
		if strings.HasPrefix(line, "R:") {
			kind = strings.TrimSpace(strings.TrimPrefix(line, "R:"))
		}

		if strings.HasPrefix(line, "T:") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "T:"))
			base = NiceName(title)
		}
		buffer = append(buffer, []byte(txt[i])...)
		buffer = append(buffer, byte('\n'))
	}
	abcFile := path.Join("tmp", base+".abc")
	f, _ := os.Create(abcFile)
	f.Write(buffer)
	f.Close()

	if xmlFile, err := Abc2Xml(abcFile, "./tmp"); err != nil {
		return fmt.Sprint(err), err
	} else {
		// Check for duplicate
		index := zixml.ComputeIndexForFile(xmlFile)
		duplicates := tuneDB.GetDuplicates(index)
		fmt.Printf("Index:%s %s (%v)\n", index, xmlFile, duplicates)
		if len(duplicates) > 0 {
			warning = "Potential duplicates:\n"
			duplicates = util.Truncate(duplicates, 5)
			for i, d := range duplicates {
				warning += fmt.Sprintf("%d - %s\n", i, d)
			}
		}

		if !direct {
			MuseEdit(xmlFile)
		} else {
			if len(duplicates) > 0 {
				warning += "Tune not imported!\n"
			} else {
				targetRep, Kind := guessTarget(kind)
				msczFile := path.Join(targetRep, base+".mscz")
				cmd := util.H.MkCmd("Xml2Mscz", map[string]string{
					"XmlFile":  xmlFile,
					"MsczFile": msczFile,
				})
				cmd.Run()
				if _, ok := util.GetModificationDate(msczFile); ok {
					tuneDB.MsczTuneSave(msczFile, Kind, time.Now())
					return warning, nil
				}
			}
		}
	}
	return warning, nil
}
