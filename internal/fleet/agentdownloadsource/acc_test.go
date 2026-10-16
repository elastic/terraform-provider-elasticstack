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

package agentdownloadsource_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const testResourceName = "elasticstack_fleet_agent_download_source.test"

var (
	minVersionFleetAgentDownloadSource     = version.Must(version.NewVersion("8.13.0"))
	minVersionFleetAgentDownloadSourceSsl  = version.Must(version.NewVersion("9.1.0"))
	minVersionFleetAgentDownloadSourceAuth = version.Must(version.NewVersion("9.4.0"))
)

func TestAccResourceFleetAgentDownloadSource_AuthBasic(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	password := fmt.Sprintf("secret-%s", random)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(password),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username", "svc-account"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password", password),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.api_key"),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthBasicImport(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	password := fmt.Sprintf("secret-%s", random)
	resourceName := "elasticstack_fleet_agent_download_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(password),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auth.username", "svc-account"),
					resource.TestCheckResourceAttr(resourceName, "auth.password", password),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.0.key", "X-A"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.0.value", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(password),
				},
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth.password"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					res := s.RootModule().Resources[resourceName]
					if res == nil || res.Primary == nil {
						return "", fmt.Errorf("resource %s not found in state", resourceName)
					}
					return fmt.Sprintf("default/%s", res.Primary.ID), nil
				},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported state, got %d", len(states))
					}
					if importedPassword, ok := states[0].Attributes["auth.password"]; ok {
						return fmt.Errorf("expected auth.password to be null after import, got %q", importedPassword)
					}
					if states[0].Attributes["auth.username"] != "svc-account" {
						return fmt.Errorf("expected auth.username to be populated from the API, got %q", states[0].Attributes["auth.username"])
					}
					if states[0].Attributes["auth.headers.#"] != "1" ||
						states[0].Attributes["auth.headers.0.key"] != "X-A" ||
						states[0].Attributes["auth.headers.0.value"] != "1" {
						return fmt.Errorf("expected auth.headers to be populated from the API, got key %q value %q", states[0].Attributes["auth.headers.0.key"], states[0].Attributes["auth.headers.0.value"])
					}
					return nil
				},
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthAPIKey(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	apiKey := fmt.Sprintf("key-%s", random)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("api_key"),
				ConfigVariables: config.Variables{
					"suffix":  config.StringVariable(random),
					"api_key": config.StringVariable(apiKey),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.api_key", apiKey),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password"),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthAPIKeyImport(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	apiKey := fmt.Sprintf("key-%s", random)
	resourceName := "elasticstack_fleet_agent_download_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("api_key"),
				ConfigVariables: config.Variables{
					"suffix":  config.StringVariable(random),
					"api_key": config.StringVariable(apiKey),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auth.api_key", apiKey),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.0.key", "X-A"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.0.value", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "auth.username"),
					resource.TestCheckNoResourceAttr(resourceName, "auth.password"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("api_key"),
				ConfigVariables: config.Variables{
					"suffix":  config.StringVariable(random),
					"api_key": config.StringVariable(apiKey),
				},
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth.api_key"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					res := s.RootModule().Resources[resourceName]
					if res == nil || res.Primary == nil {
						return "", fmt.Errorf("resource %s not found in state", resourceName)
					}
					return fmt.Sprintf("default/%s", res.Primary.ID), nil
				},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported state, got %d", len(states))
					}
					if importedAPIKey, ok := states[0].Attributes["auth.api_key"]; ok {
						return fmt.Errorf("expected auth.api_key to be null after import, got %q", importedAPIKey)
					}
					return nil
				},
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthPasswordWithAPIKey(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("password_and_api_key"),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)Invalid Attribute Combination`),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthPasswordWithoutUsername(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("password_only"),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)Invalid Attribute Combination`),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthUsernameWithoutPassword(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("username_only"),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)Invalid Attribute Combination`),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthEmpty(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_fleet_agent_download_source.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthHeadersOnly(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("headers_only"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.headers.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.headers.0.key", "X-A"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.headers.0.value", "1"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.api_key"),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthHeadersEmpty(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_headers"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.headers.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthHeadersMax(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("max"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.headers.#", "100"),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthRemove(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	password := fmt.Sprintf("secret-%s", random)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_auth"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(password),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username", "svc-account"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password", password),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("removed"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.api_key"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.headers"),
					testCheckFleetAgentDownloadSourceServerAuthCleared("elasticstack_fleet_agent_download_source.test"),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthOmittedWhenNeverConfigured(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Auth Omitted Agent Download Source %s", random)),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.api_key"),
					testCheckFleetAgentDownloadSourceServerAuthCleared("elasticstack_fleet_agent_download_source.test"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("rename"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Auth Omitted Renamed Agent Download Source %s", random)),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.api_key"),
					testCheckFleetAgentDownloadSourceServerAuthCleared("elasticstack_fleet_agent_download_source.test"),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthBasicToAPIKey(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	password := fmt.Sprintf("secret-%s", random)
	apiKey := fmt.Sprintf("key-%s", random)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(password),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username", "svc-account"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password", password),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.api_key"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("api_key"),
				ConfigVariables: config.Variables{
					"suffix":  config.StringVariable(random),
					"api_key": config.StringVariable(apiKey),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.api_key", apiKey),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password"),
					testCheckFleetAgentDownloadSourceServerBasicAuthAbsent("elasticstack_fleet_agent_download_source.test"),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthPasswordRotation(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	originalPassword := fmt.Sprintf("secret-%s", random)
	rotatedPassword := fmt.Sprintf("rotated-%s", random)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("password"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(originalPassword),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username", "svc-account"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password", originalPassword),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("password"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(rotatedPassword),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username", "svc-account"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password", rotatedPassword),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthPasswordRefresh(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	password := fmt.Sprintf("secret-%s", random)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("password"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(password),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password", password),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("password"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(password),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.username", "svc-account"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "auth.password", password),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthUsernameAndHeadersFromAPI(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	password := fmt.Sprintf("secret-%s", random)
	resourceName := "elasticstack_fleet_agent_download_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(password),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auth.username", "svc-account"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.0.key", "X-A"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.0.value", "1"),
					testCheckFleetAgentDownloadSourceServerUsernameAndHeaders(resourceName, "svc-account", "X-A", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(password),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auth.username", "svc-rotated"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.0.key", "X-B"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.0.value", "2"),
					resource.TestCheckResourceAttr(resourceName, "auth.password", password),
					testCheckFleetAgentDownloadSourceServerUsernameAndHeaders(resourceName, "svc-rotated", "X-B", "2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(password),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auth.username", "svc-rotated"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.0.key", "X-B"),
					resource.TestCheckResourceAttr(resourceName, "auth.headers.0.value", "2"),
					resource.TestCheckResourceAttr(resourceName, "auth.password", password),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthHeadersOverMax(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("over_max"),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)list must contain at most 100 elements`),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthRejectedBelowMinimumVersion(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSource, versionutils.FlavorAny)

	unsupported, err := versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSourceAuth)()
	if err != nil {
		t.Fatalf("failed to check stack version: %v", err)
	}
	if !unsupported {
		t.Skip("stack supports auth; skipping version gate test")
	}

	random := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_auth"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(random),
					"password": config.StringVariable(fmt.Sprintf("secret-%s", random)),
				},
				ExpectError: regexp.MustCompile("(?s)`auth` is not supported on this version"),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_AuthOmittedBelowAuthMinimumVersion(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSource, versionutils.FlavorAny)

	unsupported, err := versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSourceAuth)()
	if err != nil {
		t.Fatalf("failed to check stack version: %v", err)
	}
	if !unsupported {
		t.Skip("stack supports auth; skipping version gate test")
	}

	random := sdkacctest.RandString(8)
	resourceName := "elasticstack_fleet_agent_download_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_auth"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("Agent Download Source %s", random)),
					resource.TestCheckNoResourceAttr(resourceName, "auth.username"),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSource, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	var idBeforeReplacement string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "host", "https://artifacts.elastic.co/downloads/elastic-agent"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "proxy_id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "source_id", fmt.Sprintf("agent-download-source-%s", random)),
					resource.TestCheckNoResourceAttr(testResourceName, "ssl.certificate"),
					resource.TestCheckNoResourceAttr(testResourceName, "ssl.key"),
					resource.TestCheckNoResourceAttr(testResourceName, "ssl.certificate_authorities.#"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "id", fmt.Sprintf("agent-download-source-%s", random)),
					resource.TestCheckResourceAttrPair("elasticstack_fleet_agent_download_source.test", "id", "elasticstack_fleet_agent_download_source.test", "source_id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Updated Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "host", "https://artifacts.elastic.co/downloads/elastic-agent-updated"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "proxy_id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "source_id", fmt.Sprintf("agent-download-source-%s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				ResourceName:            "elasticstack_fleet_agent_download_source.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"space_ids"},
				ImportStateIdFunc:       testFleetAgentDownloadSourceImportStateIDDefaultSpace,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				ResourceName:            "elasticstack_fleet_agent_download_source.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"space_ids"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					res := s.RootModule().Resources["elasticstack_fleet_agent_download_source.test"]
					if res == nil || res.Primary == nil {
						return "", fmt.Errorf("resource elasticstack_fleet_agent_download_source.test not found in state")
					}
					return res.Primary.ID, nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("omit_optionals"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					testCheckFleetAgentDownloadSourceCaptureID(&idBeforeReplacement),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("No Optionals Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "host", "https://artifacts.elastic.co/downloads/elastic-agent-no-optionals"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "proxy_id"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_agent_download_source.test", "source_id"),
					resource.TestCheckResourceAttrPair("elasticstack_fleet_agent_download_source.test", "id", "elasticstack_fleet_agent_download_source.test", "source_id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_space_ids"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Empty Space IDs Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("replace_source_id"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					testCheckFleetAgentDownloadSourceIDChanged(&idBeforeReplacement),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "source_id", fmt.Sprintf("agent-download-source-replaced-%s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "id", fmt.Sprintf("agent-download-source-replaced-%s", random)),
					resource.TestCheckResourceAttrPair("elasticstack_fleet_agent_download_source.test", "id", "elasticstack_fleet_agent_download_source.test", "source_id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("non_default_space"),
				ConfigVariables: config.Variables{
					"suffix":               config.StringVariable(random),
					"non_default_space_id": config.StringVariable(fmt.Sprintf("fleet-agent-download-source-%s", random)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					testCheckFleetAgentDownloadSourceSpaceContains("elasticstack_fleet_agent_download_source.test", fmt.Sprintf("fleet-agent-download-source-%s", random)),
				),
			},
		},
	})
}

func TestAccResourceFleetAgentDownloadSource_SSL(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceSsl, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	var idBeforeSSLUpdate string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectSensitiveValue("elasticstack_fleet_agent_download_source.test", tfjsonpath.New("ssl").AtMapKey("key")),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testCheckFleetAgentDownloadSourceCaptureID(&idBeforeSSLUpdate),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate", "PEM-CERT"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "ssl.key", "PEM-KEY"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate_authorities.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate_authorities.*", "PEM-CA-1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate_authorities.*", "PEM-CA-2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl_updated"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate", "PEM-CERT-2"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "ssl.key", "PEM-KEY-2"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate_authorities.#", "3"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate_authorities.*", "PEM-CA-3"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate_authorities.*", "PEM-CA-4"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate_authorities.*", "PEM-CA-5"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl_updated"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				ResourceName:            "elasticstack_fleet_agent_download_source.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"space_ids", "ssl.key"}, // managed secret, redacted by Fleet
				ImportStateIdFunc:       testFleetAgentDownloadSourceImportStateIDDefaultSpace,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("clear_ssl"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					testCheckFleetAgentDownloadSourceIDUnchanged(&idBeforeSSLUpdate),
					resource.TestCheckNoResourceAttr(testResourceName, "ssl.certificate"),
					resource.TestCheckNoResourceAttr(testResourceName, "ssl.key"),
					testCheckNoFlatmapKeyPrefix("ssl.certificate_authorities."),
				),
			},
		},
	})
}

