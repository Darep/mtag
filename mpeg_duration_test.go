package mtag

import (
	"bytes"
	"testing"
	"time"

	"github.com/tommyo123/mtag/ape"
	"github.com/tommyo123/mtag/id3v2"
)

func durationTestFrames(index byte, count int) []byte {
	header := []byte{0xff, 0xfb, index << 4, 0xc4}
	frame, _ := parseMPEGHeader(header)
	data := make([]byte, frame.frameLen)
	copy(data, header)
	return bytes.Repeat(data, count)
}

func TestMPEGDurationCountsAudioOnly(t *testing.T) {
	constant := durationTestFrames(9, 1000)
	variable := append(durationTestFrames(1, 200), durationTestFrames(9, 800)...)
	apeTag := ape.New()
	apeTag.Set(ape.FieldComment, "trailing tag")
	apeBytes, err := apeTag.Encode()
	if err != nil {
		t.Fatal(err)
	}
	// A footer-bearing ID3v2.4 tag whose picture contains bytes resembling audio.
	tag := id3v2.NewTag(4)
	tag.Set(&id3v2.PictureFrame{MIME: "image/jpeg", Data: durationTestFrames(9, 200)})
	appended, err := tag.Encode(0)
	if err != nil {
		t.Fatal(err)
	}
	appended[5] |= 0x10
	footer := append([]byte("3DI"), appended[3:10]...)
	appended = append(appended, footer...)

	for name, data := range map[string][]byte{
		"constant":                   constant,
		"variable":                   variable,
		"late bitrate change":        append(durationTestFrames(9, 800), durationTestFrames(1, 200)...),
		"APEv2":                      append(append([]byte{}, variable...), apeBytes...),
		"ID3v1":                      append(append(append([]byte{}, variable...), "TAG"...), make([]byte, 125)...),
		"appended ID3v2":             append(append([]byte{}, constant...), appended...),
		"appended and leading ID3v2": append(append(append([]byte{}, appended...), constant...), appended...),
		"padding":                    append(append([]byte{}, variable...), make([]byte, 512)...),
		"damaged data":               append(append(durationTestFrames(1, 200), []byte("bad MPEG data")...), durationTestFrames(9, 800)...),
		"partial final frame":        append(append([]byte{}, variable...), durationTestFrames(9, 1)[:20]...),
	} {
		t.Run(name, func(t *testing.T) {
			for _, style := range []AudioPropertiesStyle{AudioPropertiesAccurate, AudioPropertiesAverage} {
				t.Run(style.String(), func(t *testing.T) {
					f, err := OpenBytes(data, WithSkipPictures(), WithAudioPropertiesStyle(style))
					if err != nil {
						t.Fatal(err)
					}
					defer f.Close()
					want := time.Duration(1000*1152) * time.Second / 44100
					got := f.AudioProperties().Duration
					delta := got - want
					if style == AudioPropertiesAccurate && delta != 0 || delta < -time.Millisecond || delta > time.Millisecond {
						t.Fatalf("duration=%v, want %v", got, want)
					}
				})
			}
		})
	}
}

type countedMPEGSource struct {
	*bytes.Reader
	reads int
	bytes int
}

func (r *countedMPEGSource) ReadAt(p []byte, offset int64) (int, error) {
	r.reads++
	n, err := r.Reader.ReadAt(p, offset)
	r.bytes += n
	return n, err
}

func TestMPEGDurationBatchesFrameReads(t *testing.T) {
	data := durationTestFrames(9, 1000)
	source := &countedMPEGSource{Reader: bytes.NewReader(data)}
	f, err := OpenSource(source, int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	source.reads = 0
	if f.AudioProperties().Duration == 0 {
		t.Fatal("missing duration")
	}
	if source.reads > 40 {
		t.Fatalf("1000 frames needed %d reads; expected buffered reads", source.reads)
	}
}

func TestMPEGAverageSamplesConstantBitrate(t *testing.T) {
	data := durationTestFrames(9, 10000)
	source := &countedMPEGSource{Reader: bytes.NewReader(data)}
	f, err := OpenSource(source, int64(len(data)), WithAudioPropertiesStyle(AudioPropertiesAverage))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	source.bytes = 0
	if got := int(f.AudioProperties().Duration.Seconds()); got != 261 {
		t.Fatalf("duration=%d, want 261 seconds", got)
	}
	if source.bytes > 4*(16<<10) {
		t.Fatalf("constant-bitrate estimate read %d bytes, want bounded samples", source.bytes)
	}
}
