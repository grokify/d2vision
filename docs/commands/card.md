# card

Render datasheet-style infographic cards: a fixed-size poster of stacked panels, each pairing a short text block with a simple node/edge diagram, or a comparison table. Edges can carry animated "flow" dots.

These cards are deliberately simpler than documentation diagrams: a handful of nodes per panel, explicit coordinates, one message per panel. For auto-laid-out diagrams use D2 (`generate`, `template`, `pipeline`).

## Usage

```bash
# SVG (animated with SMIL flow dots)
d2vision card render card.json --svg card.svg

# PNG screenshot 1.5 s into the loop, at 2x for crisp text (needs Chrome)
d2vision card render card.json --png card.png --at 1.5 --scale 2

# Looping animated GIF of one full loop (needs Chrome)
d2vision card render card.json --gif card.gif --fps 25
```

All outputs can be requested in one call. `--svg --animate=false` writes a fully static SVG. PNG and GIF are captured from the animated SVG by stepping the SMIL clock in headless Chrome, so a given `--at` always produces the same image. To convert an existing animated SVG to a GIF, see [svg2gif](svg2gif.md).

If Chrome cannot start because the calling process is itself sandboxed (some CI containers and agent sandboxes), add `--chrome-no-sandbox`.

## Examples

Runnable examples are in [`examples/card/`](https://github.com/grokify/d2vision/tree/main/examples/card): a hand-drawn card with sequenced steps, and a card derived from a protocol file.

## Card definition

A card is JSON, parsed strictly: unknown fields are errors.

```json
{
  "loop": 6.5,
  "panels": [
    {
      "id": "flow",
      "height": 255,
      "title": "ID-JAG",
      "body": "One short paragraph.",
      "diagram": {
        "nodes": [
          {"id": "agent", "icon": "🤖", "label": "Agent", "x": 55, "y": 150},
          {"id": "idp", "icon": "🏢", "label": "Enterprise\nIdP", "x": 215, "y": 40}
        ],
        "edges": [
          {"from": "agent", "to": "idp", "bend": -20, "label": "exchange",
           "flows": [
             {"color": "#60a5fa", "delay": 0.9},
             {"mode": "reverse", "color": "#fbbf24", "delay": 1.7}
           ]}
        ]
      }
    }
  ],
  "legend": [{"color": "#60a5fa", "label": "ID token"}],
  "footer": "example"
}
```

| Element | Notes |
|---------|-------|
| Card | `width`/`height` default 800x1000. `loop` (0 = derived), `stepGap`, `loopRest`, `artifacts`. Panels stack top to bottom; heights must fit (a footer row is reserved when `legend` or `footer` is set). |
| Panel | `kind`: `diagram` (default) or `table`. `textSide` `left`/`right`, `center` for hero panels, `ruleBelow` for a thick separator. |
| Node | `shape`: `circle` (icon in a circle, label below), `box`, `pill`. `x`,`y` is the center, relative to the panel's diagram area. `icon` is an emoji or short text; append `U+FE0F` (for example `⚙️`) so symbols that default to a text glyph render as color emoji. |
| Edge | `bend` curves the line (px, sign picks the side). `style`: `solid`/`dashed`. `arrow` adds a head. `label` draws a pill at the midpoint. |
| Flow | `mode`, `color`, `size`, `speed` (px/s), `count`, `gap`, `delay`, `pause`, `step`. `flow` is shorthand for a single entry in `flows`. |

### Flow modes

| Mode | Meaning |
|------|---------|
| `forward` | dots travel From to To |
| `reverse` | dots travel To to From |
| `alternate` | one dot goes out, rests, and returns: request/response |
| `both` | independent dots each way, deliberately out of phase: full duplex |

Every flow's animation has the card's loop as its period, so the whole card repeats seamlessly. Rendering fails with a clear error if a flow does not fit in an explicit `loop`, text overflows its panel, a table overflows, or a node lies outside its diagram area.

### Steps and timing

Give flows a `step` to sequence a story instead of hand-tuning delays:

- Within a panel, steps run in ascending order. A step starts when the previous one has fully finished, plus `stepGap` (default 0.2 s).
- Flows sharing a step number start together.
- `delay` on a stepped flow is an offset within its step.
- Each panel is its own timeline starting at zero, so panels animate side by side.
- A numbered badge is drawn on the edge for every stepped flow, so the order is readable in a static PNG.
- Leave `loop` unset and it is derived: the last finish time plus `loopRest` (default 1 s), rounded up to 50 ms. Set `loop` explicitly and rendering fails if a flow does not fit.

Flows without a `step` keep the old behavior: they start at their `delay` from the top of the loop.

### Deriving a panel from a protocol file

Instead of a hand-drawn `diagram`, a panel can have a `source` that reads a [PIDL](https://github.com/grokify/pidl) protocol definition. You choose which actors and messages to show and how; the protocol file decides who sends each message to whom, and in what order.

```json
{
  "artifacts": [
    {"id": "id-jag", "label": "ID-JAG", "color": "#fbbf24"},
    {"id": "access-token", "label": "Access token", "color": "#34d399"}
  ],
  "panels": [{
    "id": "idjag", "height": 255, "title": "ID-JAG",
    "source": {
      "pidl": "../../idjag/pidl/idjag_delegation.json",
      "actors": [
        {"entity": "agent", "icon": "🤖", "x": 55, "y": 150},
        {"entity": "idp_auth_server", "icon": "🏢", "label": "Enterprise\nIdP", "x": 215, "y": 40}
      ],
      "steps": [
        {"action": "request_idjag", "artifact": "id-jag"},
        {"action": "return_idjag", "artifact": "id-jag"}
      ],
      "edges": [{"between": ["agent", "idp_auth_server"], "bend": -20}]
    }
  }]
}
```

- `pidl` is relative to the card file. Actor labels default to the PIDL entity name.
- Each step's `action` is a PIDL flow. Its direction comes from the file, so an edge can never point the wrong way. Steps are numbered in the order you list them.
- Steps must keep the protocol's order, and each action can be used once. A step that is not in the file, that happens inside one entity (no edge to draw), or whose entity has no actor is an error.
- `artifact` picks a dot color from the card's `artifacts`. When `legend` is empty, the legend is derived from the artifacts actually used.
- A step may instead give `from` and `to` (actor ids) for something the protocol file does not model, such as a human approving out of band. It is drawn, but the CLI prints a warning that it is not backed by the protocol file. An actor with no `entity` can only take part in such steps.
- `panel.diagram` and `panel.source` are mutually exclusive.

### Stable IDs

The SVG uses stable IDs so it can be post-processed: `panel-<panel>`, `node-<panel>-<node>`, `edge-<panel>-<edge>` and `path-<panel>-<edge>` (edge ids default to `<from>-<to>`).

## Library

```go
c, err := card.ParseFile("card.json")
svg, err := c.Render(card.Options{Animate: true})
png, err := export.PNG(ctx, svg, 1.5, export.Options{Width: c.Width, Height: c.Height, Scale: 2})
frames, err := export.Frames(ctx, svg, c.Loop, 25, export.Options{Width: c.Width, Height: c.Height})
err = export.EncodeGIF(w, frames, export.GIFOptions{FPS: 25})
```

`export.EncodeGIF` uses one shared palette without dithering and stores only the changed region of each frame, with unchanged pixels transparent, so a card with a few moving dots is a few hundred KB.
