package d2vision

// ShapeType represents the visual shape of a node.
type ShapeType string

// Shape type constants.
const (
	ShapeRectangle     ShapeType = "rectangle"
	ShapeSquare        ShapeType = "square"
	ShapeCircle        ShapeType = "circle"
	ShapeOval          ShapeType = "oval"
	ShapeCylinder      ShapeType = "cylinder"
	ShapeDiamond       ShapeType = "diamond"
	ShapeHexagon       ShapeType = "hexagon"
	ShapeParallelogram ShapeType = "parallelogram"
	ShapeCloud         ShapeType = "cloud"
	ShapeDocument      ShapeType = "document"
	ShapeQueue         ShapeType = "queue"
	ShapePackage       ShapeType = "package"
	ShapePerson        ShapeType = "person"
	ShapeClass         ShapeType = "class"
	ShapeCode          ShapeType = "code"
	ShapeImage         ShapeType = "image"
	ShapeText          ShapeType = "text"
	ShapeUnknown       ShapeType = "unknown"
)

// String returns the string representation of the shape type.
func (s ShapeType) String() string {
	return string(s)
}

// NaturalName returns a human-readable name for the shape.
func (s ShapeType) NaturalName() string {
	switch s {
	case ShapeRectangle:
		return "rectangle"
	case ShapeSquare:
		return "square"
	case ShapeCircle:
		return "circle"
	case ShapeOval:
		return "oval"
	case ShapeCylinder:
		return "cylinder"
	case ShapeDiamond:
		return "diamond"
	case ShapeHexagon:
		return "hexagon"
	case ShapeParallelogram:
		return "parallelogram"
	case ShapeCloud:
		return "cloud"
	case ShapeDocument:
		return "document"
	case ShapeQueue:
		return "queue"
	case ShapePackage:
		return "package"
	case ShapePerson:
		return "person"
	case ShapeClass:
		return "class diagram"
	case ShapeCode:
		return "code block"
	case ShapeImage:
		return "image"
	case ShapeText:
		return "text"
	default:
		return "shape"
	}
}

// ArrowType represents the type of arrow on an edge endpoint.
type ArrowType string

// Arrow type constants.
const (
	ArrowNone      ArrowType = "none"
	ArrowTriangle  ArrowType = "triangle"
	ArrowDiamond   ArrowType = "diamond"
	ArrowCircle    ArrowType = "circle"
	ArrowCrowfoot  ArrowType = "crowfoot"
	ArrowCFOne     ArrowType = "cf-one"
	ArrowCFMany    ArrowType = "cf-many"
	ArrowCFOneReq  ArrowType = "cf-one-required"
	ArrowCFManyReq ArrowType = "cf-many-required"
)

// String returns the string representation of the arrow type.
func (a ArrowType) String() string {
	return string(a)
}
