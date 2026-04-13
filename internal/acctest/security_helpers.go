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
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
)

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

	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("failed to create acceptance testing client: %v", err)
	}
	esClient, err := client.GetESClient()
	if err != nil {
		t.Fatalf("failed to get Elasticsearch client: %v", err)
	}

	payload, err := json.Marshal(map[string]string{
		"grant_type": "password",
		"username":   os.Getenv("ELASTICSEARCH_USERNAME"),
		"password":   os.Getenv("ELASTICSEARCH_PASSWORD"),
	})
	if err != nil {
		t.Fatalf("failed to marshal token request: %v", err)
	}

	resp, err := esClient.Security.GetToken(
		bytes.NewReader(payload),
		esClient.Security.GetToken.WithContext(t.Context()),
	)
	if err != nil {
		t.Fatalf("failed to create Elasticsearch access token: %v", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			t.Fatalf("failed to create Elasticsearch access token: status %d (additionally failed to read error response: %v)", resp.StatusCode, readErr)
		}
		t.Fatalf("failed to create Elasticsearch access token: status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}
	if tokenResponse.AccessToken == "" {
		t.Fatalf("token response did not include an access_token")
	}

	return tokenResponse.AccessToken
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
