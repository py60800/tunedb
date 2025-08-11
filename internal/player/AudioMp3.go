// AudioMp3
package player

import (
	"io"
	"log"

	//	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"

	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/py60800/oto/v3"
	"github.com/py60800/tunedb/internal/util"
)

type Mp3Player struct {
	file        string
	isPlaying   bool
	stopRequest bool

	reader  *os.File
	decoder *mp3.Decoder

	sampleSize int
	sampleRate int
	duration   float64
	channels   int
	cmd        chan int

	progress atomic.Value

	wgPlay      sync.WaitGroup
	doneChannel chan int

	otoContext *oto.Context
	player     *oto.Player

	from float64
	to   float64

	fromIdx      int
	toIdx        int
	bufferLength int
	timeRatio    float64
	pitchScale   float64

	context map[int]*oto.Context
}

func (m *Mp3Player) selectContext(sampleRate int) *oto.Context {

	var ok bool
	if m.otoContext, ok = m.context[sampleRate]; ok {
		m.sampleRate = sampleRate
		return m.otoContext
	}
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 2,
		//		Format:       oto.FormatFloat32LE,
		Format: oto.FormatSignedInt16LE,
	}
	otoCtx, readyChan, err := oto.NewContext(op)
	util.PanicOnError(err)

	<-readyChan
	m.otoContext = otoCtx
	m.sampleRate = sampleRate
	m.context[sampleRate] = otoCtx
	return m.otoContext
}

func Mp3PlayerNew() *Mp3Player {
	m := &Mp3Player{
		cmd:         make(chan int, 1),
		doneChannel: make(chan int, 1),
		file:        "",
		from:        0.0,
		to:          0.0,
		duration:    1.0,
		timeRatio:   1.0,
		pitchScale:  1.0,
	}
	m.progress.Store(0.0)
	m.context = make(map[int]*oto.Context)

	m.progress.Store(0.0)
	return m

}
func (m *Mp3Player) SetTimeRatio(timeRatio float64) {
	m.timeRatio = timeRatio
}
func (m *Mp3Player) SetPitchScale(pitchScale float64) {
	m.pitchScale = pitchScale
}
func (m *Mp3Player) GetProgress() float64 {
	v := m.progress.Load()
	return v.(float64)
}
func (m *Mp3Player) IsPlaying() bool {
	return m.isPlaying
}
func (m *Mp3Player) CurrentFile() string {
	return m.file
}
func (m *Mp3Player) LoadFile(file string) (duration float64, err error) {
	if m.file == file {
		return m.duration, nil
	}
	m.Stop()
	if m.reader != nil {
		m.reader.Close()
	}
	m.file = file
	m.reader, err = os.Open(file)
	if err != nil {
		m.reader = nil
		m.file = ""
		return 0.0, err
	}
	m.decoder, err = mp3.NewDecoder(m.reader)
	if err != nil {
		m.reader.Close()
		m.reader = nil
		m.file = ""
		return 0.0, err
	}

	m.selectContext(m.decoder.SampleRate())

	m.channels = 2 // Hardcoded in mp3 decoder
	m.sampleSize = 4
	m.bufferLength = int(m.decoder.Length())
	m.duration = m.idxToTime(m.bufferLength)
	return m.duration, nil

}
func (m *Mp3Player) Duration() float64 {
	return m.duration
}

const (
	PMPlayOnce = iota
	PMPlayRepeat
)

func (m *Mp3Player) idxToTime(i int) float64 {
	return float64(i/m.sampleSize) / float64(m.sampleRate)
}
func (m *Mp3Player) timeToIdx(t float64) int {
	idx := int(t*float64(m.sampleRate)) * m.sampleSize
	return max(0, idx)
}

type spyReader struct {
	m          *Mp3Player
	p          *oto.Player
	r          io.Reader
	idx        int
	idxTo      int
	timeRatio  float64
	pitchScale float64
	wg         sync.WaitGroup

	dataChan chan *[]byte
	data     *[]byte
	idata    int
}

