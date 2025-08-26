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
	playLock sync.Mutex

	file string

	reader  *os.File
	decoder *mp3.Decoder

	sampleSize int
	sampleRate int
	channels   int
	//	cmd        chan int

	otoContext *oto.Context
	otoPlayer  *oto.Player

	from float64
	to   float64

	fromIdx      int
	toIdx        int
	bufferLength int

	context map[int]*oto.Context

	stopRequest atomic.Bool
	isPlaying   atomic.Bool

	dataLock sync.Mutex

	progress   float64
	timeRatio  float64
	pitchScale float64
	duration   float64
}

func (m *Mp3Player) GetSpeedAndPitch() (float64, float64) {
	m.dataLock.Lock()
	defer m.dataLock.Unlock()
	return m.timeRatio, m.pitchScale
}
func (m *Mp3Player) SetTimeRatio(timeRatio float64) {
	m.dataLock.Lock()
	defer m.dataLock.Unlock()
	m.timeRatio = timeRatio
}
func (m *Mp3Player) SetPitchScale(pitchScale float64) {
	m.dataLock.Lock()
	defer m.dataLock.Unlock()
	m.pitchScale = pitchScale
}
func (m *Mp3Player) GetProgress() float64 {
	m.dataLock.Lock()
	defer m.dataLock.Unlock()
	return m.progress
}
func (m *Mp3Player) setProgress(p float64) {
	m.dataLock.Lock()
	defer m.dataLock.Unlock()
	m.progress = p
}
func (m *Mp3Player) GetDuration() float64 {
	m.dataLock.Lock()
	defer m.dataLock.Unlock()
	return m.duration
}
func (m *Mp3Player) setDuration(d float64) {
	m.dataLock.Lock()
	defer m.dataLock.Unlock()
	m.duration = d
}

func (m *Mp3Player) selectContext(sampleRate int) *oto.Context {

	var ok bool
	if m.otoContext, ok = m.context[sampleRate]; ok {
		m.sampleRate = sampleRate
		return m.otoContext
	}
	log.Println("Oto Create Context for SampleRate:", sampleRate)
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
		//cmd:         make(chan int, 1),
		//doneChannel: make(chan int, 1),
		file:       "",
		from:       0.0,
		to:         0.0,
		duration:   1.0,
		timeRatio:  1.0,
		pitchScale: 1.0,
		progress:   0,
		context:    make(map[int]*oto.Context),
	}
	return m
}

func (m *Mp3Player) IsPlaying() bool {
	return m.isPlaying.Load()
}

func (m *Mp3Player) LoadFile(file string) (duration float64, err error) {
	m.Stop() // Before lock
	m.playLock.Lock()
	defer m.playLock.Unlock()

	log.Println("Load MP3:", file)
	if m.file == file {
		return m.GetDuration(), nil
	}

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
	duration = m.idxToTime(m.bufferLength)
	m.setDuration(duration)
	return duration, nil

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
	m     *Mp3Player
	p     *oto.Player
	r     io.Reader
	idx   int
	idxTo int

	dataChan chan *[]byte
	data     *[]byte
	idata    int
}

func (spy *spyReader) readingThread(stopRequest *atomic.Bool) {
	log.Println("MP3 Start Reader")
	var rubberBand *Rubberband = nil
	defer func() {
		log.Println("MP3 Reader finished!")
		close(spy.dataChan)
		if rubberBand != nil {
			rubberBand.Delete()
		}
	}()

	const SampleRequired = 2048
	previousTimeRatio := -1.0
	previousPitchScale := -1.0
	for finished := false; !finished && !stopRequest.Load(); {
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
		spy.m.setProgress(float64(spy.m.idxToTime(spy.idx)))

		spyTimeRatio, spyPitchScale := spy.m.GetSpeedAndPitch()

		if spyTimeRatio == 1.0 && spy.m.pitchScale == 1.0 {
			spy.dataChan <- &buffer
		} else {
			if rubberBand == nil {
				rubberBand = NewRubberband(spy.m.sampleRate, spy.m.channels, spyTimeRatio, spyPitchScale)
			}
			if spyTimeRatio != previousTimeRatio {
				rubberBand.SetTimeRatio(spyTimeRatio)
				previousTimeRatio = spyTimeRatio
			}
			if spyPitchScale != previousPitchScale {
				rubberBand.SetPitchScale(spyPitchScale)
				previousPitchScale = spyPitchScale
			}
			result := rubberBand.ProcessI16(buffer, spy.m.channels)
			spy.dataChan <- &result
		}
	}

}

func (spy *spyReader) Read(buffer []byte) (int, error) {
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

func (m *Mp3Player) singlePlay() {
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

	m.otoPlayer = m.otoContext.NewPlayer(spy)
	spy.p = m.otoPlayer
	m.otoPlayer.SetBufferSize(4096) // ~0.1 second
	m.setProgress(float64(m.idxToTime(m.fromIdx)))

	// Start reading thread
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		spy.readingThread(&m.stopRequest)
		wg.Done()
	}()

	m.otoPlayer.Play()

	wg.Wait()
	m.otoPlayer.Close()

}
func (m *Mp3Player) stopRequested() bool {
	return m.stopRequest.Load()
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

	for !m.stopRequested() {
		m.singlePlay()
		if playMode == PMPlayOnce {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (m *Mp3Player) Stop() {
	m.stopRequest.Store(true)
}
func (m *Mp3Player) Play(From float64, To float64, playMode int) {
	m.Stop()
	m.setProgress(0.0)
	m.isPlaying.Store(true)
	go func() {
		m.playLock.Lock()
		defer m.playLock.Unlock()
		m.isPlaying.Store(true)
		m.stopRequest.Store(false)
		m.play(From, To, playMode)
		m.isPlaying.Store(false)
	}()
}
