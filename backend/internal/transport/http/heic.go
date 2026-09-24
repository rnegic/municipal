package http

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"

	"github.com/gen2brain/heic"
)

func isHEIF(b []byte) bool {
	if len(b) < 12 || string(b[4:8]) != "ftyp" {
		return false
	}
	switch string(b[8:12]) {
	case "heic", "heix", "hevc", "hevx", "heim", "heis", "mif1", "msf1":
		return true
	}
	return false
}

const maxHEICPixels = 50_000_000

var errHEICTooLarge = errors.New("heic larger than 50 MP")

var heicSlot = make(chan struct{}, 1)

func heicToJPEG(b []byte) (out []byte, err error) {
	heicSlot <- struct{}{}
	defer func() { <-heicSlot }()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("heic decode panic: %v", r)
		}
	}()
	cfg, err := heic.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if cfg.Width*cfg.Height > maxHEICPixels {
		return nil, errHEICTooLarge
	}
	img, err := heic.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
