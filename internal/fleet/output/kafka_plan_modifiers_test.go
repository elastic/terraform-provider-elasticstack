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

package output

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_kafkaCompressionLevelDefaultPlanValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		compression types.String
		planValue   types.Int64
		configValue types.Int64
		want        types.Int64
	}{
		{
			name:        "keeps known plan value",
			compression: types.StringValue("gzip"),
			planValue:   types.Int64Value(6),
			configValue: types.Int64Null(),
			want:        types.Int64Value(6),
		},
		{
			name:        "defaults unknown plan value for gzip",
			compression: types.StringValue("gzip"),
			planValue:   types.Int64Unknown(),
			configValue: types.Int64Null(),
			want:        types.Int64Value(defaultKafkaGzipCompressionLevel),
		},
		{
			name:        "leaves unknown plan value for non-gzip compression",
			compression: types.StringValue("snappy"),
			planValue:   types.Int64Unknown(),
			configValue: types.Int64Null(),
			want:        types.Int64Unknown(),
		},
		{
			name:        "does not override unknown config interpolation",
			compression: types.StringValue("gzip"),
			planValue:   types.Int64Unknown(),
			configValue: types.Int64Unknown(),
			want:        types.Int64Unknown(),
		},
		{
			name:        "respects explicit config value",
			compression: types.StringValue("gzip"),
			planValue:   types.Int64Unknown(),
			configValue: types.Int64Value(7),
			want:        types.Int64Unknown(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := kafkaCompressionLevelDefaultPlanValue(
				tt.compression,
				tt.planValue,
				tt.configValue,
				defaultKafkaGzipCompressionLevel,
			)
			assert.True(t, got.Equal(tt.want))
		})
	}
}

func Test_kafkaCompressionLevelDefaultIfGzip_description(t *testing.T) {
	t.Parallel()

	modifier := kafkaCompressionLevelDefaultIfGzip()
	assert.Contains(t, modifier.Description(context.Background()), "4")
	assert.Equal(t, modifier.Description(context.Background()), modifier.MarkdownDescription(context.Background()))
}

func Test_schemaKafkaCompressionLevelHasDefaultPlanModifier(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	attr, ok := getSchema(ctx).
		Attributes["kafka"].(schema.SingleNestedAttribute).
		Attributes["compression_level"].(schema.Int64Attribute)
	require.True(t, ok)

	found := false
	for _, modifier := range attr.PlanModifiers {
		if modifier.Description(ctx) == kafkaCompressionLevelDefaultIfGzip().Description(ctx) {
			found = true
			break
		}
	}

	assert.True(t, found, "expected kafkaCompressionLevelDefaultIfGzip plan modifier on compression_level")
}
