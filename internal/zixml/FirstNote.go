// FirstNote
package zixml
import (
"github.com/py60800/tunedb/internal/util"
)
func GetFirstNote(file string) (first string, mode string, fifth int, BIndex string) {
	part, err := Parse(file)
	if err != nil || len(part.Part) == 0 || len(part.Part[0].Measures) < 2 {
		util.WarnOnError(err)
		return "", "none", 0, ""
	}
	p := part.Part[0]

	type TNote struct {
		ml   int
		note *MNote
	}
	ml := make([]TNote, 2)
	Fifths := -1
	Mode := "none"
	for i := 0; i < 2; i++ {

		m := &p.Measures[i]
		ml[i].ml = 0
		ml[i].note = nil
		for _, el := range m.Contents {
			switch v := el.Elem.(type) {
			case MAttributes:
				for _, e := range v.Contents {
					switch vi := e.Elem.(type) {
					case MKey:
						Fifths = vi.Fifths
						Mode = vi.Mode
					}
				}
			case MNote:
				ml[i].ml += int(v.Duration)
				if ml[i].note == nil {
					ml[i].note = &v
				}
			}
		}
	}
	note := ml[0].note
	if ml[0].ml < ml[1].ml {
		// truncated first Meas
		note = ml[1].note
	}
	aux := ""
	switch note.Alter {
	case 1:
		aux = "#"
	case -1:
		aux = "b"
	}
	fn := note.Step + aux
	return fn, Mode, Fifths, ComputeIndex(&part)
}
