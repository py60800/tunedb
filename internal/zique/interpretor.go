// interpretor
package zique

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/iterator"
)

const DefaultVelocity int = 100

var PlayRoll = false

type CPlayCtrl struct {
	Cmd   int
	Param string
	Coeff float64
}

type Tick struct {
	Beats        int
	BeatType     int
	XmlDivisions int
	TickTime     time.Duration
	MeasureId    int
	PassCount    int
}

const (
	CSTOP = iota
	CDRUM
	CVELOCITY
	CCHORD
	CSWING
)

type NTrail struct {
	note   int
	tickOn uint32
	tickOf uint32
}
type Player struct {
	Name           string
	XmlDivisions   int
	MeasLength     int
	MeasLengthTick int
	KeyAlter       map[byte]int

	TuneMeasureLength int

	Beats    int
	BeatType int

	Clock       uint32
	MeasStart   uint32
	MNoteLength []NTrail

	TuneStart int

	Velocity []int

	Sequence []SeqEvent

	AuxSequence []SeqEvent

	CurrentChord EChord

	MeasSeq     []uint32
	SwingFactor []int

	SwingPattern    []float64
	VelocityPattern []float64
	DrumPattern     []Beat

	PlayCtrl chan CPlayCtrl
}

func (p Player) String() string {
	return fmt.Sprintf("(%v, %v, %v, %v)", p.Name, p.XmlDivisions, p.Clock, len(p.Sequence))
}
func (p *Player) MeasureLength() int {
	t := (p.XmlDivisions * 4) / p.BeatType
	return p.Beats * t
}
func (p *Player) ComputeSwingFactor() {
	sw := make([]int, p.MeasLength)
	if len(p.SwingPattern) == 0 || p.MeasLength == 0 || p.MeasLength != p.TuneMeasureLength {
		return
	}
	segLength := p.MeasLength / len(p.SwingPattern)
	for i := range sw {
		s := i / segLength
		a0 := p.SwingPattern[s]
		a1 := p.SwingPattern[(s+1)%len(p.SwingPattern)]

		sw[i] = int(float64(MasterDivisions) * (a0 + (a1-a0)*float64(i%segLength)/float64(segLength)))
	}
	p.SwingFactor = sw
}

func (p *Player) AddSwing(s SeqEvent) SeqEvent {
	if len(p.SwingFactor) > 0 {
		nTick := uint32(int(s.Tick) +
			p.SwingFactor[(int(s.Tick-p.MeasStart))%len(p.SwingFactor)])
		if nTick < p.MeasStart+uint32(p.MeasLength) {
			s.Tick = nTick
		}
	}
	return s
}
func (p *Player) PurgeAuxSequence() {
	p.PushEvent(SeqEvent{p.Clock, DummyEvent{}})
	chordPurged := false
	drumPurged := false
	for _, auxEvent := range p.AuxSequence {
		switch v := auxEvent.Event.(type) {
		case *PChordOff:
			for _, n := range v.Notes {
				RtMidiChan <- p.AddSwing(SeqEvent{auxEvent.Tick, PNoteOff{n, ChordChannel}})
			}
			chordPurged = true
		case PTickOff:
			RtMidiChan <- p.AddSwing(auxEvent)
			drumPurged = true
		}
		if chordPurged && drumPurged {
			break
		}
	}
	p.AuxSequence = p.AuxSequence[:0]
}
func (p *Player) PushEvent(s SeqEvent) {
	i := 0
	for i < len(p.AuxSequence) && p.AuxSequence[i].Tick <= s.Tick {
		auxEvent := p.AuxSequence[i]
		// Process all previous events
		switch v := auxEvent.Event.(type) {
		case PChordOn:
			relatedChordOff := v.chordOff
			if cp, ok := ChordTemplate[p.CurrentChord.Mode]; ok {
				for _, note := range cp {
					mn := p.CurrentChord.Key + note
					Velocity := (ChordVelocity * v.Velocity) / 100
					RtMidiChan <- p.AddSwing(SeqEvent{auxEvent.Tick, PNoteOn{mn, Velocity, ChordChannel}})
					relatedChordOff.Notes = append(relatedChordOff.Notes, mn)
				}
			}
		case *PChordOff:
			for _, n := range v.Notes {
				RtMidiChan <- p.AddSwing(SeqEvent{auxEvent.Tick, PNoteOff{n, ChordChannel}})
			}
		default:
			RtMidiChan <- p.AddSwing(p.AuxSequence[i])
		}
		i++

	}
	s = p.AddSwing(s)

	if i > 0 {
		p.AuxSequence = p.AuxSequence[i:]
	}
	switch v := s.Event.(type) {
	case PNoteOn:
		p.MNoteLength = append(p.MNoteLength, NTrail{note: v.NoteNumber, tickOn: s.Tick})
	case PNoteOff:
		if len(p.MNoteLength) > 0 {
			p.MNoteLength[len(p.MNoteLength)-1].tickOf = s.Tick
		}
	}

	RtMidiChan <- s
}

