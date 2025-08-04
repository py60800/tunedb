// goxml0 project main.go
package zixml

import (
	"encoding/xml"
	//	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"google.golang.org/api/iterator"
)

// Chords
var chordStr = [...]string{
	"major", //	Triad: major third, perfect fifth.
	"minor", //	Triad: minor third, perfect fifth.

	"augmented",          //	Triad: major third, augmented fifth.
	"augmented-seventh",  //	Seventh: augmented triad, minor seventh.
	"diminished",         //	Triad: minor third, diminished fifth.
	"diminished-seventh", //	Seventh: diminished triad, diminished seventh.
	"dominant",           //	Seventh: major triad, minor seventh.
	"dominant-11th",      //	11th: dominant-ninth, perfect 11th.
	"dominant-13th",      //	13th: dominant-11th, major 13th.
	"dominant-ninth",     //	Ninth: dominant, major ninth.
	"French",             //	Functional French sixth.
	"German",             //	Functional German sixth.
	"half-diminished",    //	Seventh: diminished triad, minor seventh.
	"Italian",            //	Functional Italian sixth.
	"major-11th",         //	11th: major-ninth, perfect 11th.
	"major-13th",         //	13th: major-11th, major 13th.
	"major-minor",        //	Seventh: minor triad, major seventh.
	"major-ninth",        //	Ninth: major-seventh, major ninth.
	"major-seventh",      //	Seventh: major triad, major seventh.
	"major-sixth",        //	Sixth: major triad, added sixth.
	"minor-11th",         //	11th: minor-ninth, perfect 11th.
	"minor-13th",         //	13th: minor-11th, major 13th.
	"minor-ninth",        //	Ninth: minor-seventh, major ninth.
	"minor-seventh",      //	Seventh: minor triad, minor seventh.
	"minor-sixth",        //	Sixth: minor triad, added sixth.
	"Neapolitan",         //	Functional Neapolitan sixth.
	"none",               //	Used to explicitly encode the absence of chords or functional harmony. In this case, the <root> <numeral>, or <function> element has no meaning. When using the <root> or <numeral> element, the <root-step> or <numeral-step> text attribute should be set to the empty string to keep the root or numeral from being displayed.
	"other",              //	Used when the harmony is entirely composed of add elements.
	"pedal",              //	Pedal-point bass
	"power",              //	Perfect fifth.
	"suspended-fourth",   //	Suspended: perfect fourth, perfect fifth.
	"suspended-second",   //	Suspended: major second, perfect fifth.
	"Tristan",            // Augmented fourth, augmented sixth, augmented ninth.

}

const MasterDivisions = 1200

func ChordKind(t string) int {
	for i, c := range chordStr {
		if t == c {
			return i
		}
	}
	return -1
}

type MPartition struct {
	Work           MWork           `xml:"work"`
	Identification MIdentification `xml:"identification"`
	Part           []MPart         `xml:"part"`
}

func (m MPartition) String() string {
	r := fmt.Sprintf("Work:%v,Id:%v\n", m.Work, m.Identification)
	for _, p := range m.Part {
		r += fmt.Sprintf("Part : \n")
		for _, m := range p.Measures {
			r += fmt.Sprintf("Id %v \n", m.Id)
			for _, c := range m.Contents {
				r += fmt.Sprintf("\t%v\n", c)
			}
		}
	}
	return r
}

type MWork struct {
	Title string `xml:"work-title"`
}
type MIdentification struct {
	Source string `xml:"source"`
	Misc   MMisc  `xml:"miscellaneous"`
}
type MMisc struct {
	XMLName xml.Name `xml:"miscellaneous"`
	Field   []MField `xml:"miscellaneous-field"`
}
type MField struct {
	FName string `xml:"name,attr"`
	Value string `xml:",innerxml"`
}
type MDivision struct {
	Value int
}

func (f MField) String() string {
	return fmt.Sprintf("<%v,%v>", f.FName, f.Value)
}

type MPart struct {
	Measures []MMeasure `xml:"measure"`
}
type MRoot struct {
	RootStep  string `xml:"root-step"`
	RootAlter int    `xml:"root-alter"`
}
type MHarmony struct {
	Root MRoot  `xml:"root"`
	Kind string `xml:"kind"`
}

