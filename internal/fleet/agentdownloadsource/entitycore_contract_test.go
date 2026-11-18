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

package agentdownloadsource

import (
	"context"
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/providerfwtest"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestResource_schemaAuthExposesCredentialAndHeaderAttributes(t *testing.T) {
	t.Parallel()

	authAttr, ok := getSchema(context.Background()).Attributes["auth"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	require.True(t, authAttr.Optional)

	username, ok := authAttr.Attributes["username"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, username.Optional)
	require.False(t, username.Sensitive)

	password, ok := authAttr.Attributes["password"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, password.Optional)
	require.True(t, password.Sensitive)

	apiKey, ok := authAttr.Attributes["api_key"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, apiKey.Optional)
	require.True(t, apiKey.Sensitive)

	headers, ok := authAttr.Attributes["headers"].(schema.ListNestedAttribute)
	require.True(t, ok)
	require.True(t, headers.Optional)

	key, ok := headers.NestedObject.Attributes["key"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, key.Required)

	value, ok := headers.NestedObject.Attributes["value"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, value.Required)
}

func TestResource_schemaOmitsSecretsReferences(t *testing.T) {
	t.Parallel()

	attrs := getSchema(context.Background()).Attributes
	_, hasSecrets := attrs["secrets"]
	require.False(t, hasSecrets)

	authAttr, ok := attrs["auth"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	_, hasNestedSecrets := authAttr.Attributes["secrets"]
	require.False(t, hasNestedSecrets)

	slAttr, ok := attrs["ssl"].(schema.SingleNestedAttribute)
	require.True(t, ok, "ssl attribute must exist as a single nested attribute")
	_, hasSslNestedSecrets := slAttr.Attributes["secrets"]
	require.False(t, hasSslNestedSecrets)
}

func TestResource_schemaWriteOnlyAttributesAreWriteOnlyAndSensitive(t *testing.T) {
	t.Parallel()

	attrs := getSchema(context.Background()).Attributes
	sslAttr, ok := attrs["ssl"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	authAttr, ok := attrs["auth"].(schema.SingleNestedAttribute)
	require.True(t, ok)

	for name, nested := range map[string]map[string]schema.Attribute{
		"ssl.key_wo":       {"key_wo": sslAttr.Attributes["key_wo"]},
		"auth.password_wo": {"password_wo": authAttr.Attributes["password_wo"]},
		"auth.api_key_wo":  {"api_key_wo": authAttr.Attributes["api_key_wo"]},
	} {
		for _, a := range nested {
			wo, ok := a.(schema.StringAttribute)
			require.True(t, ok, name)
			require.True(t, wo.Optional, name)
			require.True(t, wo.WriteOnly, name)
			require.True(t, wo.Sensitive, name)
			require.False(t, wo.Computed, name)
		}
	}

	for _, legacy := range []string{"key_version", "key_wo_version"} {
		_, has := sslAttr.Attributes[legacy]
		require.False(t, has, legacy)
	}
	for _, legacy := range []string{"password_wo_version", "api_key_wo_version"} {
		_, has := authAttr.Attributes[legacy]
		require.False(t, has, legacy)
	}
}

func TestResource_embedsKibanaResource(t *testing.T) {
	t.Parallel()
	rt := reflect.TypeFor[Resource]()
	field, ok := rt.FieldByName("KibanaResource")
	require.True(t, ok)
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.KibanaResource[model]](), field.Type)
}

// Import splits "<space>/<source_id>" into structured attributes instead of
// preserving the raw import string on id (passthrough shape).
func TestResource_importState_customCompositeID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "myspace/the-src-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var sourceID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("source_id"), &sourceID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "the-src-id", sourceID.ValueString())

	var id types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "the-src-id", id.ValueString())

	var spaceIDs types.Set
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_ids"), &spaceIDs)...)
	require.False(t, resp.Diagnostics.HasError())

	var elems []types.String
	resp.Diagnostics.Append(spaceIDs.ElementsAs(ctx, &elems, false)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Len(t, elems, 1)
	require.Equal(t, "myspace", elems[0].ValueString())
}
