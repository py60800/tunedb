// Stats
package main

import (
	"github.com/py60800/tunedb/internal/zdb"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func (c *ZContext) DisplayStats() {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	grid, _ := gtk.GridNew()
	win.Add(grid)
	win.Connect("destroy", func() {
		win.Destroy()
	})
	allStats := make(map[string][]int)
	hidden := 0
	c.DB.TuneIter(func(t *zdb.DTune) {
		if !t.Hide {
			stat, ok := allStats[t.Kind]
			if !ok {
				stat = make([]int, zdb.PlayLevelMax+1)
			}
			stat[zdb.PlayLevelMax] += 1
			stat[t.Play] += 1
			allStats[t.Kind] = stat
		} else {
			hidden++
		}
	})

	sum := make([]int, zdb.PlayLevelMax+1)
	sumC := make([]int, zdb.PlayLevelMax+1)
	for _, stat := range allStats {
		for i := range stat {
			sum[i] += stat[i]
		}
	}
	cumul := 0
	for i := zdb.PlayLevelMax - 1; i >= 0; i-- {
		cumul += sum[i]
		sumC[i] = cumul
	}

	tv, _ := gtk.TreeViewNew()
	lTypes := make([]glib.Type, zdb.PlayLevelMax+2)
	lTypes[0] = glib.TYPE_STRING
	for i := 1; i < len(lTypes); i++ {
		lTypes[i] = glib.TYPE_INT
	}
	ls, _ := gtk.ListStoreNew(lTypes...)
	tv.SetModel(ls)
	for i := range lTypes {
		var title string
		var cr gtk.ICellRenderer
		switch {
		case i == 0:
			cr, _ = gtk.CellRendererTextNew()
			title = "Kind"
		case i > len(zdb.PlayLevelStr):
			title = "All"
			cr, _ = gtk.CellRendererSpinNew()
		default:
			title = zdb.PlayLevelStr[i-1]
			cr, _ = gtk.CellRendererSpinNew()
		}
		ck, _ := gtk.TreeViewColumnNewWithAttribute(title, cr, "text", i)
		tv.AppendColumn(ck)
	}
	for k, v := range allStats {
		iter := ls.Append()
		ls.SetValue(iter, 0, k)
		for i := range v {
			ls.SetValue(iter, i+1, v[i])
		}
	}
	iter := ls.Append()
	ls.SetValue(iter, 0, "SUM")
	for i := range sum {
		ls.SetValue(iter, i+1, sum[i])
	}
	iter = ls.Append()
	ls.SetValue(iter, 0, "---")
	for i := range sumC {
		ls.SetValue(iter, i+1, sumC[i])
	}

	tv.ShowAll()
	grid.Attach(tv, 0, 1, 8, 10)
	cl := MkButton("Close", func() {
		win.Destroy()
	})
	grid.Attach(cl, 0, 0, 3, 1)
	grid.ShowAll()
	win.ShowAll()

}
