// View.go
package imgprint

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"

	//	"github.com/gotk3/gotk3/glib"
	"os/exec"
	"strings"

	"os"

	"github.com/gotk3/gotk3/gtk"
)

type Printer struct {
	// ********
	win    *gtk.Window
	menu   *gtk.MenuBar
	label  *gtk.Entry
	header *gtk.Entry

	Image  *gtk.Image
	PixBuf *gdk.Pixbuf

	ImgList []string

	printSettings  *gtk.PrintSettings
	pageSettings   *gtk.PageSetup
	defaultPrinter string
}

func (p *Printer) setDefaultPrinter(ps string) {
	p.defaultPrinter = ps
	p.label.SetText(ps)

}
func Message(msg string) {
	f := gtk.DialogFlags(gtk.DIALOG_DESTROY_WITH_PARENT)
	d := gtk.MessageDialogNew(nil, f, gtk.MESSAGE_INFO, gtk.BUTTONS_CLOSE, msg)
	d.Run()
	d.Destroy()
}

func (p *Printer) draw(cr *cairo.Context, w, h float64) {

	hdr, _ := p.header.GetText()
	if len(hdr) > 0 {
		log.Println("ImgPrint:" + hdr)
	}
	// Preserve ratio
	W, H := float64(p.PixBuf.GetWidth()), float64(p.PixBuf.GetHeight())
	r := min(w/W, h/H)

	wt := (r * W)
	ht := (r * H)

	pb, _ := p.PixBuf.ScaleSimple(int(wt), int(ht), gdk.INTERP_BILINEAR)

	surf, err := gdk.CairoSurfaceCreateFromPixbuf(pb, 1, nil)
	if err != nil {
		log.Println("ImgPrint:" + fmt.Sprint(err))
	}
	cr.SetSourceSurface(surf, (w-wt)/2, 0)
	//
	cr.Rectangle((w-wt)/2, 0, wt, ht)
	cr.Fill()

}

func (p *Printer) Draw(op *gtk.PrintOperation, pc *gtk.PrintContext, pn int) {

	w, h := pc.GetWidth(), pc.GetHeight()
	log.Printf("Draw Page %v %v/%v DPI:%v/%v PixBuf: %v/%v\n", pn, w, h, pc.GetDpiX(), pc.GetDpiY(), p.PixBuf.GetWidth(), p.PixBuf.GetHeight())
	cr := pc.GetCairoContext()

	p.draw(cr, w, h)

}
func (p *Printer) Print() {
	po, _ := gtk.PrintOperationNew()

	p.printSettings.SetPrintPages(gtk.PRINT_PAGES_ALL)
	//	p.printSettings.SetResolutionXY(300, 300)
	p.printSettings.SetResolution(300)
	p.printSettings.SetPrinterLpi(150)
	p.printSettings.SetQuality(gtk.PRINT_QUALITY_HIGH)
	p.printSettings.GetPageSet(gtk.PRINT_PAGES_ALL)
	po.SetDefaultPageSetup(p.pageSettings)
	po.SetPrintSettings(p.printSettings)

	po.SetNPages(1)
	po.Connect("begin_print", func(op *gtk.PrintOperation, pc *gtk.PrintContext) {
		log.Println("ImgPrint:Start printing")
	})
	po.Connect("draw_page", func(op *gtk.PrintOperation, pc *gtk.PrintContext, pn int) {
		log.Println("ImgPrint:Print Page")
		p.Draw(op, pc, pn)
	})
	po.Connect("end_print", func(op *gtk.PrintOperation, pc *gtk.PrintContext) {
		p.Quit()
	})

	res, err := po.Run(gtk.PRINT_OPERATION_ACTION_PRINT_DIALOG, p.win)

	p.printSettings, _ = po.GetPrintSettings(p.pageSettings)
	p.printSettings.ToFile("printSettings.txt")
	p.setDefaultPrinter(p.printSettings.GetPrinter())

	log.Println("ImgPrint:", res, err)
}
func (p *Printer) Quit() {
	p.win.Hide()
}
func (p *Printer) toA4() *gdk.Pixbuf {
	W := 21 * 300 / 2.54
	H := 29.7 * 300 / 2.54
	log.Println("ImgPrint:Img file Sz:", W, H)
	w := float64(p.PixBuf.GetWidth())
	h := float64(p.PixBuf.GetHeight())
	r := min(W/w, H/h)

	res, _ := gdk.PixbufNew(gdk.COLORSPACE_RGB, false, 8, int(W), int(H))
	res.Fill(0xffffffff)
	p.PixBuf.Composite(res, 0, 0, int(min(W, r*w)), int(min(H, h*w)),
		0, 0, r, r,
		gdk.INTERP_BILINEAR, 255)
	return res
}

