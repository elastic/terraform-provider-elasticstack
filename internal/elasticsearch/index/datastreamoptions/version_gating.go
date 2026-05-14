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

package datastreamoptions

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnforceMinServerVersion returns an error diagnostic when the template object
// carries a configured data_stream_options block and the cluster is older than
// index.MinSupportedDataStreamOptionsVersion. Shared by index templates and
// component templates; both keep the data_stream_options child under template.
func EnforceMinServerVersion(tmplObj types.Object, serverVersion *version.Version) diag.Diagnostics {
	var diags diag.Diagnostics
	if serverVersion == nil {
		return diags
	}
	if tmplObj.IsNull() || tmplObj.IsUnknown() {
		return diags
	}
	dsoVal, ok := tmplObj.Attributes()["data_stream_options"]
	if !ok {
		return diags
	}
	if dsoVal.IsNull() || dsoVal.IsUnknown() {
		return diags
	}
	// Distinguish "block absent" from "block present"; unknown nested object still triggers gate when known non-null.
	if _, ok := dsoVal.(types.Object); !ok {
		return diags
	}
	if serverVersion.LessThan(index.MinSupportedDataStreamOptionsVersion) {
		diags.AddError(
			"Unsupported Elasticsearch version",
			fmt.Sprintf("'data_stream_options' is supported only for Elasticsearch v%s and above", index.MinSupportedDataStreamOptionsVersion.String()),
		)
	}
	return diags
}