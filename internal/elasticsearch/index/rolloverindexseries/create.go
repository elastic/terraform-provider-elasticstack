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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func createRolloverIndexSeries(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[Model],
) (entitycore.WriteResult[Model], diag.Diagnostics) {
	plan := req.Plan
	aliasName := plan.RolloverAlias.ValueString()
	series, diags := discoverAlias(ctx, client, aliasName)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}
	if !series.Found {
		diags = bootstrapIndex(ctx, client, plan.Prefix.ValueString(), aliasName)
		if diags.HasError() {
			return entitycore.WriteResult[Model]{Model: plan}, diags
		}
	} else if adoptDiags := rejectUnadoptable(aliasName, series); adoptDiags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, adoptDiags
	}

	id, idDiags := entitycore.CompositeIDForWrite(ctx, client, req)
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}
	plan.ID = id
	return entitycore.WriteResult[Model]{Model: plan}, diags
}

func bootstrapIndex(ctx context.Context, client *clients.ElasticsearchScopedClient, prefix, aliasName string) diag.Diagnostics {
	name := bootstrapIndexName(prefix)
	_, diags := elasticsearch.PutIndex(ctx, client, &models.Index{
		Name: name,
		Aliases: map[string]models.IndexAlias{
			aliasName: {IsWriteIndex: true},
		},
	}, &models.PutIndexParams{})
	if !diags.HasError() {
		return nil
	}
	if bootstrapNameCollision(diags) {
		return collisionDiagnostics(name, aliasName, diags)
	}
	return diags
}

func bootstrapNameCollision(diags diag.Diagnostics) bool {
	for _, d := range diags {
		if d.Severity() != diag.SeverityError {
			continue
		}
		text := d.Summary() + " " + d.Detail()
		if strings.Contains(text, "resource_already_exists_exception") {
			return true
		}
	}
	return false
}

func collisionDiagnostics(bootstrapName, aliasName string, cause diag.Diagnostics) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Bootstrap index name is already in use",
		fmt.Sprintf(
			"Cannot create bootstrap index %q because that name is already an index or alias. "+
				"Alias %q does not exist, and this resource will not attach the alias to the existing object, "+
				"delete or rename it, or choose another generation. Elasticsearch returned: %s",
			bootstrapName,
			aliasName,
			diagnosticText(cause),
		),
	)
	return diags
}

func diagnosticText(diags diag.Diagnostics) string {
	parts := make([]string, 0, len(diags))
	for _, d := range diags {
		if d.Severity() != diag.SeverityError {
			continue
		}
		parts = append(parts, strings.TrimSpace(d.Summary()+" "+d.Detail()))
	}
	return strings.Join(parts, "; ")
}
