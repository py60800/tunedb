// GtkHelper
package main

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// ListStore
type IListStoreColumn interface {
	Title() string
	Type() glib.Type
	Widget() gtk.IWidget
	Renderer() gtk.ICellRenderer
	SetValue(interface{})
	GetValue() interface{}
	BlankValue() interface{}
	Self() *ListStoreColumn
	SetWidth(int)
}

type ListStoreColumn struct {
	CTitle    string
	GType     glib.Type
	CRenderer gtk.ICellRenderer
	Storage   *gtk.ListStore
	CharWidth int
	Column    int
}

func (lc *ListStoreColumn) Self() *ListStoreColumn {
	return lc
}
func (lc *ListStoreColumn) Title() string {
	return lc.CTitle
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

func (ls *ListStoreColumnText) Widget() gtk.IWidget {
	return ls.entry
}
func ListStoreColumnTextNew(title string, w int) *ListStoreColumnText {
	lc := ListStoreColumnText{
		ListStoreColumn: ListStoreColumn{
			CTitle:    title,
			GType:     glib.TYPE_STRING,
			CharWidth: w,
		},
	}
	lc.entry, _ = gtk.EntryNew()
	lc.entry.SetWidthChars(w)
	lc.entry.SetEditable(true)
	lc.CRenderer, _ = gtk.CellRendererTextNew()
	return &lc
}
func (ls *ListStoreColumnText) SetValue(t interface{}) {
	ls.entry.SetText(t.(string))
}
func (ls *ListStoreColumnText) GetValue() interface{} {
	t, _ := ls.entry.GetText()
	return t
}
func (ls *ListStoreColumnText) BlankValue() interface{} {
	return ""
}
func (ls *ListStoreColumnText) SetWidth(w int) {
	ls.entry.ToWidget().SetSizeRequest(w, -1)
}

// Int
type ListStoreColumnInt struct {
	ListStoreColumn
	spin *gtk.SpinButton
	from int
}

func (ls *ListStoreColumnInt) Widget() gtk.IWidget {
	return ls.spin
}
func ListStoreColumnIntNew(title string, digits int, from int, to int, step int) *ListStoreColumnInt {
	lc := ListStoreColumnInt{
		ListStoreColumn: ListStoreColumn{CTitle: title, GType: glib.TYPE_INT, CharWidth: digits + 4},
	}
	lc.spin, _ = gtk.SpinButtonNewWithRange(float64(from), float64(to), float64(step))

	lc.from = from
	r, _ := gtk.CellRendererSpinNew()
	r.SetProperty("digits", digits)
	lc.CRenderer = r
	return &lc
}
func (ls *ListStoreColumnInt) SetValue(t interface{}) {
	ls.spin.SetValue(float64(t.(int)))
}
func (ls *ListStoreColumnInt) GetValue() interface{} {
	v := ls.spin.GetValue()
	return int(v)
}
func (ls *ListStoreColumnInt) BlankValue() interface{} {
	return ls.from
}
func (ls *ListStoreColumnInt) SetWidth(w int) {
	ls.spin.SetSizeRequest(w, -1)
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
		ListStoreColumn: ListStoreColumn{CTitle: title, GType: glib.TYPE_STRING, CharWidth: w + 4},
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
	fmt.Println("Unknown column:", n, ls.colNames)
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
				fmt.Printf("ListStoreGetInt: invalid type:%T\n", v)
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
				fmt.Printf("ListStoreGetBool: invalid type:%T\n", v)
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
				fmt.Printf("Error: %v %T\n", err, c)
			}
		} else {
			panic(fmt.Sprintf("Internal error : Unknown column: %v %v", k, c))
		}
	}
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

	wl.popover, _ = gtk.PopoverNew(wl.treeView)
	popoGrid, _ := gtk.GridNew()
	wl.popover.Add(popoGrid)

	wl.colNames["_ID"] = len(columns) + 1
	wl.colNames["_changed"] = len(columns)
	for i, c := range columns {
		wl.colNames[c.Title()] = i
		l, _ := gtk.LabelNewWithMnemonic(c.Title())
		popoGrid.Attach(l, 0, i, 2, 1)
		popoGrid.Attach(c.Widget(), 3, i, 5, 1)
		c.Self().Storage = ls

		tvCol, _ := gtk.TreeViewColumnNewWithAttribute(c.Title(), c.Renderer(), "text", i)
		tvCol.SetFixedWidth(c.Self().CharWidth * CharW)
		wl.treeView.AppendColumn(tvCol)
	}
	wl.treeView.SetModel(ls)
	set, _ := gtk.ButtonNewWithLabel("Set")
	cancel, _ := gtk.ButtonNewWithLabel("Cancel")
	popoGrid.Attach(set, 0, len(columns), 3, 1)
	popoGrid.Attach(cancel, 3, len(columns), 3, 1)
	cancel.Connect("clicked", func() {
		wl.popover.Hide()
	})
	set.Connect("clicked", func() {
		p1 := wl.ListStore.Append()
		for i, c := range columns {
			v := c.GetValue()
			wl.ListStore.SetValue(p1, i, v)
		}
		wl.popover.Hide()
	})

	wl.sel, _ = wl.treeView.GetSelection()
	wl.sel.SetMode(gtk.SELECTION_BROWSE)
	wl.sel.Connect("changed", func(s *gtk.TreeSelection) {
		if _, p1, ok := s.GetSelected(); ok {
			for i, c := range columns {
				v, _ := ls.GetValue(p1, i)
				gv, _ := v.GoValue()
				c.SetValue(gv)
				w := c.Widget()
				w.Set("sensitive", true)
			}
		} else {
			for _, c := range columns {
				w := c.Widget()
				w.Set("sensitive", false)
			}
		}
	})
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
		if _, p1, ok := wl.sel.GetSelected(); ok {
			if _, p2, ok := wl.sel.GetSelected(); ok {
				if ls.IterPrevious(p2) {
					ls.Swap(p1, p2)
					//	wl.sel.SelectIter(psel)
				}
			}
		}
	})
	down.Connect("clicked", func() {
		if _, p1, ok := wl.sel.GetSelected(); ok {
			if _, p2, ok := wl.sel.GetSelected(); ok {
				if ls.IterNext(p2) {
					ls.Swap(p1, p2)
					//wl.sel.SelectIter(psel)
				}
			}
		}
	})
	delete.Connect("clicked", func() {
		if _, p1, ok := wl.sel.GetSelected(); ok {
			ls.Remove(p1)
		}
	})
	wl.treeView.Connect("row-activated", func(tv *gtk.TreeView, path *gtk.TreePath) {
		fmt.Println("Row activated:", path)
		if wl.onActivate == nil {
			if iter, ok := ls.GetIter(path); ok == nil {
				for i, c := range columns {
					v, _ := ls.GetValue(iter, i)
					gv, _ := v.GoValue()
					c.SetValue(gv)
				}
			}
			wl.popover.Show()
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
	popoGrid.ShowAll()
	return wl, box
}
