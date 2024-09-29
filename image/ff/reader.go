package ff

import (
	"encoding/binary"
	"image"
	"image/color"
	"io"
)

const ffHeader = "farbfeld"

type decoder struct {
	r             io.Reader
	img           image.Image
	width, height int
	tmp           [3 * 256]byte
}

type FormatError string

func (e FormatError) Error() string {
	return "farbfeld: invalid format: " + string(e)
}

func (d *decoder) checkHeader() error {
	_, err := io.ReadFull(d.r, d.tmp[:len(ffHeader)])
	if err != nil {
		return err
	}
	if string(d.tmp[:len(ffHeader)]) != ffHeader {
		return FormatError("not a farbfeld file")
	}
	return nil
}

func Decode(r io.Reader) (image.Image, error) {
	d := &decoder{
		r: r,
	}
	if err := d.checkHeader(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}

	if _, err := io.ReadFull(d.r, d.tmp[:8]); err != nil {
		return nil, err
	}
	d.width = int(binary.BigEndian.Uint32(d.tmp[:4]))
	d.height = int(binary.BigEndian.Uint32(d.tmp[4:8]))

	rgba := image.NewRGBA64(image.Rect(0, 0, d.width, d.height))
	d.img = rgba

	for col := 0; col < d.width; col++ {
		for row := 0; row < d.height; row++ {
			pixel, err := d.parsePixel()
			if err != nil {
				return nil, err
			}
			rgba.SetRGBA64(row, col, pixel)
		}
	}

	return d.img, nil
}

func (d *decoder) parsePixel() (color.RGBA64, error) {
	if _, err := io.ReadFull(d.r, d.tmp[:8]); err != nil {
		return color.RGBA64{}, err
	}
	red := binary.BigEndian.Uint16(d.tmp[:2])
	green := binary.BigEndian.Uint16(d.tmp[2:4])
	blue := binary.BigEndian.Uint16(d.tmp[4:6])
	alpha := binary.BigEndian.Uint16(d.tmp[6:8])
	return color.RGBA64{
		R: red,
		G: green,
		B: blue,
		A: alpha,
	}, nil
}

func (d *decoder) parseWH() error {
	if _, err := io.ReadFull(d.r, d.tmp[:4]); err != nil {
		return err
	}
	width := binary.BigEndian.Uint32(d.tmp[:4])

	if _, err := io.ReadFull(d.r, d.tmp[:4]); err != nil {
		return err
	}
	height := binary.BigEndian.Uint32(d.tmp[:4])

	d.width, d.height = int(width), int(height)
	return nil
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	d := &decoder{
		r: r,
	}

	if err := d.checkHeader(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return image.Config{}, err
	}

	if err := d.parseWH(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return image.Config{}, err
	}

	return image.Config{
		ColorModel: color.RGBAModel,
		Width:      d.width,
		Height:     d.height,
	}, nil
}

func init() {
	image.RegisterFormat("ff", ffHeader, Decode, DecodeConfig)
}
