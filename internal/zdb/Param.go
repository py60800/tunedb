// Param.go
package zdb

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/py60800/tunedb/internal/util"
)

var DataBase = "FolkTune.db"
var RPatch = map[string]int{
	"Acoustic Grand Piano": 1,
	//	"Bright Acoustic Piano":   2,
	//	"Electric Grand Piano":    3,
	//	"Honky-tonk Piano":        4,
	//	"Electric Piano 1":        5,
	//	"Electric Piano 2":        6,
	//	"Harpsichord":             7,
	//	"Clavi":                   8,
	//	"Celesta":                 9,
	//	"Glockenspiel":            10,
	//	"Music Box":               11,
	//	"Vibraphone":              12,
	//	"Marimba":                 13,
	//	"Xylophone":               14,
	//	"Tubular Bells":           15,
	//	"Dulcimer":                16,
	//	"Drawbar Organ":           17,
	//	"Percussive Organ":        18,
	"Rock Organ": 19,
	//	"Church Organ":            20,
	//	"Reed Organ":              21,
	"Accordion": 22,
	"Harmonica": 23,
	//	"Tango Accordion":         24,
	//	"Acoustic Guitar (nylon)": 25,
	//	"Acoustic Guitar (steel)": 26,
	//	"Electric Guitar (jazz)":  27,
	//	"Electric Guitar (clean)": 28,
	//	"Electric Guitar (muted)": 29,
	//	"Overdriven Guitar":       30,
	//	"Distortion Guitar":       31,
	//	"Guitar harmonics":        32,
	//	"Acoustic Bass":           33,
	//	"Electric Bass (finger)":  34,
	//	"Electric Bass (pick)":    35,
	//	"Fretless Bass":           36,
	//	"Slap Bass 1":             37,
	//	"Slap Bass 2":             38,
	//	"Synth Bass 1":            39,
	//	"Synth Bass 2":            40,
	"Violin": 41,
	"Viola":  42,
	//	"Cello":                   43,
	//	"Contrabass":              44,
	//	"Tremolo Strings":         45,
	//	"Pizzicato Strings":       46,
	//	"Orchestral Harp":         47,
	"Timpani": 48,
	//	"String Ensemble 1":       49,
	//	"String Ensemble 2":       50,
	//	"SynthStrings 1":          51,
	//	"SynthStrings 2":          52,
	//	"Choir Aahs":              53,
	//	"Voice Oohs":              54,
	//	"Synth Voice":             55,
	//	"Orchestra Hit":           56,
	//	"Trumpet":                 57,
	//	"Trombone":                58,
	//	"Tuba":                    59,
	//	"Muted Trumpet":           60,
	"French Horn": 61,
	//	"Brass Section":           62,
	//	"SynthBrass 1":            63,
	//	"SynthBrass 2":            64,
	"Soprano Sax": 65,
	"Alto Sax":    66,
	//	"Tenor Sax":               67,
	//	"Baritone Sax":            68,
	"Oboe":         69,
	"English Horn": 70,
	//	"Bassoon":                 71,
	//	"Clarinet":                72,
	//	"Piccolo":                 73,
	"Flute": 74,
	//	"Recorder":                75,
	//	"Pan Flute":               76,
	//	"Blown Bottle":            77,
	//	"Shakuhachi":              78,
	"Whistle": 79,
	//	"Ocarina":                 80,
	//	"Lead 1 (square)":         81,
	//	"Lead 2 (sawtooth)":       82,
	//	"Lead 3 (calliope)":       83,
	//	"Lead 4 (chiff)":          84,
	//	"Lead 5 (charang)":        85,
	//	"Lead 6 (voice)":          86,
	//	"Lead 7 (fifths)":         87,
	//	"Lead 8 (bass + lead)":    88,
	//	"Pad 1 (new age)":         89,
	//	"Pad 2 (warm)":            90,
	//	"Pad 3 (polysynth)":       91,
	//	"Pad 4 (choir)":           92,
	//	"Pad 5 (bowed)":           93,
	//	"Pad 6 (metallic)":        94,
	//	"Pad 7 (halo)":            95,
	//	"Pad 8 (sweep)":           96,
	//	"FX 1 (rain)":             97,
	//	"FX 2 (soundtrack)":       98,
	//	"FX 3 (crystal)":          99,
	//	"FX 4 (atmosphere)":       100,
	//	"FX 5 (brightness)":       101,
	//	"FX 6 (goblins)":          102,
	//	"FX 7 (echoes)":           103,
	//	"FX 8 (sci-fi)":           104,
	//	"Sitar":                   105,
	"Banjo": 106,
	//	"Shamisen":                107,
	//	"Koto":                    108,
	//	"Kalimba":                 109,
	//	"Bag pipe":                110,
	"Fiddle": 111,
	//	"Shanai":                  112,
	//	"Tinkle Bell":             113,
	//	"Agogo":                   114,
	//	"Steel Drums":             115,
	//	"Woodblock":               116,
	//	"Taiko Drum":              117,
	//	"Melodic Tom":             118,
	//	"Synth Drum":              119,
	//	"Reverse Cymbal":          120,
	//	"Guitar Fret Noise":       121,
	//	"Breath Noise":            122,
	//	"Seashore":                123,
	//	"Bird Tweet":              124,
	//	"Telephone Ring":          125,
	//	"Helicopter":              126,
	//	"Applause":                127,
	//	"Gunshot":                 128,
}

