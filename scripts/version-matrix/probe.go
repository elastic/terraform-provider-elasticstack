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
	"fmt"
	"net/http"
	"strings"
)

type RegistryProber struct {
	Client            *http.Client
	ElasticRegistry   string
	DockerHubRegistry string
}

const elasticRegistryHost = "docker.elastic.co/"

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
	base, repo, tag := p.splitImage(image)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, strings.TrimRight(base, "/")+"/v2/"+repo+"/manifests/"+tag, nil)
	if err != nil {
		return false, fmt.Errorf("image probe request: %w", err)
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
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
