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

package validators

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type uuidValidator struct{}

func IsUUID() validator.String {
	return uuidValidator{}
}

func (v uuidValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}

	_, err := uuid.ParseUUID(req.ConfigValue.ValueString())
	if err == nil {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid UUID",
		fmt.Sprintf("Expected a valid UUID, got %s. Parsing error: %v", req.ConfigValue.ValueString(), err),
	)
}

func (v uuidValidator) Description(_ context.Context) string {
	return "value must be a valid UUID in RFC 4122 format"
}

func (v uuidValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
