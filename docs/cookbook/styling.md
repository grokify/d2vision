# Styling

D2 provides various styling options to customize the appearance of your diagrams.

## Fill Color

```d2
node: Node {
  style.fill: "#e3f2fd"
}
```

Common colors:

```d2
success: Success { style.fill: "#c8e6c9" }
warning: Warning { style.fill: "#fff3e0" }
error: Error { style.fill: "#ffcdd2" }
info: Info { style.fill: "#e3f2fd" }
```

## Stroke (Border)

```d2
node: Node {
  style.stroke: "#1976d2"
  style.stroke-width: 2
}
```

### Dashed Borders

```d2
planned: Planned {
  style.stroke-dash: 5
}
```

## Border Radius

```d2
rounded: Rounded {
  style.border-radius: 10
}
```

## Opacity

```d2
faded: Faded {
  style.opacity: 0.5
}
```

## Font Styling

```d2
node: Node {
  style.font-size: 20
  style.font-color: "#333"
  style.bold: true
  style.italic: true
}
```

## Shadow

```d2
elevated: Elevated {
  style.shadow: true
}
```

## Multiple Styles

```d2
styled: Styled Node {
  style.fill: "#e8f5e9"
  style.stroke: "#4caf50"
  style.stroke-width: 2
  style.border-radius: 8
  style.shadow: true
}
```

## Edge Styling

```d2
a -> b: {
  style.stroke: "#f44336"
  style.stroke-width: 3
}
```

### Animated Edges

```d2
a -> b: {
  style.animated: true
}
```

### Dashed Edges

```d2
a -> b: {
  style.stroke-dash: 5
}
```

## Classes

Define reusable styles with classes:

```d2
classes: {
  primary: {
    style.fill: "#1976d2"
    style.font-color: white
  }
  success: {
    style.fill: "#4caf50"
    style.font-color: white
  }
  danger: {
    style.fill: "#f44336"
    style.font-color: white
  }
}

button1: Primary Button { class: primary }
button2: Success Button { class: success }
button3: Danger Button { class: danger }
```

### Multiple Classes

```d2
classes: {
  rounded: {
    style.border-radius: 20
  }
  elevated: {
    style.shadow: true
  }
}

fancy: Fancy { class: [rounded, elevated] }
```

## Container Styling

Containers can be styled like nodes:

```d2
public: Public Zone {
  style.fill: "#e8f5e9"
  style.stroke: "#4caf50"
  style.stroke-width: 2

  server: Server
}

private: Private Zone {
  style.fill: "#fff3e0"
  style.stroke: "#ff9800"
  style.stroke-width: 2

  database: Database
}
```

## Pattern: Status Colors

```d2
classes: {
  running: {
    style.fill: "#c8e6c9"
    style.stroke: "#4caf50"
  }
  stopped: {
    style.fill: "#ffcdd2"
    style.stroke: "#f44336"
  }
  pending: {
    style.fill: "#fff3e0"
    style.stroke: "#ff9800"
  }
}

service1: Service 1 { class: running }
service2: Service 2 { class: stopped }
service3: Service 3 { class: pending }
```

## Pattern: Network Zones

```d2
dmz: DMZ {
  style.fill: "#fff8e1"
  style.stroke: "#ffc107"
  style.stroke-width: 2
  style.stroke-dash: 5

  lb: Load Balancer
}

internal: Internal {
  style.fill: "#e8f5e9"
  style.stroke: "#4caf50"
  style.stroke-width: 2

  app: Application
  db: Database
}
```

## Pattern: Highlighted Path

```d2
a -> b -> c -> d

# Highlight the critical path
b -> c: {
  style.stroke: "#f44336"
  style.stroke-width: 3
  style.animated: true
}
```

## Pattern: Deprecated Elements

```d2
current: Current API {
  style.fill: "#e3f2fd"
}

deprecated: Deprecated API {
  style.fill: "#f5f5f5"
  style.stroke: "#9e9e9e"
  style.font-color: "#9e9e9e"
  style.stroke-dash: 5
}

current -> deprecated: migration
```

## Color Reference

### Material Design Colors

```
# Red
#ffcdd2 (light), #f44336 (main), #b71c1c (dark)

# Green
#c8e6c9 (light), #4caf50 (main), #1b5e20 (dark)

# Blue
#e3f2fd (light), #1976d2 (main), #0d47a1 (dark)

# Orange
#fff3e0 (light), #ff9800 (main), #e65100 (dark)

# Yellow
#fff8e1 (light), #ffc107 (main), #ff6f00 (dark)

# Grey
#f5f5f5 (light), #9e9e9e (main), #424242 (dark)
```

## Best Practices

1. **Use color meaningfully** - Reserve red for errors, green for success, etc.

2. **Don't over-style** - Clean diagrams are easier to read

3. **Use classes for consistency** - Define styles once, use everywhere

4. **Consider accessibility** - Don't rely solely on color for meaning

5. **Match your brand** - Use company colors where appropriate
