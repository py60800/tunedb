// interpretor
package zique

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/py60800/tunedb/internal/util"
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
	Beats             int
	BeatType          int
	XmlDivisions      int
	TickTime          time.Duration
	MeasureLength     int
	MeasureLengthTune int
	MeasureId         string
	PassCount         int
}

const (
	CDRUM = iota
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
	locker      sync.Mutex
	stopRequest atomic.Bool

	Name         string
	XmlDivisions int

	MeasLength         int
	PreviousMeasLength int
	MeasLengthTick     int
	KeyAlter           map[byte]int

	TuneMeasureLength int

	Beats    int
	BeatType int

	Clock       uint32
	MeasStart   uint32
	MNoteLength []NTrail

	TuneStart int

	Velocity []int

	Sequence     []SeqEvent
	AuxSequence  []SeqEvent
	CurrentChord EChord

	MeasSeq     []uint32
	SwingFactor []int

	SwingPattern    []float64
	VelocityPattern []float64
	DrumPattern     []Beat

	PlayCtrl chan CPlayCtrl

	passCount int
}

func (p Player) String() string {
	return fmt.Sprintf("(%v, %v, %v, %v)", p.Name, p.XmlDivisions, p.Clock, len(p.Sequence))
}
func (p *Player) computeSwingFactor() {
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

func (p *Player) addSwing(s SeqEvent) SeqEvent {
	if len(p.SwingFactor) > 0 {
		nTick := uint32(int(s.Tick) +
			p.SwingFactor[(int(s.Tick-p.MeasStart))%len(p.SwingFactor)])
		if nTick < p.MeasStart+uint32(p.MeasLength) {
			s.Tick = nTick
		}
	}
	return s
}
func (p *Player) purgeAuxSequence() {
	p.pushEvent(SeqEvent{p.Clock, DummyEvent{}})
	chordPurged := false
	drumPurged := false
	for _, auxEvent := range p.AuxSequence {
		switch v := auxEvent.Event.(type) {
		case *PChordOff:
			for _, n := range v.Notes {
				RtMidiChan <- p.addSwing(SeqEvent{auxEvent.Tick, PNoteOff{n, ChordChannel}})
			}
			chordPurged = true
		case PTickOff:
			RtMidiChan <- p.addSwing(auxEvent)
			drumPurged = true
		}
		if chordPurged && drumPurged {
			break
		}
	}
	p.AuxSequence = p.AuxSequence[:0]
}
func (p *Player) pushEvent(s SeqEvent) {
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
					RtMidiChan <- p.addSwing(SeqEvent{auxEvent.Tick, PNoteOn{mn, Velocity, ChordChannel}})
					relatedChordOff.Notes = append(relatedChordOff.Notes, mn)
				}
			}
		case *PChordOff:
			for _, n := range v.Notes {
				RtMidiChan <- p.addSwing(SeqEvent{auxEvent.Tick, PNoteOff{n, ChordChannel}})
			}
		default:
			RtMidiChan <- p.addSwing(p.AuxSequence[i])
		}
		i++

	}
	s = p.addSwing(s)

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
func MakePlayer(name string) *Player {
	p := &Player{}
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

func (p *Player) xmlDuration2MidiDuration(d int) uint32 {
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
func (p *PIterator) IsMeasureStart() bool {
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
	v = min(127, v*Velocity/100)
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
		//	fmt.Println("Tied:", n.Notations.Tied.Type)
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
					p.pushEvent(SeqEvent{p.Clock, ev})
					p.Clock += uint32(rn.Duration)
					p.pushEvent(SeqEvent{p.Clock - 1, PNoteOff{k, DefaultChannel}})
					t += rn.Duration
				}
				if t != int(n.Duration) {
					err := fmt.Errorf("Internal error =>Invalid Roll length %v %v", t, n.Duration)
					util.PanicOnError(err)
				}

			}
		}
		if !doRoll {
			if n.IsGraceNote() {
				// Clock not changed
			} else {
				mn := p.KeyToInt(n.Step, n.Octave-1, n.Alter)
				if n.Notations.Tied.Type != "stop" {
					ev := PEvent(PNoteOn{mn, p.GetVelocity(int(p.Clock - p.MeasStart)), DefaultChannel})
					p.pushEvent(SeqEvent{p.Clock, ev})
				}
				p.Clock += p.xmlDuration2MidiDuration(n.Duration)
				if n.Notations.Tied.Type != "start" {
					p.pushEvent(SeqEvent{p.Clock - 1, PNoteOff{mn, DefaultChannel}})
				}
			}
		}

	} else {
		p.Clock += p.xmlDuration2MidiDuration(n.Duration)

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
		panic("Internal Error => Failed to Normalize tempo")
	}
	p.ComputeVelocity()
	p.computeSwingFactor()

	p.pushEvent(SeqEvent{p.Clock, MTimeEv{p.MeasLength, p.Beats, p.BeatType}})
}

