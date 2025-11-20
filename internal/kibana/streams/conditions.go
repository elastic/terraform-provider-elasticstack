package streams

import "encoding/json"

// Condition represents a logical condition tree that can be serialized
// into a JSON-friendly DSL structure.
//
// This is intentionally generic and internal to the provider so that we
// can evolve the exact JSON shape to match the Streams condition DSL
// without leaking details into the rest of the codebase.
type Condition interface {
	// ToDSL returns the condition encoded as a nested map/list structure
	// suitable for JSON marshaling.
	ToDSL() map[string]any
}

// FieldComparison is a simple leaf condition like:
//
//	field <op> value
//
// Example JSON (one possible shape):
//
//	{"field": "host.name", "op": "eq", "value": "web-01"}
type FieldComparison struct {
	Field string
	Op    string
	Value any
}

// ToDSL implements Condition.
func (c FieldComparison) ToDSL() map[string]any {
	return map[string]any{
		"field": c.Field,
		"op":    c.Op,
		"value": c.Value,
	}
}

// And represents a logical AND over its children.
//
// Example JSON:
//
//	{"and": [ <child1>, <child2>, ... ]}
type And struct {
	Children []Condition
}

// ToDSL implements Condition.
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

// Or represents a logical OR over its children.
//
// Example JSON:
//
//	{"or": [ <child1>, <child2>, ... ]}
type Or struct {
	Children []Condition
}

// ToDSL implements Condition.
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

// MarshalCondition encodes a Condition into JSON. This is a helper used
// by unit tests (and potentially by Streams helpers in the future).
func MarshalCondition(c Condition) ([]byte, error) {
	if c == nil {
		// Encode as JSON null for now; callers can decide how to handle this.
		return []byte("null"), nil
	}
	return json.Marshal(c.ToDSL())
}
