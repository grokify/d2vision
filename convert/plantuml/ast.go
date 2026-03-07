// Package plantuml provides parsing and conversion of PlantUML diagrams to D2.
package plantuml

// DiagramType represents the type of PlantUML diagram.
type DiagramType string

const (
	DiagramSequence  DiagramType = "sequence"
	DiagramClass     DiagramType = "class"
	DiagramComponent DiagramType = "component"
	DiagramActivity  DiagramType = "activity"
	DiagramUseCase   DiagramType = "usecase"
	DiagramState     DiagramType = "state"
	DiagramObject    DiagramType = "object"
	DiagramUnknown   DiagramType = "unknown"
)

// Document represents a parsed PlantUML document.
type Document struct {
	Type       DiagramType
	Title      string
	Participants []*Participant
	Messages     []*Message
	Groups       []*Group
	Classes      []*Class
	Packages     []*Package
	Components   []*Component
	Relations    []*Relation
	Notes        []*Note

	// Raw lines for diagnostics
	Lines []string
}

// Participant represents a participant/actor in a sequence diagram.
type Participant struct {
	ID       string
	Label    string
	Type     ParticipantType
	Stereot  string // <<stereotype>>
	Line     int
}

// ParticipantType represents the type of participant.
type ParticipantType string

const (
	ParticipantDefault   ParticipantType = "participant"
	ParticipantActor     ParticipantType = "actor"
	ParticipantBoundary  ParticipantType = "boundary"
	ParticipantControl   ParticipantType = "control"
	ParticipantEntity    ParticipantType = "entity"
	ParticipantDatabase  ParticipantType = "database"
	ParticipantCollections ParticipantType = "collections"
	ParticipantQueue     ParticipantType = "queue"
)

// ToD2Shape converts a participant type to a D2 shape.
func (t ParticipantType) ToD2Shape() string {
	switch t {
	case ParticipantActor:
		return "person"
	case ParticipantDatabase:
		return "cylinder"
	case ParticipantBoundary:
		return "rectangle"
	case ParticipantControl:
		return "circle"
	case ParticipantEntity:
		return "rectangle"
	case ParticipantQueue:
		return "queue"
	default:
		return ""
	}
}

// Message represents a message in a sequence diagram.
type Message struct {
	From      string
	To        string
	Label     string
	Style     MessageStyle
	Return    bool // Is this a return message?
	Line      int
}

// MessageStyle represents the visual style of a message.
type MessageStyle string

const (
	MessageSolid  MessageStyle = "solid"
	MessageDashed MessageStyle = "dashed"
	MessageAsync  MessageStyle = "async"
)

// Group represents a group of messages (alt, opt, loop, etc.).
type Group struct {
	Type     GroupType
	Label    string
	Messages []*Message
	Else     []*GroupElse
	Line     int
}

// GroupElse represents an else branch in a group.
type GroupElse struct {
	Label    string
	Messages []*Message
}

// GroupType represents the type of group.
type GroupType string

const (
	GroupAlt      GroupType = "alt"
	GroupOpt      GroupType = "opt"
	GroupLoop     GroupType = "loop"
	GroupPar      GroupType = "par"
	GroupBreak    GroupType = "break"
	GroupCritical GroupType = "critical"
	GroupGroup    GroupType = "group"
	GroupRef      GroupType = "ref"
)

// Note represents a note in a diagram.
type Note struct {
	Position string // left, right, over
	Target   string // participant ID or empty
	Text     string
	Line     int
}

// Class represents a class in a class diagram.
type Class struct {
	ID         string
	Label      string
	Stereotype string
	Abstract   bool
	Interface  bool
	Attributes []string
	Methods    []string
	Line       int
}

// Package represents a package/namespace container.
type Package struct {
	ID         string
	Label      string
	Type       PackageType
	Classes    []*Class
	Packages   []*Package // Nested packages
	Components []*Component
	Line       int
}

// PackageType represents the visual type of a package.
type PackageType string

const (
	PackageDefault   PackageType = "package"
	PackageNode      PackageType = "node"
	PackageFolder    PackageType = "folder"
	PackageFrame     PackageType = "frame"
	PackageCloud     PackageType = "cloud"
	PackageDatabase  PackageType = "database"
	PackageRectangle PackageType = "rectangle"
)

// Component represents a component in a component diagram.
type Component struct {
	ID         string
	Label      string
	Stereotype string
	Line       int
}

// Relation represents a relationship between elements.
type Relation struct {
	From       string
	To         string
	Label      string
	Type       RelationType
	FromCard   string // Cardinality
	ToCard     string
	Line       int
}

// RelationType represents the type of relationship.
type RelationType string

const (
	RelationAssociation  RelationType = "association"
	RelationDependency   RelationType = "dependency"
	RelationAggregation  RelationType = "aggregation"
	RelationComposition  RelationType = "composition"
	RelationInheritance  RelationType = "inheritance"
	RelationRealization  RelationType = "realization"
	RelationLink         RelationType = "link"
)

// ToD2Arrows returns source and target arrow styles for D2.
func (r RelationType) ToD2Arrows() (source, target string) {
	switch r {
	case RelationInheritance:
		return "none", "triangle"
	case RelationRealization:
		return "none", "triangle" // Use dashed style
	case RelationAggregation:
		return "diamond", "none"
	case RelationComposition:
		return "diamond", "triangle"
	case RelationDependency:
		return "none", "triangle" // Use dashed style
	case RelationAssociation:
		return "none", "triangle"
	default:
		return "none", "triangle"
	}
}
