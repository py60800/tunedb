// MidiPlay
package main

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/py60800/tunedb/internal/util"
	"github.com/py60800/tunedb/internal/zdb"
	"github.com/py60800/tunedb/internal/zique"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type MidiPlayCtrl struct {
	menuButton *gtk.MenuButton
	context    *ZContext
	popover    *gtk.Popover

	grid             *gtk.Grid
	instrumentNames  []string
	instrument       *gtk.ComboBoxText
	velocitySelector *gtk.ComboBoxText
	swingSelector    *gtk.ComboBoxText
	swingCoeff       *gtk.Scale
	drumSelector     *gtk.ComboBoxText
	speed            *gtk.Scale
	volume           *gtk.Scale
	drum             *gtk.Scale
	memoPlay         *gtk.Button

	Zique *zique.ZiquePlayer
}

func (pm *MidiPlayCtrl) Widget() gtk.IWidget {
	return pm.menuButton
}
func (pm *MidiPlayCtrl) SetTempo(tempo int, tuneKind string) {
	if tempo == 0 {
		tempo = GetContext().DB.TuneKindGetTempo(tuneKind)
	}
	pm.speed.SetValue(float64(tempo))
}
func (pm *MidiPlayCtrl) SetInstrument(instrument string) {
	if instrument == "" {
		instrument = "Acoustic Grand Piano"
	}
	for i, v := range pm.instrumentNames {
		if v == instrument {
			pm.instrument.SetActive(i)
			break
		}
	}
}
func (pm *MidiPlayCtrl) Stop() {
	pm.Zique.Stop()
}
func (c *ZContext) MkMidiPlayCtrl() (*MidiPlayCtrl, gtk.IWidget) {
	pm := &MidiPlayCtrl{}
	pm.context = c
	var msg string
	pm.Zique, msg = zique.ZiquePlayerNew(ConfigBase, util.H.Get("MidiPort"))
	if msg != "" {
		WarnOnStart(msg)
	}

	pm.menuButton, _ = gtk.MenuButtonNew()
	pm.menuButton.SetLabel("Midi Ctrl...")
	pm.popover, _ = gtk.PopoverNew(pm.menuButton)
	pm.menuButton.SetPopover(pm.popover)

	pm.grid, _ = gtk.GridNew()
	pm.grid.SetRowHomogeneous(true)
	pm.popover.Add(pm.grid)
	SetMargins(pm.grid, 5, 5)
	width := 6
	is := 0

	lInstrument, _ := gtk.FrameNew("Instrument")
	pm.grid.Attach(lInstrument, 0, is, width, 1)
	is++
	//	var err error
	pm.instrument, _ = gtk.ComboBoxTextNew()

	pm.instrumentNames = make([]string, 0, len(zdb.RPatch))
	for k := range zdb.RPatch {
		pm.instrumentNames = append(pm.instrumentNames, k)
	}
	sort.Strings(pm.instrumentNames)
	for _, k := range pm.instrumentNames {
		pm.instrument.AppendText(k)

	}
	pm.instrument.Connect("changed", func(cb *gtk.ComboBoxText) {
		v := cb.GetActiveText()
		pm.Zique.SetPatch(v)
	})
	SetMargins(pm.instrument, 3, 3)
	lInstrument.Add(pm.instrument)

	//Drum
	lDrumSelector, _ := gtk.FrameNew("Drum Pattern")
	pm.grid.Attach(lDrumSelector, 0, is, width, 1)
	is++

	pm.drumSelector, _ = gtk.ComboBoxTextNew()
	for k := range zique.PatternConfig.DrumPattern {
		pm.drumSelector.AppendText(k)
	}
	pm.drumSelector.Connect("changed", func(d *gtk.ComboBoxText) {
		pm.Zique.SetDrumPattern(d.GetActiveText())
	})
	SetMargins(pm.drumSelector, 3, 3)
	lDrumSelector.Add(pm.drumSelector)

	//Swing
	lSwingSelector, _ := gtk.FrameNew("Swing Pattern")
	pm.grid.Attach(lSwingSelector, 0, is, width, 2)
	is += 2
	lSwingBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	SetMargins(lSwingBox, 3, 3)
	lSwingSelector.Add(lSwingBox)

	pm.swingSelector, _ = gtk.ComboBoxTextNew()
	for k := range zique.PatternConfig.SwingPattern {
		pm.swingSelector.AppendText(k)
	}
	pm.swingSelector.Connect("changed", func(d *gtk.ComboBoxText) {
		pm.swingCoeff.SetValue(1.0)
		pm.Zique.SetSwingPattern(d.GetActiveText())

	})
	SetMargins(pm.swingSelector, 3, 3)
	lSwingBox.Add(pm.swingSelector)

	pm.swingCoeff, _ = gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 0, 2, 0.1)
	pm.swingCoeff.SetValue(1.0)
	pm.swingCoeff.Connect("value-changed", func(s *gtk.Scale) {
		pm.Zique.AlterSwingPattern(s.GetValue())
	})
	SetMargins(pm.swingCoeff, 3, 3)
	lSwingBox.Add(pm.swingCoeff)

	//Velocity
	lVelocitySelector, _ := gtk.FrameNew("Velocity Pattern")
	pm.grid.Attach(lVelocitySelector, 0, is, width, 2)
	is += 2
	lVelocityBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	lVelocitySelector.Add(lVelocityBox)

	pm.velocitySelector, _ = gtk.ComboBoxTextNew()
	for k := range zique.PatternConfig.VelocityPattern {
		pm.velocitySelector.AppendText(k)
	}
	fmt.Println("Velocity Pattern", zique.PatternConfig.VelocityPattern)
	lVelocityBox.Add(pm.velocitySelector)

	velocityCoeff, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 0, 3.0, 0.1)
	velocityCoeff.SetValue(1.0)
	velocityCoeff.Connect("value-changed", func(s *gtk.Scale) {
		pm.Zique.AlterVelocityPattern(s.GetValue())
	})
	pm.velocitySelector.Connect("changed", func(d *gtk.ComboBoxText) {
		velocityCoeff.SetValue(1.0)
		pm.Zique.SetVelocityPattern(d.GetActiveText())
	})
	lVelocityBox.Add(velocityCoeff)

	cursorGrid, _ := gtk.GridNew()
	cursorGrid.SetColumnHomogeneous(true)
	H := 5
	//Speed
	lSpeed, _ := gtk.LabelNew("Tempo")
	cursorGrid.Attach(lSpeed, 0, 0, 1, 1)
	pm.speed, _ = gtk.ScaleNewWithRange(gtk.ORIENTATION_VERTICAL, 10, 240, 1)
	pm.speed.SetValue(120)
	pm.speed.SetInverted(true)
	pm.speed.SetVExpand(true)
	pm.speed.Connect("value-changed", func(sp *gtk.Scale) {
		pm.Zique.SetTempo(int(sp.GetValue()))

	})
	cursorGrid.AttachNextTo(pm.speed, lSpeed, gtk.POS_BOTTOM, 1, H)

	// Volume
	lVolume, _ := gtk.LabelNew("Vol")
	cursorGrid.Attach(lVolume, 2, 0, 1, 1)
	pm.volume, _ = gtk.ScaleNewWithRange(gtk.ORIENTATION_VERTICAL, 0, 100, 1)
	pm.volume.SetValue(80)
	pm.volume.SetInverted(true)
	pm.volume.Connect("value-changed", func(d *gtk.Scale) {
		pm.Zique.SetMainVolume(int(d.GetValue()))
	})
	pm.volume.SetVExpand(true)
	cursorGrid.AttachNextTo(pm.volume, lVolume, gtk.POS_BOTTOM, 1, H)

	// Drum
	lDrum, _ := gtk.LabelNew("Drum")
	cursorGrid.Attach(lDrum, 4, 0, 1, 1)
	pm.drum, _ = gtk.ScaleNewWithRange(gtk.ORIENTATION_VERTICAL, 0, 100, 1)
	pm.drum.SetValue(0)
	pm.drum.SetInverted(true)
	pm.drum.SetVExpand(true)

	pm.drum.Connect("value-changed", func(d *gtk.Scale) {
		pm.Zique.SetDrumVolume(int(d.GetValue()))
	})
	cursorGrid.AttachNextTo(pm.drum, lDrum, gtk.POS_BOTTOM, 1, H)

	pm.grid.Attach(cursorGrid, 0, is, width, H+1)
	is += H + 1

	roll, _ := gtk.ToggleButtonNewWithLabel("Roll")
	roll.Connect("clicked", func(b *gtk.ToggleButton) {
		zique.PlayRoll = b.GetActive()
	})
	pm.grid.Attach(roll, 0, is, 2, 1)

	pm.memoPlay, _ = gtk.ButtonNew()
	pm.memoPlay.SetLabel("Memorize")
	pm.memoPlay.Connect("clicked", func() {
		if t := ActiveTune(); t != nil {
			GetContext().Stop()
			t.Tempo = int(pm.speed.GetValue())
			t.Instrument = pm.instrument.GetActiveText()
			t.VelocityPattern = pm.velocitySelector.GetActiveText()
			t.SwingPattern = pm.swingSelector.GetActiveText()
			GetContext().DB.TuneMidiPlayUpdate(t)
		}
	})
	pm.grid.Attach(pm.memoPlay, width-2, is, 2, 1)
	pm.grid.SetBorderWidth(5)

	pm.grid.ShowAll()
	return pm, pm.menuButton
}