func testFleetAgentDownloadSourceImportStateIDDefaultSpace(s *terraform.State) (string, error) {
	res := s.RootModule().Resources[testResourceName]
	if res == nil || res.Primary == nil {
		return "", fmt.Errorf("resource %s not found in state", testResourceName)
	}
	return fmt.Sprintf("default/%s", res.Primary.ID), nil
}

func testCheckFleetAgentDownloadSourceCaptureID(target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[testResourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", testResourceName)
		}
		*target = res.Primary.ID
		return nil
	}
}

func testCheckFleetAgentDownloadSourceIDChanged(previousID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[testResourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", testResourceName)
		}
		if *previousID == "" {
			return fmt.Errorf("previous ID was not captured")
		}
		if res.Primary.ID == *previousID {
			return fmt.Errorf("expected resource ID to change after source_id replacement, but remained %q", res.Primary.ID)
		}
		return nil
	}
}

func testCheckFleetAgentDownloadSourceIDUnchanged(previousID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[testResourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", testResourceName)
		}
		if *previousID == "" {
			return fmt.Errorf("previous ID was not captured")
		}
		if res.Primary.ID != *previousID {
			return fmt.Errorf("expected resource ID to remain %q, but changed to %q", *previousID, res.Primary.ID)
		}
		return nil
	}
}

func testCheckFleetAgentDownloadSourceSpaceContains(resourceName, spaceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		for k, v := range res.Primary.Attributes {
			if k == "space_ids.#" {
				continue
			}
			if strings.HasPrefix(k, "space_ids.") && v == spaceID {
				return nil
			}
		}
		return fmt.Errorf("expected space_ids to contain %q", spaceID)
	}
}

