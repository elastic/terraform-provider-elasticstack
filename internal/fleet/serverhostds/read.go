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

package serverhostds

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func readDataSource(ctx context.Context, kbClient *clients.KibanaScopedClient, _ string, spaceID string, config serverHostModel) (serverHostModel, bool, diag.Diagnostics) {
	fleetClient := kbClient.GetFleetClient()

	switch {
	case typeutils.IsKnown(config.HostID):
		host, diags := fleet.GetFleetServerHost(ctx, fleetClient, config.HostID.ValueString(), spaceID)
		if diags.HasError() || host == nil {
			return config, false, diags
		}
		return applyHost(ctx, config, host)
	case typeutils.IsKnown(config.Name):
		name := config.Name.ValueString()
		return readFromList(ctx, fleetClient, spaceID, config,
			func(host kbapi.ServerHost) bool { return host.Name == name },
			fmt.Sprintf("no Fleet server host found matching name %q", name),
			fmt.Sprintf("multiple Fleet server hosts match name %q", name),
		)
	case config.Default.ValueBool():
		return readFromList(ctx, fleetClient, spaceID, config,
			func(host kbapi.ServerHost) bool { return typeutils.Deref(host.IsDefault) },
			"no default Fleet server host found",
			"multiple Fleet server hosts are marked as default",
		)
	default:
		return config, false, nil
	}
}

func readFromList(
	ctx context.Context,
	fleetClient *fleet.Client,
	spaceID string,
	config serverHostModel,
	matches func(kbapi.ServerHost) bool,
	notFoundDetail string,
	ambiguousDetail string,
) (serverHostModel, bool, diag.Diagnostics) {
	hosts, diags := fleet.ListFleetServerHosts(ctx, fleetClient, spaceID)
	if diags.HasError() {
		return config, false, diags
	}

	var matched []*kbapi.ServerHost
	for i := range hosts {
		if matches(hosts[i]) {
			matched = append(matched, &hosts[i])
		}
	}

	switch len(matched) {
	case 0:
		diags.AddError("Fleet server host not found", notFoundDetail)
		return config, false, diags
	case 1:
		return applyHost(ctx, config, matched[0])
	default:
		diags.AddError("Multiple Fleet server hosts match", fmt.Sprintf("%s: %s", ambiguousDetail, hostIDList(matched)))
		return config, false, diags
	}
}

func applyHost(ctx context.Context, config serverHostModel, host *kbapi.ServerHost) (serverHostModel, bool, diag.Diagnostics) {
	diags := (&config).populateFromAPI(ctx, host)
	if diags.HasError() {
		return config, false, diags
	}
	return config, true, diags
}

func hostIDList(hosts []*kbapi.ServerHost) string {
	parts := make([]string, len(hosts))
	for i, host := range hosts {
		parts[i] = fmt.Sprintf("host_id %q", host.Id)
	}
	return strings.Join(parts, ", ")
}
