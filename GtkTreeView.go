// GtkHelper
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// TreeStore
type ITreeStoreColumn interface {
	Title() string
	Type() glib.Type
	Widget() gtk.IWidget
	Renderer() gtk.ICellRenderer
	SetValue(any)
	GetValue() any
	BlankValue() any
	Attribute() string
	Self() *TreeStoreColumn
	SetWidth(int)
}

type TreeStoreColumn struct {
	CTitle    string
	GType     glib.Type
	Attr      string
	CRenderer gtk.ICellRenderer
	Storage   *gtk.TreeStore
	CharWidth int
	Column    int
	widget    gtk.IWidget
}

func (lc *TreeStoreColumn) Self() *TreeStoreColumn {
	return lc
}
func (lc *TreeStoreColumn) Attribute() string {
	return lc.Attr
}
func (lc *TreeStoreColumn) Title() string {
	return lc.CTitle
}
func (lc *TreeStoreColumn) Widget() gtk.IWidget {
	return lc.widget
}

func (lc *TreeStoreColumn) Type() glib.Type {
	return lc.GType
}
func (lc *TreeStoreColumn) Renderer() gtk.ICellRenderer {
	return lc.CRenderer
}

// --------------------------------------------
type TreeStoreColumnText struct {
	TreeStoreColumn
	entry *gtk.Entry
}

func TreeStoreColumnTextNew(title string, w int) *TreeStoreColumnText {
	r, _ := gtk.CellRendererTextNew()
	entry, _ := gtk.EntryNew()
	entry.SetWidthChars(w)
	entry.SetEditable(true)

	return &TreeStoreColumnText{
		TreeStoreColumn: TreeStoreColumn{
			CTitle:    title,
			GType:     glib.TYPE_STRING,
			CharWidth: w,
			Attr:      "text",
			CRenderer: r,
			widget:    entry,
		},
		entry: entry,
	}
}
func (ls *TreeStoreColumnText) SetValue(val any) {
	if v, ok := val.(string); ok {
		ls.entry.SetText(v)
	}
}
func (ls *TreeStoreColumnText) GetValue() any {
	t, _ := ls.entry.GetText()
	return t
}

func (ls *TreeStoreColumnText) BlankValue() any {
	return ""
}
func (ls *TreeStoreColumnText) SetWidth(w int) {
	ls.widget.(*gtk.Entry).SetSizeRequest(w, -1)
}

// Int -------------------------------------------------------------------------
type TreeStoreColumnInt struct {
	TreeStoreColumn
	spin *gtk.SpinButton
	from int
}

func TreeStoreColumnIntNew(title string, digits int, from int, to int, step int) *TreeStoreColumnInt {
	r, _ := gtk.CellRendererSpinNew()
	r.SetProperty("digits", digits)

	spin, _ := gtk.SpinButtonNewWithRange(float64(from), float64(to), float64(step))

	return &TreeStoreColumnInt{
		TreeStoreColumn: TreeStoreColumn{
			CTitle:    title,
			GType:     glib.TYPE_INT,
			CharWidth: digits + 4,
			Attr:      "text",
			CRenderer: r,
			widget:    spin,
		},
		spin: spin,
		from: from,
	}
}
func (ls *TreeStoreColumnInt) SetValue(val any) {
	if v, ok := val.(int); ok {
		ls.spin.SetValue(float64(v))
	}
}
func (ls *TreeStoreColumnInt) GetValue() any {
	return ls.spin.GetValueAsInt()
}
func (ls *TreeStoreColumnInt) BlankValue() any {
	return ls.from
}
func (ls *TreeStoreColumnInt) SetWidth(w int) {
	ls.widget.(*gtk.SpinButton).SetSizeRequest(w, -1)
}

type TreeStoreColumnIcon struct {
	TreeStoreColumn
}

