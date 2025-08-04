// View_test
package imgprint

import (
	//	"fmt"
	"testing"

	"github.com/gotk3/gotk3/gtk"
)

var imgs = []string{
	"/home/yvon/Music/Mscz/Reel/img/Charlie_Mulvihill's-1.svg",
	"/home/yvon/Music/Mscz/Reel/img/Heather_Breeze,_The-1.svg",
}

func TestMain(m *testing.M) {

	gtk.Init(nil)
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	grid, _ := gtk.GridNew()
	win.Add(grid)
	button, _ := gtk.ButtonNewWithLabel("Run test")
	grid.Attach(button, 0, 0, 4, 1)
	quit, _ := gtk.ButtonNewWithLabel("Quit")
	grid.Attach(quit, 0, 1, 4, 1)
	quit.Connect("clicked", func() {
		gtk.MainQuit()
	})

	printer := PrinterNew()
	button.Connect("clicked", func() {
		printer.Run(imgs)
	})
	win.ShowAll()
	gtk.Main()

}
