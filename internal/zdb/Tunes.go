// Tunes.go
package zdb

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

type DTuneReference struct {
	ID            int `gorm:"primaryKey"`
	Title         string
	NiceName      string
	Kind          string
	LastRehearsal time.Time
}

type DTune struct {
	ID       int    `gorm:"primaryKey"`
	File     string `gorm:"index"`
	Date     time.Time
	FileDate time.Time
	Hide     bool
	//
	FileType int
	XRef     int
	AbcHash  string

	Img           string
	Xml           string `gorm:"index"`
	NiceName      string
	Kind          string
	Title         string
	Fun           FunLevel
	Play          PlayLevel
	Flute         int
	LastRehearsal time.Time
	Comment       string
	Tempo         int

	Instrument      string
	VelocityPattern string
	SwingPattern    string
	FirstNote       string
	Mode            string
	Fifth           int
	BreathnachCode  string `gorm:"index"`

	Lists []TuneListBase `gorm:"-:all"`
}

type TuneKind struct {
	ID    int `gorm:"primaryKey"`
	Kind  string
	Tempo int
}

// Tune Kind *******************************************************************
func (db *TuneDB) TuneKindGetAllNames() []string {
	var res []string
	db.cnx.Raw("select kind from tune_kinds order by kind asc").Scan(&res)
	return res
}

func (db *TuneDB) TuneKindGetAll() []TuneKind {
	var tuneKinds []TuneKind
	db.cnx.Find(&tuneKinds)
	return tuneKinds
}
func (db *TuneDB) TuneKindUpdateAll(tk []TuneKind) {
	var currTK []int
	db.cnx.Raw("select id from tune_kinds").Scan(&currTK)
	update := make(map[int]int)
	for _, k := range tk {
		update[k.ID] = 0
		db.cnx.Save(&k)
	}
	for _, id := range currTK {
		if _, ok := update[id]; !ok {
			db.cnx.Where("id = ?", id).Delete(&TuneKind{})
		}
	}
}
func (db *TuneDB) TuneKindGetTempo(kind string) int {
	type Tempo struct {
		Tempo int
	}
	t := Tempo{120}
	db.cnx.Model(&TuneKind{}).Where("kind = ?", kind).First(&t)
	return t.Tempo
}
func TuneSort(tunes []DTuneReference, sortMethod string) {
	var sorter func(i, j int) bool
	switch sortMethod {
	case "Practice Date":
		sorter = func(i, j int) bool {
			return tunes[i].LastRehearsal.Before(tunes[j].LastRehearsal)
		}
	case "Practice Date (Inv)":
		sorter = func(i, j int) bool {
			return tunes[i].LastRehearsal.After(tunes[j].LastRehearsal)
		}
	case "Date":
		sorter = func(i, j int) bool {
			return tunes[i].ID < tunes[j].ID
		}
	default:
		fallthrough
	case "Date (Inv)":
		sorter = func(i, j int) bool {
			return tunes[i].ID >= tunes[j].ID
		}
	case "Name (Inv)":
		sorter = func(i, j int) bool {
			return tunes[i].NiceName > tunes[j].NiceName
		}
	case "Name":
		sorter = func(i, j int) bool {
			return tunes[i].NiceName < tunes[j].NiceName
		}
	case "Random":
		sorter = nil

	}
	if sorter != nil {
		sort.Slice(tunes, sorter)
	} else {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(tunes), func(i, j int) {
			tunes[i], tunes[j] = tunes[j], tunes[i]
		})
	}
}
func (db *TuneDB) TuneGetSimilar(tune *DTune) *DTune {
	var tunes []DTune
	r := db.cnx.Where("breathnach_code = ?", tune.BreathnachCode).Find(&tunes)
	warnOnDbError(r)
	log.Printf("[%v] %v => %d \n", tune.Title, tune.BreathnachCode, len(tunes))
	for i, t := range tunes {
		if t.ID == tune.ID {
			return &tunes[(i+1)%len(tunes)]
		}
	}
	return tune
}
func (db *TuneDB) TuneSearch(filter Filter) []DTuneReference {
	wdb := db.cnx
	if filter.List != 0 {
		wdb = wdb.Joins("inner join tune_list_items on tune_list_id = ? and d_tune_id = d_tunes.id", filter.List)
	}
	if filter.Kind != "" && filter.Kind != "*" {
		wdb = wdb.Where("kind = ?", filter.Kind)
	}
	if filter.PartialName != "" {
		pn := strings.ReplaceAll(filter.PartialName, " ", "%")
		wdb = wdb.Where("nice_name LIKE ?", "%"+pn+"%")
	}
	if filter.Fifth >= -4 && filter.Fifth <= 4 {
		wdb = wdb.Where("fifth = ?", filter.Fifth)
	}

	// Play level
	if filter.PlayLevelFrom > 0 {
		wdb = wdb.Where("play >= ?", filter.PlayLevelFrom)
	}
	if filter.PlayLevelTo < PlayLevelMax-1 {
		wdb = wdb.Where("play <= ?", filter.PlayLevelTo)
	}

	if filter.FunLevelFrom > 0 {
		wdb = wdb.Where("fun >= ?", filter.FunLevelFrom)
	}
	if filter.FunLevelTo < FunLevelMax-1 {
		wdb = wdb.Where("fun <= ?", filter.FunLevelTo)
	}
	if filter.Mode != "*" {
		wdb = wdb.Where("mode = ?", filter.Mode)
	}

	/*	if filter.LearnPriority != 0 {
		wdb = wdb.Where("learn_priority & ?", filter.LearnPriority)
	}*/
	if filter.FirstNote != "" {
		wdb = wdb.Where("first_note = ?", filter.FirstNote)
	}

	if filter.RehearsalTo > time.Second {
		date := time.Now().Add(-filter.RehearsalTo)
		wdb = wdb.Where("last_rehearsal < ?", date)
	}
	if filter.RehearsalFrom > filter.RehearsalTo && filter.RehearsalFrom < (100000-10)*time.Hour {
		date := time.Now().Add(-filter.RehearsalFrom)
		wdb = wdb.Where("last_rehearsal > ?", date)
	}

	// Not hidden
	if !filter.IncludeHidden {
		wdb = wdb.Where("(hide is null) or (hide <> 1)")
	}
	var tunes []DTuneReference
	wdb = wdb.Model(&DTune{}).Find(&tunes)
	warnOnDbError(wdb)

	if filter.List == 0 {
		TuneSort(tunes, filter.SortMethod)
	} else {
		listItems := db.GetTuneListByID(filter.List)
		tuneId2rank := make(map[int]int)
		for i, t := range tunes {
			tuneId2rank[t.ID] = i
		}
		ntunes := make([]DTuneReference, 0, len(listItems.Tunes))
		for _, l := range listItems.Tunes {
			if tid, ok := tuneId2rank[l.DTuneID]; ok {
				ntunes = append(ntunes, tunes[tid])
			}
		}
		return ntunes
	}
	return tunes
}