func (ls *TreeStoreColumnIcon) Widget() gtk.IWidget {
	return nil
}
func TreeStoreColumnIconNew(title string, digits int) *TreeStoreColumnIcon {
	cr, _ := gtk.CellRendererPixbufNew()
	lc := TreeStoreColumnIcon{
		TreeStoreColumn: TreeStoreColumn{
			CTitle: title,
			//GType:     glib.TYPE_STRING,
			GType:     glib.TYPE_OBJECT,
			CharWidth: digits + 4,
			//Attr:      "icon-name",
			Attr:      "pixbuf",
			CRenderer: cr,
		},
	}
	return &lc
}

func (ls *TreeStoreColumnIcon) SetValue(t any) {
	panic("Unimplemented")
}

func (ls *TreeStoreColumnIcon) GetValue() any {
	return nil
}

func (ls *TreeStoreColumnIcon) BlankValue() any {
	return 0
}
func (ls *TreeStoreColumnIcon) SetWidth(w int) {
	//ls.w.SetSizeRequest(w, -1)
}

// Combo
type TreeStoreColumnCombo struct {
	TreeStoreColumn
	combo  *gtk.ComboBoxText
	values []string
}

func (ls *TreeStoreColumnCombo) Widget() gtk.IWidget {
	return ls.combo
}
func TreeStoreColumnComboNew(title string, w int, values []string) *TreeStoreColumnCombo {
	lc := TreeStoreColumnCombo{
		TreeStoreColumn: TreeStoreColumn{CTitle: title,
			GType: glib.TYPE_STRING, CharWidth: w + 4, Attr: "text"},
	}
	lc.combo, _ = gtk.ComboBoxTextNew()
	lc.values = values
	for _, s := range values {
		lc.combo.AppendText(s)
	}
	lc.CRenderer, _ = gtk.CellRendererTextNew()
	return &lc
}
func (ls *TreeStoreColumnCombo) SetValue(ti any) {
	t := ti.(string)
	for i, v := range ls.values {
		if v == t {
			ls.combo.SetActive(i)
		}
	}
}
func (ls *TreeStoreColumnCombo) GetValue() any {
	t := ls.combo.GetActiveText()
	return t
}

func (ls *TreeStoreColumnCombo) BlankValue() any {
	return ls.values[0]
}
func (ls *TreeStoreColumnCombo) SetWidth(w int) {
	ls.combo.SetSizeRequest(w, -1)
}

type WTreeStore struct {
	container *gtk.ScrolledWindow
	treeView  *gtk.TreeView
	popover   *gtk.Popover
	columns   []ITreeStoreColumn

	TreeStore *gtk.TreeStore

	colNames   map[string]int
	ColCount   int
	sel        *gtk.TreeSelection
	onActivate func(map[string]any)
	onGroup    func(wt *WTreeStore, data [][]any) []any
	onUnGroup  func(wt *WTreeStore, data [][]any)
	expand     *gtk.ToggleButton
}

func (ls *WTreeStore) SetActivate(handler func(map[string]any)) {
	ls.onActivate = handler
}
func (ls *WTreeStore) SetOnGroup(handler func(wt *WTreeStore, data [][]any) []any,
	handlerU func(wt *WTreeStore, data [][]any)) {
	ls.onGroup = handler
	ls.onUnGroup = handlerU
}

func (ls *WTreeStore) GetColIdx(n string) int {
	if i, ok := ls.colNames[n]; ok {
		return i
	}
	log.Println("GetColIdx Unknown column:", n, ls.colNames)
	return -1
}

