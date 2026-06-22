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

package schema_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitElasticsearchConnectionCAFingerprintConflictsWithCAFile(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_elasticsearch_info" "test" {
  elasticsearch_connection {
    ca_fingerprint = "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
    ca_file        = "/path/to/ca.pem"
  }
}
`,
				ExpectError: regexp.MustCompile(`(?s)(Invalid Attribute Combination|ca_fingerprint.*ca_file|ca_file.*ca_fingerprint)`),
			},
		},
	})
}

func TestUnitElasticsearchConnectionCAFingerprintConflictsWithCAData(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_elasticsearch_info" "test" {
  elasticsearch_connection {
    ca_fingerprint = "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
    ca_data        = "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----"
  }
}
`,
				ExpectError: regexp.MustCompile(`(?s)(Invalid Attribute Combination|ca_fingerprint.*ca_data|ca_data.*ca_fingerprint)`),
			},
		},
	})
}

func TestUnitEphemeralElasticsearchConnectionCAFingerprintConflictsWithCAFile(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

ephemeral "elasticstack_elasticsearch_security_api_key" "test" {
  name = "test-key"
  elasticsearch_connection {
    ca_fingerprint = "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
    ca_file        = "/path/to/ca.pem"
  }
}
`,
				ExpectError: regexp.MustCompile(`(?s)(Invalid Attribute Combination|ca_fingerprint.*ca_file|ca_file.*ca_fingerprint)`),
			},
		},
	})
}
