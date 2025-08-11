package zique

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	RtMidiChan     chan SeqEvent
	RtCtrlChan     chan int
	RtMidiSinkChan chan []byte
)

/*
	type DrumBeat struct {
		Instrument int
		Velocity   int
		Duration   float64
	}
*/
var (
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

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ln := scanner.Text()
		if !strings.HasPrefix(ln, "#") {
			lines = append(lines, scanner.Text())
		}
	}
	return lines, scanner.Err()
}

type ZiquePlayer struct {
	ZiqueCtrl       chan interface{}
	DrumPattern     string
	VelocityPattern string
	SwingPattern    string
	player          *Player
	FeedBack        chan string
	TickBack        chan Tick
	RtFeedBack      chan RtMeasureTick
}

func ZiquePlayerNew(context string, midiPort string) (*ZiquePlayer, string) {
	z := ZiquePlayer{
		ZiqueCtrl: make(chan interface{}, 10),
		FeedBack:  make(chan string, 2),
		TickBack:  make(chan Tick, 2),
	}
	msg := z.init(context, midiPort)
	return &z, msg
}

func (z *ZiquePlayer) init(context string, midiPort string) string {

	InitPattern(context)
	if midiPort == "" {
		midiPort = "Synth"
	}

	RtMidiChan = make(chan SeqEvent, 3)
	RtCtrlChan = make(chan int, 3)
	RtMidiSinkChan = make(chan []byte, 0)
	RtStartMsg := make(chan string)

	z.RtFeedBack = make(chan RtMeasureTick, 2)

	go RtMidiSink(midiPort, RtMidiSinkChan, RtStartMsg)
	go MidiSink(RtMidiChan, RtCtrlChan, RtMidiSinkChan, z.RtFeedBack)

	go z.mainLoop(z.ZiqueCtrl)
	msg := <-RtStartMsg
	return msg
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
	z.player.PlayCtrl <- CPlayCtrl{CDRUM, dp, 1.0}

}
func (z *ZiquePlayer) SetSwingPattern(dp string) {
	z.SwingPattern = dp
	z.player.PlayCtrl <- CPlayCtrl{CSWING, dp, 1.0}

}
func (z *ZiquePlayer) AlterSwingPattern(coeff float64) {
	z.player.PlayCtrl <- CPlayCtrl{CSWING, z.SwingPattern, coeff}

}

func (z *ZiquePlayer) SetVelocityPattern(dp string) {
	z.VelocityPattern = dp
	z.player.PlayCtrl <- CPlayCtrl{CVELOCITY, dp, 1.0}

}
func (z *ZiquePlayer) AlterVelocityPattern(coeff float64) {
	z.player.PlayCtrl <- CPlayCtrl{CVELOCITY, z.VelocityPattern, coeff}
}

func (z *ZiquePlayer) Kill() {
	z.ZiqueCtrl <- 0
}
func (z *ZiquePlayer) Play(tune string) {
	set := []SetElem{SetElem{File: tune, Count: 0}}
	z.PlaySet(set)
}
func (z *ZiquePlayer) PlaySet(tunes MusicSet) {
	RtCtrlChan <- 3 // Resume Pause
	log.Println("Midi-ZPlay:", tunes)
	z.ZiqueCtrl <- tunes
}
func (z *ZiquePlayer) Pause() {
	RtCtrlChan <- 1
}
func (z *ZiquePlayer) Stop() {
	log.Print("Midi Stop")
	z.ZiqueCtrl <- 1
	log.Println("...ping!")
}

func (z *ZiquePlayer) mainLoop(Cmd chan interface{}) {
	var wg sync.WaitGroup
	pl := MakePlayer("Dummy")
	z.player = &pl
	var set MusicSet
	Playing := false
	var passCount int
	PlayingThread := func() {
		log.Println("Midi-Start playing")
		wg.Add(1)
		defer func() {
			Playing = false
			wg.Done()
			log.Println("Midi-Exit playing")
		}()
		log.Println("Midi-Play:", set)
		for iset, t := range set {
			log.Println("Midi-Parse:", t.File)
			select {
			case z.FeedBack <- t.File:
			default:
			}
			partition, err := Parse(t.File)
			log.Println("Midi-Parse:", err, len(partition.Part))
			if err != nil || len(partition.Part) == 0 {
				log.Println(partition)
				break
			}
			count := t.Count
			passCount = 1
			barCount := -1
			signalBar := -1
			for Playing {
				barCount = pl.PlayTune(partition.Part[0], signalBar, func() {
					if iset != len(set)-1 {
						select {
						case z.FeedBack <- set[iset+1].File:
						default:
						}

					}
				})
				passCount++
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
			if !Playing {
				break
			}
		}

	}

	for {
		select {
		case MeasureTick := <-z.RtFeedBack:

			select {
			case z.TickBack <- Tick{Beats: z.player.Beats, BeatType: z.player.BeatType,
				XmlDivisions: z.player.XmlDivisions,
				TickTime:     MeasureTick.Time,
				MeasureId:    MeasureTick.MeasureId,
				PassCount:    passCount}:
			default:
			}

		case cmd := <-Cmd:
			log.Println("Midi-Player: Cmd received:", cmd)
			switch v := cmd.(type) {
			case MusicSet:
				if Playing {
					pl.PlayCtrl <- CPlayCtrl{CSTOP, "", 0.0}
					Playing = false
					log.Println("Midi-Player: Wg wait")
					wg.Wait()
				}

				set = v
				Playing = true
				go PlayingThread()

			case int:
				switch v {
				case 0:
					pl.PlayCtrl <- CPlayCtrl{CSTOP, "", 0.0}
					if Playing {
						Playing = false
						log.Println("Midi-Player: Wg wait")
						wg.Wait()
					}
					log.Println("Midi-Player: Exit playing thread")
					return
				case 1:
					if Playing {
						pl.PlayCtrl <- CPlayCtrl{CSTOP, "", 0.0}
						Playing = false
						log.Println("Midi-Player: Wg wait")
						wg.Wait()
					}
					log.Println("Midi-Player: Stopped")

				}
			}
		}
	}
}