func (ls *WTreeStore) Clear() {
	ls.TreeStore.Clear()
}
func TreeStoreGetString(ls *gtk.TreeStore, iter *gtk.TreeIter, col int) string {
	if val, err := ls.GetValue(iter, col); err == nil {
		if str, err := val.GetString(); err == nil {
			return str
		}
	}
	return ""
}
func TreeStoreGetInt(ls *gtk.TreeStore, iter *gtk.TreeIter, col int) int {
	if val, err := ls.GetValue(iter, col); err == nil {
		if gv, err := val.GoValue(); err == nil {
			switch v := gv.(type) {
			case int:
				return v
			default:
				log.Printf("TreeStoreGetInt: invalid type:%T\n", v)
			}
		}
	}
	return 0
}
func TreeStoreGetBool(ls *gtk.TreeStore, iter *gtk.TreeIter, col int) bool {
	if val, err := ls.GetValue(iter, col); err == nil {
		if gv, err := val.GoValue(); err == nil {
			switch v := gv.(type) {
			case bool:
				return v
			default:
				log.Printf("TreeStoreGetBool: invalid type:%T\n", v)
			}
		}
	}
	return false
}

func (wl *WTreeStore) insert(pos *gtk.TreeIter, data map[string]any) {
	for k, c := range data {
		if i := wl.GetColIdx(k); i >= 0 {
			err := wl.TreeStore.SetValue(pos, i, c)
			if err != nil {
				log.Printf("Col Insert Error: %v %T\n", err, c)
			}
		} else {
			panic(fmt.Sprintf("Internal error : Unknown column: %v %v", k, c))
		}
	}
	wl.sel.UnselectAll()
	wl.sel.SelectIter(pos)

}
func (wl *WTreeStore) AppendM(parent *gtk.TreeIter, data map[string]any) *gtk.TreeIter {
	p1 := wl.TreeStore.Append(parent)
	wl.insert(p1, data)
	return p1
}

/*
	func (wl *WTreeStore) InsertM(data map[string]any) {
		var p *gtk.TreeIter
		if _, p1, ok := wl.sel.GetSelected(); ok {
			p = wl.TreeStore.InsertAfter(nil, p1)
		} else {
			p = wl.TreeStore.Append(nil)
		}
		wl.insert(p, data)
	}
*/
func (m *WTreeStore) getLSValue(iter *gtk.TreeIter, c int) any {
	val, _ := m.TreeStore.GetValue(iter, c)
	r, _ := val.GoValue()
	return r

}
func (m *WTreeStore) GetValues() []map[string]any {
	fmt.Println("Get Values")
	res := make([]map[string]any, 0)
	m.TreeStore.ForEach(func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
		if model.IterHasChild(iter) {
			return false
		}
		row := make(map[string]any)
		values := m.getRow(iter)
		for i, c := range m.columns {
			row[c.Title()] = values[i]
		}
		n := len(m.columns)
		row["_changed"] = m.getLSValue(iter, n)
		row["_ID"] = m.getLSValue(iter, n+1)

		res = append(res, row)
		return false
	})

	fmt.Println(res)
	return res
}
func (wl *WTreeStore) mkEditPopover() {
	wl.popover, _ = gtk.PopoverNew(wl.treeView)
	popoGrid, _ := gtk.GridNew()
	wl.popover.Add(popoGrid)

	for i, c := range wl.columns {
		wl.colNames[c.Title()] = i
		l, _ := gtk.LabelNewWithMnemonic(c.Title())
		popoGrid.Attach(l, 0, i, 2, 1)
		popoGrid.Attach(c.Widget(), 3, i, 5, 1)
	}

	set := MkButton("Set", func() {
		p1 := wl.TreeStore.Append(nil)
		for i, c := range wl.columns {
			v := c.GetValue()
			wl.TreeStore.SetValue(p1, i, v)
		}
		wl.popover.Hide()
	})

	cancel := MkButton("Cancel", func() {
		wl.popover.Hide()
	})
	y := len(wl.columns)
	popoGrid.Attach(set, 0, y, 3, 1)
	popoGrid.Attach(cancel, 3, y, 3, 1)
}
func (wl *WTreeStore) Expand() {
	wl.expand.Activate()
	wl.treeView.ExpandAll()
}
func (wl *WTreeStore) fillRow(iter *gtk.TreeIter, data []any) {
	for i, v := range data {
		wl.TreeStore.SetValue(iter, i, v)
	}
}
func (wl *WTreeStore) Degroup() {
	fmt.Println("Degroup")
	paths := wl.selection()
	if len(paths) != 1 {
		return
	}
	gpath := paths[0]
	giter := wl.iter(gpath)
	if !wl.TreeStore.IterHasChild(giter) {
		// nothing to ungroup
		return
	}
	fmt.Println("Degroup:", gpath)

	data := make([][]any, 0)

	chIter := &gtk.TreeIter{}
	if !wl.TreeStore.IterChildren(giter, chIter) {
		fmt.Println("Cannot iter children")
		return
	}
	fmt.Println("Copy && remove")
	for {
		v := wl.getRow(chIter)
		data = append(data, v)
		if !wl.TreeStore.Remove(chIter) {
			break
		}
	}
	fmt.Println(data)
	if wl.onUnGroup != nil {
		wl.onUnGroup(wl, data)
	}

	ip := wl.iter(gpath)

	for _, row := range data {
		ip = wl.TreeStore.InsertAfter(nil, ip)
		wl.fillRow(ip, row)
	}
	wl.TreeStore.Remove(wl.iter(gpath))

}

