// SetPlayer
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/py60800/tunedb/internal/zdb"

	"github.com/gotk3/gotk3/gtk"
)

type ListMgr struct {
	tuneList        []zdb.TuneListBase
	currentTuneList *zdb.TuneList

	menuButton *gtk.MenuButton

	lName        *gtk.Label
	listSelector *gtk.ComboBoxText

	tuneSelector  *STuneSelector
	popo          *gtk.Popover
	searchEntry   *gtk.SearchEntry
	setCombo      *gtk.ComboBoxText
	suspendChange bool

	tsSelectorIdx []int
	listStore     *WListStore
	name          *gtk.Entry
	tag           *gtk.Entry
	comment       *gtk.TextView
}

func MkListMgr() (*ListMgr, gtk.IWidget) {
	sp := &ListMgr{}
	sp.menuButton, _ = gtk.MenuButtonNew()
	sp.menuButton.SetLabel("Lists...")
	sp.menuButton.Connect("clicked", func() {
		sp.GetAllTunelist()
		sp.fillSelector()
		sp.popo.ShowAll()
	})
	is := 0
	gw := 6
	sp.popo, _ = gtk.PopoverNew(sp.menuButton)
	sp.menuButton.SetPopover(sp.popo)

	mainGrid, _ := gtk.GridNew()

	sp.popo.Add(mainGrid)
	sp.lName, _ = gtk.LabelNew("...")
	mainGrid.Attach(sp.lName, 0, is, gw, 1)
	is++

	sp.listSelector, _ = gtk.ComboBoxTextNew()
	selB := MkButton("Select", func() { // Select a tune set
		idx := sp.listSelector.GetActive()
		if idx >= 0 && idx < len(sp.tuneList) {
			sp.SelectTuneList(&sp.tuneList[idx])
		}
	})
	mainGrid.Attach(sp.listSelector, 0, is, gw-2, 1)
	mainGrid.AttachNextTo(selB, sp.listSelector, gtk.POS_RIGHT, 2, 1)
	is++
	//
	lName, _ := gtk.LabelNew("Name")
	sp.name, _ = gtk.EntryNew()
	x := 0
	mainGrid.Attach(lName, 0, is, 1, 1)
	x++
	mainGrid.Attach(sp.name, x, is, gw-4, 1)
	x += gw - 4
	lTag, _ := gtk.LabelNew("Tag")
	sp.tag, _ = gtk.EntryNew()
	mainGrid.Attach(lTag, x, is, 1, 1)
	x++
	mainGrid.Attach(sp.tag, x, is, 1, 1)
	is++
	lcomment, _ := gtk.FrameNew("Comment")
	sp.comment, _ = gtk.TextViewNew()
	lcomment.Add(sp.comment)
	mainGrid.Attach(lcomment, 0, is, gw, 2)
	is += 2

	var w gtk.IWidget
	sp.tuneSelector, w = STuneSelectorNew(func(ref *zdb.DTuneReference) {
		tune := GetContext().DB.TuneGetByID(ref.ID)
		sp.listStore.InsertM(map[string]any{
			"ID":    tune.ID,
			"_ID":   tune.ID,
			"Title": tune.Title,
			"Kind":  tune.Kind,
			"PlayL": tune.Play.String(),
		})
	})
	mainGrid.Attach(w, 0, is, gw, 1)
	is++
	columns := []IListStoreColumn{
		ListStoreColumnIntNew("ID", 2, 0, 1000000, 1),
		ListStoreColumnTextNew("Kind", 5),
		ListStoreColumnTextNew("Title", 30),
		ListStoreColumnTextNew("PlayL", 5),
	}
	vadj, _ := gtk.AdjustmentNew(0, 0, 100, 1, 0, 0)
	scw, _ := gtk.ScrolledWindowNew(nil, vadj)
	sp.listStore, w = WListStoreNew(scw, columns, false)
	scw.SetSizeRequest(450, 400)

	mainGrid.Attach(w, 0, is, gw, 10)
	sp.listStore.SetActivate(func(data map[string]interface{}) {
		if id, ok := data["ID"].(int); ok {
			GetContext().LoadTuneByID(id, false, true)
		}
	})

	is += 10

	save := MkButton("Save", func() {
		sp.SaveTuneList()
		sp.TuneCtxRefresh()
	})
	apply := MkButton("Save&Apply", func() {
		if sp.SaveTuneList() {
			sp.apply()
			sp.popo.Hide()
		}
		sp.TuneCtxRefresh()
	})
	deDup := MkButton("DeDup", func() {
		sp.DeDup()
	})

	clear := MkButton("Clear", func() {
		sp.clear()

	})
	del := MkButton("Del", func() {
		if sp.currentTuneList != nil && sp.currentTuneList.ID != 0 {
			if len(sp.currentTuneList.Tunes) < 5 || MessageConfirm(fmt.Sprintf("Delete tune list ?")) {
				GetContext().DB.TuneListRemove(sp.currentTuneList)
				zdb.TuneTagUpdate(GetContext().DB)
				sp.TuneCtxRefresh()
			}
		}
		sp.clear()
		sp.UpdateCombo()
	})
	info := MkListInfo(sp)
	mainGrid.Attach(save, 0, is, 1, 1)
	mainGrid.Attach(apply, 1, is, 1, 1)
	mainGrid.Attach(clear, 2, is, 1, 1)
	mainGrid.Attach(deDup, 3, is, 1, 1)
	mainGrid.Attach(del, 4, is, 1, 1)
	mainGrid.Attach(info, 5, is, 1, 1)
	return sp, sp.menuButton

}
func (sp *ListMgr) TuneCtxRefresh() {
	TuneTagUpdated = true
	GetContext().tuneCtx.Refresh()
}
func (sp *ListMgr) fillSelector() {
	sp.suspendChange = true
	defer func() {
		sp.suspendChange = false
	}()
	sp.listSelector.RemoveAll()
	for _, ts := range sp.tuneList {
		txt := fmt.Sprintf("[%d|%s]%s", ts.ID, ts.Tag, ts.Name)
		sp.listSelector.AppendText(txt)
	}
	sp.listSelector.SetActive(0)
}
func (sp *ListMgr) apply() {
	tunes := sp.listStore.GetValues()
	ids := make([]int, len(tunes))
	for i, t := range tunes {
		ids[i] = t["ID"].(int)
	}
	GetContext().tuneSelector.RefreshFromList(ids)
}
func (sp *ListMgr) clear() {
	sp.currentTuneList = &zdb.TuneList{}
	sp.lName.SetLabel("...")
	sp.name.SetText("")
	sp.listStore.Clear()
}
func (sp *ListMgr) SelectTuneList(ts *zdb.TuneListBase) {
	sp.currentTuneList = GetContext().DB.GetTuneListByID(ts.ID)
	if sp.currentTuneList == nil {
		Message(fmt.Sprintf("[%v] not found", ts.Name))
	}
	sp.lName.SetLabel(ts.Name)
	sp.name.SetText(ts.Name)
	sp.tag.SetText(ts.Tag)
	b, _ := sp.comment.GetBuffer()
	b.SetText(ts.Comment)

	ls := sp.listStore
	ls.Clear()
	for _, tis := range sp.currentTuneList.Tunes {
		tune := GetContext().DB.TuneGetByID(tis.DTuneID)
		if tune.ID != 0 {

			sp.listStore.AppendM(map[string]any{
				"_ID":      tis.DTuneID,
				"_changed": false,
				"ID":       tune.ID,
				"Title":    tune.Title,
				"Kind":     tune.Kind,
				"PlayL":    tune.Play.String(),
			})
		}
	}
}
func (sp *ListMgr) GetAllTunelist() {
	sp.tuneList = GetContext().DB.TuneListGetAll()
}
func (sp *ListMgr) UpdateCombo() {
	sp.GetAllTunelist()
	sp.fillSelector()
}
func (sp *ListMgr) DeDup() {
	tunes := sp.listStore.GetValues()
	found := make(map[int]int)
	for i := 0; i < len(tunes); {
		if _, ok := found[tunes[i]["ID"].(int)]; ok {
			// duplicate
			if i == len(tunes)-1 {
				tunes = tunes[:i]
				break
			} else {
				tunes = append(tunes[:i], tunes[i+1:]...)
			}
		} else {
			found[tunes[i]["ID"].(int)] = 0
			i++
		}
	}
	sp.listStore.Clear()
	for _, t := range tunes {
		sp.listStore.AppendM(t)
	}
}
func (sp *ListMgr) SaveTuneList() bool {
	name, _ := sp.name.GetText()
	if name == "" {
		Message("Name Required")
		return false
	}
	if sp.currentTuneList == nil {
		sp.currentTuneList = &zdb.TuneList{}
	}
	tag, _ := sp.tag.GetText()
	if len(tag) >= 3 {
		tag = tag[0:3]
	}

	b, _ := sp.comment.GetBuffer()
	start, end := b.GetBounds()
	comment, _ := b.GetText(start, end, true)

	if tag != "" && tag != sp.currentTuneList.Tag {
		tags := GetContext().DB.TuneListTags()
		for _, t := range tags {
			if t.Tag == tag {
				Message("Duplicate tag:" + tag)
				tag = ""
				break
			}
		}
		if len(tags) >= 12 {
			Message("Too many tags")
			tag = ""
		}
	}

	sp.currentTuneList.Tag = tag
	sp.currentTuneList.Comment = comment
	sp.currentTuneList.Name = name

	tunes := sp.listStore.GetValues()
	if sp.currentTuneList == nil || sp.currentTuneList.ID == 0 {
		sp.currentTuneList = &zdb.TuneList{}
	}
	sp.currentTuneList.Tunes = make([]zdb.TuneListItem, len(tunes))
	sp.currentTuneList.Name, _ = sp.name.GetText()

	for i := range tunes {
		t := tunes[i]
		sp.currentTuneList.Tunes[i] = zdb.TuneListItem{
			Rank:    i,
			DTuneID: t["ID"].(int),
		}
	}
	GetContext().DB.TuneListSave(sp.currentTuneList)
	zdb.TuneTagUpdate(GetContext().DB)

	sp.UpdateCombo()
	return true
	//	sp.selectTuneSet(sp.currentTuneSet.Name)
}

func MkListInfo(sp *ListMgr) gtk.IWidget {
	b, _ := gtk.MenuButtonNew()
	b.SetLabel("Info...")
	popo, _ := gtk.PopoverNew(b)
	entry, _ := gtk.TextViewNew()
	popo.Add(entry)
	b.SetPopover(popo)
	b.Connect("clicked", func() {

		buffer, _ := entry.GetBuffer()
		tunes := sp.listStore.GetValues()
		stats := make(map[string]int)
		for _, t := range tunes {
			pl := t["PlayL"].(string)
			stats[pl]++
		}
		var s = &strings.Builder{}
		sk := make([]string, 0, len(stats))
		for k := range stats {
			sk = append(sk, k)
		}
		sort.Strings(sk)
		cpt := len(tunes)
		fmt.Fprintf(s, "%d Tunes\n", cpt)
		for _, k := range sk {
			n := stats[k]
			fmt.Fprintf(s, "%-50s : %3d | %3d %2d%%\n", k, n, cpt, (cpt*100)/len(tunes))
			cpt -= n
		}
		buffer.SetText(s.String())
		entry.Show()
	})
	return b
}
