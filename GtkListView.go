// GtkHelper
package main

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// ListStore
type IListStoreColumn interface {
	Title() string
	Type() glib.Type
	Widget() gtk.IWidget
	Renderer() gtk.ICellRenderer
	SetValue(any)
	GetValue() any
	BlankValue() any
	Attribute() string
	Self() *ListStoreColumn
	SetWidth(int)
}

type ListStoreColumn struct {
	CTitle    string
	GType     glib.Type
	Attr      string
	CRenderer gtk.ICellRenderer
	Storage   *gtk.ListStore
	CharWidth int
	Column    int
	widget    gtk.IWidget
}

func (lc *ListStoreColumn) Self() *ListStoreColumn {
	return lc
}
func (lc *ListStoreColumn) Attribute() string {
	return lc.Attr
}
func (lc *ListStoreColumn) Title() string {
	return lc.CTitle
}
func (lc *ListStoreColumn) Widget() gtk.IWidget {
	return lc.widget
}

func (lc *ListStoreColumn) Type() glib.Type {
	return lc.GType
}
func (lc *ListStoreColumn) Renderer() gtk.ICellRenderer {
	return lc.CRenderer
}

// --------------------------------------------
type ListStoreColumnText struct {
	ListStoreColumn
	entry *gtk.Entry
}

