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

package cloudconnector

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSchemaTopLevelAttributes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	s := getSchema(ctx)

	expected := []string{
		attrID,
		attrCloudConnectorID,
		attrSpaceID,
		attrName,
		attrCloudProvider,
		attrAccountType,
		attrForceDelete,
		attrAWSBlock,
		attrAzureBlock,
		attrVarsMap,
		attrNamespace,
		attrPackagePolicyCount,
		attrVerificationStatus,
		attrVerificationStartedAt,
		attrVerificationFailedAt,
		attrCreatedAt,
		attrUpdatedAt,
	}

	for _, name := range expected {
		_, ok := s.Attributes[name]
		assert.True(t, ok, "expected top-level attribute %q", name)
	}
	assert.Len(t, s.Attributes, len(expected))
}

func TestGetSchemaDualPopulationAttributesOptionalComputed(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())

	varsAttr, ok := s.Attributes[attrVarsMap].(schema.MapNestedAttribute)
	require.True(t, ok)
	assert.True(t, varsAttr.Optional, "%s should be optional", attrVarsMap)
	assert.True(t, varsAttr.Computed, "%s should be optional+computed for dual representation", attrVarsMap)
}

func TestGetSchemaWriteOnlyParentBlocksNotComputed(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())

	for _, name := range []string{attrAWSBlock, attrAzureBlock} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			attr, ok := s.Attributes[name].(schema.SingleNestedAttribute)
			require.True(t, ok)
			assert.True(t, attr.Optional, "%s should be optional", name)
			assert.False(t, attr.Computed, "%s must not be computed when it contains write-only children", name)
		})
	}
}
func TestGetSchemaRequiresReplacePlanModifiers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	s := getSchema(ctx)

	for _, name := range []string{attrCloudProvider, attrCloudConnectorID, attrSpaceID} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			attr, ok := s.Attributes[name].(schema.StringAttribute)
			require.True(t, ok, "expected %q to be a StringAttribute", name)
			assert.True(t, hasRequiresReplaceStringPlanModifier(ctx, attr.PlanModifiers), "expected RequiresReplace on %q", name)
		})
	}
}

func TestGetSchemaSecretSensitiveAttributes(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())

	awsAttr, ok := s.Attributes[attrAWSBlock].(schema.SingleNestedAttribute)
	require.True(t, ok)
	externalID, ok := awsAttr.Attributes[attrAWSExternalID].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, externalID.WriteOnly)
	assert.True(t, externalID.Sensitive)
	assert.False(t, externalID.Computed)

	azureAttr, ok := s.Attributes[attrAzureBlock].(schema.SingleNestedAttribute)
	require.True(t, ok)
	tenantID, ok := azureAttr.Attributes[attrAzureTenantID].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, tenantID.WriteOnly)
	assert.True(t, tenantID.Sensitive)
	assert.False(t, tenantID.Computed)
	clientID, ok := azureAttr.Attributes[attrAzureClientID].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, clientID.WriteOnly)
	assert.True(t, clientID.Sensitive)
	assert.False(t, clientID.Computed)

	varsAttr, ok := s.Attributes[attrVarsMap].(schema.MapNestedAttribute)
	require.True(t, ok)
	secretValue, ok := varsAttr.NestedObject.Attributes[attrVarsSecretValue].(schema.StringAttribute)
	require.True(t, ok)
	assert.False(t, secretValue.WriteOnly)
	assert.True(t, secretValue.Sensitive)
	assert.False(t, secretValue.Computed)
}

func TestGetSchemaUseStateForUnknownOnDualPopulationFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	s := getSchema(ctx)

	azureAttr, ok := s.Attributes[attrAzureBlock].(schema.SingleNestedAttribute)
	require.True(t, ok)
	for _, name := range []string{attrAzureTenantIDSecretRef, attrAzureClientIDSecretRef} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			attr, ok := azureAttr.Attributes[name].(schema.SingleNestedAttribute)
			require.True(t, ok)
			assert.NotEmpty(t, attr.PlanModifiers, "%s should have plan modifiers", name)
			assert.True(t, hasUseStateForUnknownObjectPlanModifier(ctx, attr.PlanModifiers))
		})
	}
}

func TestGetSchemaVarsInnerDualPopulationAttributesOptionalComputed(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())

	varsAttr, ok := s.Attributes[attrVarsMap].(schema.MapNestedAttribute)
	require.True(t, ok)

	for _, name := range []string{attrVarsType, attrVarsValue, attrVarsFrozen} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			switch attr := varsAttr.NestedObject.Attributes[name].(type) {
			case schema.StringAttribute:
				assert.True(t, attr.Optional, "%s should be optional", name)
				assert.True(t, attr.Computed, "%s should be computed", name)
			case schema.BoolAttribute:
				assert.True(t, attr.Optional, "%s should be optional", name)
				assert.True(t, attr.Computed, "%s should be computed", name)
			default:
				t.Fatalf("unexpected attribute type for %q: %T", name, varsAttr.NestedObject.Attributes[name])
			}
		})
	}
}

func hasRequiresReplaceStringPlanModifier(ctx context.Context, modifiers []planmodifier.String) bool {
	for _, modifier := range modifiers {
		desc := strings.ToLower(modifier.Description(ctx))
		if strings.Contains(desc, "destroy and recreate") {
			return true
		}
	}
	return false
}

func hasUseStateForUnknownObjectPlanModifier(ctx context.Context, modifiers []planmodifier.Object) bool {
	for _, modifier := range modifiers {
		desc := strings.ToLower(modifier.Description(ctx))
		if strings.Contains(desc, "once set") {
			return true
		}
	}
	return false
}
