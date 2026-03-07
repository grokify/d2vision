package plantuml

import (
	"testing"
)

func TestParseSequenceDiagram(t *testing.T) {
	source := `@startuml
participant Alice
participant Bob
Alice -> Bob: Hello
Bob --> Alice: Hi
@enduml`

	doc, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if doc.Type != DiagramSequence {
		t.Errorf("Expected DiagramSequence, got %s", doc.Type)
	}

	if len(doc.Participants) < 2 {
		t.Errorf("Expected at least 2 participants, got %d", len(doc.Participants))
	}

	if len(doc.Messages) < 2 {
		t.Errorf("Expected at least 2 messages, got %d", len(doc.Messages))
	}
}

func TestParseSequenceWithActors(t *testing.T) {
	source := `@startuml
actor User
participant "Web Server" as WS
database DB

User -> WS: Request
WS -> DB: Query
DB --> WS: Result
WS --> User: Response
@enduml`

	doc, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(doc.Participants) < 3 {
		t.Errorf("Expected at least 3 participants, got %d", len(doc.Participants))
	}

	// Check participant types
	typeMap := make(map[ParticipantType]bool)
	for _, p := range doc.Participants {
		typeMap[p.Type] = true
	}

	if !typeMap[ParticipantActor] {
		t.Error("Expected actor participant type")
	}
	if !typeMap[ParticipantDatabase] {
		t.Error("Expected database participant type")
	}
}

func TestParseClassDiagram(t *testing.T) {
	source := `@startuml
class Animal {
  +name: String
  +age: int
  +makeSound()
}

class Dog {
  +breed: String
  +bark()
}

Animal <|-- Dog
@enduml`

	doc, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if doc.Type != DiagramClass {
		t.Errorf("Expected DiagramClass, got %s", doc.Type)
	}

	if len(doc.Classes) < 2 {
		t.Errorf("Expected at least 2 classes, got %d", len(doc.Classes))
	}

	// Check for relations
	if len(doc.Relations) < 1 {
		t.Errorf("Expected at least 1 relation, got %d", len(doc.Relations))
	}
}

func TestParseComponentDiagram(t *testing.T) {
	source := `@startuml
package "Frontend" {
  [Web UI]
  [Mobile App]
}

package "Backend" {
  [API Server]
  [Database]
}

[Web UI] --> [API Server]
[Mobile App] --> [API Server]
[API Server] --> [Database]
@enduml`

	doc, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if doc.Type != DiagramComponent {
		t.Errorf("Expected DiagramComponent, got %s", doc.Type)
	}

	if len(doc.Packages) < 2 {
		t.Errorf("Expected at least 2 packages, got %d", len(doc.Packages))
	}

	if len(doc.Relations) < 3 {
		t.Errorf("Expected at least 3 relations, got %d", len(doc.Relations))
	}
}

func TestConvertSequenceToD2(t *testing.T) {
	source := `@startuml
participant Alice
participant Bob
Alice -> Bob: Hello
Bob --> Alice: Hi
@enduml`

	converter := NewConverter()
	result, err := converter.Convert(source)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Spec == nil {
		t.Fatal("Expected non-nil spec")
	}

	// Should have sequence
	if len(result.Spec.Sequences) == 0 {
		t.Error("Expected sequence in spec")
	}

	seq := result.Spec.Sequences[0]

	// Should have actors
	if len(seq.Actors) < 2 {
		t.Errorf("Expected at least 2 actors, got %d", len(seq.Actors))
	}

	// Should have steps
	if len(seq.Steps) < 2 {
		t.Errorf("Expected at least 2 steps, got %d", len(seq.Steps))
	}
}

func TestConvertComponentToD2(t *testing.T) {
	source := `@startuml
package "System" {
  [Component A]
  [Component B]
}
[Component A] --> [Component B]
@enduml`

	converter := NewConverter()
	result, err := converter.Convert(source)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Spec == nil {
		t.Fatal("Expected non-nil spec")
	}

	// Should have container for package
	if len(result.Spec.Containers) == 0 {
		t.Error("Expected containers in spec")
	}

	// Should have edges
	if len(result.Spec.Edges) == 0 {
		t.Error("Expected edges in spec")
	}
}