func (pm *MPart) Attributes() (int, int, int) {
	var div int
	var beat int
	var beatType int

	iter := pm.CreateIterator()
	for {
		el, err := iter.Next()
		if err != nil {
			break
		}
		switch v := el.Elem.(type) {
		case MAttributes:
			for _, mx := range v.Contents {
				switch ve := mx.Elem.(type) {
				case MDivision:
					div = ve.Value
				case MTime:
					beat, beatType = ve.Beats, ve.BeatType
				}
			}

		}
	}
	return div, beat, beatType
}
func MakePlayer(name string) Player {
	var p Player
	p.Name = name
	p.Sequence = make([]SeqEvent, 0, 10000)
	p.AuxSequence = make([]SeqEvent, 0, 100)
	p.Velocity = make([]int, 0)

	p.MeasSeq = make([]uint32, 0)

	p.Clock = 0
	p.MeasStart = p.Clock
	p.MNoteLength = make([]NTrail, 0, 20)
	p.DrumPattern = make([]Beat, 0)

	p.PlayCtrl = make(chan CPlayCtrl, 2)
	return p
}

func (p *Player) PartInit(pm *MPart) {
	p.XmlDivisions, p.Beats, p.BeatType = pm.Attributes()
}

func (p *Player) XmlDuration2MidiDuration(d int) uint32 {
	return uint32(d * MasterDivisions / p.XmlDivisions)
}

// **************************************************************

type PIterator struct {
	mp   *MPart
	i, j int
}

func (mp *MPart) CreateIterator() PIterator {
	return PIterator{mp, 0, 0}
}

func (p *PIterator) Done() error {
	if p.i >= len(p.mp.Measures) {
		return iterator.Done
	}
	return nil
}
func (p *PIterator) IsStartMeasure() bool {
	return p.j == 0 && p.i < len(p.mp.Measures)
}
func (p *PIterator) MeasureIndex() int {
	return p.i
}
func (p *PIterator) CurrentMeasure() *MMeasure {
	return &p.mp.Measures[p.i]
}

func (p *PIterator) Next() (*Mixed, error) {
	if p.i < len(p.mp.Measures) {
		r := &p.mp.Measures[p.i].Contents[p.j]
		p.j++
		if p.j >= len(p.mp.Measures[p.i].Contents) {
			p.j = 0
			p.i++
		}

		return r, nil
	}
	return nil, iterator.Done
}

// **************************************************

type MElem interface {
	Process(*Player)
}

func (p *Player) GetVelocity(iMes int) int {
	v := DefaultVelocity
	if iMes < len(p.Velocity) {
		v = p.Velocity[iMes]
	}
	v = (v * Velocity) / 100
	if v > 127 {
		v = 127
	}
	return v
}
func (p *Player) ComputeVelocity() {
	p.Velocity = make([]int, p.MeasLength)
	if len(p.VelocityPattern) == 0 {
		for i := range p.Velocity {
			p.Velocity[i] = DefaultVelocity
		}
	} else {
		segLength := p.MeasLength / (len(p.VelocityPattern) - 1)
		for i := range p.Velocity {
			s := i / segLength
			a0 := p.VelocityPattern[s]
			a1 := p.VelocityPattern[(s + 1)]

			p.Velocity[i] = int(float64(Velocity) * (a0 + (a1-a0)*float64(i%segLength)/float64(segLength)))
		}
	}

}
func (h MHarmony) Process(p *Player) {
	p.CurrentChord = EChord{p.KeyToInt(h.Root.RootStep, 2, h.Root.RootAlter), h.Kind}
}

