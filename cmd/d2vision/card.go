package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/grokify/d2vision/card"
	"github.com/grokify/d2vision/export"
)

var (
	cardNoSandbox bool
	cardSVG       string
	cardPNG       string
	cardGIF       string
	cardAnimate   bool
	cardAt        float64
	cardScale     float64
	cardFPS       float64
)

var cardCmd = &cobra.Command{
	Use:   "card",
	Short: "Render datasheet-style infographic cards",
}

var cardRenderCmd = &cobra.Command{
	Use:   "render <card.json>",
	Short: "Render a card definition to SVG, PNG and/or animated GIF",
	Long: `Render a datasheet card from a JSON definition.

A card is a fixed-size poster of stacked panels; each panel pairs a text block
with a simple node/edge diagram, or is a comparison table. Edges may define a
"flow" (animated dots). The same definition produces every output:

  --svg   SVG; animated with SMIL flow dots unless --animate=false
  --png   PNG screenshot at --at seconds into the loop (needs Chrome)
  --gif   looping animated GIF of one full loop (needs Chrome)

Examples:
  d2vision card render card.json --svg card.svg
  d2vision card render card.json --png card.png --at 1.5 --scale 2
  d2vision card render card.json --gif card.gif --fps 25
`,
	Args: cobra.ExactArgs(1),
	RunE: runCardRender,
}

func init() {
	f := cardRenderCmd.Flags()
	f.StringVar(&cardSVG, "svg", "", "write SVG to this file")
	f.StringVar(&cardPNG, "png", "", "write PNG to this file")
	f.StringVar(&cardGIF, "gif", "", "write animated GIF to this file")
	f.BoolVar(&cardAnimate, "animate", true, "include animated flow dots in the SVG")
	f.Float64Var(&cardAt, "at", 1.5, "seconds into the loop to capture for --png")
	f.Float64Var(&cardScale, "scale", 1, "device scale factor for --png (2 = retina)")
	f.Float64Var(&cardFPS, "fps", 25, "frames per second for --gif")
	f.BoolVar(&cardNoSandbox, "chrome-no-sandbox", false, "launch Chrome with --no-sandbox (for sandboxed environments)")
	cardCmd.AddCommand(cardRenderCmd)
}

func runCardRender(cmd *cobra.Command, args []string) error {
	if cardSVG == "" && cardPNG == "" && cardGIF == "" {
		return fmt.Errorf("nothing to do: set at least one of --svg, --png, --gif")
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	c, err := card.ParseFile(args[0])
	if err != nil {
		return err
	}
	for _, w := range c.Warnings() {
		cmd.PrintErrf("warning: %s\n", w)
	}

	// PNG and GIF capture the animated SVG so dots appear; --animate only
	// affects what is written to --svg.
	animated, err := c.Render(card.Options{Animate: true})
	if err != nil {
		return err
	}
	if cardSVG != "" {
		out := animated
		if !cardAnimate {
			if out, err = c.Render(card.Options{}); err != nil {
				return err
			}
		}
		if err := os.WriteFile(cardSVG, out, 0o644); err != nil { // #nosec G306 -- generated asset
			return err
		}
		cmd.PrintErrf("wrote %s\n", cardSVG)
	}

	opt := export.Options{Width: c.Width, Height: c.Height, Scale: cardScale, NoSandbox: cardNoSandbox}
	if cardPNG != "" {
		data, err := export.PNG(ctx, animated, cardAt, opt)
		if err != nil {
			return err
		}
		if err := os.WriteFile(cardPNG, data, 0o644); err != nil { // #nosec G306 -- generated asset
			return err
		}
		cmd.PrintErrf("wrote %s\n", cardPNG)
	}
	if cardGIF != "" {
		n, err := writeGIF(ctx, animated, opt, c.LoopSeconds(), cardFPS, cardGIF)
		if err != nil {
			return err
		}
		cmd.PrintErrf("wrote %s (%d frames)\n", cardGIF, n)
	}
	return nil
}
