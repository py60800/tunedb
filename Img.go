package main

import (
	//"fmt"

	"github.com/py60800/tunedb/internal/util"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type TImage struct {
	previousH    int
	previousW    int
	previousFile string
	PixBuf       *gdk.Pixbuf
	Image        *gtk.DrawingArea
	WPixBuf      int
	HPixBuf      int

	X0, X1, Y0, Y1 float64
	selectAll      bool
	ButtonPress    bool
}

func (c *TImage) Refresh() {
	c.Image.QueueDraw()
}
func (c *TImage) ResetSelection() {
	c.X0, c.X1, c.Y0, c.Y1 = 0., 0., 0., 0.
	c.selectAll = false
}
func (c *TImage) SelectAll() {
	c.selectAll = !c.selectAll
	c.Image.QueueDraw()
}

func (c *TImage) Draw(img *gtk.DrawingArea, cr *cairo.Context) {
	Hd := img.GetAllocatedHeight()
	Wd := img.GetAllocatedWidth()
	tune := ActiveTune()
	if tune == nil || tune.Img == "" {
		cr.SetSourceRGB(255, 0, 255)
		cr.Rectangle(0, 0, float64(Wd), float64(Hd))
		cr.Fill()
		c.previousH, c.previousW = 0, 0
		return
	}

	imgFile := tune.Img
	ctx := GetContext()
	if ctx.svgt != nil {
		if _, ok := util.GetModificationDate(ctx.svgt.TempFile); ok {
			imgFile = ctx.svgt.TempFile
		}
	}

	var err error
	c.PixBuf, err = gdk.PixbufNewFromFileAtScale(imgFile, Wd, Hd, true)
	util.WarnOnError(err)
	c.WPixBuf = c.PixBuf.GetWidth()
	c.HPixBuf = c.PixBuf.GetHeight()
	c.previousFile = imgFile
	c.previousH, c.previousW = Hd, Wd
	if err == nil {
		gtk.GdkCairoSetSourcePixBuf(cr, c.PixBuf, 0.0, 0.0)
		cr.Paint()
	} else {
		util.WarnOnError(err)
	}

	if c.ButtonPress {
		cr.SetSourceRGBA(0, 0, 255, 50)
		cr.Rectangle(c.X0, c.Y0, c.X1-c.X0, c.Y1-c.Y0)
		cr.SetLineWidth(2)
		cr.Stroke()
	}
	if ctx.svgt != nil {
		if ctx.Image.selectAll {
			ctx.svgt.ViewTab(img, cr, float64(c.WPixBuf), float64(c.HPixBuf),
				0, float64(Wd), 0, float64(Hd))
		} else {
			ctx.svgt.ViewTab(img, cr, float64(c.WPixBuf), float64(c.HPixBuf),
				c.X0, c.X1, c.Y0, c.Y1)
		}
	}
}
func (c *TImage) SetSizeRequest(w, l int) {
	c.Image.SetSizeRequest(w, l)
}
func (c *TImage) ForceUpdate() {
	c.previousFile = ""
}

func (c *ZContext) MkImage() (*TImage, *gtk.Widget) {
	imgCtrl := &TImage{}
	imgCtrl.Image, _ = gtk.DrawingAreaNew()
	imgCtrl.Image.SetSizeRequest(800, 800)

	imgCtrl.Image.SetHExpand(true)
	imgCtrl.Image.SetVExpand(true)
	imgCtrl.Image.AddEvents(int(gdk.BUTTON_PRESS_MASK | gdk.BUTTON_MOTION_MASK | gdk.BUTTON_RELEASE_MASK))
	imgCtrl.Image.Connect("button-press-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
		evb := gdk.EventButtonNewFromEvent(ev)
		imgCtrl.ButtonPress = true
		imgCtrl.X0 = evb.X()
		imgCtrl.X1 = imgCtrl.X0
		imgCtrl.Y0 = evb.Y()
		imgCtrl.Y1 = imgCtrl.Y0

		/*if evb.Type() == 4 {
			w := da.GetAllocatedWidth()
			Message(fmt.Sprintf("Click:", evb.X(), evb.Y(), w))
			return true
		}*/
		if c.svgt != nil {
			c.svgt.Click(evb.X(), evb.Y(), float64(imgCtrl.WPixBuf), float64(imgCtrl.HPixBuf))
		}
		return false
	})
	imgCtrl.Image.Connect("button-release-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
		//evb := gdk.EventButtonNewFromEvent(ev)
		/*if evb.Type() == 4 {
			w := da.GetAllocatedWidth()
			Message(fmt.Sprintf("Click:", evb.X(), evb.Y(), w))
			return true
		}*/
		imgCtrl.ButtonPress = false
		da.QueueDraw()

		return false
	})
	imgCtrl.Image.Connect("motion-notify-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
		evb := gdk.EventButtonNewFromEvent(ev)
		X, Y := evb.X(), evb.Y()
		imgCtrl.X0 = min(imgCtrl.X0, X)
		imgCtrl.X1 = max(imgCtrl.X1, X)
		imgCtrl.Y0 = min(imgCtrl.Y0, Y)
		imgCtrl.Y1 = max(imgCtrl.Y1, Y)
		imgCtrl.selectAll = false
		da.QueueDraw()
		return false
	})

	imgCtrl.Image.Connect("draw", func(img *gtk.DrawingArea, cr *cairo.Context) {
		imgCtrl.Draw(img, cr)
	})
	return imgCtrl, imgCtrl.Image.ToWidget()
}
