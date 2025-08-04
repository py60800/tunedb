// TuneList.go
package zdb

import (
	"sort"
)

type TuneListTag struct {
	Tag string
	ID  int
}
type TuneListBase struct {
	ID      int    `gorm:"primaryKey"`
	Name    string `gorm:"uniqueIndex"`
	Tag     string
	Comment string
}
type TuneList struct {
	TuneListBase
	Tunes []TuneListItem
}
type TuneListItem struct {
	TuneListID int
	Rank       int
	DTuneID    int // Exclusive: Tune  or Tune Set
	TuneSetID  int
}

// Tune lists ******************************************************************
func (db *TuneDB) TuneListGetAll() []TuneListBase {
	var result []TuneListBase
	r := db.cnx.Model(&TuneList{}).Select("id", "name", "tag", "comment").Scan(&result)
	warnOnDbError(r)
	return result
}
func (db *TuneDB) TuneListGetId(name string) int {
	var id int
	r := db.cnx.Model(&TuneList{}).Select("id").Where("name = ?", name).Scan(&id)
	warnOnDbError(r)
	return id
}
func (db *TuneDB) TuneListTags() []TuneListTag {
	var result []TuneListTag
	db.cnx.Model(&TuneList{}).Select("tag", "id").Where("tag <> ''").Order("id asc").Scan(&result)
	warnOnDbError(db.cnx)
	return result
}
func (db *TuneDB) GetTuneListByID(id int) *TuneList {
	var res TuneList
	res.ID = id
	r := db.cnx.Find(&res)
	warnOnDbError(r)
	r = db.cnx.Where("tune_list_id = ?", res.ID).Find(&res.Tunes)
	sort.Slice(res.Tunes, func(i, j int) bool {
		return res.Tunes[i].Rank < res.Tunes[j].Rank
	})
	return &res
}

func (db *TuneDB) TuneListSave(tl *TuneList) {
	if tl.ID != 0 {
		db.cnx.Where("tune_list_id = ?", tl.ID).Delete(&TuneListItem{})
	}
	r := db.cnx.Save(tl)
	warnOnDbError(r)
}
func (db *TuneDB) TuneListRemove(tl *TuneList) {
	if tl.ID != 0 {
		db.cnx.Where("tune_list_id = ?", tl.ID).Delete(&TuneListItem{})
		db.cnx.Where("id = ?", tl.ID).Delete(&TuneList{})
	}
}
func (db *TuneDB) TuneListFastUpdate(tune *DTune, listID int, set bool) {
	if !set {
		t := db.cnx.Where("tune_list_id = ? and d_tune_id = ?", listID, tune.ID).Delete(&TuneListItem{})
		warnOnDbError(t)
	} else {
		rank := -1
		//db.cnx.Raw("select max(rank) from tune_list_items where tune_list_id = ?", listID).Scan(&rank)
		r0 := db.cnx.Model(&TuneListItem{}).Select("max(rank)").Where("tune_list_id = ?", listID).Scan(&rank)
		warnOnDbError(r0)
		it := TuneListItem{DTuneID: tune.ID, Rank: rank + 1, TuneListID: listID}
		r := db.cnx.Create(&it) // Save doesn't work because no ID
		warnOnDbError(r)
	}
}
func (tl *TuneList) RemoveDuplicate() {
	found := make(map[int]int)
	for i := 0; i < len(tl.Tunes); {
		if _, ok := found[tl.Tunes[i].DTuneID]; ok {
			// duplicate
			if i == len(tl.Tunes)-1 {
				tl.Tunes = tl.Tunes[:i]
				break
			} else {
				tl.Tunes = append(tl.Tunes[:i], tl.Tunes[i+1:]...)
			}
		} else {
			i++
		}
	}

}
