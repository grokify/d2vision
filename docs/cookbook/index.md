# Cookbook Overview

This cookbook covers common D2 diagram patterns and solutions to layout challenges.

## The Layout Problem

D2 uses layout engines (ELK, dagre) that automatically position elements. While this is convenient, it can produce unexpected results:

- Containers stack vertically when you want them side-by-side
- Elements don't align as expected
- Cross-container edges cause layout shifts

This cookbook teaches you how to control D2's layout behavior.

## Quick Reference

| Problem | Solution |
|---------|----------|
| Containers stack vertically | Use `grid-columns: N` |
| Elements flow wrong direction | Use `direction: right/down` |
| Need horizontal grouping in vertical layout | Use invisible container |
| Cross-container edges cause stacking | Add `grid-columns` at root |
| Alignment issues | Use grid with explicit dimensions |

## Cookbook Sections

### [Layout Fundamentals](layout-fundamentals.md)

Core concepts: direction, grid-columns, grid-rows, and how the layout engine works.

### [Side-by-Side Layouts](side-by-side.md)

How to place containers horizontally when they want to stack vertically.

### [Container Patterns](containers.md)

Nesting, invisible containers, and container styling.

### [Cross-Container Edges](cross-container-edges.md)

Managing edges between containers without breaking layout.

### [Sequence Diagrams](sequence-diagrams.md)

Actors, messages, spans, and groups.

### [SQL Tables](sql-tables.md)

Entity-relationship diagrams with proper relationships.

### [Styling](styling.md)

Colors, borders, fonts, and classes.

## Key Insight

**The layout engine optimizes for edge routing.** When you have edges between containers, the engine positions containers to minimize edge crossings and lengths. This often means vertical stacking.

To override this behavior, use `grid-columns` or `grid-rows` to explicitly control top-level positioning. The layout engine will then only optimize within each cell.
