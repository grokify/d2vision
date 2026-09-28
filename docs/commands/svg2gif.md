# svg2gif

Convert an animated SVG to a looping GIF, for platforms that need a GIF (LinkedIn, for example).

Most SVG rasterizers (`rsvg`, `resvg`, ImageMagick) ignore SVG animation and produce a single still. `svg2gif` loads the SVG in headless Chrome, steps its animation clock one frame at a time, and encodes the frames.

## Usage

```bash
d2vision svg2gif card.svg -o card.gif
d2vision svg2gif diagram.svg -o diagram.gif --loop 4 --fps 20
```

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | required | Output GIF file |
| `--fps` | 25 | Frames per second |
| `--loop` | longest animation in the SVG | Loop length in seconds |
| `--width`, `--height` | from the SVG | Pixel size; otherwise `width`/`height`, falling back to `viewBox` |
| `--chrome-no-sandbox` | off | Launch Chrome with `--no-sandbox`, for sandboxed environments |

## What it animates

- **SMIL** (`<animate>`, `<animateMotion>`): the card renderer's flow dots.
- **CSS animations**: D2's `style.animated: true` edges.

The loop length is the longest `dur` / `animation-duration` found. If an SVG mixes durations that are not multiples of one another it only repeats seamlessly at their common multiple, so pass `--loop` explicitly. If no duration is found the command asks for `--loop`.

## Output size

One shared palette is built from all frames and mapped without dithering, so flat colors and text do not flicker between frames. After the first frame only the changed region is stored, with unchanged pixels transparent. A card with a few moving dots is a few hundred KB.

Converting a card's SVG gives a byte-identical GIF to `card render --gif`.

Requires Chrome or Chromium.
