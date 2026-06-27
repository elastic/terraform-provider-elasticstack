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

package tag_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	minKibanaTagAccTestVersion = version.Must(version.NewVersion("9.5.0-SNAPSHOT"))

	tagResourceAddr    = "elasticstack_kibana_tag.test"
	tagsDataSourceAddr = "data.elasticstack_kibana_tags.test"

	accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"
)

func skipKibanaTagUnsupported() func() (bool, error) {
	return versionutils.CheckIfVersionIsUnsupported(minKibanaTagAccTestVersion)
}

func TestAccResourceKibanaTag_basic(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tagResourceAddr, "id"),
					resource.TestCheckResourceAttrSet(tagResourceAddr, "tag_id"),
					testAccCheckCompositeIDFormat(tagResourceAddr, clients.DefaultSpaceID),
					resource.TestCheckResourceAttr(tagResourceAddr, "space_id", clients.DefaultSpaceID),
					resource.TestCheckResourceAttr(tagResourceAddr, "name", fmt.Sprintf("tf-acc-tag-%s", suffix)),
					resource.TestCheckResourceAttr(tagResourceAddr, "color", "#FF0000"),
					resource.TestCheckResourceAttr(tagResourceAddr, "description", "initial description"),
					resource.TestCheckResourceAttrSet(tagResourceAddr, "created_at"),
					resource.TestCheckResourceAttrSet(tagResourceAddr, "updated_at"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tagResourceAddr, "name", fmt.Sprintf("tf-acc-tag-updated-%s", suffix)),
					resource.TestCheckResourceAttr(tagResourceAddr, "color", "#FF0000"),
					resource.TestCheckNoResourceAttr(tagResourceAddr, "description"),
				),
			},
		},
	})
}

func TestAccResourceKibanaTag_noColor(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}
	var initialColor string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tagResourceAddr, "color"),
					captureTagColor(tagResourceAddr, &initialColor),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tagResourceAddr, "name", fmt.Sprintf("tf-acc-tag-nocolor-updated-%s", suffix)),
					testAccCheckTagColorEquals(tagResourceAddr, &initialColor),
				),
			},
		},
	})
}

func TestAccResourceKibanaTag_withTagID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	tagID := uuid.NewString()
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
		"tag_id": config.StringVariable(tagID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tagResourceAddr, "tag_id", tagID),
					resource.TestCheckResourceAttr(tagResourceAddr, "id", fmt.Sprintf("%s/%s", clients.DefaultSpaceID, tagID)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ResourceName:             tagResourceAddr,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateId:            fmt.Sprintf("%s/%s", clients.DefaultSpaceID, tagID),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
			},
		},
	})
}

func TestAccResourceKibanaTag_duplicateTagID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	tagID := uuid.NewString()
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
		"tag_id": config.StringVariable(tagID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("duplicate"),
				ConfigVariables:          vars,
				ExpectError:              regexp.MustCompile("terraform import"),
			},
		},
	})
}

func TestAccResourceKibanaTag_duplicateName(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("duplicate"),
				ConfigVariables:          vars,
				ExpectError:              regexp.MustCompile("(?i)(conflict|duplicate|already exists)"),
			},
		},
	})
}

func TestAccResourceKibanaTag_import(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ResourceName:             tagResourceAddr,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[tagResourceAddr]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", tagResourceAddr)
					}
					return rs.Primary.ID, nil
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("import"),
				ConfigVariables: vars,
			},
		},
	})
}

func TestAccResourceKibanaTag_colorUpdate(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tagResourceAddr, "color", "#FF0000"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tagResourceAddr, "color", "#00FF00"),
				),
			},
		},
	})
}

