package main

import (
	"fmt"

	"regexp"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"unicode"

	"github.com/py60800/tunedb/internal/zdb"
)

type ExtLinkT struct {
	zdb.ExtLink
	NotNew    bool
	Changed   bool
	LinkValid bool
	Deleted   bool
}
type ExtLinkCtrl struct {
	menuButton *gtk.MenuButton
	popover    *gtk.Popover
	grid       *gtk.Grid
	clip       *gtk.Clipboard
	noTune     *gtk.Label
	links      []*ExtLinkT
	reCheck    *regexp.Regexp
	tuneID     int
	idx        int
	rowCount   int
}

func (l *ExtLinkCtrl) Widget() gtk.IWidget {
	return l.menuButton
}
func (l *ExtLinkCtrl) Append(linkReference *ExtLinkT) {
	linkEntry, _ := gtk.EntryNew()
	linkEntry.SetEditable(false) // Only copy
	remove, _ := gtk.ButtonNewWithLabel("Delete")

	linkComment, _ := gtk.EntryNew()
	activeLink, _ := gtk.LinkButtonNew("Paste/Follow Link")
	linkEntry.SetText(linkReference.Link)
	linkEntry.Connect("paste-clipboard", func() {
		txt, _ := l.clip.WaitForText()
		if l.reCheck.Match([]byte(txt)) {
			linkReference.Changed = true
			linkReference.LinkValid = true
			linkReference.Link = txt
			linkReference.DTuneID = l.tuneID
			linkComment.SetSensitive(true)

			remove.SetSensitive(true)
			activeLink.SetSensitive(true)
			activeLink.SetUri(txt)
			linkEntry.SetText(txt)
			l.appendDummyLink()
		} else {
			linkEntry.SetText("Paste Link Here")
		}
	})
	tIdx := fmt.Sprint(l.idx)
	linkEntry.SetName(tIdx)
	l.idx++
	l.grid.Attach(linkEntry, 0, l.rowCount, 20, 1)
	l.grid.AttachNextTo(linkComment, linkEntry, gtk.POS_RIGHT, 6, 1)
	linkComment.SetText(linkReference.Comment)
	linkComment.Connect("changed", func() {
		linkReference.Changed = true
		linkReference.Comment, _ = linkComment.GetText()
	})
	activeLink.SetUri(linkReference.Link)
	activeLink.Connect("activate-link", func() bool {

		l.clip.SetText(linkReference.Link)
		return false
	})
	l.grid.AttachNextTo(activeLink, linkComment, gtk.POS_RIGHT, 2, 1)
	if !linkReference.NotNew {
		remove.SetSensitive(false)
		linkComment.SetSensitive(false)
		linkComment.SetText("Description")
		linkEntry.SetText("Paste Link Here")
		activeLink.SetSensitive(false)

	}
	l.grid.AttachNextTo(remove, activeLink, gtk.POS_RIGHT, 2, 1)
	remove.Connect("clicked", func() {
		for idxRemove := 0; idxRemove < l.rowCount; idxRemove++ {
			child, _ := l.grid.GetChildAt(0, idxRemove)
			if child != nil {
				entry := child.(*gtk.Entry)
				name, _ := entry.GetName()
				if name == tIdx {
					l.grid.RemoveRow(idxRemove)
					linkReference.LinkValid = false
					l.rowCount--
					break
				}
			}
		}
	})
	l.rowCount++
	l.grid.ShowAll()
}
func (l *ExtLinkCtrl) appendDummyLink() {
	l.links = append(l.links, &ExtLinkT{})
	l.Append(l.links[len(l.links)-1])
	l.grid.ShowAll()

}
func (l *ExtLinkCtrl) UpdateTuneLinks(c *ZContext, tuneID int) {
	if tuneID == l.tuneID {
		return
	}
	l.tuneID = tuneID
	// Tune ID has changed
	l.links = make([]*ExtLinkT, 0)
	// Cleanup Widgets
	for i := l.rowCount; i >= 0; i-- {
		l.grid.RemoveRow(i)
	}
	l.rowCount = 0
	// Get New
	if tuneID == 0 {
		l.menuButton.SetLabel(fmt.Sprintf("[]XLinks"))
		return
	}
	actualLinks := c.DB.ExtLinkGet(tuneID)
	l.menuButton.SetLabel(fmt.Sprintf("[%d]XLinks", len(actualLinks)))
	for _, lnk := range actualLinks {
		l.links = append(l.links, &ExtLinkT{ExtLink: lnk, NotNew: true, LinkValid: true})
		l.Append(l.links[len(l.links)-1])
	}
	l.appendDummyLink()
}
func (l *ExtLinkCtrl) SaveTuneList(c *ZContext) {
	for i := range l.links {
		switch {
		case l.links[i].LinkValid && l.links[i].Changed:

			ref := l.links[i].ExtLink
			c.DB.ExtLinkSave(&ref)
			l.links[i].Changed = false
		case !l.links[i].LinkValid && !l.links[i].Deleted && l.links[i].ID != 0:
			ref := l.links[i].ExtLink
			c.DB.ExtLinkDelete(&ref)
			l.links[i].Deleted = true
		default:
			//Nothing done
		}
	}
}

func (c *ZContext) MkExtLinkCtrl() (*ExtLinkCtrl, gtk.IWidget) {

	var l ExtLinkCtrl
	l.clip, _ = gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)

	//	c.extLinkCtrl = &l
	l.links = make([]*ExtLinkT, 0)
	l.tuneID = -1
	l.menuButton, _ = gtk.MenuButtonNew()
	l.menuButton.SetLabel("[]XLinks...")
	l.popover, _ = gtk.PopoverNew(l.menuButton)
	l.menuButton.SetPopover(l.popover)
	l.popover.Connect("closed", func() {
		l.SaveTuneList(c)
	})
	l.grid, _ = gtk.GridNew()
	l.popover.Add(l.grid)
	l.reCheck = regexp.MustCompile("http(s)?://.*/")
	l.noTune, _ = gtk.LabelNew("No tune selected")
	l.grid.Attach(l.noTune, 0, 0, 5, 1)
	l.popover.ShowAll()
	l.popover.Popdown()

	return &l, l.menuButton
}

func TheSessionCtrlNew() gtk.IWidget {
	b, _ := gtk.LinkButtonNew("TheSession...")
	//	linkButton, _ := gtk.LinkButtonNew("dummy")
	b.Connect("activate-link", func() bool {
		if t := ActiveTune(); t != nil {
			s := ""
			for _, c := range t.Title {
				switch {
				case unicode.IsLetter(c):
					s += string(c)
				case unicode.IsSpace(c):
					s += "+"
				}
			}
			link := "https://thesession.org/tunes/search?q=" + s
			b.SetUri(link)
			return false

		}
		return true
	})
	return b
}
