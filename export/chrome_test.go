package export

import (
	"bytes"
	"context"
	"image/png"
	"os"
	"os/exec"
	"testing"
	"time"
)

// chromeInstalled reports whether a Chrome that chromedp would find is
// present. The browser tests are skipped without one, so a build without
// Chrome still passes.
func chromeInstalled() bool {
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "chrome"} {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	_, err := os.Stat("/Applications/Google Chrome.app")
	return err == nil
}

const squareSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="80" height="60" viewBox="0 0 80 60"><rect x="10" y="10" width="60" height="40" fill="#336699"/></svg>`

// A page is loaded in Chrome and comes back as a PNG of the size asked for,
// with the drawing in it.
func TestPNGRendersTheSVGInChrome(t *testing.T) {
	if !chromeInstalled() {
		t.Skip("Chrome is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	data, err := PNG(ctx, []byte(squareSVG), 0, Options{Width: 80, Height: 60, NoSandbox: true})
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 80 || b.Dy() != 60 {
		t.Errorf("size = %dx%d, want 80x60", b.Dx(), b.Dy())
	}
	// The middle of the square is the fill color, and the corner is not.
	r, g, bl, _ := img.At(40, 30).RGBA()
	if r>>8 != 0x33 || g>>8 != 0x66 || bl>>8 != 0x99 {
		t.Errorf("middle pixel = #%02x%02x%02x, want #336699", r>>8, g>>8, bl>>8)
	}
	if r2, g2, b2, _ := img.At(2, 2).RGBA(); r2>>8 == 0x33 && g2>>8 == 0x66 && b2>>8 == 0x99 {
		t.Error("the corner has the fill color, so the whole page is filled")
	}
}

func TestOpenNeedsASize(t *testing.T) {
	if _, err := Open(context.Background(), []byte(squareSVG), Options{}); err == nil {
		t.Error("a session without a size was opened")
	}
}
