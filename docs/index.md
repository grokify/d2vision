# d2vision

Tools for [D2](https://d2lang.com) diagram parsing, generation, and AI-assisted creation.

## Overview

d2vision provides a complete toolkit for working with D2 diagrams:

- **Parse**: Extract structure from D2-generated SVGs
- **Generate**: Create D2 code from structured specifications
- **Template**: Quick-start patterns for common diagrams
- **Learn**: Reverse engineer D2 code from existing SVGs
- **Lint**: Check D2 files for common layout issues
- **Diff**: Compare two diagrams
- **Watch**: Auto-render D2 files on changes
- **Analyze**: Analyze layout and provide generation hints
- **Icons**: Browse and search D2's icon library (185+ SVG icons)

## Why d2vision?

D2 is a powerful diagram language, but creating well-laid-out diagrams requires understanding its layout engine quirks. d2vision helps by:

1. **Providing templates** - Start with proven patterns instead of from scratch
2. **Linting D2 files** - Catch layout issues before rendering
3. **Learning from examples** - Reverse engineer existing diagrams to understand what works
4. **Token-efficient output** - TOON format uses ~40% fewer tokens than JSON, ideal for LLM consumption

## Quick Start

```bash
# Install
go install github.com/grokify/d2vision/cmd/d2vision@latest

# Generate a diagram from a template
d2vision template microservices --d2 | d2 - microservices.svg

# Parse an existing SVG
d2vision parse diagram.svg

# Lint a D2 file
d2vision lint diagram.d2
```

## For AI Assistants

d2vision is designed to help AI assistants create D2 diagrams effectively:

- **TOON format** reduces token usage by ~40%
- **Templates** provide starting points for common patterns
- **Lint** catches issues that cause poor layouts
- **Learn** helps understand existing diagram structures
- **Analyze** provides layout insights and generation hints
- **Parse --for-generation** outputs D2 code skeletons for recreation

```bash
# AI workflow: template → modify → render
d2vision template network-boundary --d2 > diagram.d2
# ... AI modifies diagram.d2 ...
d2 diagram.d2 output.svg

# AI workflow: analyze existing → recreate
d2vision parse existing.svg --for-generation
# ... use insights and skeleton to create new diagram ...
```