func (wl *WTreeStore) getRow(iter *gtk.TreeIter) []any {
	val := make([]any, wl.ColCount)
	for i := 0; i < wl.ColCount; i++ {
		v, _ := wl.TreeStore.GetValue(iter, i)
		val[i], _ = v.GoValue()
	}
	return val
}
func (wl *WTreeStore) iter(p *gtk.TreePath) *gtk.TreeIter {
	iter, _ := wl.TreeStore.GetIter(p)
	return iter
}
func (wl *WTreeStore) MakeGroup() {
	fmt.Println("Group")
	liter := wl.selection()
	if len(liter) <= 1 {
		return
	}

	// Group
	// No subgroups
	for _, p := range liter {
		iter := wl.iter(p)
		if wl.TreeStore.IterHasChild(iter) {
			return
		}
	}
	fmt.Println("Insert Before")
	p0, _ := liter[0].Copy()
	// Copy data
	data := make([][]any, len(liter))
	for i, p := range liter {
		data[i] = wl.getRow(wl.iter(p))
	}

	fmt.Println("Delete")
	for i := range liter {
		wl.TreeStore.Remove(wl.iter(liter[len(liter)-i-1]))
	}

	fmt.Println("Fill")
	var ds *gtk.TreeIter
	if p0.Prev() {
		ds = wl.TreeStore.InsertAfter(nil, wl.iter(p0))
	} else {
		ds = wl.TreeStore.Insert(nil, 0)
	}
	if wl.onGroup != nil {
		wl.fillRow(ds, wl.onGroup(wl, data))
	}
	for i, d := range data {
		if i == 0 {
			ds = wl.TreeStore.InsertAfter(ds, nil)
		} else {
			ds = wl.TreeStore.InsertAfter(nil, ds)
		}

		wl.fillRow(ds, d)
	}

}
func (wl *WTreeStore) selection() []*gtk.TreePath {
	lst := wl.sel.GetSelectedRows(wl.TreeStore)
	liter := make([]*gtk.TreePath, 0)
	lst.Foreach(func(v any) {
		if iter, ok := v.(*gtk.TreePath); ok {
			liter = append(liter, iter)
		}

	})
	return liter
}

