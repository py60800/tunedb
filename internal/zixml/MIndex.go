// Index
package zixml

import (
	"fmt"
)

func INote(n *MNote) int {
	return MidiKeyInt(n.Step, n.Octave) + n.Alter
}
func MIndex(m *MMeasure, Divisor int, FirstNote int) (int, []int) {
	r := make([]int, Divisor)
	tick := 0
	length := m.Length()
	var note int
	tn := make([]int, length)
	for _, n := range m.Contents {
		switch v := n.Elem.(type) {

		case MNote:
			if v.Duration == 0 {
				continue
			}
			if v.Rest.Local != "rest" {
				note = INote(&v)
				if FirstNote == 0 {
					FirstNote = note
				}
			}
			for i := 0; i < v.Duration; i++ {
				tn[tick+i] = note
			}
			tick += v.Duration
		}
	}
	for i := range r {
		r[i] = tn[i*length/Divisor] - FirstNote
	}
	if tick != len(tn) {
		fmt.Println("Internal Error Mindex")
		return 0, r
	}
	return FirstNote, r

}
func ComputeIndex(p *MPartition) string {
	if p == nil || len(p.Part) < 1 {
		return ""
	}
	partition := p.Part[0]
	if len(partition.Measures) < 4 {
		return "" //[]int{}
	}
	iStart := 0
	l0 := partition.Measures[0].Length()
	l1 := partition.Measures[1].Length()
	if l0 < l1 {
		iStart = 1
	}

	//	TimeDiv, Beats, BeatType := partition.Attributes()
	_, Beats, BeatType := partition.Attributes()
	Divisor := 4
	switch {
	case Beats == 4 && BeatType == 4:
		Divisor = 4
	case Beats == 2 && BeatType == 2:
		Divisor = 4
	case Beats == 2 && BeatType == 4:
		Divisor = 4
	case Beats == 6 && BeatType == 8:
		Divisor = 2
	case Beats == 3 && BeatType == 4:
		Divisor = 3
	case Beats == 9 && BeatType == 8:
		Divisor = 3
	case Beats == 12 && BeatType == 8:
		Divisor = 4
	default:
		fmt.Printf("Unhandled time sig %d/%d", Beats, Beats)
		Divisor = Beats
	}
	FN, m0 := MIndex(&partition.Measures[iStart], Divisor, 0)
	_, m1 := MIndex(&partition.Measures[iStart+1], Divisor, FN)
	res := append(m0, m1...)
	// encode
	code := fmt.Sprint(Divisor)
	//code := ""
	for _, i := range res {
		var c rune
		switch {
		case i == 0:
			c = '-'
		case i > 0:
			c = rune(int('A') + i)
		case i < 0:
			c = rune(int('a') - i)
		}
		code += string(c)
	}
	return code
}
func ComputeIndexForFile(file string) string {
	part, _ := Parse(file)
	return ComputeIndex(&part)
}