var RDrum = map[string]int{
	"Acoustic Bass Drum": 35,
	"Bass Drum 1":        36,
	"Side Stick":         37,
	"Acoustic Snare":     38,
	"Hand Clap":          39,
	"Electric Snare":     40,
	"Low Floor Tom":      41,
	"Closed Hi Hat":      42,
	"High Floor Tom":     43,
	"Pedal Hi-Hat":       44,
	"Low Tom":            45,
	"Open Hi-Hat":        46,
	"Low-Mid Tom":        47,
	"Hi-Mid Tom":         48,
	"Crash Cymbal 1":     49,
	"High Tom":           50,
	"Ride Cymbal 1":      51,
	"Chinese Cymbal":     52,
	"Ride Bell":          53,
	"Tambourine":         54,
	"Splash Cymbal":      55,
	"Cowbell":            56,
	"Crash Cymbal 2":     57,
	"Vibraslap":          58,
	"Ride Cymbal 2":      59,
	"Hi Bongo":           60,
	"Low Bongo":          61,
	"Mute Hi Conga":      62,
	"Open Hi Conga":      63,
	"Low Conga":          64,
	"High Timbale":       65,
	"Low Timbale":        66,
	"High Agogo":         67,
	"Low Agogo":          68,
	"Cabasa":             69,
	"Maracas":            70,
	"Short Whistle":      71,
	"Long Whistle":       72,
	"Short Guiro":        73,
	"Long Guiro":         74,
	"Claves":             75,
	"Hi Wood Block":      76,
	"Low Wood Block":     77,
	"Mute Cuica":         78,
	"Open Cuica":         79,
	"Mute Triangle":      80,
	"Open Triangle":      81,
}

// Tune kind -----------------------------------------------
var TuneKindDefaults = []TuneKind{
	TuneKind{0, "Reel", 180},
	TuneKind{0, "Jig", 140},
	TuneKind{0, "SlipJig", 140},
	TuneKind{0, "Hornpipe", 160},
	TuneKind{0, "Polka", 140},
	TuneKind{0, "SetDance", 140},
	TuneKind{0, "March", 100},
	TuneKind{0, "Airs", 80},
	TuneKind{0, "Mazurka", 120},
	TuneKind{0, "Waltz", 100},
	TuneKind{0, "O'Carolan", 120},
	TuneKind{0, "Fling", 140},
	TuneKind{0, "BarnDance", 140},
	TuneKind{0, "Misc", 120},
}
var TuneKindStr []string
var TuneKindTempo map[string]int
var TuneKindIdx map[string]int

