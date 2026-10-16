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
	"regexp"
	"strings"
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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
				ImportStateVerifyIgnore: []string{"space_ids"},
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
