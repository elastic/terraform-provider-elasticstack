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

package alertingrule

import (
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                   = newResource()
	_ resource.ResourceWithConfigure      = newResource()
	_ resource.ResourceWithImportState    = newResource()
	_ resource.ResourceWithValidateConfig = newResource()
	_ resource.ResourceWithUpgradeState   = newResource()
	_ resource.ResourceWithModifyPlan     = newResource()
)

//go:embed resource-description.md
var resourceDescription string

type Resource struct {
	*entitycore.KibanaResource[alertingRuleModel]
	*entitycore.KibanaSpaceImporter
}

func newResource() *Resource {
	return &Resource{
		KibanaResource: entitycore.NewKibanaResource[alertingRuleModel](
			entitycore.ComponentKibana,
			"alerting_rule",
			entitycore.KibanaResourceOptions[alertingRuleModel]{
				Schema: getSchema,
				Read:   readAlertingRule,
				Delete: entitycore.SimpleKibanaDelete[alertingRuleModel](kibanaoapi.DeleteAlertingRule),
				Create: createAlertingRule,
				Update: updateAlertingRule,
			},
		),
		KibanaSpaceImporter: entitycore.NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id")),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newResource()
}
