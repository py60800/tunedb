// ConfigCtrl
package main

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
	"github.com/py60800/tunedb/internal/zdb"
)

var sourceType = []string{
	"Mscz",
	"Mp3",
	// "Abc",
}

type SourceRepositoryConfigurator struct {
	win        *gtk.Window
	wListStore *WListStore
}

func fillConfig(wLs *WListStore, c []zdb.SourceRepository) {
	wLs.Clear()
	for _, el := range c {
		r := "-"
		if el.Recurse {
			r = "Recurse"
		}
		wLs.AppendM(map[string]any{
			"Directory":    el.Location,
			"Resources":    el.Type,
			"Default Kind": el.DefaultKind,
			"Recurse":      r,
			"_ID":          el.ID,
		})

	}
}
func mkConfigurationWindow(c *ZContext) *SourceRepositoryConfigurator {
	cfg := &SourceRepositoryConfigurator{}
	cfg.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	cfg.win.SetTitle("Sources configuration")
	hConfigBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)

	cfg.win.Add(hConfigBox)
	cfg.win.SetModal(true)
	cfg.win.Connect("destroy", func() {
		cfg.win.Hide()
	})

	recurse := []string{"-", "Recurse"}
	tuneKinds := []string{"*"}
	tuneKinds = append(tuneKinds, c.DB.TuneKindGetAllNames()...)
	columns := []IListStoreColumn{
		ListStoreColumnTextNew("Directory", 50),
		ListStoreColumnComboNew("Resources", 5, sourceType),
		ListStoreColumnComboNew("Recurse", 6, recurse),
		ListStoreColumnComboNew("Default Kind", 10, tuneKinds),
	}

	configurator, w := WListStoreNew(nil, columns, true)
	cfg.wListStore = configurator
	hConfigBox.Add(w)

	lastLineBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1)
	lastLineBox.SetSpacing(50)

	save, _ := gtk.ButtonNewWithLabel("Save & Scan")
	save.Connect("clicked", func() {
		data := cfg.wListStore.GetValues()
		fmt.Println(data)
		sr := make([]zdb.SourceRepository, len(data))
		for i, d := range data {
			sr[i] = zdb.SourceRepository{
				Location:    d["Directory"].(string),
				Type:        d["Resources"].(string),
				DefaultKind: d["Default Kind"].(string),
				Recurse:     d["Recurse"].(string) == "Recurse",
				ID:          d["_ID"].(int),
			}
		}
		c.DB.SourceRepositoryUpdateAll(sr)
		c.sourceRepositories = c.DB.SourceRepositoryGetAll()
		cfg.win.Hide()

	})
	lastLineBox.Add(save)

	refresh, _ := gtk.ButtonNewWithLabel("Refresh From DB")
	refresh.Connect("clicked", func() {
		fillConfig(sourceRepositoryConfigurator.wListStore, c.DB.SourceRepositoryGetAll())
	})
	lastLineBox.Add(refresh)
	closeButton, _ := gtk.ButtonNewWithLabel("Close")
	closeButton.Connect("clicked", func() {
		cfg.win.Hide()
	})
	lastLineBox.Add(closeButton)
	lastLineBox.ShowAll()
	hConfigBox.Add(lastLineBox)
	hConfigBox.ShowAll()

	return cfg
}

var sourceRepositoryConfigurator *SourceRepositoryConfigurator

func (c *ZContext) SourceConfiguration() {
	if sourceRepositoryConfigurator == nil {
		sourceRepositoryConfigurator = mkConfigurationWindow(c)

	}
	currConfig := c.DB.SourceRepositoryGetAll()
	fillConfig(sourceRepositoryConfigurator.wListStore, currConfig)
	sourceRepositoryConfigurator.win.SetModal(true)
	sourceRepositoryConfigurator.win.ShowAll()
}

// ------------------------------------------------------------------------------
type TuneKindConfigurator struct {
	win        *gtk.Window
	wListStore *WListStore
}

var tuneKindConfigurator *TuneKindConfigurator

