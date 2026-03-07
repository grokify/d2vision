package plantuml

import (
	"github.com/grokify/d2vision/generate"
)

// SequenceConverter converts PlantUML sequence diagrams to DiagramSpec.
type SequenceConverter struct{}

// Convert transforms a parsed PlantUML sequence document into a DiagramSpec.
func (c *SequenceConverter) Convert(doc *Document) *generate.DiagramSpec {
	spec := &generate.DiagramSpec{}

	seq := generate.SequenceSpec{
		ID:    "sequence",
		Label: doc.Title,
	}

	if seq.Label == "" {
		seq.Label = "Sequence Diagram"
	}

	// Convert participants
	for _, part := range doc.Participants {
		actorSpec := generate.ActorSpec{
			ID:    part.ID,
			Label: part.Label,
			Shape: part.Type.ToD2Shape(),
		}
		seq.Actors = append(seq.Actors, actorSpec)
	}

	// If no explicit participants, derive from messages
	if len(seq.Actors) == 0 {
		seenActors := make(map[string]bool)
		for _, msg := range doc.Messages {
			if !seenActors[msg.From] {
				seq.Actors = append(seq.Actors, generate.ActorSpec{
					ID:    msg.From,
					Label: msg.From,
				})
				seenActors[msg.From] = true
			}
			if !seenActors[msg.To] {
				seq.Actors = append(seq.Actors, generate.ActorSpec{
					ID:    msg.To,
					Label: msg.To,
				})
				seenActors[msg.To] = true
			}
		}
	}

	// Convert messages
	for _, msg := range doc.Messages {
		seq.Steps = append(seq.Steps, generate.MessageSpec{
			From:  msg.From,
			To:    msg.To,
			Label: msg.Label,
		})
	}

	// Convert groups
	for _, group := range doc.Groups {
		groupSpec := generate.GroupSpec{
			ID:    string(group.Type),
			Label: group.Label,
		}

		for _, msg := range group.Messages {
			groupSpec.Messages = append(groupSpec.Messages, generate.MessageSpec{
				From:  msg.From,
				To:    msg.To,
				Label: msg.Label,
			})
		}

		// Handle else branches
		for _, elseBranch := range group.Else {
			// Add else messages to the group
			// Note: D2's sequence diagram support may not have full else support
			for _, msg := range elseBranch.Messages {
				groupSpec.Messages = append(groupSpec.Messages, generate.MessageSpec{
					From:  msg.From,
					To:    msg.To,
					Label: msg.Label,
				})
			}
		}

		seq.Groups = append(seq.Groups, groupSpec)
	}

	spec.Sequences = append(spec.Sequences, seq)
	return spec
}
