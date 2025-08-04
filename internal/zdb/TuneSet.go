// TuneSet
package zdb

type TuneInSet struct {
	ID        int `gorm:"primaryKey"`
	TuneSetID int
	DTuneID   int
	Count     int
	Tempo     int
	Rank      int
	Title     string
}
type TuneSet struct {
	Name  string `gorm:"uniqueIndex;not null"`
	ID    int    `gorm:"primaryKey"`
	Tempo int
	Tunes []TuneInSet `gorm:"-:all"`
}

// TuneSet *********************************************************************
func (db *TuneDB) TuneSetRemove(tuneSet *TuneSet) {
	if tuneSet.ID != 0 {
		db.cnx.Where("tune_set_id = ?", tuneSet.ID).Delete(&TuneInSet{})
		db.cnx.Where("id = ?", tuneSet.ID).Delete(tuneSet)
	}
}

func (db *TuneDB) TuneSetSave(tune *TuneSet) {
	r := db.cnx.Save(tune)
	warnOnDbError(r)
	db.cnx.Where("tune_set_id = ?", tune.ID).Delete(&TuneInSet{})
	for i := range tune.Tunes { // Gorm is not perfect
		tune.Tunes[i].TuneSetID = tune.ID
		r := db.cnx.Save(&tune.Tunes[i])
		warnOnDbError(r)
	}
	return
}
func (db *TuneDB) TuneSetGetAll() []TuneSet {
	var ts []TuneSet
	db.cnx.Find(&ts)
	for i := range ts {
		ts[i].Tunes = db.tuneSetGetTuneInSet(ts[i].ID)
	}
	return ts
}
func (db *TuneDB) tuneSetGetTuneInSet(tsID int) []TuneInSet {
	var tis []TuneInSet
	db.cnx.Where("tune_set_id = ?", tsID).Order("Rank").Find(&tis)
	return tis
}
func (db *TuneDB) TuneSetGetForId(tuneID int) []TuneSet {
	var ts []TuneSet
	exec := db.cnx.Joins("inner join tune_in_sets on d_tune_id = ? and tune_in_sets.tune_set_id = tune_sets.id", tuneID).Find(&ts)
	warnOnDbError(exec)
	for i := range ts {
		ts[i].Tunes = db.tuneSetGetTuneInSet(ts[i].ID)
	}
	return ts
}

func (db *TuneDB) TuneSetGetCount(tuneID int) int {
	var count int64
	db.cnx.Model(&TuneInSet{}).Where("d_tune_id = ?", tuneID).Count(&count)
	return int(count)
}
func (ts *TuneSet) ToText() string {
	txt := "[" + ts.Name + "]"
	for i, ti := range ts.Tunes {
		sep := "/"
		if i == 0 {
			sep = ""
		}
		txt = txt + sep + ti.Title
	}
	return txt
}
