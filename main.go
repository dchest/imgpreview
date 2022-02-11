package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/chai2010/webp"

	"github.com/disintegration/imaging"
)

var (
	fSize    = flag.Int("s", 42, "maximum preview side size")
	fQuality = flag.Int("q", 1, "WebP quality (0-100)")
	fBlur    = flag.Int("b", 10, "blur")
	fTag     = flag.Bool("tag", false, "output img tag")
	fWebP    = flag.Bool("webp", false, "output the preview WebP to stdout")
	fSvg     = flag.Bool("svg", false, "output the preview SVG to stdout")
	fBase64  = flag.Bool("base64", false, "encode SVG as Base64 instead of quoting")
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	img, err := imaging.Open(flag.Arg(0), imaging.AutoOrientation(true))
	if err != nil {
		log.Fatal(err)
	}
	if !isOpaque(img) {
		log.Printf("************\nWARNING! Image has transparency, see caveats in README.\n************")
	}
	resizedImg := imaging.Fit(img, *fSize, *fSize, imaging.Lanczos)
	imageData, err := webp.EncodeRGBA(resizedImg, float32(*fQuality))
	if err != nil {
		log.Fatal(err)
	}
	if *fWebP {
		os.Stdout.Write(imageData)
		return
	}
	encodedImage := base64.StdEncoding.EncodeToString(imageData)
	svg := strings.ReplaceAll(svgTemplate, "{{BLUR}}", strconv.Itoa(*fBlur))
	svg = strings.ReplaceAll(svg, "{{WEBP}}", encodedImage)
	if *fSvg {
		os.Stdout.Write([]byte(svg))
		return
	}
	if *fBase64 {
		svg = ";base64," + base64.StdEncoding.EncodeToString([]byte(svg))
	} else {
		svg = strings.ReplaceAll(svg, `"`, `%22`)
		svg = strings.ReplaceAll(svg, `<`, `%3C`)
		svg = strings.ReplaceAll(svg, `>`, `%3E`)
		svg = strings.ReplaceAll(svg, `#`, `%23`)
		svg = "," + svg
	}
	css := fmt.Sprintf("background: url('data:image/svg+xml%s') no-repeat 100%%", svg)
	if *fTag {
		fmt.Printf("<img src=\"%s\" width=\"%d\" height=\"%d\" alt=\"\" style=\"%s\">\n",
			filepath.Base(flag.Arg(0)), img.Bounds().Dx(), img.Bounds().Dy(), css)
	} else {
		fmt.Println(css)
	}
}

func init() {
	// Minify svgtemplate.
	// Doing it at runtime for template readability.
	var buf bytes.Buffer
	var p byte
	for _, c := range []byte(svgTemplate) {
		switch c {
		case ' ', '\t', '\n', '\r':
			switch p {
			case 0, '>', ' ', '\t', '\n', '\r':
				// ignore
			default:
				buf.WriteByte(c)
			}
		default:
			buf.WriteByte(c)
		}
		p = c
	}
	svgTemplate = buf.String()
}

func isOpaque(m image.Image) bool {
	switch m := m.(type) {
	case *image.NRGBA:
		return m.Opaque()
	case *image.NRGBA64:
		return m.Opaque()
	case *image.RGBA:
		return m.Opaque()
	case *image.RGBA64:
		return m.Opaque()
	case *image.Paletted:
		return m.Opaque()
	}
	return true
}

var svgTemplate = `
<svg xmlns="http://www.w3.org/2000/svg">
  <defs>
    <filter id="f">
      <feGaussianBlur stdDeviation="{{BLUR}}"/>
    </filter>
  </defs>
  <image width="100%" height="100%" filter="url(#f)" href="data:image/webp;base64,{{WEBP}}"/>
</svg>`
