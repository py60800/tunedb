package zdb

import (
	"fmt"
	"os"
	"path"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	FileTypeMscz int = 0
	FileTypeAbc  int = 1
)

type TuneDB struct {
	cnx *gorm.DB
}

type ExtLink struct {
	ID      int `gorm:"primaryKey"`
	DTuneID int `gorm:"index"`
	Link    string
	Comment string
}

func (db *TuneDB) Close() {
	// ???
}

type SourceRepository struct {
	ID          int `gorm:"primaryKey"`
	Location    string
	Type        string
	DefaultKind string
	Recurse     bool
}

type Button struct {
	Note int
	Side int
	Row  int
	Rank int
	Pull bool
}

type CButton struct {
	DTuneID int `gorm:"index"`
	Idx     int
	Button
}

// Tina ************************************************************************
func (db *TuneDB) TuneGetButtons(tuneId int) []CButton {
	var buttons []CButton
	db.cnx.Where("d_tune_id = ?", tuneId).Order("idx asc").Find(&buttons)
	return buttons
}
func (db *TuneDB) TuneSaveButtons(tuneId int, buttons []CButton) {
	db.cnx.Where("d_tune_id = ?", tuneId).Delete(&CButton{})
	for i := range buttons {
		buttons[i].DTuneID = tuneId
		buttons[i].Idx = i
	}
	db.cnx.Save(&buttons)
}

// *****************************************************************************
var tuneDB *TuneDB

func TuneDBNew() *TuneDB {
	gormConfig := &gorm.Config{}
	//gormConfig.Logger = logger.Default.LogMode(logger.Info)
	gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	_, dbExists := GetModificationDate(DataBase)

	cnx, err := gorm.Open(sqlite.Open(DataBase), gormConfig)

	if err != nil {
		panic(fmt.Errorf("Open database : %v", err))
	}
	tuneDB = &TuneDB{cnx: cnx}
	tuneDB.SchemaUpdate()
	if !dbExists {
		tuneDB.TuneKindUpdateAll(TuneKindDefaults)
	}
	fmt.Println("Database opened")
	return tuneDB

}

func (db *TuneDB) SchemaUpdate() {
	db.cnx.AutoMigrate(&DTune{})
	db.cnx.AutoMigrate(&TuneSet{})
	db.cnx.AutoMigrate(&TuneInSet{})
	db.cnx.AutoMigrate(&MP3File{})
	db.cnx.AutoMigrate(&SourceRepository{})
	db.cnx.AutoMigrate(&ExtLink{})
	db.cnx.AutoMigrate(&TuneKind{})
	db.cnx.AutoMigrate(&MP3TuneSet{})
	db.cnx.AutoMigrate(&MP3TuneInSet{})
	//	db.cnx.AutoMigrate(&TuneTag{})
	db.cnx.AutoMigrate(&TuneList{})
	db.cnx.AutoMigrate(&TuneListItem{})
	db.cnx.AutoMigrate(&CButton{})
}

// Repositories ****************************************************************
func (db *TuneDB) SourceRepositoryGetAll() []SourceRepository {
	var sr []SourceRepository
	db.cnx.Find(&sr)
	if len(sr) == 0 {
		// Not yet configured => Generate default
		sr = db.SourceRepositoryGenerateDefault()
		db.cnx.Save(&sr)
	}
	return sr
}
func (db *TuneDB) SourceRepositoryUpdateAll(newSrc []SourceRepository) {
	prev := db.SourceRepositoryGetAll()
	prevM := make(map[int]SourceRepository)
	for _, t := range prev {
		prevM[t.ID] = t
	}
	for _, n := range newSrc {
		db.cnx.Save(&n)
		delete(prevM, n.ID)
	}
	for _, el := range prevM {
		db.cnx.Delete(&el)
	}
}
func (db *TuneDB) SourceRepositoryGenerateDefault() []SourceRepository {
	sr := make([]SourceRepository, 0)
	home, _ := os.UserHomeDir()
	home = path.Join(home, "Music")

	sr = append(sr, SourceRepository{Location: path.Join(home, "mp3"), Type: "Mp3", DefaultKind: "-", Recurse: true})
	for _, k := range TuneKindDefaults {
		kind := k.Kind
		sr = append(sr, SourceRepository{Location: path.Join(home, "MuseScore", kind), Type: "Mscz", DefaultKind: kind, Recurse: false})
	}
	return sr
}

// Texternal link **************************************************************
func (db *TuneDB) ExtLinkSave(l *ExtLink) {
	db.cnx.Save(l)
}
func (db *TuneDB) ExtLinkDelete(l *ExtLink) {
	db.cnx.Where("id = ?", l.ID).Delete(l)

}
func (db *TuneDB) ExtLinkGet(tuneID int) []ExtLink {
	var extLink []ExtLink
	db.cnx.Where("d_tune_id = ?", tuneID).Find(&extLink)
	return extLink
}

// *****************************************************************************
var ref_tables = []string{
	"tune_in_sets",
	//	"mp3_refs",
	"ext_links",
	"mp3_tune_in_sets",
	"mp3_tune_in_sets",
	"tune_list_items",
	"c_buttons",
}
