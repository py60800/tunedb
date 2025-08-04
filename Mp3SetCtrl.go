// Mp3Set
package main

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/py60800/tunedb/internal/player"
	"github.com/py60800/tunedb/internal/zdb"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type Mp3SetConfigurator struct {
	title        *gtk.Entry
	button       *gtk.MenuButton
	popover      *gtk.Popover
	comment      *gtk.TextView
	mp3Selector  *Mp3Selector
	playWidget   *Mp3PlayWidget
	tuneSelector *STuneSelector
	tuneList     *STuneList
	current      *zdb.MP3TuneSet
	selectedFile string
}

type STuneSelector struct {
	search        *gtk.SearchEntry
	combo         *gtk.ComboBoxText
	tuneRef       []zdb.DTuneReference
	suspendChange bool
	frame         *gtk.Frame
}

type STuneList struct {
	tv     *gtk.TreeView
	ls     *gtk.ListStore
	parent *Mp3SetConfigurator
}

func (tl *STuneList) fixNum() {
	if iter, ok := tl.ls.GetIterFirst(); ok {
		for i := 1; ; i++ {
			tl.ls.SetValue(iter, 0, i)
			if !tl.ls.IterNext(iter) {
				break
			}
		}
	}
}

const (
	stNum = iota
	stTitle
	stFrom
	stTo
	stPref
	stId
	stTisId
)

func (tl *STuneList) Append(t *zdb.DTuneReference) {
	iter := tl.ls.Append()
	var title string
	var id int
	if t == nil {
		title = "---"
		id = 0
	} else {
		title = t.Title
		id = t.ID
	}

	tl.ls.SetValue(iter, stTitle, title)
	tl.ls.SetValue(iter, stFrom, "-")
	tl.ls.SetValue(iter, stTo, "-")
	tl.ls.SetValue(iter, stId, id)
	tl.ls.SetValue(iter, stPref, "")
	tl.ls.SetValue(iter, stTisId, 0)
	tl.fixNum()
}
func (tl *STuneList) Reset() {
	tl.ls.Clear()
}
func prefToText(p bool) string {
	if p {
		return "*"
	}
	return ""
}
func (tl *STuneList) AppendT(t *zdb.MP3TuneInSet) {
	tune := GetContext().DB.TuneGetByID(t.DTuneID)
	iter := tl.ls.Append()
	tl.ls.SetValue(iter, stTitle, tune.Title)
	tl.ls.SetValue(iter, stFrom, fmt.Sprintf("%3.1f", t.From))
	tl.ls.SetValue(iter, stTo, fmt.Sprintf("%3.1f", t.To))
	tl.ls.SetValue(iter, stPref, prefToText(t.Pref))
	tl.ls.SetValue(iter, stId, tune.ID)
	tl.ls.SetValue(iter, stTisId, t.ID)

	tl.fixNum()

}
func (tl *STuneList) Update(m []float64) {
	if iter, ok := tl.ls.GetIterFirst(); ok {
		for i := range m {
			tl.ls.SetValue(iter, stFrom, fmt.Sprintf("%3.1f", m[i]))
			if i < len(m) {
				tl.ls.SetValue(iter, stTo, fmt.Sprintf("%3.1f", m[i+1]))
			} else {
				tl.ls.SetValue(iter, stTo, fmt.Sprintf("???"))
			}
			if !tl.ls.IterNext(iter) {
				break
			}
		}
	}
}
func (tl *STuneList) GetTuneId(idx int) int {
	iter, _ := tl.ls.GetIterFirst()
	for i := 0; ; i++ {
		if i == idx {
			return ListStoreGetInt(tl.ls, iter, stId)
		}
		if !tl.ls.IterNext(iter) {
			break
		}
	}
	return 0
}

type slTune struct {
	title string
	from  float64
	to    float64
	id    int
	pref  bool
}

func (tl *STuneList) getTune(iter *gtk.TreeIter) slTune {
	if !tl.ls.IterIsValid(iter) {
		return slTune{}
	}
	sFrom := ListStoreGetString(tl.ls, iter, stFrom)
	sTo := ListStoreGetString(tl.ls, iter, stTo)
	from, err := strconv.ParseFloat(sFrom, 64)
	if err != nil {
		from = 0.0
	}
	to, err := strconv.ParseFloat(sTo, 64)
	if err != nil {
		to = 0.0
	}
	return slTune{
		title: ListStoreGetString(tl.ls, iter, stTitle),
		id:    ListStoreGetInt(tl.ls, iter, stId),
		to:    to,
		from:  from,
		pref:  ListStoreGetString(tl.ls, iter, stPref) == "*",
	}
}
func (tl *STuneList) locateTune(idx int, pos float64) int {
	if iter, ok := tl.ls.GetIterFirst(); ok {
		for i := 0; ; i++ {
			if i+1 == idx {
				t := tl.getTune(iter)
				return t.id
			}
			if !tl.ls.IterNext(iter) {
				break
			}
		}
	}
	return 0
}