func (p MPart) String() string {
	r := ""
	for _, m := range p.Measures {
		r += fmt.Sprintf("%v\n", m)
	}
	return r
}

type Mixed struct {
	Type string
	Elem any
}

func (dv *MDivision) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var err error
	t, _ := d.Token()
	dv.Value, err = strconv.Atoi(string(t.(xml.CharData)))
	d.Skip()
	return err
}
func (m *Mixed) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	m.Type = start.Name.Local
	err := fmt.Errorf("Unexpected element " + m.Type)

	switch start.Name.Local {
	case "note":
		var e MNote
		err = d.DecodeElement(&e, &start)
		m.Elem = e
	case "attributes":
		var e MAttributes
		err = d.DecodeElement(&e, &start)
		m.Elem = e
	case "key":
		var e MKey
		err = d.DecodeElement(&e, &start)
		m.Elem = e
	case "time":
		var e MTime
		err = d.DecodeElement(&e, &start)
		m.Elem = e

	case "barline":
		var e MBarline
		err = d.DecodeElement(&e, &start)
		m.Elem = e
	case "harmony":
		var e MHarmony
		err = d.DecodeElement(&e, &start)
		m.Elem = e
	case "print", "clef":
		m.Elem = nil
		d.Skip()
		return nil
	default:
		//fmt.Printf("unknown element!: %s\n", start)
		m.Elem = nil
		d.Skip()
		return nil
	}

	return err
}

type MMeasure struct {
	Id       int     `xml:"number,attr"`
	Contents []Mixed `xml:",any"`
}

type MAttributes struct {
	Contents []AMixed `xml:",any"`
}
type AMixed struct {
	Type string
	Elem any
}

func (m *AMixed) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	m.Type = start.Name.Local
	err := fmt.Errorf("Unexpected element " + m.Type)

	switch start.Name.Local {
	case "divisions":
		var e MDivision
		err = d.DecodeElement(&e, &start)
		m.Elem = e
	case "key":
		var e MKey
		err = d.DecodeElement(&e, &start)
		m.Elem = e
	case "time":
		var e MTime
		err = d.DecodeElement(&e, &start)
		m.Elem = e
	case "clef":
		m.Elem = nil
		d.Skip()
		return nil
	default:
		fmt.Printf("unknown element: %s\n", start)
		d.Skip()
		return nil
	}

	return err
}

type MEnding struct {
	Number int    `xml:"number,attr"`
	Type   string `xml:"type,attr"`
}
type MBarline struct {
	Repeat MDirection `xml:"repeat"`
	Ending MEnding    `xml:"ending"`
}
type MDirection struct {
	Direction string `xml:"direction,attr"`
}
type MKey struct {
	XMLName xml.Name `xml:"key"`
	Fifths  int      `xml:"fifths"`
	Mode    string   `xml:"mode"`
}
type MTime struct {
	Beats    int `xml:"beats"`
	BeatType int `xml:"beat-type"`
}

type MNote struct {
	//	Pitch    MPitch   `xml:"pitch"`
	Step     string   `xml:"pitch>step"`
	Octave   int      `xml:"pitch>octave"`
	Alter    int      `xml:"pitch>alter"`
	Rest     xml.Name `xml:"rest"`
	Duration int      `xml:"duration"`

	Type      string     `xml:"type"`
	Dot       xml.Name   `xml:"dot"`
	TimeMod   MTimeMod   `xml:"time-modification`
	Notations MNotations `xml:"notations"`
}

func (n *MNote) IsRest() bool {
	return n.Rest.Local == "rest"
}
func (n *MNote) IsGraceNote() bool {
	return n.Duration == 0
}

type MTimeMod struct {
	Actual int `xml:"actual-notes"`
	Normal int `xml:"normal-notes"`
}
type MNotations struct {
	Tuplet       MTuplet  `xml:"tuplet"`
	StrongAccent xml.Name `xml:"articulations>strong-accent"`
}
type MTuplet struct {
}

var baseNotes = map[rune]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}