type fleetAgentDownloadSourcePayload struct {
	Item struct {
		Auth *struct {
			Username *string `json:"username"`
			Password *string `json:"password"`
			APIKey   *string `json:"api_key"`
			Headers  []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"headers"`
		} `json:"auth"`
		Secrets *struct {
			Auth *struct {
				Password json.RawMessage `json:"password"`
				APIKey   json.RawMessage `json:"api_key"`
			} `json:"auth"`
		} `json:"secrets"`
	} `json:"item"`
}

func readFleetAgentDownloadSourcePayload(s *terraform.State, resourceName string) (fleetAgentDownloadSourcePayload, error) {
	var payload fleetAgentDownloadSourcePayload

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok || rs.Primary == nil {
		return payload, fmt.Errorf("resource %s not found in state", resourceName)
	}

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return payload, err
	}

	resp, diags := fleet.GetAgentDownloadSource(context.Background(), client.GetFleetClient(), rs.Primary.ID, getOperationalSpaceFromResourceState(rs))
	if diags.HasError() {
		return payload, diagutil.FwDiagsAsError(diags)
	}
	if resp == nil || resp.JSON200 == nil {
		return payload, fmt.Errorf("agent download source %s was not found", rs.Primary.ID)
	}
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

func testCheckFleetAgentDownloadSourceServerUsernameAndHeaders(resourceName, username, headerKey, headerValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		payload, err := readFleetAgentDownloadSourcePayload(s, resourceName)
		if err != nil {
			return err
		}
		if payload.Item.Auth == nil || payload.Item.Auth.Username == nil {
			return fmt.Errorf("expected server auth username %q, got none", username)
		}
		if *payload.Item.Auth.Username != username {
			return fmt.Errorf("expected server auth username %q, got %q", username, *payload.Item.Auth.Username)
		}
		if len(payload.Item.Auth.Headers) != 1 {
			return fmt.Errorf("expected 1 server auth header, got %d", len(payload.Item.Auth.Headers))
		}
		header := payload.Item.Auth.Headers[0]
		if header.Key != headerKey || header.Value != headerValue {
			return fmt.Errorf("expected server auth header %s=%s, got %s=%s", headerKey, headerValue, header.Key, header.Value)
		}
		return nil
	}
}

func testCheckFleetAgentDownloadSourceServerBasicAuthAbsent(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		payload, err := readFleetAgentDownloadSourcePayload(s, resourceName)
		if err != nil {
			return err
		}
		if payload.Item.Auth != nil && payload.Item.Auth.Username != nil {
			return fmt.Errorf("expected server auth username to be absent, got %q", *payload.Item.Auth.Username)
		}
		if payload.Item.Auth != nil && payload.Item.Auth.Password != nil {
			return fmt.Errorf("expected server auth password to be absent")
		}
		if payload.Item.Secrets != nil && payload.Item.Secrets.Auth != nil && jsonRawPresent(payload.Item.Secrets.Auth.Password) {
			return fmt.Errorf("expected server-side password secret to be absent")
		}
		return nil
	}
}

func testCheckFleetAgentDownloadSourceServerAuthCleared(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		payload, err := readFleetAgentDownloadSourcePayload(s, resourceName)
		if err != nil {
			return err
		}
		if payload.Item.Auth != nil {
			if payload.Item.Auth.Username != nil {
				return fmt.Errorf("expected server auth username to be cleared, got %q", *payload.Item.Auth.Username)
			}
			if payload.Item.Auth.Password != nil {
				return fmt.Errorf("expected server auth password to be cleared")
			}
			if payload.Item.Auth.APIKey != nil {
				return fmt.Errorf("expected server auth api_key to be cleared")
			}
			if len(payload.Item.Auth.Headers) > 0 {
				return fmt.Errorf("expected server auth headers to be cleared, got %d", len(payload.Item.Auth.Headers))
			}
		}
		if payload.Item.Secrets != nil && payload.Item.Secrets.Auth != nil && (jsonRawPresent(payload.Item.Secrets.Auth.Password) || jsonRawPresent(payload.Item.Secrets.Auth.APIKey)) {
			return fmt.Errorf("expected server-side auth secrets to be cleared")
		}
		return nil
	}
}

func jsonRawPresent(raw json.RawMessage) bool {
	return len(raw) > 0 && string(raw) != "null"
}

func checkResourceFleetAgentDownloadSourceDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_agent_download_source" {
			continue
		}

		fleetClient := client.GetFleetClient()
		spaceID := getOperationalSpaceFromResourceState(rs)
		resp, diags := fleet.GetAgentDownloadSource(context.Background(), fleetClient, rs.Primary.ID, spaceID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if resp != nil && resp.JSON200 != nil {
			return fmt.Errorf("fleet agent download source id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}

func getOperationalSpaceFromResourceState(rs *terraform.ResourceState) string {
	for k, v := range rs.Primary.Attributes {
		if strings.HasPrefix(k, "space_ids.") && k != "space_ids.#" && v != "" {
			return v
		}
	}
	return ""
}

func TestAccResourceFleetAgentDownloadSource_SSLChildRemoval(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceSsl, versionutils.FlavorAny)

	for _, transition := range []string{"drop_key", "drop_cas", "drop_cert"} {
		random := sdkacctest.RandString(8)
		var idBeforeChildRemoval string

		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acctest.PreCheck(t) },
			CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
			Steps: []resource.TestStep{
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl"),
					ConfigVariables: config.Variables{
						"suffix": config.StringVariable(random),
					},
					Check: testCheckFleetAgentDownloadSourceCaptureID(&idBeforeChildRemoval),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory(transition),
					ConfigVariables: config.Variables{
						"suffix": config.StringVariable(random),
					},
					Check: resource.ComposeTestCheckFunc(
						testCheckFleetAgentDownloadSourceIDUnchanged(&idBeforeChildRemoval),
						childRemovalChecks(transition),
					),
				},
			},
		})
	}
}

func childRemovalChecks(transition string) resource.TestCheckFunc {
	const (
		caList = "ssl.certificate_authorities"
		ca1    = "PEM-CA-1"
		ca2    = "PEM-CA-2"
	)

	retainedCAs := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(testResourceName, caList+".#", "2"),
		resource.TestCheckResourceAttr(testResourceName, caList+".0", ca1),
		resource.TestCheckResourceAttr(testResourceName, caList+".1", ca2),
	)

	switch transition {
	case "drop_key":
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(testResourceName, "ssl.certificate", "PEM-CERT"),
			retainedCAs,
			resource.TestCheckNoResourceAttr(testResourceName, "ssl.key"),
		)
	case "drop_cas":
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(testResourceName, "ssl.certificate", "PEM-CERT"),
			resource.TestCheckResourceAttr(testResourceName, "ssl.key", "PEM-KEY"),
			testCheckNoFlatmapKeyPrefix(caList+"."),
		)
	case "drop_cert":
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(testResourceName, "ssl.key", "PEM-KEY"),
			retainedCAs,
			resource.TestCheckNoResourceAttr(testResourceName, "ssl.certificate"),
		)
	}
	return func(*terraform.State) error { return fmt.Errorf("unknown transition %q", transition) }
}

// testCheckNoFlatmapKeyPrefix asserts no state flatmap key carries the given
// prefix, catching leftover list elements that a key-absence check on the
// "#" count alone would miss.
func testCheckNoFlatmapKeyPrefix(prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[testResourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", testResourceName)
		}
		for k := range res.Primary.Attributes {
			if strings.HasPrefix(k, prefix) {
				return fmt.Errorf("expected no state keys with prefix %q, found %q", prefix, k)
			}
		}
		return nil
	}
}

func TestAccResourceFleetAgentDownloadSource_SSLValidation(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceSsl, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ssl_empty_object"),
				ExpectError:              regexp.MustCompile(`(?s)Empty ssl configuration.*At least one of certificate_authorities, certificate, key, or key_wo`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ssl_empty_certificate"),
				ExpectError:              regexp.MustCompile(`(?i)certificate.*string length must be at least 1`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ssl_empty_key"),
				ExpectError:              regexp.MustCompile(`(?i)key.*string length must be at least 1`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ssl_empty_ca_list"),
				ExpectError:              regexp.MustCompile(`(?i)certificate_authorities.*must contain at least 1`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ssl_empty_ca_element"),
				ExpectError:              regexp.MustCompile(`(?s)(?i)certificate_authorities.*string length must be at least 1`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ssl_ca_11"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(sdkacctest.RandString(8)),
				},
				ExpectError: regexp.MustCompile(`(?i)certificate_authorities.*at most 10`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ssl_ca_10"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(sdkacctest.RandString(8)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate_authorities.#", "10"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate_authorities.*", "PEM-CA-1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "ssl.certificate_authorities.*", "PEM-CA-10"),
				),
			},
		},
	})
}

// TestAccResourceFleetAgentDownloadSource_ProxyID exercises the proxy_id attribute with a
// real value, wiring in an elasticstack_fleet_proxy resource. This covers the non-empty
// ProxyId code paths in create.go/update.go, and the update path back to an unset proxy_id.
func TestAccResourceFleetAgentDownloadSource_ProxyID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSource, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_proxy"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "proxy_id", fmt.Sprintf("agent-download-source-proxy-%s", random)),
					resource.TestCheckResourceAttrPair("elasticstack_fleet_agent_download_source.test", "proxy_id", "elasticstack_fleet_proxy.test", "proxy_id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_proxy"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "proxy_id"),
				),
			},
		},
	})
}

// TestAccResourceFleetAgentDownloadSource_DefaultTrue exercises the IsDefault boolean
// round-trip in read.go for the true branch, and toggling the value via update.
// The final step returns default to false because the Fleet API rejects deleting
// the current default download source.
func TestAccResourceFleetAgentDownloadSource_DefaultTrue(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSource, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("default_false"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("default_true"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("default_false"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "false"),
				),
			},
		},
	})
}

// TestAccResourceFleetAgentDownloadSource_SpaceIDsUpdate exercises an in-place update that
// changes space_ids from a single-element set to a different, larger (2+ element) set
// that does not include the prior operational space. This covers update.go's prior-space
// resolution: the update must target the space where the resource currently exists
// (prior state), not an arbitrary element of the new plan set.
func TestAccResourceFleetAgentDownloadSource_SpaceIDsUpdate(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSource, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	secondSpaceID := fmt.Sprintf("fleet-agent-download-source-space-update-%s", random)
	thirdSpaceID := fmt.Sprintf("fleet-agent-download-source-space-update-b-%s", random)
	var idBeforeUpdate string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("single_space"),
				ConfigVariables: config.Variables{
					"suffix":          config.StringVariable(random),
					"second_space_id": config.StringVariable(secondSpaceID),
					"third_space_id":  config.StringVariable(thirdSpaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					testCheckFleetAgentDownloadSourceCaptureID(&idBeforeUpdate),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multi_space"),
				ConfigVariables: config.Variables{
					"suffix":          config.StringVariable(random),
					"second_space_id": config.StringVariable(secondSpaceID),
					"third_space_id":  config.StringVariable(thirdSpaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", secondSpaceID),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", thirdSpaceID),
					// id mirrors source_id for this resource, so a stable id is a
					// sanity check that source_id did not change — not proof of an
					// in-place update versus destroy/recreate with the same source_id.
					resource.TestCheckResourceAttrWith("elasticstack_fleet_agent_download_source.test", "id", func(value string) error {
						if value != idBeforeUpdate {
							return fmt.Errorf("expected id to remain %q after space_ids update, got %q", idBeforeUpdate, value)
						}
						return nil
					}),
				),
			},
		},
	})
}

// Write-only secret acceptance tests. They share testdata/write_only, a single
// configuration whose optional inputs default to null, and need Terraform 1.11+
// for write-only attributes.

const writeOnlyConfigDirectory = "testdata/write_only"

func writeOnlyTerraformVersionChecks() []tfversion.TerraformVersionCheck {
	return []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)}
}

// writeOnlyVars builds ConfigVariables for the shared write-only configuration
// from key/value pairs. Unset variables default to null in the configuration.
func writeOnlyVars(suffix string, kv ...string) config.Variables {
	vars := config.Variables{"suffix": config.StringVariable(suffix)}
	for i := 0; i+1 < len(kv); i += 2 {
		if kv[i] == "headers" {
			vars[kv[i]] = config.BoolVariable(kv[i+1] == "true")
			continue
		}
		vars[kv[i]] = config.StringVariable(kv[i+1])
	}
	return vars
}

func writeOnlyStep(suffix string, kv ...string) resource.TestStep {
	return resource.TestStep{
		ProtoV6ProviderFactories: acctest.Providers,
		ConfigDirectory:          config.StaticDirectory(writeOnlyConfigDirectory),
		ConfigVariables:          writeOnlyVars(suffix, kv...),
	}
}

func writeOnlyUnchangedStep(suffix string, kv ...string) resource.TestStep {
	step := writeOnlyStep(suffix, kv...)
	step.ConfigPlanChecks = resource.ConfigPlanChecks{
		PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}
	return step
}

func writeOnlyPlannedUpdateStep(suffix string, checks resource.TestCheckFunc, kv ...string) resource.TestStep {
	step := writeOnlyStep(suffix, kv...)
	step.ConfigPlanChecks = resource.ConfigPlanChecks{
		PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionUpdate)},
	}
	step.Check = checks
	return step
}

// fetchServerItem reads the download source as returned by Fleet.
func fetchServerItem(s *terraform.State) (map[string]any, error) {
	rs, ok := s.RootModule().Resources[testResourceName]
	if !ok || rs.Primary == nil {
		return nil, fmt.Errorf("resource %s not found in state", testResourceName)
	}
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return nil, err
	}
	resp, diags := fleet.GetAgentDownloadSource(context.Background(), client.GetFleetClient(), rs.Primary.ID, getOperationalSpaceFromResourceState(rs))
	if diags.HasError() {
		return nil, diagutil.FwDiagsAsError(diags)
	}
	if resp == nil || resp.JSON200 == nil {
		return nil, fmt.Errorf("agent download source %s was not found", rs.Primary.ID)
	}
	var body struct {
		Item map[string]any `json:"item"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return nil, err
	}
	return body.Item, nil
}

func serverNested(item map[string]any, parent, field string) any {
	m, ok := item[parent].(map[string]any)
	if !ok {
		return nil
	}
	return m[field]
}

// serverSecretID returns the managed secret id stored at secrets.<parent>.<field>,
// or "" when Fleet has no secret reference there.
func serverSecretID(item map[string]any, parent, field string) string {
	secrets, ok := item["secrets"].(map[string]any)
	if !ok {
		return ""
	}
	ref, ok := serverNested(secrets, parent, field).(map[string]any)
	if !ok {
		return ""
	}
	id, _ := ref["id"].(string)
	return id
}

func testCheckServerSecretPresent(parent, field string, capture *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		item, err := fetchServerItem(s)
		if err != nil {
			return err
		}
		id := serverSecretID(item, parent, field)
		if id == "" {
			return fmt.Errorf("expected Fleet to hold a managed secret at secrets.%s.%s, got none (item: %v)", parent, field, item["secrets"])
		}
		if capture != nil {
			*capture = id
		}
		return nil
	}
}

