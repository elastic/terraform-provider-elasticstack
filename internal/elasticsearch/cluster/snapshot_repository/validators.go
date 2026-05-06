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

	set := 0
	if !data.Fs.IsNull() && !data.Fs.IsUnknown() {
		set++
	}
	if !data.Url.IsNull() && !data.Url.IsUnknown() {
		set++
	}
	if !data.Gcs.IsNull() && !data.Gcs.IsUnknown() {
		set++
	}
	if !data.Azure.IsNull() && !data.Azure.IsUnknown() {
		set++
	}
	if !data.S3.IsNull() && !data.S3.IsUnknown() {
		set++
	}
	if !data.Hdfs.IsNull() && !data.Hdfs.IsUnknown() {
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
