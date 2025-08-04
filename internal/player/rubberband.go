package player

// #cgo LDFLAGS: -l rubberband
// #include "rubberband/rubberband-c.h"
/*
static RubberBandState native(void *p){
    return (RubberBandState)p;
}
static const float *const * ptr(void * p){
	return (const float *const *) p;
	}

*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"runtime"
	"time"
	"unsafe"
)

type Rubberband struct {
	ptr            unsafe.Pointer
	SampleRequired int
}

func NewRubberband(sampleRate int, channels int, timeRatio float64, pitchRatio float64) *Rubberband {
	var rb Rubberband
	options := C.RubberBandOptionProcessRealTime | C.RubberBandOptionDetectorCompound
	rb.ptr = unsafe.Pointer(C.rubberband_new(C.uint(sampleRate), C.uint(channels), C.int(options), C.double(timeRatio), C.double(pitchRatio)))
	rb.SampleRequired = int(C.rubberband_get_samples_required(C.native(rb.ptr)))
	fmt.Println("Rubberband new")
	return &rb
}
func (rb *Rubberband) SetTimeRatio(timeRatio float64) {
	C.rubberband_set_time_ratio(C.native(rb.ptr), C.double(timeRatio))
}
func (rb *Rubberband) SetPitchScale(pitchScale float64) {
	C.rubberband_set_pitch_scale(C.native(rb.ptr), C.double(pitchScale))
}

var rbCount = 0
var rbTime time.Duration

func (rb *Rubberband) Delete() {
	C.rubberband_delete(C.native(rb.ptr))
	if rbCount > 0 {
		fmt.Printf("Rubberband Delete: %d %v\n", rbCount, rbTime/time.Duration(rbCount))
	} else {
		fmt.Printf("Rubberband Delete: unused\n")
	}
	rb.ptr = nil
}

func (rb *Rubberband) ProcessI16(data []byte, channels int) []byte {
	rbCount++
	start := time.Now()
	defer func() {
		rbTime += time.Now().Sub(start)
	}()

	var pinner runtime.Pinner
	c := make([][]C.float, channels)
	sz := len(data) / (channels * 2)
	for i := range c {
		c[i] = make([]C.float, sz)
	}
	t := make([]int16, len(data)/2)
	reader := bytes.NewReader(data)
	binary.Read(reader, binary.LittleEndian, &t)
	for i := 0; i < sz; i++ {
		for j := range c {
			c[j][i] = C.float(t[j+i*channels])
		}
	}

	d := make([]unsafe.Pointer, channels)
	pinner.Pin(&d)
	defer pinner.Unpin()
	for i := range d {
		d[i] = unsafe.Pointer(&(c[i][0]))
		pinner.Pin(d[i])
	}

	C.rubberband_process(C.native(rb.ptr), C.ptr(unsafe.Pointer(&(d[0]))), C.uint(sz), 0)
	available := C.rubberband_available(C.native(rb.ptr))
	if available == 0 {
		fmt.Println("lazy rubberband")
		return []byte{}
	}

	res := make([][]C.float, channels)
	for i := range channels {
		res[i] = make([]C.float, available)
	}
	pres := make([]unsafe.Pointer, channels)
	for i := range channels {
		pres[i] = unsafe.Pointer(&(res[i][0]))
		pinner.Pin(pres[i])
	}

	n := int(C.rubberband_retrieve(C.native(rb.ptr), C.ptr(unsafe.Pointer(&(pres[0]))), C.uint(available)))
	bres := make([]int16, n*channels)
	for i := range n {
		for j := range channels {
			bres[i*channels+j] = int16(res[j][i])
		}
	}
	resBuffer := make([]byte, len(bres)*2)
	binary.Encode(resBuffer, binary.LittleEndian, &bres)
	return resBuffer
}