func (n MNote) Process(p *Player) {
	if !n.IsRest() && !MelodyOff {
		doRoll := false
		if PlayRoll && n.Notations.StrongAccent.Local == "strong-accent" {
			rollPattern, ok := RollPattern[RPattern{n.Step, n.Alter, n.Type, n.Dot.Local == "dot"}]
			if ok {
				doRoll = true
				t := 0
				for _, rn := range rollPattern {
					octave := n.Octave - 1
					step := rn.Note
					if strings.HasPrefix(step, "^") {
						octave++
						step = strings.TrimPrefix(step, "^")

					}
					k := MidiKeyInt(step, octave)
					ev := PEvent(PNoteOn{k, (p.GetVelocity(int(p.Clock-p.MeasStart)) * rn.RelVelocity) / 100, DefaultChannel})
					p.PushEvent(SeqEvent{p.Clock, ev})
					p.Clock += uint32(rn.Duration)
					p.PushEvent(SeqEvent{p.Clock - 1, PNoteOff{k, DefaultChannel}})
					t += rn.Duration
				}
				if t != int(n.Duration) {
					panic(fmt.Sprintf("Invalid Roll length %v %v", t, n.Duration))
				}

			}
		}
		if !doRoll {
			if n.IsGraceNote() {
				// Clock not changed
			} else {
				mn := p.KeyToInt(n.Step, n.Octave-1, n.Alter)
				ev := PEvent(PNoteOn{mn, p.GetVelocity(int(p.Clock - p.MeasStart)), DefaultChannel})
				p.PushEvent(SeqEvent{p.Clock, ev})
				p.Clock += p.XmlDuration2MidiDuration(n.Duration)
				p.PushEvent(SeqEvent{p.Clock - 1, PNoteOff{mn, DefaultChannel}})
			}
		}

	} else {
		p.Clock += p.XmlDuration2MidiDuration(n.Duration)

	}

}
func (a MAttributes) Process(p *Player) {
	for _, t := range a.Contents {
		if t.Elem != nil {
			t.Elem.Process(p)
		}
	}
}
func (b MBarline) Process(p *Player) {}
func (d MDivision) Process(p *Player) {
	p.XmlDivisions = d.Value
}
func (d MTime) Process(p *Player) {
	p.Beats = d.Beats
	p.BeatType = d.BeatType
	p.MeasLength = d.Beats * ((p.XmlDivisions * 4) / d.BeatType)
	p.MeasLengthTick = (p.Beats * MasterDivisions * 4) / p.BeatType
	if p.MeasLength != p.MeasLengthTick {
		panic("Normlize raté")
	}
	p.ComputeVelocity()
	p.ComputeSwingFactor()

	p.PushEvent(SeqEvent{p.Clock, MTimeEv{p.MeasLength, p.Beats, p.BeatType}})
}

func (k MKey) Process(p *Player) {
	//	Nothing to do
}

func (p *Player) ActualMeasureLength(m *MMeasure) int {
	MeasLength := 0
	for _, el := range m.Contents {
		switch v := el.Elem.(type) {
		case MNote:
			MeasLength += int((v.Duration * MasterDivisions) / p.XmlDivisions)
		}
	}
	return MeasLength

}
func (p *Player) processCmd(C CPlayCtrl) bool {
	switch C.Cmd {
	case CSTOP:
		return false
	case CDRUM:
		p.DrumPattern = GetDrumPattern(C.Param)
	case CSWING:
		p.SwingPattern = GetSwingPattern(C.Param)
		for i := range p.SwingPattern {
			p.SwingPattern[i] = p.SwingPattern[i] * C.Coeff
		}
		p.ComputeSwingFactor()
	case CVELOCITY:
		p.VelocityPattern = GetVelocityPattern(C.Param)
		// Get min max
		for i, v := range p.VelocityPattern {
			p.VelocityPattern[i] = math.Max(0.0, 1.0-(1.0-v)*C.Coeff)
		}
		p.ComputeVelocity()
	default:
		fmt.Println("Unimplemented:", C)
	}
	return true
}

