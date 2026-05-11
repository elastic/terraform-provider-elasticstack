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

package snapshot_repository

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// blockRequiredAttrValidator is a validator.Object placed on a SingleNestedBlock
// to enforce that specific attributes within the block are non-null when the
// block is present in the configuration.  Because the enclosing block is
// optional, the framework does not allow marking its child attributes as
// Required (that would demand them unconditionally).  This validator bridges
// that gap: when the block IS configured the named attributes are mandatory.
type blockRequiredAttrValidator struct {
	attrNames []string
}

func requireBlockAttrs(attrNames ...string) validator.Object {
	return blockRequiredAttrValidator{attrNames: attrNames}
}

func (v blockRequiredAttrValidator) Description(_ context.Context) string {
	return fmt.Sprintf("when block is configured, the following attributes are required: %s", strings.Join(v.attrNames, ", "))
}

func (v blockRequiredAttrValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v blockRequiredAttrValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// Block not present in config – nothing to validate.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	for _, name := range v.attrNames {
		val, ok := attrs[name]
		if !ok || val.IsUnknown() {
			// Unknown means it will be resolved later; skip.
			continue
		}
		if val.IsNull() {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName(name),
				"Missing required attribute",
				fmt.Sprintf("The %q attribute is required when the block is configured.", name),
			)
		}
	}
}

var urlProtocolRegex = regexp.MustCompile("^(file:|ftp:|http:|https:|jar:)")

// s3EndpointValidator validates that an S3 endpoint is a valid HTTP/HTTPS URL.
type s3EndpointValidator struct{}

func (v s3EndpointValidator) Description(_ context.Context) string {
	return "Value must be a valid HTTP/HTTPS URL"
}

func (v s3EndpointValidator) MarkdownDescription(_ context.Context) string {
	return "Value must be a valid HTTP/HTTPS URL"
}

func (v s3EndpointValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	if val == "" {
		return
	}
	parsed, err := url.ParseRequestURI(val)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid S3 endpoint",
			fmt.Sprintf("%q must be a valid HTTP/HTTPS URL", val),
		)
		return
	}
	if !strings.HasPrefix(val, "http://") && !strings.HasPrefix(val, "https://") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid S3 endpoint",
			fmt.Sprintf("%q must start with http:// or https://", val),
		)
	}
}

// validateConfigExactlyOneType ensures exactly one repository type block is set.
type validateConfigExactlyOneType struct{}

var typeBlockNames = []string{"fs", "url", "gcs", "azure", "s3", "hdfs"}

func (v validateConfigExactlyOneType) Description(_ context.Context) string {
	return "Exactly one repository type block must be set"
}

func (v validateConfigExactlyOneType) MarkdownDescription(_ context.Context) string {
	return "Exactly one repository type block must be set"
}

func (v validateConfigExactlyOneType) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data Data
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation when any type block is unknown (e.g. derived from a variable).
	if data.Fs.IsUnknown() || data.URL.IsUnknown() || data.Gcs.IsUnknown() || data.Azure.IsUnknown() || data.S3.IsUnknown() || data.Hdfs.IsUnknown() {
		return
	}

	set := 0
	if !data.Fs.IsNull() {
		set++
	}
	if !data.URL.IsNull() {
		set++
	}
	if !data.Gcs.IsNull() {
		set++
	}
	if !data.Azure.IsNull() {
		set++
	}
	if !data.S3.IsNull() {
		set++
	}
	if !data.Hdfs.IsNull() {
		set++
	}

	if set == 0 {
		resp.Diagnostics.AddError(
			"Missing repository type",
			fmt.Sprintf("Exactly one of the following blocks must be set: %s", strings.Join(typeBlockNames, ", ")),
		)
	} else if set > 1 {
		resp.Diagnostics.AddError(
			"Multiple repository types",
			fmt.Sprintf("Exactly one of the following blocks must be set, but %d are set: %s", set, strings.Join(typeBlockNames, ", ")),
		)
	}
}
