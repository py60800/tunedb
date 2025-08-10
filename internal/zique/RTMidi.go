// rtmidi.go
package zique

import (
	"fmt"
	"time"

	rtmidi "gitlab.com/gomidi/midi/v2"

	// _ "gitlab.com/gomidi/midi/v2/drivers/portmididrv" // autoregisters driver
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
	//	"sync"
)

func (n PNoteOn) GetRtMidiEvent() []byte {
	return rtmidi.NoteOn(uint8(n.Channel), uint8(n.NoteNumber), uint8(n.Velocity))
}
func (n PNoteOff) GetRtMidiEvent() []byte {
	return rtmidi.NoteOff(uint8(n.Channel), uint8(n.NoteNumber))
}
func (n PTickOn) GetRtMidiEvent() []byte {
	return rtmidi.NoteOn(uint8(DrumChannel), uint8(n.NoteNumber), uint8(n.Velocity))
}

func (n PTickOff) GetRtMidiEvent() []byte {
	return rtmidi.NoteOff(uint8(DrumChannel), uint8(n.NoteNumber))
}
func (n MStart) GetRtMidiEvent() []byte {
	return []byte{}
}
func (n MTimeEv) GetRtMidiEvent() []byte {
	//p.ComputeSwingFactor(n.MLength, n.Beats, n.BeatType)
	return []byte{}
}
func (c PChordOn) GetRtMidiEvent() []byte {
	return []byte{}
}
func (c *PChordOff) GetRtMidiEvent() []byte {
	return []byte{}
}
func (p PProgramChange) GetRtMidiEvent() []byte {
	return rtmidi.ProgramChange(uint8(DefaultChannel), uint8(p.Patch))
}
func (p PTempoChange) GetRtMidiEvent() []byte {
	return []byte{}
}

func GetMidiSink(midiPort string) (func(rtmidi.Message) error, string) {
	Ports := rtmidi.GetOutPorts()
	fmt.Println("Available Midi Ports:")
	for _, p := range Ports {
		fmt.Println("-", p)
	}
	if len(Ports) == 0 {
		fmt.Println("No Midi devices")
		return nil, "Major failure => No MIDI device"
	}
	if out, err := rtmidi.FindOutPort(midiPort); err == nil {
		fmt.Println("Got synth")
		send, err := rtmidi.SendTo(out)
		if err != nil {
			fmt.Println("Midi connection error")
			return nil, "Can't connect to MIDI port"
		}
		return send, ""
	} else {
		// Try a fall back to last channel
		for i := range Ports {
			j := len(Ports) - 1 - i
			send, err := rtmidi.SendTo(Ports[j])
			if err == nil {
				fmt.Println("Can't connect to MIDI port:", midiPort)
				return send, fmt.Sprint("Fallback to midi port:", Ports[j])
			}
		}
	}
	panic("Can't locate midi port")
}

func RtMidiSink(midiPort string, bChan chan []byte, rt chan string) {
	defer rtmidi.CloseDriver()
	send, msg := GetMidiSink(midiPort)
	rt <- msg
	for {
		data, ok := <-bChan
		if !ok {
			break
		}
		if len(data) > 0 {
			if send != nil {
				send(data)
			}
		}
	}
}

type RtMeasureTick struct {
	Time      time.Duration
	MeasureId string
}

func MidiSink(mchan chan SeqEvent, ctrlChan chan int, sendChan chan []byte, feedBack chan RtMeasureTick) {
	defer func() {
		close(sendChan)
	}()
	silence := func() {
		for _, m := range rtmidi.SilenceChannel(-1) {
			sendChan <- m
		}
	}
	silence()
	sendChan <- rtmidi.ProgramChange(uint8(DefaultChannel), uint8(MainInstrument))
	sendChan <- rtmidi.ProgramChange(uint8(ChordChannel), uint8(ChordInstrument))

	TickTime := time.Minute / time.Duration(Tempo*MasterDivisions)

	clock := uint32(0)
	lastEvent := SeqEvent{0, DummyEvent{}}
	for {
		select {
		case evt := <-mchan:
			if evt.Tick == 0 {
				evt.Tick = clock
			}
			switch v := evt.Event.(type) {
			case PTempoChange:
				Tempo = v.Value
				TickTime = time.Minute / time.Duration(Tempo*MasterDivisions)
			case MStart:
				select {
				case feedBack <- RtMeasureTick{Time: TickTime, MeasureId: v.MeasureId}:
				default:
				}

			}
			ntime := evt.Tick
			wait := time.Duration(ntime-clock) * TickTime
			if wait < 0 || wait > 2*time.Second {
				fmt.Printf("Sequence error Event:%v (PrevEvent %d:%v) ntime:%v clock%v ntime-nclock:%v clock-ntime:%v\n ",
					evt.Event.String(),
					lastEvent.Tick, lastEvent.Event.String(),
					ntime, clock, ntime-clock, clock-ntime)

			} else {
				time.Sleep(wait)
			}
			clock = ntime
			sendChan <- evt.Event.GetRtMidiEvent()
			lastEvent = evt

		case cmd := <-ctrlChan:
			switch cmd {
			case 0:
				silence()
				return
			case 1: // Pause
				silence()
				cmd = <-ctrlChan
				if cmd == 0 {
					return
				}
			case 2: // Stop
				silence()
				return
			case 3: // resume pause

			}
		}

	}
}
