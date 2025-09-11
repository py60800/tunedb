// TuneCtx
package main

import (
	"fmt"

	"time"

	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zdb"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"strings"
)

type TuneCtx struct {
	context *ZContext
	// Tune Ctx
	kind         *gtk.ComboBoxText
	fstNote      *gtk.Entry
	tuneLayout   *gtk.Grid
	playLevel    *gtk.ComboBoxText
	funLevel     *gtk.ComboBoxText
	fluteLevel   *gtk.ComboBoxText
	fifths       *gtk.ComboBoxText
	mode         *gtk.ComboBoxText
	practiceDate *gtk.Entry
	comment      *gtk.TextView
	tuneHide     *gtk.CheckButton
	tagButtons   []*gtk.ToggleButton
	tagGrid      *gtk.Grid

	suspendChange bool
}

func MkTuneInfo(c *ZContext) gtk.IWidget {
	b, _ := gtk.MenuButtonNew()
	b.SetLabel("Info...")
	popo, _ := gtk.PopoverNew(b)
	grid, _ := gtk.GridNew()
	popo.Add(grid)
	entry, _ := gtk.TextViewNew()
	grid.Attach(entry, 0, 0, 8, 8)
	baseReloc := ""
	needReloc := false
	reloc := MkButton("Relocate", func() {
		c := Context()
		c.DB.TuneRelocate(c.ActiveTune.ID, baseReloc)
		c.LoadTuneByID(c.ActiveTune.ID, true, true)
	})
	close := MkButton("Close", func() {
		popo.Hide()
	})
	grid.Attach(reloc, 0, 9, 2, 1)
	grid.Attach(close, 2, 9, 2, 1)
	b.SetPopover(popo)
	b.Connect("clicked", func() {

		buffer, _ := entry.GetBuffer()

		if tune := c.ActiveTune; tune != nil {
			var s = &strings.Builder{}
			fmt.Fprintf(s, "ID:%d\n", tune.ID)
			var fileType string
			switch tune.FileType {
			case zdb.FileTypeAbc:
				fileType = "Abc"
			case zdb.FileTypeMscz:
				fileType = "Mscz"
			default:
				fileType = "Unknown"
			}
			fmt.Fprintf(s, "FileType: %s\n", fileType)
			fmt.Fprintf(s, "Title: %s\n", tune.Title)
			fmt.Fprintf(s, "Date: %v\n", tune.Date)
			fmt.Fprintf(s, "File: %s\n", tune.File)
			fmt.Fprintf(s, "Xml: %s\n", tune.Xml)
			fmt.Fprintf(s, "Img: %s\n", tune.Img)
			fmt.Fprintf(s, "Kind: %v\n", tune.Kind)
			fmt.Fprintf(s, "Fifth: %v\n", tune.Fifth)
			fmt.Fprintf(s, "Mode: %v\n", tune.Mode)
			fmt.Fprintf(s, "Play Level: %v\n", tune.Play)
			fmt.Fprintf(s, "Fun Level: %v\n", tune.Fun)
			fmt.Fprintf(s, "Tempo: %v\n", tune.Tempo)
			fmt.Fprintf(s, "Instrument: %v\n", tune.Instrument)
			fmt.Fprintf(s, "Rehearsal Date: %v\n", tune.LastRehearsal)
			fmt.Fprintf(s, "Fun Level:%v\n", tune.Fun)
			fmt.Fprintf(s, "Hide: %v\n", tune.Hide)
			fmt.Fprintf(s, "Index: %v\n", tune.BreathnachCode)

			buffer.SetText(s.String())
			baseReloc, needReloc = DB().TuneNeedsReloc(tune.File, tune.Kind)
			reloc.SetSensitive(needReloc)
		} else {
			buffer.SetText("No tune!")
		}
		grid.ShowAll()
		popo.Show()

	})
	return b
}
func (tc *TuneCtx) Refresh() {
	tune := Context().ActiveTune
	if tune != nil && tune.ID != 0 {
		tc.LoadTune(tune, true)
	}
}
func (tc *TuneCtx) LoadTune(tune *zdb.DTune, keepPlayContext bool) {
	tc.suspendChange = true
	defer func() {
		tc.suspendChange = false
	}()
	if i, ok := zdb.TuneKindIdx[tune.Kind]; ok {
		tc.kind.SetActive(i + 1)

	} else {
		tc.kind.SetActive(0)
	}
	tc.playLevel.SetActive(int(tune.Play))
	tc.funLevel.SetActive(int(tune.Fun))
	tc.practiceDate.SetText(util.StrTime(tune.LastRehearsal))
	b, _ := tc.comment.GetBuffer()
	b.SetText(tune.Comment)
	tc.context.win.SetTitle(fmt.Sprintf("%v [%v]", tune.Title, tune.ID))

	tc.fstNote.SetText(tune.FirstNote)

	tc.tuneHide.SetActive(tune.Hide)
	tc.fluteLevel.SetActive(tune.Flute)
	tc.fifths.SetActive(zdb.FifthIdx(tune.Fifth))
	// Mode
	tc.mode.RemoveAll()
	modes := zdb.ModesForFifth(zdb.Note(tune.Fifth))
	for _, s := range modes {
		tc.mode.AppendText(s)
	}
	tc.mode.SetActive(zdb.ModeIdx(tune.Mode))

	//Tags
	tc.TagUpdate()

}