func (p *Printer) HRPrint() {
	pCmd := "img2pdf  --pagesize A4 -b 10:10 - | tee spy.pdf | lpr"
	if p.defaultPrinter != "" {
		pCmd += " -P " + p.defaultPrinter
	}
	cmd := exec.Command("bash", "-c", pCmd)
	writer, _ := cmd.StdinPipe()
	var out strings.Builder
	cmd.Stdout = &out
	var stderr strings.Builder
	cmd.Stderr = &stderr
	log.Println("ImgPrint:Start Printing")
	err := cmd.Start()
	if err != nil {
		log.Println("ImgPrint:", err)
		return
	} else {
		log.Println("ImgPrint:Start success")
	}

	pb := p.toA4()
	errw := pb.WriteJPEG(writer, 100)
	writer.Close()
	log.Println("ImgPrint:Writing:", errw)

	cmd.Wait()
	log.Println("ImgPrint:Print Result:", out.String(), stderr.String())
}
func (p *Printer) PrintSettings() {
	po, _ := gtk.PrintOperationNew()

	po.SetNPages(0)

	po.Run(gtk.PRINT_OPERATION_ACTION_PRINT_DIALOG, p.win)

	p.printSettings, _ = po.GetPrintSettings(p.pageSettings)
	p.printSettings.ToFile("printSettings.txt")
	p.setDefaultPrinter(p.printSettings.GetPrinter())

}

func (p *Printer) ToFile() {
	p.toA4().SaveJPEG("img.jpg", 100)
}
func (p *Printer) SetImg(imgs []string) {

	if len(p.ImgList) > 0 {
		for _, f := range p.ImgList {
			os.Remove(f)
		}
		p.ImgList = make([]string, 0)
	}

	pixBuf := make([]*gdk.Pixbuf, len(imgs))
	w, h := 0, 0
	for i, file := range imgs {

		f := svgPatch(file)
		if f == "" {
			continue
		}
		p.ImgList = append(p.ImgList, f)

		var err error
		pixBuf[i], err = gdk.PixbufNewFromFile(f)
		if err != nil {
			log.Println("ImgPrint:", f, err)
		}
		log.Printf("File: %v Temp:%v W:%v H:%v \n", file, f, pixBuf[i].GetWidth(), pixBuf[i].GetHeight())
		w = max(w, pixBuf[i].GetWidth())
		h += pixBuf[i].GetHeight()
	}
	log.Println("ImgPrint:R Size:", w, h)
	all, _ := gdk.PixbufNew(pixBuf[0].GetColorspace(), pixBuf[0].GetHasAlpha(),
		pixBuf[0].GetBitsPerSample(), w, h)
	y := 0
	for _, p := range pixBuf {
		log.Println("ImgPrint:All:", all.GetWidth(), all.GetHeight(), "Part", p.GetWidth(), p.GetHeight(), "Y:", y)
		p.Composite(all, 0, y, p.GetWidth(), p.GetHeight(), 0, float64(y), 1.0, 1.0, gdk.INTERP_TILES, 255)
		y += p.GetHeight()
	}
	p.PixBuf = all
	p.Image.QueueDraw()
}

func (p *Printer) MakeUI() {

	// Main window ---------------------------
	p.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	p.win.SetSizeRequest(800, 600)

	p.header, _ = gtk.EntryNew()

	firstLine, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	grid, _ := gtk.GridNew()
	p.win.Add(grid)
	p.win.Connect("destroy", func() {
		p.Quit()
	})
	grid.Attach(firstLine, 0, 0, 12, 1)

	print, _ := gtk.ButtonNewWithLabel("Print...")
	print.Connect("clicked", func() {
		p.Print()
	})
	firstLine.Add(print)

	p.label, _ = gtk.EntryNew()
	pr := p.printSettings.GetPrinter()
	p.label.SetText(pr)
	firstLine.Add(p.label)

	cancel, _ := gtk.ButtonNewWithLabel("Cancel")
	cancel.Connect("clicked", func() {
		p.Quit()
	})
	firstLine.Add(cancel)
	toFile, _ := gtk.ButtonNewWithLabel("ToFile")
	toFile.Connect("clicked", func() {
		p.ToFile()
		p.win.Hide()
	})
	firstLine.Add(toFile)
	printSet, _ := gtk.ButtonNewWithLabel("Print Settings..")
	printSet.Connect("clicked", func() {
		p.PrintSettings()
	})
	firstLine.Add(printSet)

	hrPrint, _ := gtk.ButtonNewWithLabel("HRPrint")
	hrPrint.Connect("clicked", func() {
		p.HRPrint()
		p.win.Hide()
	})
	firstLine.Add(hrPrint)

	p.Image, _ = gtk.ImageNew()
	p.Image.SetHExpand(true)
	p.Image.SetVExpand(true)
	grid.Attach(p.Image, 0, 1, 12, 11)
	p.Image.Connect("draw", func(img *gtk.Image, cr *cairo.Context) bool {
		w, h := img.GetAllocatedWidth(), img.GetAllocatedHeight()
		p.draw(cr, float64(w), float64(h))
		return true
	})

	grid.ShowAll()
	p.win.Maximize()
}
func PrinterNew() *Printer {
	var p Printer
	p.pageSettings, _ = gtk.PageSetupNew()
	p.printSettings, _ = gtk.PrintSettingsNew()
	p.printSettings.LoadFile("printSettings.txt")

	p.MakeUI()

	log.Println("ImgPrint:" + p.printSettings.GetDefaultSource())
	log.Println("ImgPrint:" + p.printSettings.GetPrinter())
	p.pageSettings.PageSetupToFile("pageSetup.txt")

	return &p
}
func (p *Printer) Run(img []string) {
	p.SetImg(img)
	p.win.ShowNow()
}