func (tl *STuneList) playTune(iter *gtk.TreeIter) {
	tune := tl.getTune(iter)
	if tune.id == 0 {
		return
	}
	GetContext().LoadTuneByID(tune.id, true, false)
	tl.parent.playWidget.play(tune.from, tune.to, player.PMPlayRepeat)
}

func STuneListNew() (*STuneList, gtk.IWidget) {
	tl := &STuneList{}
	frame, _ := gtk.FrameNew("Tune List")
	grid, _ := gtk.GridNew()
	SetMargins(grid, 5, 5)
	frame.Add(grid)
	tl.ls, _ = gtk.ListStoreNew(glib.TYPE_INT, glib.TYPE_STRING, glib.TYPE_STRING,
		glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_INT, glib.TYPE_INT)
	tl.tv, _ = gtk.TreeViewNew()

	cr0, _ := gtk.CellRendererSpinNew()
	cl0, _ := gtk.TreeViewColumnNewWithAttribute("Num", cr0, "text", 0)
	tl.tv.AppendColumn(cl0)

	cr1, _ := gtk.CellRendererTextNew()
	cl1, _ := gtk.TreeViewColumnNewWithAttribute("Title", cr1, "text", 1)
	tl.tv.AppendColumn(cl1)

	cr2, _ := gtk.CellRendererTextNew()
	cl2, _ := gtk.TreeViewColumnNewWithAttribute("From", cr2, "text", 2)
	tl.tv.AppendColumn(cl2)

	cr3, _ := gtk.CellRendererTextNew()
	cl3, _ := gtk.TreeViewColumnNewWithAttribute("To", cr3, "text", 3)
	tl.tv.AppendColumn(cl3)

	cr4, _ := gtk.CellRendererTextNew()
	cl4, _ := gtk.TreeViewColumnNewWithAttribute("Pref", cr4, "text", 4)
	tl.tv.AppendColumn(cl4)

	grid.Attach(tl.tv, 0, 0, 6, 1)
	tl.tv.SetModel(tl.ls)
	tl.tv.SetHExpand(true)
	sel, _ := tl.tv.GetSelection()
	sel.SetMode(gtk.SELECTION_SINGLE)
	dummy := MkButton("Blank", func() {
		tl.Append(nil)
		tl.parent.playWidget.AppendMarker(-1, -1)

	})
	del := MkButton("Delete", func() {
		if _, iter, ok := sel.GetSelected(); ok {
			tl.ls.Remove(iter)
			if tl.parent.playWidget != nil {
				tl.parent.playWidget.RemoveLastMarker()
			}
		}

	})
	Up := MkButton("Up", func() {
		if _, iter, ok := sel.GetSelected(); ok {
			//prev, _ := iter.Copy()
			_, prev, _ := sel.GetSelected()
			if tl.ls.IterPrevious(prev) {
				tl.ls.Swap(iter, prev)
				sel.SelectIter(iter)
			}
			tl.fixNum()
		}
		runtime.GC()
	})
	Pref := MkButton("Pref", func() {
		if _, iter, ok := sel.GetSelected(); ok {
			v, _ := tl.ls.GetValue(iter, stPref)
			st, _ := v.GoValue()
			s := st.(string)
			if s == "*" {
				s = ""
			} else {
				s = "*"
			}
			tl.ls.SetValue(iter, stPref, s)
			tuneId := ListStoreGetInt(tl.ls, iter, stId)
			if tsetId := ListStoreGetInt(tl.ls, iter, stTisId); tsetId != 0 {
				GetContext().DB.Mp3SetPreference(tuneId, tsetId, s == "*")
			}
		}
		runtime.GC()
	})
	tl.tv.Connect("row-activated", func(tv *gtk.TreeView, path *gtk.TreePath) {
		if iter, ok := tl.ls.GetIter(path); ok == nil {
			tl.playTune(iter)
			tl.parent.popover.Hide()
		}
	})
	grid.Attach(dummy, 0, 1, 1, 1)
	grid.Attach(del, 2, 1, 1, 1)
	grid.Attach(Up, 4, 1, 1, 1)
	grid.Attach(Pref, 5, 1, 1, 1)
	return tl, frame

}
func STuneSelectorNew(Append func(r *zdb.DTuneReference)) (*STuneSelector, gtk.IWidget) {
	ts := &STuneSelector{}
	ts.frame, _ = gtk.FrameNew("Tune Selector")
	grid, _ := gtk.GridNew()
	ts.frame.Add(grid)
	SetMargins(grid, 5, 5)
	ts.search = MkDeferedSearchEntry(&ts.suspendChange, func(what string) {
		filter := zdb.FilterNew()
		filter.PartialName = what
		ts.tuneRef = GetContext().DB.TuneSearch(filter)
		if len(ts.tuneRef) > 50 {
			ts.tuneRef = ts.tuneRef[:50]
		}
		ts.combo.RemoveAll()
		for _, t := range ts.tuneRef {
			ts.combo.AppendText(fmt.Sprintf("[%d]%v: %s", t.ID, t.Kind, t.Title))
		}
		ts.combo.SetActive(0)
	})
	ts.combo, _ = gtk.ComboBoxTextNew()
	grid.Attach(ts.search, 0, 0, 6, 1)
	grid.Attach(ts.combo, 0, 1, 4, 1)
	b := MkButton("Add", func() {
		idx := ts.combo.GetActive()
		if idx >= 0 && idx < len(ts.tuneRef) {
			Append(&ts.tuneRef[idx])
		}
	})
	grid.Attach(b, 4, 1, 1, 1)
	curr := MkButton("Cur", func() {
		t := GetContext().ActiveTune
		if t != nil && t.ID != 0 {
			ref := zdb.DTuneReference{
				ID:            t.ID,
				Title:         t.Title,
				NiceName:      t.NiceName,
				Kind:          t.Kind,
				LastRehearsal: t.LastRehearsal,
			}
			Append(&ref)
		}
	})
	grid.Attach(curr, 5, 1, 1, 1)
	ts.search.SetHExpand(true)
	ts.frame.SetHExpand(true)
	return ts, ts.frame

}
func (s *STuneSelector) SetSensitive(v bool) {
	s.frame.SetSensitive(v)
}

