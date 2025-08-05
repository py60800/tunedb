package imgprint

import (
	"fmt"

	"os"
	"strings"

	xml "github.com/subchen/go-xmldom"
)

func ReadXml(file string) *xml.Document {
	doc, err := xml.ParseFile(file)
	if err != nil {
		fmt.Println("Read Xml:", err)
		return nil
	}
	return doc
}
func WriteXml(file string, doc *xml.Document) {
	os.WriteFile(file, []byte(doc.XMLPretty()), 0666)
}

func f2t(f float64) string {
	return fmt.Sprintf("%4.3f", f)
}

func svgPatch(file string) string {
	dest, err := os.CreateTemp("tmp", "tmp.*.svg")
	if err != nil {
		return ""
	}
	doc := ReadXml(file)
	if doc == nil {
		return ""
	}
	svg := doc.Root

	Hs := svg.GetAttributeValue("height")
	//	Ws := svg.GetAttributeValue("width")
	if !strings.HasSuffix(Hs, "px") {
		vb := svg.GetAttributeValue("viewBox")
		var x0, y0, w, h float64
		if n, err := fmt.Sscanf(vb, "%f %f %f %f", &x0, &y0, &w, &h); n == 4 && err == nil {
			svg.RemoveAttribute("height")
			svg.RemoveAttribute("width")
			svg.SetAttributeValue("height", f2t(h+5)+"px")
			svg.SetAttributeValue("width", f2t(w)+"px")
			svg.RemoveAttribute("viewBox")
			svg.SetAttributeValue("viewBox", fmt.Sprintf("%2.1f %2.1f %2.1f %2.1f", x0, y0, w, h+4))
			fmt.Println("Img Patched")
		}

	}
	dest.Write([]byte(doc.XMLPretty()))
	dest.Close()
	return dest.Name()
}