func (p *Player) ComputeTuneMeasureLength(fp MPart) int {
	ml := make(map[int]int)
	for i := 0; i < 8 && i < len(fp.Measures); i++ {
		mLength := p.ActualMeasureLength(&fp.Measures[i])
		if n, ok := ml[mLength]; ok {
			ml[mLength] = n + 1
		} else {
			ml[mLength] = 1
		}
	}
	mx := 0
	mlx := 0
	for k, n := range ml {
		if n > mx {
			mx = n
			mlx = k
		}
	}

	return mlx
}

func (p *Player) PlayTune(fp MPart, barSignal int, callBack func()) int {
	if p.Clock == 0 {
		p.PartInit(&fp)
		p.TuneMeasureLength = p.ComputeTuneMeasureLength(fp)

	} else {
		p.TuneMeasureLength = p.ComputeTuneMeasureLength(fp)
		lastMeasDuration := int(p.Clock - p.MeasStart)

		p.PartInit(&fp)
		FirstMeasLength := p.ActualMeasureLength(&(fp.Measures[0]))
		MeasureLength := p.ActualMeasureLength(&(fp.Measures[1]))
		FirstMeasLength %= MeasureLength

		switch {
		case lastMeasDuration+FirstMeasLength == MeasureLength:
			fmt.Printf("Raccord Type 1!! lm:%v, fm:%v, ml:%v\n", lastMeasDuration, FirstMeasLength, MeasureLength)
		case lastMeasDuration == FirstMeasLength && FirstMeasLength == MeasureLength:
			fmt.Printf("Raccord Type 2!! lm:%v, fm:%v, ml:%v\n", lastMeasDuration, FirstMeasLength, MeasureLength)
		case lastMeasDuration+FirstMeasLength < MeasureLength:
			// Transient Measure requires completion
			delta := MeasureLength - (lastMeasDuration + FirstMeasLength)
			p.Clock += uint32(delta)
		case lastMeasDuration+FirstMeasLength > MeasureLength:
			delta := (lastMeasDuration + FirstMeasLength) % MeasureLength
			p.Clock -= uint32(delta)
			fmt.Printf("Raccord Zarbi!! LastM:%v, FirstM:%v, MLength:%v (%v)\n", lastMeasDuration, FirstMeasLength, MeasureLength, delta)
		default:
			fmt.Println("Unexpected Measure config\n", lastMeasDuration, FirstMeasLength, MeasureLength)

		}

	}
	barCount := 0
	PassNumber := 1
	iter := fp.CreateIterator()
	RestartPoint := iter
	Replay := false
loop:
	for {
		select {
		case cmd := <-p.PlayCtrl:
			if !p.processCmd(cmd) {
				return -1
			}
		default:
		}
		if iter.IsStartMeasure() || Replay {
			{
				p.MNoteLength = make([]NTrail, 0, 10)
			}
			p.PurgeAuxSequence()
			actualMeasureLength := p.ActualMeasureLength(iter.CurrentMeasure())
			p.MeasLength = actualMeasureLength
			p.ComputeVelocity()
			p.ComputeSwingFactor()
			p.AddChord(p.Clock, uint32(actualMeasureLength))
			p.AddDrum(p.Clock, uint32(actualMeasureLength))

			sort.Slice(p.AuxSequence, func(i, j int) bool { return p.AuxSequence[i].Tick < p.AuxSequence[j].Tick })

			p.MeasSeq = append(p.MeasSeq, p.Clock)
			p.MeasStart = p.Clock
			p.PushEvent(SeqEvent{p.Clock, MStart{MeasureId: iter.CurrentMeasure().Id}})
			Replay = false
		}

		el, err := iter.Next()
		if err != nil {
			break
		} // end of file

		switch v := el.Elem.(type) {
		case MBarline:
			if barCount == barSignal {
				callBack()
			}
			barCount++
			switch {
			case v.Repeat.Direction == "backward":
				if PassNumber == 1 {
					Replay = true
					iter = RestartPoint
					PassNumber++
					continue loop
				}
			case v.Repeat.Direction == "forward":
				RestartPoint = iter
				PassNumber = 1
			case v.Ending.Number == PassNumber && v.Ending.Type == "start":
				// First pass continue
			case v.Ending.Number != PassNumber && v.Ending.Type == "start":
				//fmt.Println("Second Pass")
				p.AuxSequence = p.AuxSequence[:0] // a voir
				//Skip until stop bar
				for { // Search for Barline
					el, err := iter.Next()

					if err != nil {
						fmt.Println("No second pass")
						break loop
					}
					switch v := el.Elem.(type) {
					case MBarline:
						if v.Ending.Type == "stop" {
							continue loop
						}
					}
				}
				fmt.Println("end not found")
			}
		default:
			//fmt.Println("Process ", el)
			if el.Elem != nil {
				el.Elem.Process(p)
			}
		}

	}
	return barCount
}

