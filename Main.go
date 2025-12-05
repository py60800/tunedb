package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/py60800/tunedb/internal/imgprint"
	"github.com/py60800/tunedb/internal/player"
	"github.com/py60800/tunedb/internal/svgtab"
	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zdb"
	"github.com/py60800/tunedb/internal/zique"

	"debug/buildinfo"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type ZContext struct {
	// ********
	win   *gtk.Window
	menu  *gtk.MenuBar
	Image *TImage
	clip  *gtk.Clipboard

	DB                 *zdb.TuneDB
	mp3Collection      *zdb.MP3Collection
	sourceRepositories []zdb.SourceRepository

	tuneCtx    *TuneCtx
	ActiveTune *zdb.DTune

	tuneSelector *TuneSelector
	metronome    *WMetronome
	cursor       *gtk.DrawingArea

	playCtrl     *gtk.Box
	extLinkCtrl  *ExtLinkCtrl
	listMgr      *ListMgr
	midiPlayCtrl *MidiPlayCtrl

	setPlayCtrl   *SetPlayCtrl
	abcImport     *AbcImporter
	mp3PlayButton *gtk.Button

	mp3Player      *player.Mp3Player
	mp3SetPlayer   *Mp3SetConfigurator
	printer        *imgprint.Printer
	svgt           *svgtab.SvgTab
	concertinaCtrl *ConcertinaCtrl

	xchgCtrl *XchgCtrl
}

var ziqueContext ZContext

func Context() *ZContext {
	return &ziqueContext
}
func DB() *zdb.TuneDB {
	return ziqueContext.DB
}

func ActiveTune() *zdb.DTune {
	if ziqueContext.ActiveTune == nil || ziqueContext.ActiveTune.ID == 0 {
		return nil
	}
	return ziqueContext.ActiveTune
}

func (c *ZContext) ScanMp3() {
	msgC := make(chan string)
	go func() {
		msgC <- "Mp3 scan"
		zdb.MP3DBUpdate(c.DB)
		c.mp3Collection = zdb.MP3CollectionNew(c.DB)
		msgC <- ""
	}()
	OpWait("Please Wait", msgC)

}

var msgRelocate = `
Message relocation aims to move files to
their proper location according to their kind.
It can break the database.
Backup files before processing.
Confirm relocation ?
`

func (c *ZContext) MassRelocate() {
	if MessageConfirm(msgRelocate) {
		msgC := make(chan string)
		go func() {
			msgC <- "Mass Relocate"
			c.DB.MassRelocate(msgC)
			msgC <- ""
		}()
		OpWait("Please Wait", msgC)
	}

}

func (c *ZContext) MkMenu() {
	c.menu, _ = gtk.MenuBarNew()
	appMenu, _ := gtk.MenuItemNewWithLabel("App")
	c.menu.Append(appMenu)
	configMenu, _ := gtk.MenuItemNewWithLabel("Config")
	c.menu.Append(configMenu)

	appMenuGroup, _ := gtk.MenuNew()
	appMenu.SetSubmenu(appMenuGroup)
	refreshEntry, _ := gtk.MenuItemNewWithLabel("Scan MuseScore Repositories")
	refreshEntry.Connect("activate", func() { c.TUpdate() })

	appMenuGroup.Append(refreshEntry)
	fmMp3, _ := gtk.MenuItemNewWithLabel("Scan MP3 Repositories")
	fmMp3.Connect("activate", func() {
		c.ScanMp3()

	})
	appMenuGroup.Append(fmMp3)
	relocate, _ := gtk.MenuItemNewWithLabel("Relocate Tunes")
	relocate.Connect("activate", func() {
		c.MassRelocate()

	})
	appMenuGroup.Append(relocate)

	/*	refreshAbc, _ := gtk.MenuItemNewWithLabel("Scan ABC Repositories")
		refreshAbc.Connect("activate", func() { zdb.AbcDBUpdate(c.DB) })
		appMenuGroup.Append(refreshAbc)*/

	statEntry, _ := gtk.MenuItemNewWithLabel("Stats")
	appMenuGroup.Append(statEntry)
	statEntry.Connect("activate", func() {
		c.DisplayStats()
	})
	cleanup, _ := gtk.MenuItemNewWithLabel("Clean Up Temporary Files")
	appMenuGroup.Append(cleanup)
	cleanup.Connect("activate", func() {
		zdb.Cleanup()
	})

	buildInfo, _ := gtk.MenuItemNewWithLabel("Build info")
	appMenuGroup.Append(buildInfo)
	buildInfo.Connect("activate", func() {
		exe, _ := os.Executable()
		data, err := buildinfo.ReadFile(exe)
		var msg string
		if err != nil {
			msg = "BuildInfo:" + err.Error()
		} else {
			msg = data.String()
		}
		log.Println(msg)

		Message(msg)
	})

	quitEntry, _ := gtk.MenuItemNewWithLabel("Quit")
	appMenuGroup.Append(quitEntry)
	quitEntry.Connect("activate", func() {
		gtk.MainQuit()
	})

	configMenuGroup, _ := gtk.MenuNew()
	configMenu.SetSubmenu(configMenuGroup)
	configSource, _ := gtk.MenuItemNewWithLabel("Configure Sources")
	configSource.Connect("activate", func() {
		c.SourceConfiguration()
	})
	configMenuGroup.Append(configSource)
	configTK, _ := gtk.MenuItemNewWithLabel("Configure Tune Kinds")
	configTK.Connect("activate", func() {
		c.TuneKindConfiguration()
	})
	configMenuGroup.Append(configTK)

}
func (c *ZContext) LoadTuneByID(id int, keepPlayContext bool, forceReload bool) {
	if c.GetCurrentTuneID() == id && !forceReload {
		return
	}
	tune := c.DB.TuneGetByID(id)
	if tune.ID != 0 {
		c.LoadTune(&tune, keepPlayContext)
	}
}