func (k MKey) Process(p *Player) {
	//	Nothing to do
}
func (p *Player) processCmd(C CPlayCtrl) {
	switch C.Cmd {
	case CDRUM:
		p.DrumPattern = GetDrumPattern(C.Param)
	case CSWING:
		p.SwingPattern = GetSwingPattern(C.Param)
		for i := range p.SwingPattern {
			p.SwingPattern[i] = p.SwingPattern[i] * C.Coeff
		}
		p.computeSwingFactor()
	case CVELOCITY:
		p.VelocityPattern = GetVelocityPattern(C.Param)
		// Get min max
		for i, v := range p.VelocityPattern {
			p.VelocityPattern[i] = math.Max(0.0, 1.0-(1.0-v)*C.Coeff)
		}
		p.ComputeVelocity()
	default:
		log.Println("Unimplemented:", C)
	}
}

func (p *Player) ComputeTuneMeasureLength(fp MPart) int {
	mx := 0
	for i := 0; i < 4 && i < len(fp.Measures); i++ {
		mx = max(mx, fp.Measures[i].length)
	}
	return mx
}
func (p *Player) stopRequested() bool {
	return p.stopRequest.Load()
}
func (p *Player) Stop() {
	p.stopRequest.Store(true)
}
func (p *Player) playTune(fp MPart, barSignal int, callBack func()) int {
	log.Println("MidiPlay:", len(fp.Measures))
	p.stopRequest.Store(false)

	barCount := 0
	PassNumber := 1
	iter := fp.CreateIterator()

	if p.Clock == 0 {
		// First tune or first pass
		p.PartInit(&fp)
		p.TuneMeasureLength = p.ComputeTuneMeasureLength(fp)

	} else {

		// Process tune junction !

		p.PreviousMeasLength = p.MeasLength
		lastMeasureIsClean := p.TuneMeasureLength == p.PreviousMeasLength
		missingDuration := p.TuneMeasureLength - p.PreviousMeasLength

		//
		p.TuneMeasureLength = p.ComputeTuneMeasureLength(fp)
		p.PartInit(&fp)

		FirstMeasLength := fp.Measures[0].length
		cleanStart := FirstMeasLength == p.TuneMeasureLength

		if lastMeasureIsClean {
			if cleanStart {
				// Perfect match => nothing to do
				log.Println("Midi: clean junction")
			} else {
				// Ignore pickup
				log.Println("Midi: Skip pickup")
				for {
					_, err := iter.Next()
					if err != nil {
						return -1
					}
					if iter.IsMeasureStart() {
						break
					}
				}
			}
		} else {
			if cleanStart {
				// Last Measure was incomplete => Add Silence
				log.Println("Midi: Incomplete Meas && CleanStart")
				p.Clock += uint32(missingDuration)
			} else {
				combinedDuration := p.PreviousMeasLength + FirstMeasLength
				switch {
				case combinedDuration == p.TuneMeasureLength:
					log.Println("Midi: Perfect junction")
				case combinedDuration < p.TuneMeasureLength:
					log.Println("Midi: Add missing rest:", p.TuneMeasureLength-combinedDuration)
					p.Clock += uint32(p.TuneMeasureLength - combinedDuration)
				default: // combined junction too long
					log.Println("Midi: Drop pickup && add rest")
					for {
						_, err := iter.Next()
						if err != nil {
							return -1
						}
						if iter.IsMeasureStart() {
							break
						}
						p.Clock += uint32(missingDuration)
					}

				}
			}
		}
	}
	RestartPoint := iter
	Replay := false

loop:
	for !p.stopRequested() {
		// Process pending commands
		select {
		case cmd := <-p.PlayCtrl:
			p.processCmd(cmd)
		default:
		}

		if iter.IsMeasureStart() || Replay {

			p.MNoteLength = make([]NTrail, 0, 10)
			p.purgeAuxSequence()
			p.MeasLength = iter.CurrentMeasure().length

			p.ComputeVelocity()
			p.computeSwingFactor()
			p.AddChord(p.Clock, uint32(p.MeasLength))
			p.AddDrum(p.Clock, uint32(p.MeasLength))

			sort.Slice(p.AuxSequence, func(i, j int) bool { return p.AuxSequence[i].Tick < p.AuxSequence[j].Tick })

			p.MeasSeq = append(p.MeasSeq, p.Clock)
			p.MeasStart = p.Clock
			p.pushEvent(SeqEvent{p.Clock, MStart{
				MeasureId:         iter.CurrentMeasure().Id,
				MeasureLength:     iter.CurrentMeasure().length,
				MeasureLengthTune: p.TuneMeasureLength,
			}})

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
						break loop
					}
					switch v := el.Elem.(type) {
					case MBarline:
						if v.Ending.Type == "stop" {
							continue loop
						}
					}
				}
				log.Println("end not found")
			}
		default:

			// Process element => main job
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
			mlength := 0
			for ic, item := range measure.Contents {
				switch v := item.Elem.(type) {
				case MAttributes:
					for ia, a := range v.Contents {
						switch d := a.Elem.(type) {
						case MDivision:
							nd := d.Value
							if nd > MasterDivisions {
								panic("Major error XML Divisions too large")
							}
							coeff = MasterDivisions / nd
							if coeff*nd != MasterDivisions {
								panic("Major error Divisions")
							}
							d.Value = MasterDivisions
							p.Part[ip].Measures[im].Contents[ic].Elem.(MAttributes).Contents[ia].Elem = d
						}
					}
				case MNote:
					if coeff == 0 {
						panic("Missing division parameter")
					}
					v.Duration *= coeff
					p.Part[ip].Measures[im].Contents[ic].Elem = v
					mlength += v.Duration
				}
			}
			part.Measures[im].length = mlength

		}
	}

}
func (p *Player) PlaySet(set MusicSet, FeedBack chan string) {
	defer func() {
		log.Println("Midi-Exit playing")
	}()
	p.locker.Lock()
	p.stopRequest.Store(false)
	defer p.locker.Unlock()

	log.Println("Midi-Play:", set)
	for iset, t := range set {
		if p.stopRequested() {
			return
		}

		log.Println("Midi-Parse:", t.File)

		select { // Update the displayed score
		case FeedBack <- t.File:
		default:
		}

		partition, err := Parse(t.File)
		log.Println("Midi-Parse:", err, len(partition.Part))
		if err != nil || len(partition.Part) == 0 || len(partition.Part[0].Measures) < 3 {
			log.Println("Invalid Tune:", t.File, err)
			break
		}
		count := t.Count
		p.passCount = 1
		barCount := -1
		signalBar := -1
		for !p.stopRequested() {
			// Play the tune
			// As much as possible the display must be updated before the end
			barCount = p.playTune(partition.Part[0], signalBar, func() {
				if iset != len(set)-1 {
					select {
					case FeedBack <- set[iset+1].File:
					default:
					}
				}
			})
			p.passCount++
			if count > 0 { // count == 0 => infinite
				count--
				if count == 1 && barCount > 0 {
					signalBar = barCount - 2
					barCount = -1
				}
				if count == 0 {
					break
				}
			}
		}
	}

}
