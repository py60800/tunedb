// ConcertinaTab
package main

import (
	"github.com/py60800/tunedb/internal/svgtab"
	"github.com/py60800/tunedb/internal/zdb"
	"github.com/py60800/tunedb/internal/zixml"

	"log"

	"github.com/gotk3/gotk3/gtk"
)

func BConv(b zdb.Button) svgtab.Button {
	return svgtab.Button{
		Row:  b.Row,
		Rank: b.Rank,
		Pull: b.Pull,
		Side: b.Side,
	}
}
func BConv2(b svgtab.Button) zdb.Button {
	return zdb.Button{
		Row:  b.Row,
		Rank: b.Rank,
		Pull: b.Pull,
		Side: b.Side,
	}
}

// *****************************************************************************
type ConcertinaCtrl struct {
	win         *gtk.Window
	ctrlButton  *gtk.ToggleButton
	needRefresh bool
}

func (cc *ConcertinaCtrl) Hide() {
	cc.ctrlButton.SetActive(false)
	ctx := Context()
	ctx.svgt = nil
	ctx.Image.Refresh()
	cc.win.Hide()
}
func (cc *ConcertinaCtrl) Quit() {
	if Context().svgt.Changed {
		if MessageConfirm("Save changes ?") {
			cc.SaveFull()
		}
	}
	log.Println("Need Refresh:", cc.needRefresh)
	if cc.needRefresh {
		Context().RefreshTune()
	}
	cc.Hide()

}
func (cc *ConcertinaCtrl) Show(c *ZContext) {
	if tune := c.ActiveTune; tune != nil && tune.ID != 0 {
		c.svgt = svgtab.SvgTabNew(ConfigBase, tune.Img)
		mButtons := c.DB.TuneGetButtons(tune.ID)
		buttons := make([]svgtab.Button, len(mButtons))
		for i := range mButtons {
			buttons[i] = BConv(mButtons[i].Button)
		}
		c.svgt.SetNotes(zixml.GetNoteList(tune.Xml), buttons)
		c.Image.Refresh()
		cc.win.ShowAll()
	}
}
func (cc *ConcertinaCtrl) Save() {
	c := Context()
	if tune := ActiveTune(); tune != nil && tune.ID != 0 {
		if s := c.svgt; s != nil {
			btns := make([]zdb.CButton, len(s.Buttons))
			for i := range btns {
				btns[i].Button = BConv2(s.Buttons[i])
				btns[i].Idx = i
				btns[i].DTuneID = tune.ID
			}
			c.DB.TuneSaveButtons(tune.ID, btns)
			s.Changed = false
		}
	}
}
func (cc *ConcertinaCtrl) SaveFull() {
	cc.Save()

	if tune := ActiveTune(); tune != nil && tune.ID != 0 {
		log.Println("Concertina Save Full:", tune.File)
		if s := Context().svgt; s != nil {
			s.MsczUpdate(tune.File)
		}
	}
	cc.needRefresh = true

}
func (cc *ConcertinaCtrl) FirstFinger() {
	ctx := Context()
	if s := ctx.svgt; s != nil {
		s.FirstFinger()
		ctx.Image.ResetSelection()
		ctx.Image.Refresh()
	}

}
func (cc *ConcertinaCtrl) FirstRow() {
	ctx := Context()
	if s := ctx.svgt; s != nil {
		s.FirstRow()
		ctx.Image.ResetSelection()
		ctx.Image.Refresh()
	}
}
func (cc *ConcertinaCtrl) SecondRow() {
	ctx := Context()
	if s := ctx.svgt; s != nil {
		s.SecondRow()
		ctx.Image.ResetSelection()
		ctx.Image.Refresh()
	}
}

func ConcertinaCtrlNew() (*ConcertinaCtrl, gtk.IWidget) {
	cc := &ConcertinaCtrl{}
	cc.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	cc.win.SetKeepAbove(true)
	cc.win.SetSizeRequest(200, 400)
	cc.win.Connect("delete-event", func() bool {
		cc.Quit()
		return true
	})

	button, _ := gtk.ToggleButtonNewWithLabel("Concertina")
	button.Connect("toggled", func(b *gtk.ToggleButton) {
		if b.GetActive() {
			cc.Show(Context())
			cc.needRefresh = false
		} else {
			cc.Hide()
		}
	})
	cc.ctrlButton = button

	hBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	label, _ := gtk.LabelNew("Concertina Tab")
	selAll := MkButton("Select All", func() {
		Context().Image.SelectAll()
	})
	ff := MkButton("1st Finger", func() {
		cc.FirstFinger()
	})
	fr := MkButton("1st Row", func() {
		cc.FirstRow()
	})
	sr := MkButton("2d Row", func() {
		cc.SecondRow()
	})
	save := MkButton("Save Button", func() {
		log.Println("Concertina Save Button")
		cc.Save()
		/*
			c := Context()
			if tune := ActiveTune(); tune != nil && tune.ID != 0 {
				if s := c.svgt; s != nil {
					btns := make([]zdb.CButton, len(s.Buttons))
					for i := range btns {
						btns[i].Button = BConv2(s.Buttons[i])
						btns[i].Idx = i
						btns[i].DTuneID = tune.ID
					}
					c.DB.TuneSaveButtons(tune.ID, btns)
				}
			}
		*/
	})
	saveFull := MkButton("Save Full", func() {
		cc.SaveFull()
		/*
			c := Context()
			if tune := ActiveTune(); tune != nil && tune.ID != 0 {
			log.Println("Concertina Save Full:",tune.File)
				if s := c.svgt; s != nil {
					btns := make([]zdb.CButton, len(s.Buttons))
					for i := range btns {
						btns[i].Button = BConv2(s.Buttons[i])
						btns[i].Idx = i
						btns[i].DTuneID = tune.ID
					}
					c.DB.TuneSaveButtons(tune.ID, btns)
					s.MsczUpdate(c.ActiveTune.File)
				}
			}
		*/
	})
	done := MkButton("Done", func() {
		cc.Quit()
	})
	sep, _ := gtk.LabelNew("---")
	cleanUp := MkButton("CleanUp", func() {
		c := Context()
		if tune := ActiveTune(); tune != nil && tune.ID != 0 {
			log.Println("Concertina Cleanup")
			if s := c.svgt; s != nil {
				s.MsczCleanUp(c.ActiveTune.File)
			}
			cc.needRefresh = true
		}
	})
	hBox.Add(label)
	hBox.Add(selAll)
	hBox.Add(ff)
	hBox.Add(fr)
	hBox.Add(sr)
	hBox.Add(save)
	hBox.Add(saveFull)
	hBox.Add(sep)
	hBox.Add(cleanUp)
	hBox.Add(done)
	cc.win.Add(hBox)
	return cc, button
}
