package uploader

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "github.com/biessek/golang-ico"
	"github.com/chai2010/webp"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func ConvertToWebP(data []byte, ext string) ([]byte, error) {
	switch strings.ToLower(ext) {
	case ".gif":
		return data, nil
	case ".heic", ".heif":
		return convertHEICToWebP(data)
	case ".svg":
		return convertSVGToWebP(data)
	case ".jpg", ".jpeg", ".png", ".bmp", ".ico", ".tiff", ".webp":
		return convertRasterToWebP(data)
	default:
		return data, nil
	}
}

func convertRasterToWebP(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)

	var buf bytes.Buffer
	if err := webp.Encode(&buf, rgba, &webp.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}
	return buf.Bytes(), nil
}

func convertSVGToWebP(data []byte) ([]byte, error) {
	icon, err := oksvg.ReadIconStream(bytes.NewReader(data), oksvg.WarnErrorMode)
	if err != nil {
		return nil, fmt.Errorf("decode svg: %w", err)
	}

	w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
	if w <= 0 || h <= 0 {
		w, h = 512, 512
	}

	if w < 512 || h < 512 {
		scale := 512.0 / float64(min(w, h))
		w = int(float64(w) * scale)
		h = int(float64(h) * scale)
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	dasher := rasterx.NewDasher(w, h, scanner)
	icon.SetTarget(0, 0, float64(w), float64(h))
	icon.Draw(dasher, 1.0)

	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("encode svg to webp: %w", err)
	}
	return buf.Bytes(), nil
}

func convertHEICToWebP(data []byte) ([]byte, error) {
	// HEIC/HEIF decoding requires libheif (CGO).
	// To enable: install libheif, then implement decode using github.com/MaestroError/go-libheif.
	return data, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
