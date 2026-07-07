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

package acctest

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/security/gettoken"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/accesstokengranttype"
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
)

// entityStoreCleanupTimeout bounds how long CleanupEntityStore waits for the
// store to reach not_installed. Test code has no resource ctx/timeouts block,
// so it uses a local timeout matching the provider Delete default cadence.
const entityStoreCleanupTimeout = 5 * time.Minute

// CleanupEntityStore uninstalls the Security Entity Store in the given space and
// waits until the store reports not_installed. It is intended to be registered
// with t.Cleanup so that each acceptance test leaves the per-space singleton
// store in a clean state, preventing cross-test contamination from the Kibana
// merge-on-install behavior. It is best-effort and idempotent: uninstalling an
// already-uninstalled store is treated as success and never fails the test.
func CleanupEntityStore(t *testing.T, spaceID string) {
	t.Helper()
	// Guard on TF_ACC directly rather than SkipIfNotAcceptanceTest: this runs
	// inside t.Cleanup, where calling t.Skip would re-skip an already-skipped
	// test and emit confusing duplicate skip logs.
	if os.Getenv("TF_ACC") == "" {
		return
	}

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Logf("CleanupEntityStore: failed to create Kibana client, skipping cleanup: %v", err)
		return
	}
	oapiClient := client.GetKibanaOapiClient()

	ctx, cancel := context.WithTimeout(context.Background(), entityStoreCleanupTimeout)
	defer cancel()

	t.Logf("CleanupEntityStore: uninstalling entity store in space %q", spaceID)
	if diags := kibanaoapi.UninstallSecurityEntityStore(ctx, oapiClient, spaceID, kbapi.PostSecurityEntityStoreUninstallJSONRequestBody{}); diags.HasError() {
		// Idempotency: uninstalling an already not_installed store may error;
		// log and continue to the wait rather than failing the test.
		t.Logf("CleanupEntityStore: uninstall returned diagnostics (continuing): %v", diags.Errors())
	}

	checker := func(ctx context.Context) (bool, error) {
		resp, diags := kibanaoapi.GetSecurityEntityStoreStatus(ctx, oapiClient, spaceID, false)
		if diags.HasError() {
			t.Logf("CleanupEntityStore: transient error reading status, retrying: %v", diags.Errors())
			return false, nil
		}
		var status struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(resp.Body, &status); err != nil {
			t.Logf("CleanupEntityStore: transient error decoding status, retrying: %v", err)
			return false, nil
		}
		return status.Status == string(kbapi.SecurityEntityAnalyticsAPIStoreStatusNotInstalled), nil
	}

	if err := asyncutils.WaitForStateTransition(ctx, "security entity store", spaceID, checker, asyncutils.WithPollInterval(5*time.Second)); err != nil {
		t.Logf("CleanupEntityStore: store did not reach not_installed within %s: %v", entityStoreCleanupTimeout, err)
		return
	}
	t.Logf("CleanupEntityStore: entity store in space %q reached not_installed", spaceID)
}

type TLSMaterial struct {
	CAPEM    string
	CertPEM  string
	KeyPEM   string
	CAFile   string
	CertFile string
	KeyFile  string
}

func CreateESAccessToken(t *testing.T) string {
	t.Helper()
	SkipIfNotAcceptanceTest(t)

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("failed to create acceptance testing client: %v", err)
	}
	typedClient := client.GetESClient()

	username := os.Getenv("ELASTICSEARCH_USERNAME")
	password := os.Getenv("ELASTICSEARCH_PASSWORD")
	grantType := accesstokengranttype.Password

	resp, err := typedClient.Security.GetToken().
		Request(&gettoken.Request{
			GrantType: &grantType,
			Username:  &username,
			Password:  &password,
		}).
		Do(t.Context())
	if err != nil {
		t.Fatalf("failed to create Elasticsearch access token: %v", err)
	}

	if resp.AccessToken == "" {
		t.Fatalf("token response did not include an access_token")
	}

	return resp.AccessToken
}

func CreateTLSMaterial(t *testing.T, commonName string) TLSMaterial {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	if commonName == "" {
		commonName = "terraform-provider-elasticstack-test"
	}

	certificate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER}))
	keyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}))

	tempDir := t.TempDir()
	caFile := filepath.Join(tempDir, "ca.pem")
	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")

	for path, contents := range map[string]string{
		caFile:   certPEM,
		certFile: certPEM,
		keyFile:  keyPEM,
	} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatalf("failed to write TLS test file %s: %v", path, err)
		}
	}

	return TLSMaterial{
		CAPEM:    certPEM,
		CertPEM:  certPEM,
		KeyPEM:   keyPEM,
		CAFile:   caFile,
		CertFile: certFile,
		KeyFile:  keyFile,
	}
}
