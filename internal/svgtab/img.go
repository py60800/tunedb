package svgtab

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

func (s *SvgTab) ViewTab(img *gtk.DrawingArea, cr *cairo.Context, W, H float64,
	x0, x1, y0, y1 float64) {
	s.selection = make([]int, 0)
	for i, p := range s.Dots {
		y := s.Staves[p.Staff].bottom
		switch {
		case p.X*W >= x0 && p.X*W <= x1 && p.Y*H >= y0 && p.Y*H <= y1:
			cr.SetSourceRGB(255, 0, 0)
			s.selection = append(s.selection, i)
		case p.Highlight:
			cr.SetSourceRGB(0, 255, 0)
		default:
			cr.SetSourceRGB(0, 0, 0)
		}
		cr.SelectFontFace("Times", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
		cr.SetFontSize(18)
		cr.MoveTo(p.X*W, y*H)

		txt, pull := s.Buttons[i].Text()
		txtEx := cr.TextExtents(txt)
		cr.ShowText(txt)

		if pull {

			cr.MoveTo(p.X*W, y*H+4)
			cr.LineTo(p.X*W+txtEx.Width, y*H+4)
			cr.SetLineWidth(3)
			cr.Stroke()
		}

	}

}
