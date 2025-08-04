package zique

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	_ "embed"

	"github.com/py60800/tunedb/internal/util"
)

//go:embed pattern.yml
var defaultPattern string

// ******************************************************
const (
	C  = 0
	Db = 1
	D  = 2
	Eb = 3
	E  = 4
	F  = 5
	Gb = 6
	G  = 7
	Ab = 8
	A  = 9
	Bb = 10
	B  = 11
)

var ChordTemplate = map[string][]int{
	"major":         []int{C, E, G},
	"minor":         []int{C, Eb, G},
	"augmented":     []int{C, E, Ab},
	"diminished":    []int{C, Eb, Gb},
	"major-seventh": []int{C, E, G, B},
	"minor-seventh": []int{C, Eb, G, Bb},
	"none":          []int{},
}

type CRoll struct {
	Note     string
	Type     string
	Dot      bool
	Seq      []string
	Timing   []int
	Velocity []int
}
type CPatternConfig struct {
	Roll            []CRoll `yaml:"roll"`
	DrumPatch       map[string]int
	DrumBeat        map[string][]string
	DrumPattern     map[string][]string
	VelocityPattern map[string]CVelocityPattern
	SwingPattern    map[string]CSwingPattern
	ChordPattern    map[string]CChordPattern
}

var PatternConfig CPatternConfig

type RPattern struct {
	Note  string
	Alter int
	NType string
	Dot   bool
}
type RNote struct {
	Note        string
	Duration    int
	RelVelocity int
}
type RNoteSeq []RNote

var RollPattern map[RPattern]RNoteSeq

// **********************
type CVelocityPattern struct {
	Measure  int
	Velocity []float64
}

type CSwingPattern struct {
	Measure int
	Delta   []float64
}
type CChord []string

type CChordPattern struct {
	Measure  int
	Duration []float64
	Velocity []int
}

// ****************************************************
type Beat struct {
	Instrument int
	Velocity   int
	Duration   float64
}
type CDrumPattern struct {
	Measure int `yaml:"measure"`
	Beats   []Beat
}

func rPattern(n string, ntypen string, dot bool) RPattern {
	alter := 0
	if strings.HasSuffix(n, "#") {
		alter = 1
	}
	if strings.HasSuffix(n, "b") {
		alter = -1
	}
	str := string(n[0])
	ntype := "quarter"
	switch ntypen {
	case "quarter":
	case "q":
		ntype = "quarter"
	}
	return RPattern{str, alter, ntype, dot}

}
func noteSeq(notes []string, timing []int, Velocity []int) RNoteSeq {
	var r RNoteSeq
	r = make([]RNote, len(notes))
	for i, n := range notes {
		r[i] = RNote{n, timing[i], Velocity[i]}
	}
	return r
}
func GetDrumPattern(name string) []Beat {
	if dp, ok := PatternConfig.DrumPattern[name]; ok {
		res := make([]Beat, len(dp))
		for i, b := range dp {
			if beat, ok := PatternConfig.DrumBeat[b]; ok {
				patch, duration, velocity := 0, 1.0, 100
				switch len(beat) {
				case 3:
					velocity, _ = strconv.Atoi(beat[2])
					fallthrough
				case 2:
					duration, _ = strconv.ParseFloat(beat[1], 64)
					fallthrough
				case 1:
					var ok bool
					if patch, ok = PatternConfig.DrumPatch[beat[0]]; !ok {
						fmt.Println("Invalid drum patch:", beat[0])
						return []Beat{}
					}
				}
				res[i] = Beat{Instrument: patch, Duration: duration, Velocity: velocity}
			} else {
				fmt.Println("Invalid Beat:", b)
				return []Beat{}
			}
		}
		return res
	}
	fmt.Println("Drum Pattern not found:", name, PatternConfig.DrumPattern)
	return []Beat{}
}
func normalize(n []float64) []float64 {
	max := 0.0
	r := make([]float64, len(n))
	for _, t := range n {
		if t > max {
			max = t
		}
	}
	if max == 0 {
		max = 1.0
	}
	for i, t := range n {
		r[i] = t / max
	}
	return r
}
func GetVelocityPattern(name string) []float64 {
	if dp, ok := PatternConfig.VelocityPattern[name]; ok {
		return normalize(dp.Velocity)
	}
	fmt.Println("Velocity Pattern not found:", name)
	return []float64{1.0}

}
func GetSwingPattern(name string) []float64 {
	if dp, ok := PatternConfig.SwingPattern[name]; ok {
		return dp.Delta
	}
	fmt.Println("Swing Pattern not found:", name)
	return []float64{0.0}

}

type ChordStroke struct {
	Duration float64
	Velocity int
}

func GetChordPattern(name string) []ChordStroke {
	if len(name) == 0 {
		return []ChordStroke{}
	}
	if ch, ok := PatternConfig.ChordPattern[name]; ok {
		if len(ch.Duration) != len(ch.Velocity) {
			fmt.Println("Invalid Chord Pattern:", name)
			return []ChordStroke{}
		}
		res := make([]ChordStroke, len(ch.Duration))
		for i := range ch.Duration {
			res[i] = ChordStroke{ch.Duration[i], ch.Velocity[i]}
		}
		return res
	}
	fmt.Println("Chord pattern not found:", name)
	return []ChordStroke{}
}

func InitPattern(context string) {
	file := path.Join(context, "pattern.yml")
	PatternConfig = util.ReadConfig[CPatternConfig](file, defaultPattern)
	RollPattern = make(map[RPattern]RNoteSeq)

	for _, p := range PatternConfig.Roll {
		RollPattern[rPattern(p.Note, p.Type, p.Dot)] = noteSeq(p.Seq, p.Timing, p.Velocity)
	}
}
