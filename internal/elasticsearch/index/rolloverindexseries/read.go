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

package rolloverindexseries

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func readRolloverIndexSeries(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	resourceID string,
	state Model,
) (Model, bool, diag.Diagnostics) {
	series, diags := discoverAlias(ctx, client, resourceID)
	if diags.HasError() {
		return state, false, diags
	}
	if !series.Found {
		return state, false, nil
	}
	if streamDiags := rejectDataStreamTargets(resourceID, series); streamDiags.HasError() {
		return state, false, streamDiags
	}
	diags.Append(writerHealthWarning(resourceID, series)...)

	id, idDiags := client.ID(ctx, resourceID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	return state.withObservedSeries(id.String(), series), true, diags
}

func writerHealthWarning(aliasName string, series seriesView) diag.Diagnostics {
	var diags diag.Diagnostics
	if !series.HasWriter {
		diags.AddWarning(
			"Rollover alias has no effective writer",
			fmt.Sprintf(
				"Alias %q has no effective writer. Writes and automatic rollover may fail. This resource will not repair the alias or its members.",
				aliasName,
			),
		)
		return diags
	}
	if hasGenerationSuffix(series.Writer) {
		return nil
	}
	diags.AddWarning(
		"Rollover writer cannot support automatic rollover",
		fmt.Sprintf(
			"Alias %q has effective writer %q, which does not end in -<digits>. "+
				"Automatic rollover cannot derive a successor name, and writes may fail. "+
				"This resource will not repair the alias or rename the writer.",
			aliasName,
			series.Writer,
		),
	)
	return diags
}
