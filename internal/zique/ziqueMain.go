package zique

import (
	"log"
	"sync/atomic"
)

var (
	RtMidiChan      chan SeqEvent
	RtCtrlChan      chan int
	RtMidiSinkChan  chan []byte
	Tempo           = 120
	MainInstrument  = 0
	Velocity        = 100
	ChordInstrument = 0
	ChordVelocity   = 0
	DrumVelocity    = 0
	MelodyOff       = false
	ChordPattern    []ChordStroke
)

const MasterDivisions = 1200

type SetElem struct {
	File  string
	Count int
	Tempo int
}
type MusicSet []SetElem

type ZiquePlayer struct {
	ZiqueCtrl chan interface{}

	DrumPattern     string
	VelocityPattern string
	SwingPattern    string

	midiPlayer *Player

	FeedBack   chan string
	TickBack   chan Tick
	RtFeedBack chan RtMeasureTick

	playing   atomic.Bool
	passCount int
}

func ZiquePlayerNew(context string, midiPort string) (*ZiquePlayer, string) {
	z := ZiquePlayer{
		ZiqueCtrl: make(chan interface{}, 10),
		FeedBack:  make(chan string, 2),
		TickBack:  make(chan Tick, 2),
	}
	z.playing.Store(false)
	msg := z.init(context, midiPort)
	return &z, msg
}

func (z *ZiquePlayer) init(context string, midiPort string) string {
	if midiPort == "" {
		midiPort = "Synth"
	}

	RtMidiChan = make(chan SeqEvent, 3)
	RtCtrlChan = make(chan int, 3)
	RtStartMsg := make(chan string)

	z.RtFeedBack = make(chan RtMeasureTick, 2)

	go MidiSink(RtMidiChan, midiPort, RtStartMsg, RtCtrlChan, z.RtFeedBack)

	go z.mainLoop(z.ZiqueCtrl)
	msg := <-RtStartMsg
	return msg
}
func (z *ZiquePlayer) IsPlaying() bool {
	return z.playing.Load()
}
func (z *ZiquePlayer) RtSend(evt PEvent) {
	RtMidiSinkChan <- evt.GetRtMidiEvent()
}
func (z *ZiquePlayer) SetTempo(tempo int) {
	RtMidiChan <- SeqEvent{0, PTempoChange{tempo}}
}
func (z *ZiquePlayer) SetPatch(patch string) {
	p, ok := RPatch[patch]
	if ok {
		RtMidiChan <- SeqEvent{0, PProgramChange{p - 1}}
	}

}
func (z *ZiquePlayer) SetMainVolume(v int) {
	Velocity = v
}
func (z *ZiquePlayer) SetDrumVolume(v int) {
	DrumVelocity = v
}
func (z *ZiquePlayer) SetDrumPattern(dp string) {
	z.DrumPattern = dp
	z.midiPlayer.PlayCtrl <- CPlayCtrl{CDRUM, dp, 1.0}
}
func (z *ZiquePlayer) SetSwingPattern(dp string) {
	z.SwingPattern = dp
	z.midiPlayer.PlayCtrl <- CPlayCtrl{CSWING, dp, 1.0}
}
func (z *ZiquePlayer) AlterSwingPattern(coeff float64) {
	z.midiPlayer.PlayCtrl <- CPlayCtrl{CSWING, z.SwingPattern, coeff}
}

func (z *ZiquePlayer) SetVelocityPattern(dp string) {
	z.VelocityPattern = dp
	z.midiPlayer.PlayCtrl <- CPlayCtrl{CVELOCITY, dp, 1.0}
}
func (z *ZiquePlayer) AlterVelocityPattern(coeff float64) {
	z.midiPlayer.PlayCtrl <- CPlayCtrl{CVELOCITY, z.VelocityPattern, coeff}
}
func (z *ZiquePlayer) stopRequest() {
	z.midiPlayer.Stop()
}

func (z *ZiquePlayer) Kill() {
	z.ZiqueCtrl <- 0
}
func (z *ZiquePlayer) Play(tune string) {
	set := []SetElem{SetElem{File: tune, Count: 0}}
	z.PlaySet(set)
}
func (z *ZiquePlayer) PlaySet(tunes MusicSet) {
	z.Stop()
	log.Println("Midi-ZPlay:", tunes)
	z.ZiqueCtrl <- tunes
}
func (z *ZiquePlayer) Stop() {
	log.Print("Midi Stop")
	z.ZiqueCtrl <- 1
}

func (z *ZiquePlayer) mainLoop(Cmd chan interface{}) {
	log.Println("Midi start Midi Loop Ctrl")

	z.midiPlayer = MakePlayer("Dummy")
	z.playing.Store(false)

	for {
		select {
		case MeasureTick := <-z.RtFeedBack:
			select {
			case z.TickBack <- Tick{
				Beats:             z.midiPlayer.Beats,
				BeatType:          z.midiPlayer.BeatType,
				XmlDivisions:      z.midiPlayer.XmlDivisions,
				TickTime:          MeasureTick.Time,
				MeasureId:         MeasureTick.MeasureId,
				MeasureLength:     MeasureTick.MeasureLength,
				MeasureLengthTune: MeasureTick.MeasureLengthTune,
				PassCount:         z.midiPlayer.passCount}:
			default:
				log.Println("Faild to send Tick:", MeasureTick)

			}

		case cmd := <-Cmd:
			log.Println("Midi-Player: Cmd received:", cmd)
			switch v := cmd.(type) {
			case MusicSet:
				z.midiPlayer.Stop()
				z.playing.Store(true)
				go func() {
					z.midiPlayer.PlaySet(v, z.FeedBack)
					z.playing.Store(false)
				}()

			case int:
				z.midiPlayer.Stop()
				//log.Println("Midi-Player: Exit playing thread")
			}
		}
	}
}
