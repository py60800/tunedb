// yml
package svgtab

import (
	"fmt"
	"sort"
)

type CButton [2]string
type ButtonLayout struct {
	Left  [][]CButton
	Right [][]CButton
}
type PrintLayout struct {
	Left  [][]string
	Right [][]string
}
type Config struct {
	Layout ButtonLayout
	Coding PrintLayout
}

type Button struct {
	Note int
	Side int
	Row  int
	Rank int
	Pull bool
}

type Note2Button map[int][]Button

var NoteList = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
var baseNotes = map[rune]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}
var note2button Note2Button

func text2note(note string) int {
	if len(note) >= 1 {
		if base, ok := baseNotes[rune(note[0])]; ok {
			for _, c := range note[1:] {
				switch rune(c) {
				case '#':
					base++
				case 'b':
					base--
				case '\'':
					base += 12
				case ',':
					base -= 12
				default:
					fmt.Println("Invalid Note:", note)
				}
			}
			return base + 4*12
		}
	}
	fmt.Println("Invalid note:", note)
	return 0
}

type Button2Text map[Button]string

func GetButton(text string, underline bool) Button {
	coding := config.Coding
	for i, row := range coding.Left {
		for j, t := range row {
			if t == text {
				return Button{Side: 0, Row: i, Rank: j, Pull: underline}
			}
		}
	}
	for i, row := range coding.Right {
		for j, t := range row {
			if t == text {
				return Button{Side: 1, Row: i, Rank: j, Pull: underline}
			}
		}
	}
	return Button{}
}
func (b *Button) Text() (string, bool) {
	cd := &config.Coding
	var l [][]string
	if b.Side == 0 {
		l = cd.Left
	} else {
		l = cd.Right
	}
	if b.Row < len(l) {
		r := l[b.Row]
		if b.Rank < len(r) {
			return r[b.Rank], b.Pull
		}
	}
	return "??", false
}

func configAnalysis(c *Config) Note2Button {
	layout := &c.Layout
	n2b := make(map[int][]Button)
	mapper := func(layout [][]CButton, side int) {
		for ir, row := range layout {
			for jr, button := range row {
				if len(button) != 2 {
					panic(fmt.Errorf("No Push pull for %v %v", ir, jr))
				}
				pushNote := text2note(button[0])
				pullNote := text2note(button[1])
				n2b[pushNote] = append(n2b[pushNote], Button{Side: side, Row: ir, Rank: jr, Pull: false})
				n2b[pullNote] = append(n2b[pullNote], Button{Side: side, Row: ir, Rank: jr, Pull: true})
				fmt.Println(ir, jr, button)
			}
		}
	}
	mapper(layout.Left, 0)
	mapper(layout.Right, 1)

	for k, l := range n2b {
		sort.Slice(l, func(i, j int) bool {
			return l[i].Rank < l[j].Rank
		})

		n2b[k] = l
	}

	return n2b
}

var config Config

func defaultButton(n int) Button {
	if b, ok := note2button[n]; ok {
		return b[0]
	}
	return Button{}
}