type TuneFile struct {
	File     string
	FileDate time.Time
}

func (db *TuneDB) TuneGetAll() []DTune {
	var tunes []DTune
	r := db.cnx.Find(&tunes)
	warnOnDbError(r)
	return tunes
}

func (db *TuneDB) TunesGetFileList() []TuneFile {
	var tunes []TuneFile
	r := db.cnx.Model(&DTune{}).Scan(&tunes)
	warnOnDbError(r)
	return tunes
}
func (db *TuneDB) TuneSearchM(filter map[string]any) []DTuneReference {
	wdb := db.cnx
	includeHidden := false
	sort := ""
	list := ""
	for k, v := range filter {
		switch k {
		case "PartialName":
			pn := strings.ReplaceAll(v.(string), " ", "%")
			wdb = wdb.Where("nice_name like ?", "%"+pn+"%")
		case "Kind":
			wdb = wdb.Where("kind = ?", v.(string))
		case "Fifth":
			wdb = wdb.Where("fifth = ?", v.(int))
		case "PlayLevelFrom":
			wdb = wdb.Where("play >= ?", v.(int))
		case "PlayLevelTo":
			wdb = wdb.Where("play <= ?", v.(int))
		case "FunLevelFrom":
			wdb = wdb.Where("fun >= ?", v.(int))
		case "FunLevelTo":
			wdb = wdb.Where("fun <= ?", v.(int))
		case "Mode":
			wdb = wdb.Where("mode = ?", v.(string))
		case "First":
			wdb = wdb.Where("first_note = ?", v.(string))
		case "Hidden":
			includeHidden = true
		case "PracticeTo":
			if practiceTo, ok := v.(time.Duration); ok && practiceTo > time.Second {
				date := time.Now().Add(-practiceTo)
				wdb = wdb.Where("last_rehearsal < ?", date)
			}
		case "PracticeFrom":
			if practiceFrom, ok := v.(time.Duration); ok && practiceFrom < (100000*time.Hour) {
				date := time.Now().Add(-practiceFrom)
				wdb = wdb.Where("last_rehearsal > ?", date)
			}
		case "Sort":
			sort = v.(string)
		case "List":
			list = v.(string)

		}

	}
	if !includeHidden {
		wdb = wdb.Where("(hide is null) or (hide <> 1)")
	}

	if list != "" {
		id := db.TuneListGetId(list)
		wdb = wdb.Joins("inner join tune_list_items on tune_list_id = ? and d_tune_id = d_tunes.id", id)
		sort = ""
	}

	var tunes []DTuneReference
	wdb = wdb.Model(&DTune{}).Find(&tunes)
	warnOnDbError(wdb)
	if sort != "" {
		TuneSort(tunes, sort)
	}

	return tunes
}

func (db *TuneDB) TuneSearchByIds(ids []int) []DTuneReference {
	tunes := make([]DTuneReference, len(ids))
	for i, id := range ids {
		db.cnx.Model(&DTune{}).Where("ID = ?", id).Find(&tunes[i])
	}
	return tunes

}

