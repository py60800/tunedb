// TuneSelector
package main

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zdb"
)

/*
ts.playLevelRS.Reset()
ts.funLevelRS.Reset()
ts.practiceDateRS.Reset()
ts.listFilter = 0
ts.fnEntry.SetText("")
ts.hidden.SetActive(false)

ts.filterKind.SetActive(0)
ts.filterText.SetText("")
ts.resetMode()    X
ts.resetFifth()
ts.resetListSelector() X
*/
type TuneSelector struct {
	TuneRefs []zdb.DTuneReference
	IdxTune  int

	context      *ZContext
	filterKind   *gtk.ComboBoxText
	filterText   *gtk.SearchEntry
	filter       zdb.Filter
	filterBox    *gtk.Box
	countDisplay *gtk.Entry

	sortMethod    *gtk.ComboBoxText
	suspendChange bool

	//lstList    []*gtk.CheckButton
	listFilter int
	fnEntry    *gtk.Entry

	hidden *gtk.CheckButton

	playLevelRS    *RangeSelector
	funLevelRS     *RangeSelector
	practiceDateRS *RangeSelector

	modeSelector  *gtk.ComboBoxText
	fifthSelector *gtk.ComboBoxText
	listSelector  *gtk.ComboBoxText
}

type TSContext struct {
	Filter zdb.Filter
	Index  int
}

var contextFile = "searchCtx.gob"

func (ts *TuneSelector) SaveContext() {
	ctx := TSContext{
		Filter: ts.filter,
		Index:  ts.IdxTune,
	}
	util.GobSave(path.Join("context", contextFile), ctx)
}
func (ts *TuneSelector) RetrieveContext() TSContext {
	var ctx TSContext
	ctx, err := util.GobLoad[TSContext](path.Join("context", contextFile))
	if err != nil {
		ctx.Filter = zdb.FilterNew()
		ctx.Index = 0
	}
	return ctx
}

func (ts *TuneSelector) fillMode() {
	modes := ts.context.DB.GetTuneModes()
	ts.modeSelector.RemoveAll()
	ts.modeSelector.AppendText("*")
	idx := -1
	for i, m := range modes {
		ts.modeSelector.AppendText(m)
		if m == ts.filter.Mode {
			idx = i
		}
	}
	if idx < 0 {
		ts.modeSelector.SetActive(0)
	} else {
		ts.modeSelector.SetActive(idx)
	}
}

func (ts *TuneSelector) fillListSelector() {
	ts.listSelector.RemoveAll()
	ts.listSelector.Append("0", "*")
	tl := GetContext().DB.TuneListGetAll()
	for _, t := range tl {
		ts.listSelector.Append(strconv.Itoa(t.ID), fmt.Sprintf("[%s]%s", t.Tag, t.Name))
	}
	ts.listSelector.SetActiveID(fmt.Sprint(ts.filter.List))

}

func (ts *TuneSelector) fillFifth() {
	ts.fifthSelector.RemoveAll()
	ts.fifthSelector.Append("1000", "*")
	ts.fifthSelector.Append("0", "-")
	for i := 1; i < 4; i++ {
		t := ""
		for j := 0; j < i; j++ {
			t += "#"
		}
		ts.fifthSelector.Append(fmt.Sprint(i), t)
	}
	for i := 1; i < 4; i++ {
		t := ""
		for j := 0; j < i; j++ {
			t += "b"
		}
		ts.fifthSelector.Append(fmt.Sprint(-i), t)
	}
	//ts.fifthSelector.SetActiveID("1000")
	ts.fifthSelector.SetActiveID(fmt.Sprint(ts.filter.Fifth))
}

func (ts *TuneSelector) ChangeFile(d int) {
	if len(ts.TuneRefs) == 0 {
		return
	}
	ts.IdxTune = (ts.IdxTune + d + len(ts.TuneRefs)) % len(ts.TuneRefs)
	tune := ts.context.DB.TuneGetByID(ts.TuneRefs[ts.IdxTune].ID)
	ts.context.LoadTune(&tune, false)
	ts.countDisplay.SetText(fmt.Sprintf("%d/%d", ts.IdxTune+1, len(ts.TuneRefs)))

}

var SortMethod = []string{
	"Random", "Name", "Name (Inv)", "Date", "Date (Inv)", "Practice Date", "Practice Date (Inv)",
}

func (ts *TuneSelector) Refresh(first bool) {
	ts.TuneRefs = ts.context.DB.TuneSearch(ts.filter)
	ts.DoUpdate(first)
}
func (ts *TuneSelector) RefreshFromList(ids []int) {
	ts.TuneRefs = ts.context.DB.TuneSearchByIds(ids)
	ts.DoUpdate(false)
}