// *****************************************************************************
type TickEvent struct {
	date      time.Time
	TickStep  int
	TickCount int
}

type WMetronome struct {
	Metronome *gtk.DrawingArea
	TickCount int
	TickStep  int

	ticking bool

	pendingEvent []TickEvent
	nextEvent    int

	PassCount int
	MeasureId string
}

func (c *WMetronome) Ticker() {
	context := GetContext()
	select {
	case currentTune := <-context.midiPlayCtrl.Zique.FeedBack:
		if aTune := ActiveTune(); aTune != nil && aTune.Xml != currentTune {
			tune := context.DB.TuneGetByXmlFile(currentTune)
			context.LoadTune(&tune, true)
		}
	case tick := <-context.midiPlayCtrl.Zique.TickBack:
		c.pendingEvent = make([]TickEvent, 0)
		prev := 100 * time.Millisecond
		c.PassCount = tick.PassCount
		c.MeasureId = tick.MeasureId
		XD := time.Duration(tick.XmlDivisions)
		delay := tick.TickTime * XD
		switch tick.BeatType {
		case 1:
			delay = tick.TickTime * XD * 4
		case 2:
			//			undefined
			delay = tick.TickTime * XD * 2
		case 4:
			delay = tick.TickTime * XD
		case 8:
			delay = tick.TickTime * XD / 2
		case 16:
			delay = tick.TickTime * XD / 4

		}
		count := tick.Beats
		now := time.Now()
		for i := 0; i < count*2; i++ {
			tickN := i + 1
			c.pendingEvent = append(c.pendingEvent, TickEvent{
				date:      now.Add(time.Duration(tickN)*delay - prev),
				TickStep:  tickN % count,
				TickCount: count,
			})
		}
		cancelTime := now.Add(time.Second)
		if len(c.pendingEvent) > 0 {
			cancelTime = c.pendingEvent[len(c.pendingEvent)-1].date.Add(time.Second)
		}
		c.pendingEvent = append(c.pendingEvent, TickEvent{
			date:      cancelTime,
			TickStep:  0,
			TickCount: 0,
		})

		c.nextEvent = 0
		// Force start
		c.TickStep = 0
		c.Metronome.QueueDraw()
	default:
	}
	if c.nextEvent < len(c.pendingEvent) && time.Now().After(c.pendingEvent[c.nextEvent].date) {
		tickEvent := c.pendingEvent[c.nextEvent]
		c.TickCount = tickEvent.TickCount
		c.TickStep = tickEvent.TickStep
		c.Metronome.QueueDraw()
		c.nextEvent++
	}

}

