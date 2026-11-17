package tests

import (
	"slices"
	"strings"
	"testing"

	"github.com/tommyo123/mtag/id3v2"
)

// FuzzID3v2DecodeString checks that every encoding round-trips a
// NUL-free string with its canonical terminator, and that a UTF-16
// string ended by a single NUL byte, as some taggers write it, decodes
// the same through both DecodeString and SplitStrings. Arbitrary bytes
// must never make either function panic.
func FuzzID3v2DecodeString(f *testing.F) {
	f.Add(byte(id3v2.EncUTF16), "Stam1na", []byte{0xFF, 0xFE, 'a', 0, 0})
	f.Add(byte(id3v2.EncUTF16BE), "Käskyä", []byte{0x4E, 0x00, 0})
	f.Add(byte(id3v2.EncUTF16), "😀", []byte{0xFF, 0xFE, 0x3D, 0xD8})
	f.Add(byte(id3v2.EncUTF8), "", []byte{})
	f.Fuzz(func(t *testing.T, encByte byte, s string, raw []byte) {
		enc := id3v2.Encoding(encByte % 4)
		_, _ = id3v2.DecodeString(enc, raw)
		_, _ = id3v2.SplitStrings(enc, raw)

		if strings.ContainsRune(s, 0) {
			return
		}
		switch enc {
		case id3v2.EncISO8859:
			s = strings.Map(func(r rune) rune {
				if r > 0xFF {
					return '?'
				}
				return r
			}, s)
		case id3v2.EncUTF16, id3v2.EncUTF16BE:
			s = string([]rune(s)) // the UTF-16 encoder turns each invalid byte into U+FFFD
		}
		data, err := id3v2.EncodeString(enc, s, true)
		if err != nil {
			t.Fatal(err)
		}
		if got, err := id3v2.DecodeString(enc, data); err != nil || got != s {
			t.Fatalf("round trip of %q via %x = %q, %v", s, data, got, err)
		}
		if enc != id3v2.EncUTF16 && enc != id3v2.EncUTF16BE {
			return
		}
		single := data[:len(data)-1]
		if got, err := id3v2.DecodeString(enc, single); err != nil || got != s {
			t.Fatalf("DecodeString of single-NUL %x = %q, %v; want %q", single, got, err, s)
		}
		var want []string
		if s != "" {
			want = []string{s}
		}
		if got, err := id3v2.SplitStrings(enc, single); err != nil || !slices.Equal(got, want) {
			t.Fatalf("SplitStrings of single-NUL %x = %q, %v; want %q", single, got, err, want)
		}
	})
}
