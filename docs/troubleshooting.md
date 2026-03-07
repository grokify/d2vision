# Troubleshooting

Common issues and solutions when working with d2vision.

## Pipeline Command Issues

### Disconnected graph with parallel stages

**Symptom**: Stages inside a `parallel` block appear as isolated clusters with no edges connecting them to downstream stages.

**Cause**: In earlier versions, when a downstream stage referenced parallel substages via `joinFrom`, d2vision emitted unqualified IDs (`substage_a -> downstream`) instead of qualified IDs (`parent.substage_a -> downstream`). D2 treats unqualified names as new top-level nodes.

**Solution**: Upgrade to d2vision v0.1.0 or later. This was fixed to properly qualify nested stage references.

**Example of correct output**:

```d2
# Parallel stages are nested inside parent
split: Split {
  process_a: Process A { ... }
  process_b: Process B { ... }
}

# Fan-in uses qualified paths
split.process_a -> merge
split.process_b -> merge
```

---

## SVG Rendering Issues

### SVG dimensions don't match content

**Symptom**: Rendered SVG has excessive whitespace or content is clipped.

**Cause**: D2's layout engine determines dimensions based on content. Complex diagrams with many containers may have unexpected sizing.

**Solution**: Use the `rotate` command to adjust orientation, or post-process the SVG to adjust the viewBox:

```bash
# Rotate landscape to portrait
d2vision rotate diagram.svg --angle 90 -o rotated.svg
```

---

## PDF Generation

When embedding d2vision-generated SVGs in PDFs, consider these approaches.

### Converting SVG to PNG for PDF embedding

LaTeX engines handle PNG more reliably than SVG. Use `rsvg-convert` for high-quality conversion:

```bash
# Generate SVG
d2vision pipeline spec.json --svg -o diagram.svg

# Convert to 300dpi PNG
rsvg-convert -d 300 -p 300 diagram.svg -o diagram.png
```

### Rotating diagrams for portrait PDFs

Landscape diagrams can be rotated for portrait PDF pages with left-side binding:

```bash
# Rotate 90° counter-clockwise (top goes to binding side)
d2vision rotate diagram.svg --angle 90 -o portrait.svg
```

### Font issues with special characters

If your PDF workflow reports missing glyphs for arrow characters (`→`) or other Unicode symbols, replace them with ASCII equivalents in your source files or configure font fallbacks in your PDF engine.

---

## Format Conversion Issues

### Mermaid/PlantUML conversion warnings

**Symptom**: Conversion succeeds but reports unsupported features.

**Cause**: Not all Mermaid/PlantUML features have D2 equivalents. The converter skips unsupported features and reports them.

**Solution**: Use `--lint-only` to check compatibility before converting:

```bash
d2vision convert --lint-only diagram.mmd
```

Unsupported features are listed with suggestions for manual adjustment.

### Auto-detection picks wrong format

**Symptom**: Converter misidentifies the source format.

**Solution**: Explicitly specify the format:

```bash
d2vision convert --from mermaid diagram.txt
d2vision convert --from plantuml diagram.txt
```

---

## Parse Command Issues

### Empty or minimal output from parse

**Symptom**: Parsing an SVG returns few or no nodes.

**Cause**: d2vision parses D2-generated SVGs by decoding base64-encoded IDs from CSS class names. SVGs from other tools (Mermaid, PlantUML, hand-drawn) don't have this encoding.

**Solution**: Only use `parse` with SVGs rendered by D2. For other diagram formats, use the `convert` command to translate to D2 first.

### Shape detection returns "unknown"

**Symptom**: Nodes have `shape: unknown` instead of the expected shape type.

**Cause**: D2 doesn't embed explicit shape metadata in SVGs. d2vision infers shapes from geometry (aspect ratio, path patterns). Complex or custom shapes may not be recognized.

**Solution**: This is expected behavior for non-standard shapes. The node bounds and connectivity are still parsed correctly.

---

## CLI Issues

### Command not found after installation

**Symptom**: `d2vision: command not found` after `go install`.

**Solution**: Ensure `$GOPATH/bin` (or `$GOBIN`) is in your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Large binary size

**Symptom**: The d2vision binary is ~34MB.

**Cause**: d2vision embeds the D2 rendering library for self-contained operation. This includes layout engines, font handling, and SVG rendering.

**Solution**: This is expected. The tradeoff is no external D2 CLI dependency.

---

## Getting Help

If you encounter an issue not covered here:

1. Check the [GitHub Issues](https://github.com/grokify/d2vision/issues) for existing reports
2. Run with verbose output where available
3. Open a new issue with:
   - d2vision version (`d2vision --version`)
   - Command and flags used
   - Input file (or minimal reproduction)
   - Expected vs actual output
