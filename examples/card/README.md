# Card Examples

Datasheet-style infographic cards rendered with `d2vision card render`.

| File | Description |
|------|-------------|
| [`request-flow.card.json`](request-flow.card.json) | Hand-drawn diagram: sequenced steps with numbered badges, an alternating request/response edge, and derived loop timing |
| [`from-protocol.card.json`](from-protocol.card.json) | Diagram derived from a protocol file: [`gateway.pidl.json`](gateway.pidl.json) decides direction and order |

## Render

```bash
# Animated SVG
d2vision card render examples/card/request-flow.card.json --svg request-flow.svg

# PNG (needs Chrome) and looping GIF (needs Chrome)
d2vision card render examples/card/request-flow.card.json \
  --png request-flow.png --scale 2 --at 1.5 --gif request-flow.gif

# Or convert an animated SVG you already have
d2vision svg2gif request-flow.svg -o request-flow.gif
```

Open the `.svg` in a browser to see the dots move. See the [card command guide](../../docs/commands/card.md) for the full definition format.
