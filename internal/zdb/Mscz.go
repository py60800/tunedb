// Mscz.go
package zdb

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zixml"

	"gorm.io/gorm"
)

func (db *TuneDB) MsczTuneSave(mscz string, defaultKind string, msczDate time.Time) {
	base := strings.TrimSuffix(path.Base(mscz), ".mscz")
	xmlDir := path.Join(path.Dir(mscz), "xml")
	imgDir := path.Join(path.Dir(mscz), "img")

	//msczDate, _ := GetModificationDate(mscz)

	var tuneInDB DTune
	var updatedTune DTune

	res := db.cnx.Where("File = ?", mscz).First(&tuneInDB)

	if res.Error == gorm.ErrRecordNotFound {
		updatedTune.File = mscz
		updatedTune.FileDate = msczDate
		updatedTune.Kind = defaultKind
		updatedTune.Title = MsczGetTitle(mscz)
		updatedTune.NiceName = NiceName(updatedTune.Title)
		updatedTune.Date = time.Now()
		updatedTune.Hide = false
	} else {
		if tuneInDB.FileDate.Unix() == msczDate.Unix() {
			// Same MSCZ date => No change
			if !strings.HasSuffix(tuneInDB.Img, ".svg") {
				updatedTune.Img = path.Join(imgDir, base+"-1.svg")
				tuneInDB.Img = updatedTune.Img

			}
			_, okXml := GetModificationDate(tuneInDB.Xml)
			_, okImg := GetModificationDate(tuneInDB.Img)
			//			if okXml && okImg { //
			if okXml && okImg && tuneInDB.BreathnachCode != "" { // Temp test for DB update
				return
			}
		}
		updatedTune.ID = tuneInDB.ID
		updatedTune.File = mscz
	}

	updatedTune.Xml = path.Join(xmlDir, base+".musicxml")
	xmlDate, okXml := GetModificationDate(updatedTune.Xml)
	if !okXml || xmlDate.Before(msczDate) {
		updatedTune.MsczRefreshXml()
	}

	// Locate image
	updatedTune.Img = path.Join(imgDir, base+"-1.svg")
	imgDate, okImg := GetModificationDate(updatedTune.Img)
	if !okImg || imgDate.Before(msczDate) {
		updatedTune.MsczRefreshImg()
	}
	updatedTune.FileDate = msczDate

	if _, ok := GetModificationDate(updatedTune.Xml); ok { // xml exists
		var mode string
		updatedTune.FirstNote, mode, updatedTune.Fifth, updatedTune.BreathnachCode = zixml.GetFirstNote(updatedTune.Xml)
		updatedTune.Mode = ModeCommonName(Note(updatedTune.Fifth), XmlModeXRef[mode])
		if tuneInDB.ID != 0 {
			updatedTune.Mode = tuneInDB.Mode
			updatedTune.Fifth = tuneInDB.Fifth
		}
	} else {
		fmt.Println("xml file is missing", updatedTune.Xml) // should panic
	}
	if tuneInDB.ID != 0 {
		db.cnx.Model(&updatedTune).Updates(&updatedTune)
	} else {
		db.cnx.Create(&updatedTune)
	}
}
func (db *TuneDB) MsczContentUpdate() {
	repositories := db.SourceRepositoryGetAll()
	t0 := time.Now()
	lTunes := db.TunesGetFileList()
	m := make(map[string]int)
	for i := range lTunes {
		m[lTunes[i].File] = i
	}
	da := time.Now().Sub(t0)
	var d0 time.Duration
	var d1 time.Duration
	for _, repo := range repositories {
		if repo.Type != "Mscz" {
			continue
		}

		if _, ok := GetModificationDate(repo.Location); !ok {
			os.MkdirAll(repo.Location, 0777) // Create dir any way
			fmt.Println("Directory not found:", repo.Location)
			continue
		}

		util.MkSubDirs(repo.Location) // Create directories anyway
		ta := time.Now()
		lMscz := GetFileListR2(repo.Location, ".mscz", repo.Recurse)
		d0 += time.Now().Sub(ta)
		tb := time.Now()
		for _, mscz := range lMscz {
			if idx, ok := m[mscz.Name]; !ok || mscz.Date.Unix() != lTunes[idx].FileDate.Unix() {
				fmt.Println("Update:", mscz.Name)
				db.MsczTuneSave(mscz.Name, repo.DefaultKind, mscz.Date)
			}
		}
		d1 += time.Now().Sub(tb)
	}
	fmt.Printf("MszUpdate: T0 : %v D0: %v D1: %v (Total:%v)\n", da, d0, d1, time.Now().Sub(t0))
}
func (db *TuneDB) PurgeMscz() {
	var tunes []DTune
	db.cnx.Find(&tunes)
	purge, _ := os.Create("purge.sql")
	fpurge, _ := os.Create("purgeFile.sh")
	for _, t := range tunes {
		_, ok := GetModificationDate(t.File)
		if !ok {
			for _, tbl := range ref_tables {
				fmt.Fprintf(purge, "delete from %s where d_tune_id = %d;\n", tbl, t.ID)
			}
			fmt.Fprintf(purge, "delete from d_tunes where id = %d ;\n", t.ID)
			fmt.Fprintf(fpurge, "rm \"%s\"\n", t.Img)
			fmt.Fprintf(fpurge, "rm \"%s\"\n", t.Xml)
		}
	}
	purge.Close()
	fpurge.Close()
}
