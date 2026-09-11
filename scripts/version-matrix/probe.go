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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RegistryProber struct {
	Client            *http.Client
	ElasticRegistry   string
	DockerHubRegistry string
}

const (
	elasticRegistryHost = "docker.elastic.co/"
	manifestAccept      = "application/vnd.docker.distribution.manifest.v2+json, " +
		"application/vnd.oci.image.manifest.v1+json, " +
		"application/vnd.docker.distribution.manifest.list.v2+json, " +
		"application/vnd.oci.image.index.v1+json"
)

func ProbeComposeStack(ctx context.Context, prober RegistryProber, version string) (bool, error) {
	for _, image := range ComposeStackImages(version) {
		ok, err := prober.manifestExists(ctx, image)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (p RegistryProber) manifestExists(ctx context.Context, image string) (bool, error) {
	resp, err := p.headManifest(ctx, image, "")
	if err != nil {
		return false, fmt.Errorf("image probe %s: %w", image, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, terr := p.fetchAnonymousToken(ctx, resp.Header.Get("WWW-Authenticate"))
		if terr != nil {
			return false, fmt.Errorf("image probe %s: %w", image, terr)
		}
		_ = resp.Body.Close()
		resp, err = p.headManifest(ctx, image, token)
		if err != nil {
			return false, fmt.Errorf("image probe %s: %w", image, err)
		}
		defer resp.Body.Close()
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("image probe %s: unexpected status %s", image, resp.Status)
	}
}

func (p RegistryProber) headManifest(ctx context.Context, image, token string) (*http.Response, error) {
	base, repo, tag := p.splitImage(image)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, strings.TrimRight(base, "/")+"/v2/"+repo+"/manifests/"+tag, nil)
	if err != nil {
		return nil, fmt.Errorf("image probe request: %w", err)
	}
	req.Header.Set("Accept", manifestAccept)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return p.httpClient().Do(req)
}

func (p RegistryProber) fetchAnonymousToken(ctx context.Context, challenge string) (string, error) {
	realm, service, scope, ok := parseBearerChallenge(challenge)
	if !ok {
		return "", fmt.Errorf("unrecognized registry auth challenge")
	}
	u, err := url.Parse(realm)
	if err != nil {
		return "", fmt.Errorf("registry auth realm: %w", err)
	}
	q := u.Query()
	if service != "" {
		q.Set("service", service)
	}
	if scope != "" {
		q.Set("scope", scope)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("registry token request: %w", err)
	}
	resp, err := p.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("registry token fetch: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("registry token read: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry token endpoint returned %s", resp.Status)
	}
	var payload struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("registry token parse: %w", err)
	}
	if payload.Token != "" {
		return payload.Token, nil
	}
	if payload.AccessToken != "" {
		return payload.AccessToken, nil
	}
	return "", fmt.Errorf("registry token response missing token")
}

func parseBearerChallenge(header string) (realm, service, scope string, ok bool) {
	header = strings.TrimSpace(header)
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return "", "", "", false
	}
	params := header[len("Bearer "):]
	for part := range strings.SplitSeq(params, ",") {
		key, value, found := strings.Cut(strings.TrimSpace(part), "=")
		if !found {
			continue
		}
		value = strings.Trim(value, `"`)
		switch strings.ToLower(key) {
		case "realm":
			realm = value
		case "service":
			service = value
		case "scope":
			scope = value
		}
	}
	return realm, service, scope, realm != ""
}

func (p RegistryProber) httpClient() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}

func (p RegistryProber) splitImage(image string) (base, repo, tag string) {
	name, tag, _ := strings.Cut(image, ":")
	if strings.HasPrefix(name, elasticRegistryHost) {
		return p.ElasticRegistry, strings.TrimPrefix(name, elasticRegistryHost), tag
	}
	return p.DockerHubRegistry, name, tag
}

func ComposeStackImages(version string) []string {
	return []string{
		"docker.elastic.co/elasticsearch/elasticsearch:" + version,
		"docker.elastic.co/kibana/kibana:" + version,
		AgentImage(version),
	}
}

func AgentImage(version string) string {
	major, minor, _ := parseLooseVersion(version)
	if major == 8 && minor <= 1 {
		return "elastic/elastic-agent:" + version
	}
	return "docker.elastic.co/elastic-agent/elastic-agent:" + version
}
