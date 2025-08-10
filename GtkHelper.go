// GtkHelper
package main

import (
	"fmt"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func DelayedAction(w gtk.IWidget, d time.Duration, action func()) {
	widget := w.ToWidget()
	timeOut := time.Now().Add(d)
	widget.AddTickCallback(func(widget *gtk.Widget, frameClock *gdk.FrameClock) bool {
		if time.Now().After(timeOut) {
			action()
			return false
		}
		return true
	})
}

func MkDeferedSearchEntry(pSuspend *bool, action func(searchText string)) *gtk.SearchEntry {
	filterText, _ := gtk.SearchEntryNew()
	var timeFilterTo time.Time
	ticking := false
	suspend := pSuspend
	filterText.Connect("changed", func() {
		if suspend != nil && *suspend {
			return
		}
		timeFilterTo = time.Now().Add((400 * time.Millisecond))
		if !ticking {
			ticking = true
			filterText.AddTickCallback(func(widget *gtk.Widget, frameClock *gdk.FrameClock) bool {
				if time.Now().After(timeFilterTo) {
					searchText, _ := filterText.GetText()
					action(searchText)
					ticking = false
					return false
				}
				return true
			})
		}
	})
	return filterText
}

func SetMargins(widget gtk.IWidget, wm, hm int) {
	w := widget.ToWidget()
	w.SetMarginBottom(hm)
	w.SetMarginTop(hm)
	w.SetMarginStart(wm)
	w.SetMarginEnd(wm)

}
func MkButton(label string, ft func()) *gtk.Button {
	b, _ := gtk.ButtonNewWithLabel(label)
	b.Connect("clicked", ft)
	return b
}
func MkButtonIcon(label string, ft func()) *gtk.Button {
	fmt.Println("Icon:", label)
	b, _ := gtk.ButtonNewFromIconName(label, gtk.ICON_SIZE_BUTTON)
	b.Connect("clicked", ft)
	return b
}

func MkLabel(l string) *gtk.Label {
	label, _ := gtk.LabelNewWithMnemonic(l)
	return label
}
func Message(msg string) {
	f := gtk.DialogFlags(gtk.DIALOG_DESTROY_WITH_PARENT)
	d := gtk.MessageDialogNew(GetContext().win, f, gtk.MESSAGE_INFO, gtk.BUTTONS_CLOSE, msg)
	d.Run()
	d.Destroy()
}
func MessageConfirm(msg string) bool {
	f := gtk.DialogFlags(gtk.DIALOG_DESTROY_WITH_PARENT)
	d := gtk.MessageDialogNew(GetContext().win, f, gtk.MESSAGE_INFO, gtk.BUTTONS_OK_CANCEL, msg)
	r := d.Run()
	d.Destroy()
	return r == gtk.RESPONSE_OK
}

func OpWait(msg string, msgChan chan string) {
	f := gtk.DialogFlags(gtk.DIALOG_DESTROY_WITH_PARENT)
	d := gtk.MessageDialogNew(GetContext().win, f, gtk.MESSAGE_INFO, 0, msg)
	d.ToWidget().AddTickCallback(func(widget *gtk.Widget, frameClock *gdk.FrameClock) bool {
		select {
		case msg := <-msgChan:
			if msg == "" {
				d.Destroy()
			} else {
				d.FormatSecondaryText("Hello:%v", msg)
			}
		default:
		}
		return true
	})

	d.Run()
	d.Destroy()

}
