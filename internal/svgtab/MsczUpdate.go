package svgtab

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"log"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
	"github.com/py60800/tunedb/internal/util"
	xml "github.com/subchen/go-xmldom"
)

func CopyFile(src, dst string) error {
	srcfile, err := os.Open(src)
	if err != nil {
		util.WarnOnError(err)
		return err
	}
	defer srcfile.Close()

	info, err := srcfile.Stat()
	if err != nil {
		util.WarnOnError(err)
		return err
	}
	if info.IsDir() {
		log.Println("Copy dir attempt")
		return errors.New("cannot read from directories")
	}

	// create new file, truncate if exists and apply same permissions as the original one
	dstfile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer dstfile.Close()

	_, err = io.Copy(dstfile, srcfile)
	return err
}
func MsczUpdate(file string, buttons []Button) (err error) {
	defer func() {
		log.Println("MsczUpdate error:", err)
	}()
	// Save file
	dir := path.Dir(file)
	archive := path.Join(dir, "archive")
	os.MkdirAll(archive, 0777)
	ext := time.Now().Format("2006-01-02,15:04:05")
	archiveFile := path.Join(archive, path.Base(file)+ext)
	if err := CopyFile(file, archiveFile); err != nil {
		return err
	}

	r, err := zip.OpenReader(archiveFile)
	if err != nil {
		return err
	}
	defer r.Close()

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, f := range r.File {
		rc, er0 := f.Open()
		if er0 != nil {
			return er0
		}
		defer rc.Close()
		data, er1 := ioutil.ReadAll(rc)
		if er1 != nil {
			return er1
		}

		if strings.HasSuffix(f.Name, ".mscx") {
			// Patch
			doc, er := xml.ParseXML(string(data))
			if er != nil {
				return er
			}
			notes := doc.Root.Query("//Chord")
			if len(notes) != len(buttons) {
				return errors.New(fmt.Sprintf("Inconsistant data: %v,%v", len(notes), len(buttons)))
			}

			for i := range notes {
				// Create text node
				textNode := &xml.Node{Name: "text"}
				inner := textNode
				if buttons[i].Side == 1 {
					inner = inner.CreateNode("b")
				}
				if buttons[i].Pull {
					inner = inner.CreateNode("u")
				}
				inner.Text, _ = buttons[i].Text()

				cNote := notes[i]

				if l := cNote.FindOneByName("Lyrics"); l != nil {
					t := l.FindOneByName("text")
					l.RemoveChild(t)
					l.AppendChild(textNode)
				} else {
					lyrics := &xml.Node{Name: "Lyrics"}
					lyrics.AppendChild(textNode)
					notes[i].AppendChild(lyrics)
					length := len(cNote.Children)
					for i, n := range cNote.Children {
						if n.Name == "durationType" {
							i++
							for j := length - 1; j > i; j-- {
								cNote.Children[j] = cNote.Children[j-1]
							}
							cNote.Children[i] = lyrics
						}
					}
				}
			}
			data = []byte(doc.XMLPretty())
		}
		to, er2 := w.Create(f.Name)
		if er2 != nil {
			return er2
		}
		_, er3 := to.Write(data)
		if err != nil {
			return er3
		}
		// to.Close()

	}
	w.Close()
	//	dest := path.Join(dir, "tmp", path.Base(file))

	target, _ := os.Create(file)
	_, erf := buf.WriteTo(target)
	target.Close()

	return erf
}
func MsczBackup(file string) (string, error) {
	dir := path.Dir(file)
	archive := path.Join(dir, "archive")
	os.MkdirAll(archive, 0777)
	ext := time.Now().Format("2006-01-02,15:04:05")
	archiveFile := path.Join(archive, path.Base(file)+ext)
	return archiveFile, CopyFile(file, archiveFile)
}
func MsczCleanUp(file string) (err error) {
	defer func() {
		log.Println("MsczCleanUp error:", err)
	}()
	var archiveFile string
	if archiveFile, err = MsczBackup(file); err != nil {
		util.WarnOnError(err)
		return err
	}

	r, err := zip.OpenReader(archiveFile)
	if err != nil {
		util.WarnOnError(err)
		return err
	}
	defer r.Close()

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, f := range r.File {
		rc, er0 := f.Open()
		if er0 != nil {
			return er0
		}
		defer rc.Close()
		data, er1 := ioutil.ReadAll(rc)
		if er1 != nil {
			return er1
		}

		if strings.HasSuffix(f.Name, ".mscx") {
			// Patch
			doc, er := xml.ParseXML(string(data))
			if er != nil {
				return er
			}
			notes := doc.Root.Query("//Chord")

			for i := range notes {
				cNote := notes[i]
				if l := cNote.FindOneByName("Lyrics"); l != nil {
					cNote.RemoveChild(l)
				}
			}
			data = []byte(doc.XMLPretty())
		}
		to, er2 := w.Create(f.Name)
		if er2 != nil {
			return er2
		}
		_, er3 := to.Write(data)
		if err != nil {
			return er3
		}
		// to.Close()

	}
	w.Close()
	//	dest := path.Join(dir, "tmp", path.Base(file))

	target, _ := os.Create(file)
	_, erf := buf.WriteTo(target)
	target.Close()

	return erf
}
