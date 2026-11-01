# watch

Watch D2 files and automatically re-render on changes.

## Usage

```bash
d2vision watch <file.d2> [output.svg] [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `<input>.svg` | Output SVG file |
| `--lint` | `false` | Run d2vision lint before rendering |
| `--debounce` | `100` | Debounce delay in milliseconds |
| `--d2-args` | | Additional arguments to pass to d2 |
| `--on-success` | | Command to run after successful render |

## Basic Usage

```bash
# Watch a file (outputs to diagram.svg)
d2vision watch diagram.d2

# Watch with explicit output
d2vision watch diagram.d2 output.svg

# Using -o flag
d2vision watch diagram.d2 -o output.svg
```

## Features

### Automatic Re-rendering

The watch command monitors your D2 file for changes and automatically re-renders when you save:

```
$ d2vision watch diagram.d2
Watching diagram.d2 → diagram.svg
Press Ctrl+C to stop

[09:15:32] Rendering diagram.d2...
[09:15:32] ✓ Rendered diagram.svg (245ms)
[09:16:45] Rendering diagram.d2...
[09:16:45] ✓ Rendered diagram.svg (198ms)
```

### Linting Before Render

Check for layout issues before rendering:

```bash
d2vision watch diagram.d2 --lint
```

Output with lint warnings:

```
[09:15:32] Linting diagram.d2...
[09:15:32] ⚠ Found 2 issue(s):
  Line 5: Cross-container edge may cause vertical stacking
  Line 12: Consider adding grid-columns for side-by-side layout
[09:15:32] Rendering diagram.d2...
[09:15:32] ✓ Rendered diagram.svg (245ms)
```

### Custom D2 Arguments

Pass additional arguments to the d2 renderer:

```bash
# Use a specific theme
d2vision watch diagram.d2 --d2-args="--theme=200"

# Use dark theme with specific layout engine
d2vision watch diagram.d2 --d2-args="--theme=200 --layout=elk"

# Multiple arguments
d2vision watch diagram.d2 --d2-args="--sketch --pad=50"
```

### Post-Render Hooks

Run a command after successful renders:

```bash
# Open the SVG (macOS)
d2vision watch diagram.d2 --on-success="open %s"

# Refresh browser (requires browser-sync or similar)
d2vision watch diagram.d2 --on-success="browser-sync reload"

# Copy to clipboard (macOS)
d2vision watch diagram.d2 --on-success="cat %s | pbcopy"

# Notify on completion
d2vision watch diagram.d2 --on-success="osascript -e 'display notification \"Diagram updated\"'"
```

The `%s` placeholder is replaced with the output file path.

### Debouncing

Prevent multiple renders when files are saved rapidly:

```bash
# Longer debounce for slower systems
d2vision watch diagram.d2 --debounce=500

# Shorter debounce for quick feedback
d2vision watch diagram.d2 --debounce=50
```

## Workflow Examples

### Development Workflow

```bash
# Terminal 1: Watch and render
d2vision watch architecture.d2 --lint

# Terminal 2: Edit with your favorite editor
vim architecture.d2
```

### Live Preview Setup

```bash
# Watch with browser auto-open (macOS)
d2vision watch diagram.d2 --on-success="open %s"

# With live reload server
npx live-server --watch=output.svg &
d2vision watch diagram.d2 -o output.svg
```

### CI-like Local Validation

```bash
# Lint + render with strict checking
d2vision watch diagram.d2 --lint --d2-args="--bundle=false"
```

## Error Handling

When rendering fails, the watch command continues running:

```
[09:15:32] Rendering diagram.d2...
error: failed to compile diagram.d2:1:1: syntax error
[09:15:32] ✘ Render failed (45ms)
[09:16:10] Rendering diagram.d2...
[09:16:10] ✓ Rendered diagram.svg (198ms)
```

## Tips

1. **Use with split terminal**: Keep your editor and watch output visible simultaneously
2. **Enable lint mode**: Catch layout issues early with `--lint`
3. **Set up auto-refresh**: Use `--on-success` with a browser refresh command for live preview
4. **Adjust debounce**: If your editor auto-saves frequently, increase the debounce value