func testCheckServerSecretAbsent(parent, field string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		item, err := fetchServerItem(s)
		if err != nil {
			return err
		}
		if id := serverSecretID(item, parent, field); id != "" {
			return fmt.Errorf("expected no managed secret at secrets.%s.%s, got id %q", parent, field, id)
		}
		return nil
	}
}

// testCheckServerSecretUnchanged asserts the managed secret id equals the
// captured one: the resource referenced it by {id} instead of resending.
func testCheckServerSecretUnchanged(parent, field string, previous *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		item, err := fetchServerItem(s)
		if err != nil {
			return err
		}
		if id := serverSecretID(item, parent, field); id != *previous {
			return fmt.Errorf("expected secrets.%s.%s id to stay %q, got %q", parent, field, *previous, id)
		}
		return nil
	}
}

// testCheckServerSecretRotated asserts a new managed secret id was minted
// (a literal was sent) and records it as the new reference.
func testCheckServerSecretRotated(parent, field string, previous *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		item, err := fetchServerItem(s)
		if err != nil {
			return err
		}
		id := serverSecretID(item, parent, field)
		if id == "" || id == *previous {
			return fmt.Errorf("expected secrets.%s.%s to hold a new secret id, previous %q, got %q", parent, field, *previous, id)
		}
		*previous = id
		return nil
	}
}

