//go:build ignore

// Command winicon writes the Windows icon for the guix build, the same way
// `wails build` does (generateWindowsIcon in github.com/wailsapp/wails/cmd,
// v1.16.9, MIT licence): the PNG scaled to each size with Catmull-Rom and
// stored as PNG images inside an .ico.
//
// Usage: go run contrib/guix/winicon.go appicon.png vertcoin-ocm.ico
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/draw"
)

type icoHeader struct {
	_          uint16
	imageType  uint16
	imageCount uint16
}

type icoDescriptor struct {
	width  uint8
	height uint8
	colors uint8
	_      uint8
	planes uint16
	bpp    uint16
	size   uint32
	offset uint32
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: winicon in.png out.ico")
	}
	in, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	src, err := png.Decode(in)
	in.Close()
	if err != nil {
		log.Fatal(err)
	}

	sizes := []int{256, 128, 64, 48, 32, 16}
	var descriptors []icoDescriptor
	var images [][]byte
	for _, size := range sizes {
		rect := image.Rect(0, 0, size, size)
		dst := image.NewRGBA(rect)
		draw.CatmullRom.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
		var buf bytes.Buffer
		if err := png.Encode(&buf, dst); err != nil {
			log.Fatal(err)
		}
		dim := size
		if dim >= 256 {
			dim = 0
		}
		descriptors = append(descriptors, icoDescriptor{
			width:  uint8(dim),
			height: uint8(dim),
			planes: 1,
			bpp:    32,
			size:   uint32(buf.Len()),
		})
		images = append(images, buf.Bytes())
	}

	var out bytes.Buffer
	binary.Write(&out, binary.LittleEndian, icoHeader{imageType: 1, imageCount: uint16(len(sizes))})
	offset := uint32(6 + 16*len(sizes))
	for _, d := range descriptors {
		d.offset = offset
		binary.Write(&out, binary.LittleEndian, d)
		offset += d.size
	}
	for _, img := range images {
		out.Write(img)
	}
	if err := os.WriteFile(os.Args[2], out.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