func (ts *TuneSelector) DoUpdate(first bool) {
	if len(ts.TuneRefs) > 0 {
		if first {
			ts.IdxTune = ts.IdxTune % len(ts.TuneRefs)
		} else {
			ts.IdxTune = 0
		}
		tune := ts.context.DB.TuneGetByID(ts.TuneRefs[ts.IdxTune].ID)
		ts.context.LoadTune(&tune, false)
		ts.countDisplay.SetText(fmt.Sprintf("%d/%d", ts.IdxTune+1, len(ts.TuneRefs)))

	} else {
		ts.countDisplay.SetText("---")
	}
}

type RangeSelector struct {
	idxMin        int
	idxMax        int
	from          *gtk.ComboBoxText
	to            *gtk.ComboBoxText
	item          []string
	suspendChange bool
}

func (rs *RangeSelector) GetLimits() (int, int) {
	return rs.idxMin, rs.idxMax
}

/*
	func (rs *RangeSelector) Reset() {
		rs.idxMin = 0
		rs.idxMax = len(rs.item) - 1
		rs.display()
	}
*/
func (rs *RangeSelector) display() {
	rs.suspendChange = true
	rs.from.SetActive(rs.idxMin)
	rs.to.SetActive(rs.idxMax)
	rs.suspendChange = false
}

func (rs *RangeSelector) set(_min, _max int) {
	_min = max(0, min(_min, len(rs.item)-1))
	_max = max(_min, min(len(rs.item)-1, _max))
	rs.idxMin = _min
	rs.idxMax = _max

	rs.display()
}
func (ts *TuneSelector) MkRange(label string, item []string, fromP, toP int) (*RangeSelector, gtk.IWidget) {
	rs := &RangeSelector{
		item:   item,
		idxMin: 0,
		idxMax: len(item) - 1,
	}
	rs.suspendChange = true
	frame, _ := gtk.FrameNew(label)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2)
	frame.Add(box)
	lFrom, _ := gtk.LabelNew("From:")
	rs.from, _ = gtk.ComboBoxTextNew()
	rs.from.Connect("changed", func() {
		if rs.suspendChange {
			return
		}
		i := rs.from.GetActive()
		rs.set(i, rs.idxMax)
	})
	lTo, _ := gtk.LabelNew("To:")
	rs.to, _ = gtk.ComboBoxTextNew()

	rs.to.Connect("changed", func() {
		if rs.suspendChange {
			return
		}
		i := rs.to.GetActive()
		rs.set(rs.idxMin, i)
	})
	rs.from.RemoveAll()
	for i := 0; i < len(item); i++ {
		rs.from.AppendText(rs.item[i])
	}

	rs.to.RemoveAll()
	for i := 0; i < len(item); i++ {
		rs.to.AppendText(rs.item[i])
	}

	box.Add(lFrom)
	box.Add(rs.from)
	box.Add(lTo)
	box.Add(rs.to)

	//rs.Reset()
	rs.set(fromP, toP)
	rs.suspendChange = false
	return rs, frame

}
func (ts *TuneSelector) fillTuneKind() {
	ts.filterKind.RemoveAll()
	ts.filterKind.AppendText("*")
	idx := -1
	for i, s := range zdb.TuneKindStr {
		ts.filterKind.AppendText(s)
		if ts.filter.Kind == s {
			idx = i
		}
	}
	ts.filterKind.SetActive(max(idx, 0))
}
func (ts *TuneSelector) fillSortMethod() {
	idx := -1
	for i, t := range SortMethod {
		ts.sortMethod.AppendText(t)
		if ts.filter.SortMethod == t {
			idx = i
		}
	}
	if idx < 0 {
		idx = 4
	}
	ts.sortMethod.SetActive(idx)

}

type rAge struct {
	durationT string
	label     string
	duration  time.Duration
}

var practiceDates = []rAge{rAge{durationT: "0s", label: "Now"}, rAge{durationT: "72h", label: "3 Days ago"},
	rAge{durationT: "168h", label: "1 Week ago"}, rAge{durationT: "672h", label: "1 Month ago"},
	rAge{durationT: "2160h", label: "3 Months ago"}, rAge{durationT: "4360h", label: "6 Months ago"},
	rAge{durationT: "100000h", label: "can't remember"}}

func practiceDateLabels() []string {
	l := make([]string, 0, len(practiceDates))
	for i := range practiceDates {
		l = append(l, practiceDates[i].label)
	}
	return l
}
func pdComp(i, j int) bool {
	return practiceDates[i].duration > practiceDates[j].duration
}
func init() {
	for i := range practiceDates {
		practiceDates[i].duration, _ = time.ParseDuration(practiceDates[i].durationT)
	}
	sort.Slice(practiceDates, pdComp)
}

