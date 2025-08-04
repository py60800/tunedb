package zique

import (
	"fmt"
	"sync"
)

type MStart struct {
	MeasureId int
}
type MTimeEv struct {
	MLength  int
	Beats    int
	BeatType int
}

func (m MTimeEv) String() string {
	return fmt.Sprintf("TimeEvt: %v %v/%v", m.MLength, m.Beats, m.BeatType)
}

type PNoteOn struct {
	NoteNumber int
	Velocity   int
	Channel    int
}

type PNoteOff struct {
	NoteNumber int
	Channel    int
}
type PTickOn struct {
	NoteNumber int
	Velocity   int
}
type PTickOff struct {
	NoteNumber int
}
type PChordOn struct {
	p        *Player
	chordOff *PChordOff
	Velocity int
}
type PProgramChange struct {
	Patch int
}
type PTempoChange struct {
	Value int
}

func (t PTempoChange) String() string {
	return fmt.Sprintf("Tempo:", t.Value)
}
func (c PChordOn) String() string {
	return fmt.Sprintf("ChordOn (%v)", c.Velocity)
}

type PChordOff struct {
	Notes []int
}

func (c *PChordOff) String() string {
	return fmt.Sprintf("ChordOff")
}

type PEvent interface {
	GetRtMidiEvent() []byte
	String() string
}

type SeqEvent struct {
	Tick  uint32
	Event PEvent
}
type SeqChord struct {
	Tick  uint32
	Chord EChord
}
type EChord struct {
	Key  int
	Mode string
}

func (p *Player) KeyToInt(note string, octave int, alter int) int {
	n := MidiKeyInt(note, octave)
	n += alter
	return n
}

var DefaultChannel int = 0
var ChordChannel int = 2
var DrumChannel int = 9

func (n PNoteOn) String() string {
	return fmt.Sprintf("NoteOn %v %v", n.NoteNumber, n.Velocity)
}
func (n PTickOn) String() string {
	return fmt.Sprintf("--TickOn %v %v", n.NoteNumber, n.Velocity)
}
func (n PNoteOff) String() string {
	return fmt.Sprintf("NoteOff %v", n.NoteNumber)
}
func (n PTickOff) String() string {
	return fmt.Sprintf("--TickOff %v", n.NoteNumber)
}

func (p PProgramChange) String() string {
	return fmt.Sprintf("PChange: %d", p.Patch)
}
func (m MStart) String() string {
	return fmt.Sprintf("MeasureStart")
}

type DummyEvent struct {
}

func (p DummyEvent) GetRtMidiEvent() []byte {
	return []byte{}
}
func (p DummyEvent) String() string {
	return "Dummy Event"
}

func MidiRecord(pl *Player, mchan chan SeqEvent, ctrlChan chan int, wg *sync.WaitGroup) {
	for {
		select {
		case evt := <-mchan:
			pl.Sequence = append(pl.Sequence, evt)
			//			fmt.Println("Midi:", evt)

		case cmd := <-ctrlChan:
			switch cmd {
			case 0:
				fmt.Println("Midi: Finish")
				wg.Done()
				return
			case 1:
			}
		}

	}

}

// ****************************************************
