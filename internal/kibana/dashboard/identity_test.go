// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