func practiceDateToInt(a time.Duration) int {
	i := sort.Search(len(practiceDates), func(i int) bool {
		return practiceDates[i].duration <= a
	})
	return i
}
func practiceDateToIntR(a time.Duration) int {
	return practiceDateToInt(a)
}

func (ts *TuneSelector) MkFilter() gtk.IWidget {
	menuB, _ := gtk.MenuButtonNew()
	menuB.SetLabel("Filter...")

	// Popoover --------------------------------------------------------------
	popo, _ := gtk.PopoverNew(menuB)
	menuB.SetPopover(popo)

	filterGrid, _ := gtk.GridNew()
	is := 0

	lKind, _ := gtk.LabelNew("Kind")
	var w gtk.IWidget
	ts.filterKind, _ = gtk.ComboBoxTextNew()
	ts.fillTuneKind()
	ts.filterKind.Connect("changed", func() {
		if !ts.suspendChange {
			ts.filter.Kind = ts.filterKind.GetActiveText()
			ts.Refresh(false)
		}
	})
	filterGrid.Attach(lKind, 0, 0, 3, 1)
	filterGrid.Attach(ts.filterKind, 3, 0, 9, 1)
	is++

	lSort, _ := gtk.LabelNew("Sort")
	ts.sortMethod, _ = gtk.ComboBoxTextNew()
	ts.fillSortMethod()
	ts.sortMethod.Connect("changed", func() {
		ts.Refresh(false)
	})
	filterGrid.Attach(lSort, 0, is, 3, 1)
	filterGrid.Attach(ts.sortMethod, 3, is, 9, 1)
	is++

	lMode, _ := gtk.LabelNew("Mode")
	ts.modeSelector, _ = gtk.ComboBoxTextNew()
	ts.fillMode()
	filterGrid.Attach(lMode, 0, is, 3, 1)
	filterGrid.Attach(ts.modeSelector, 3, is, 9, 1)
	is++

	lFifth, _ := gtk.LabelNew("Fifth")
	ts.fifthSelector, _ = gtk.ComboBoxTextNew()
	ts.fillFifth()
	filterGrid.Attach(lFifth, 0, is, 3, 1)
	filterGrid.Attach(ts.fifthSelector, 3, is, 9, 1)
	is++

	lListS, _ := gtk.LabelNew("List")
	ts.listSelector, _ = gtk.ComboBoxTextNew()
	ts.fillListSelector()
	filterGrid.Attach(lListS, 0, is, 3, 1)
	filterGrid.Attach(ts.listSelector, 3, is, 9, 1)
	is++

	ts.funLevelRS, w = ts.MkRange("Fun Level", zdb.FunLevelStr, ts.filter.FunLevelFrom, ts.filter.FunLevelTo)
	filterGrid.Attach(w, 0, is, 12, 1)
	is++

	ts.playLevelRS, w = ts.MkRange("Play Level", zdb.PlayLevelStr, ts.filter.PlayLevelFrom, ts.filter.PlayLevelTo)
	filterGrid.Attach(w, 0, is, 12, 1)
	is++

	/*	type rAge struct {
				d        string
				l        string
				duration time.Duration
			}
			rhP := []rAge{rAge{d: "0s", l: "Now"}, rAge{d: "72h", l: "3 Days ago"},
				rAge{d: "168h", l: "1 Week ago"}, rAge{d: "672h", l: "1 Month ago"},
				rAge{d: "2160h", l: "3 Months ago"}, rAge{d: "4360h", l: "6 Months ago"},
				rAge{d: "100000h", l: "can't remember"}}
			for i := range rhP {
				rhP[i].duration, _ = time.ParseDuration(rhP[i].d)
			}
			sort.Slice(rhP, func(i, j int) bool {
				return rhP[i].duration > rhP[j].duration
			})

		rhPLabel := make([]string, len(rhP))
		for i, t := range rhP {
			rhPLabel[i] = t.l
		}*/
	ts.practiceDateRS, w = ts.MkRange("Last Practice Date", practiceDateLabels(),
		practiceDateToInt(ts.filter.RehearsalFrom), practiceDateToInt(ts.filter.RehearsalTo))
	filterGrid.Attach(w, 0, is, 12, 1)
	is++

	fn, _ := gtk.LabelNew("First Note:")
	filterGrid.Attach(fn, 0, is, 5, 1)
	ts.fnEntry, _ = gtk.EntryNew()
	ts.fnEntry.SetText(ts.filter.FirstNote)
	filterGrid.Attach(ts.fnEntry, 6, is, 5, 1)
	is++

	bc, _ := gtk.ButtonNewWithLabel("Done")
	bc.Connect("clicked", func() {
		popo.Popdown()
	})

	filterGrid.Attach(bc, 6, is, 5, 1)
	ts.hidden, _ = gtk.CheckButtonNewWithLabel("Hidden")
	ts.hidden.SetActive(ts.filter.IncludeHidden)
	filterGrid.Attach(ts.hidden, 0, is, 4, 1)
	is++
	filterGrid.ShowAll()
	popo.Add(filterGrid)

	popo.Connect("closed", func() {
		ts.filter.Kind = ts.filterKind.GetActiveText()
		ts.filter.PartialName, _ = ts.filterText.GetText()
		ts.filter.PlayLevelFrom, ts.filter.PlayLevelTo = ts.playLevelRS.GetLimits()
		ts.filter.FunLevelFrom, ts.filter.FunLevelTo = ts.funLevelRS.GetLimits()
		d1, d2 := ts.practiceDateRS.GetLimits()
		ts.filter.RehearsalFrom = practiceDates[d1].duration
		ts.filter.RehearsalTo = practiceDates[d2].duration

		//		ts.filter.LearnPriority = ts.listFilter
		ts.filter.FirstNote, _ = ts.fnEntry.GetText()
		ts.filter.IncludeHidden = ts.hidden.GetActive()
		ts.filter.Mode = ts.modeSelector.GetActiveText()
		// fifth
		fifth := ts.fifthSelector.GetActiveID()
		var err error
		ts.filter.Fifth, err = strconv.Atoi(fifth)
		if err != nil {
			ts.filter.Fifth = 999
		}
		ts.filter.SortMethod = ts.sortMethod.GetActiveText()
		ts.filter.List, _ = strconv.Atoi(ts.listSelector.GetActiveID())
		if !ts.suspendChange {
			ts.Refresh(false)
		}
	})
	// Popover End -------------------------------------------------------------
	return menuB
}