// testCheckServerNoPlainSecret asserts Fleet does not expose the secret value as
// a plain field.
func testCheckServerNoPlainSecret(parent, field string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		item, err := fetchServerItem(s)
		if err != nil {
			return err
		}
		if v := serverNested(item, parent, field); v != nil {
			return fmt.Errorf("expected %s.%s to be absent from the Fleet response, got a value", parent, field)
		}
		return nil
	}
}

func writeOnlyStateChecks(nullPaths ...tfjsonpath.Path) []statecheck.StateCheck {
	checks := make([]statecheck.StateCheck, 0, len(nullPaths))
	for _, p := range nullPaths {
		checks = append(checks, statecheck.ExpectKnownValue(testResourceName, p, knownvalue.Null()))
	}
	return checks
}

var (
	pathSSLKey         = tfjsonpath.New("ssl").AtMapKey("key")
	pathSSLKeyWO       = tfjsonpath.New("ssl").AtMapKey("key_wo")
	pathAuthPassword   = tfjsonpath.New("auth").AtMapKey("password")
	pathAuthPasswordWO = tfjsonpath.New("auth").AtMapKey("password_wo")
	pathAuthAPIKey     = tfjsonpath.New("auth").AtMapKey("api_key")
	pathAuthAPIKeyWO   = tfjsonpath.New("auth").AtMapKey("api_key_wo")
)

// TestAccResourceFleetAgentDownloadSource_SSLKeyWriteOnly covers the key_wo-only
// ssl lifecycle (tasks 8.1, 8.9): the parent ssl object survives although no
// ordinary attribute is set, key_wo never reaches state, an unchanged value is
// a no-op that keeps the managed secret, an unrelated update references the
// secret by id, and a changed value mints a new secret.
func TestAccResourceFleetAgentDownloadSource_SSLKeyWriteOnly(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceSsl, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	var secretID string

	create := writeOnlyStep(random, "key_wo", "key-one")
	create.ConfigStateChecks = append(
		writeOnlyStateChecks(pathSSLKeyWO, pathSSLKey),
		statecheck.ExpectKnownValue(testResourceName, tfjsonpath.New("ssl"), knownvalue.NotNull()),
	)
	create.Check = resource.ComposeTestCheckFunc(
		resource.TestCheckNoResourceAttr(testResourceName, "ssl.key_wo"),
		resource.TestCheckNoResourceAttr(testResourceName, "ssl.key"),
		testCheckServerSecretPresent("ssl", "key", &secretID),
		testCheckServerNoPlainSecret("ssl", "key"),
	)

	rename := writeOnlyStep(random, "key_wo", "key-one", "label", "renamed")
	rename.ConfigStateChecks = writeOnlyStateChecks(pathSSLKeyWO, pathSSLKey)
	rename.Check = resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(testResourceName, "name", fmt.Sprintf("WO Agent Download Source %s renamed", random)),
		testCheckServerSecretUnchanged("ssl", "key", &secretID),
	)

	rotate := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
		resource.TestCheckNoResourceAttr(testResourceName, "ssl.key_wo"),
		testCheckServerSecretRotated("ssl", "key", &secretID),
		testCheckServerNoPlainSecret("ssl", "key"),
	), "key_wo", "key-two", "label", "renamed")
	rotate.ConfigStateChecks = writeOnlyStateChecks(pathSSLKeyWO, pathSSLKey)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		CheckDestroy:           checkResourceFleetAgentDownloadSourceDestroy,
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps: []resource.TestStep{
			create,
			writeOnlyUnchangedStep(random, "key_wo", "key-one"),
			rename,
			writeOnlyUnchangedStep(random, "key_wo", "key-one", "label", "renamed"),
			rotate,
			writeOnlyUnchangedStep(random, "key_wo", "key-two", "label", "renamed"),
		},
	})
}

// TestAccResourceFleetAgentDownloadSource_AuthPasswordWriteOnly covers the
// password_wo lifecycle (task 8.2) with a username: neither write-only value
// nor the plain sibling reaches state, an unchanged value keeps the managed
// secret, an unrelated update references it by id, and a changed value mints a
// new one.
func TestAccResourceFleetAgentDownloadSource_AuthPasswordWriteOnly(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	var secretID string

	create := writeOnlyStep(random, "username", "svc-account", "password_wo", "pass-one")
	create.ConfigStateChecks = append(
		writeOnlyStateChecks(pathAuthPasswordWO, pathAuthPassword, pathAuthAPIKeyWO, pathAuthAPIKey),
		statecheck.ExpectKnownValue(testResourceName, tfjsonpath.New("auth").AtMapKey("username"), knownvalue.StringExact("svc-account")),
	)
	create.Check = resource.ComposeTestCheckFunc(
		resource.TestCheckNoResourceAttr(testResourceName, "auth.password_wo"),
		resource.TestCheckNoResourceAttr(testResourceName, "auth.password"),
		testCheckServerSecretPresent("auth", "password", &secretID),
		testCheckServerNoPlainSecret("auth", "password"),
	)

	rename := writeOnlyStep(random, "username", "svc-account", "password_wo", "pass-one", "label", "renamed")
	rename.ConfigStateChecks = writeOnlyStateChecks(pathAuthPasswordWO, pathAuthPassword)
	rename.Check = testCheckServerSecretUnchanged("auth", "password", &secretID)

	rotate := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
		resource.TestCheckNoResourceAttr(testResourceName, "auth.password_wo"),
		testCheckServerSecretRotated("auth", "password", &secretID),
	), "username", "svc-account", "password_wo", "pass-two", "label", "renamed")
	rotate.ConfigStateChecks = writeOnlyStateChecks(pathAuthPasswordWO, pathAuthPassword)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		CheckDestroy:           checkResourceFleetAgentDownloadSourceDestroy,
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps: []resource.TestStep{
			create,
			writeOnlyUnchangedStep(random, "username", "svc-account", "password_wo", "pass-one"),
			rename,
			writeOnlyUnchangedStep(random, "username", "svc-account", "password_wo", "pass-one", "label", "renamed"),
			rotate,
			writeOnlyUnchangedStep(random, "username", "svc-account", "password_wo", "pass-two", "label", "renamed"),
		},
	})
}

