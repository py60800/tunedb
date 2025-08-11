package zdb

import (
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/py60800/tunedb/internal/search"
	"github.com/py60800/tunedb/internal/util"

	mp3tag "github.com/dhowden/tag"
)

type MP3Collection struct {
	list []MP3File
	//	hasContent map[int]bool
	index map[int]int
	node0 *search.Node
}

func MP3CollectionNew(db *TuneDB) *MP3Collection {
	mp3Files := db.Mp3DBLoad()
	node0 := search.NodeNew()
	index := make(map[int]int)
	for i, d := range mp3Files {
		words := d.Artist + " " + d.Title
		node0.IndexWords(words, i)
		index[d.ID] = i
	}
	//	hasContent := make(map[int]bool)
	ids := db.Mp3SetIds() // mp3FilesId
	for _, id := range ids {
		mp3Files[index[id]].HasContent = true
	}
	log.Printf("Mp3Collection: %d files, %d nodes\n", len(mp3Files), node0.Count())
	return &MP3Collection{
		list:  mp3Files,
		node0: node0,
		index: index,
	}
}
func (c *MP3Collection) MarkContent(id int) {
	for i := range c.list {
		if c.list[i].ID == id {
			c.list[i].HasContent = true
			break
		}
	}
}
func (c *MP3Collection) GetByID(ID int) *MP3File {
	idx := c.index[ID]
	if idx >= 0 && idx < len(c.list) {
		return &c.list[idx]
	}
	return &MP3File{}
}
func (c *MP3Collection) GetByFileName(file string) *MP3File {
	for i := range c.list {
		if c.list[i].File == file {
			return &c.list[i]
		}
	}
	return nil

}
func (c *MP3Collection) GetByIds(id []int) []MP3File {
	m := make([]MP3File, 0)
	for _, i := range id {
		for j := range c.list {
			if i == c.list[j].ID {
				m = append(m, c.list[j])
			}
		}
	}
	return m
}
func (c *MP3Collection) Search(what string, withContent bool, limit int) []MP3File {
	res := make([]MP3File, 0)
	var l []int
	if what == "" {
		l = make([]int, len(c.list))
		for i := range l {
			l[i] = i
		}
	} else {
		l = c.node0.Search(what)

	}
	for _, n := range l {
		res = append(res, c.list[n])
		if len(res) >= limit {
			break
		}
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].HasContent != res[j].HasContent {
			return res[i].HasContent
		}
		if res[i].Artist != res[j].Artist {
			return res[i].Artist < res[j].Artist
		}
		return res[i].Title < res[j].Title
	})

	return res
}

func MP3GetTag(file string, date time.Time, mp3File *MP3File) (*MP3File, error) {
	if mp3File == nil {
		// default
		mp3File = &MP3File{
			File:   file,
			Title:  strings.TrimSuffix(path.Base(file), ".mp3"),
			Artist: "Unknown",
		}
	}
	mp3File.FileDate = date
	if f, err := os.Open(file); err == nil {
		defer f.Close()
		if m, err := mp3tag.ReadFrom(f); err == nil {
			mp3File.Track, _ = m.Track()
			title := m.Title()
			if title == "" {
				mp3File.Title = strings.TrimSuffix(path.Base(file), ".mp3")
			} else {
				mp3File.Title = title
			}
			artist := m.Artist()
			if artist == "" {
				mp3File.Artist = "Unknow"
			} else {
				mp3File.Artist = artist
			}
			mp3File.Album = m.Album()
		}
	} else {
		util.CheckErr(err, file)
		return mp3File, err
	}
	return mp3File, nil
}

func MP3DBUpdate(db *TuneDB) {

	dirs := db.SourceRepositoryGetAll()
	for _, d := range dirs {
		if d.Type == "Mp3" {
			start := time.Now()
			files := GetFileListR2(d.Location, ".mp3", d.Recurse)
			stepT := time.Now()
			log.Printf("GetFileList %v : Found: %v , D:%v\n", d.Location, len(files), stepT.Sub(start))
			db.Mp3DBStore2(files)
			log.Println("Mp3DBStore Duration:", time.Now().Sub(stepT))
		}
	}
}
func (c *MP3Collection) GetByTuneID(tuneID int) []MP3File {
	mp3 := tuneDB.Mp3SetGetByTuneID(tuneID)
	m := make([]MP3File, len(mp3))
	for i, id := range mp3 {
		m[i] = *tuneDB.Mp3FileGetByID(id)
	}
	return m
}
