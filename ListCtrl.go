// SetPlayer
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zdb"

	_ "embed"
)

type ListMgr struct {
	tuneList        []zdb.TuneListBase
	currentTuneList *zdb.TuneList

	menuButton *gtk.MenuButton

	lName        *gtk.Label
	listSelector *gtk.ComboBoxText

	tuneSelector  *STuneSelector
	popo          *gtk.Popover
	searchEntry   *gtk.SearchEntry
	setCombo      *gtk.ComboBoxText
	suspendChange bool

	tsSelectorIdx []int
	listStore     *WTreeStore
	name          *gtk.Entry
	tag           *gtk.Entry
	comment       *gtk.TextView
}

func TitleCombine(lst []string) string {
	title := ""
	for i, t := range lst {
		p := ""
		if i > 0 {
			p = "/"
		}
		title += p + util.STruncate(t, 10)
	}
	return util.STruncate(title, 40)
}

func doGroup(tv *WTreeStore, val [][]any) []any {
	res := make([]any, len(val[0]))
	res[tv.GetColIdx("ID")] = len(val)
	res[tv.GetColIdx("Kind")] = "Set"
	titles := make([]string, 0)
	for i := range val {
		t := val[i][tv.GetColIdx("Title")]
		if s, ok := t.(string); ok {
			titles = append(titles, s)
		}
	}
	res[tv.GetColIdx("Title")] = TitleCombine(titles)
	ic := tv.GetColIdx("Tie")
	ig := tv.GetColIdx("_group")
	for i := range val {
		switch i {
		case 0:
			val[i][ic] = icons[zdb.TL_GroupStart]
			val[i][ig] = zdb.TL_GroupStart
		case len(val) - 1:
			val[i][ic] = icons[zdb.TL_GroupEnd]
			val[i][ig] = zdb.TL_GroupEnd
		default:
			val[i][ic] = icons[zdb.TL_GroupMid]
			val[i][ig] = zdb.TL_GroupMid
		}
	}
	return res
}
func unGroup(tv *WTreeStore, val [][]any) {
	for i := range val {
		val[i][tv.GetColIdx("Tie")] = icons[zdb.TL_GroupNone]
		val[i][tv.GetColIdx("_group")] = zdb.TL_GroupNone
	}
}

