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

	for _, name := range []string{attrAWSBlock, attrAzureBlock, attrVarsMap} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			switch attr := s.Attributes[name].(type) {
			case schema.SingleNestedAttribute:
				assert.True(t, attr.Optional, "%s should be optional", name)
				assert.True(t, attr.Computed, "%s should be computed", name)
			case schema.MapNestedAttribute:
				assert.True(t, attr.Optional, "%s should be optional", name)
				assert.True(t, attr.Computed, "%s should be computed", name)
			default:
				t.Fatalf("unexpected attribute type for %q: %T", name, s.Attributes[name])
			}
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

func TestGetSchemaWriteOnlySensitiveAttributes(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())

	awsAttr, ok := s.Attributes[attrAWSBlock].(schema.SingleNestedAttribute)
	require.True(t, ok)
	externalID, ok := awsAttr.Attributes[attrAWSExternalID].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, externalID.WriteOnly)
	assert.True(t, externalID.Sensitive)
	assert.False(t, externalID.Computed)

	varsAttr, ok := s.Attributes[attrVarsMap].(schema.MapNestedAttribute)
	require.True(t, ok)
	secretValue, ok := varsAttr.NestedObject.Attributes[attrVarsSecretValue].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, secretValue.WriteOnly)
	assert.True(t, secretValue.Sensitive)
	assert.False(t, secretValue.Computed)
}

func TestGetSchemaUseStateForUnknownOnDualPopulationFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	s := getSchema(ctx)

	awsAttr, ok := s.Attributes[attrAWSBlock].(schema.SingleNestedAttribute)
	require.True(t, ok)
	assert.NotEmpty(t, awsAttr.PlanModifiers, "aws block should have plan modifiers")

	azureAttr, ok := s.Attributes[attrAzureBlock].(schema.SingleNestedAttribute)
	require.True(t, ok)
	assert.NotEmpty(t, azureAttr.PlanModifiers, "azure block should have plan modifiers")

	varsAttr, ok := s.Attributes[attrVarsMap].(schema.MapNestedAttribute)
	require.True(t, ok)
	assert.NotEmpty(t, varsAttr.PlanModifiers, "vars map should have plan modifiers")

	assert.True(t, hasUseStateForUnknownObjectPlanModifier(ctx, awsAttr.PlanModifiers))
	assert.True(t, hasUseStateForUnknownObjectPlanModifier(ctx, azureAttr.PlanModifiers))
	assert.True(t, hasUseStateForUnknownMapPlanModifier(ctx, varsAttr.PlanModifiers))
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

func hasUseStateForUnknownMapPlanModifier(ctx context.Context, modifiers []planmodifier.Map) bool {
	for _, modifier := range modifiers {
		desc := strings.ToLower(modifier.Description(ctx))
		if strings.Contains(desc, "once set") {
			return true
		}
	}
	return false
}
