// MPlayer
package main

import (
	"fmt"

	"math"
	"strconv"
	"time"

	"github.com/py60800/tunedb/internal/player"
	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zdb"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const Mp3GWidth int = 12

type Marker struct {
	Idx      int
	Location float64
	button   *gtk.Button
}

type Mp3PlayWidget struct {
	file string

	frame *gtk.Frame
	grid  *gtk.Grid

	mainCursor *gtk.DrawingArea // On main Window
	cursor     *gtk.DrawingArea
	fileName   *gtk.Entry

	Markers            []Marker
	signal             func(idx int, pos float64)
	signalMarkerChange func()
	hide               func()
	iSignal            int
	tickCbId           int

	from    *gtk.SpinButton
	to      *gtk.SpinButton
	adjFrom *gtk.Adjustment
	adjTo   *gtk.Adjustment

	From     float64
	To       float64
	Duration float64
	Position float64
	prevPos  float64

	markerBox *gtk.Grid

	position *gtk.Entry
	speed    *gtk.SpinButton
	pitch    *gtk.SpinButton
}

func triangle(cr *cairo.Context, xS float64, yS float64, base float64, height float64) {
	cr.MoveTo(xS, yS)
	cr.LineTo(xS-base/2, yS+height)
	cr.LineTo(xS+base/2, yS+height)
	cr.ClosePath()
	cr.Fill()

}
func (m *Mp3PlayWidget) SetSignal(signal func(int, float64)) {
	m.signal = signal
}
func (m *Mp3PlayWidget) SetHide(hide func()) {
	m.hide = hide
}
func (m *Mp3PlayWidget) RemoveLastMarker() {
	switch len(m.Markers) {
	case 0:
		// nothing
	case 1, 2:
		m.Markers = make([]Marker, 0)
		m.markerBox.RemoveRow(0)
	default:
		m.Markers = m.Markers[:len(m.Markers)-1]
		m.markerUpdate()
	}
}

func (m *Mp3PlayWidget) markerSet(b *gtk.Button) {
	l := len(m.Markers)
	name, _ := b.GetName()
	idx, _ := strconv.Atoi(name)
	switch {
	case l == 0:
	// nothing
	case l == 1:
		m.Markers[0].Location = m.Position
	case idx == 0 && l > 1: // First
		if m.Position <= m.Markers[1].Location {
			m.Markers[0].Location = m.Position
		}
	case idx == l-1 && l > 1: // Last
		if m.Position >= m.Markers[l-2].Location {
			m.Markers[l-1].Location = m.Position
		}
	default:
		if l < 3 {
			panic("Marker logic error => Report error")
		}
		if m.Position >= m.Markers[idx-1].Location && m.Position <= m.Markers[idx+1].Location {
			m.Markers[idx].Location = m.Position
		}

	}
	m.Markers[idx].Location = m.Position
	m.signalMarkerChange()
}
func (m *Mp3PlayWidget) markerUpdate() {
	m.markerBox.RemoveRow(0)
	for i, mrk := range m.Markers {
		mrk.button.SetName(strconv.Itoa(i))
		var label string
		switch i {
		case 0:
			label = "...[1"
		case len(m.Markers) - 1:
			label = fmt.Sprintf("%d]...", len(m.Markers)-1)
		default:
			label = fmt.Sprintf("%v][%v", i, i+1)
		}
		mrk.button.SetLabel(label)
		m.markerBox.Attach(mrk.button, 2*i, 0, 2, 1)

	}
	m.markerBox.ShowAll()
}
func (m *Mp3PlayWidget) AppendMarker(from, to float64) {
	idx := len(m.Markers)
	if idx == 0 {
		from = max(from, 0.0)
		b, _ := gtk.ButtonNew()
		b.Connect("clicked", func(b *gtk.Button) {
			m.markerSet(b)
		})
		m.Markers = append(m.Markers, Marker{Location: from, button: b})
	}
	if to < 0.0 {
		to = m.Duration
	}

	b, _ := gtk.ButtonNew()
	b.Connect("clicked", func(b *gtk.Button) {
		m.markerSet(b)
	})
	m.Markers = append(m.Markers, Marker{Location: to, button: b})
	m.markerUpdate()
}

func (m *Mp3PlayWidget) GetMarkers() []float64 {
	res := make([]float64, len(m.Markers))
	for i, m := range m.Markers {
		res[i] = m.Location
	}
	return res
}
func (m *Mp3PlayWidget) ResetMarker() {
	m.markerBox.RemoveRow(0)
	m.Markers = make([]Marker, 0)
}
func (m *Mp3PlayWidget) Player() *player.Mp3Player {
	return GetContext().mp3Player
}
func (m *Mp3PlayWidget) Reset() {
	m.ResetMarker()
	m.Player().Stop()
	m.SelectFile(nil, 0, 0)
}
func (p *Mp3PlayWidget) CursorDraw(da *gtk.DrawingArea, cr *cairo.Context) {
	w := float64(da.GetAllocatedWidth())
	h := float64(da.GetAllocatedHeight())
	cr.SetSourceRGB(128, 128, 128)
	cr.Rectangle(0, 0, w, h)
	cr.Fill()
	cr.SetSourceRGB(0, 0, 0)
	cr.SetLineWidth(1)
	cr.Rectangle(0, 0, w, h)
	cr.Stroke()

	cr.SetSourceRGB(0, 0, 255)
	cr.SetLineWidth(3)
	cr.MoveTo(0, h/2.)
	cr.LineTo(w, h/2.)
	cr.Stroke()
	cr.SetSourceRGB(255, 0, 0)
	cr.SetLineWidth(1.0)
	duration := p.Player().GetDuration()
	if duration < 1 {
		return
	}
	step := w / duration
	for i := 0; i < int(duration); i += 10 {
		cr.MoveTo(float64(i)*step, h/2-5)
		cr.LineTo(float64(i)*step, h/2+5)
		cr.Stroke()
	}
	cr.SetSourceRGB(0, 255, 0)
	triangle(cr, step*p.From, h/2, 8, -12)
	cr.SetSourceRGB(0, 255, 255)
	triangle(cr, step*p.To, h/2, 8, -12)
	for _, mk := range p.Markers {
		triangle(cr, step*mk.Location, h/2, 10, -20)
	}
	cr.SetSourceRGB(255, 0, 0)
	triangle(cr, step*p.Player().GetProgress(), h/2, 10, 20)
}
func (m *Mp3PlayWidget) play(from, to float64, mode int) {
	GetContext().Stop()
	if m.tickCbId != 0 {
		m.mainCursor.RemoveTickCallback(m.tickCbId)
		m.tickCbId = 0
	}
	m.prevPos = 36000.0
	duration := m.Player().GetDuration()

	m.setDuration(duration)
	if to < 0 {
		to = duration
	}
	m.Player().Play(from, to, mode)
	m.startTickCallBack(mode, from, to, duration)
}

func (m *Mp3PlayWidget) startTickCallBack(mode int, from, to, duration float64) {
	nextTo := time.Now().Add(100 * time.Millisecond)
	type event struct {
		timeOut float64
		idx     int
	}
	if to <= from {
		to = duration
	}
	schedule := make([]event, 0)
	for i, v := range m.Markers {
		if v.Location > (from-1) && v.Location < to-1 {
			schedule = append(schedule, event{timeOut: v.Location - 2.0, idx: i + 1})
		}
	}
	if mode == player.PMPlayRepeat {
		schedule = append(schedule, event{timeOut: to - 2.0, idx: 0})
	}
	m.iSignal = 0
	m.tickCbId = m.mainCursor.AddTickCallback(func(widget *gtk.Widget, frameClock *gdk.FrameClock) bool {
		m.Position = m.Player().GetProgress()
		if time.Now().After(nextTo) {

			nextTo = time.Now().Add(100 * time.Millisecond)
			m.Position = m.Player().GetProgress()
			m.cursor.QueueDraw()
			m.mainCursor.QueueDraw()
			m.position.SetText(fmt.Sprintf("%3.1f/%3.1f", m.Position, m.Duration))
			if len(m.Markers) > 0 && m.signal != nil {
				if m.Position < m.prevPos {
					m.signal(0, m.Position)
					m.iSignal = 0
				}
				m.prevPos = m.Position
				for i := m.iSignal; i < len(schedule); i++ {
					if schedule[i].timeOut < m.Position {
						m.signal(schedule[i].idx, m.Position)
						m.iSignal = i + 1
						break
					}
				}
			}
		}
		stop := m.Player().IsPlaying()
		return stop
	})

}
func fRound(d float64) float64 {
	return math.Round(d*10.0) / 10.0
}
func Mp3PlayWidgetNew(signalMarkerChange func(), mainCursor *gtk.DrawingArea) (*Mp3PlayWidget, gtk.IWidget) {
	m := &Mp3PlayWidget{}
	m.frame, _ = gtk.FrameNew("Mp3 player")
	grid, _ := gtk.GridNew()
	m.frame.Add(grid)
	SetMargins(grid, 5, 5)
	m.fileName, _ = gtk.EntryNew()
	m.fileName.SetEditable(false)
	is := 0
	grid.Attach(m.fileName, 0, 0, 8, 1)
	is++
	//	hbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	iw := 0
	pgrid := func(w gtk.IWidget, width int) {
		grid.Attach(w, iw, is, width, 1)
		iw += width
	}
	// Plain play
	play := MkButtonIcon("media-playback-start", func() {
		m.play(0.0, -1.0, player.PMPlayOnce)

	})
	pgrid(play, 1)

	repeat := MkButtonIcon("media-playlist-repeat", func() {
		if m.hide != nil {
			m.hide()
			DelayedAction(m.frame, 1*time.Second, func() {
				m.play(0.0, -1.0, player.PMPlayRepeat)

			})
		} else {
			m.play(0.0, -1.0, player.PMPlayRepeat)

		}
	})
	pgrid(repeat, 1)
	stop := MkButtonIcon("media-playback-stop", func() {
		m.Player().Stop()
	})
	pgrid(stop, 1)
	lSpeed, _ := gtk.LabelNew("S:")
	pgrid(lSpeed, 1)
	m.speed, _ = gtk.SpinButtonNewWithRange(0.25, 2.0, 0.01)
	m.speed.Connect("value-changed", func(sb *gtk.SpinButton) {
		m.Player().SetTimeRatio(1.0 / sb.GetValue())
	})
	m.speed.SetValue(1.0)
	pgrid(m.speed, 1)

	lPitch, _ := gtk.LabelNew("P:")
	pgrid(lPitch, 1)
	m.pitch, _ = gtk.SpinButtonNewWithRange(-2.5, 2.5, 0.1)
	m.pitch.SetValue(0.0)
	m.pitch.Connect("value-changed", func(sb *gtk.SpinButton) {
		pitchScale := sb.GetValue()
		if pitchScale == 0.0 {
			m.Player().SetPitchScale(1.0) // make sure exact value
		} else {
			m.Player().SetPitchScale(math.Pow(math.Pow(2.0, 1.0/12.0), pitchScale))
		}
	})
	pgrid(m.pitch, 1)

	m.position, _ = gtk.EntryNew()
	m.position.SetEditable(false)
	pgrid(m.position, 1)

	is++

	currentPosition, _ := gtk.EntryNew()
	currentPosition.SetEditable(false)

	m.cursor, _ = gtk.DrawingAreaNew()
	m.cursor.AddEvents(4) // EVENT_MOTION_NOTIFY
	m.cursor.SetHAlign(gtk.ALIGN_FILL)
	m.cursor.SetVAlign(gtk.ALIGN_FILL)
	m.cursor.SetHExpand(true)
	m.cursor.SetVExpand(true)
	m.cursor.SetSizeRequest(400, 50)
	m.cursor.AddEvents(int(gdk.BUTTON_PRESS_MASK))
	m.cursor.QueueDraw()

	m.mainCursor = mainCursor
	m.mainCursor.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
		m.CursorDraw(da, cr)
	})

	m.Position = 0.0
	// Direct
	m.cursor.Connect("button-press-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
		evb := gdk.EventButtonNewFromEvent(ev)
		if evb.Type() == 4 {
			w := da.GetAllocatedWidth()
			start := (evb.X() * m.Player().GetDuration()) / float64(w)
			m.play(start, 0.0, player.PMPlayOnce)
			return true
		}
		return false
	})
	m.cursor.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
		m.CursorDraw(da, cr)
	})
	grid.Attach(m.cursor, 0, is, 8, 1)
	is++

	// Partial playing
	iw = 0
	pPLabel := MkLabel("Selection")
	pgrid(pPLabel, 1)
	pPplay := MkButtonIcon("media-playback-start", func() {
		m.play(m.From, m.To, player.PMPlayOnce)
	})
	pgrid(pPplay, 1)

	pPrepeat := MkButtonIcon("media-playlist-repeat", func() {
		if m.hide != nil {
			m.hide()
			DelayedAction(m.frame, 1*time.Second, func() {
				m.play(m.From, m.To, player.PMPlayRepeat)

			})
		} else {
			m.play(m.From, m.To, player.PMPlayRepeat)
		}
	})
	pgrid(pPrepeat, 1)
	pPstop := MkButtonIcon("media-playback-stop", func() {
		m.Player().Stop()
	})
	pgrid(pPstop, 1)

	// To From Button
	bFrom := MkButton("From", func() {
		m.From = m.Position
		m.from.SetValue(fRound(m.From))
	})
	m.adjFrom, _ = gtk.AdjustmentNew(0.0, 0.0, 100.0, 0.1, 1.0, 0)
	m.adjTo, _ = gtk.AdjustmentNew(0.0, 0.0, 100.0, 0.1, 1.0, 0)
	m.from, _ = gtk.SpinButtonNew(m.adjFrom, 0.1, 1)
	m.from.SetSensitive(false)
	pgrid(bFrom, 1)
	pgrid(m.from, 1)
	bTo := MkButton("To", func() {
		m.To = m.Position
		m.to.SetValue(fRound(m.To))

	})
	m.to, _ = gtk.SpinButtonNew(m.adjTo, 0.1, 1)
	m.to.SetSensitive(false)

	m.from.Connect("value-changed", func(sb *gtk.SpinButton) {
		m.From = sb.GetValue()
	})
	m.to.Connect("value-changed", func(sb *gtk.SpinButton) {
		m.To = sb.GetValue()
	})
	pgrid(bTo, 1)
	pgrid(m.to, 1)

	is++
	// Markers
	if signalMarkerChange != nil {
		m.markerBox, _ = gtk.GridNew()
		grid.Attach(m.markerBox, 0, is, 8, 1)
		is++
		m.signalMarkerChange = signalMarkerChange
	}

	return m, m.frame
}
func (m *Mp3PlayWidget) GetBounds() (float64, float64) {
	return m.From, m.To
}
func (m *Mp3PlayWidget) SelectFile(mp3file *zdb.MP3File, from, to float64) {
	if mp3file != nil {
		m.file = mp3file.File
		m.From = from
		m.To = to
		m.fileName.SetText(mp3file.Title)
		m.fileName.SetTooltipText(fmt.Sprintf("Artist:%v\nAlbum:%v\nTitle:%v\nFile:%v",
			mp3file.Artist, mp3file.Album, mp3file.Title, mp3file.File))

	} else {
		m.file = ""
		m.setDuration(0.5)
	}
	if m.file != "" {
		m.Duration, _ = m.Player().LoadFile(m.file)
	}
	m.setDuration(m.Duration)
	m.from.SetValue(fRound(m.From))
	m.to.SetValue(fRound(m.To))
	m.position.SetText(fmt.Sprintf("%3.1f/%3.1f", 0.0, m.Duration))
	m.speed.SetValue(1.0)
	m.pitch.SetValue(0.0)
}
func (p *Mp3PlayWidget) setDuration(d float64) {
	p.Duration = d
	if p.Duration > 1.0 {
		if p.To <= p.From {
			p.To = p.Duration
		}
		p.adjFrom, _ = gtk.AdjustmentNew(p.From, 0.0, p.Duration, 0.1, 1.0, 0)
		p.adjTo, _ = gtk.AdjustmentNew(p.To, 0.0, p.Duration, 0.1, 1.0, 0)
		p.from.SetAdjustment(p.adjFrom)
		p.to.SetAdjustment(p.adjTo)
		p.from.SetSensitive(true)
		p.to.SetSensitive(true)
	} else {
		p.from.SetSensitive(false)
		p.to.SetSensitive(false)
	}
}

