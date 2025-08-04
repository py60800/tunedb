package svgtab

import (
	"fmt"

	svgutil "github.com/hookttg/svgparser/utils"
	xml "github.com/subchen/go-xmldom"

	_ "embed"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/py60800/tunedb/internal/util"
)

//go:embed concertina.yml
var defaultConfig string

func convert(m []float64, x, y float64) (xn float64, yn float64) {
	return m[0]*x + m[2]*y + m[4], m[1]*x + m[3]*y + m[5]
}

var charHeight = 35.0

func ReadXml(file string) *xml.Document {
	doc, err := xml.ParseFile(file)
	if err != nil {
		panic(err)
	}
	return doc
}
func WriteXml(file string, doc *xml.Document) {
	os.WriteFile(file, []byte(doc.XMLPretty()), 0666)
}

type Dot struct {
	Staff     int
	X, Y      float64
	Highlight bool
}

func origin(ps string) (float64, float64) {
	data, _ := svgutil.PathParser(ps)
	for _, sp := range data.Subpaths {
		for _, c := range sp.Commands {
			switch c.Symbol {
			case "M":
				return c.Params[0], c.Params[1]
			}
		}
	}
	return 0, 0

}

type Staff struct {
	x0, x1 float64
	y      float64
	Y      float64
	W      float64
	top    float64
	bottom float64
}

func f2t(f float64) string {
	return fmt.Sprintf("%4.3f", f)
}

type SvgTab struct {
	doc       *xml.Document
	TempFile  string
	W, H      float64
	Staves    []Staff
	Dots      []Dot
	Notes     []int
	Buttons   []Button
	Idx       []int
	selection []int
}

func (s *SvgTab) SetNotes(Notes []int, buttons []Button) {
	s.Notes = Notes
	if len(buttons) != len(Notes) {
		s.Buttons = make([]Button, len(Notes))
		for i, n := range Notes {
			s.Buttons[i] = defaultButton(n)
		}
	} else {
		s.Buttons = buttons
	}
	s.Idx = make([]int, len(Notes))
	for i, stv := range s.Staves {
		s.Staves[i].top = stv.Y - stv.W
		s.Staves[i].bottom = stv.Y + stv.W*1.5
	}
}
func (s *SvgTab) GetButtons() []Button {
	return s.Buttons
}
func dist(x, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}
func (s *SvgTab) Click(X, Y, W, H float64) {
	fmt.Println("Click:", X, Y, W, H)
	x, y := X/W, Y/H
	ib := -1
	dmax := 1.0
	for i, p := range s.Dots {
		if y > s.Staves[p.Staff].top && y < s.Staves[p.Staff].bottom {

			d := math.Sqrt((x - p.X) * (x - p.X))
			if d < dmax {
				dmax = d
				ib = i
			}
		}
	}
	if ib < 0 || dmax > 0.1 {
		fmt.Println("Not found:", ib, x, y, dmax)
		return
	} else {
		fmt.Println("Found ", ib, s.Dots[ib], dmax)
	}
	btns := note2button[s.Notes[ib]]
	s.Idx[ib] = (s.Idx[ib] + 1) % len(btns)
	s.Buttons[ib] = btns[s.Idx[ib]]
	s.highLight()
}

var configInit bool = false

func fixHeight(doc *xml.Document, delta float64) (float64, float64) {
	svg := doc.Root
	_X := strings.TrimSuffix(svg.GetAttributeValue("width"), "px")
	_Y := strings.TrimSuffix(svg.GetAttributeValue("height"), "px")
	var H, W float64
	vb := svg.GetAttributeValue("viewBox")
	var x0, y0, w, h float64
	if n, err := fmt.Sscanf(vb, "%f %f %f %f", &x0, &y0, &w, &h); n == 4 && err == nil {
		W = w
		H = h
	} else {
		W, _ = strconv.ParseFloat(_X, 64)
		H, _ = strconv.ParseFloat(_Y, 64)
	}
	H = H + delta
	return H, W
}