func MkListMgr() (*ListMgr, gtk.IWidget) {
	sp := &ListMgr{}
	sp.menuButton, _ = gtk.MenuButtonNew()
	sp.menuButton.SetLabel("Lists...")
	sp.menuButton.Connect("clicked", func() {
		sp.GetAllTunelist()
		sp.fillSelector()
		sp.popo.ShowAll()
	})
	is := 0
	gw := 6
	sp.popo, _ = gtk.PopoverNew(sp.menuButton)
	sp.menuButton.SetPopover(sp.popo)

	mainGrid, _ := gtk.GridNew()

	sp.popo.Add(mainGrid)
	sp.lName, _ = gtk.LabelNew("...")
	mainGrid.Attach(sp.lName, 0, is, gw, 1)
	is++

	sp.listSelector, _ = gtk.ComboBoxTextNew()
	selB := MkButton("Select", func() { // Select a tune set
		idx := sp.listSelector.GetActive()
		if idx >= 0 && idx < len(sp.tuneList) {
			sp.SelectTuneList(&sp.tuneList[idx])
		}
	})
	mainGrid.Attach(sp.listSelector, 0, is, gw-2, 1)
	mainGrid.AttachNextTo(selB, sp.listSelector, gtk.POS_RIGHT, 2, 1)
	is++
	//
	lName, _ := gtk.LabelNew("Name")
	sp.name, _ = gtk.EntryNew()
	x := 0
	mainGrid.Attach(lName, 0, is, 1, 1)
	x++
	mainGrid.Attach(sp.name, x, is, gw-4, 1)
	x += gw - 4
	lTag, _ := gtk.LabelNew("Tag")
	sp.tag, _ = gtk.EntryNew()
	mainGrid.Attach(lTag, x, is, 1, 1)
	x++
	mainGrid.Attach(sp.tag, x, is, 1, 1)
	is++
	lcomment, _ := gtk.FrameNew("Comment")
	sp.comment, _ = gtk.TextViewNew()
	lcomment.Add(sp.comment)
	mainGrid.Attach(lcomment, 0, is, gw, 2)
	is += 2

	var w gtk.IWidget
	sp.tuneSelector, w = STuneSelectorNew(func(ref *zdb.DTuneReference) {
		tune := DB().TuneGetByID(ref.ID)
		sp.listStore.AppendM(nil, map[string]any{
			"ID":     tune.ID,
			"_ID":    tune.ID,
			"Title":  tune.Title,
			"Kind":   tune.Kind,
			"PlayL":  tune.Play.String(),
			"Tie":    icons[0],
			"_group": zdb.TL_GroupNone,
		})
	})
	mainGrid.Attach(w, 0, is, gw, 1)
	is++
	columns := []ITreeStoreColumn{
		TreeStoreColumnIntNew("ID", 2, 0, 1000000, 1),
		TreeStoreColumnTextNew("Kind", 5),
		TreeStoreColumnTextNew("Title", 30),
		TreeStoreColumnTextNew("PlayL", 5),
		TreeStoreColumnIconNew("Tie", 2),
		TreeStoreColumnIntNew("_group", 0, 0, 4, 1),
	}
	vadj, _ := gtk.AdjustmentNew(0, 0, 100, 1, 0, 0)
	scw, _ := gtk.ScrolledWindowNew(nil, vadj)
	sp.listStore, w = WTreeStoreNew(scw, columns, false)
	scw.SetSizeRequest(450, 400)

	mainGrid.Attach(w, 0, is, gw, 10)
	sp.listStore.SetActivate(func(data map[string]interface{}) {
		if id, ok := data["ID"].(int); ok {
			Context().LoadTuneByID(id, false, true)
		}
	})
	sp.listStore.SetOnGroup(doGroup, unGroup)

	is += 10

	save := MkButton("Save", func() {
		sp.SaveTuneList()
		sp.TuneCtxRefresh()
	})
	apply := MkButton("Save&Apply", func() {
		if sp.SaveTuneList() {
			sp.apply()
			sp.popo.Hide()
		}
		sp.TuneCtxRefresh()
	})
	deDup := MkButton("DeDup", func() {
		log.Println("Deduplicate TuneList")
		sp.DeDup()
	})

	clear := MkButton("Clear", func() {
		sp.clear()
	})

	export := MkButton("Export", func() {
		sp.Export()
	})
	del := MkButton("Del", func() {
		if sp.currentTuneList != nil && sp.currentTuneList.ID != 0 {
			if len(sp.currentTuneList.Tunes) < 5 || MessageConfirm(fmt.Sprintf("Delete tune list ?")) {
				log.Println("Delete tune list:", sp.currentTuneList.ID)
				DB().TuneListRemove(sp.currentTuneList)
				zdb.TuneTagUpdate(DB())
				sp.TuneCtxRefresh()
			}
		}
		sp.clear()
		sp.UpdateCombo()
	})
	info := MkListInfo(sp)
	mainGrid.Attach(save, 0, is, 1, 1)
	mainGrid.Attach(apply, 1, is, 1, 1)
	mainGrid.Attach(clear, 2, is, 1, 1)
	mainGrid.Attach(deDup, 3, is, 1, 1)
	mainGrid.Attach(del, 4, is, 1, 1)
	mainGrid.Attach(info, 5, is, 1, 1)
	mainGrid.Attach(export, 6, is, 1, 1)
	return sp, sp.menuButton

}
func (sp *ListMgr) doExport(file string, name string) {
	tunes := sp.listStore.GetValues()
	f, _ := os.Create(file)
	defer f.Close()
	csf := csv.NewWriter(f)
	defer csf.Flush()
	csf.Write([]string{"Title", "Group", "Kind", "Mscz", "Img", "Xml"})

	for _, t := range tunes {
		tune := DB().TuneGetByID(t["ID"].(int))
		csf.Write([]string{
			tune.Title,
			fmt.Sprint(t["_group"].(int)),
			tune.Kind,
			tune.File,
			tune.Img,
			tune.Xml,
		})
	}

}
func (sp *ListMgr) Export() {
	name, _ := sp.name.GetText()
	fc, _ := gtk.FileChooserNativeDialogNew("Select export",
		Context().win, gtk.FILE_CHOOSER_ACTION_SAVE, "OK", "Cancel")
	fc.SetCurrentName(name + ".csv")
	if fc.Run() == int(gtk.RESPONSE_ACCEPT) {

		fmt.Println("Accept", fc.GetFilename())
		sp.doExport(fc.GetFilename(), name)
	}
}
func (sp *ListMgr) TuneCtxRefresh() {
	TuneTagUpdated = true
	Context().tuneCtx.Refresh()
}
func (sp *ListMgr) fillSelector() {
	sp.suspendChange = true
	defer func() {
		sp.suspendChange = false
	}()
	sp.listSelector.RemoveAll()
	for _, ts := range sp.tuneList {
		txt := fmt.Sprintf("[%d|%s]%s", ts.ID, ts.Tag, ts.Name)
		sp.listSelector.AppendText(txt)
	}
	sp.listSelector.SetActive(0)
}
func (sp *ListMgr) apply() {
	tunes := sp.listStore.GetValues()
	ids := make([]int, len(tunes))
	for i, t := range tunes {
		ids[i] = t["ID"].(int)
	}
	Context().tuneSelector.RefreshFromList(ids)
}
func (sp *ListMgr) clear() {
	sp.currentTuneList = &zdb.TuneList{}
	sp.lName.SetLabel("...")
	sp.name.SetText("")
	sp.listStore.Clear()
}
func (sp *ListMgr) SelectTuneList(ts *zdb.TuneListBase) {
	sp.currentTuneList = DB().GetTuneListByID(ts.ID)
	if sp.currentTuneList == nil {
		msg := fmt.Sprintf("[%v] not found", ts.Name)
		log.Println("Select Tune List", msg)
		Message(msg)
	}
	sp.lName.SetLabel(ts.Name)
	sp.name.SetText(ts.Name)
	sp.tag.SetText(ts.Tag)
	b, _ := sp.comment.GetBuffer()
	b.SetText(ts.Comment)

	ls := sp.listStore
	ls.Clear()
	var parent *gtk.TreeIter
	tl := sp.currentTuneList.Tunes
	tunes := make([]zdb.DTune, len(tl))
	for i := range tl {
		tunes[i] = DB().TuneGetByID(tl[i].DTuneID)
	}
	for i, tis := range tl {
		// tune := GetDB().TuneGetByID(tis.DTuneID)
		tune := &tunes[i]
		if tune.ID != 0 {
			if tis.Group == zdb.TL_GroupStart {
				titles := make([]string, 0)
				for j := i; j < len(tl); j++ {
					titles = append(titles, tunes[j].Title)
					if tl[j].Group == zdb.TL_GroupEnd {
						break
					}
				}
				parent = sp.listStore.AppendM(nil, map[string]any{
					"_ID":      0,
					"_changed": false,
					"ID":       0,
					"Title":    TitleCombine(titles),
					"Kind":     "Set",
					"PlayL":    "",
					"Tie":      icons[0],
					"_group":   0,
				})
			}
			sp.listStore.AppendM(parent, map[string]any{
				"_ID":      tis.DTuneID,
				"_changed": false,
				"ID":       tune.ID,
				"Title":    tune.Title,
				"Kind":     tune.Kind,
				"PlayL":    tune.Play.String(),
				"Tie":      icons[tis.Group],
				"_group":   tis.Group,
			})
			if tis.Group == zdb.TL_GroupEnd {
				parent = nil
			}
		}

	}
	sp.listStore.Expand()
}
func (sp *ListMgr) GetAllTunelist() {
	sp.tuneList = DB().TuneListGetAll()
}
func (sp *ListMgr) UpdateCombo() {
	sp.GetAllTunelist()
	sp.fillSelector()
}
func (sp *ListMgr) DeDup() {
	tunes := sp.listStore.GetValues()
	found := make(map[int]int)
	for i := 0; i < len(tunes); {
		if _, ok := found[tunes[i]["ID"].(int)]; ok {
			// duplicate
			if i == len(tunes)-1 {
				tunes = tunes[:i]
				break
			} else {
				tunes = append(tunes[:i], tunes[i+1:]...)
			}
		} else {
			found[tunes[i]["ID"].(int)] = 0
			i++
		}
	}
	sp.listStore.Clear()
	for _, t := range tunes {
		sp.listStore.AppendM(nil, t)
	}
}
func (sp *ListMgr) SaveTuneList() bool {
	name, _ := sp.name.GetText()
	if name == "" {
		Message("Name Required")
		return false
	}
	if sp.currentTuneList == nil {
		sp.currentTuneList = &zdb.TuneList{}
	}
	tag, _ := sp.tag.GetText()
	if len(tag) >= 3 {
		tag = tag[0:3]
	}

	b, _ := sp.comment.GetBuffer()
	start, end := b.GetBounds()
	comment, _ := b.GetText(start, end, true)

	if tag != "" && tag != sp.currentTuneList.Tag {
		tags := DB().TuneListTags()
		for _, t := range tags {
			if t.Tag == tag {
				Message("Duplicate tag:" + tag)
				tag = ""
				break
			}
		}
		if len(tags) >= 12 {
			Message("Too many tags")
			tag = ""
		}
	}

	sp.currentTuneList.Tag = tag
	sp.currentTuneList.Comment = comment
	sp.currentTuneList.Name = name

	tunes := sp.listStore.GetValues()
	if sp.currentTuneList == nil || sp.currentTuneList.ID == 0 {
		sp.currentTuneList = &zdb.TuneList{}
	}
	sp.currentTuneList.Tunes = make([]zdb.TuneListItem, len(tunes))
	sp.currentTuneList.Name, _ = sp.name.GetText()

	for i := range tunes {
		t := tunes[i]
		sp.currentTuneList.Tunes[i] = zdb.TuneListItem{
			Rank:    i,
			DTuneID: t["ID"].(int),
			Group:   t["_group"].(int),
		}
	}
	log.Println("Save Tune List:", name)
	DB().TuneListSave(sp.currentTuneList)
	zdb.TuneTagUpdate(DB())

	sp.UpdateCombo()
	return true
	//	sp.selectTuneSet(sp.currentTuneSet.Name)
}

