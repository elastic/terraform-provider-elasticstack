package kbapi

import (
	"encoding/json"
	"fmt"
)

// PostAlertingRuleIdJSONBody_Params is generated as a wrapper type to support
// oneOf/AdditionalProperties. The generated struct does not include custom JSON
// marshalling, which would otherwise cause params to always serialize as `{}`.
//
// We intentionally treat this as a free-form object: if AdditionalProperties is
// set, marshal that object. If union is set (e.g. via UnmarshalJSON), prefer it.
func (p PostAlertingRuleIdJSONBody_Params) MarshalJSON() ([]byte, error) {
	if len(p.union) > 0 {
		return p.union, nil
	}
	if p.AdditionalProperties == nil {
		return []byte(`{}`), nil
	}
	return json.Marshal(p.AdditionalProperties)
}

func (p *PostAlertingRuleIdJSONBody_Params) UnmarshalJSON(b []byte) error {
	if p == nil {
		return fmt.Errorf("PostAlertingRuleIdJSONBody_Params: UnmarshalJSON on nil receiver")
	}

	// Store raw bytes for round-tripping.
	p.union = append(p.union[:0], b...)

	// Params are expected to be a JSON object.
	var values map[string]interface{}
	if err := json.Unmarshal(b, &values); err != nil {
		return err
	}
	p.AdditionalProperties = values
	return nil
}