func WTreeStoreNew(container *gtk.ScrolledWindow, columns []ITreeStoreColumn, withAdd bool) (*WTreeStore, gtk.IWidget) {
	const CharW = 10

	wl := &WTreeStore{container: container, columns: columns, ColCount: len(columns) + 2}
	wl.colNames = make(map[string]int)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	wl.treeView, _ = gtk.TreeViewNew()
	wl.treeView.SetGridLines(gtk.TREE_VIEW_GRID_LINES_BOTH)

	lTypes := make([]glib.Type, len(columns))
	for i, c := range columns {
		lTypes[i] = c.Type()
	}
	lTypes = append(lTypes, []glib.Type{glib.TYPE_BOOLEAN, glib.TYPE_INT}...)
	ls, _ := gtk.TreeStoreNew(lTypes...)

	wl.TreeStore = ls

	wl.colNames["_ID"] = len(columns) + 1
	wl.colNames["_changed"] = len(columns)
	for i, c := range columns {
		wl.colNames[c.Title()] = i
		c.Self().Storage = ls

		if !strings.HasPrefix(c.Title(), "_") { // hidden col

			tvCol, _ := gtk.TreeViewColumnNewWithAttribute(c.Title(), c.Renderer(), c.Attribute(), i)

			tvCol.SetFixedWidth(c.Self().CharWidth * CharW)
			wl.treeView.AppendColumn(tvCol)
		}
	}
	wl.treeView.SetModel(ls)

	wl.sel, _ = wl.treeView.GetSelection()
	wl.sel.SetMode(gtk.SELECTION_MULTIPLE)

	if container != nil {
		container.Add(wl.treeView)
		box.PackStart(container, true, true, 0)
	} else {
		box.PackStart(wl.treeView, true, true, 0)
	}
	actionBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1)
	if withAdd {
		add, _ := gtk.ButtonNewWithLabel("Add...")
		add.Connect("clicked", func() {
			//			wl.currentIter = nil
			for _, c := range columns {
				c.SetValue(c.BlankValue())
			}
			wl.popover.Show()
		})
		actionBox.Add(add)
	}

	group := MkButton("Group", func() {
		wl.MakeGroup()
	})
	ungroup := MkButton("UnGroup", func() {
		wl.Degroup()
	})
	wl.expand, _ = gtk.ToggleButtonNewWithLabel("Expand")
	wl.expand.Connect("clicked", func() {
		if wl.expand.GetActive() {
			wl.treeView.ExpandAll()
		} else {
			wl.treeView.CollapseAll()
		}
	})

	up := MkButton("Up", func() {
		liter := wl.selection()
		if len(liter) > 0 {
			p0 := wl.iter(liter[0])
			pn := wl.iter(liter[len(liter)-1])

			if ls.IterPrevious(p0) {
				ls.MoveAfter(p0, pn)

			}
		}
	})
	down := MkButton("Down", func() {
		liter := wl.selection()
		if len(liter) > 0 {
			p0 := wl.iter(liter[0])
			pn := wl.iter(liter[len(liter)-1])
			if ls.IterNext(pn) {
				ls.MoveBefore(pn, p0)
			}
		}
	})
	delete := MkButton("Delete", func() {
		liter := wl.selection()
		for _, p := range liter {
			iter := wl.iter(p)
			ls.Remove(iter)
		}
	})
	actionBox.Add(up)
	actionBox.Add(down)
	actionBox.Add(delete)
	actionBox.Add(group)
	actionBox.Add(ungroup)
	actionBox.Add(wl.expand)
	box.PackStart(actionBox, false, false, 0)

	wl.treeView.Connect("row-activated", func(tv *gtk.TreeView, path *gtk.TreePath) {
		if wl.onActivate == nil {
			if wl.popover == nil {
				wl.mkEditPopover()
			}
			if iter, ok := ls.GetIter(path); ok == nil {
				for i, c := range columns {
					v, _ := ls.GetValue(iter, i)
					gv, _ := v.GoValue()
					c.SetValue(gv)
				}
			}
			wl.popover.ShowAll()
		} else {
			data := make(map[string]any)
			if iter, ok := ls.GetIter(path); ok == nil {
				for i, c := range columns {
					v, _ := ls.GetValue(iter, i)
					gv, _ := v.GoValue()
					data[c.Title()] = gv
				}
				wl.onActivate(data)
			}
		}
	})
	wl.sel.UnselectAll()
	return wl, box
}