func CleanTuneKind(t string) string {
	cl := []rune(strings.ToLower(t))
	clt := make([]rune, 0)
	for _, r := range cl {
		if r == '%' {
			break
		}
		if unicode.IsLetter(r) {
			clt = append(clt, r)
		}
	}
	str := string(clt)
	iMax := -1
	max := -1
	for i, k := range TuneKindStr {
		kt := strings.ToLower(k)
		if kt == str {
			return k
		}
		j := 0
		for j = 0; j < len(kt) && j < len(str); j++ {
			if str[j] != kt[j] {
				break
			}
		}
		if j > max {
			max = j
			iMax = i
		}
	}
	if iMax >= 0 {
		return TuneKindStr[iMax]
	}
	return "*"

}

// Fun level -------------------------------------------------------
type FunLevel int

const (
	FlUnClassified FunLevel = iota
	FlNiceToKnow
	FlPrettyGood
	FlGreatTune
	FlGreatFun
)

var FunLevelStr = []string{"UnClassified", "NiceToKnow", "PrettyGood", "GreatTune", "GreatFun"}
var FunLevelMax = len(FunLevelStr)

func (n FunLevel) String() string {
	return fmt.Sprintf("%d %s", int(n), FunLevelStr[n])
}

// PlayLevel ---------------------------------------------------------
type PlayLevel int

const (
	PlToLearn PlayLevel = iota
	PlVeryBad
	PlOnTheWay
	PlAlmost
	PlPrettyClose
	PlGood
	PlGreat
	PlLimit
)

var PlayLevelStr = []string{"ToLearn", "VeryBad", "OnTheWay", "Almost", "PrettyClose", "Good", "Great"}
var PlayLevelMax = len(PlayLevelStr)

func (n PlayLevel) String() string {
	return fmt.Sprintf("%d %s", int(n), PlayLevelStr[n])
}

// Notes and modes ----------------------------------
var NoteListStr = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

type Note int

const (
	_C Note = iota
	_Db
	_D
	_Eb
	_E
	_F
	_Gb
	_G
	_Ab
	_A
	_Bb
	_B
)

var NodeToInt = map[string]Note{
	"C": _C, "C#": _Db, "Cb": _B,
	"D": _D, "D#": _Eb, "Db": _Db,
	"E": _E, "E#": _F, "Eb": _Eb,
	"F": _F, "F#": _Gb, "Fb": _E,
	"G": _G, "G#": _Ab, "Gb": _Gb,
	"A": _A, "A#": _Bb, "Ab": _Ab,
	"B": _B, "B#": _C, "Bb": _Bb,
}
var CircleOfFifthMaj = map[Note]Note{
	0: _C,
	1: _G, 2: _D, 3: _A, 4: _E, 5: _B, 6: _Gb,
	-1: _F, -2: _Bb, -3: _Eb, -4: _Ab, -5: _Db, -6: _Gb,
}
var ModeDeltaToMaj = map[string]Note{
	"Maj": 0,
	"min": 9,
	"Mix": 7,
	"Dor": 2,
	"Phr": 4,
	"Lyd": 5,
	"Loc": 11,
}
var ModeStrOrdered = []string{"Maj", "min", "Mix", "Dor", "Phr", "Lyd", "Loc"}
var XmlModeXRef = map[string]string{
	"none":       "-",
	"major":      "Maj",
	"minor":      "min",
	"mixolidian": "Mix",
	"dorian":     "Dor",
	"phygian":    "Phr",
	"lydian":     "Lyd",
	"locrian":    "Loc",
	"ionian":     "Maj",
	"aeolian":    "Aeo",
}
var XmlModeStr = []string{"none", "major", "minor", "dorian", "phrygian", "lydian", "mixolydian", "aeolian", "ionian", "locrian"}

