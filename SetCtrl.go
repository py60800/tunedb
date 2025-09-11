// SetPlayer
package main

import (
	"fmt"

	"time"

	"github.com/py60800/tunedb/internal/search"
	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zdb"
	"github.com/py60800/tunedb/internal/zique"

	"github.com/gotk3/gotk3/gtk"
)

type TuneSetSelector struct {
	nodeSearch    *search.Node
	tsSelector    *gtk.ComboBoxText
	tsSelectorIdx []int
	selectTuneSet func(*zdb.TuneSet)
	searchEntry   *gtk.SearchEntry
	setCombo      *gtk.ComboBoxText
	tuneSets      []zdb.TuneSet
	suspendChange bool
}

func TuneSetSelectorNew(tsSelect func(*zdb.TuneSet)) (gtk.IWidget, *TuneSetSelector) {
	setSelector := &TuneSetSelector{
		selectTuneSet: tsSelect,
	}
	frame, _ := gtk.FrameNew("Set Selector")
	grid, _ := gtk.GridNew()
	frame.Add(grid)

	setSelector.searchEntry = MkDeferedSearchEntry(&setSelector.suspendChange, func(what string) {
		setSelector.UpdateCombo(what)
	})
	is := 0
	gw := 6
	grid.Attach(setSelector.searchEntry, 0, is, gw-2, 1)
	reset := MkButton("Rst", func() {
		setSelector.searchEntry.SetText("")
		setSelector.GetAllTuneSets()

	})

	curB := MkButton("Cur", func() {
		if id := Context().GetCurrentTuneID(); id != 0 {
			setSelector.GetTuneSetsForId(id)
		}
	})
	grid.AttachNextTo(reset, setSelector.searchEntry, gtk.POS_RIGHT, 1, 1)
	grid.AttachNextTo(curB, reset, gtk.POS_RIGHT, 1, 1)

	is++
	setSelector.tsSelector, _ = gtk.ComboBoxTextNew()
	selB := MkButton("Select", func() { // Select a tune set
		idx := setSelector.tsSelector.GetActive()
		if idx >= 0 && idx < len(setSelector.tsSelectorIdx) {
			selectedId := setSelector.tsSelectorIdx[idx]
			if selectedId < len(setSelector.tuneSets) {
				setSelector.selectTuneSet(&setSelector.tuneSets[selectedId])

			}
		}
	})
	grid.Attach(setSelector.tsSelector, 0, is, gw-2, 1)
	grid.AttachNextTo(selB, setSelector.tsSelector, gtk.POS_RIGHT, 2, 1)
	is++

	return frame, setSelector
}
func (setSelector *TuneSetSelector) UpdateCombo(txt string) {
	l := setSelector.nodeSearch.Search(txt)
	setSelector.fillSelector(l)
}
func (setSelector *TuneSetSelector) fillSelector(lIdx []int) {
	setSelector.suspendChange = true
	defer func() {
		setSelector.suspendChange = false
	}()
	setSelector.tsSelector.RemoveAll()

	if len(lIdx) == 0 {
		setSelector.tsSelectorIdx = make([]int, len(setSelector.tuneSets))
		for i := range setSelector.tsSelectorIdx {
			setSelector.tsSelectorIdx[i] = i
		}
	} else {
		setSelector.tsSelectorIdx = lIdx
	}
	for i, n := range setSelector.tsSelectorIdx {
		txt := fmt.Sprint(i, util.STruncate(setSelector.tuneSets[n].ToText(), 60))
		setSelector.tsSelector.AppendText(txt)
	}
	setSelector.tsSelector.SetActive(0)
}
func (setSelector *TuneSetSelector) GetAllTuneSets() {
	setSelector.tuneSets = DB().TuneSetGetAll()
	setSelector.RefreshTuneSetSelector()
}
func (setSelector *TuneSetSelector) GetTuneSetsForId(id int) {
	setSelector.tuneSets = DB().TuneSetGetForId(id)
	setSelector.RefreshTuneSetSelector()
}