func SvgTabNew(context string, file string) *SvgTab {
	if !configInit {
		config = util.ReadConfig[Config](path.Join(context, "concertina.yml"), defaultConfig)
		note2button = configAnalysis(&config)
		configInit = true
	}
	svgt := SvgTab{}

	dots := make([]Dot, 0, 150)
	svgt.doc = ReadXml(file)
	if svgt.doc == nil {
		fmt.Println("SvgTab Can't read :", file)
		return nil
	}
	svg := svgt.doc.Root
	H, W := fixHeight(svgt.doc, charHeight*3.0)
	svgt.H = H
	svgt.W = W
	svg.RemoveAttribute("height")
	svg.RemoveAttribute("width")
	// Extend view for tabs
	svg.SetAttributeValue("height", f2t(svgt.H)+"px")
	svg.SetAttributeValue("width", f2t(svgt.W)+"px")
	svg.RemoveAttribute("viewBox")
	//	fmt.Println("Size:", _X, _Y, vb, svg.GetAttributeValue("height"), svg.GetAttributeValue("width"))
	// Locate staves
	Staves := make([]Staff, 0)
	staffWidth := 0.0
	{
		staves := svg.Query("/polyline[@class='StaffLines']")
		sum := 0.0
		var y0 float64
		for i, s := range staves {
			if i%5 == 0 {
				sum = 0
			}
			var x0, x1, y float64
			if pos := s.GetAttribute("points"); pos != nil {
				pts := strings.Split(pos.Value, " ")
				fmt.Sscanf(pts[0], "%f,%f", &x0, &y)
				sum += y
				if i%5 == 0 {
					y0 = y
				}
				fmt.Sscanf(pts[1], "%f", &x1)
			}
			if i%5 == 4 {
				fmt.Println(y, x0, x1)
				Staves = append(Staves, Staff{y: sum / 5.0, x0: x0, x1: x1,
					Y: sum / (5.0 * svgt.H), W: (y - y0) / svgt.H})
				if staffWidth == 0.0 {
					staffWidth = y - y0
				}
			}
		}
	}
	svgt.Staves = Staves

	// Search notes
	notes := svg.Query("/path[@class='Note']")
	for _, n := range notes {
		var x, y float64
		path := "NoPath"
		transform := "No Transform"
		m := make([]float64, 0)

		if p := n.GetAttribute("d"); p != nil {
			path = p.Value
			x, y = origin(path)
			path = path[:min(len(path), 10)]
		}

		if t := n.GetAttribute("transform"); t != nil {
			transform = strings.TrimSpace(t.Value)
			if strings.HasPrefix(transform, "matrix") {
				content := transform[strings.Index(transform, "(")+1 : strings.Index(transform, ")")]
				d := strings.Split(content, ",")
				for _, s := range d {
					f, _ := strconv.ParseFloat(s, 64)
					m = append(m, f)
				}
			}
			xm, ym := convert(m, x, y)
			is, mx := 0, svgt.H
			for i, sp := range Staves {
				d := math.Abs(ym - sp.y)
				if d < mx {
					mx = d
					is = i
				}
			}
			dots = append(dots, Dot{is, xm / svgt.W, ym / svgt.H, false})
		}
	}

	sort.Slice(dots, func(i, j int) bool {
		switch {
		case dots[i].Staff < dots[j].Staff:
			return true
		case dots[i].Staff > dots[j].Staff:
			return false
		default:
			return dots[i].X < dots[j].X

		}
	})
	svgt.Dots = dots

	remainder := make([]*xml.Node, 0)
	for _, s := range svg.Children {
		if attr := s.GetAttribute("class"); attr == nil || attr.Value != "Lyrics" {
			remainder = append(remainder, s)
		}
	}
	svg.Children = remainder

	baseFile := path.Base(file)
	svgt.TempFile = path.Join("tmp", baseFile)
	// Create temporary file
	WriteXml(svgt.TempFile, svgt.doc)
	svgt.highLight()
	return &svgt
}
func (s *SvgTab) fMap(ft func(i int, dts []Button)) {
	for _, i := range s.selection {
		if dts, ok := note2button[s.Notes[i]]; ok {
			ft(i, dts)
		}
	}

}
func (s *SvgTab) FirstFinger() {
	s.fMap(func(i int, dts []Button) {
		s.Buttons[i] = dts[0]
		s.Idx[i] = 0
	})
	s.highLight()
}
func (s *SvgTab) FirstRow() {
	s.fMap(func(i int, dts []Button) {
		j := 0
		for j = range dts {
			if dts[j].Row == 0 {
				break
			}
		}
		s.Buttons[i] = dts[j]
		s.Idx[i] = j
	})
	s.highLight()
}
func (s *SvgTab) SecondRow() {
	s.fMap(func(i int, dts []Button) {
		j := 0
		for j = range dts {
			if dts[j].Row == 1 {
				break
			}
		}
		s.Buttons[i] = dts[j]
		s.Idx[i] = j
	})
	s.highLight()
}
func (s *SvgTab) highLight() {
	for i := range s.Dots {
		s.Dots[i].Highlight = false
	}
	for i := 1; i < len(s.Buttons); i++ {
		if s.Buttons[i].Side == s.Buttons[i-1].Side &&
			s.Buttons[i].Rank == s.Buttons[i-1].Rank &&
			s.Buttons[i].Row != s.Buttons[i-1].Row {
			s.Dots[i].Highlight = true
			s.Dots[i-1].Highlight = true
		}
	}

}
func (s *SvgTab) MsczUpdate(mscz string) {
	MsczUpdate(mscz, s.Buttons)
}
func (s *SvgTab) MsczCleanUp(mscz string) {
	MsczCleanUp(mscz)
}

// *********************************************************

func SvgEnhance(file string) {
	fmt.Println("SvgEnhance", file)
	doc := ReadXml(file)
	if doc == nil {
		return
	}
	svg := doc.Root
	/*
	   _H := strings.TrimSuffix(svg.GetAttributeValue("height"), "px")
	   H, _ := strconv.ParseFloat(_H, 64)
	   H = H + 10.0
	   svg.SetAttributeValue("height", f2t(H)+"px")
	   svg.RemoveAttribute("viewBox")
	*/
	fixHeight(doc, 10)
	underlines := svg.Query("/polyline[@class='Lyrics']")
	for _, ln := range underlines {
		sw := ln.GetAttributeValue("stroke-width")
		v, _ := strconv.ParseFloat(sw, 64)
		ln.SetAttributeValue("stroke-width", f2t(v*2.5))
	}
	if len(underlines) > 0 {
		WriteXml(file, doc)
	}
}