func (p *MPartition) NormalizeDivisions(MasterDivisions int) {
	for ip, part := range p.Part {
		coeff := 0

		for im, measure := range part.Measures {
			for ic, item := range measure.Contents {
				switch v := item.Elem.(type) {
				case MAttributes:
					for ia, a := range v.Contents {
						switch d := a.Elem.(type) {
						case MDivision:
							nd := d.Value
							if nd > MasterDivisions {
								panic("XML Divisions too large")
							}
							coeff = MasterDivisions / nd
							if coeff*nd != MasterDivisions {
								panic("Odd Divisions")
							}
							fmt.Printf("Div:%d\n", p.Part[ip].Measures[im].Contents[ic].Elem.(MAttributes).Contents[ia].Elem.(MDivision).Value)
							d.Value = MasterDivisions
							p.Part[ip].Measures[im].Contents[ic].Elem.(MAttributes).Contents[ia].Elem = d
						}
					}
				case MNote:
					if coeff == 0 {
						panic("Division unset")
					}
					v.Duration *= coeff
					p.Part[ip].Measures[im].Contents[ic].Elem = v
					//fmt.Println("NM", v)
				}
			}

		}
	}

}

type PIterator struct {
	mp   *MPart
	i, j int
}

func (mp *MPart) CreateIterator() PIterator {
	return PIterator{mp, 0, 0}
}

func (p *PIterator) Done() error {
	if p.i >= len(p.mp.Measures) {
		return iterator.Done
	}
	return nil
}
func (p *PIterator) IsStartMeasure() bool {
	return p.j == 0 && p.i < len(p.mp.Measures)
}
func (p *PIterator) MeasureIndex() int {
	return p.i
}
func (p *PIterator) CurrentMeasure() *MMeasure {
	return &p.mp.Measures[p.i]
}

func (p *PIterator) Next() (*Mixed, error) {
	if p.i < len(p.mp.Measures) {
		r := &p.mp.Measures[p.i].Contents[p.j]
		p.j++
		if p.j >= len(p.mp.Measures[p.i].Contents) {
			p.j = 0
			p.i++
		}

		return r, nil
	}
	return nil, iterator.Done
}

func (pm *MPart) Attributes() (int, int, int) {
	var div int
	var beat int
	var beatType int

	iter := pm.CreateIterator()
	for {
		el, err := iter.Next()
		if err != nil {
			break
		}
		switch v := el.Elem.(type) {
		case MAttributes:
			for _, mx := range v.Contents {
				switch ve := mx.Elem.(type) {
				case MDivision:
					div = ve.Value
				case MTime:
					beat, beatType = ve.Beats, ve.BeatType
				}
			}

		}
	}
	return div, beat, beatType
}

func GetNoteList(fileName string) []int {
	notes := make([]int, 0)
	part, err := Parse(fileName)
	if err != nil {
		return notes
	}
	for _, m := range part.Part[0].Measures {
		for _, elem := range m.Contents {
			switch v := elem.Elem.(type) {
			case MNote:

				if !v.IsRest() {
					note := baseNotes[rune(v.Step[0])]
					note += 12 * v.Octave
					note += v.Alter
					notes = append(notes, note)
				}
			}
		}
	}
	return notes
}
func Parse(fileName string) (MPartition, error) {
	var partition MPartition
	xmlFile, err := os.Open(fileName)
	defer xmlFile.Close()

	if err != nil {
		fmt.Printf("Failed to read %v (%v)\n", fileName, err)
		return partition, err
	}
	byteValue, _ := ioutil.ReadAll(xmlFile)
	xml.Unmarshal(byteValue, &partition)
	partition.NormalizeDivisions(MasterDivisions)
	return partition, nil

}

func (m *MMeasure) Length() int {
	d := 0
	for _, el := range m.Contents {
		switch v := el.Elem.(type) {
		case MNote:
			d += v.Duration
		default:
			// ignore
		}
	}
	return d
}

func (p *MPart) ComputeMLength() (int, int, int) {
	return p.Measures[0].Length(), p.Measures[1].Length(), p.Measures[len(p.Measures)-1].Length()
}
