package id3v2

import (
	"slices"
	"testing"
)

func TestDecodeUTF16SingleNULTerminator(t *testing.T) {
	for _, tt := range []struct {
		name string
		enc  Encoding
		data []byte
		want string
	}{
		{"UTF-16LE with BOM, single NUL", EncUTF16, []byte{0xFF, 0xFE, 'O', 0, 'v', 0, 'i', 0, 0}, "Ovi"},
		{"UTF-16LE with BOM, double NUL", EncUTF16, []byte{0xFF, 0xFE, 'O', 0, 'v', 0, 'i', 0, 0, 0}, "Ovi"},
		{"UTF-16LE with BOM, no NUL", EncUTF16, []byte{0xFF, 0xFE, 'O', 0, 'v', 0, 'i', 0}, "Ovi"},
		{"UTF-16LE non-ASCII, single NUL", EncUTF16, []byte{0xFF, 0xFE, 'K', 0, 0xE4, 0, 0}, "Kä"},
		{"UTF-16LE with BOM, double NUL plus pad NUL", EncUTF16, []byte{0xFF, 0xFE, 'O', 0, 'v', 0, 'i', 0, 0, 0, 0}, "Ovi"},
		{"UTF-16BE with BOM, single NUL", EncUTF16, []byte{0xFE, 0xFF, 0x01, 0x00, 0}, "Ā"},
		{"UTF-16BE, single NUL", EncUTF16BE, []byte{0x4E, 0x00, 0}, "一"},
		{"UTF-16BE surrogate pair, single NUL", EncUTF16BE, []byte{0xD8, 0x3D, 0xDE, 0x00, 0}, "😀"},
		{"UTF-16BE, double NUL", EncUTF16BE, []byte{0x4E, 0x00, 0, 0}, "一"},
		{"UTF-16BE, double NUL plus pad NUL", EncUTF16BE, []byte{0x01, 0x00, 0, 0, 0}, "Ā"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeString(tt.enc, tt.data)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("DecodeString = %q, want %q", got, tt.want)
			}
			values, err := SplitStrings(tt.enc, tt.data)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(values, []string{tt.want}) {
				t.Errorf("SplitStrings = %q, want [%q]", values, tt.want)
			}
		})
	}
}

func TestSplitStringsUTF16LastValueSingleNUL(t *testing.T) {
	data := []byte{0xFF, 0xFE, 'A', 0, 0, 0, 0xFF, 0xFE, 'B', 0, 0}
	values, err := SplitStrings(EncUTF16, data)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"A", "B"}; !slices.Equal(values, want) {
		t.Errorf("SplitStrings = %q, want %q", values, want)
	}
}
