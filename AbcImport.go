// AbcImport
package main

import (
	//	"log"

	"github.com/gotk3/gotk3/gtk"
	//	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zdb"
)

type AbcImporter struct {
	textView *gtk.TextView
	importer *zdb.AbcImporter
}

func (l *AbcImporter) Start() error {
	b, _ := l.textView.GetBuffer()
	start, end := b.GetBounds()
	txt, _ := b.GetText(start, end, true)
	message, err := l.importer.Start(txt)
	if message != "" {
		Message(message)
	}
	return err
}
func (c *ZContext) MkAbcImport() (*AbcImporter, gtk.IWidget) {

	l := &AbcImporter{
		importer: zdb.NewAbcImporter(),
	}

	menuButton, _ := gtk.MenuButtonNew()
	menuButton.SetLabel("ABC Import...")
	popover, _ := gtk.PopoverNew(menuButton)
	menuButton.SetPopover(popover)
	grid, _ := gtk.GridNew()
	grid.SetColumnHomogeneous(true)
	menuButton.Connect("clicked", func() {
		popover.ShowAll()
	})
	popover.Add(grid)
	l.textView, _ = gtk.TextViewNew()
	l.textView.SetSizeRequest(100, 200)
	is := 0
	grid.Attach(l.textView, 0, 0, 6, 6)
	is += 6

	importB := MkButton("MuseScore Import", func() {
		err := l.Start()
		if err != nil {
			Message(err.Error())
		}
		msg, hasDup := l.importer.CheckDuplicates()
		if hasDup {
			msg += "Proceed ?"
			if !MessageConfirm(msg) {
				return
			}
		}
		l.importer.MuseImport()
		popover.Popdown()
	})
	importD := MkButton("Direct Import", func() {
		err := l.Start()
		if err != nil {
			Message(err.Error())
		}
		msg, hasDup := l.importer.CheckDuplicates()
		if hasDup {
			msg += "Proceed ?"
			if !MessageConfirm(msg) {
				return
			}
		}
		l.importer.DirectImport()
		popover.Popdown()
	})
	clearB := MkButton("Clear", func() {
		b, _ := l.textView.GetBuffer()
		b.SetText("")
	})

	cancelB := MkButton("Cancel", func() {
		b, _ := l.textView.GetBuffer()
		b.SetText("")
		popover.Popdown()
	})
	grid.Attach(importB, 0, is, 1, 1)
	grid.Attach(importD, 1, is, 1, 1)
	grid.Attach(clearB, 2, is, 1, 1)
	grid.Attach(cancelB, 3, is, 1, 1)

	return l, menuButton
}