func ModeCommonName(fifth Note, mode string) string {
	base := CircleOfFifthMaj[fifth]
	delta := ModeDeltaToMaj[mode]
	return NoteToString(base+delta, fifth > 0) + mode
}
func NoteToString(n Note, p bool) string {
	var c1 = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	var c2 = []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}
	if p {
		return c1[n%12]
	}
	return c2[n%12]
}
func ModesForFifth(fifth Note) []string {
	s := make([]string, len(ModeStrOrdered))
	for i, k := range ModeStrOrdered {
		s[i] = NoteToString(CircleOfFifthMaj[fifth]+ModeDeltaToMaj[k], fifth > 0) + k
	}
	return s
}
func ModeIdx(str string) int {
	b := []byte(str)
	if len(b) >= 3 {
		mod := string(b[len(b)-3:])
		for i := range ModeStrOrdered {
			if mod == ModeStrOrdered[i] {
				return i
			}
		}
	}
	return -1
}
func ModeXmlConvert(fifth Note, xmlMode string) string {
	base := CircleOfFifthMaj[fifth]
	if stdMode, ok := XmlModeXRef[xmlMode]; ok {
		if delta, ok := ModeDeltaToMaj[stdMode]; ok {
			return NoteToString(base+delta, fifth > 0) + stdMode
		}
	}
	return NoteToString(base, fifth > 0)

}

// According to musicxml
var ModeStr = []string{"none", "major", "minor", "dorian", "phrygian", "lydian", "mixolydian", "aeolian", "ionian", "locrian"}

var FifthsStr = []string{"bbbbbb", "bbbbb", "bbbb", "bbb", "bb", "b", "-", "#", "##", "###", "####", "#####", "######"}

func FifthIdx(f int) int {
	return f + 6
}
func FifthIdxR(f int) int {
	return f - 6
}

var TuneTags []TuneListTag
var TuneTagUpdated bool

// *****************************************************************************
func tuneKindInit(db *TuneDB) {
	tk := db.TuneKindGetAll()

	if len(tk) == 0 {
		db.TuneKindUpdateAll(TuneKindDefaults)
		tk = TuneKindDefaults
	}
	TuneKindStr = make([]string, 0)
	TuneKindTempo = make(map[string]int)
	TuneKindIdx = make(map[string]int)
	for _, t := range tk {
		TuneKindStr = append(TuneKindStr, t.Kind)
		TuneKindTempo[t.Kind] = t.Tempo
	}
	sort.Strings(TuneKindStr)
	for i, tk := range TuneKindStr {
		TuneKindIdx[tk] = i
	}

}
func TuneTagUpdate(db *TuneDB) {
	TuneTags = db.TuneListTags()
	util.Truncate(TuneTags, 12)
	TuneTagUpdated = true
}
func ParamInit(db *TuneDB) {
	tuneKindInit(db)
	TuneTagUpdate(db)
	practiceDateInit()
}

// Practice date ***************************************************************
type PracticeAge struct {
	d        string
	l        string
	Duration time.Duration
}

var PracticeAgeStr []string // initialized in int
var PracticeAgeParam = []PracticeAge{
	PracticeAge{d: "0s", l: "Now"},
	PracticeAge{d: "72h", l: "3 Days ago"},
	PracticeAge{d: "168h", l: "1 Week ago"},
	PracticeAge{d: "672h", l: "1 Month ago"},
	PracticeAge{d: "2160h", l: "3 Months ago"},
	PracticeAge{d: "4360h", l: "6 Months ago"},
	PracticeAge{d: "100000h", l: "can't remember"},
}

func practiceDateInit() {
	for i := range PracticeAgeParam {
		PracticeAgeStr = append(PracticeAgeStr, PracticeAgeParam[i].l)
		PracticeAgeParam[i].Duration, _ = time.ParseDuration(PracticeAgeParam[i].d)
	}
}
