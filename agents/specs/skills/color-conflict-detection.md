---
name: color-conflict-detection
description: Detects color conflicts in diagram legends where different categories use the same or similar colors
triggers:
  - color conflict
  - legend colors
  - color clarity
dependencies:
  - d2vision
---

# Color Conflict Detection Skill

Detect when multiple legend categories use the same or similar colors, causing visual confusion.

## Detection Algorithm

### 1. Extract Legend Colors

Parse D2 file for style definitions in legend containers:

```d2
legend: Legend {
  stride: STRIDE {
    s: S { style.fill: "#ffebee"; style.stroke: "#c62828" }  # Extract these
  }
  mitre: ATT&CK {
    ta0010: TA0010 { style.fill: "#ffcdd2"; style.stroke: "#b71c1c" }  # And these
  }
}
```

### 2. Color Similarity Check

Colors are considered similar if:
- **Exact match**: Same hex code
- **Near match**: Within 30 units in RGB space

```
distance = sqrt((r1-r2)² + (g1-g2)² + (b1-b2)²)
similar = distance < 30
```

### 3. Conflict Categories

| Conflict Type | Severity | Example |
|---------------|----------|---------|
| Exact match across sections | Error | STRIDE.S and Assets.crown both #ffcdd2 |
| Near match across sections | Warning | #ffcdd2 vs #ffebee (both red family) |
| Same color within section | Info | Acceptable if items are related |

### 4. Color Family Classification

Map colors to semantic families:

| Family | Hue Range | Common Uses |
|--------|-----------|-------------|
| Red | 0-30, 330-360 | Danger, threats, critical |
| Orange | 30-60 | Warnings, high value |
| Yellow | 60-90 | Caution, attention |
| Green | 90-150 | Safe, approved |
| Teal | 150-190 | Data, neutral |
| Blue | 190-250 | Information, process |
| Purple | 250-290 | Special, collection |
| Pink | 290-330 | Credential access |

### 5. Cross-Section Analysis

Check for semantic conflicts:

```
For each color C in section A:
  For each color C' in section B:
    If similar(C, C'):
      If semantically_related(A, B):
        severity = "info"  # OK for related concepts
      Else:
        severity = "error"  # Confusing for unrelated concepts
```

## Usage

Parse D2 source and check for conflicts:

```bash
# Extract styles
grep -E "style\.(fill|stroke):" diagram.d2

# Compare with expected palette
```

## Output

```json
{
  "color_conflicts": {
    "exact_matches": [
      {
        "color": "#ffcdd2",
        "locations": [
          {"section": "STRIDE", "item": "S", "property": "fill"},
          {"section": "Assets", "item": "crown", "property": "fill"}
        ],
        "severity": "error",
        "message": "Same red fill used for STRIDE threats and Crown Jewel assets"
      }
    ],
    "near_matches": [
      {
        "colors": ["#ffebee", "#ffcdd2"],
        "distance": 22,
        "family": "red",
        "locations": [
          {"section": "STRIDE", "item": "S"},
          {"section": "ATT&CK", "item": "TA0010"}
        ],
        "severity": "warning",
        "message": "Both use red family - may be confused"
      }
    ],
    "recommendations": [
      {
        "section": "ATT&CK",
        "current_family": "red",
        "suggested_family": "blue",
        "rationale": "Use blue gradient for attack progression, distinct from STRIDE red"
      },
      {
        "section": "Assets",
        "current_family": "red",
        "suggested_family": "gold",
        "rationale": "Use gold for value classification, distinct from threat colors"
      }
    ]
  }
}
```

## Recommended Color Assignments

To avoid conflicts in threat modeling diagrams:

| Category | Recommended Family | Hex Range |
|----------|-------------------|-----------|
| STRIDE Spoofing | Red | #ffebee, #c62828 |
| STRIDE Elevation | Green | #e8f5e9, #2e7d32 |
| STRIDE Info Disclosure | Blue | #e3f2fd, #1565c0 |
| ATT&CK Tactics | Blue gradient | #e3f2fd → #283593 |
| Crown Jewel Assets | Gold | #fff8e1, #ff8f00 |
| High Value Assets | Teal | #e0f2f1, #00897b |
| Malicious Actors | Red | #ffcdd2, #b71c1c |
| Trust Boundaries | Purple | #f3e5f5, #7b1fa2 |
