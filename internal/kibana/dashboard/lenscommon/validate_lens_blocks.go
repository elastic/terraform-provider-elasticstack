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

package lenscommon

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ValidateLensBlocks returns a non-nil Diagnostics error if blocks is nil.
// Call at the top of every PopulateFromAttributes implementation before writing to blocks.
func ValidateLensBlocks(blocks *models.LensByValueChartBlocks, configName string) diag.Diagnostics {
	if blocks == nil {
		var d diag.Diagnostics
		d.AddError("Lens chart blocks missing",
			fmt.Sprintf("cannot populate %s without chart blocks", configName))
		return d
	}
	return nil
}
