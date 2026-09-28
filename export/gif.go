package export

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
	"sort"
)

// GIFOptions controls GIF encoding.
type GIFOptions struct {
	FPS       float64 // frames per second; 25 if zero
	LoopCount int     // 0 loops forever
}

// EncodeGIF writes frames as an animated GIF.
//
// One palette is built from all frames and mapped without dithering: cards are
// flat colors and anti-aliased text, and dithering noise would differ from
// frame to frame and flicker. After the first frame, only the bounding box of
// changed pixels is stored, which keeps loops small (a few moving dots on a
// static card compress to a fraction of the full-frame size).
func EncodeGIF(w io.Writer, frames []image.Image, o GIFOptions) error {
	if len(frames) == 0 {
		return errors.New("export: no frames")
	}
	if o.FPS == 0 {
		o.FPS = 25
	}
	delay := int(100/o.FPS + 0.5) // GIF delay is in 1/100 s
	if delay < 2 {
		delay = 2 // browsers clamp shorter delays
	}

	// One slot is reserved for a transparent index, used to skip unchanged
	// pixels inside a delta patch so they compress to long runs.
	opaque := buildPalette(frames, 255)
	pal := append(append(color.Palette{}, opaque...), color.RGBA{})
	transparent := uint8(len(pal) - 1)
	index := newIndexer(opaque)

	out := &gif.GIF{LoopCount: o.LoopCount}
	b := frames[0].Bounds()
	out.Config = image.Config{ColorModel: pal, Width: b.Dx(), Height: b.Dy()}

	var prev *image.Paletted
	for _, f := range frames {
		cur := index.paletted(f, pal)
		if prev == nil {
			out.Image = append(out.Image, cur)
		} else {
			box := diffBounds(prev, cur)
			if box.Empty() {
				box = image.Rect(0, 0, 1, 1) // identical frame: keep timing with a 1px patch
			}
			patch := image.NewPaletted(box, pal)
			draw.Draw(patch, box, cur, box.Min, draw.Src)
			for y := box.Min.Y; y < box.Max.Y; y++ {
				for x := box.Min.X; x < box.Max.X; x++ {
					if prev.ColorIndexAt(x, y) == cur.ColorIndexAt(x, y) {
						patch.SetColorIndex(x, y, transparent)
					}
				}
			}
			out.Image = append(out.Image, patch)
		}
		out.Delay = append(out.Delay, delay)
		out.Disposal = append(out.Disposal, gif.DisposalNone)
		prev = cur
	}
	return gif.EncodeAll(w, out)
}

// diffBounds returns the smallest rectangle containing every differing pixel.
func diffBounds(a, b *image.Paletted) image.Rectangle {
	r := a.Bounds()
	minX, minY, maxX, maxY := r.Max.X, r.Max.Y, r.Min.X, r.Min.Y
	for y := r.Min.Y; y < r.Max.Y; y++ {
		ra := a.Pix[a.PixOffset(r.Min.X, y):a.PixOffset(r.Max.X, y)]
		rb := b.Pix[b.PixOffset(r.Min.X, y):b.PixOffset(r.Max.X, y)]
		for i := range ra {
			if ra[i] == rb[i] {
				continue
			}
			x := r.Min.X + i
			minX, maxX = min(minX, x), max(maxX, x+1)
			minY, maxY = min(minY, y), max(maxY, y+1)
		}
	}
	if maxX <= minX {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX, maxY)
}

// indexer maps RGBA colors to palette indexes with a cache; cards contain few
// distinct colors so nearest-color search is rare.
type indexer struct {
	pal   color.Palette
	cache map[uint32]uint8
}

func newIndexer(p color.Palette) *indexer {
	return &indexer{pal: p, cache: make(map[uint32]uint8, 1024)}
}

func (ix *indexer) paletted(src image.Image, pal color.Palette) *image.Paletted {
	b := src.Bounds()
	dst := image.NewPaletted(image.Rect(0, 0, b.Dx(), b.Dy()), pal)
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			r, g, bl, _ := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			key := r>>8<<16 | g>>8<<8 | bl>>8
			i, ok := ix.cache[key]
			if !ok {
				i = uint8(ix.pal.Index(color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), 255}))
				ix.cache[key] = i
			}
			dst.Pix[y*dst.Stride+x] = i
		}
	}
	return dst
}

// buildPalette returns up to limit colors representing all frames. If the
// frames use no more than limit distinct colors the palette is exact;
// otherwise median cut.
func buildPalette(frames []image.Image, limit int) color.Palette {
	counts := map[uint32]int{}
	for _, f := range frames {
		b := f.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				r, g, bl, _ := f.At(x, y).RGBA()
				counts[r>>8<<16|g>>8<<8|bl>>8]++
			}
		}
	}
	type entry struct {
		rgb   [3]uint8
		count int
	}
	entries := make([]entry, 0, len(counts))
	for k, n := range counts {
		entries = append(entries, entry{[3]uint8{uint8(k >> 16), uint8(k >> 8), uint8(k)}, n})
	}
	// Deterministic order regardless of map iteration.
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i].rgb, entries[j].rgb
		if a[0] != b[0] {
			return a[0] < b[0]
		}
		if a[1] != b[1] {
			return a[1] < b[1]
		}
		return a[2] < b[2]
	})

	if len(entries) <= limit {
		pal := make(color.Palette, len(entries))
		for i, e := range entries {
			pal[i] = color.RGBA{e.rgb[0], e.rgb[1], e.rgb[2], 255}
		}
		return pal
	}

	// Median cut over weighted colors.
	type box []entry
	boxes := []box{entries}
	for len(boxes) < limit {
		// Split the box with the widest channel range.
		bi, bch, bspan := -1, 0, 0
		for i, bx := range boxes {
			if len(bx) < 2 {
				continue
			}
			for ch := 0; ch < 3; ch++ {
				lo, hi := 255, 0
				for _, e := range bx {
					lo, hi = min(lo, int(e.rgb[ch])), max(hi, int(e.rgb[ch]))
				}
				if hi-lo > bspan {
					bi, bch, bspan = i, ch, hi-lo
				}
			}
		}
		if bi < 0 {
			break
		}
		bx := boxes[bi]
		sort.Slice(bx, func(i, j int) bool { return bx[i].rgb[bch] < bx[j].rgb[bch] })
		total := 0
		for _, e := range bx {
			total += e.count
		}
		acc, cut := 0, 1
		for i, e := range bx {
			acc += e.count
			if acc*2 >= total {
				cut = min(max(i+1, 1), len(bx)-1)
				break
			}
		}
		boxes[bi] = bx[:cut]
		boxes = append(boxes, bx[cut:])
	}
	pal := make(color.Palette, 0, len(boxes))
	for _, bx := range boxes {
		var r, g, b, n int
		for _, e := range bx {
			r += int(e.rgb[0]) * e.count
			g += int(e.rgb[1]) * e.count
			b += int(e.rgb[2]) * e.count
			n += e.count
		}
		pal = append(pal, color.RGBA{uint8(r / n), uint8(g / n), uint8(b / n), 255})
	}
	return pal
}