func (spy *spyReader) readingThread() {
	log.Println("Start Reader")
	var rubberBand *Rubberband = nil
	defer func() {
		log.Println("Reader finished!")
		close(spy.dataChan)
		if rubberBand != nil {
			rubberBand.Delete()
		}
		spy.wg.Done()
	}()
	spy.timeRatio = spy.m.timeRatio
	spy.pitchScale = spy.m.pitchScale
	const SampleRequired = 2048

	for finished := false; !finished && !spy.m.stopRequest; {
		buffer := make([]byte, SampleRequired*spy.m.sampleSize)
		// Read data
		for iRead := 0; iRead < len(buffer); {
			n, err := spy.r.Read(buffer[iRead:])
			if err != nil {
				finished = true
				for ; iRead < len(buffer); iRead++ {
					buffer[iRead] = 0
				}
			} else {
				iRead += n
			}
		}
		spy.idx += len(buffer)
		finished = spy.idx >= spy.idxTo
		spy.m.progress.Store(float64(spy.m.idxToTime(spy.idx)))

		if spy.m.timeRatio == 1.0 && spy.m.pitchScale == 1.0 {
			spy.dataChan <- &buffer
		} else {
			if rubberBand == nil {
				spy.timeRatio, spy.pitchScale = spy.m.timeRatio, spy.m.pitchScale
				rubberBand = NewRubberband(spy.m.sampleRate, spy.m.channels, spy.timeRatio, spy.pitchScale)
			}
			if spy.timeRatio != spy.m.timeRatio {
				spy.timeRatio = spy.m.timeRatio
				rubberBand.SetTimeRatio(spy.timeRatio)
			}
			if spy.pitchScale != spy.m.pitchScale {
				spy.pitchScale = spy.m.pitchScale
				rubberBand.SetPitchScale(spy.pitchScale)
			}
			result := rubberBand.ProcessI16(buffer, spy.m.channels)
			spy.dataChan <- &result
		}
	}

}

func (spy *spyReader) Read(buffer []byte) (int, error) {
	/*	for i := range buffer {
			var ok bool
			if buffer[i], ok = <-spy.m.dataChannel; !ok {
				return 0, io.EOF
			}
		}
		return len(buffer), nil
	*/
	i := 0
	for i = 0; i < len(buffer); {
		if spy.data != nil {
			n := copy(buffer[i:], (*spy.data)[spy.idata:])
			i += n
			spy.idata += n
			if spy.idata >= len(*spy.data) {
				spy.data = nil
			}
		} else {
			var ok bool
			if spy.data, ok = <-spy.dataChan; !ok {
				return 0, io.EOF
			}
			spy.idata = 0
		}
	}
	return i, nil
}

func (m *Mp3Player) _play() {
	m.stopRequest = false
	m.decoder.Seek(int64(m.fromIdx), 0)
	spy := &spyReader{
		m:        m,
		r:        m.decoder,
		idx:      m.fromIdx,
		idxTo:    m.toIdx,
		dataChan: make(chan *[]byte, 4),
		data:     nil,
		idata:    0,
	}

	m.player = m.otoContext.NewPlayer(spy)
	spy.p = m.player
	m.player.SetBufferSize(4096) // ~0.1 second
	m.progress.Store(float64(m.idxToTime(m.fromIdx)))

	go func() {
		spy.wg.Add(1)
		spy.readingThread()
	}()
	m.player.Play()
	for m.player.IsPlaying() {
		time.Sleep(100 * time.Millisecond)
	}
	m.player.Close()
	spy.wg.Wait()
	m.doneChannel <- 0
}
func (m *Mp3Player) play(From float64, To float64, playMode int) {
	if m.file == "" {
		return
	}
	m.fromIdx = m.timeToIdx(From - 0.2)
	m.toIdx = m.timeToIdx(To + 0.2)
	if m.toIdx <= m.fromIdx || m.toIdx > m.bufferLength {
		m.toIdx = m.bufferLength
	}

	go m._play()

	for {
		select {
		case <-m.cmd:
			//m.player.Close()
			<-m.doneChannel
			return
		case <-m.doneChannel:
			switch playMode {
			case PMPlayOnce:
				return
			case PMPlayRepeat:
				time.Sleep(500 * time.Millisecond)
				go m._play()
			}
		}
	}
}
func (m *Mp3Player) Stop() {
	m.stopRequest = true
	select {
	case m.cmd <- 0:
	default:
	}
	m.wgPlay.Wait()
}
func (m *Mp3Player) Play(From float64, To float64, playMode int) {
	m.Stop()
	m.progress.Store(0.0)
	m.isPlaying = true
	go func() {
		m.wgPlay.Add(1)
		for len(m.cmd) > 0 {
			<-m.cmd
		}
		m.play(From, To, playMode)
		m.isPlaying = false
		m.wgPlay.Done()
	}()
}
