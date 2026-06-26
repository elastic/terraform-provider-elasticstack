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

package tag

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const managedTagErrorDetail = "This tag is managed by Kibana and cannot be controlled by the elasticstack_kibana_tag resource. " +
	"Use the elasticstack_kibana_tags data source to read managed tags."

func managedTagDiagnostic() diag.Diagnostic {
	return diag.NewErrorDiagnostic("Managed Kibana tag", managedTagErrorDetail)
}

func checkManagedTag(detail *kibanaoapi.TagDetail) diag.Diagnostics {
	if detail != nil && detail.Managed != nil && *detail.Managed {
		return diag.Diagnostics{managedTagDiagnostic()}
	}
	return nil
}
