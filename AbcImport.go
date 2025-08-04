// AbcImport
package main

import (
	"github.com/py60800/tunedb/internal/zdb"

	"github.com/gotk3/gotk3/gtk"
)

type AbcImporter struct {
}

func (c *ZContext) MkAbcImport() (*AbcImporter, gtk.IWidget) {

	l := &AbcImporter{}

	menuButton, _ := gtk.MenuButtonNew()
	menuButton.SetLabel("AbcImport...")
	popover, _ := gtk.PopoverNew(menuButton)
	menuButton.SetPopover(popover)
	grid, _ := gtk.GridNew()
	grid.SetColumnHomogeneous(true)
	menuButton.Connect("clicked", func() {
		popover.ShowAll()
	})
	popover.Add(grid)
	textView, _ := gtk.TextViewNew()
	textView.SetSizeRequest(100, 200)
	is := 0
	grid.Attach(textView, 0, 0, 8, 6)
	is += 8
	getText := func() string {
		b, _ := textView.GetBuffer()
		start, end := b.GetBounds()
		txt, _ := b.GetText(start, end, true)
		return txt
	}
	importB, _ := gtk.ButtonNewWithLabel("MuseScore Import")
	importB.Connect("clicked", func() {
		msg, _ := zdb.AbcImport(getText(), false)
		if msg != "" {
			Message(msg)
		}
		popover.Popdown()
	})
	importD, _ := gtk.ButtonNewWithLabel("Direct Import")
	importD.Connect("clicked", func() {
		msg, err := zdb.AbcImport(getText(), true)
		if msg != "" {
			Message(msg)
		}
		if err == nil {
			c.TUpdate()
			popover.Popdown()
		}
	})
	clearB, _ := gtk.ButtonNewWithLabel("Clear")
	clearB.Connect("clicked", func() {
		b, _ := textView.GetBuffer()
		b.SetText("")
	})
	cancelB, _ := gtk.ButtonNewWithLabel("Cancel")
	cancelB.Connect("clicked", func() {
		popover.Popdown()
	})
	grid.Attach(importB, 0, is, 2, 1)
	grid.Attach(importD, 2, is, 2, 1)
	grid.Attach(clearB, 4, is, 1, 1)
	grid.Attach(cancelB, 5, is, 1, 1)

	return l, menuButton
}
