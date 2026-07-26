package document

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func upsertReferences(parent *yaml.Node, references []MetadataReference) {
	seq := findOrCreateSequence(parent, "references")
	for _, reference := range references {
		existing := findReferenceNode(seq, reference)
		if existing == nil {
			existing = &yaml.Node{Kind: yaml.MappingNode}
			seq.Content = append(seq.Content, existing)
		}
		setOptionalNodeString(existing, "id", reference.ID)
		setNodeString(existing, "name", reference.Name)
		setNodeString(existing, "type", reference.Type)
		setNodeString(existing, "target", reference.Target)
		setOptionalNodeString(existing, "selector_type", reference.SelectorType)
		setOptionalNodeString(existing, "selector", reference.Selector)
		setNodeString(existing, "relation", reference.Relation)
		setNodeString(existing, "read_policy", reference.ReadPolicy)
		setNodeString(existing, "used_for", reference.UsedFor)
		setNodeString(existing, "status", reference.Status)
	}
}

func findReferenceNode(seq *yaml.Node, reference MetadataReference) *yaml.Node {
	for _, item := range seq.Content {
		if item.Kind != yaml.MappingNode {
			continue
		}
		if strings.TrimSpace(reference.ID) != "" &&
			strings.EqualFold(getNodeString(item, "id"), reference.ID) {
			return item
		}
		if strings.EqualFold(getNodeString(item, "name"), reference.Name) &&
			strings.EqualFold(getNodeString(item, "type"), reference.Type) &&
			getNodeString(item, "target") == reference.Target &&
			getNodeString(item, "selector_type") == reference.SelectorType &&
			getNodeString(item, "selector") == reference.Selector {
			return item
		}
	}
	return nil
}

func upsertSkills(parent *yaml.Node, skills []MetadataSkill) {
	seq := findOrCreateSequence(parent, "skills")
	for _, skill := range skills {
		existing := findSkillNode(seq, skill)
		if existing == nil {
			existing = &yaml.Node{Kind: yaml.MappingNode}
			seq.Content = append(seq.Content, existing)
		}
		setNodeString(existing, "name", skill.Name)
		setNodeString(existing, "source", skill.Source)
		setNodeString(existing, "path", skill.Path)
		setNodeString(existing, "trigger", skill.Trigger)
		setNodeBool(existing, "required", skill.Required)
	}
}

func findSkillNode(seq *yaml.Node, skill MetadataSkill) *yaml.Node {
	for _, item := range seq.Content {
		if item.Kind != yaml.MappingNode {
			continue
		}
		if strings.EqualFold(getNodeString(item, "name"), skill.Name) && getNodeString(item, "path") == skill.Path {
			return item
		}
	}
	return nil
}

func findOrCreateMapping(parent *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value != key {
			continue
		}
		if parent.Content[i+1].Kind != yaml.MappingNode {
			parent.Content[i+1] = &yaml.Node{Kind: yaml.MappingNode}
		}
		return parent.Content[i+1]
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	valueNode := &yaml.Node{Kind: yaml.MappingNode}
	parent.Content = append(parent.Content, keyNode, valueNode)
	return valueNode
}

func findOrCreateSequence(parent *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value != key {
			continue
		}
		if parent.Content[i+1].Kind != yaml.SequenceNode {
			parent.Content[i+1] = &yaml.Node{Kind: yaml.SequenceNode}
		}
		return parent.Content[i+1]
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	valueNode := &yaml.Node{Kind: yaml.SequenceNode}
	parent.Content = append(parent.Content, keyNode, valueNode)
	return valueNode
}

func setNodeString(parent *yaml.Node, key, value string) {
	setNode(parent, key, &yaml.Node{Kind: yaml.ScalarNode, Value: value})
}

func setOptionalNodeString(parent *yaml.Node, key, value string) {
	if strings.TrimSpace(value) == "" {
		removeNode(parent, key)
		return
	}
	setNodeString(parent, key, value)
}

func setNodeInt(parent *yaml.Node, key string, value int) {
	setNode(parent, key, &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", value)})
}

func setNodeBool(parent *yaml.Node, key string, value bool) {
	scalar := &yaml.Node{Kind: yaml.ScalarNode, Value: "false"}
	if value {
		scalar.Value = "true"
	}
	setNode(parent, key, scalar)
}

func setNode(parent *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			parent.Content[i+1] = value
			return
		}
	}
	parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, value)
}

func removeNode(parent *yaml.Node, key string) {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value != key {
			continue
		}
		parent.Content = append(parent.Content[:i], parent.Content[i+2:]...)
		return
	}
}

func getNodeString(parent *yaml.Node, key string) string {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			return parent.Content[i+1].Value
		}
	}
	return ""
}
