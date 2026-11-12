# legibility

Check whether a diagram's labels can be read where it is shown: on a Word page,
a slide, or a screen.

## Usage

```bash
d2vision legibility <file.svg> [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, --target` | all | Target to check, repeatable |
| `--width`, `--height` | | A custom target's area, in inches |
| `--min-font` | `9` | A custom target's smallest readable label, in points |
| `--check` | `false` | Exit with status 1 if a label is too small on a target or an arrow is drawn over a box |
| `--show` | `5` | Too-small labels to list per target in text output |
| `-f, --format` | `text` | Output format: text, toon, json |

## How It Works

A diagram is scaled to fit the area it is shown in, keeping its shape. A wide
diagram on a narrow page shrinks, and its labels shrink with it. The command
reads each label's font size from the SVG, works out the size it is shown at
on each target, and compares it with the smallest size that reads comfortably
there.

| Target | Area | Smallest label |
|--------|------|----------------|
| `docx-portrait` | Word, US Letter portrait, 1-inch margins: 6.5 x 9 in | 9 pt |
| `docx-landscape` | Word, US Letter landscape, 1-inch margins: 9 x 6.5 in | 9 pt |
| `tabloid-landscape` | Tabloid landscape, 1-inch margins: 15 x 9 in | 9 pt |
| `slide-16x9` | 16:9 slide below a title: 12.33 x 5.9 in | 18 pt |
| `screen` | Browser window, 1280 x 720 CSS pixels, below a title | 9 pt (12 px) |

For a diagram that is not readable, the advice says:

- How many times too wide or tall it is, and the largest size it can be with its smallest label still readable
- To wrap labels when a line is longer than 40 characters, to about 28
- To use fewer columns, or fewer items in each column
- To raise the font size, when the diagram is less than one and a half times too large; otherwise to split it
- Which other targets it is already readable on

It also lists every arrow drawn over a box that is neither of its ends, which
hides the box and makes the arrow hard to follow. A grid draws arrows straight,
over whatever is in the way; a layout engine that routes them, such as ELK
(`d2 --layout elk`), draws them around boxes. `--check` fails on either a label
that is too small or an arrow over a box.

Edge labels are not counted, since the SVG does not record their size.

## Example

```bash
d2vision legibility theory-of-change.svg --target docx-portrait --show 1
```

```text
theory-of-change.svg: 4574 x 1138 pixels, 35 labels, longest line 110 characters

docx-portrait (Word, US Letter portrait, 1-inch margins, from 9 pt): ✗ 35 of 35 labels too small
  shown at 14% (fits the width); smallest label 1.6 pt, median 1.6 pt
  - activities.int-capacity-programs: 1.6 pt ("Deliver capacity building trainings, workshops, a…")
  … and 34 more
  → The diagram is 4574 x 1138 pixels and 5.5 times too wide: to show its 16-pixel labels at 9 points on docx-portrait it must fit within 832 x 1152.
  → Wrap labels to about 28 characters a line: the longest line has 110, and long lines make wide boxes.
  → Use fewer columns, or a layout whose shape is closer to the target's.
  → Split the diagram into smaller ones, such as one per group of items.
```

## In Go

```go
import (
    "github.com/grokify/d2vision"
    "github.com/grokify/d2vision/legibility"
)

d, err := d2vision.ParseFile("diagram.svg")
t, _ := legibility.TargetByName("docx-portrait")
report, err := legibility.Analyze(d, t)
if !report.Clear() {
    fmt.Println(report.Advice)
}
crossings := legibility.Crossings(d) // arrows over boxes, on any target

label := legibility.Wrap("Deliver capacity building trainings and workshops", 28)
```

`legibility.Wrap` breaks a label into lines at spaces. The `generate` package
keeps the line breaks, so wrapped labels make taller, narrower boxes. The
`legibility` package does not link the D2 engine.