func (c *ZContext) LoadTune(tune *zdb.DTune, keepPlayContext bool) {
	c.ActiveTune = tune
	c.svgt = nil
	if c.concertinaCtrl != nil {
		c.concertinaCtrl.Hide()
	}
	c.tuneCtx.LoadTune(tune, keepPlayContext)

	if !keepPlayContext {
		c.midiPlayCtrl.SetTempo(tune.Tempo, tune.Kind)
		c.midiPlayCtrl.SetInstrument(tune.Instrument)
	}

	c.mp3SetPlayer.ShowCount(tune.ID)
	c.extLinkCtrl.UpdateTuneLinks(tune.ID)
	c.setPlayCtrl.SetCount(c.DB.TuneSetGetCount(tune.ID))

	c.Image.Refresh()
	c.MarkOK()
}

func Quit() {
	log.Println("TuneDb Quit")
	c := Context()
	c.tuneSelector.SaveContext()
	c.midiPlayCtrl.Stop()
	c.midiPlayCtrl.Zique().Kill()
	c.mp3Player.Stop()
	util.RemovePidFile()
	gtk.MainQuit()
}
func (c *ZContext) SetHeader(change bool) {
	header := wHeader
	if tune := c.ActiveTune; tune != nil && tune.ID != 0 {
		header += "\t\t" + fmt.Sprintf("%s [%d]", tune.Title, tune.ID)
	}
	if change {
		header += " *"
	}
	c.win.SetTitle(header)
}

func (c *ZContext) MarkChange() {
	c.SetHeader(true)
}
func (c *ZContext) MarkOK() {
	c.SetHeader(false)
}
func (c *ZContext) GetCurrentTuneID() int {
	if c.ActiveTune == nil {
		return 0
	}
	return c.ActiveTune.ID
}
func (c *ZContext) Refresh() {
	if c.ActiveTune != nil {
		c.tuneCtx.LoadTune(c.ActiveTune, false)
	}
}

// Tune Context

