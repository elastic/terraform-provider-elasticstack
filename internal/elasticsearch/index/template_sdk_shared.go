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

package index

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// validateDataStreamOptionsVersion returns an error diagnostic if data_stream_options is configured and the server is too old.
// Used by tests; index template resource logic lives in the Plugin Framework template package.
func validateDataStreamOptionsVersion(serverVersion *version.Version, templ *models.Template) diag.Diagnostics {
	if templ != nil && templ.DataStreamOptions != nil && serverVersion.LessThan(MinSupportedDataStreamOptionsVersion) {
		return diag.FromErr(fmt.Errorf("'data_stream_options' is supported only for Elasticsearch v%s and above", MinSupportedDataStreamOptionsVersion.String()))
	}
	return nil
}
