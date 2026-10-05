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

package alias

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	frameworkvalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func schemaWriteIndexNameValidator(t *testing.T) frameworkvalidator.String {
	t.Helper()

	resourceSchema := getSchemaFactory(context.Background())
	writeIndex := resourceSchema.Attributes["write_index"].(schema.SingleNestedAttribute)
	name := writeIndex.Attributes[attrName].(schema.StringAttribute)

	require.Len(t, name.Validators, 1)
	return name.Validators[0]
}

func TestWriteIndexNameValidator_SkipsUnknownValues(t *testing.T) {
	t.Parallel()

	var resp frameworkvalidator.StringResponse
	schemaWriteIndexNameValidator(t).ValidateString(context.Background(), frameworkvalidator.StringRequest{
		ConfigValue: types.StringUnknown(),
	}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
}

func TestWriteIndexNameValidator_RejectsWildcard(t *testing.T) {
	t.Parallel()

	var resp frameworkvalidator.StringResponse
	schemaWriteIndexNameValidator(t).ValidateString(context.Background(), frameworkvalidator.StringRequest{
		ConfigValue: types.StringValue("logs-*"),
	}, &resp)

	require.True(t, resp.Diagnostics.HasError())
}

func TestWriteIndexNameValidator_RejectsQuestionMark(t *testing.T) {
	t.Parallel()

	var resp frameworkvalidator.StringResponse
	schemaWriteIndexNameValidator(t).ValidateString(context.Background(), frameworkvalidator.StringRequest{
		ConfigValue: types.StringValue("logs-?"),
	}, &resp)

	require.True(t, resp.Diagnostics.HasError())
}

func TestWriteIndexNameValidator_RejectsCommaList(t *testing.T) {
	t.Parallel()

	var resp frameworkvalidator.StringResponse
	schemaWriteIndexNameValidator(t).ValidateString(context.Background(), frameworkvalidator.StringRequest{
		ConfigValue: types.StringValue("logs-1,logs-2"),
	}, &resp)

	require.True(t, resp.Diagnostics.HasError())
}

func TestWriteIndexNameValidator_RejectsLeadingExclusion(t *testing.T) {
	t.Parallel()

	var resp frameworkvalidator.StringResponse
	schemaWriteIndexNameValidator(t).ValidateString(context.Background(), frameworkvalidator.StringRequest{
		ConfigValue: types.StringValue("-logs-1"),
	}, &resp)

	require.True(t, resp.Diagnostics.HasError())
}

func TestWriteIndexNameValidator_RejectsDateMath(t *testing.T) {
	t.Parallel()

	var resp frameworkvalidator.StringResponse
	schemaWriteIndexNameValidator(t).ValidateString(context.Background(), frameworkvalidator.StringRequest{
		ConfigValue: types.StringValue("<logs-{now/d}>"),
	}, &resp)

	require.True(t, resp.Diagnostics.HasError())
}

func TestWriteIndexNameValidator_RejectsAllIndicesSelector(t *testing.T) {
	t.Parallel()

	var resp frameworkvalidator.StringResponse
	schemaWriteIndexNameValidator(t).ValidateString(context.Background(), frameworkvalidator.StringRequest{
		ConfigValue: types.StringValue("_all"),
	}, &resp)

	require.True(t, resp.Diagnostics.HasError())
}

func TestWriteIndexNameValidator_AcceptsSingleTargetNames(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"my.index_1-v2", ".internal-index", "logs-current", "_system_index"} {
		t.Run(name, func(t *testing.T) {
			var resp frameworkvalidator.StringResponse
			schemaWriteIndexNameValidator(t).ValidateString(context.Background(), frameworkvalidator.StringRequest{
				ConfigValue: types.StringValue(name),
			}, &resp)

			require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
		})
	}
}
