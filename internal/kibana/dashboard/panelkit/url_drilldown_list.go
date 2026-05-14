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

package panelkit

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// drilldownListMaxSize is the Kibana Dashboard API's per-panel drilldown cap.
const drilldownListMaxSize = 100

// URLDrilldownListAttribute returns an optional ListNestedAttribute whose element is
// URLDrilldownSchema(opts), with the Kibana API's 100-item size cap pre-applied.
func URLDrilldownListAttribute(markdownDescription string, opts URLDrilldownOptions) schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: markdownDescription,
		Optional:            true,
		NestedObject:        URLDrilldownSchema(opts),
		Validators: []validator.List{
			listvalidator.SizeAtMost(drilldownListMaxSize),
		},
	}
}
