// MidiHelper
package zique

import (
	"strings"
)

var NotesToInt = map[string]int{
	"C":  0,
	"C#": 1,
	"DB": 1,
	"D":  2,
	"D#": 3,
	"EB": 3,
	"E":  4,
	"F":  5,
	"F#": 6,
	"GB": 6,
	"G":  7,
	"G#": 8,
	"AB": 8,
	"A":  9,
	"A#": 10,
	"BB": 10,
	"B":  11,
}

// KeyInt converts an A-G note notation to a midi note number value.
func MidiKeyInt(n string, octave int) int {
	key := NotesToInt[strings.ToUpper(n)]
	// octave starts at -2 but first note is at 0
	return key + (octave+2)*12
}
