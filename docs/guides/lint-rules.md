# D2 Lint Rules

`d2vision lint <file.d2>` runs deterministic, static checks over a D2 diagram
and exits non-zero when any issue is found. Each finding carries a stable
**code** — the machine contract used by CI filters, suppressions, and
dashboards (think `net/http` status constants). The codes are defined in Go in
`lint/code.go`; the same registry backs `--explain` and this guide.

## Usage

```sh
d2vision lint diagram.d2                 # lint one file (exit 1 on findings)
d2vision lint diagram.d2 --format json   # machine-readable, for CI
d2vision lint diagram.d2 --config .d2vision.yaml
d2vision lint --list                     # list every rule (code, severity, title)
d2vision lint --explain corner-near      # full remediation for one code
```

Severities: **error** fails the run; **warning** and **info** are advisory
(the CLI still exits non-zero if any issue is present).

## Configuration

A single YAML config declares which rules run and how, so one CLI pass applies a
whole policy (in the spirit of golangci-lint). It is resolved in this order:
`--config <path>`, else the nearest `.d2vision.yaml` walking up from the target
file, else the built-in defaults. A partial config only overrides what it names.

```yaml
rules:
  text-overlap:          # opt-in: runs a full layout pass
    enabled: true
    settings:
      layout: elk        # check the engine you actually render with
  corner-near:
    severity: warning    # override the default severity
  deep-nesting:
    enabled: false       # turn a rule off
```

Per rule: `enabled` (bool), `severity` (`error|warning|info`), and rule-specific
`settings`. Unset fields fall back to the registry defaults below. All enabled
rules run in a single invocation.

## Rules

### corner-near

**Corner-pinned element** · severity: `error`

An object sets `near` to a corner constant (`top-left`, `top-right`,
`bottom-left`, `bottom-right`). A corner pin places the element in one quadrant
of the canvas while the graph body flows into another, so the two sit
diagonally apart and roughly half the canvas is empty.

**Remediation:** use one of the four edge-center constants instead —
`top-center`, `bottom-center`, `center-left`, `center-right` — so the element
stacks cleanly above/below/left/right of the diagram and shares an axis with it.
Choose the edge on the diagram's **short axis** to minimize whitespace: a tall
(top-down) flow → `center-left`/`center-right`; a wide flow →
`top-center`/`bottom-center`.

Detection is structural (via the D2 compiler's constant-`near` resolution), so
an object literally named `top-left` is not mis-flagged.

### text-overlap

**Overlapping text** · severity: `warning` · opt-in

A connection label's rendered rectangle overlaps another label or an unrelated
node, making the text hard to read. Unlike the other rules this is a
**post-layout** check: it runs a full layout pass, so it is **engine-dependent**
(the same source can overlap under `dagre` yet be clean under `elk` — common when
several edges share a node pair) and **opt-in** via config.

Enable it and pick the engine you render with:

```yaml
rules:
  text-overlap:
    enabled: true
    settings:
      layout: elk   # or dagre
```

**Remediation:** switch the layout engine (ELK spaces parallel-edge labels
better), reduce parallel edges between the same node pair (or add an
intermediate node), or shorten/relocate the label.

### cross-container-edge

**Cross-container edge** · severity: `warning`

An edge connects nodes in different top-level containers, which can force the
layout engine to stack the containers vertically.

**Remediation:** add `grid-columns: N` at the root to control the arrangement,
or reconsider whether the edge should cross the boundary.

### duplicate-node

**Duplicate node definition** · severity: `warning`

A node id is defined more than once; later definitions silently merge into the
first.

**Remediation:** consolidate the attributes into a single definition.

### deep-nesting

**Deep container nesting** · severity: `info`

Container nesting deeper than ~3 levels can slow layout and hurt readability.

**Remediation:** flatten the hierarchy or split the diagram.

### missing-grid

**Multiple root containers without a grid** · severity: `info`

Several root-level containers with no `grid-columns` leave their arrangement to
the layout engine, which often stacks them.

**Remediation:** add `grid-columns: N` at the root to place them deliberately.

### mixed-directions

**Mixed direction settings** · severity: `info`

Different `direction` values across the graph can produce an inconsistent flow.

**Remediation:** prefer one direction unless a sub-graph genuinely needs its own.

### compile-skipped

**Structural checks skipped** · severity: `info`

The D2 did not compile, so graph-based checks were skipped; only the text-based
checks ran.

**Remediation:** fix the reported compile error and re-lint.

---

Detection is deterministic and needs no LLM. Correcting a finding — e.g.
choosing which edge to move a legend to — is where an agent such as
`legend-reviewer` helps.