func (c *WMetronome) MetronomeHide() {
	c.TickCount = 0
	c.Metronome.QueueDraw()
}
func (c *WMetronome) MetronomeShow() {

}
func (c *WMetronome) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	om := 2.0 * math.Pi

	X := float64(da.GetAllocatedWidth())
	Y := float64(da.GetAllocatedHeight())

	Sz := math.Min(X, Y)
	R := Sz/2 - 2.0
	Xc := X / 2 // Centre
	Yc := Y / 2
	if c.TickCount == 0 {
		cr.SetSourceRGB(127, 127, 127)
		cr.Rectangle(0, 0, X, Y)
		cr.Fill()
		return
	} else {

		dlt := -om / 4.0
		cr.SetLineWidth(3)

		if c.TickStep == 0 {
			cr.SetSourceRGB(255, 0, 0)
			cr.Arc(Xc, Yc, R, 0.0, om)
			cr.Fill()

		} else {
			cr.SetSourceRGB(0, 0, 127)
			cr.MoveTo(Xc, Yc)
			cr.Arc(Xc, Yc, R, dlt, dlt+om*float64(c.TickStep)/(float64(c.TickCount)))
			cr.Fill()

			if (c.TickStep * 2) == c.TickCount {
				cr.SetSourceRGB(0, 0, 0)
				cr.SetLineWidth(3)
				cr.MoveTo(Xc, Yc-R)
				cr.LineTo(Xc, Yc+R)
				cr.Stroke()
			}
		}
	}
	cr.Fill()

	{ // Beat display
		cr.SetSourceRGB(0, 0, 0)
		cr.SetLineWidth(3)
		cr.Arc(Xc, Yc, float64(R), 0.0, om)
		cr.Stroke()
		cr.SetLineWidth(1)
		cr.Arc(Xc, Yc, float64(R/3), 0.0, om)
		cr.Stroke()
		cr.SetSourceRGB(255, 255, 255)
		cr.Arc(Xc, Yc, float64(R/3), 0.0, om)
		cr.Fill()
		cr.SetSourceRGB(0, 0, 0)
		cr.SetFontSize(R / 4)
		txt := fmt.Sprint(c.TickStep + 1)
		ext := cr.TextExtents(txt)
		cr.MoveTo(Xc-ext.Width/2, Yc+ext.Height/2)
		cr.ShowText(txt)
		cr.Fill()
	}

	{ // Measure display
		cr.SetFontSize(R / 4)
		txt := fmt.Sprint(c.MeasureId)
		ext := cr.TextExtents(txt)
		cr.MoveTo(X-ext.Width-3, ext.Height)
		cr.ShowText(txt)
		cr.Fill()
	}
	// Pass display
	if c.PassCount > 0 {
		cr.SetFontSize(R / 4)
		txt := fmt.Sprint(c.PassCount)
		ext := cr.TextExtents(txt)
		cr.MoveTo(3, ext.Height)
		cr.ShowText(txt)
		cr.Fill()
	}

}

func (ctx *ZContext) MkMetro() (*WMetronome, gtk.IWidget) {
	c := &WMetronome{}
	c.Metronome, _ = gtk.DrawingAreaNew()
	c.Metronome.SetHAlign(gtk.ALIGN_FILL)
	c.Metronome.SetVAlign(gtk.ALIGN_FILL)
	c.Metronome.SetHExpand(true)
	c.Metronome.SetVExpand(true)
	c.Metronome.SetSizeRequest(100, 100)
	c.Metronome.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
		c.Draw(da, cr)
	})

	c.pendingEvent = make([]TickEvent, 0)
	c.Metronome.AddTickCallback(func(widget *gtk.Widget, frameClock *gdk.FrameClock) bool {
		c.Ticker()
		return true
	})

	return c, c.Metronome
}
