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

package connectors_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/connectors"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const connectorResourceName = "elasticstack_kibana_action_connector.test"

// TestAccResourceKibanaConnectorInvalidConnectorID covers the negative path of
// the `connector_id` attribute's IsUUID() validator: a non-UUID value must be
// rejected at plan time, before any API call is made.
func TestAccResourceKibanaConnectorInvalidConnectorID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid"),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)Invalid UUID`),
			},
		},
	})
}

// TestAccResourceKibanaConnectorTypeChangeRequiresReplace verifies that
// changing `connector_type_id` on an existing connector triggers a
// destroy-then-create plan action, matching the attribute's documented
// RequiresReplace plan modifier.
func TestAccResourceKibanaConnectorTypeChangeRequiresReplace(t *testing.T) {
	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceKibanaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step1"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".index"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step2"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							connectorResourceName,
							plancheck.ResourceActionDestroyBeforeCreate,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".webhook"),
				),
			},
		},
	})
}

// TestAccResourceKibanaConnectorIDChangeRequiresReplace verifies that changing
// a user-supplied `connector_id` on an existing connector triggers a
// destroy-then-create plan action, matching the attribute's documented
// RequiresReplace plan modifier.
func TestAccResourceKibanaConnectorIDChangeRequiresReplace(t *testing.T) {
	minSupportedVersion := connectors.MinVersionSupportingPreconfiguredIDs

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	connectorIDOne := uuid.NewString()
	connectorIDTwo := uuid.NewString()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceKibanaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step1"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
					"connector_id":   config.StringVariable(connectorIDOne),
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".index"),
					resource.TestCheckResourceAttr(connectorResourceName, "connector_id", connectorIDOne),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step2"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
					"connector_id":   config.StringVariable(connectorIDTwo),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							connectorResourceName,
							plancheck.ResourceActionDestroyBeforeCreate,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".index"),
					resource.TestCheckResourceAttr(connectorResourceName, "connector_id", connectorIDTwo),
				),
			},
		},
	})
}

// TestAccResourceKibanaConnectorSpaceIDChangeRequiresReplace verifies that
// changing `space_id` on an existing connector triggers a destroy-then-create
// plan action, matching kbschema.ResourceSpaceIDAttributeRequiresReplaceOnly's
// RequiresReplace plan modifier.
func TestAccResourceKibanaConnectorSpaceIDChangeRequiresReplace(t *testing.T) {
	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	spaceIDOne := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	spaceIDTwo := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	vars := config.Variables{
		"connector_name": config.StringVariable(connectorName),
		"space_id_a":     config.StringVariable(spaceIDOne),
		"space_id_b":     config.StringVariable(spaceIDTwo),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceKibanaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step1"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".index"),
					resource.TestCheckResourceAttr(connectorResourceName, "space_id", spaceIDOne),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step2"),
				ConfigVariables:          vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							connectorResourceName,
							plancheck.ResourceActionDestroyBeforeCreate,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".index"),
					resource.TestCheckResourceAttr(connectorResourceName, "space_id", spaceIDTwo),
				),
			},
		},
	})
}
