// Package export rasterizes SVG documents, including SMIL-animated ones, using
// headless Chrome, and encodes frame sequences as GIF.
//
// Capturing the browser's own rendering keeps every SVG feature (fonts, emoji,
// SMIL motion paths) identical to what a viewer sees, and works for SVG from
// any source: the card renderer or D2 output enriched with flow dots.
package export

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Options controls capture.
type Options struct {
	Width, Height int           // CSS pixels; must match the SVG viewBox size
	Scale         float64       // device scale factor; 1 if zero, 2 for retina-quality PNG
	Timeout       time.Duration // overall limit; 2 minutes if zero
	// NoSandbox passes --no-sandbox to Chrome. Needed when the calling process
	// is itself sandboxed (some CI containers, agent sandboxes) because Chrome's
	// sandbox cannot nest. The document rendered is our own generated SVG.
	NoSandbox bool
}

// Session is a headless Chrome page holding one SVG document.
type Session struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// Open launches headless Chrome and loads svg. Close must be called.
func Open(ctx context.Context, svg []byte, o Options) (*Session, error) {
	if o.Width <= 0 || o.Height <= 0 {
		return nil, errors.New("export: width and height are required")
	}
	if o.Scale == 0 {
		o.Scale = 1
	}
	if o.Timeout == 0 {
		o.Timeout = 2 * time.Minute
	}
	ctx, cancelTimeout := context.WithTimeout(ctx, o.Timeout)
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("force-color-profile", "srgb"), // stable colors across machines
	)
	if o.NoSandbox {
		allocOpts = append(allocOpts, chromedp.NoSandbox)
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	cancel := func() { cancelBrowser(); cancelAlloc(); cancelTimeout() }

	doc := `<!doctype html><html><head><meta charset="utf-8"></head><body style="margin:0;overflow:hidden">` + string(svg) + `</body></html>`
	err := chromedp.Run(browserCtx,
		chromedp.EmulateViewport(int64(o.Width), int64(o.Height), chromedp.EmulateScale(o.Scale)),
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			tree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(tree.Frame.ID, doc).Do(ctx)
		}),
		chromedp.WaitReady("svg"),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("export: load svg in chrome (is Chrome installed?): %w", err)
	}
	return &Session{ctx: browserCtx, cancel: cancel}, nil
}

// Close shuts the browser down.
func (s *Session) Close() { s.cancel() }

// FrameAt pauses SMIL animation at t seconds and returns a screenshot.
func (s *Session) FrameAt(t float64) (image.Image, error) {
	var buf []byte
	err := chromedp.Run(s.ctx,
		chromedp.Evaluate(fmt.Sprintf(`(() => {
			const s = document.querySelector('svg');
			s.pauseAnimations(); s.setCurrentTime(%[1]g);                     // SMIL
			document.getAnimations().forEach(a => { a.pause(); a.currentTime = %[1]g * 1000; }); // CSS
		})()`, t), nil),
		chromedp.CaptureScreenshot(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("export: capture at %.3fs: %w", t, err)
	}
	return png.Decode(bytes.NewReader(buf))
}

// PNG returns a PNG screenshot of svg with animation paused at t seconds.
func PNG(ctx context.Context, svg []byte, t float64, o Options) ([]byte, error) {
	s, err := Open(ctx, svg, o)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	img, err := s.FrameAt(t)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// Frames captures a full loop of `loop` seconds at fps frames per second.
func Frames(ctx context.Context, svg []byte, loop, fps float64, o Options) ([]image.Image, error) {
	if loop <= 0 || fps <= 0 {
		return nil, errors.New("export: loop and fps must be > 0")
	}
	s, err := Open(ctx, svg, o)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	n := int(loop * fps)
	frames := make([]image.Image, 0, n)
	for i := 0; i < n; i++ {
		img, err := s.FrameAt(float64(i) / fps)
		if err != nil {
			return nil, err
		}
		frames = append(frames, img)
	}
	return frames, nil
}
