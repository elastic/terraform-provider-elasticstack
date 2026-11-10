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

package rule

import (
	"context"
	"encoding/json"
	"strconv"
	"unicode/utf8"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	maxArtifactDataKeys      = 32
	maxArtifactDataKeyLength = 256
)

// artifactDataValidator accepts an artifacts[].data value that is a JSON
// object with at most maxArtifactDataKeys top-level keys, each 1 to
// maxArtifactDataKeyLength characters long.
func artifactDataValidator() validator.String {
	return artifactDataJSONObject{}
}

type artifactDataJSONObject struct{}

func (artifactDataJSONObject) Description(_ context.Context) string {
	return "must be a JSON object with at most " + strconv.Itoa(maxArtifactDataKeys) +
		" top-level keys, each 1-" + strconv.Itoa(maxArtifactDataKeyLength) + " characters long"
}

func (v artifactDataJSONObject) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (artifactDataJSONObject) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}
	validators.StringIsJSONObject{}.ValidateString(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &object); err != nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(req.Path, "must be a JSON object", err.Error()))
		return
	}
	if len(object) > maxArtifactDataKeys {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(req.Path,
			"must have at most "+strconv.Itoa(maxArtifactDataKeys)+" top-level keys", strconv.Itoa(len(object))+" keys"))
		return
	}
	for key := range object {
		if length := utf8.RuneCountInString(key); length < 1 || length > maxArtifactDataKeyLength {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(req.Path,
				"must have top-level keys of 1-"+strconv.Itoa(maxArtifactDataKeyLength)+" characters", `key "`+key+`"`))
			return
		}
	}
}

// artifactIDsUniqueValidator rejects an artifacts list in which two elements
// share an id. Unknown elements and ids are skipped until they resolve.
func artifactIDsUniqueValidator() validator.List {
	return artifactIDsUnique{}
}

type artifactIDsUnique struct{}

func (artifactIDsUnique) Description(_ context.Context) string {
	return "artifact ids must be unique within the rule"
}

func (v artifactIDsUnique) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (artifactIDsUnique) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}
	seen := make(map[string]struct{}, len(req.ConfigValue.Elements()))
	for index, element := range req.ConfigValue.Elements() {
		artifact, ok := element.(types.Object)
		if !ok || !typeutils.IsKnown(artifact) {
			continue
		}
		id, ok := artifact.Attributes()["id"].(types.String)
		if !ok || !typeutils.IsKnown(id) {
			continue
		}
		if _, duplicate := seen[id.ValueString()]; duplicate {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtListIndex(index).AtName("id"),
				"Duplicate artifact id",
				`Artifact id "`+id.ValueString()+`" is used more than once. Each artifact id must be unique within the rule.`,
			)
			continue
		}
		seen[id.ValueString()] = struct{}{}
	}
}

// nonEmptyListValidator rejects a configured empty list. The API rejects
// empty arrays, so the attribute must be omitted instead.
func nonEmptyListValidator() validator.List {
	return nonEmptyList{}
}

type nonEmptyList struct{}

func (nonEmptyList) Description(_ context.Context) string {
	return "must contain at least one item; omit the attribute instead of configuring an empty list"
}

func (v nonEmptyList) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (nonEmptyList) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if !typeutils.IsKnown(req.ConfigValue) || len(req.ConfigValue.Elements()) > 0 {
		return
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Empty list not allowed",
		"The list must contain at least one item. The API rejects empty lists, so omit the attribute instead of setting it to [].",
	)
}