// TestAccResourceFleetAgentDownloadSource_AuthAPIKeyWriteOnly covers the
// api_key_wo lifecycle (task 8.2) and the removal of api_key_wo while the auth
// object stays (headers only), which relies on omitting secrets.auth.api_key
// clearing the remote secret (task 8.10).
func TestAccResourceFleetAgentDownloadSource_AuthAPIKeyWriteOnly(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	var secretID string

	create := writeOnlyStep(random, "api_key_wo", "key-one", "headers", "true")
	create.ConfigStateChecks = writeOnlyStateChecks(pathAuthAPIKeyWO, pathAuthAPIKey)
	create.Check = resource.ComposeTestCheckFunc(
		resource.TestCheckNoResourceAttr(testResourceName, "auth.api_key_wo"),
		resource.TestCheckNoResourceAttr(testResourceName, "auth.api_key"),
		testCheckServerSecretPresent("auth", "api_key", &secretID),
		testCheckServerNoPlainSecret("auth", "api_key"),
	)

	rename := writeOnlyStep(random, "api_key_wo", "key-one", "headers", "true", "label", "renamed")
	rename.ConfigStateChecks = writeOnlyStateChecks(pathAuthAPIKeyWO, pathAuthAPIKey)
	rename.Check = testCheckServerSecretUnchanged("auth", "api_key", &secretID)

	rotate := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
		testCheckServerSecretRotated("auth", "api_key", &secretID),
	), "api_key_wo", "key-two", "headers", "true", "label", "renamed")
	rotate.ConfigStateChecks = writeOnlyStateChecks(pathAuthAPIKeyWO, pathAuthAPIKey)

	// Only api_key_wo is removed: auth survives with its headers, nothing else
	// changes, and Fleet must drop the secret.
	remove := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
		testCheckServerSecretAbsent("auth", "api_key"),
		resource.TestCheckResourceAttr(testResourceName, "auth.headers.#", "1"),
	), "headers", "true", "label", "renamed")
	remove.ConfigStateChecks = writeOnlyStateChecks(pathAuthAPIKeyWO, pathAuthAPIKey)

	readd := writeOnlyStep(random, "api_key_wo", "key-two", "headers", "true", "label", "renamed")
	readd.Check = testCheckServerSecretPresent("auth", "api_key", &secretID)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		CheckDestroy:           checkResourceFleetAgentDownloadSourceDestroy,
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps: []resource.TestStep{
			create,
			writeOnlyUnchangedStep(random, "api_key_wo", "key-one", "headers", "true"),
			rename,
			rotate,
			writeOnlyUnchangedStep(random, "api_key_wo", "key-two", "headers", "true", "label", "renamed"),
			remove,
			writeOnlyUnchangedStep(random, "headers", "true", "label", "renamed"),
			readd,
			writeOnlyUnchangedStep(random, "api_key_wo", "key-two", "headers", "true", "label", "renamed"),
		},
	})
}

// TestAccResourceFleetAgentDownloadSource_WriteOnlySecretsIndependent rotates
// one write-only secret at a time and asserts the others keep their managed
// secret ids (tasks 8.1, 8.2), then switches the auth credential family.
func TestAccResourceFleetAgentDownloadSource_WriteOnlySecretsIndependent(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	var sslID, authID string

	base := []string{"certificate", "PEM-CERT", "username", "svc-account"}
	vars := func(kv ...string) []string { return append(append([]string{}, base...), kv...) }

	create := writeOnlyStep(random, vars("key_wo", "key-one", "password_wo", "pass-one")...)
	create.Check = resource.ComposeTestCheckFunc(
		testCheckServerSecretPresent("ssl", "key", &sslID),
		testCheckServerSecretPresent("auth", "password", &authID),
	)

	rotatePassword := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
		testCheckServerSecretUnchanged("ssl", "key", &sslID),
		testCheckServerSecretRotated("auth", "password", &authID),
	), vars("key_wo", "key-one", "password_wo", "pass-two")...)

	rotateKey := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
		testCheckServerSecretRotated("ssl", "key", &sslID),
		testCheckServerSecretUnchanged("auth", "password", &authID),
	), vars("key_wo", "key-two", "password_wo", "pass-two")...)

	// Switching the credential family drops the password secret and adds an
	// api_key secret while the ssl secret is referenced by id.
	toAPIKey := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
		testCheckServerSecretUnchanged("ssl", "key", &sslID),
		testCheckServerSecretAbsent("auth", "password"),
		testCheckServerSecretPresent("auth", "api_key", &authID),
	), "certificate", "PEM-CERT", "key_wo", "key-two", "api_key_wo", "api-key-one")
	toAPIKey.ConfigStateChecks = writeOnlyStateChecks(pathAuthPassword, pathAuthPasswordWO, pathAuthAPIKey, pathAuthAPIKeyWO, pathSSLKey, pathSSLKeyWO)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		CheckDestroy:           checkResourceFleetAgentDownloadSourceDestroy,
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps: []resource.TestStep{
			create,
			writeOnlyUnchangedStep(random, vars("key_wo", "key-one", "password_wo", "pass-one")...),
			rotatePassword,
			writeOnlyUnchangedStep(random, vars("key_wo", "key-one", "password_wo", "pass-two")...),
			rotateKey,
			writeOnlyUnchangedStep(random, vars("key_wo", "key-two", "password_wo", "pass-two")...),
			toAPIKey,
			writeOnlyUnchangedStep(random, "certificate", "PEM-CERT", "key_wo", "key-two", "api_key_wo", "api-key-one"),
		},
	})
}

// TestAccResourceFleetAgentDownloadSource_WriteOnlyRemoveReadd covers removal
// and re-add of write-only secrets (task 8.10). Removing an unchanged key_wo or
// password_wo with no other attribute change must still plan an update that
// clears the remote secret; re-adding the same value sends a literal again and
// records a fresh hash, so the follow-up plan is empty.
func TestAccResourceFleetAgentDownloadSource_WriteOnlyRemoveReadd(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	var sslID, authID string

	withBoth := []string{"certificate", "PEM-CERT", "key_wo", "key-one", "username", "svc-account", "password_wo", "pass-one", "headers", "true"}

	create := writeOnlyStep(random, withBoth...)
	create.Check = resource.ComposeTestCheckFunc(
		testCheckServerSecretPresent("ssl", "key", &sslID),
		testCheckServerSecretPresent("auth", "password", &authID),
	)

	removeKey := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
		testCheckServerSecretAbsent("ssl", "key"),
		testCheckServerSecretUnchanged("auth", "password", &authID),
		resource.TestCheckResourceAttr(testResourceName, "ssl.certificate", "PEM-CERT"),
	), "certificate", "PEM-CERT", "username", "svc-account", "password_wo", "pass-one", "headers", "true")
	removeKey.ConfigStateChecks = writeOnlyStateChecks(pathSSLKey, pathSSLKeyWO)

	readdKey := writeOnlyPlannedUpdateStep(random, testCheckServerSecretRotated("ssl", "key", &sslID), withBoth...)
	readdKey.ConfigStateChecks = writeOnlyStateChecks(pathSSLKey, pathSSLKeyWO)

	// Drop the username together with password_wo; headers keep auth present.
	removePassword := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
		testCheckServerSecretAbsent("auth", "password"),
		testCheckServerSecretUnchanged("ssl", "key", &sslID),
		resource.TestCheckNoResourceAttr(testResourceName, "auth.username"),
		resource.TestCheckResourceAttr(testResourceName, "auth.headers.#", "1"),
	), "certificate", "PEM-CERT", "key_wo", "key-one", "headers", "true")
	removePassword.ConfigStateChecks = writeOnlyStateChecks(pathAuthPassword, pathAuthPasswordWO)

	readdPassword := writeOnlyPlannedUpdateStep(random, testCheckServerSecretPresent("auth", "password", &authID), withBoth...)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		CheckDestroy:           checkResourceFleetAgentDownloadSourceDestroy,
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps: []resource.TestStep{
			create,
			removeKey,
			writeOnlyUnchangedStep(random, "certificate", "PEM-CERT", "username", "svc-account", "password_wo", "pass-one", "headers", "true"),
			readdKey,
			writeOnlyUnchangedStep(random, withBoth...),
			removePassword,
			writeOnlyUnchangedStep(random, "certificate", "PEM-CERT", "key_wo", "key-one", "headers", "true"),
			readdPassword,
			writeOnlyUnchangedStep(random, withBoth...),
		},
	})
}