func (c *ZContext) TUpdate() {
	msgC := make(chan string)
	go func() {
		msgC <- "Mscz update"
		c.DB.MsczContentUpdate()
		msgC <- ""
	}()
	OpWait("Please Wait", msgC)
	c.Refresh()
	c.tuneSelector.Refresh(false)
	c.DB.PurgeMscz()
}
func (c *ZContext) Stop() {
	c.midiPlayCtrl.Stop()
	c.mp3Player.Stop()
	c.metronome.MetronomeHide()
}
func (c *ZContext) PlayMidi() {
	c.Stop()
	if tune := c.ActiveTune; tune != nil && tune.ID != 0 && len(tune.Xml) > 0 {
		c.metronome.MetronomeShow()
		c.midiPlayCtrl.Zique().Play(tune.Xml)
	}

}
func (c *ZContext) MkPlayCtrl() {
	c.playCtrl, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1)
	midiPlay := MkButton("Play:Midi", func() {
		c.PlayMidi()
	})
	c.playCtrl.Add(midiPlay)
	c.mp3PlayButton = MkButton("Play:MP3", func() {
		c.Stop()
		c.mp3SetPlayer.PlayDefault()
	})
	c.playCtrl.Add(c.mp3PlayButton)

	bstop := MkButton("Stop", func() {
		c.Stop()
	})
	c.playCtrl.Add(bstop)

}
func (c *ZContext) CursorDraw() {
	c.cursor.QueueDraw()
}

var previousAlloc *gtk.Allocation

func (c *ZContext) RefreshTune() {
	if tune := c.ActiveTune; tune != nil && tune.ID != 0 {
		if tune.FileType == zdb.FileTypeMscz {
			log.Println("Do RefreshTune:", tune.Title)
			date, _ := util.GetModificationDate(tune.File)
			c.DB.MsczTuneSave(tune.File, "", date)
			c.Image.ForceUpdate()
			c.LoadTuneByID(tune.ID, true, true)
		}
	}
}

var StartupMessages []string

func OnceCheck() {
	Context().tuneSelector.Refresh(true)

	for _, m := range StartupMessages {
		log.Println("Startup message:", m)
		Message(m)
	}
	StartupMessages = []string{}
}
func WarnOnStart(msg string) {
	StartupMessages = append(StartupMessages, msg)
}

