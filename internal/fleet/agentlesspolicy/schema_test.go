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

package agentlesspolicy

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetSchema_noKibanaConnectionBlock asserts this schema factory does not
// define `kibana_connection` (or any Blocks) itself: the resource is built
// on the entitycore.KibanaResource[T] envelope, which injects that block
// (see internal/entitycore/base_envelope.go). Defining it here too would be
// redundant (silently overwritten) and would diverge from the pattern used
// by other envelope-based Fleet resources (internal/fleet/proxy,
// internal/fleet/serverhost, internal/fleet/output).
func TestGetSchema_noKibanaConnectionBlock(t *testing.T) {
	t.Parallel()
	s := getSchema(context.Background())
	assert.Empty(t, s.Blocks, "expected the schema factory to leave block injection to the entitycore envelope")
}

// TestGetSchema_identityAttributes checks Optional/Computed/Required and the
// force-replacement plan modifiers for the identity attributes against
// specs/fleet-agentless-policy/spec.md's "Schema attributes" requirement.
func TestGetSchema_identityAttributes(t *testing.T) {
	t.Parallel()
	s := getSchema(context.Background())

	idAttr, ok := s.Attributes["id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, idAttr.Computed)
	assert.False(t, idAttr.Optional)
	assert.False(t, idAttr.Required)

	policyID, ok := s.Attributes["policy_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, policyID.Optional)
	assert.True(t, policyID.Computed)
	assert.True(t, hasStringRequiresReplace(policyID.PlanModifiers), "policy_id should force replacement on change")

	name, ok := s.Attributes["name"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, name.Required)
	assert.False(t, name.Optional)
	assert.True(t, hasStringRequiresReplace(name.PlanModifiers), "name should force replacement on change")

	description, ok := s.Attributes["description"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, description.Optional)
	assert.False(t, description.Computed)
	assert.False(t, hasStringRequiresReplace(description.PlanModifiers), "description is updatable in-place, not RequiresReplace")

	namespace, ok := s.Attributes["namespace"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, namespace.Optional)
	assert.True(t, namespace.Computed)
	assert.True(t, hasStringRequiresReplace(namespace.PlanModifiers), "namespace should force replacement on change")

	spaceIDs, ok := s.Attributes["space_ids"].(schema.SetAttribute)
	require.True(t, ok)
	assert.True(t, spaceIDs.Optional)
	assert.True(t, spaceIDs.Computed)
	assert.Equal(t, types.StringType, spaceIDs.ElementType)
	require.NotEmpty(t, spaceIDs.PlanModifiers)
}

// TestGetSchema_package checks that package.name/version force replacement
// while package.title is updatable in-place (the spike-informed exception
// documented in update.go and design.md Decision 3).
func TestGetSchema_package(t *testing.T) {
	t.Parallel()
	s := getSchema(context.Background())

	pkg, ok := s.Attributes["package"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	assert.True(t, pkg.Required)

	name, ok := pkg.Attributes["name"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, name.Required)
	assert.True(t, hasStringRequiresReplace(name.PlanModifiers), "package.name should force replacement on change")

	version, ok := pkg.Attributes["version"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, version.Required)
	assert.True(t, hasStringRequiresReplace(version.PlanModifiers), "package.version should force replacement on change")

	title, ok := pkg.Attributes["title"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, title.Optional)
	assert.True(t, title.Computed)
	assert.False(t, hasStringRequiresReplace(title.PlanModifiers), "package.title is updatable in-place per the Task 3 spike findings")
}

// TestGetSchema_policyTemplate checks policy_template forces replacement.
func TestGetSchema_policyTemplate(t *testing.T) {
	t.Parallel()
	s := getSchema(context.Background())

	policyTemplate, ok := s.Attributes["policy_template"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, policyTemplate.Optional)
	assert.True(t, hasStringRequiresReplace(policyTemplate.PlanModifiers))
}

// TestGetSchema_varsJSONAndInputs checks that vars_json and inputs reuse the
// shared policyshape custom types (no local reimplementation), and that
// inputs is Optional+Computed with UseStateForUnknown.
func TestGetSchema_varsJSONAndInputs(t *testing.T) {
	t.Parallel()
	s := getSchema(context.Background())

	varsJSON, ok := s.Attributes["vars_json"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, varsJSON.Optional)
	assert.True(t, varsJSON.Computed)
	assert.Equal(t, "policyshape.VarsJSONType", varsJSON.CustomType.String())

	inputs, ok := s.Attributes["inputs"].(schema.MapNestedAttribute)
	require.True(t, ok)
	assert.True(t, inputs.Optional)
	assert.True(t, inputs.Computed)
	assert.Equal(t, "policyshape.InputsType", inputs.CustomType.String())
	require.NotEmpty(t, inputs.PlanModifiers)

	// Element-level shape: enabled/condition/vars/streams, matching the
	// integration_policy convention (vars keyed as "vars", not "vars_json").
	_, hasEnabled := inputs.NestedObject.Attributes["enabled"]
	_, hasCondition := inputs.NestedObject.Attributes["condition"]
	_, hasVars := inputs.NestedObject.Attributes["vars"]
	_, hasStreams := inputs.NestedObject.Attributes["streams"]
	assert.True(t, hasEnabled)
	assert.True(t, hasCondition)
	assert.True(t, hasVars)
	assert.True(t, hasStreams)

	// Input-level `vars` must be Computed with UseStateForUnknown (not purely
	// Optional, unlike integration_policy's equivalent attribute): some
	// packages (e.g. cloud_security_posture/CSPM) populate informational
	// input-level vars server-side regardless of config, which trips
	// "Provider produced inconsistent result after apply" without Computed.
	// See schema.go's getInputsNestedObject doc comment and this change's
	// review-fix note in spec.md's "Variables and inputs" requirement.
	inputVars, ok := inputs.NestedObject.Attributes["vars"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, inputVars.Computed, "input-level vars should be Computed to tolerate server-populated informational vars")
	assert.True(t, hasStringUseStateForUnknown(inputVars.PlanModifiers), "input-level vars should have UseStateForUnknown")
	_, hasDefaults := inputs.NestedObject.Attributes["defaults"]
	assert.False(t, hasDefaults, "agentless inputs intentionally omit the package-defaults object modeled by integration_policy")

	varGroupSelections, ok := s.Attributes["var_group_selections"].(schema.MapAttribute)
	require.True(t, ok)
	assert.True(t, varGroupSelections.Optional)
	assert.Equal(t, types.StringType, varGroupSelections.ElementType)
}

// TestGetSchema_cloudConnector checks the whole cloud_connector object forces
// replacement on change, and that target_csp is validated against the three
// supported cloud service providers.
func TestGetSchema_cloudConnector(t *testing.T) {
	t.Parallel()
	s := getSchema(context.Background())

	cc, ok := s.Attributes["cloud_connector"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	assert.True(t, cc.Optional)
	require.NotEmpty(t, cc.PlanModifiers, "cloud_connector should force replacement on change at the object level")

	targetCSP, ok := cc.Attributes["target_csp"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, targetCSP.Optional)
	require.NotEmpty(t, targetCSP.Validators)

	for _, valid := range []string{"aws", "azure", "gcp"} {
		t.Run("valid/"+valid, func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{ConfigValue: types.StringValue(valid)}
			var resp validator.StringResponse
			for _, v := range targetCSP.Validators {
				v.ValidateString(context.Background(), req, &resp)
			}
			assert.False(t, resp.Diagnostics.HasError())
		})
	}

	t.Run("invalid/value", func(t *testing.T) {
		t.Parallel()
		req := validator.StringRequest{ConfigValue: types.StringValue("invalid-csp")}
		var resp validator.StringResponse
		for _, v := range targetCSP.Validators {
			v.ValidateString(context.Background(), req, &resp)
		}
		assert.True(t, resp.Diagnostics.HasError())
	})

	for _, name := range []string{"enabled", "cloud_connector_id", "name"} {
		_, ok := cc.Attributes[name]
		assert.True(t, ok, "expected cloud_connector to define %q", name)
	}
}

// TestGetSchema_extrasAndOperationFlags checks the remaining spec attributes:
// global_data_tags, additional_datastreams_permissions,
// create_dataset_templates, force, force_delete, skip_topology_check,
// created_at, updated_at.
func TestGetSchema_extrasAndOperationFlags(t *testing.T) {
	t.Parallel()
	s := getSchema(context.Background())

	globalDataTags, ok := s.Attributes["global_data_tags"].(schema.ListNestedAttribute)
	require.True(t, ok)
	assert.True(t, globalDataTags.Optional)
	_, hasName := globalDataTags.NestedObject.Attributes["name"]
	_, hasValue := globalDataTags.NestedObject.Attributes["value"]
	assert.True(t, hasName)
	assert.True(t, hasValue)

	additionalPerms, ok := s.Attributes["additional_datastreams_permissions"].(schema.ListAttribute)
	require.True(t, ok)
	assert.True(t, additionalPerms.Optional)
	assert.Equal(t, types.StringType, additionalPerms.ElementType)

	createDatasetTemplates, ok := s.Attributes["create_dataset_templates"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, createDatasetTemplates.Optional)
	assert.False(t, createDatasetTemplates.Computed)
	require.Empty(t, createDatasetTemplates.PlanModifiers, "create_dataset_templates is create-only, not RequiresReplace")

	force, ok := s.Attributes["force"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, force.Optional)
	assert.False(t, force.Computed)

	forceDelete, ok := s.Attributes["force_delete"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, forceDelete.Optional)
	assert.True(t, forceDelete.Computed)
	require.NotNil(t, forceDelete.Default)

	skipTopologyCheck, ok := s.Attributes["skip_topology_check"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, skipTopologyCheck.Optional)
	assert.False(t, skipTopologyCheck.Computed, "skip_topology_check is a client-side preflight toggle, not API-persisted")
	require.Empty(t, skipTopologyCheck.PlanModifiers, "skip_topology_check is not RequiresReplace")

	createdAt, ok := s.Attributes["created_at"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, createdAt.Computed)
	assert.False(t, createdAt.Optional)
	// created_at carries UseStateForUnknown because it never legitimately
	// changes after the resource is created (Kibana never updates it), so
	// there is no risk of the plan promising "unchanged" and apply then
	// producing a different value -- unlike updated_at just below.
	assert.True(t, hasStringUseStateForUnknown(createdAt.PlanModifiers), "created_at should have UseStateForUnknown")

	updatedAt, ok := s.Attributes["updated_at"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, updatedAt.Computed)
	assert.False(t, updatedAt.Optional)
	// updated_at deliberately must NOT carry UseStateForUnknown: it
	// legitimately changes on every real Update (Kibana bumps it), so a plan
	// modifier that pre-commits it to "stays equal to prior state" would be
	// actively wrong and produce a live "Provider produced inconsistent
	// result after apply" error whenever a real content change actually
	// happens -- confirmed empirically via the acceptance test step in
	// acc_test.go. onlyCreateOnlyFlagsChanged (update.go) is instead written
	// to never depend on updated_at's plan value at all -- see its doc
	// comment and TestOnlyCreateOnlyFlagsChanged.
	assert.False(t, hasStringUseStateForUnknown(updatedAt.PlanModifiers), "updated_at must NOT have UseStateForUnknown")
}

// TestAgentlessPolicyModel_coversAllSchemaAttributes cross-checks that every
// top-level schema attribute has a corresponding tfsdk-tagged field on
// agentlessPolicyModel, catching drift between schema.go and models.go (the
// two files Task 4 edits together).
func TestAgentlessPolicyModel_coversAllSchemaAttributes(t *testing.T) {
	t.Parallel()
	s := getSchema(context.Background())

	tfsdkTags := map[string]bool{}
	rt := reflect.TypeFor[agentlessPolicyModel]()
	collectTFSDKTags(rt, tfsdkTags)

	for attrName := range s.Attributes {
		assert.True(t, tfsdkTags[attrName], "schema attribute %q has no corresponding tfsdk-tagged model field", attrName)
	}
}

func collectTFSDKTags(rt reflect.Type, out map[string]bool) {
	for f := range rt.Fields() {
		if f.Anonymous {
			ft := f.Type
			if ft.Kind() == reflect.Struct {
				collectTFSDKTags(ft, out)
			}
			continue
		}
		if tag, ok := f.Tag.Lookup("tfsdk"); ok && tag != "" && tag != "-" {
			out[tag] = true
		}
	}
}

const requiresReplaceDescription = "If the value of this attribute changes, Terraform will destroy and recreate the resource."
const useStateForUnknownDescription = "Once set, the value of this attribute in state will not change."

func hasStringRequiresReplace(mods []planmodifier.String) bool {
	for _, m := range mods {
		if m.Description(context.Background()) == requiresReplaceDescription {
			return true
		}
	}
	return false
}

func hasStringUseStateForUnknown(mods []planmodifier.String) bool {
	for _, m := range mods {
		if m.Description(context.Background()) == useStateForUnknownDescription {
			return true
		}
	}
	return false
}

// TestGetSchema_descriptionAndNamespaceRejectEmptyString covers a review
// finding: populateFromPackagePolicy/populateFromCreateResponse fold the
// API's `""` back to null for both description and namespace (see
// typeutils.NonEmptyStringOrNull in models_convert.go), because Kibana
// returns an explicit "" (not an omitted field) once a description has been
// cleared. Without a validator, a config that explicitly sets
// `description = ""` or `namespace = ""` would be indistinguishable from
// "unset" after the first apply's Read, producing a permanent,
// non-converging diff. A LengthAtLeast(1) validator rejects the empty
// string upfront instead.
func TestGetSchema_descriptionAndNamespaceRejectEmptyString(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := getSchema(ctx)

	description, ok := s.Attributes["description"].(schema.StringAttribute)
	require.True(t, ok)
	diags := validateStringValidators(ctx, description.Validators, types.StringValue(""), path.Root("description"))
	assert.True(t, diags.HasError(), "description = \"\" should be rejected by a validator, not silently accepted")

	namespace, ok := s.Attributes["namespace"].(schema.StringAttribute)
	require.True(t, ok)
	diags = validateStringValidators(ctx, namespace.Validators, types.StringValue(""), path.Root("namespace"))
	assert.True(t, diags.HasError(), "namespace = \"\" should be rejected by a validator, not silently accepted")
}

func validateStringValidators(ctx context.Context, validators []validator.String, value types.String, p path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	req := validator.StringRequest{ConfigValue: value, Path: p}
	for _, v := range validators {
		var resp validator.StringResponse
		v.ValidateString(ctx, req, &resp)
		diags.Append(resp.Diagnostics...)
	}
	return diags
}