var icons []*gdk.Pixbuf

//go:embed icons/blank.svg
var blankIcon []byte

//go:embed icons/start.svg
var startIcon []byte

//go:embed icons/middle.svg
var middleIcon []byte

//go:embed icons/end.svg
var endIcon []byte

func init() {
	var icn = [][]byte{
		blankIcon,
		startIcon,
		middleIcon,
		endIcon,
	}
	icons = make([]*gdk.Pixbuf, 0)
	for i := range icn {
		//		o := glib.Object{}
		ic, err := gdk.PixbufNewFromBytesOnly(icn[i])
		if err != nil {
			panic(err)
		}
		//o.SetProperty("pixbuf", ic)
		icons = append(icons, ic)
	}
}

func MkListInfo(sp *ListMgr) gtk.IWidget {
	b, _ := gtk.MenuButtonNew()
	b.SetLabel("Info...")
	popo, _ := gtk.PopoverNew(b)
	entry, _ := gtk.TextViewNew()
	popo.Add(entry)
	b.SetPopover(popo)
	b.Connect("clicked", func() {

		buffer, _ := entry.GetBuffer()
		tunes := sp.listStore.GetValues()
		stats := make(map[string]int)
		for _, t := range tunes {
			pl := t["PlayL"].(string)
			stats[pl]++
		}
		var s = &strings.Builder{}
		sk := make([]string, 0, len(stats))
		for k := range stats {
			sk = append(sk, k)
		}
		sort.Strings(sk)
		cpt := len(tunes)
		fmt.Fprintf(s, "%d Tunes\n", cpt)
		for _, k := range sk {
			n := stats[k]
			fmt.Fprintf(s, "%-50s : %3d | %3d %2d%%\n", k, n, cpt, (cpt*100)/len(tunes))
			cpt -= n
		}
		buffer.SetText(s.String())
		entry.Show()
	})
	return b
}