func (db *TuneDB) TuneMemoRehearsal(tune *DTune) {
	tune.LastRehearsal = time.Now()
	db.cnx.Model(&tune).Updates(DTune{Play: tune.Play, LastRehearsal: tune.LastRehearsal})
}

func (db *TuneDB) TuneHideUpdate(tune *DTune) {
	db.cnx.Model(&tune).Update("hide", tune.Hide)
}
func (db *TuneDB) TuneFlUpdate(tune *DTune) {
	db.cnx.Model(&tune).Update("flute", tune.Flute)
}
func (db *TuneDB) TuneFieldUpdate(tune *DTune, field string, value interface{}) {
	db.cnx.Model(&tune).Update(field, value)
}
func (db *TuneDB) TuneMidiPlayUpdate(tune *DTune) {
	db.cnx.Model(&tune).Updates(
		DTune{Tempo: tune.Tempo,
			Instrument:      tune.Instrument,
			VelocityPattern: tune.VelocityPattern,
			SwingPattern:    tune.SwingPattern})
}

// Misc ************************************************************************
func (db *TuneDB) GetTuneModes() []string {
	var mode []string
	db.cnx.Raw("select mode from d_tunes group by mode order by mode asc").Scan(&mode)
	return mode
}

// DTune ***********************************************************************
func (db *TuneDB) TuneGetByID(id int) DTune {
	var tune DTune
	tune.ID = id
	db.cnx.Find(&tune)
	r := db.cnx.Model(&TuneList{}).Joins("inner join tune_list_items on tune_list_id = tune_lists.id").Where("d_tune_id = ?", id).Find(&tune.Lists)
	warnOnDbError(r)

	return tune
}
func (db *TuneDB) TuneGetByXmlFile(xmlFile string) DTune {
	var tune DTune
	res := db.cnx.Where("xml = ?", xmlFile).Find(&tune)
	warnOnDbError(res)
	return tune
}

func (db *TuneDB) GetDuplicates(index string) []string {
	type pTune struct {
		ID    int
		Kind  string
		Title string
	}
	var res []pTune
	r := db.cnx.Model(DTune{}).Where("breathnach_code = ?", index).Find(&res)
	warnOnDbError(r)
	result := make([]string, len(res))
	for i, r := range res {
		result[i] = fmt.Sprintf("[%d] %s/%s", r.ID, r.Kind, r.Title)
	}
	return result
}
func (db *TuneDB) TuneIter(tIter func(t *DTune)) {
	rows, _ := db.cnx.Model(&DTune{}).Rows()
	defer rows.Close()
	for rows.Next() {
		var tune DTune
		db.cnx.ScanRows(rows, &tune)
		tIter(&tune)
	}
}

// *****************************************************************************
var SortMethod = []string{
	"Random", "Name", "Name (Inv)", "Date", "Date (Inv)", "Practice Date", "Practice Date (Inv)",
}

/*
	func TuneSort(tunes []DTuneReference, sortMethod string) {
		var sorter func(i, j int) bool
		switch sortMethod {
		case "Practice Date":
			sorter = func(i, j int) bool {
				return tunes[i].LastRehearsal.Before(tunes[j].LastRehearsal)
			}
		case "Practice Date (Inv)":
			sorter = func(i, j int) bool {
				return tunes[i].LastRehearsal.After(tunes[j].LastRehearsal)
			}
		case "Date":
			sorter = func(i, j int) bool {
				return tunes[i].ID < tunes[j].ID
			}
		default:
			fallthrough
		case "Date (Inv)":
			sorter = func(i, j int) bool {
				return tunes[i].ID >= tunes[j].ID
			}
		case "Name (Inv)":
			sorter = func(i, j int) bool {
				return tunes[i].NiceName > tunes[j].NiceName
			}
		case "Name":
			sorter = func(i, j int) bool {
				return tunes[i].NiceName < tunes[j].NiceName
			}
		case "Random":
			sorter = nil

		}
		if sorter != nil {
			sort.Slice(tunes, sorter)
		} else {
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(tunes), func(i, j int) {
				tunes[i], tunes[j] = tunes[j], tunes[i]
			})
		}
	}
*/
func (db *TuneDB) TuneDelete(id int) {
	tune := db.TuneGetByID(id)
	if tune.ID == 0 {
		return
	}
	db.cnx.Where("d_tune_id = ?", id).Delete(&TuneInSet{})
	db.cnx.Where("d_tune_id = ?", id).Delete(&ExtLink{})
	db.cnx.Where("d_tune_id = ?", id).Delete(&MP3TuneInSet{})
	db.cnx.Where("d_tune_id = ?", id).Delete(&TuneListItem{})
	db.cnx.Where("d_tune_id = ?", id).Delete(&CButton{})

	db.cnx.Where("id = ?", id).Delete(&DTune{})

	os.Remove(tune.File)
	os.Remove(tune.Xml)
	os.Remove(tune.Img)
}