func ListStoreColumnTextNew(title string, w int) *ListStoreColumnText {
	r, _ := gtk.CellRendererTextNew()
	entry, _ := gtk.EntryNew()
	entry.SetWidthChars(w)
	entry.SetEditable(true)

	return &ListStoreColumnText{
		ListStoreColumn: ListStoreColumn{
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
func (ls *ListStoreColumnText) SetValue(val any) {
	if v, ok := val.(string); ok {
		ls.entry.SetText(v)
	}
}
func (ls *ListStoreColumnText) GetValue() any {
	t, _ := ls.entry.GetText()
	return t
}

func (ls *ListStoreColumnText) BlankValue() any {
	return ""
}
func (ls *ListStoreColumnText) SetWidth(w int) {
	ls.widget.(*gtk.Entry).SetSizeRequest(w, -1)
}

// Int -------------------------------------------------------------------------
type ListStoreColumnInt struct {
	ListStoreColumn
	spin *gtk.SpinButton
	from int
}

func ListStoreColumnIntNew(title string, digits int, from int, to int, step int) *ListStoreColumnInt {
	r, _ := gtk.CellRendererSpinNew()
	r.SetProperty("digits", digits)

	spin, _ := gtk.SpinButtonNewWithRange(float64(from), float64(to), float64(step))

	return &ListStoreColumnInt{
		ListStoreColumn: ListStoreColumn{
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
func (ls *ListStoreColumnInt) SetValue(val any) {
	if v, ok := val.(int); ok {
		ls.spin.SetValue(float64(v))
	}
}
func (ls *ListStoreColumnInt) GetValue() any {
	return ls.spin.GetValueAsInt()
}
func (ls *ListStoreColumnInt) BlankValue() interface{} {
	return ls.from
}
func (ls *ListStoreColumnInt) SetWidth(w int) {
	ls.widget.(*gtk.SpinButton).SetSizeRequest(w, -1)
}

type ListStoreColumnIcon struct {
	ListStoreColumn
}

func (ls *ListStoreColumnIcon) Widget() gtk.IWidget {
	return nil
}
func ListStoreColumnIconNew(title string, digits int) *ListStoreColumnIcon {
	cr, _ := gtk.CellRendererPixbufNew()
	lc := ListStoreColumnIcon{
		ListStoreColumn: ListStoreColumn{
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

func (ls *ListStoreColumnIcon) SetValue(t interface{}) {
	panic("Unimplemented")
}

func (ls *ListStoreColumnIcon) GetValue() interface{} {
	return nil
}

func (ls *ListStoreColumnIcon) BlankValue() any {
	return 0
}
func (ls *ListStoreColumnIcon) SetWidth(w int) {
	//ls.w.SetSizeRequest(w, -1)
}

// Combo
type ListStoreColumnCombo struct {
	ListStoreColumn
	combo  *gtk.ComboBoxText
	values []string
}

func (ls *ListStoreColumnCombo) Widget() gtk.IWidget {
	return ls.combo
}
func ListStoreColumnComboNew(title string, w int, values []string) *ListStoreColumnCombo {
	lc := ListStoreColumnCombo{
		ListStoreColumn: ListStoreColumn{CTitle: title,
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
func (ls *ListStoreColumnCombo) SetValue(ti interface{}) {
	t := ti.(string)
	for i, v := range ls.values {
		if v == t {
			ls.combo.SetActive(i)
		}
	}
}
func (ls *ListStoreColumnCombo) GetValue() interface{} {
	t := ls.combo.GetActiveText()
	return t
}

func (ls *ListStoreColumnCombo) BlankValue() interface{} {
	return ls.values[0]
}
func (ls *ListStoreColumnCombo) SetWidth(w int) {
	ls.combo.SetSizeRequest(w, -1)
}

/*
	type ListStoreColumnToggle struct {
		ListStoreColumn
		cb    *gtk.CheckButton
		value bool
	}

	func (ls *ListStoreColumnToggle) Widget() gtk.IWidget {
		return ls.cb
	}

	func (ls *ListStoreColumnToggle) Attribute() string {
		return "active"
	}

	func ListStoreColumnToggleNew(title string) *ListStoreColumnToggle {
		lc := ListStoreColumnToggle{
			ListStoreColumn: ListStoreColumn{CTitle: title, GType: glib.TYPE_BOOLEAN},
		}
		lc.cb, _ = gtk.CheckButtonNew()
		cr, _ := gtk.CellRendererToggleNew()
		cr.SetActivatable(true)
		cr.SetRadio(true)
		cr.Connect("toggled", func(cr *gtk.CellRendererToggle, path string) {
			iter, _ := lc.Storage.GetIterFromString(path)
			v, _ := lc.Storage.GetValue(iter, lc.Column)
			gv, _ := v.GoValue()
			val := gv.(bool)
			lc.Storage.SetValue(iter, lc.Column, !val)
		})
		lc.CRenderer = cr
		return &lc
	}

	func (ls *ListStoreColumnToggle) SetValue(ti interface{}) {
		ls.value = ti.(bool)
		ls.cb.SetActive(ls.value)
	}

	func (ls *ListStoreColumnToggle) GetValue() interface{} {
		t := ls.cb.GetActive()
		return t
	}

	func (ls *ListStoreColumnToggle) BlankValue() interface{} {
		return false
	}
*/
type WListStore struct {
	container *gtk.ScrolledWindow
	treeView  *gtk.TreeView
	popover   *gtk.Popover
	columns   []IListStoreColumn
	ListStore *gtk.ListStore

	colNames   map[string]int
	ColCount   int
	sel        *gtk.TreeSelection
	onActivate func(map[string]interface{})
}

func (ls *WListStore) SetActivate(handler func(map[string]interface{})) {
	ls.onActivate = handler
}

func (ls *WListStore) GetColIdx(n string) int {
	if i, ok := ls.colNames[n]; ok {
		return i
	}
	log.Println("GetColIdx Unknown column:", n, ls.colNames)
	return -1
}

func (ls *WListStore) Clear() {
	ls.ListStore.Clear()
}
func ListStoreGetString(ls *gtk.ListStore, iter *gtk.TreeIter, col int) string {
	if val, err := ls.GetValue(iter, col); err == nil {
		if str, err := val.GetString(); err == nil {
			return str
		}
	}
	return ""
}
func ListStoreGetInt(ls *gtk.ListStore, iter *gtk.TreeIter, col int) int {
	if val, err := ls.GetValue(iter, col); err == nil {
		if gv, err := val.GoValue(); err == nil {
			switch v := gv.(type) {
			case int:
				return v
			default:
				log.Printf("ListStoreGetInt: invalid type:%T\n", v)
			}
		}
	}
	return 0
}
func ListStoreGetBool(ls *gtk.ListStore, iter *gtk.TreeIter, col int) bool {
	if val, err := ls.GetValue(iter, col); err == nil {
		if gv, err := val.GoValue(); err == nil {
			switch v := gv.(type) {
			case bool:
				return v
			default:
				log.Printf("ListStoreGetBool: invalid type:%T\n", v)
			}
		}
	}
	return false
}

func (wl *WListStore) insert(pos *gtk.TreeIter, data map[string]any) {
	for k, c := range data {
		if i := wl.GetColIdx(k); i >= 0 {
			err := wl.ListStore.SetValue(pos, i, c)
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
func (wl *WListStore) AppendM(data map[string]any) {
	p1 := wl.ListStore.Append()
	wl.insert(p1, data)
}
func (wl *WListStore) InsertM(data map[string]any) {
	var p *gtk.TreeIter
	if _, p1, ok := wl.sel.GetSelected(); ok {
		p = wl.ListStore.InsertAfter(p1)
	} else {
		p = wl.ListStore.Append()
	}
	wl.insert(p, data)
}
func (m *WListStore) getLSValue(iter *gtk.TreeIter, c int) interface{} {
	val, _ := m.ListStore.GetValue(iter, c)
	r, _ := val.GoValue()
	return r

}
func (m *WListStore) GetValues() []map[string]interface{} {
	res := make([]map[string]interface{}, 0)
	if iter, ok := m.ListStore.GetIterFirst(); ok {
		for {
			row := make(map[string]interface{})
			for i, c := range m.columns {
				row[c.Title()] = m.getLSValue(iter, i)
			}
			n := len(m.columns)
			row["_changed"] = m.getLSValue(iter, n)
			row["_ID"] = m.getLSValue(iter, n+1)
			res = append(res, row)
			if !m.ListStore.IterNext(iter) {
				break
			}
		}
	}
	return res
}
func (wl *WListStore) mkEditPopover() {
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
		p1 := wl.ListStore.Append()
		for i, c := range wl.columns {
			v := c.GetValue()
			wl.ListStore.SetValue(p1, i, v)
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

func (wl *WListStore) GroupAction() {

}
func (wl *WListStore) selection() []*gtk.TreePath {
	lst := wl.sel.GetSelectedRows(wl.ListStore)
	liter := make([]*gtk.TreePath, 0)
	lst.Foreach(func(v any) {
		if iter, ok := v.(*gtk.TreePath); ok {
			liter = append(liter, iter)
		}
	})
	return liter
}
func (wl *WListStore) iter(p *gtk.TreePath) *gtk.TreeIter {
	iter, _ := wl.ListStore.GetIter(p)
	return iter
}

func WListStoreNew(container *gtk.ScrolledWindow, columns []IListStoreColumn, withAdd bool) (*WListStore, gtk.IWidget) {
	const CharW = 10

	wl := &WListStore{container: container, columns: columns, ColCount: len(columns)}
	wl.colNames = make(map[string]int)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	wl.treeView, _ = gtk.TreeViewNew()
	wl.treeView.SetGridLines(gtk.TREE_VIEW_GRID_LINES_BOTH)

	lTypes := make([]glib.Type, len(columns))
	for i, c := range columns {
		lTypes[i] = c.Type()
	}
	lTypes = append(lTypes, []glib.Type{glib.TYPE_BOOLEAN, glib.TYPE_INT}...)
	ls, _ := gtk.ListStoreNew(lTypes...)

	wl.ListStore = ls

	wl.colNames["_ID"] = len(columns) + 1
	wl.colNames["_changed"] = len(columns)
	for i, c := range columns {
		wl.colNames[c.Title()] = i
		c.Self().Storage = ls

		tvCol, _ := gtk.TreeViewColumnNewWithAttribute(c.Title(), c.Renderer(), c.Attribute(), i)

		tvCol.SetFixedWidth(c.Self().CharWidth * CharW)
		wl.treeView.AppendColumn(tvCol)
	}
	wl.treeView.SetModel(ls)

	wl.sel, _ = wl.treeView.GetSelection()
	wl.sel.SetMode(gtk.SELECTION_BROWSE)

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

	delete, _ := gtk.ButtonNewWithLabel("Delete")
	up, _ := gtk.ButtonNewWithLabel("Up")
	down, _ := gtk.ButtonNewWithLabel("Down")
	actionBox.Add(up)
	actionBox.Add(down)
	actionBox.Add(delete)
	box.PackStart(actionBox, false, false, 0)

	up.Connect("clicked", func() {
		liter := wl.selection()
		if len(liter) > 0 {
			p0 := wl.iter(liter[0])
			pn := wl.iter(liter[len(liter)-1])

			if ls.IterPrevious(p0) {
				ls.MoveAfter(p0, pn)

			}
		}

	})
	down.Connect("clicked", func() {
		liter := wl.selection()
		if len(liter) > 0 {
			p0 := wl.iter(liter[0])
			pn := wl.iter(liter[len(liter)-1])
			if ls.IterNext(pn) {
				ls.MoveBefore(pn, p0)
			}
		}

	})
	delete.Connect("clicked", func() {
		liter := wl.selection()
		for _, p := range liter {
			iter := wl.iter(p)
			ls.Remove(iter)
		}

	})
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
			data := make(map[string]interface{})
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