type Mp3Selector struct {
	searchText    *gtk.SearchEntry
	searchResult  *gtk.ComboBoxText
	withContent   *gtk.CheckButton
	suspendChange bool
	mp3Files      []zdb.MP3File
}

func (m *Mp3Selector) updateResults() {
	m.searchResult.RemoveAll()
	m.mp3Files = util.Truncate(m.mp3Files, 100)

	for _, f := range m.mp3Files {
		cnt := " "
		if f.HasContent {
			cnt = "*"
		}
		txt := fmt.Sprintf("%s%s: %s", cnt, f.Artist, f.Title)
		if len(txt) > 70 {
			txt = txt[:70]
		}
		m.searchResult.AppendText(txt)
	}
	m.searchResult.SetActive(0)

}
func (m *Mp3Selector) doSearch(what string) {
	withContent := m.withContent.GetActive()
	m.mp3Files = GetContext().mp3Collection.Search(what, withContent, 100)
	m.updateResults()
}
func (m *Mp3Selector) doSearchCurrent() {
	m.suspendChange = true
	if tune := ActiveTune(); tune != nil {
		title := tune.Title

		m.searchText.SetText(title)

		l := GetContext().DB.Mp3SetGetByTuneID(tune.ID)
		if len(l) > 0 {
			m.mp3Files = GetContext().mp3Collection.GetByIds(l)
			m.updateResults()
		} else {
			m.doSearch(title)
		}
	}
	m.suspendChange = false
}

