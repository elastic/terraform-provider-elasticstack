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

package fleet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetPackage reads a specific package from the API.
func GetPackage(ctx context.Context, client *Client, name, version, spaceID string) (*kbapi.KibanaHTTPAPIsGetPackageInfo, diag.Diagnostics) {
	params := kbapi.GetFleetEpmPackagesPkgnamePkgversionParams{}

	resp, err := client.API.GetFleetEpmPackagesPkgnamePkgversionWithResponse(ctx, name, version, &params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// InstallPackageOptions holds the options for installing a package.
type InstallPackageOptions struct {
	SpaceID                   string
	Force                     bool
	Prerelease                bool
	IgnoreMappingUpdateErrors *bool
	SkipDataStreamRollover    *bool
	IgnoreConstraints         bool
}

// InstallPackage installs a package.
func InstallPackage(ctx context.Context, client *Client, name, version string, opts InstallPackageOptions) diag.Diagnostics {
	params := kbapi.PostFleetEpmPackagesPkgnamePkgversionParams{
		Prerelease:                &opts.Prerelease,
		IgnoreMappingUpdateErrors: opts.IgnoreMappingUpdateErrors,
		SkipDataStreamRollover:    opts.SkipDataStreamRollover,
	}
	body := kbapi.PostFleetEpmPackagesPkgnamePkgversionJSONRequestBody{
		Force:             &opts.Force,
		IgnoreConstraints: &opts.IgnoreConstraints,
	}

	resp, err := client.API.PostFleetEpmPackagesPkgnamePkgversionWithResponse(ctx, name, version, &params, body, kibanautil.SpaceAwarePathRequestEditor(opts.SpaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// Uninstall uninstalls a package.
func Uninstall(ctx context.Context, client *Client, name, version string, spaceID string, force bool) diag.Diagnostics {
	params := kbapi.DeleteFleetEpmPackagesPkgnamePkgversionParams{
		Force: &force,
	}
	resp, err := client.API.DeleteFleetEpmPackagesPkgnamePkgversionWithResponse(ctx, name, version, &params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		msg := resp.JSON400.Message
		if msg == fmt.Sprintf("%s is not installed", name) {
			return nil
		}
		return diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	case http.StatusNotFound:
		return nil
	default:
		return diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// InstallKibanaAssets installs Kibana assets for an already-installed package into a specific space.
func InstallKibanaAssets(ctx context.Context, client *Client, name, version string, spaceID string, force bool) diag.Diagnostics {
	spaceIDs := []string{spaceID}
	body := kbapi.PostFleetEpmPackagesPkgnamePkgversionKibanaAssetsJSONRequestBody{
		Force:    &force,
		SpaceIds: &spaceIDs,
	}

	resp, err := client.API.PostFleetEpmPackagesPkgnamePkgversionKibanaAssetsWithResponse(ctx, name, version, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// installSpaceDeleteRejectedMsg is the stable substring of the Fleet 9.5+
// error returned when DELETE .../kibana_assets targets the package's install
// space while the package is still installed in other spaces. It must be kept
// in sync with the upstream Fleet wording.
const installSpaceDeleteRejectedMsg = "space where the package was installed"

func normalizeDiagnosticText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// ContainsInstallSpaceDeleteRejection reports whether diags contains an error
// diagnostic whose Detail or Summary contains the Fleet install-space
// rejection message. The comparison normalises whitespace so embedded newlines
// in JSON-serialised response bodies do not prevent matching.
func ContainsInstallSpaceDeleteRejection(diags diag.Diagnostics) bool {
	for _, d := range diags {
		if d.Severity() != diag.SeverityError {
			continue
		}
		if strings.Contains(normalizeDiagnosticText(d.Detail()), installSpaceDeleteRejectedMsg) {
			return true
		}
		if strings.Contains(normalizeDiagnosticText(d.Summary()), installSpaceDeleteRejectedMsg) {
			return true
		}
	}
	return false
}

// forceQueryParamEditor returns a RequestEditorFn that appends ?force=<bool> to
// the request URL. Used for generated endpoints that do not expose a typed force
// parameter (e.g. the kibana_assets DELETE endpoint).
func forceQueryParamEditor(force bool) kbapi.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		q := req.URL.Query()
		q.Set("force", strconv.FormatBool(force))
		req.URL.RawQuery = q.Encode()
		return nil
	}
}

// DeleteKibanaAssets removes Kibana assets for a package from a specific space.
func DeleteKibanaAssets(ctx context.Context, client *Client, name, version string, spaceID string, force bool) diag.Diagnostics {
	resp, err := client.API.DeleteFleetEpmPackagesPkgnamePkgversionKibanaAssetsWithResponse(ctx, name, version, kibanautil.SpaceAwarePathRequestEditor(spaceID), forceQueryParamEditor(force))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleDeleteResponse(resp.StatusCode(), resp.Body)
}

// GetPackages returns information about the latest packages known to Fleet.
// If spaceID is non-empty and not "default", the request will be scoped to that Kibana space.
func unpackGetPackagesItems(resp *kbapi.KibanaHTTPAPIsGetPackagesResponse, contentType string) ([]kbapi.KibanaHTTPAPIsPackageListItem, diag.Diagnostics) {
	if resp == nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unexpected Fleet response",
				fmt.Sprintf(
					"Fleet returned HTTP 200 for the packages list endpoint but the response body could not be decoded as JSON. "+
						"Unexpected Content-Type: %q. Verify the Kibana Fleet endpoint is reachable and returns a JSON Content-Type.",
					contentType,
				),
			),
		}
	}
	return resp.Items, nil
}

func GetPackages(ctx context.Context, client *Client, prerelease bool, spaceID string) ([]kbapi.KibanaHTTPAPIsPackageListItem, diag.Diagnostics) {
	params := kbapi.GetFleetEpmPackagesParams{
		Prerelease: &prerelease,
	}

	resp, err := client.API.GetFleetEpmPackagesWithResponse(ctx, &params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		items, diags := unpackGetPackagesItems(resp.JSON200, resp.ContentType())
		if diags.HasError() {
			return nil, diags
		}
		return items, nil
	case http.StatusBadRequest:
		// Older Kibana versions (pre-8.7) do not recognise the prerelease query
		// parameter and return 400 with "definition for this key is missing".
		// Retry without the parameter so we remain compatible.
		if strings.Contains(string(resp.Body), "prerelease") {
			retryParams := kbapi.GetFleetEpmPackagesParams{}
			retryResp, retryErr := client.API.GetFleetEpmPackagesWithResponse(ctx, &retryParams, kibanautil.SpaceAwarePathRequestEditor(spaceID))
			if retryErr != nil {
				return nil, diagutil.FrameworkDiagFromError(retryErr)
			}
			if retryResp.StatusCode() == http.StatusOK {
				items, diags := unpackGetPackagesItems(retryResp.JSON200, retryResp.ContentType())
				if diags.HasError() {
					return nil, diags
				}
				return items, nil
			}
			return nil, diagutil.ReportUnknownHTTPError(retryResp.StatusCode(), retryResp.Body)
		}
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// UploadPackageOptions holds the options for uploading a custom integration package.
type UploadPackageOptions struct {
	// PackagePath is the path to the package archive to upload (.zip or .tar.gz/.tgz).
	PackagePath string
	// ContentType is the MIME type of the package file (e.g. "application/zip").
	ContentType string
	// IgnoreMappingUpdateErrors suppresses mapping update errors during install.
	IgnoreMappingUpdateErrors bool
	// SkipDataStreamRollover skips data stream rollover during install.
	SkipDataStreamRollover bool
	// SpaceID scopes the request to a specific Kibana space.
	SpaceID string
}

// UploadPackageResult holds the result of uploading a custom integration package.
type UploadPackageResult struct {
	// PackageName is the name of the uploaded package as returned by Fleet.
	PackageName string
	// PackageVersion is the installed version resolved from the package list.
	PackageVersion string
}

// UploadPackage uploads a custom integration package to Fleet and returns the
// resolved package name and installed version. It opens the file at
// opts.PackagePath, posts it to the Fleet EPM packages endpoint, extracts the
// package name from the response, and then queries the package list to resolve
// the installed version.
func UploadPackage(ctx context.Context, client *Client, opts UploadPackageOptions) (*UploadPackageResult, diag.Diagnostics) {
	f, err := os.Open(opts.PackagePath)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("opening package file %q: %w", opts.PackagePath, err))
	}
	defer f.Close()

	params := kbapi.PostFleetEpmPackagesParams{
		IgnoreMappingUpdateErrors: &opts.IgnoreMappingUpdateErrors,
		SkipDataStreamRollover:    &opts.SkipDataStreamRollover,
	}

	resp, err := client.API.PostFleetEpmPackagesWithBodyWithResponse(ctx, &params, opts.ContentType, io.NopCloser(f), kibanautil.SpaceAwarePathRequestEditor(opts.SpaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	// Kibana rate-limits EPM uploads with HTTP 429 ("Please wait Xs before uploading
	// again."). Retry once after the requested delay so sequential tests that upload
	// multiple packages do not fail due to back-to-back upload attempts.
	if resp.StatusCode() == http.StatusTooManyRequests {
		wait := 15 * time.Second
		if m := regexp.MustCompile(`\b(\d+)s\b`).FindSubmatch(resp.Body); m != nil {
			if secs, parseErr := strconv.Atoi(string(m[1])); parseErr == nil && secs > 0 {
				wait = time.Duration(secs+2) * time.Second
			}
		}
		select {
		case <-ctx.Done():
			return nil, diagutil.FrameworkDiagFromError(ctx.Err())
		case <-time.After(wait):
		}
		if _, seekErr := f.Seek(0, io.SeekStart); seekErr != nil {
			return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("rewinding package file for retry after rate limit: %w", seekErr))
		}
		resp, err = client.API.PostFleetEpmPackagesWithBodyWithResponse(ctx, &params, opts.ContentType, io.NopCloser(f), kibanautil.SpaceAwarePathRequestEditor(opts.SpaceID))
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusCreated:
		// intentional fall-through
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}

	// The response body does not have a typed JSON200 field; unmarshal manually.
	// The field that carries the package name and version changed across Kibana versions:
	//   - newer Kibana (8.8+): _meta.name / _meta.version
	//   - older Kibana (8.0–8.7): items[0].name / items[0].version
	// Try both paths; if neither yields a name, fall back to parsing the
	// zip manifest directly (version-independent but zip-only).
	var uploadResp struct {
		Meta struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"_meta"`
		Items []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"items"`
	}
	// Best-effort unmarshal; an error here is non-fatal — we fall through to
	// the zip-manifest fallback below.
	_ = json.Unmarshal(resp.Body, &uploadResp)

	packageName := uploadResp.Meta.Name
	packageVersion := uploadResp.Meta.Version
	if packageName == "" && len(uploadResp.Items) > 0 {
		packageName = uploadResp.Items[0].Name
		packageVersion = uploadResp.Items[0].Version
	}
	if packageName == "" {
		// Last resort: parse the name (and version) from the package archive. This is
		// reliable across all Kibana versions and supports both zip and tar.gz archives.
		var archErr error
		packageName, packageVersion, archErr = parsePackageInfo(opts.PackagePath)
		if archErr != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid upload response",
					"Fleet did not return a package name and the archive manifest could not be parsed: "+archErr.Error(),
				),
			}
		}
	} else if packageVersion == "" {
		// Have name from response but no version; fill in from zip manifest.
		_, packageVersion, _ = parsePackageInfo(opts.PackagePath)
	}

	// Resolve the installed version by querying the package list and filtering by
	// name and status. This is the post-upload verification source for the
	// package version that we persist in state.
	//
	// When multiple versions of the same package are listed, pick the highest
	// semver among entries with status "installed" so that state always tracks the
	// most recent confirmed installation rather than a registry-only entry.
	packages, diags := GetPackages(ctx, client, true, opts.SpaceID)
	if diags.HasError() {
		return nil, diags
	}

	var highestSemver *semver.Version
	var resolvedVersion string
	for _, pkg := range packages {
		if pkg.Name != packageName {
			continue
		}
		if pkg.Status == nil || *pkg.Status != "installed" {
			continue
		}
		v, parseErr := semver.NewVersion(pkg.Version)
		if parseErr != nil {
			// Non-semver version string: use it only if no valid candidate yet.
			if resolvedVersion == "" {
				resolvedVersion = pkg.Version
			}
			continue
		}
		if highestSemver == nil || v.GreaterThan(highestSemver) {
			highestSemver = v
			resolvedVersion = pkg.Version
		}
	}
	if resolvedVersion != "" {
		if diags := waitForPackageInstalled(ctx, client, packageName, resolvedVersion, opts.SpaceID); diags.HasError() {
			return nil, diags
		}
		return &UploadPackageResult{
			PackageName:    packageName,
			PackageVersion: resolvedVersion,
		}, nil
	}

	if packageVersion != "" {
		pkg, pkgDiags := GetPackage(ctx, client, packageName, packageVersion, opts.SpaceID)
		if !pkgDiags.HasError() && pkg != nil {
			if diags := waitForPackageInstalled(ctx, client, packageName, packageVersion, opts.SpaceID); diags.HasError() {
				return nil, diags
			}
			return &UploadPackageResult{
				PackageName:    packageName,
				PackageVersion: packageVersion,
			}, nil
		}
	}

	detail := fmt.Sprintf(
		"Fleet accepted the upload for package %q, but neither the packages list nor the package info API returned a matching installed package.",
		packageName,
	)
	if packageVersion != "" {
		detail = fmt.Sprintf(
			"Fleet accepted the upload for package %q and the upload/archive metadata resolved version %q, but neither the packages list nor the package info API returned a matching installed package.",
			packageName,
			packageVersion,
		)
	}
	detail += " The provider requires a matching installed package to verify the upload result."

	return nil, diag.Diagnostics{
		diag.NewErrorDiagnostic(
			"Package not found after upload",
			detail,
		),
	}
}

// isInstalledStatus reports whether the given installation-info/status pair indicates a package
// is fully installed. It checks the newer InstallationInfo.InstallStatus enum first, falling back
// to the legacy Status string field when the enum is absent or inconclusive. Both
// [kbapi.KibanaHTTPAPIsGetPackageInfo] and [kbapi.KibanaHTTPAPIsPackageListItem] expose these same
// two fields, so this helper serves both the package-info and packages-list response shapes.
func isInstalledStatus(installationInfo *kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo, status *string) bool {
	if installationInfo != nil {
		switch installationInfo.InstallStatus {
		case kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstalled:
			return true
		case kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstallFailed:
			return false
		}
	}
	return status != nil && strings.EqualFold(*status, "installed")
}

// isInstallFailedStatus is the failure-side counterpart to [isInstalledStatus].
func isInstallFailedStatus(installationInfo *kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo, status *string) bool {
	if installationInfo != nil {
		switch installationInfo.InstallStatus {
		case kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstallFailed:
			return true
		case kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstalled:
			return false
		}
	}
	return status != nil && strings.EqualFold(*status, "install_failed")
}

// IsPackageInstalled reports whether pkg's installation status indicates it is fully installed.
func IsPackageInstalled(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo) bool {
	if pkg == nil {
		return false
	}
	return isInstalledStatus(pkg.InstallationInfo, pkg.Status)
}

// IsPackageInstallFailed reports whether pkg's installation status indicates installation failed.
func IsPackageInstallFailed(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo) bool {
	if pkg == nil {
		return false
	}
	return isInstallFailedStatus(pkg.InstallationInfo, pkg.Status)
}

// WaitForPackageInstalled polls Fleet until packageName/packageVersion reports an installed status
// in spaceID, or timeout elapses. When fallbackToListScan is true and the package-info endpoint has
// not yet reported a conclusive status, the packages list is also scanned for a matching
// name/version entry; this is needed by upload flows where the packages list can reflect the
// resolved status before the package-info endpoint does.
func WaitForPackageInstalled(ctx context.Context, client *Client, packageName, packageVersion, spaceID string, timeout time.Duration, fallbackToListScan bool) error {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return asyncutils.WaitForStateTransition(waitCtx, "fleet package", fmt.Sprintf("%s/%s", packageName, packageVersion), func(ctx context.Context) (bool, error) {
		pkg, diags := GetPackage(ctx, client, packageName, packageVersion, spaceID)
		if diags.HasError() {
			return false, fmt.Errorf("failed to read package installation status: %s", diags[0].Summary())
		}
		if IsPackageInstalled(pkg) {
			return true, nil
		}
		if IsPackageInstallFailed(pkg) {
			return false, fmt.Errorf("package %s/%s installation failed", packageName, packageVersion)
		}
		if !fallbackToListScan {
			return false, nil
		}

		packages, diags := GetPackages(ctx, client, true, spaceID)
		if diags.HasError() {
			return false, fmt.Errorf("failed to list packages during verification: %s", diags[0].Summary())
		}
		for _, candidate := range packages {
			if candidate.Name != packageName || candidate.Version != packageVersion {
				continue
			}
			if isInstalledStatus(candidate.InstallationInfo, candidate.Status) {
				return true, nil
			}
			if isInstallFailedStatus(candidate.InstallationInfo, candidate.Status) {
				return false, fmt.Errorf("package %s/%s installation failed", packageName, packageVersion)
			}
		}
		return false, nil
	})
}

func waitForPackageInstalled(ctx context.Context, client *Client, packageName, packageVersion, spaceID string) diag.Diagnostics {
	if waitErr := WaitForPackageInstalled(ctx, client, packageName, packageVersion, spaceID, 30*time.Second, false); waitErr != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Package not ready after upload",
				fmt.Sprintf("Package %s/%s did not reach an installed state after upload: %s", packageName, packageVersion, waitErr.Error()),
			),
		}
	}
	return nil
}