func (c *ZContext) MkTuneSelector() (*TuneSelector, gtk.IWidget) {
	ts := &TuneSelector{}
	ts.context = c
	ts.suspendChange = true

	previousContext := ts.RetrieveContext()
	ts.filter = previousContext.Filter
	ts.IdxTune = previousContext.Index

	tsGrid, _ := gtk.GridNew()
	xGrid := 0
	addGrid := func(w gtk.IWidget, sz int) {
		tsGrid.Attach(w, xGrid, 0, sz, 1)
		xGrid += sz
	}

	ts.filterText = MkDeferedSearchEntry(&ts.suspendChange, func(what string) {
		if !ts.suspendChange {
			ts.filter.PartialName = what
			ts.Refresh(false)
		}

	})
	ts.filterText.SetText(ts.filter.PartialName)

	addGrid(ts.filterText, 3)

	wFilter := ts.MkFilter()
	addGrid(wFilter, 2)

	reset, _ := gtk.ButtonNewWithLabel("Reset ")
	reset.Connect("clicked", func() {
		ts.suspendChange = true
		ts.filter = zdb.FilterNew()
		ts.playLevelRS.set(ts.filter.PlayLevelFrom, ts.filter.PlayLevelTo)
		ts.funLevelRS.set(ts.filter.FunLevelFrom, ts.filter.FunLevelTo)
		ts.practiceDateRS.set(practiceDateToInt(ts.filter.RehearsalFrom), practiceDateToIntR(ts.filter.RehearsalTo))
		ts.listFilter = 0
		ts.fnEntry.SetText("")
		ts.hidden.SetActive(false)

		ts.filterKind.SetActive(0)
		ts.filterText.SetText("")
		ts.fillMode()
		ts.fillFifth()
		ts.fillListSelector()
		ts.suspendChange = false

		ts.Refresh(false)

	})
	addGrid(reset, 2)

	ts.countDisplay, _ = gtk.EntryNew()
	addGrid(ts.countDisplay, 1)

	bPrev, _ := gtk.ButtonNewWithLabel("<<")
	bPrev.Connect("clicked", func() {
		ts.context.midiPlayCtrl.Zique().Stop()
		ts.ChangeFile(-1)

	})
	addGrid(bPrev, 1)

	bNext, _ := gtk.ButtonNewWithLabel(">>")
	bNext.Connect("clicked", func() {
		ts.context.midiPlayCtrl.Zique().Stop()
		ts.ChangeFile(+1)
	})
	addGrid(bNext, 1)
	bSim := MkButton("~>~", func() {
		c := GetContext()
		if t := c.ActiveTune; t != nil {
			c.LoadTune(c.DB.TuneGetSimilar(t), true)
		}
	})
	addGrid(bSim, 1)
	ts.suspendChange = false
	return ts, tsGrid
}