// TestAccResourceFleetAgentDownloadSource_PlainToWriteOnly switches ssl.key,
// auth.password and auth.api_key between their plain and write-only forms. The
// plain forms are sent as managed-secret literals too, so Fleet holds a secret
// throughout, and the write-only value must never leak into the plain sibling.
func TestAccResourceFleetAgentDownloadSource_PlainToWriteOnly(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	var sslID, authID string

	plain := []string{"certificate", "PEM-CERT", "key", "plain-key", "username", "svc-account", "password", "plain-pass"}
	writeOnly := []string{"certificate", "PEM-CERT", "key_wo", "wo-key", "username", "svc-account", "password_wo", "wo-pass"}

	toPlain := func(kv []string) resource.TestStep {
		step := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(testResourceName, "ssl.key", "plain-key"),
			resource.TestCheckResourceAttr(testResourceName, "auth.password", "plain-pass"),
			testCheckServerSecretRotated("ssl", "key", &sslID),
			testCheckServerSecretRotated("auth", "password", &authID),
		), kv...)
		step.ConfigStateChecks = writeOnlyStateChecks(pathSSLKeyWO, pathAuthPasswordWO)
		return step
	}

	create := writeOnlyStep(random, plain...)
	create.Check = resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(testResourceName, "ssl.key", "plain-key"),
		resource.TestCheckResourceAttr(testResourceName, "auth.password", "plain-pass"),
		testCheckServerSecretPresent("ssl", "key", &sslID),
		testCheckServerSecretPresent("auth", "password", &authID),
	)

	toWriteOnly := writeOnlyPlannedUpdateStep(random, resource.ComposeTestCheckFunc(
		resource.TestCheckNoResourceAttr(testResourceName, "ssl.key"),
		resource.TestCheckNoResourceAttr(testResourceName, "auth.password"),
		testCheckServerSecretRotated("ssl", "key", &sslID),
		testCheckServerSecretRotated("auth", "password", &authID),
		testCheckServerNoPlainSecret("ssl", "key"),
	), writeOnly...)
	toWriteOnly.ConfigStateChecks = writeOnlyStateChecks(pathSSLKey, pathSSLKeyWO, pathAuthPassword, pathAuthPasswordWO)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		CheckDestroy:           checkResourceFleetAgentDownloadSourceDestroy,
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps: []resource.TestStep{
			create,
			toWriteOnly,
			writeOnlyUnchangedStep(random, writeOnly...),
			toPlain(plain),
			writeOnlyUnchangedStep(random, plain...),
		},
	})
}

// pemKey returns a PEM-shaped value well over bcrypt's 72-byte input limit
// whose first ~100 bytes are identical across variants.
func pemKey(variant string) string {
	var b strings.Builder
	b.WriteString("-----BEGIN PRIVATE KEY-----\n")
	for range 24 {
		b.WriteString(strings.Repeat("MIIEvQIBADANBgkqhkiG9w0BAQEFAASC", 2) + "\n")
	}
	b.WriteString(variant + "\n-----END PRIVATE KEY-----")
	return b.String()
}

// TestAccResourceFleetAgentDownloadSource_SSLKeyWriteOnlyPEM exercises the
// writeonlyhash 72-byte fix end to end (task 8.5): a PEM-sized key_wo creates,
// refreshes without drift, and a change confined to the tail of the value is
// still detected.
func TestAccResourceFleetAgentDownloadSource_SSLKeyWriteOnlyPEM(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceSsl, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	var secretID string
	keyA, keyB := pemKey("AAAA"), pemKey("BBBB")

	create := writeOnlyStep(random, "key_wo", keyA)
	create.Check = testCheckServerSecretPresent("ssl", "key", &secretID)
	create.ConfigStateChecks = writeOnlyStateChecks(pathSSLKeyWO, pathSSLKey)

	rotate := writeOnlyPlannedUpdateStep(random, testCheckServerSecretRotated("ssl", "key", &secretID), "key_wo", keyB)
	rotate.ConfigStateChecks = writeOnlyStateChecks(pathSSLKeyWO, pathSSLKey)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		CheckDestroy:           checkResourceFleetAgentDownloadSourceDestroy,
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps: []resource.TestStep{
			create,
			writeOnlyUnchangedStep(random, "key_wo", keyA),
			rotate,
			writeOnlyUnchangedStep(random, "key_wo", keyB),
		},
	})
}

// TestAccResourceFleetAgentDownloadSource_WriteOnlyValidation covers the
// configuration rejection matrix for the write-only attributes (tasks 8.1, 8.3,
// 8.8). These fail at validation, so no stack is needed.
func TestAccResourceFleetAgentDownloadSource_WriteOnlyValidation(t *testing.T) {
	combination := regexp.MustCompile(`(?s)Invalid Attribute Combination`)

	cases := []struct {
		name   string
		vars   []string
		expect *regexp.Regexp
	}{
		{"key_and_key_wo", []string{"key", "k", "key_wo", "k"}, combination},
		{"empty_key_wo", []string{"key_wo", "<empty>"}, regexp.MustCompile(`(?is)key_wo.*string length must be at least 1`)},
		{"username_without_password", []string{"username", "u"}, combination},
		{"password_without_username", []string{"password", "p"}, combination},
		{"password_wo_without_username", []string{"password_wo", "p"}, combination},
		{"password_and_password_wo", []string{"username", "u", "password", "p", "password_wo", "p"}, combination},
		{"api_key_and_api_key_wo", []string{"api_key", "k", "api_key_wo", "k"}, combination},
		{"password_wo_and_api_key", []string{"username", "u", "password_wo", "p", "api_key", "k"}, combination},
		{"password_and_api_key_wo", []string{"username", "u", "password", "p", "api_key_wo", "k"}, combination},
		{"password_wo_and_api_key_wo", []string{"username", "u", "password_wo", "p", "api_key_wo", "k"}, combination},
	}

	steps := make([]resource.TestStep, 0, len(cases))
	for _, tc := range cases {
		step := writeOnlyStep("validation", tc.vars...)
		step.PlanOnly = true
		step.ExpectError = tc.expect
		steps = append(steps, step)
	}

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps:                  steps,
	})
}

// TestAccResourceFleetAgentDownloadSource_WriteOnlyImport asserts the
// write-only attributes are null after import (task 8.7): they have no API
// representation. The managed secrets themselves are not hydrated into the plain
// siblings.
func TestAccResourceFleetAgentDownloadSource_WriteOnlyImport(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	vars := []string{"certificate", "PEM-CERT", "key_wo", "key-one", "username", "svc-account", "password_wo", "pass-one"}
	apiKeyVars := []string{"certificate", "PEM-CERT", "key_wo", "key-one", "api_key_wo", "api-key-one"}

	importStep := func(kv []string) resource.TestStep {
		step := writeOnlyStep(random, kv...)
		step.ResourceName = testResourceName
		step.ImportState = true
		step.ImportStateIdFunc = testFleetAgentDownloadSourceImportStateIDDefaultSpace
		step.ImportStateCheck = func(states []*terraform.InstanceState) error {
			if len(states) != 1 {
				return fmt.Errorf("expected 1 imported state, got %d", len(states))
			}
			for k, v := range states[0].Attributes {
				for _, wo := range []string{"ssl.key_wo", "auth.password_wo", "auth.api_key_wo", "ssl.key", "auth.password", "auth.api_key"} {
					if k == wo {
						return fmt.Errorf("expected %s to be null after import, got %q", wo, v)
					}
				}
			}
			return nil
		}
		return step
	}

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		CheckDestroy:           checkResourceFleetAgentDownloadSourceDestroy,
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps: []resource.TestStep{
			writeOnlyStep(random, vars...),
			importStep(vars),
			writeOnlyStep(random, apiKeyVars...),
			importStep(apiKeyVars),
		},
	})
}