func (c *ZContext) StartUserInterface() {
	log.Println("Gtk Init")
	gtk.Init(nil)
	var w gtk.IWidget

	// Main window ---------------------------
	c.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	c.win.SetSizeRequest(800, 600)

	grid, _ := gtk.GridNew()
	firstLine, _ := gtk.GridNew()
	c.win.Add(grid)
	c.win.Connect("destroy", Quit)

	c.MkMenu()
	grid.Attach(c.menu, 0, 0, 6, 1)
	grid.Attach(firstLine, 0, 1, 12, 1)

	body, _ := gtk.GridNew()
	firstCol, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)

	var wTuneCtx gtk.IWidget
	c.tuneCtx, wTuneCtx = c.MkTuneCtx()
	firstCol.Add(wTuneCtx)
	var wm gtk.IWidget
	c.metronome, wm = c.MkMetro()
	firstCol.Add(wm)
	xc := 0
	body.Attach(firstCol, 0, 0, 3, 1)
	xc += 3
	firstCol.SetHAlign(gtk.ALIGN_START)

	// Image
	var wImg *gtk.Widget
	c.Image, wImg = c.MkImage()
	body.Attach(wImg, xc, 0, 8, 1)
	wImg.SetHAlign(gtk.ALIGN_CENTER)
	xc += 8

	//-------------------------- Tune Filter
	c.tuneSelector, w = c.MkTuneSelector()
	firstLine.Attach(w, 0, 1, 3, 1)

	c.cursor, _ = gtk.DrawingAreaNew()
	c.cursor.SetMarginStart(20)
	c.cursor.SetMarginEnd(20)
	c.cursor.SetSizeRequest(400, 30)

	c.MkPlayCtrl()

	firstLine.AttachNextTo(c.cursor, w, gtk.POS_RIGHT, 3, 1)
	firstLine.AttachNextTo(c.playCtrl, c.cursor, gtk.POS_RIGHT, 8, 1)

	cpy := MkButton("^C", func() {
		if tune := c.ActiveTune; tune != nil {
			c.clip.SetText(tune.Title)
		}
	})
	firstLine.AttachNextTo(cpy, c.playCtrl, gtk.POS_RIGHT, 1, 1)

	rightColumn, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)

	edit := MkButton("Muse Edit", func() {
		if tune := c.ActiveTune; tune != nil {
			zdb.MuseEdit(tune.File)
		}
	})
	rightColumn.Add(edit)
	refresh := MkButton("Refresh Tune", func() {
		c.RefreshTune()
	})
	rightColumn.Add(refresh)

	c.abcImport, w = c.MkAbcImport()
	rightColumn.Add(w)

	c.midiPlayCtrl, w = c.MkMidiPlayCtrl()
	rightColumn.Add(w)

	c.mp3SetPlayer, w = MkMp3SetConfigurator(c.cursor)
	rightColumn.Add(w)

	if WithConcertina {
		c.concertinaCtrl, w = ConcertinaCtrlNew()
		rightColumn.Add(w)
	}

	c.setPlayCtrl, w = MkSetPlayCtrl()
	rightColumn.Add(w)

	c.listMgr, w = MkListMgr()
	rightColumn.Add(w)

	c.extLinkCtrl, w = c.MkExtLinkCtrl()
	rightColumn.Add(w)

	w = TheSessionCtrlNew()
	rightColumn.Add(w)
	if WithXchg {
		c.xchgCtrl, w = c.MkXchgCtrl()
		rightColumn.Add(w)
	}
	print := MkButton("Print...", func() {
		if c.ActiveTune != nil && c.ActiveTune.ID != 0 {
			file := c.ActiveTune.Img
			if c.svgt != nil {
				file = c.svgt.TempFile
			}
			c.printer.Run([]string{file})
		}
	})
	rightColumn.Add(print)
	body.Attach(rightColumn, xc, 0, 2, 1)
	grid.Attach(body, 0, 2, 12, 1)
	c.win.Maximize()
	c.win.ShowAll()
	c.win.Connect("size-allocate", func(w *gtk.Window) {
		sc := w.GetScreen()
		di, _ := sc.GetDisplay()
		mo, _ := di.GetMonitor(0)
		geo := mo.GetGeometry()

		bodyAllocation := body.GetAllocation()
		rightA := firstCol.GetAllocation()
		leftA := rightColumn.GetAllocation()
		remainder := leftA.GetX() - (rightA.GetX() + rightA.GetWidth())

		if geo != nil && bodyAllocation.GetWidth() > geo.GetWidth() {
			w.Resize(bodyAllocation.GetWidth()-50, geo.GetHeight()-50)
			c.Image.SetSizeRequest(geo.GetWidth()/2, geo.GetHeight()/2)
		} else {

			c.Image.SetSizeRequest(remainder-50, bodyAllocation.GetHeight()-50)

		}
		winAllocation := w.GetAllocation()
		if previousAlloc == nil || *previousAlloc != *winAllocation {
			previousAlloc = winAllocation
		}

	})
	display, _ := gdk.DisplayGetDefault()
	c.clip, _ = gtk.ClipboardGetForDisplay(display, gdk.SELECTION_CLIPBOARD)
	c.printer = imgprint.PrinterNew()

	c.win.AddTickCallback(func(widget *gtk.Widget, frameClock *gdk.FrameClock) bool {
		OnceCheck()
		return false
	})
	log.Println("GTK Main")
	/*
		c.win.AddEvents(int(gdk.EVENT_KEY_PRESS))
		c.win.Connect("key-press-event", func(w *gtk.Window, evt *gdk.Event) {
			fmt.Printf("Evt: %T %v\n", evt, *evt)
		})*/

	gtk.Main()
}

// ****************************************************************************
var WithConcertina bool
var WithXchg bool

func main() {
	var workingDir string
	flag.BoolVar(&WithConcertina, "concertina", false, "Add concertina tab generator")
	flag.BoolVar(&WithXchg, "x", false, "Tune Xchg")
	flag.StringVar(&workingDir, "d", "", "Working directory")
	flag.Parse()

	MakeHomeContext(workingDir)

	c := Context()
	log.Println("Open database")
	c.DB = zdb.TuneDBNew()
	log.Println("Helper init")
	util.HelperInit(ConfigBase)
	log.Println("Param init")
	zdb.ParamInit(c.DB)
	if msg, err := zdb.CheckHelpers(); err != nil {
		WarnOnStart(msg)
	}
	zique.InitPattern("context")

	c.sourceRepositories = c.DB.SourceRepositoryGetAll()
	log.Println("Mscz Content Update")
	c.DB.MsczContentUpdate()

	log.Println("Retrieve MP3 from DB")
	c.mp3Collection = zdb.MP3CollectionNew(c.DB)
	c.mp3Player = player.Mp3PlayerNew()
	log.Println("Create user interface")
	c.StartUserInterface()
}
