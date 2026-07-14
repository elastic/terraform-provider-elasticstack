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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultKafkaGzipCompressionLevel int64 = 4

func kafkaCompressionLevelDefaultIfGzip() planmodifier.Int64 {
	return kafkaCompressionLevelDefaultModifier{defaultLevel: defaultKafkaGzipCompressionLevel}
}

type kafkaCompressionLevelDefaultModifier struct {
	defaultLevel int64
}

func (m kafkaCompressionLevelDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf(
		"Sets kafka.compression_level to %d when kafka.compression is gzip and no level is configured.",
		m.defaultLevel,
	)
}

func (m kafkaCompressionLevelDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m kafkaCompressionLevelDefaultModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	resp.PlanValue = kafkaCompressionLevelDefaultPlanValue(
		kafkaCompressionFromPlanOrConfig(ctx, req.Plan, req.Config, req.Path, resp),
		req.PlanValue,
		req.ConfigValue,
		m.defaultLevel,
	)
}

func kafkaCompressionFromPlanOrConfig(
	ctx context.Context,
	plan tfsdk.Plan,
	config tfsdk.Config,
	currentPath path.Path,
	resp *planmodifier.Int64Response,
) types.String {
	compressionPath := currentPath.ParentPath().AtName("compression")

	var compression types.String
	diags := plan.GetAttribute(ctx, compressionPath, &compression)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return types.StringUnknown()
	}

	if !compression.IsUnknown() && !compression.IsNull() {
		return compression
	}

	diags = config.GetAttribute(ctx, compressionPath, &compression)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return types.StringUnknown()
	}

	return compression
}

func kafkaCompressionLevelDefaultPlanValue(
	compression types.String,
	planValue, configValue types.Int64,
	defaultLevel int64,
) types.Int64 {
	if !planValue.IsUnknown() {
		return planValue
	}

	if configValue.IsUnknown() {
		return planValue
	}

	if !configValue.IsNull() {
		return planValue
	}

	if compression.ValueString() == kafkaCompressionGzip {
		return types.Int64Value(defaultLevel)
	}

	return planValue
}
