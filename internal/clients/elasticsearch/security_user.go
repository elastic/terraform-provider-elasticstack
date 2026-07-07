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

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func PutUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, user *types.User, password, passwordHash *string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	req := typedClient.Security.PutUser(user.Username).
		Enabled(user.Enabled).
		Roles(user.Roles...)

	if user.Email != nil {
		req.Email(*user.Email)
	}
	if user.FullName != nil {
		req.FullName(*user.FullName)
	}
	if user.Metadata != nil {
		req.Metadata(user.Metadata)
	}

	if password != nil {
		req.Password(*password)
	}
	if passwordHash != nil {
		req.PasswordHash(*passwordHash)
	}

	_, err := req.Do(ctx)
	if err != nil {
		diags.AddError("Unable to create or update a user", err.Error())
		return diags
	}

	return diags
}

func GetUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string) (*types.User, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, err := typedClient.Security.GetUser().Username(username).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	if user, ok := res[username]; ok {
		return &user, nil
	}

	return nil, fwdiag.Diagnostics{
		fwdiag.NewErrorDiagnostic(
			"Unable to find a user in the cluster",
			fmt.Sprintf(`Unable to find "%s" user in the cluster`, username),
		),
	}
}

func DeleteUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	_, err := typedClient.Security.DeleteUser(username).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Unable to delete a user", err.Error())
		return diags
	}

	return diags
}

func EnableUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	_, err := typedClient.Security.EnableUser(username).Do(ctx)
	if err != nil {
		diags.AddError(
			"Unable to enable system user",
			err.Error(),
		)
		return diags
	}

	return diags
}

func DisableUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	_, err := typedClient.Security.DisableUser(username).Do(ctx)
	if err != nil {
		diags.AddError(
			"Unable to disable system user",
			err.Error(),
		)
		return diags
	}

	return diags
}

func ChangeUserPassword(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string, password, passwordHash *string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	req := typedClient.Security.ChangePassword().Username(username)
	if password != nil {
		req.Password(*password)
	}
	if passwordHash != nil {
		req.PasswordHash(*passwordHash)
	}

	_, err := req.Do(ctx)
	if err != nil {
		diags.AddError(
			"Unable to change user's password",
			err.Error(),
		)
		return diags
	}

	return diags
}
