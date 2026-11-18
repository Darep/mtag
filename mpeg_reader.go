package mtag

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/tommyo123/mtag/id3v2"
)

// MPEG scans touch every frame header. Cache adjacent reads so a file on a
// network share does not need a separate read for each header or damaged byte.
type mpegReaderAt struct {
	src    io.ReaderAt
	buffer [16 << 10]byte
	offset int64
	size   int
}

func (r *mpegReaderAt) ReadAt(p []byte, offset int64) (int, error) {
	if offset < 0 || len(p) > len(r.buffer) {
		return r.src.ReadAt(p, offset)
	}
	if offset < r.offset || offset+int64(len(p)) > r.offset+int64(r.size) {
		n, err := r.src.ReadAt(r.buffer[:], offset)
		r.offset, r.size = offset, n
		if n < len(p) {
			return copy(p, r.buffer[:n]), err
		}
	}
	copy(p, r.buffer[offset-r.offset:offset-r.offset+int64(len(p))])
	return len(p), nil
}

// mpegAudioEnd excludes tail metadata regardless of tag read/skip options.
func mpegAudioEnd(r io.ReaderAt, end int64) int64 {
	var tail [128]byte
	for {
		if end >= 128 {
			if _, err := r.ReadAt(tail[:3], end-128); err == nil && string(tail[:3]) == "TAG" {
				end -= 128
				continue
			}
		}
		if end >= 32 {
			if _, err := r.ReadAt(tail[:32], end-32); err == nil && string(tail[:8]) == "APETAGEX" {
				size := int64(binary.LittleEndian.Uint32(tail[12:16]))
				if binary.LittleEndian.Uint32(tail[20:24])&(1<<31) != 0 {
					size += 32 // APE's size excludes its optional header.
				}
				if size >= 32 && size <= end {
					end -= size
					continue
				}
			}
		}
		if end >= 20 {
			if _, err := r.ReadAt(tail[:10], end-10); err == nil && string(tail[:3]) == "3DI" && tail[3] == 4 && tail[5]&0x10 != 0 {
				size := int64(id3v2.DecodeSynchsafe(tail[6:10])) + 20
				if size <= end {
					var header [10]byte
					if _, err := r.ReadAt(header[:], end-size); err == nil && string(header[:3]) == "ID3" && bytes.Equal(header[3:], tail[3:10]) {
						end -= size
						continue
					}
				}
			}
		}
		return end
	}
}
