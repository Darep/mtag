package mtag

import (
	"encoding/binary"
	"testing"

	"github.com/tommyo123/mtag/id3v2"
)

func TestWAVFieldsFallBackToINFO(t *testing.T) {
	data := buildWAVInfoFile(map[string]string{"INAM": "Title", "IART": "Artist", "IPRD": "Album"})
	tag := id3v2.NewTag(4)
	tag.SetText(id3v2.FramePart, "2/4")
	raw, err := tag.Encode(0)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, "id3 "...)
	data = binary.LittleEndian.AppendUint32(data, uint32(len(raw)))
	data = append(data, raw...)
	if len(raw)%2 != 0 {
		data = append(data, 0)
	}
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(data)-8))
	f, err := OpenBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if f.Title() != "Title" || f.Artist() != "Artist" || f.Album() != "Album" || f.Disc() != 2 {
		t.Fatalf("title=%q artist=%q album=%q disc=%d", f.Title(), f.Artist(), f.Album(), f.Disc())
	}
	f.ID3v2().SetText(id3v2.FrameTitle, "ID3 title")
	if got := f.Title(); got != "ID3 title" {
		t.Fatalf("title=%q, want ID3 title", got)
	}
}
