package export

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"testing"
)

func solidFrame(w, h int, bg color.RGBA, dot image.Point) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Rect, image.NewUniform(bg), image.Point{}, draw.Src)
	img.SetRGBA(dot.X, dot.Y, color.RGBA{255, 0, 0, 255})
	return img
}

func TestEncodeGIFRoundTrip(t *testing.T) {
	bg := color.RGBA{10, 10, 10, 255}
	frames := []image.Image{
		solidFrame(40, 30, bg, image.Pt(5, 5)),
		solidFrame(40, 30, bg, image.Pt(6, 5)),
		solidFrame(40, 30, bg, image.Pt(7, 5)),
	}
	var buf bytes.Buffer
	if err := EncodeGIF(&buf, frames, GIFOptions{FPS: 25}); err != nil {
		t.Fatal(err)
	}
	g, err := gif.DecodeAll(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Image) != 3 || g.Delay[0] != 4 {
		t.Fatalf("frames=%d delay=%v", len(g.Image), g.Delay)
	}
	// Later frames are small patches, not full frames.
	if b := g.Image[1].Bounds(); b.Dx() >= 40 || b.Dy() >= 30 {
		t.Errorf("frame 1 should be a delta patch, bounds %v", b)
	}
	// Compositing the patches reproduces the last frame.
	canvas := image.NewRGBA(image.Rect(0, 0, 40, 30))
	for _, p := range g.Image {
		draw.Draw(canvas, p.Bounds(), p, p.Bounds().Min, draw.Over)
	}
	want := frames[2]
	for y := 0; y < 30; y++ {
		for x := 0; x < 40; x++ {
			if canvas.At(x, y) != want.At(x, y) {
				t.Fatalf("pixel (%d,%d) = %v, want %v", x, y, canvas.At(x, y), want.At(x, y))
			}
		}
	}
}

func TestEncodeGIFManyColors(t *testing.T) {
	// More than 255 distinct colors forces median cut.
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(x * 4), uint8(y * 4), uint8((x + y) * 2), 255})
		}
	}
	var buf bytes.Buffer
	if err := EncodeGIF(&buf, []image.Image{img, img}, GIFOptions{}); err != nil {
		t.Fatal(err)
	}
	g, err := gif.DecodeAll(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(g.Image[0].Palette); n > 256 {
		t.Errorf("palette has %d entries", n)
	}
	if len(g.Image) != 2 {
		t.Errorf("identical frames must still produce a frame for timing, got %d", len(g.Image))
	}
}

func TestEncodeGIFNoFrames(t *testing.T) {
	if err := EncodeGIF(&bytes.Buffer{}, nil, GIFOptions{}); err == nil {
		t.Error("expected error for no frames")
	}
}
