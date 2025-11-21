package streams

import "encoding/json"

// Condition represents a logical condition tree that can be serialized
// into a JSON-friendly DSL structure.

type Condition interface {
	ToDSL() map[string]any
}

type FieldComparison struct {
	Field string
	Op    string
	Value any
}

func (c FieldComparison) ToDSL() map[string]any {
	return map[string]any{
		"field": c.Field,
		"op":    c.Op,
		"value": c.Value,
	}
}

type And struct {
	Children []Condition
}

func (a And) ToDSL() map[string]any {
	children := make([]any, 0, len(a.Children))
	for _, child := range a.Children {
		if child == nil {
			continue
		}
		children = append(children, child.ToDSL())
	}
	return map[string]any{
		"and": children,
	}
}

type Or struct {
	Children []Condition
}

func (o Or) ToDSL() map[string]any {
	children := make([]any, 0, len(o.Children))
	for _, child := range o.Children {
		if child == nil {
			continue
		}
		children = append(children, child.ToDSL())
	}
	return map[string]any{
		"or": children,
	}
}

func MarshalCondition(c Condition) ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}
	return json.Marshal(c.ToDSL())
}
