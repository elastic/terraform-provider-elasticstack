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

	"github.com/elastic/go-elasticsearch/v9/typedapi/ccr/follow"
	"github.com/elastic/go-elasticsearch/v9/typedapi/ccr/putautofollowpattern"
	"github.com/elastic/go-elasticsearch/v9/typedapi/ccr/resumefollow"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

func CreateFollowerIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, indexName string, req *follow.Request) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Ccr.Follow(indexName).Request(req).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetFollowerIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, indexName string) (*types.FollowerIndex, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()
	res, err := typedClient.Ccr.FollowInfo(indexName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if len(res.FollowerIndices) > 0 {
		return &res.FollowerIndices[0], nil
	}
	return nil, nil
}

func PauseFollowerIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, indexName string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Ccr.PauseFollow(indexName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func ResumeFollowerIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, indexName string, req *resumefollow.Request) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Ccr.ResumeFollow(indexName).Request(req).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func CloseIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, indexName string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Indices.Close(indexName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func UnfollowIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, indexName string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Ccr.Unfollow(indexName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func OpenIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, indexName string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Indices.Open(indexName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func PutAutoFollowPattern(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, req *putautofollowpattern.Request) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Ccr.PutAutoFollowPattern(name).Request(req).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetAutoFollowPattern(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) (*types.AutoFollowPatternSummary, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()
	res, err := typedClient.Ccr.GetAutoFollowPattern().Name(name).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if len(res.Patterns) > 0 {
		return &res.Patterns[0].Pattern, nil
	}
	return nil, nil
}

func PauseAutoFollowPattern(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Ccr.PauseAutoFollowPattern(name).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func ResumeAutoFollowPattern(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Ccr.ResumeAutoFollowPattern(name).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func DeleteAutoFollowPattern(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Ccr.DeleteAutoFollowPattern(name).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