func (c *ZContext) TuneKindConfiguration() {
	if tuneKindConfigurator == nil {
		tkc := &TuneKindConfigurator{}
		tuneKindConfigurator = tkc
		tkc.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
		tkc.win.SetTitle("Tune Kind configuration")
		hConfigBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
		columns := []IListStoreColumn{
			ListStoreColumnTextNew("Tune Kind", 30),
			ListStoreColumnIntNew("Tempo", 3, 40, 250, 5),
		}
		var w gtk.IWidget
		tkc.wListStore, w = WListStoreNew(nil, columns, true)
		hConfigBox.Add(w)

		bottomBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
		save, _ := gtk.ButtonNewWithLabel("Save & Close")
		save.Connect("clicked", func() {
			data := tkc.wListStore.GetValues()
			res := make([]zdb.TuneKind, 0, len(data))
			for _, d := range data {
				res = append(res, zdb.TuneKind{
					Kind:  d["Kind"].(string),
					Tempo: d["Tempo"].(int),
					ID:    d["_ID"].(int),
				})
			}
			c.DB.TuneKindUpdateAll(res)
			c.tuneSelector.UpdateTK()
			c.tuneCtx.UpdateTuneKind()
			/*
				fmt.Println(tkc)
				ls := tkc.wListStore.GetListStore()
				res := make([]TuneKind, 0)
				iter, ok := ls.GetIterFirst()
				if ok {
					for {
						res = append(res,
							TuneKind{
								Kind:  ListStoreGetString(ls, iter, 0),
								Tempo: ListStoreGetInt(ls, iter, 1),
								ID:    ListStoreGetInt(ls, iter, tkc.wListStore.GetIdIdx()),
							})
						if !ls.IterNext(iter) {
							break
						}

					}
					c.DB.TuneKindUpdateAll(res)
					c.tuneSelector.UpdateTK()
					c.tuneCtx.UpdateTuneKind()
				}
			*/
			tkc.win.Hide()
		})
		cancel, _ := gtk.ButtonNewWithLabel("Cancel")
		cancel.Connect("clicked", func() {
			tkc.win.Hide()
		})

		bottomBox.Add(save)
		bottomBox.Add(cancel)
		hConfigBox.Add(bottomBox)
		tkc.win.Add(hConfigBox)
		tkc.win.SetModal(true)
		tkc.win.Connect("destroy", func() {
			tkc.win.Hide()
		})

	}
	currConfig := c.DB.TuneKindGetAll()
	ls := tuneKindConfigurator.wListStore
	ls.Clear()
	for _, tk := range currConfig {
		ls.AppendM(map[string]any{
			"_ID":       tk.ID,
			"_changed":  false,
			"Tune Kind": tk.Kind,
			"Tempo":     tk.Tempo,
		})
	}

	tuneKindConfigurator.win.SetModal(true)
	tuneKindConfigurator.win.ShowAll()
}

// =============================================================================

// Tune tags
/*
type TuneTagConfigurator struct {
	win        *gtk.Window
	wListStore *WListStore
}

var tuneTagConfigurator *TuneTagConfigurator

func (c *ZContext) TuneTagConfiguration() {
	ttc := tuneTagConfigurator
	if ttc == nil {
		ttc = &TuneTagConfigurator{}
		tuneTagConfigurator = ttc
		ttc.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
		ttc.win.SetTitle("Tune Tag configuration")
		hConfigBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
		columns := []IListStoreColumn{
			ListStoreColumnTextNew("Tag", 3),
			ListStoreColumnTextNew("Long Name", 30),
		}
		var w gtk.IWidget
		ttc.wListStore, w = WListStoreNew(nil, columns, true)
		hConfigBox.Add(w)

		bottomBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
		save, _ := gtk.ButtonNewWithLabel("Save & Close")
		save.Connect("clicked", func() {
			data := ttc.wListStore.GetValues()
			if len(data) > 12 {
				Message("Can't create more than 12 tags")
				return
			}
			newTags := make([]TuneTag, len(data))
			for i := range data {
				sn := data[i]["Tag"].(string)
				ln := data[i]["Long Name"].(string)
				switch len(sn) {
				case 0:
					sn = fmt.Sprint(i + 1)
				case 1, 2, 3:
					//OK
				default: //  3
					sn = sn[:3]
				}
				newTags[i] = TuneTag{i, sn, ln}
			}
			GetContext().DB.TuneTagSaveAll(newTags)
			//tuneTagInit(GetContext().DB)
			Message("Please Relaunch Application")
			ttc.win.Hide()
		})
		cancel, _ := gtk.ButtonNewWithLabel("Cancel")
		cancel.Connect("clicked", func() {
			ttc.win.Hide()
		})

		bottomBox.Add(save)
		bottomBox.Add(cancel)
		hConfigBox.Add(bottomBox)
		ttc.win.Add(hConfigBox)
		ttc.win.SetModal(true)
		ttc.win.Connect("destroy", func() {
			ttc.win.Hide()
		})

	}
	currTags := c.DB.TuneTagGetAll()
	ttc.wListStore.Clear()

	for _, tk := range currTags {
		ttc.wListStore.Append(tk.ShortName, tk.LongName)
	}
	ttc.win.SetModal(true)
	ttc.win.ShowAll()
}
*/
