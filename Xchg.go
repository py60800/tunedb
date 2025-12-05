// Xchg
package main

import (
	"github.com/gotk3/gotk3/gtk"
	// "github.com/py60800/tunedb/internal/zdb"
	"fmt"
	"strings"
)

type XchgCtrl struct {
}

func (c *ZContext) MkXchgCtrl() (*XchgCtrl, gtk.IWidget) {

	l := &XchgCtrl{}

	menuButton, _ := gtk.MenuButtonNew()
	menuButton.SetLabel("Xchg...")
	popover, _ := gtk.PopoverNew(menuButton)
	menuButton.SetPopover(popover)
	grid, _ := gtk.GridNew()
	menuButton.Connect("clicked", func() {
		popover.ShowAll()
	})
	popover.Add(grid)
	textView, _ := gtk.TextViewNew()
	textView.SetSizeRequest(100, 50)
	is := 0
	grid.Attach(textView, 0, 0, 8, 2)
	is += 2
	importB := MkButton("Import", func() {
		b, _ := textView.GetBuffer()
		start, end := b.GetBounds()
		txt, _ := b.GetText(start, end, true)
		s := strings.Split(txt, ";")
		msg := "Wrong format"
		switch len(s) {
		case 0:
		case 1:
			msg = Context().DB.TuneImport(s[0], "-")
		default:
			msg = Context().DB.TuneImport(s[0], s[1])
		}
		if msg != "" {
			Message(msg)
		}
		popover.Popdown()
	})
	exportB := MkButton("Export", func() {
		if tune := ActiveTune(); tune != nil {
			Context().clip.SetText(fmt.Sprintf("%s;%s", tune.File, tune.Kind))
		}
		popover.Popdown()
	})
	clearB := MkButton("Clear", func() {
		b, _ := textView.GetBuffer()
		b.SetText("")
	})
	cancelB := MkButton("Cancel", func() {
		popover.Popdown()
	})
	grid.Attach(importB, 0, is, 2, 1)
	grid.Attach(exportB, 2, is, 2, 1)
	grid.Attach(clearB, 4, is, 2, 1)
	grid.Attach(cancelB, 6, is, 2, 1)

	return l, menuButton
}