var TuneTagUpdated bool

func (tc *TuneCtx) TagUpdate() {
	if TuneTagUpdated {
		tc.UpdateTuneTags()
		TuneTagUpdated = false
	}
	if tune := Context().ActiveTune; tune != nil {
		for i, t := range zdb.TuneTags {
			set := false
			for j := range tune.Lists {
				if t.ID == tune.Lists[j].ID {
					set = true
					break
				}
			}
			tc.tagButtons[i].SetActive(set)
		}
	}
}
func (tc *TuneCtx) UpdateTuneKind() {
	sc := tc.suspendChange
	tc.suspendChange = true
	tc.kind.RemoveAll()
	tc.kind.AppendText("?")
	for _, s := range zdb.TuneKindStr {
		tc.kind.AppendText(s)
	}
	if tune := ActiveTune(); tune != nil {
		if i, ok := zdb.TuneKindIdx[tune.Kind]; ok {
			tc.kind.SetActive(i + 1)
		}
	}

	tc.suspendChange = sc
}
func (tc *TuneCtx) UpdateTuneTags() {
	if len(tc.tagButtons) == 0 {
		tc.tagGrid, _ = gtk.GridNew()
		tc.tagButtons = make([]*gtk.ToggleButton, 12)
		for i := range tc.tagButtons {
			tc.tagButtons[i], _ = gtk.ToggleButtonNewWithLabel("-")
			tc.tagGrid.Attach(tc.tagButtons[i], i%6, i/6, 1, 1)
			tc.tagButtons[i].SetSensitive(false)
			idx := i // Binding for func
			tc.tagButtons[i].Connect("clicked", func(b *gtk.ToggleButton) {
				if tune := ActiveTune(); tune != nil && idx < len(zdb.TuneTags) {
					set := b.GetActive()
					tc.context.DB.TuneListFastUpdate(tune, zdb.TuneTags[idx].ID, set)
				}
			})
		}
	}
	for i, t := range zdb.TuneTags {
		if i >= 12 {
			break
		}
		tc.tagButtons[i].SetLabel(t.Tag)
		tc.tagButtons[i].SetSensitive(true)
	}
	for i := len(zdb.TuneTags); i < 12; i++ {
		tc.tagButtons[i].SetLabel("-")
		tc.tagButtons[i].SetActive(false)
		tc.tagButtons[i].SetSensitive(false)
	}
	TuneTagUpdated = false
}
func (c *ZContext) MkTuneCtx() (*TuneCtx, gtk.IWidget) {
	tc := &TuneCtx{}
	tc.context = c

	tc.tuneLayout, _ = gtk.GridNew()
	g := tc.tuneLayout // Alias

	is := 0
	gw := 6
	// Kind -------------------------
	kind, _ := gtk.LabelNew("Kind")
	tc.kind, _ = gtk.ComboBoxTextNew()
	tc.UpdateTuneKind()

	tc.kind.Connect("changed", func(cb *gtk.ComboBoxText) {
		if tune := ActiveTune(); tune != nil && !tc.suspendChange {
			tune.Kind = tc.kind.GetActiveText()
			tc.context.DB.TuneFieldUpdate(tune, "kind", tune.Kind)
		}
	})
	g.Attach(kind, 0, 0, 1, 1)
	g.AttachNextTo(tc.kind, kind, gtk.POS_RIGHT, gw-1, 1)
	is++

	// Mode
	kLabel, _ := gtk.LabelNew("Fifth")
	tc.fifths, _ = gtk.ComboBoxTextNew()
	for _, s := range zdb.FifthsStr {
		tc.fifths.AppendText(s)
	}
	tc.fifths.Connect("changed", func() {
		if tune := ActiveTune(); tune != nil && !tc.suspendChange {
			tune.Fifth = zdb.FifthIdxR(tc.fifths.GetActive())
			tc.context.DB.TuneFieldUpdate(tune, "fifth", tune.Fifth)
			tc.mode.RemoveAll()
			modes := zdb.ModesForFifth(zdb.Note(tune.Fifth))
			for _, s := range modes {
				tc.mode.AppendText(s)
			}
		}

	})

	mLabel, _ := gtk.LabelNew("Mode")
	tc.mode, _ = gtk.ComboBoxTextNew()
	tc.mode.Connect("changed", func() {
		if tune := ActiveTune(); tune != nil && !tc.suspendChange {
			tune.Mode = tc.mode.GetActiveText()
			tc.context.DB.TuneFieldUpdate(tune, "mode", tune.Mode)
		}
	})

	g.Attach(kLabel, 0, is, 1, 1)
	g.AttachNextTo(tc.fifths, kLabel, gtk.POS_RIGHT, gw-1, 1)
	is++

	g.Attach(mLabel, 0, is, 1, 1)
	g.AttachNextTo(tc.mode, mLabel, gtk.POS_RIGHT, gw-1, 1)
	is++

	// Fun level ----------------------
	lFunLevel, _ := gtk.LabelNew("Fun")
	tc.funLevel, _ = gtk.ComboBoxTextNew()
	for i := zdb.FlUnClassified; i <= zdb.FlGreatFun; i++ {
		tc.funLevel.AppendText(zdb.FunLevel(i).String())
	}
	tc.funLevel.Connect("changed", func(cb *gtk.ComboBoxText) {
		if tune := ActiveTune(); tune != nil && !tc.suspendChange {
			tune.Fun = zdb.FunLevel(cb.GetActive())
			tc.context.DB.TuneFieldUpdate(tune, "fun", tune.Fun)
		}
	})
	g.Attach(lFunLevel, 0, is, 1, 1)
	g.AttachNextTo(tc.funLevel, lFunLevel, gtk.POS_RIGHT, gw-1, 1)
	is++

	// Play level -----------------------
	lPlayLevel, _ := gtk.LabelNew("Level")
	tc.playLevel, _ = gtk.ComboBoxTextNew()
	for i := zdb.PlToLearn; i <= zdb.PlGreat; i++ {
		tc.playLevel.AppendText(zdb.PlayLevel(i).String())
	}
	tc.playLevel.Connect("changed", func(cb *gtk.ComboBoxText) {
		if tune := ActiveTune(); tune != nil && !tc.suspendChange {
			tune.Play = zdb.PlayLevel(cb.GetActive())
			tc.context.DB.TuneFieldUpdate(tune, "play", tune.Play)
		}
	})
	g.Attach(lPlayLevel, 0, is, 1, 1)
	g.AttachNextTo(tc.playLevel, lPlayLevel, gtk.POS_RIGHT, gw-1, 1)
	is++

	// Rehearsa
	practiceFrame, _ := gtk.FrameNew("Last practice")
	pBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2)
	practiceFrame.Add(pBox)
	tc.practiceDate, _ = gtk.EntryNew()
	tc.practiceDate.SetEditable(false)
	reButton, _ := gtk.ButtonNew()
	reButton.SetLabel("Save Pract. Date")
	reButton.Connect("clicked", func() {
		if tune := ActiveTune(); tune != nil && !tc.suspendChange {
			tune.Play = zdb.PlayLevel(tc.playLevel.GetActive())
			tune.Fun = zdb.FunLevel(tc.funLevel.GetActive())
			tc.context.DB.TuneMemoRehearsal(tune)
			tc.practiceDate.SetText(util.StrTime(tune.LastRehearsal))
		}
	})

	pBox.Add(tc.practiceDate)
	pBox.Add(reButton)
	SetMargins(pBox, 3, 3)

	g.Attach(practiceFrame, 0, is, gw, 1)
	is++

	firstNoteLabel, _ := gtk.LabelNew("Fst")
	tc.fstNote, _ = gtk.EntryNew()
	tc.fstNote.SetEditable(false)

	g.Attach(firstNoteLabel, 0, is, 1, 1)
	g.AttachNextTo(tc.fstNote, firstNoteLabel, gtk.POS_RIGHT, 1, 1)

	// No idx++

	// Flute level
	flLabel, _ := gtk.LabelNew("Flute")
	tc.fluteLevel, _ = gtk.ComboBoxTextNew()
	var Fl = [...]string{"No", "Mid", "OK"}
	for _, fl := range Fl {
		tc.fluteLevel.AppendText(fl)
	}
	tc.fluteLevel.Connect("changed", func(cb *gtk.ComboBoxText) {
		if tune := ActiveTune(); tune != nil && !tc.suspendChange {
			fl := cb.GetActive()
			if fl != tune.Flute {
				tune.Flute = fl
				tc.context.DB.TuneFlUpdate(tune)
			}
		}
	})
	g.AttachNextTo(flLabel, tc.fstNote, gtk.POS_RIGHT, 1, 1)
	g.AttachNextTo(tc.fluteLevel, flLabel, gtk.POS_RIGHT, 1, 1)
	is++

	// Comment Area ------------------
	tc.comment, _ = gtk.TextViewNew()
	scWin, _ := gtk.ScrolledWindowNew(nil, nil)
	scWin.Add(tc.comment)
	commentFrame, _ := gtk.FrameNew("Comment")
	scWin.SetVExpand(true)
	commentFrame.Add(scWin)
	//	tc.commentFrame = frame

	var commentTimeOut time.Time
	changePending := false
	delayedChange := func(widget *gtk.Widget, frameClock *gdk.FrameClock) bool {
		if time.Now().After(commentTimeOut) {
			if tune := ActiveTune(); tune != nil {
				b, _ := tc.comment.GetBuffer()
				start, end := b.GetBounds()
				tune.Comment, _ = b.GetText(start, end, true)
				DB().TuneFieldUpdate(tune, "comment", tune.Comment)
				c.MarkOK()
			}
			changePending = false
			return false
		} else {
			return true
		}
	}

	//	tc.comment = textArea
	tc.comment.Connect("key-press-event", func() {
		commentTimeOut = time.Now().Add(1 * time.Second)
		if !changePending {
			changePending = true
			tc.comment.AddTickCallback(delayedChange)
		}
		tc.context.MarkChange()
	})

	commentFrame.SetVExpand(true)

	g.Attach(commentFrame, 0, is, gw, 6)
	is += 6

	tc.tagGrid, _ = gtk.GridNew()
	tc.UpdateTuneTags()
	g.Attach(tc.tagGrid, 0, is, gw, 2) // 2 rows
	is += 2

	info := MkTuneInfo(c)
	g.Attach(info, 0, is, 3, 1)
	deleteTune := MkButton("Delete", func() {
		tune := ActiveTune()
		if tune != nil {
			if MessageConfirm("WARNING : Deleting the tune with delete all related files and references!\n Are you sure ?\n") {
				if MessageConfirm(fmt.Sprintf("Please confirm tune delete of %v", tune.Title)) {
					tc.context.DB.TuneDelete(tune.ID)
					tc.context.TUpdate()
				}
			}
		}
	})
	g.Attach(deleteTune, 3, is, 2, 1)
	is++

	tc.tuneHide, _ = gtk.CheckButtonNewWithLabel("Hide this tune")
	tc.tuneHide.Connect("toggled", func(b *gtk.CheckButton) {
		if tune := ActiveTune(); tune != nil {
			tune.Hide = b.GetActive()
			tc.context.DB.TuneHideUpdate(tune)
		}

	})
	g.Attach(tc.tuneHide, 0, is, 3, 1)
	is++

	g.SetBorderWidth(5)

	return tc, g
}