func TestAccResourceKibanaTag_nonDefaultSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{
		"suffix":   config.StringVariable(suffix),
		"space_id": config.StringVariable(spaceID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tagResourceAddr, "space_id", spaceID),
					resource.TestMatchResourceAttr(tagResourceAddr, "id", regexp.MustCompile("^"+regexp.QuoteMeta(spaceID)+"/")),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tagResourceAddr, "name", fmt.Sprintf("tf-acc-tag-space-updated-%s", suffix)),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaTags_basic(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tagsDataSourceAddr, "tags.#", "2"),
					testAccCheckDataSourceContainsTagName(tagsDataSourceAddr, fmt.Sprintf("tf-acc-tag-a-%s", suffix)),
					testAccCheckDataSourceContainsTagName(tagsDataSourceAddr, fmt.Sprintf("tf-acc-tag-b-%s", suffix)),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaTags_query(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	queryName := fmt.Sprintf("tf-acc-tag-query-%s", suffix)
	vars := config.Variables{
		"suffix":     config.StringVariable(suffix),
		"query_name": config.StringVariable(queryName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tagsDataSourceAddr, "tags.#", "1"),
					testAccCheckDataSourceContainsTagName(tagsDataSourceAddr, queryName),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaTags_nonDefaultSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaTagAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{
		"suffix":   config.StringVariable(suffix),
		"space_id": config.StringVariable(spaceID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkTagDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKibanaTagUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tagsDataSourceAddr, "space_id", spaceID),
					resource.TestCheckResourceAttr(tagsDataSourceAddr, "tags.#", "1"),
					testAccCheckDataSourceContainsTagName(tagsDataSourceAddr, fmt.Sprintf("tf-acc-tag-space-%s", suffix)),
				),
			},
		},
	})
}

var checkTagDestroy = checks.KibanaResourceDestroyCheckCompositeID(
	"elasticstack_kibana_tag",
	func(ctx context.Context, client *kibanaoapi.Client, spaceID, tagID string) (bool, error) {
		detail, diags := kibanaoapi.GetTag(ctx, client, spaceID, tagID)
		if diags.HasError() {
			return false, fmt.Errorf("failed to get tag %q in space %q: %v", tagID, spaceID, diags)
		}
		return detail != nil, nil
	},
)

func testAccCheckCompositeIDFormat(resourceAddr, spaceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceAddr)
		}

		tagID, ok := rs.Primary.Attributes["tag_id"]
		if !ok || tagID == "" {
			return fmt.Errorf("resource %q has no tag_id in state", resourceAddr)
		}

		expectedID := fmt.Sprintf("%s/%s", spaceID, tagID)
		if rs.Primary.ID != expectedID {
			return fmt.Errorf("expected id %q, got %q", expectedID, rs.Primary.ID)
		}
		return nil
	}
}

func captureTagColor(addr string, color *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("resource %q not found in state", addr)
		}
		value, ok := rs.Primary.Attributes["color"]
		if !ok || value == "" {
			return fmt.Errorf("resource %q has no color in state", addr)
		}
		*color = value
		return nil
	}
}

func testAccCheckTagColorEquals(addr string, expectedColor *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if expectedColor == nil || *expectedColor == "" {
			return fmt.Errorf("no expected color captured for %q", addr)
		}
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("resource %q not found in state", addr)
		}
		if got := rs.Primary.Attributes["color"]; got != *expectedColor {
			return fmt.Errorf("expected color %q, got %q", *expectedColor, got)
		}
		return nil
	}
}

func testAccCheckDataSourceContainsTagName(dataSourceAddr, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceAddr]
		if !ok {
			return fmt.Errorf("data source %q not found in state", dataSourceAddr)
		}

		countStr, ok := rs.Primary.Attributes["tags.#"]
		if !ok {
			return fmt.Errorf("data source %q has no tags list", dataSourceAddr)
		}

		var count int
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return fmt.Errorf("invalid tags count %q: %w", countStr, err)
		}

		for i := range count {
			attrName := fmt.Sprintf("tags.%d.name", i)
			if rs.Primary.Attributes[attrName] == name {
				return nil
			}
		}

		return fmt.Errorf("data source %q does not contain tag named %q", dataSourceAddr, name)
	}
}
