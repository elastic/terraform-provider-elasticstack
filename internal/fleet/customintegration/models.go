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

package customintegration

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type customIntegrationModel struct {
	ID                        types.String `tfsdk:"id"`
	KibanaConnection          types.List   `tfsdk:"kibana_connection"`
	PackagePath               types.String `tfsdk:"package_path"`
	PackageName               types.String `tfsdk:"package_name"`
	PackageVersion            types.String `tfsdk:"package_version"`
	Checksum                  types.String `tfsdk:"checksum"`
	IgnoreMappingUpdateErrors types.Bool   `tfsdk:"ignore_mapping_update_errors"`
	SkipDataStreamRollover    types.Bool   `tfsdk:"skip_data_stream_rollover"`
	SkipDestroy               types.Bool   `tfsdk:"skip_destroy"`
	SpaceID                   types.String `tfsdk:"space_id"`
}

// getPackageID returns the deterministic Terraform ID for an uploaded package.
// Matches the pattern used by elasticstack_fleet_integration for consistency
// across the two resources — both key on (name, version) of the installed
// package in Fleet.
func getPackageID(name, version string) string {
	hash, _ := schemautil.StringToHash(name + version)
	if hash == nil {
		return ""
	}
	return *hash
}

// detectContentType maps a file path extension to the Content-Type header
// expected by POST /api/fleet/epm/packages. The Fleet API accepts
// application/zip and application/gzip; we default to zip for unknown
// extensions (elastic-package produces .zip files by default).
func detectContentType(packagePath string) string {
	lower := strings.ToLower(packagePath)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"), strings.HasSuffix(lower, ".gz"):
		return "application/gzip"
	default:
		return "application/zip"
	}
}

// sha256File returns the SHA-256 hex digest of the file at path.
// Errors are returned with the wrapped file path for diagnostic clarity.
func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("read %q: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
