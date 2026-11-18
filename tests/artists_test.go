package tests

import (
	"bytes"
	"slices"
	"testing"

	"github.com/tommyo123/mtag"
	"github.com/tommyo123/mtag/ape"
	"github.com/tommyo123/mtag/flac"
	"github.com/tommyo123/mtag/id3v1"
	"github.com/tommyo123/mtag/id3v2"
)

func TestArtistsReadsAllValues(t *testing.T) {
	tag := id3v2.NewTag(4)
	tag.SetText(id3v2.FrameArtist, "", "First Artist", " ", "Second Artist", "")
	id3, err := tag.Encode(0)
	if err != nil {
		t.Fatal(err)
	}
	vc := &flac.VorbisComment{Vendor: "test", Fields: []flac.Field{
		{Name: "ARTIST", Value: ""},
		{Name: "ARTIST", Value: "First Artist"},
		{Name: "ARTIST", Value: " "},
		{Name: "ARTIST", Value: "Second Artist"},
	}}
	vorbis := buildFLACWithBlocks(t, flac.Block{Type: flac.BlockVorbisComment, Body: flac.EncodeVorbisComment(vc)})
	apeTag := ape.New()
	apeTag.Set(ape.FieldArtist, "\x00First Artist\x00 \x00Second Artist\x00")
	apeData, err := apeTag.Encode()
	if err != nil {
		t.Fatal(err)
	}
	var artistAtoms []byte
	for _, value := range []string{"", "First Artist", " ", "Second Artist", ""} {
		artistAtoms = append(artistAtoms, mp4Atom("data", append([]byte{0, 0, 0, 1, 0, 0, 0, 0}, value...))...)
	}
	frame := make([]byte, 417)
	copy(frame, []byte{0xff, 0xfb, 0x90, 0xc4})
	for name, data := range map[string][]byte{
		"ID3v2":  append(id3, bytes.Repeat([]byte{0}, 256)...),
		"Vorbis": vorbis,
		"MP4":    buildMinimalMP4(mp4Atom("\xa9ART", artistAtoms), nil),
		"APEv2":  append(bytes.Repeat(frame, 2), apeData...),
	} {
		t.Run(name, func(t *testing.T) {
			f, err := mtag.OpenBytes(data)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			want := []string{"First Artist", "Second Artist"}
			got := f.Artists()
			if !slices.Equal(got, want) {
				t.Fatalf("Artists=%q, want %q", got, want)
			}
			got[0] = "changed"
			if f.Artist() != want[0] || !slices.Equal(f.Artists(), want) {
				t.Fatal("changing the returned slice changed metadata or Artist lost the primary performer")
			}
		})
	}
}

func TestArtistsUsesTagPriorityAndFallback(t *testing.T) {
	tag := id3v2.NewTag(4)
	tag.SetText(id3v2.FrameArtist, "ID3 Artist")
	path := buildTestFile(t, tag, &id3v1.Tag{Artist: "Legacy Artist"}, make([]byte, 256))
	f, err := mtag.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if got := f.Artists(); !slices.Equal(got, []string{"ID3 Artist"}) {
		t.Fatalf("Artists=%q, want ID3 Artist", got)
	}
	f.ID3v2().Remove(id3v2.FrameArtist)
	if got := f.Artists(); !slices.Equal(got, []string{"Legacy Artist"}) {
		t.Fatalf("Artists=%q, want Legacy Artist", got)
	}
	var empty mtag.File
	if got := empty.Artists(); len(got) != 0 {
		t.Fatalf("zero File Artists=%q", got)
	}
}
