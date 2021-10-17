package exifremove

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image/jpeg"
	"image/png"

	"github.com/dsoprea/go-exif"
	"github.com/dsoprea/go-jpeg-image-structure"
	"github.com/dsoprea/go-png-image-structure"
)

func Remove(data []byte) ([]byte, error) {

	const (
		JpegMediaType  = "jpeg"
		PngMediaType   = "png"
		OtherMediaType = "other"
		StartBytes     = 0
		EndBytes       = 0
	)

	type MediaContext struct {
		MediaType string
		RootIfd   *exif.Ifd
		RawExif   []byte
		Media     interface{}
	}

	jmp := jpegstructure.NewJpegMediaParser()
	pmp := pngstructure.NewPngMediaParser()
	var before, after []byte

	// copy data not to effect args
	before = append([]byte{}, data...)

	if jmp.LooksLikeFormat(before) {

		sl, err := jmp.ParseBytes(before)
		if err != nil {
			return nil, err
		}

		_, rawExif, err := sl.Exif()
		if err != nil {
			return data, nil
		}

		startExifBytes := StartBytes
		endExifBytes := EndBytes

		for i := 0; i < len(before)-len(rawExif); i++ {
			if bytes.Compare(before[i:i+len(rawExif)], rawExif) == 0 {
				startExifBytes = i
				endExifBytes = i + len(rawExif)
				break
			}
		}
		fill := make([]byte, len(before[startExifBytes:endExifBytes]))
		copy(before[startExifBytes:endExifBytes], fill)

		after = before

		_, err = jpeg.Decode(bytes.NewReader(after))
		if err != nil {
			return nil, errors.New("EXIF removal corrupted " + err.Error())
		}

	} else if pmp.LooksLikeFormat(before) {

		cs, err := pmp.ParseBytes(before)
		if err != nil {
			return nil, err
		}

		_, rawExif, err := cs.Exif()
		if err != nil {
			return data, nil
		}

		startExifBytes := StartBytes
		endExifBytes := EndBytes

		for i := 0; i < len(before)-len(rawExif); i++ {
			if bytes.Compare(before[i:i+len(rawExif)], rawExif) == 0 {
				startExifBytes = i
				endExifBytes = i + len(rawExif)
				break
			}
		}
		fill := make([]byte, len(before[startExifBytes:endExifBytes]))
		copy(before[startExifBytes:endExifBytes], fill)

		after = before

		chunks := readPNGChunks(bytes.NewReader(after))

		for _, chunk := range chunks {
			if !chunk.CRCIsValid() {
				offset := int(chunk.Offset) + 8 + int(chunk.Length)
				crc := chunk.CalculateCRC()

				buf := new(bytes.Buffer)
				binary.Write(buf, binary.BigEndian, crc)
				crcBytes := buf.Bytes()

				copy(after[offset:], crcBytes)
			}
		}

		chunks = readPNGChunks(bytes.NewReader(after))
		for _, chunk := range chunks {
			if !chunk.CRCIsValid() {
				return nil, errors.New("EXIF removal failed CRC")
			}
		}

		_, err = png.Decode(bytes.NewReader(after))
		if err != nil {
			return nil, errors.New("EXIF removal corrupted " + err.Error())
		}

	}

	return after, nil
}