func (setSelector *TuneSetSelector) RefreshTuneSetSelector() {
	setSelector.suspendChange = true
	setSelector.tsSelector.RemoveAll()
	setSelector.nodeSearch = search.NodeNew()
	setSelector.fillSelector([]int{})
	for i, ts := range setSelector.tuneSets {
		txt := ts.ToText()
		setSelector.nodeSearch.IndexWords(txt, i)
	}
	setSelector.suspendChange = false
}

// *****************************************************************************
type SetPlayCtrl struct {
	currentTuneSet zdb.TuneSet
	menuButton     *gtk.MenuButton
	setSelector    *TuneSetSelector
	lName          *gtk.Label

	tuneSelector  *STuneSelector
	popo          *gtk.Popover
	suspendChange bool

	listStore *WListStore
	name      *gtk.Entry
	tempo     *gtk.SpinButton
}

func MkSetPlayCtrl() (*SetPlayCtrl, gtk.IWidget) {
	sp := &SetPlayCtrl{}
	sp.menuButton, _ = gtk.MenuButtonNew()
	sp.menuButton.SetLabel("Sets...")

	sp.menuButton.Connect("clicked", func() {
		sp.setSelector.GetAllTuneSets()
		sp.popo.ShowAll()
	})
	is := 0
	gw := 6
	sp.popo, _ = gtk.PopoverNew(sp.menuButton)
	sp.menuButton.SetPopover(sp.popo)

	mainGrid, _ := gtk.GridNew()
	var w gtk.IWidget
	w, sp.setSelector = TuneSetSelectorNew(func(ts *zdb.TuneSet) {
		sp.selectTuneSet(ts)
	})
	sp.popo.Add(mainGrid)
	sp.lName, _ = gtk.LabelNew("...")
	mainGrid.Attach(sp.lName, 0, is, gw, 1)
	is++
	mainGrid.Attach(w, 0, is, gw, 2)
	is += 2
	sp.tuneSelector, w = STuneSelectorNew(func(ref *zdb.DTuneReference) {
		tune := DB().TuneGetByID(ref.ID)
		// 3 : default count
		//sp.listStore.Insert(tune.Title, 3, tune.Tempo, false, tune.ID)
		sp.listStore.InsertM(map[string]any{
			"Title": tune.Title,
			"Tempo": tune.Tempo,
			"Count": 3,
			"_ID":   tune.ID,
		})
	})
	mainGrid.Attach(w, 0, is, gw, 1)
	is++
	columns := []IListStoreColumn{
		ListStoreColumnTextNew("Title", 50),
		ListStoreColumnIntNew("Count", 3, 1, 9, 1),
		ListStoreColumnIntNew("Tempo", 3, 40, 240, 5),
	}
	sp.listStore, w = WListStoreNew(nil, columns, false)
	mainGrid.Attach(w, 0, is, gw, 1)
	is++

	lName, _ := gtk.LabelNew("Name")
	sp.name, _ = gtk.EntryNew()
	sp.tempo, _ = gtk.SpinButtonNewWithRange(0, 240, 5)
	sp.tempo.SetValue(120)
	lTempo, _ := gtk.LabelNew("Tempo")
	x := 0
	mainGrid.Attach(lName, 0, is, 1, 1)
	x++
	mainGrid.Attach(sp.name, x, is, gw-4, 1)
	x += gw - 4
	mainGrid.Attach(lTempo, x, is, 1, 1)
	x++
	mainGrid.Attach(sp.tempo, x, is, 2, 1)
	is++
	play := MkButton("Play", func() {
		sp.play()
	})

	stop := MkButton("Stop", func() {
		Context().Stop()
	})
	mainGrid.Attach(play, 0, is, 2, 1)
	mainGrid.Attach(stop, gw-2, is, 2, 1)
	is++

	save := MkButton("Save", func() {
		sp.SaveCurrentTuneSet()
	})
	clear := MkButton("Clear", func() {
		sp.clear()

	})
	del := MkButton("Del", func() {
		if sp.currentTuneSet.ID != 0 {
			DB().TuneSetRemove(&sp.currentTuneSet)
		}
		sp.clear()
	})
	print := MkButton("Print...", func() {
		files := make([]string, 0)
		for _, t := range sp.currentTuneSet.Tunes {
			tune := DB().TuneGetByID(t.DTuneID)
			files = append(files, tune.Img)
		}
		if len(files) > 0 {
			Context().printer.Run(files)
		}
	})
	mainGrid.Attach(save, 0, is, 2, 1)
	mainGrid.Attach(clear, 2, is, 2, 1)
	mainGrid.Attach(del, gw-2, is, 1, 1)
	mainGrid.Attach(print, gw-1, is, 1, 1)
	return sp, sp.menuButton

}
func (sp *SetPlayCtrl) SetCount(count int) {
	sp.menuButton.SetLabel(fmt.Sprintf("[%d]Sets...", count))
}
func (sp *SetPlayCtrl) clear() {
	sp.currentTuneSet = zdb.TuneSet{}
	sp.lName.SetLabel("...")
	sp.name.SetText("")
	sp.tempo.SetValue(120)
	sp.listStore.Clear()
}
func (sp *SetPlayCtrl) selectTuneSet(ts *zdb.TuneSet) {
	sp.currentTuneSet = *ts
	sp.lName.SetLabel(ts.Name)
	sp.name.SetText(ts.Name)

	sp.tempo.SetValue(float64(ts.Tempo))
	ls := sp.listStore
	ls.Clear()
	for _, tis := range ts.Tunes {
		ls.AppendM(map[string]any{
			"Title":    tis.Title,
			"Count":    tis.Count,
			"Tempo":    tis.Tempo,
			"_ID":      tis.DTuneID,
			"_changed": false,
		})
	}
}
func (sp *SetPlayCtrl) SaveCurrentTuneSet() {
	tunes := sp.listStore.GetValues()
	if len(tunes) == 0 {
		return
	}
	sp.currentTuneSet.Tunes = make([]zdb.TuneInSet, len(tunes))

	sp.currentTuneSet.Tempo = sp.tempo.GetValueAsInt()
	titles := make([]string, 0)
	for i := range sp.currentTuneSet.Tunes {
		t := tunes[i]
		sp.currentTuneSet.Tunes[i] = zdb.TuneInSet{
			Title:   t["Title"].(string),
			Count:   t["Count"].(int),
			Tempo:   t["Tempo"].(int),
			Rank:    i,
			DTuneID: t["_ID"].(int),
		}
		titles = append(titles, t["Title"].(string))
	}
	name, _ := sp.name.GetText()

	if name == "" {
		name = TitleCombine(titles)
	}
	sp.currentTuneSet.Name = name
	DB().TuneSetSave(&sp.currentTuneSet)
	sp.setSelector.GetAllTuneSets()
	//	sp.selectTuneSet(sp.currentTuneSet.Name)
}

func (sp *SetPlayCtrl) play() {
	Context().Stop()
	tunes := sp.listStore.GetValues()
	if len(tunes) == 0 {
		return
	}
	playSet := make([]zique.SetElem, len(tunes))
	firstID := tunes[0]["_ID"].(int)

	for i, t := range tunes {
		Count := t["Count"].(int)
		TuneID := t["_ID"].(int)
		theTune := DB().TuneGetByID(TuneID)
		playSet[i] = zique.SetElem{File: theTune.Xml, Count: Count}
	}
	c := Context()
	c.LoadTuneByID(firstID, true, false)
	if tune := ActiveTune(); tune != nil {
		c.midiPlayCtrl.SetTempo(sp.tempo.GetValueAsInt(), tune.Kind)
	}

	//
	DelayedAction(sp.menuButton, 2*time.Second, func() {
		Context().midiPlayCtrl.Zique().PlaySet(playSet)
		c.metronome.MetronomeShow()
	})
	sp.popo.Popdown()
}
