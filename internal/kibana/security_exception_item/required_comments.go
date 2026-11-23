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

package securityexceptionitem

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func requiredCommentValidators() []validator.List {
	return []validator.List{
		listvalidator.ValueStringsAre(nonWhitespaceStringValidator{}),
		uniqueValuesValidator{},
	}
}

type nonWhitespaceStringValidator struct{}

func (nonWhitespaceStringValidator) Description(_ context.Context) string {
	return "value must not be empty or contain only whitespace"
}

func (nonWhitespaceStringValidator) MarkdownDescription(ctx context.Context) string {
	return nonWhitespaceStringValidator{}.Description(ctx)
}

func (nonWhitespaceStringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Unknown values are validated once resolved, but a null element must be
	// treated like an empty one: it would otherwise reach the API as a missing
	// comment body instead of a plan-time diagnostic.
	if req.ConfigValue.IsUnknown() {
		return
	}
	if strings.TrimSpace(req.ConfigValue.ValueString()) == "" {
		resp.Diagnostics.AddError(
			"Invalid required_comments entry",
			"Comment text must not be empty or contain only whitespace characters.",
		)
	}
}

type uniqueValuesValidator struct{}

func (uniqueValuesValidator) Description(_ context.Context) string {
	return "values must be unique"
}

func (uniqueValuesValidator) MarkdownDescription(ctx context.Context) string {
	return uniqueValuesValidator{}.Description(ctx)
}

func (uniqueValuesValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}
	seen := make(map[string]struct{}, len(req.ConfigValue.Elements()))
	for _, elem := range req.ConfigValue.Elements() {
		if elem.(types.String).IsUnknown() {
			continue
		}
		text := elem.(types.String).ValueString()
		if _, ok := seen[text]; ok {
			resp.Diagnostics.AddError(
				"Duplicate required_comments entry",
				"The same comment text appears more than once in required_comments: "+text,
			)
			continue
		}
		seen[text] = struct{}{}
	}
}

// minAppendOnlyCommentBodyVersion is the first Kibana version whose update
// endpoint accepts a comments array containing only the new texts. Earlier
// versions treat the array as the full desired log, so they require the
// stored comments echoed back with their ids (elastic/kibana#145575, fixed
// in 8.6.0).
var minAppendOnlyCommentBodyVersion = version.Must(version.NewVersion("8.6.0"))

func supportsMissingOnlyCommentBody(kibanaVersion *version.Version) bool {
	return kibanaVersion.GreaterThanOrEqual(minAppendOnlyCommentBodyVersion)
}

func missingCommentTexts(required, storedTexts []string) []string {
	stored := make(map[string]struct{}, len(storedTexts))
	for _, text := range storedTexts {
		stored[text] = struct{}{}
	}
	var missing []string
	for _, text := range required {
		if _, ok := stored[text]; !ok {
			missing = append(missing, text)
		}
	}
	return missing
}

// buildAppendComments builds the update-request comments array that appends
// the missing texts. Servers before 8.6.0 require the stored comments echoed
// back as {id, comment} in stored order before the new texts (they treat the
// array as the full desired log); servers 8.6.0 and later accept the missing
// texts alone. New texts never carry an id and never precede id-bearing
// entries, and a stored text is never resent without its id.
func buildAppendComments(
	stored []kbapi.SecurityExceptionsAPIExceptionListItemComment,
	missing []string,
	missingOnly bool,
) kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray {
	body := make(kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray, 0, len(stored)+len(missing))
	if !missingOnly {
		for _, comment := range stored {
			id := comment.Id
			body = append(body, kbapi.SecurityExceptionsAPIUpdateExceptionListItemComment{
				Id:      &id,
				Comment: comment.Comment,
			})
		}
	}
	for _, text := range missing {
		body = append(body, kbapi.SecurityExceptionsAPIUpdateExceptionListItemComment{
			Comment: text,
		})
	}
	return body
}

type commentReconciler struct {
	readStoredComments      func(ctx context.Context) (stored []kbapi.SecurityExceptionsAPIExceptionListItemComment, found bool, diags diag.Diagnostics)
	supportsMissingOnlyBody func(ctx context.Context) (bool, diag.Diagnostics)
}

func (r commentReconciler) reconcile(ctx context.Context, required []string) (*kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray, diag.Diagnostics) {
	var diags diag.Diagnostics

	stored, found, d := r.readStoredComments(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	if !found {
		diags.AddError(
			"Failed to read exception item comments",
			"The exception item could not be read back before update, so its current comment log is unknown and required_comments cannot be reconciled.",
		)
		return nil, diags
	}

	storedTexts := make([]string, len(stored))
	for i, comment := range stored {
		storedTexts[i] = comment.Comment
	}
	missing := missingCommentTexts(required, storedTexts)
	if len(missing) == 0 {
		return nil, diags
	}

	missingOnly, d := r.supportsMissingOnlyBody(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	body := buildAppendComments(stored, missing, missingOnly)
	return &body, diags
}

func requiredCommentTexts(list types.List) ([]string, bool) {
	if !typeutils.IsKnown(list) {
		return nil, false
	}
	texts := make([]string, 0, len(list.Elements()))
	for _, elem := range list.Elements() {
		texts = append(texts, elem.(types.String).ValueString())
	}
	return texts, true
}
