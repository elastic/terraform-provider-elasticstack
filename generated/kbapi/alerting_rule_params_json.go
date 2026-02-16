package kbapi

import (
	"encoding/json"
	"fmt"
)

// NOTE: This file is hand-maintained.
//
// The Kibana API types in this package are generated via `generated/kbapi/Makefile`
// using `oapi-codegen`. Some generated wrapper types (notably oneOf +
// AdditionalProperties unions) require custom JSON marshalling to preserve
// free-form object payloads. Keeping this file in the `kbapi` package ensures
// the methods attach to the generated types without modifying generated output.
//
// The current codegen workflow does not wipe this directory; it generates into
// a fixed output file configured by `oapi-config.yaml`.
//
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
