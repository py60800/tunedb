// Mp3.go
package zdb

import (
	"fmt"
	"log"
	"sort"
	"time"

	"gorm.io/gorm"
)

type MP3FileBase struct {
	ID       int
	File     string
	FileDate time.Time
}
type MP3File struct {
	ID         int    `gorm:"primaryKey"`
	File       string `gorm:"uniqueIndex;not null"`
	FileDate   time.Time
	Artist     string
	Album      string
	Track      int
	Title      string
	HasContent bool `gorm:"-:all"`
}
type MP3TuneInSet struct {
	ID           int `gorm:"primaryKey"`
	MP3TuneSetID int `gorm:"index"`
	DTuneID      int
	Pref         bool
	From         float64
	To           float64
}
type MP3TuneSet struct {
	ID        int `gorm:"primaryKey"`
	MP3FileID int `gorm:"index"`
	Comment   string
	Tunes     []MP3TuneInSet
	From      float64
	To        float64
}

func (db *TuneDB) Mp3SetPreference(tuneId int, mp3TuneInSetId int, val bool) {
	db.cnx.Model(&MP3TuneInSet{}).Where("d_tune_id = ?", tuneId).Update("pref", false)
	db.cnx.Model(&MP3TuneInSet{}).Where("id = ?", mp3TuneInSetId).Update("pref", val)
}
func (db *TuneDB) Mp3DBStore(data []MP3File) {
	for _, item := range data {
		var rec MP3File
		r := db.cnx.Where("file = ?", item.File).First(&rec)
		if r.Error == gorm.ErrRecordNotFound {
			db.cnx.Create(&item)
		} else {
			rec.Artist = item.Artist
			rec.Album = item.Album
			rec.Track = item.Track
			rec.Title = item.Title
			db.cnx.Save(rec)
		}
	}
}

func (db *TuneDB) Mp3DBStore2(files []FileInfo) {
	start := time.Now()
	var knownMp3 []MP3File
	r := db.cnx.Find(&knownMp3)
	warnOnDbError(r)
	m := make(map[string]int)
	for i := range knownMp3 {
		m[knownMp3[i].File] = i
	}
	log.Println("GetAllMp3:", time.Now().Sub(start))
	update := 0
	new := 0
	for i := range files {
		if idx, ok := m[files[i].Name]; ok {
			if files[i].Date.Unix() != knownMp3[idx].FileDate.Unix() {
				log.Println("Update:", knownMp3[idx].File, knownMp3[idx].FileDate, files[i].Date)
				mp3File, _ := MP3GetTag(files[i].Name, files[i].Date, &knownMp3[idx])
				err := db.cnx.Save(mp3File)
				warnOnDbError(err)
				update++
			}
		} else {
			log.Println("New:", files[i].Name)
			if mp3File, err := MP3GetTag(files[i].Name, files[i].Date, nil); err == nil {
				err := db.cnx.Create(mp3File)
				warnOnDbError(err)
				new++
			}
		}
	}
	log.Println("Mp3:", len(files), update, new)
}
func (db *TuneDB) Mp3DBLoad() []MP3File {
	var lmp3 []MP3File
	db.cnx.Find(&lmp3)
	return lmp3
}

func (db *TuneDB) Mp3FileGetByID(id int) *MP3File {
	var m MP3File
	r := db.cnx.Where("id = ?", id).Find(&m)
	if m.ID != 0 {
		return &m
	}
	warnOnDbError(r)
	return nil
}
func (db *TuneDB) Mp3SetGetByTuneID(i int) []int {
	var ids []int
	r := db.cnx.Raw(`select mp3_files.id from mp3_files 
	            join mp3_tune_sets, mp3_tune_in_sets 
				where mp3_files.id = mp3_tune_sets.mp3_file_id 
				and mp3_tune_in_sets.mp3_tune_set_id = mp3_tune_sets.id 
				and mp3_tune_in_sets.d_tune_id = ? 
				group by mp3_files.id
				order by mp3_tune_in_sets.pref  desc
				`, i).Scan(&ids)
	warnOnDbError(r)
	return ids
}
func (db *TuneDB) Mp3SetGetBySetID(i int) *MP3TuneSet {
	var ts MP3TuneSet
	db.cnx.Where("mp3_file_id = ?", i).First(&ts)
	if ts.ID == 0 {
		log.Println("Mp3SetGetBySetID:", i, "Not found")
		return nil
	}
	var tunes []MP3TuneInSet
	db.cnx.Where("mp3_tune_set_id = ?", ts.ID).Find(&tunes)
	sort.Slice(tunes, func(i, j int) bool {
		return tunes[i].To < tunes[j].To
	})
	fixTuneInSets(tunes)
	ts.Tunes = tunes
	return &ts
}
func (db *TuneDB) Mp3SetSave(ms *MP3TuneSet) {
	if ms.ID != 0 {
		db.cnx.Where("mp3_tune_set_id = ?", ms.ID).Delete(&MP3TuneInSet{})
	}
	db.cnx.Save(ms)
}
func (db *TuneDB) Mp3SetIds() []int {
	var id []int
	db.cnx.Raw("select mp3_file_id from mp3_tune_sets").Scan(&id)
	return id
}
func (db *TuneDB) Mp3SetDelete(id int) {
	db.cnx.Where("id = ?", id).Delete(&MP3TuneSet{})
}
func (db *TuneDB) Mp3TuneSetUpdateComment(id int, comment string) {
	if id != 0 {
		db.cnx.Model(&MP3TuneSet{}).Where("id = ?", id).Update("comment", comment)
	}
}
func fixTuneInSets(tunes []MP3TuneInSet) {
	pTo := 0.0
	for i, t := range tunes {
		if t.From < pTo {
			t.From = pTo
		}
		if t.To < t.From {
			t.To = t.From
		}
		pTo = t.To
		tunes[i] = t
	}
}
func (m *MP3File) ShortName() string {
	hasContent := " "
	if m.HasContent {
		hasContent = "*"
	}
	r := fmt.Sprintf("%s%s/%s", hasContent, m.Artist, m.Title)
	return r[:min(80, len(r))]
}
