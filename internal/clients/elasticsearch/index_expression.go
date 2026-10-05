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

package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/expandwildcard"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

type IndexTargetKind string

const (
	RegularIndexTarget IndexTargetKind = "index"
	DataStreamTarget   IndexTargetKind = "data_stream"
)

type ResolvedIndexTargets struct {
	Kind  IndexTargetKind
	Names []string
}

func ResolveIndexExpression(ctx context.Context, client *clients.ElasticsearchScopedClient, expression string, excludedAliases ...string) (ResolvedIndexTargets, fwdiags.Diagnostics) {
	response, err := client.GetESClient().Indices.ResolveIndex(expression).
		ExpandWildcards(expandwildcard.All).
		AllowNoIndices(true).
		Do(ctx)
	if err != nil {
		return ResolvedIndexTargets{}, diagutil.FrameworkDiagFromError(err)
	}
	excludedAliasNames := make(map[string]struct{}, len(excludedAliases))
	for _, alias := range excludedAliases {
		excludedAliasNames[alias] = struct{}{}
	}
	for _, alias := range response.Aliases {
		if _, excluded := excludedAliasNames[alias.Name]; excluded {
			continue
		}
		return ResolvedIndexTargets{}, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(
				"Invalid Configuration",
				fmt.Sprintf("Index expression %q resolves to alias %q, which cannot be an alias target", expression, alias.Name),
			),
		}
	}

	resolvedDataStreams := make(map[string]struct{}, len(response.DataStreams))
	for _, dataStream := range response.DataStreams {
		resolvedDataStreams[dataStream.Name] = struct{}{}
	}

	targets := ResolvedIndexTargets{Kind: RegularIndexTarget}
	for _, index := range response.Indices {
		if index.DataStream != nil {
			if _, resolved := resolvedDataStreams[*index.DataStream]; resolved {
				continue
			}
		}
		if strings.Contains(index.Name, ":") {
			return ResolvedIndexTargets{}, remoteTargetDiagnostic(expression, index.Name)
		}
		targets.Names = append(targets.Names, index.Name)
	}
	if len(targets.Names) > 0 && len(response.DataStreams) > 0 {
		return ResolvedIndexTargets{}, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(
				"Invalid Configuration",
				fmt.Sprintf("Index expression %q resolves to both regular indices and data streams", expression),
			),
		}
	}
	if len(response.DataStreams) > 0 {
		targets.Kind = DataStreamTarget
		for _, dataStream := range response.DataStreams {
			if strings.Contains(dataStream.Name, ":") {
				return ResolvedIndexTargets{}, remoteTargetDiagnostic(expression, dataStream.Name)
			}
			targets.Names = append(targets.Names, dataStream.Name)
		}
	}
	return targets, nil
}

func remoteTargetDiagnostic(expression, target string) fwdiags.Diagnostics {
	return fwdiags.Diagnostics{
		fwdiags.NewErrorDiagnostic(
			"Invalid Configuration",
			fmt.Sprintf("Index expression %q resolves to remote target %q, which cannot be an alias target", expression, target),
		),
	}
}