// *****************************************************************************
var mp3SetConfigurator *Mp3SetConfigurator

func (m *Mp3SetConfigurator) HandleSignal(idx int, pos float64) {
	id := m.tuneList.locateTune(idx, pos)
	if id > 0 {
		GetContext().LoadTuneByID(id, false, false)
	}
}
func (m *Mp3SetConfigurator) CommentHandler() gtk.IWidget {
	fComment, _ := gtk.FrameNew("Comment")
	m.comment, _ = gtk.TextViewNew()
	fComment.Add(m.comment)
	var commentTimeOut time.Time
	changePending := false
	delayedChange := func(widget *gtk.Widget, frameClock *gdk.FrameClock) bool {
		if time.Now().After(commentTimeOut) {
			b, _ := m.comment.GetBuffer()
			start, end := b.GetBounds()
			m.current.Comment, _ = b.GetText(start, end, true)
			if m.current.ID != 0 {
				GetContext().DB.Mp3TuneSetUpdateComment(m.current.ID, m.current.Comment)
			}
			changePending = false
			return false
		} else {
			return true
		}
	}

	m.comment.Connect("key-press-event", func() {
		commentTimeOut = time.Now().Add(1 * time.Second)
		if !changePending {
			changePending = true
			m.comment.AddTickCallback(delayedChange)
		}
	})
	return fComment
}

