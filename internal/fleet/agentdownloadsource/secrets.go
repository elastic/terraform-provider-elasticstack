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

package agentdownloadsource

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/writeonlyhash"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Terraform attribute paths of the write-only secrets. They double as the
// private-state hash keys (see secret_hash.go) and resolver keys.
const (
	sslKeyWOPath       = "ssl.key_wo"
	authPasswordWOPath = "auth.password_wo"
	authAPIKeyWOPath   = "auth.api_key_wo"
)

var writeOnlyPaths = []string{sslKeyWOPath, authPasswordWOPath, authAPIKeyWOPath}

var secretHasher = writeonlyhash.New("elasticstack_fleet_agent_download_source")

// secretValue is one Fleet secret request value: either a literal (Fleet mints
// a new managed secret) or a reference to an existing secret by id.
type secretValue struct {
	literal *string
	id      *string
}

func (v secretValue) MarshalJSON() ([]byte, error) {
	if v.id != nil {
		return json.Marshal(struct {
			ID string `json:"id"`
		}{ID: *v.id})
	}
	return json.Marshal(v.literal)
}

// secretResolver decides how a configured write-only value, identified by its
// attribute path, is sent.
type secretResolver func(attrPath, value string) secretValue

// literalSecrets sends every value as a literal. It is the resolver for Create,
// where no secret exists yet.
func literalSecrets(_, value string) secretValue {
	return secretValue{literal: &value}
}

// secretWriteBranch implements the shared literal-vs-id decision for a
// write-only value. The id of the secret Fleet currently holds is referenced
// only when the value's hash is unchanged since the last successful apply and
// Fleet reported an id; otherwise the literal is sent. A field with a
// configured value is never omitted, since omission clears the Fleet secret.
func secretWriteBranch(value string, storedHash []byte, existingID *string) secretValue {
	if len(storedHash) > 0 && existingID != nil && *existingID != "" && secretHasher.Matches(value, storedHash) {
		return secretValue{id: existingID}
	}
	return secretValue{literal: &value}
}

type secretsPayload struct {
	Auth *authSecretsPayload `json:"auth,omitempty"`
	Ssl  *sslSecretsPayload  `json:"ssl,omitempty"`
}

type authSecretsPayload struct {
	APIKey   *secretValue `json:"api_key,omitempty"`
	Password *secretValue `json:"password,omitempty"`
}

type sslSecretsPayload struct {
	Key *secretValue `json:"key,omitempty"`
}

// secretsPayload builds the request `secrets` object from the model's secret
// fields, or nil when none is configured. Plain ssl.key is always a literal;
// write-only values go through resolve.
func (m model) secretsPayload(ctx context.Context, resolve secretResolver) (*secretsPayload, diag.Diagnostics) {
	var payload secretsPayload

	ssl, diags := objectValueToSSL(ctx, m.Ssl)
	if diags.HasError() {
		return nil, diags
	}
	if ssl != nil {
		switch {
		case ssl.Key != nil:
			payload.Ssl = &sslSecretsPayload{Key: &secretValue{literal: ssl.Key}}
		case ssl.KeyWO != nil:
			v := resolve(sslKeyWOPath, *ssl.KeyWO)
			payload.Ssl = &sslSecretsPayload{Key: &v}
		}
	}

	if m.Auth != nil {
		var auth authSecretsPayload
		if m.Auth.PasswordWO.ValueStringPointer() != nil {
			v := resolve(authPasswordWOPath, m.Auth.PasswordWO.ValueString())
			auth.Password = &v
		}
		if m.Auth.APIKeyWO.ValueStringPointer() != nil {
			v := resolve(authAPIKeyWOPath, m.Auth.APIKeyWO.ValueString())
			auth.APIKey = &v
		}
		if auth.Password != nil || auth.APIKey != nil {
			payload.Auth = &auth
		}
	}

	if payload.Auth == nil && payload.Ssl == nil {
		return nil, diags
	}
	return &payload, diags
}

// setSecrets copies payload into a generated request body's anonymous
// `secrets` field. The generated create and update bodies declare identical
// but distinct union types for each secret, so the payload is transferred
// through its JSON form rather than per-union constructors.
func setSecrets[T any](dst **T, payload *secretsPayload) error {
	if payload == nil {
		return nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return err
	}
	*dst = &out
	return nil
}