// TestAccResourceFleetAgentDownloadSource_WriteOnlyDeleteRecreate destroys and
// re-creates a resource with write-only secrets under the same configuration
// (task 8.6). Delete clears the resource's secret-hash private-state keys, so
// the re-create must send the literals again rather than trust a stale hash.
func TestAccResourceFleetAgentDownloadSource_WriteOnlyDeleteRecreate(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	vars := []string{"key_wo", "key-one", "username", "svc-account", "password_wo", "pass-one"}

	present := resource.ComposeTestCheckFunc(
		testCheckServerSecretPresent("ssl", "key", nil),
		testCheckServerSecretPresent("auth", "password", nil),
	)
	create := writeOnlyStep(random, vars...)
	create.Check = present
	recreate := writeOnlyStep(random, vars...)
	recreate.Check = present

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		CheckDestroy:           checkResourceFleetAgentDownloadSourceDestroy,
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps: []resource.TestStep{
			create,
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("destroyed"),
				Check: func(s *terraform.State) error {
					if _, ok := s.RootModule().Resources[testResourceName]; ok {
						return fmt.Errorf("expected %s to be destroyed", testResourceName)
					}
					return nil
				},
			},
			recreate,
			writeOnlyUnchangedStep(random, vars...),
		},
	})
}

// noSecretStorageProxy fronts Kibana and imitates a Fleet Server that cannot
// store secrets: responses never carry secret ids, and Kibana merges the
// submitted secret literals into the plain ssl.key / auth.password /
// auth.api_key fields instead. It records every secret value the resource sends.
type noSecretStorageProxy struct {
	server *httptest.Server

	mu       sync.Mutex
	literals map[string]string // "ssl.key", "auth.password", "auth.api_key" from the last write
	writes   []map[string]any  // the secrets object of each POST/PUT
}

func newNoSecretStorageProxy(t *testing.T, kibanaEndpoint string) *noSecretStorageProxy {
	t.Helper()

	target, err := url.Parse(kibanaEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	p := &noSecretStorageProxy{literals: map[string]string{}}
	rp := &httputil.ReverseProxy{}
	rp.Rewrite = func(pr *httputil.ProxyRequest) {
		pr.SetURL(target)
		pr.Out.Host = target.Host
		pr.Out.Header.Del("Accept-Encoding")
		if pr.Out.Method == http.MethodPost || pr.Out.Method == http.MethodPut {
			p.recordWrite(pr.Out)
		}
	}
	rp.ModifyResponse = p.rewriteResponse
	p.server = httptest.NewServer(rp)
	t.Cleanup(p.server.Close)
	return p
}

func isDownloadSourcePath(r *http.Request) bool {
	return strings.Contains(r.URL.Path, "/api/fleet/agent_download_sources")
}

func (p *noSecretStorageProxy) recordWrite(r *http.Request) {
	if !isDownloadSourcePath(r) || r.Body == nil {
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	r.Body = io.NopCloser(strings.NewReader(string(raw)))

	var body struct {
		Secrets map[string]map[string]any `json:"secrets"`
	}
	_ = json.Unmarshal(raw, &body)

	p.mu.Lock()
	defer p.mu.Unlock()
	p.literals = map[string]string{}
	secrets := map[string]any{}
	for parent, fields := range body.Secrets {
		for field, v := range fields {
			secrets[parent+"."+field] = v
			if str, ok := v.(string); ok {
				p.literals[parent+"."+field] = str
			}
		}
	}
	p.writes = append(p.writes, secrets)
}

func (p *noSecretStorageProxy) rewriteResponse(resp *http.Response) error {
	if !isDownloadSourcePath(resp.Request) || resp.StatusCode != http.StatusOK ||
		!strings.Contains(resp.Header.Get("Content-Type"), "json") {
		return nil
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	var body map[string]any
	if json.Unmarshal(raw, &body) == nil {
		if item, ok := body["item"].(map[string]any); ok {
			delete(item, "secrets")
			p.mu.Lock()
			for key, literal := range p.literals {
				parent, field, _ := strings.Cut(key, ".")
				obj, _ := item[parent].(map[string]any)
				if obj == nil {
					obj = map[string]any{}
				}
				obj[field] = literal
				item[parent] = obj
			}
			p.mu.Unlock()
			raw, _ = json.Marshal(body)
		}
	}
	resp.Body = io.NopCloser(strings.NewReader(string(raw)))
	resp.ContentLength = int64(len(raw))
	resp.Header.Set("Content-Length", fmt.Sprint(len(raw)))
	return nil
}

// secretsSent returns the secret value sent for key (for example "ssl.key") by
// each write, in order; entries are strings (literals) or maps ({id} references).
func (p *noSecretStorageProxy) secretsSent(key string) []any {
	p.mu.Lock()
	defer p.mu.Unlock()
	var out []any
	for _, w := range p.writes {
		if v, ok := w[key]; ok {
			out = append(out, v)
		}
	}
	return out
}

// TestAccResourceFleetAgentDownloadSource_WriteOnlySecretStorageUnavailable
// covers a Fleet Server that cannot store secrets (task 8.4). The acceptance
// stack always has storage, so a proxy in front of Kibana strips secret ids and
// echoes the submitted literals into the plain fields, as Kibana does without
// storage. The resource must resend literals rather than {id} references, must
// not error, and must not copy the echoed write-only literal into ssl.key.
func TestAccResourceFleetAgentDownloadSource_WriteOnlySecretStorageUnavailable(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetAgentDownloadSourceAuth, versionutils.FlavorAny)
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests require TF_ACC")
	}

	proxy := newNoSecretStorageProxy(t, os.Getenv("KIBANA_ENDPOINT"))
	t.Setenv("KIBANA_ENDPOINT", proxy.server.URL)

	random := sdkacctest.RandString(8)
	vars := []string{"key_wo", "key-one", "username", "svc-account", "password_wo", "pass-one"}
	wantNullState := writeOnlyStateChecks(pathSSLKeyWO, pathSSLKey, pathAuthPasswordWO, pathAuthPassword)

	assertSent := func(key string, want ...string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			got := proxy.secretsSent(key)
			if len(got) != len(want) {
				return fmt.Errorf("expected %s to be sent %d times as %v, got %v", key, len(want), want, got)
			}
			for i, v := range got {
				if v != want[i] {
					return fmt.Errorf("write %d: expected %s literal %q, got %v", i, key, want[i], v)
				}
			}
			return nil
		}
	}
	noPlainEcho := resource.ComposeTestCheckFunc(
		resource.TestCheckNoResourceAttr(testResourceName, "ssl.key"),
		resource.TestCheckNoResourceAttr(testResourceName, "auth.password"),
	)

	create := writeOnlyStep(random, vars...)
	create.ConfigStateChecks = wantNullState
	create.Check = resource.ComposeTestCheckFunc(noPlainEcho, assertSent("ssl.key", "key-one"), assertSent("auth.password", "pass-one"))

	// An unrelated update has no secret id to reference, so literals are resent.
	rename := writeOnlyStep(random, append(append([]string{}, vars...), "label", "renamed")...)
	rename.ConfigStateChecks = wantNullState
	rename.Check = resource.ComposeTestCheckFunc(noPlainEcho, assertSent("ssl.key", "key-one", "key-one"), assertSent("auth.password", "pass-one", "pass-one"))

	rotate := writeOnlyStep(random, "key_wo", "key-two", "username", "svc-account", "password_wo", "pass-one", "label", "renamed")
	rotate.ConfigStateChecks = wantNullState
	rotate.Check = resource.ComposeTestCheckFunc(noPlainEcho, assertSent("ssl.key", "key-one", "key-one", "key-two"))

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: writeOnlyTerraformVersionChecks(),
		Steps: []resource.TestStep{
			create,
			writeOnlyUnchangedStep(random, vars...),
			rename,
			rotate,
		},
	})
}