func MkMp3SetConfigurator(mainCursor *gtk.DrawingArea) (*Mp3SetConfigurator, gtk.IWidget) {
	m := &Mp3SetConfigurator{}
	m.button, _ = gtk.MenuButtonNew()
	m.button.SetLabel("MP3 Set...")
	popover, _ := gtk.PopoverNew(m.button)
	m.popover = popover
	m.button.SetPopover(popover)
	m.button.Connect("clicked", func() {
		m.mp3Selector.Preselect(GetContext().GetCurrentTuneID())
	})

	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	popover.Add(box)
	box.SetSizeRequest(200, -1)

	m.title, _ = gtk.EntryNew()

	var w gtk.IWidget
	m.mp3Selector, w = MkMp3Selector(func(mp3file *zdb.MP3File) {
		m.SelectMp3(mp3file)
		m.tuneSelector.SetSensitive(true)
	})
	box.Add(w)
	m.tuneSelector, w = STuneSelectorNew(func(t *zdb.DTuneReference) {
		m.tuneList.Append(t)
		m.playWidget.AppendMarker(-1, -1)
	})
	m.tuneSelector.SetSensitive(false)
	box.Add(w)

	m.playWidget, w = Mp3PlayWidgetNew(func() {
		m.tuneList.Update(m.playWidget.GetMarkers())
	}, mainCursor)
	m.playWidget.SetSignal(func(i int, pos float64) {
		m.HandleSignal(i, pos)
	})
	m.playWidget.SetHide(func() {
		popover.Hide()
	})
	box.Add(w)

	m.tuneList, w = STuneListNew()
	m.tuneList.parent = m
	box.Add(w)
	box.Add(m.CommentHandler())

	b, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)

	save := MkButton("Save", func() {
		m.tuneList.Update(m.playWidget.GetMarkers())
		tunes := make([]zdb.MP3TuneInSet, 0)
		ls := m.tuneList.ls
		if iter, ok := ls.GetIterFirst(); ok {
			for {
				t := m.tuneList.getTune(iter)
				if t.id != 0 {
					tunes = append(tunes, zdb.MP3TuneInSet{
						DTuneID: t.id,
						From:    t.from,
						To:      t.to,
						Pref:    t.pref,
						//						Title:   t.title,
					})
				}
				if !ls.IterNext((iter)) {
					break
				}
			}

			m.current.Tunes = tunes
			m.current.From, m.current.To = m.playWidget.GetBounds()
			GetContext().DB.Mp3SetSave(m.current)
			GetContext().mp3Collection.MarkContent(m.current.MP3FileID)
		}

	})
	del := MkButton("Delete", func() {
		if MessageConfirm("Delete Set Configuration ?") {
			id := m.current.ID
			if id != 0 {
				GetContext().DB.Mp3SetDelete(id)
				m.tuneList.Reset()
				m.playWidget.Reset()
			}
		}
	})
	audacity := MkButton("Audacity", func() {
		zdb.LaunchAudacity(m.selectedFile)
	})
	print := MkButton("Print...", func() {
		files := make([]string, 0)
		ls := m.tuneList.ls
		if iter, ok := ls.GetIterFirst(); ok {
			for {
				t := m.tuneList.getTune(iter)
				if t.id != 0 {
					tune := GetContext().DB.TuneGetByID(t.id)
					files = append(files, tune.Img)
				}
				if !ls.IterNext((iter)) {
					break
				}
			}
		}
		if len(files) > 0 {
			GetContext().printer.Run(files)
		}

	})

	b.Add(save)
	b.Add(audacity)
	b.Add(print)
	b.Add(del)
	box.Add(b)
	close := MkButton("Close", func() {
		popover.Popdown()
	})
	box.Add(close)
	box.ShowAll()
	return m, m.button
}
func (m *Mp3SetConfigurator) ShowCount(id int) {
	l := GetContext().DB.Mp3SetGetByTuneID(id)
	m.button.SetLabel(fmt.Sprintf("[%d]MP3...", len(l)))
}

func (m *Mp3SetConfigurator) PlayDefault() {
	if id := GetContext().GetCurrentTuneID(); id != 0 {
		l := GetContext().DB.Mp3SetGetByTuneID(id)
		if len(l) > 0 {
			tuneSet := GetContext().DB.Mp3SetGetBySetID(l[0])
			mp3file := GetContext().DB.Mp3FileGetByID(tuneSet.MP3FileID)
			if mp3file != nil {
				m.SelectMp3(mp3file)
				for _, t := range tuneSet.Tunes {
					if t.DTuneID == id {
						m.playWidget.play(t.From, t.To, player.PMPlayRepeat)
						return
					}
				}
			}
		}
	}
}
func (m *Mp3SetConfigurator) SelectMp3(mp3file *zdb.MP3File) {
	m.title.SetText(mp3file.Title)
	m.title.SetEditable(false)
	m.title.SetTooltipText(fmt.Sprintf("Artist:%v\nAlbum:%v\nTitle:%v\nFile:%v",
		mp3file.Artist, mp3file.Album, mp3file.Title, mp3file.File))

	m.current = GetContext().DB.Mp3SetGetBySetID(mp3file.ID)
	if m.current == nil {
		m.current = &zdb.MP3TuneSet{
			MP3FileID: mp3file.ID,
			Tunes:     make([]zdb.MP3TuneInSet, 0),
			From:      0.0,
			To:        0.0,
			Comment:   "",
		}
	}
	b, _ := m.comment.GetBuffer()
	b.SetText(m.current.Comment)
	m.selectedFile = mp3file.File
	m.playWidget.SelectFile(mp3file, m.current.From, m.current.To)
	m.tuneList.Reset()

	m.playWidget.ResetMarker()

	prevTo := 0.0
	for _, t := range m.current.Tunes {
		if t.From-prevTo > 1.0 {
			m.tuneList.Append(nil)
			m.playWidget.AppendMarker(prevTo, t.From)
		}
		m.tuneList.AppendT(&t)
		m.playWidget.AppendMarker(t.From, t.To)
		prevTo = t.To
	}
}