// ********************************************

func (p *Player) AddChord(Clock uint32, actualMeasureLength uint32) {
	if len(p.MeasSeq) == 0 || ChordVelocity == 0 || len(ChordPattern) == 0 {
		return
	}
	end := Clock + actualMeasureLength

	for _, ch := range ChordPattern {
		Duration := uint32(float64(MasterDivisions) * ch.Duration)
		if Clock+Duration >= end {
			return
		}
		Velocity := (ch.Velocity * ChordVelocity) / 100
		chordOff := PChordOff{make([]int, 0)}
		last := len(p.AuxSequence)
		p.AuxSequence = append(p.AuxSequence, SeqEvent{Clock + Duration - 1, &chordOff})
		aChordOff, _ := (p.AuxSequence[last].Event.(*PChordOff))
		chordOn := PChordOn{p, aChordOff, Velocity}
		p.AuxSequence = append(p.AuxSequence, SeqEvent{Clock, chordOn})
		Clock += Duration
	}
}
func (p *Player) AddDrum(Clock uint32, actualMeasureLength uint32) {
	if len(p.MeasSeq) == 0 || DrumVelocity == 0 || len(p.DrumPattern) == 0 {
		return
	}
	//	sClock := Clock
	end := Clock + actualMeasureLength
	for _, drm := range p.DrumPattern {
		Duration := uint32(drm.Duration*float64(MasterDivisions)) - 1
		if Clock+Duration > end {
			break
		}
		if drm.Instrument != 0 {
			Velocity := (drm.Velocity * DrumVelocity) / 100
			p.AuxSequence = append(p.AuxSequence, SeqEvent{Clock, PTickOn{drm.Instrument, Velocity}})
			p.AuxSequence = append(p.AuxSequence, SeqEvent{Clock + Duration, PTickOff{drm.Instrument}})
		}
		Clock += Duration + 1
	}
}

func (p *MPartition) NormalizeDivisions(MasterDivisions int) {
	for ip, part := range p.Part {
		coeff := 0

		for im, measure := range part.Measures {
			for ic, item := range measure.Contents {
				switch v := item.Elem.(type) {
				case MAttributes:
					for ia, a := range v.Contents {
						switch d := a.Elem.(type) {
						case MDivision:
							nd := d.Value
							if nd > MasterDivisions {
								panic("XML Divisions too large")
							}
							coeff = MasterDivisions / nd
							if coeff*nd != MasterDivisions {
								panic("Odd Divisions")
							}
							d.Value = MasterDivisions
							p.Part[ip].Measures[im].Contents[ic].Elem.(MAttributes).Contents[ia].Elem = d
						}
					}
				case MNote:
					if coeff == 0 {
						panic("Division unset")
					}
					v.Duration *= coeff
					p.Part[ip].Measures[im].Contents[ic].Elem = v
				}
			}

		}
	}

}
