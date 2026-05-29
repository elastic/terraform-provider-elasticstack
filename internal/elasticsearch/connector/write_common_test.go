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

package connector

import (
	"context"
	"testing"

	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestApiKeyChanged(t *testing.T) {
	t.Parallel()

	prior := &ContentConnectorData{
		APIKeyID:       fwtypes.StringValue("a"),
		APIKeySecretID: fwtypes.StringValue("b"),
	}
	unchanged := ContentConnectorData{
		APIKeyID:       fwtypes.StringValue("a"),
		APIKeySecretID: fwtypes.StringValue("b"),
	}
	require.False(t, apiKeyChanged(unchanged, prior))

	changedID := ContentConnectorData{
		APIKeyID:       fwtypes.StringValue("z"),
		APIKeySecretID: fwtypes.StringValue("b"),
	}
	require.True(t, apiKeyChanged(changedID, prior))

	require.True(t, apiKeyChanged(ContentConnectorData{}, nil))
}

func TestConfigurationSchemaNotRegisteredDiagnostic(t *testing.T) {
	t.Parallel()
	require.Equal(t, configurationSchemaNotRegisteredTitle, "Connector configuration schema not yet registered")
	detail := configurationSchemaNotRegisteredDetail("postgresql")
	require.Contains(t, detail, `service_type "postgresql"`)
	require.Contains(t, detail, configurationSchemaNotRegisteredURL)
}

func TestSkipAspectOnUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	nullObj := fwtypes.ObjectNull(pipelineModelAttrTypes())
	knownObj, diags := fwtypes.ObjectValueFrom(ctx, pipelineModelAttrTypes(), PipelineModel{
		Name:                 fwtypes.StringValue("p"),
		ExtractBinaryContent: fwtypes.BoolValue(true),
		ReduceWhitespace:     fwtypes.BoolValue(true),
		RunMlInference:       fwtypes.BoolValue(false),
	})
	require.False(t, diags.HasError())

	require.True(t, skipAspectOnUpdate(nullObj, nullObj))
	require.False(t, skipAspectOnUpdate(knownObj, nullObj))
	require.False(t, skipAspectOnUpdate(knownObj, knownObj))
}

func TestPlanObjectSet_andPlanMapSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	require.False(t, planObjectSet(fwtypes.ObjectNull(pipelineModelAttrTypes())))
	require.False(t, planMapSet(fwtypes.MapNull(fwtypes.ObjectType{AttrTypes: configurationValueModelAttrTypes()})))

	knownMap, diags := fwtypes.MapValueFrom(ctx, fwtypes.ObjectType{AttrTypes: configurationValueModelAttrTypes()}, map[string]ConfigurationValueModel{
		"k": {String: fwtypes.StringValue("v")},
	})
	require.False(t, diags.HasError())
	require.True(t, planMapSet(knownMap))
}
