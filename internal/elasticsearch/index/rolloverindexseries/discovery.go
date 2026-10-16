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
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/expandwildcard"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type targetKind int

const (
	targetKindIndex targetKind = iota
	targetKindDataStream
)

type aliasMember struct {
	Name         string
	IsWriteIndex *bool
	Kind         targetKind
}

type seriesView struct {
	Found     bool
	Members   []aliasMember
	Writer    string
	HasWriter bool
}

var generationSuffix = regexp.MustCompile(`-[0-9]+$`)

func hasGenerationSuffix(name string) bool {
	return generationSuffix.MatchString(name)
}

func effectiveWriter(members []aliasMember) (string, bool) {
	var explicit []string
	for _, member := range members {
		if member.IsWriteIndex != nil && *member.IsWriteIndex {
			explicit = append(explicit, member.Name)
		}
	}
	if len(explicit) == 1 {
		return explicit[0], true
	}
	if len(explicit) > 0 {
		return "", false
	}
	if len(members) == 1 && members[0].IsWriteIndex == nil {
		return members[0].Name, true
	}
	return "", false
}

// discoverAlias loads one concrete alias. GetAlias uses the same object shape
// for an index and a data stream, so each target's kind comes from resolve-index.
func discoverAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, aliasName string) (seriesView, diag.Diagnostics) {
	return discoverUntilTargetsResolve(ctx, client, aliasName, func() (map[string]types.IndexAliases, diag.Diagnostics) {
		return elasticsearch.GetAlias(ctx, client, aliasName)
	})
}

// discoverAliasForDelete requests open, closed, and hidden indices. The alias
// argument stays a concrete name; expand_wildcards selects index state only.
func discoverAliasForDelete(ctx context.Context, client *clients.ElasticsearchScopedClient, aliasName string) (seriesView, diag.Diagnostics) {
	return discoverUntilTargetsResolve(ctx, client, aliasName, func() (map[string]types.IndexAliases, diag.Diagnostics) {
		return elasticsearch.CallOrNotFound(func() (map[string]types.IndexAliases, error) {
			return client.GetESClient().Indices.GetAlias().
				Name(aliasName).
				ExpandWildcards(expandwildcard.Open, expandwildcard.Closed, expandwildcard.Hidden).
				Do(ctx)
		}, "Unable to get index alias")
	})
}

func discoverUntilTargetsResolve(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	aliasName string,
	fetch func() (map[string]types.IndexAliases, diag.Diagnostics),
) (seriesView, diag.Diagnostics) {
	var previousDisappeared []string
	for {
		if ctx.Err() != nil {
			return seriesView{}, aliasResolutionTimeout(aliasName)
		}
		indices, diags := fetch()
		if diags.HasError() {
			return seriesView{}, diags
		}
		view, disappeared, interpretDiags := interpretAlias(ctx, client, aliasName, indices)
		if interpretDiags.HasError() {
			return seriesView{}, interpretDiags
		}
		if len(disappeared) == 0 {
			return view, nil
		}
		slices.Sort(disappeared)
		if slices.Equal(previousDisappeared, disappeared) {
			return seriesView{}, unresolvedAliasTargets(aliasName, disappeared)
		}
		previousDisappeared = disappeared
		if !waitForRediscovery(ctx) {
			return seriesView{}, aliasResolutionTimeout(aliasName)
		}
	}
}

func aliasResolutionTimeout(aliasName string) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Timed out resolving alias membership",
		fmt.Sprintf("Resolving alias %q did not finish before the operation timeout. Terraform kept this resource in state.", aliasName),
	)
	return diags
}

func unresolvedAliasTargets(aliasName string, names []string) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Unable to resolve alias target type",
		fmt.Sprintf(
			"Alias %q member %s could not be resolved as an index or a data stream. This resource will not treat it as an ordinary index.",
			aliasName,
			strings.Join(names, ", "),
		),
	)
	return diags
}

func interpretAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, aliasName string, indices map[string]types.IndexAliases) (seriesView, []string, diag.Diagnostics) {
	if len(indices) == 0 {
		return seriesView{}, nil, nil
	}

	members := make([]aliasMember, 0, len(indices))
	for targetName, indexAliases := range indices {
		definition, ok := indexAliases.Aliases[aliasName]
		if !ok {
			continue
		}
		members = append(members, aliasMember{
			Name:         targetName,
			IsWriteIndex: definition.IsWriteIndex,
		})
	}
	if len(members) == 0 {
		return seriesView{}, nil, nil
	}

	var diags diag.Diagnostics
	var disappeared []string
	sawDataStream := false
	for i := range members {
		kind, missing, kindDiags := resolveTargetKind(ctx, client, members[i].Name, aliasName)
		diags.Append(kindDiags...)
		if diags.HasError() {
			return seriesView{}, nil, diags
		}
		if missing {
			disappeared = append(disappeared, members[i].Name)
			continue
		}
		members[i].Kind = kind
		if kind == targetKindDataStream {
			sawDataStream = true
		}
	}
	if len(disappeared) > 0 {
		if !sawDataStream {
			return seriesView{}, disappeared, nil
		}
		members = omitNamedMembers(members, disappeared)
	}

	writer, hasWriter := effectiveWriter(members)
	return seriesView{
		Found:     true,
		Members:   members,
		Writer:    writer,
		HasWriter: hasWriter,
	}, nil, nil
}

func omitNamedMembers(members []aliasMember, names []string) []aliasMember {
	skipped := make(map[string]struct{}, len(names))
	for _, name := range names {
		skipped[name] = struct{}{}
	}
	remaining := make([]aliasMember, 0, len(members))
	for _, member := range members {
		if _, ok := skipped[member.Name]; !ok {
			remaining = append(remaining, member)
		}
	}
	return remaining
}

func resolveTargetKind(ctx context.Context, client *clients.ElasticsearchScopedClient, targetName, aliasName string) (targetKind, bool, diag.Diagnostics) {
	resolved, diags := elasticsearch.ResolveIndexExpression(ctx, client, targetName, aliasName)
	if diags.HasError() {
		return targetKindIndex, false, diags
	}
	if resolved.Kind == elasticsearch.DataStreamTarget {
		return targetKindDataStream, false, nil
	}
	if len(resolved.Names) == 0 {
		return targetKindIndex, true, nil
	}
	return targetKindIndex, false, nil
}

func rejectDataStreamTargets(aliasName string, series seriesView) diag.Diagnostics {
	names := dataStreamNames(series.Members)
	if len(names) == 0 {
		return nil
	}
	var diags diag.Diagnostics
	diags.AddError(
		"Rollover alias targets a data stream",
		fmt.Sprintf(
			"Alias %q resolves to data stream %s. This resource manages ordinary index aliases only and will not adopt, modify, or delete data stream targets.",
			aliasName,
			strings.Join(names, ", "),
		),
	)
	return diags
}

func rejectUnadoptable(aliasName string, series seriesView) diag.Diagnostics {
	if diags := rejectDataStreamTargets(aliasName, series); diags.HasError() {
		return diags
	}
	if !series.HasWriter {
		var diags diag.Diagnostics
		diags.AddError(
			"Rollover alias has no effective writer",
			fmt.Sprintf(
				"Alias %q has no effective writer among members [%s]. "+
					"A member must be marked is_write_index=true, or be the sole member with that flag unset. "+
					"This resource will not choose a writer or modify the alias.",
				aliasName,
				strings.Join(memberNames(series.Members), ", "),
			),
		)
		return diags
	}
	if !hasGenerationSuffix(series.Writer) {
		var diags diag.Diagnostics
		diags.AddError(
			"Rollover writer cannot support automatic rollover",
			fmt.Sprintf(
				"Alias %q has effective writer %q, which does not end in -<digits>. Automatic rollover cannot derive a successor name from that writer, and this resource will not rename it or modify the alias.",
				aliasName,
				series.Writer,
			),
		)
		return diags
	}
	return nil
}

func memberNames(members []aliasMember) []string {
	names := make([]string, len(members))
	for i, member := range members {
		names[i] = member.Name
	}
	slices.Sort(names)
	return names
}

func dataStreamNames(members []aliasMember) []string {
	names := make([]string, 0)
	for _, member := range members {
		if member.Kind == targetKindDataStream {
			names = append(names, member.Name)
		}
	}
	slices.Sort(names)
	return names
}

// rediscoveryDelay pauses between membership passes for delete and alias
// resolution. The wait stops when the caller's context ends, so it cannot run
// past that operation's deadline.
var rediscoveryDelay = 50 * time.Millisecond

func waitForRediscovery(ctx context.Context) bool {
	timer := time.NewTimer(rediscoveryDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
