package dashboard

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
)

func TestDashboardIdentitySchema(t *testing.T) {
	t.Parallel()

	r := &Resource{}
	var resp resource.IdentitySchemaResponse
	r.IdentitySchema(context.Background(), resource.IdentitySchemaRequest{}, &resp)

	attr, ok := resp.IdentitySchema.Attributes["id"]
	if !ok {
		t.Fatalf("expected identity schema to include attribute %q", "id")
	}

	stringAttr, ok := attr.(identityschema.StringAttribute)
	if !ok {
		t.Fatalf("expected identity attribute %q to be identityschema.StringAttribute, got %T", "id", attr)
	}

	if !stringAttr.RequiredForImport {
		t.Fatalf("expected identity attribute %q to be RequiredForImport", "id")
	}
}
