package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/grokify/d2vision/export"
)

var (
	s2gOutput    string
	s2gFPS       float64
	s2gLoop      float64
	s2gWidth     int
	s2gHeight    int
	s2gNoSandbox bool
)

var svg2gifCmd = &cobra.Command{
	Use:   "svg2gif <animated.svg>",
	Short: "Convert an animated SVG to a looping GIF",
	Long: `Convert an animated SVG to a looping GIF, for platforms that need a GIF.

Most SVG rasterizers ignore SVG animation. This command loads the SVG in
headless Chrome, steps its animation clock (SMIL and CSS animations) one frame
at a time, and encodes the frames. Size comes from the SVG's width/height or
viewBox, and the loop length from its longest animation duration; override
either with flags.

Examples:
  d2vision svg2gif card.svg -o card.gif
  d2vision svg2gif diagram.svg -o diagram.gif --loop 4 --fps 20
`,
	Args: cobra.ExactArgs(1),
	RunE: runSVG2GIF,
}

func init() {
	f := svg2gifCmd.Flags()
	f.StringVarP(&s2gOutput, "output", "o", "", "output GIF file (required)")
	f.Float64Var(&s2gFPS, "fps", 25, "frames per second")
	f.Float64Var(&s2gLoop, "loop", 0, "loop length in seconds (default: longest animation in the SVG)")
	f.IntVar(&s2gWidth, "width", 0, "pixel width (default: from the SVG)")
	f.IntVar(&s2gHeight, "height", 0, "pixel height (default: from the SVG)")
	f.BoolVar(&s2gNoSandbox, "chrome-no-sandbox", false, "launch Chrome with --no-sandbox (for sandboxed environments)")
	if err := svg2gifCmd.MarkFlagRequired("output"); err != nil {
		panic(err) // programming error: the flag is defined just above
	}
}

func runSVG2GIF(cmd *cobra.Command, args []string) error {
	svg, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	info, err := export.InspectSVG(svg)
	if err != nil && (s2gWidth == 0 || s2gHeight == 0) {
		return err
	}
	if s2gWidth > 0 {
		info.Width = s2gWidth
	}
	if s2gHeight > 0 {
		info.Height = s2gHeight
	}
	if s2gLoop > 0 {
		info.Loop = s2gLoop
	}
	if info.Loop <= 0 {
		return errors.New("no animation duration found in the SVG; pass --loop")
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	n, err := writeGIF(ctx, svg, export.Options{Width: info.Width, Height: info.Height, NoSandbox: s2gNoSandbox}, info.Loop, s2gFPS, s2gOutput)
	if err != nil {
		return err
	}
	cmd.PrintErrf("wrote %s (%d frames, %dx%d, %.2fs loop)\n", s2gOutput, n, info.Width, info.Height, info.Loop)
	return nil
}

// writeGIF captures one loop of svg and writes it to path, returning the frame
// count. Shared by svg2gif and card render.
func writeGIF(ctx context.Context, svg []byte, opt export.Options, loop, fps float64, path string) (int, error) {
	opt.Scale = 1 // GIF size grows with the square of the scale
	frames, err := export.Frames(ctx, svg, loop, fps, opt)
	if err != nil {
		return 0, err
	}
	f, err := os.Create(path) // #nosec G304 -- caller-supplied output path
	if err != nil {
		return 0, err
	}
	if err := export.EncodeGIF(f, frames, export.GIFOptions{FPS: fps}); err != nil {
		_ = f.Close() // encode error takes precedence
		return 0, fmt.Errorf("encode gif: %w", err)
	}
	return len(frames), f.Close()
}
