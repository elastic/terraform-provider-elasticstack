package rolemapping

import (
	"encoding/json"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

func TestRoleTemplatesToJSON(t *testing.T) {
	// Simulate parsing config JSON into typed struct
	configJSON := `[{"format":"json","template":"{\"source\":\"{{#tojson}}groups{{/tojson}}\"}"}]`

	var roleTemplates []types.RoleTemplate
	if err := json.Unmarshal([]byte(configJSON), &roleTemplates); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Simulate ES round-trip (marshal then unmarshal)
	sent, _ := json.Marshal(roleTemplates)

	var fromES []types.RoleTemplate
	if err := json.Unmarshal(sent, &fromES); err != nil {
		t.Fatalf("es parse error: %v", err)
	}

	result, err := roleTemplatesToJSON(fromES)
	if err != nil {
		t.Fatalf("roleTemplatesToJSON error: %v", err)
	}

	t.Logf("config:   %s", configJSON)
	t.Logf("sent:     %s", string(sent))
	t.Logf("fromES.Source: %q", *fromES[0].Template.Source)
	t.Logf("result:   %s", result)

	if result != configJSON {
		t.Errorf("roleTemplatesToJSON mismatch\nexpected: %s\ngot:      %s", configJSON, result)
	}
}
