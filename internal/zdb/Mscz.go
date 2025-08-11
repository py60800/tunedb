// Mscz.go
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
		log.Println("xml file is missing", updatedTune.Xml) // should panic
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
			os.Mkdir(path.Join(repo.Location, "xml"), 0777)
			os.Mkdir(path.Join(repo.Location, "img"), 0777)
			log.Println("Directory not found:", repo.Location)
			continue
		}

		util.MkSubDirs(repo.Location) // Create directories anyway
		ta := time.Now()
		lMscz := GetFileListR2(repo.Location, ".mscz", repo.Recurse)
		d0 += time.Now().Sub(ta)
		tb := time.Now()
		for _, mscz := range lMscz {
			if idx, ok := m[mscz.Name]; !ok || mscz.Date.Unix() != lTunes[idx].FileDate.Unix() {
				log.Println("Update:", mscz.Name)
				db.MsczTuneSave(mscz.Name, repo.DefaultKind, mscz.Date)
			}
		}
		d1 += time.Now().Sub(tb)
	}
	log.Printf("MszUpdate: T0 : %v D0: %v D1: %v (Total:%v)\n", da, d0, d1, time.Now().Sub(t0))
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

func (db *TuneDB) TuneImport(file string, kind string) string {
	repo, k := guessTarget(kind)
	dst := path.Join(repo, path.Base(file))
	if !strings.HasSuffix(file, ".mscz") {
		return "Wrong file type"
	}
	if _, ok := util.GetModificationDate(dst); ok {
		return dst + "File exists!"
	}
	if err := util.CopyFile(file, dst); err != nil {
		return "Failed to copy file"
	}

	db.MsczTuneSave(dst, k, time.Now())

	return "Import OK"

}
func (db *TuneDB) TuneNeedsReloc_base(repo []SourceRepository, file string, kind string) (string, bool) {
	dir := path.Dir(file)
	aRepo := ""
	for i := range repo {
		if repo[i].Type != "Mscz" {
			continue
		}
		if dir == repo[i].Location && repo[i].DefaultKind == kind {
			return "", false
		}
		if repo[i].DefaultKind == kind && aRepo == "" {
			aRepo = repo[i].Location
		}
	}
	if aRepo != "" {
		return aRepo, true
	}
	return "", false
}

func (db *TuneDB) TuneNeedsReloc(file string, kind string) (string, bool) {
	repo := db.SourceRepositoryGetAll()
	return db.TuneNeedsReloc_base(repo, file, kind)
}

func (db *TuneDB) TuneRelocate(id int, target string) error {
	tune := db.TuneGetByID(id)
	if tune.ID == 0 {
		return fmt.Errorf("No tune")
	}
	newFile := path.Join(target, path.Base(tune.File))
	if err := os.Rename(tune.File, newFile); err != nil {
		return err
	}
	xmlTarget := path.Join(target, "xml")
	imgTarget := path.Join(target, "img")
	if _, err := os.Stat(xmlTarget); err != nil {
		os.MkdirAll(xmlTarget, 0777)
		os.MkdirAll(imgTarget, 0777)
	}
	r := db.cnx.Model(&tune).Update("file", newFile)
	warnOnDbError(r)

	newXml := path.Join(xmlTarget, path.Base(tune.Xml))
	os.Rename(tune.Xml, newXml)
	r = db.cnx.Model(&tune).Update("xml", newXml)
	warnOnDbError(r)

	newImg := path.Join(imgTarget, path.Base(tune.Img))
	os.Rename(tune.Img, newImg)
	r = db.cnx.Model(&tune).Update("img", newImg)
	warnOnDbError(r)
	return db.cnx.Error
}
func (db *TuneDB) MassRelocate(feedback chan string) {
	tunes := db.TuneGetAll()
	repo := db.SourceRepositoryGetAll()
	for _, t := range tunes {
		kloc, needsReloc := db.TuneNeedsReloc_base(repo, t.File, t.Kind)
		if needsReloc && kloc != "" {
			select {
			case feedback <- t.File:
			default:
			}
			db.TuneRelocate(t.ID, kloc)
		}
	}
}
