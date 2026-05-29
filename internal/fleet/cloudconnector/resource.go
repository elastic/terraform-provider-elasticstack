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

package cloudconnector

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                     = newResource()
	_ resource.ResourceWithConfigure        = newResource()
	_ resource.ResourceWithConfigValidators = newResource()
)

// Resource is the Fleet cloud connector resource.
// CRUD implementations are placeholders until Task 6 replaces readCloudConnector,
// deleteCloudConnector, and the write callbacks wired below.
type Resource struct {
	*entitycore.KibanaResource[cloudConnectorModel]
}

func newResource() *Resource {
	placeholder := entitycore.PlaceholderKibanaWriteCallback[cloudConnectorModel]()
	return &Resource{
		KibanaResource: entitycore.NewKibanaResource[cloudConnectorModel](
			entitycore.ComponentFleet,
			"cloud_connector",
			entitycore.KibanaResourceOptions[cloudConnectorModel]{
				Schema: getSchema,
				Read:   readCloudConnector,
				Delete: deleteCloudConnector,
				Create: placeholder,
				Update: placeholder,
			},
		),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newResource()
}

func (r *Resource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot(attrAWSBlock),
			path.MatchRoot(attrAzureBlock),
			path.MatchRoot(attrVarsMap),
		),
		providerBlockMatchesCloudProvider{},
	}
}

func readCloudConnector(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model cloudConnectorModel) (cloudConnectorModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	diags.AddError(
		"Fleet cloud connector read not implemented",
		"Read is implemented in Task 6 of the fleet-cloud-connector change.",
	)
	return model, false, diags
}

func deleteCloudConnector(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ cloudConnectorModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Fleet cloud connector delete not implemented",
		"Delete is implemented in Task 6 of the fleet-cloud-connector change.",
	)
	return diags
}