func MkMp3Selector(selectFile func(mp3 *zdb.MP3File)) (*Mp3Selector, gtk.IWidget) {
	m := &Mp3Selector{}
	frame, _ := gtk.FrameNew("Mp3 Selector")
	grid, _ := gtk.GridNew()
	SetMargins(grid, 5, 5)
	frame.Add(grid)
	is := 0
	/*	lFileChooser, _ := gtk.LabelNewWithMnemonic("File Search")
		fileChooser, _ := gtk.FileChooserButtonNew("Select file", gtk.FILE_CHOOSER_ACTION_OPEN)
		homeDir, _ := os.UserHomeDir()
		fileChooser.SetCurrentFolder(homeDir)

		fileChooser.Connect("file-set", func() {
			file := fileChooser.GetFilename()
			if strings.HasSuffix(file, ".mp3") {
				mp3File := GetContext().mp3Collection.GetByFileName(file)
				if mp3File != nil {
					selectFile(mp3File)
				} else {
					Message(fmt.Sprintf("File %s not registred", file))
				}
			}
		})
		grid.Attach(lFileChooser, 0, is, 2, 1)
		grid.Attach(fileChooser, 2, is, Mp3GWidth-2, 1)
		is++
	*/

	// Search entry
	search := MkButton("Search", func() {
		txt, _ := m.searchText.GetText()
		m.doSearch(txt)

	})
	m.searchResult, _ = gtk.ComboBoxTextNew()

	m.searchText = MkDeferedSearchEntry(&m.suspendChange, func(what string) {
		m.doSearch(what)
	})
	bCurrent := MkButton("Cur", func() {
		m.doSearchCurrent()
	})
	bScan := MkButton("Scan", func() {
		GetContext().ScanMp3()
	})
	m.withContent, _ = gtk.CheckButtonNewWithLabel("W/ Content")
	grid.Attach(m.searchText, 0, is, Mp3GWidth-5, 1)
	grid.Attach(search, Mp3GWidth-5, is, 1, 1)
	grid.Attach(bCurrent, Mp3GWidth-4, is, 1, 1)
	grid.Attach(bScan, Mp3GWidth-3, is, 1, 1)
	grid.Attach(m.withContent, Mp3GWidth-2, is, 2, 1)
	m.withContent.Connect("toggled", func() {
		txt, _ := m.searchText.GetText()
		m.doSearch(txt)

	})
	is++

	sel := MkButton("Select", func() {
		i := m.searchResult.GetActive()
		if i >= 0 && i <= len(m.mp3Files) {
			selectFile(&m.mp3Files[i])
		}
	})
	grid.Attach(m.searchResult, 0, is, Mp3GWidth-2, 1)
	grid.Attach(sel, Mp3GWidth-2, is, 2, 1)
	m.searchResult.SetHExpand(true)
	selB := MkButton("Sel", func() {
		i := m.searchResult.GetActive()
		if i >= 0 && i <= len(m.mp3Files) {
			selectFile(&m.mp3Files[i])
		}
	})
	grid.Attach(selB, Mp3GWidth-2, is, 2, 1)
	is++
	frame.SetHExpand(true)
	grid.SetHExpand(true)
	return m, frame
}
func (m *Mp3Selector) Preselect(id int) {
	if id == 0 {
		return
	}
	ids := GetContext().DB.Mp3SetGetByTuneID(id)
	if len(ids) > 0 {
		m.mp3Files = GetContext().mp3Collection.GetByIds(ids)
		m.updateResults()
	}
}
func (m *Mp3Selector) SetSearchText(txt string) {
	m.suspendChange = true
	m.searchText.SetText(txt)
	m.suspendChange = false
}