func TestLintSequenceDiagram(t *testing.T) {
	source := `@startuml
participant Alice
participant Bob
Alice -> Bob: Hello
note left: This is a note
activate Bob
Bob --> Alice: Hi
deactivate Bob
@enduml`

	converter := NewConverter()
	lintResult, err := converter.Lint(source)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if !lintResult.Convertible {
		t.Error("Expected diagram to be convertible")
	}

	// Should have unsupported features for note and activate
	features := make(map[string]bool)
	for _, u := range lintResult.Unsupported {
		features[u.Feature] = true
	}

	if !features["note"] {
		t.Error("Expected unsupported feature for note")
	}
	if !features["activate/deactivate"] {
		t.Error("Expected unsupported feature for activate/deactivate")
	}
}

func TestParticipantTypeToShape(t *testing.T) {
	tests := []struct {
		pType     ParticipantType
		wantShape string
	}{
		{ParticipantActor, "person"},
		{ParticipantDatabase, "cylinder"},
		{ParticipantDefault, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.pType), func(t *testing.T) {
			got := tt.pType.ToD2Shape()
			if got != tt.wantShape {
				t.Errorf("Expected shape %s, got %s", tt.wantShape, got)
			}
		})
	}
}

func TestRelationTypeToArrows(t *testing.T) {
	tests := []struct {
		relType      RelationType
		wantSource   string
		wantTarget   string
	}{
		{RelationInheritance, "none", "triangle"},
		{RelationAggregation, "diamond", "none"},
		{RelationComposition, "diamond", "triangle"},
	}

	for _, tt := range tests {
		t.Run(string(tt.relType), func(t *testing.T) {
			source, target := tt.relType.ToD2Arrows()
			if source != tt.wantSource {
				t.Errorf("Expected source arrow %s, got %s", tt.wantSource, source)
			}
			if target != tt.wantTarget {
				t.Errorf("Expected target arrow %s, got %s", tt.wantTarget, target)
			}
		})
	}
}

func TestParseSequenceGroups(t *testing.T) {
	source := `@startuml
participant Alice
participant Bob

alt success
    Alice -> Bob: Request
    Bob --> Alice: Response
else failure
    Alice -> Bob: Request
    Bob --> Alice: Error
end
@enduml`

	doc, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(doc.Groups) == 0 {
		t.Error("Expected at least one group")
	}

	group := doc.Groups[0]
	if group.Type != GroupAlt {
		t.Errorf("Expected GroupAlt, got %s", group.Type)
	}

	if len(group.Messages) == 0 {
		t.Error("Expected messages in group")
	}
}

func TestMessageStyles(t *testing.T) {
	tests := []struct {
		arrow     string
		wantStyle MessageStyle
	}{
		{"->", MessageSolid},
		{"-->", MessageDashed},
		{"->>", MessageAsync},
	}

	for _, tt := range tests {
		t.Run(tt.arrow, func(t *testing.T) {
			got := parseArrowStyle(tt.arrow)
			if got != tt.wantStyle {
				t.Errorf("Expected style %s, got %s", tt.wantStyle, got)
			}
		})
	}
}

func TestConvertClassDiagramToD2(t *testing.T) {
	source := `@startuml
class Animal {
  +name: String
  +makeSound()
}

class Dog
Animal <|-- Dog
@enduml`

	converter := NewConverter()
	result, err := converter.Convert(source)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Spec == nil {
		t.Fatal("Expected non-nil spec")
	}

	// Should have containers for classes
	if len(result.Spec.Containers) == 0 {
		t.Error("Expected containers in spec")
	}

	// Should have edges for relationships
	if len(result.Spec.Edges) == 0 {
		t.Error("Expected edges in spec")
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"with space", "with_space"},
		{"with-dash", "with_dash"},
		{"123start", "_123start"},
		{"MixedCase", "MixedCase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeID(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
