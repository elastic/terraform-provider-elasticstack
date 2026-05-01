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

package systemuser_test

import (
	_ "embed"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"golang.org/x/crypto/bcrypt"
)

const remoteMonitoringUser = "remote_monitoring_user"
const systemUserResourceName = "elasticstack_elasticsearch_security_system_user.remote_monitoring_user"

func TestAccResourceSecuritySystemUser(t *testing.T) {
	password1 := "new_password_1"
	password2 := "new_password_2"
	password3 := "new_password_3"
	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(password3), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password for test: %s", err)
	}
	passwordHash := string(passwordHashBytes)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(systemUserResourceName, "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr(systemUserResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(systemUserResourceName, "id"),
					resource.TestCheckNoResourceAttr(systemUserResourceName, "password"),
					resource.TestCheckNoResourceAttr(systemUserResourceName, "password_hash"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
					"password": config.StringVariable(password1),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(systemUserResourceName, "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr(systemUserResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(systemUserResourceName, "id"),
					resource.TestCheckResourceAttrSet(systemUserResourceName, "password"),
					checks.CheckUserCanAuthenticate(remoteMonitoringUser, password1),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(systemUserResourceName, "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr(systemUserResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(systemUserResourceName, "id"),
					resource.TestCheckNoResourceAttr(systemUserResourceName, "password"),
					resource.TestCheckNoResourceAttr(systemUserResourceName, "password_hash"),
					checks.CheckUserCanAuthenticate(remoteMonitoringUser, password1),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("disabled"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(systemUserResourceName, "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr(systemUserResourceName, "enabled", "false"),
					resource.TestCheckResourceAttrSet(systemUserResourceName, "id"),
					resource.TestCheckNoResourceAttr(systemUserResourceName, "password"),
					resource.TestCheckNoResourceAttr(systemUserResourceName, "password_hash"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("reenabled"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
					"password": config.StringVariable(password2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(systemUserResourceName, "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr(systemUserResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(systemUserResourceName, "id"),
					resource.TestCheckResourceAttrSet(systemUserResourceName, "password"),
					checks.CheckUserCanAuthenticate(remoteMonitoringUser, password2),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("password_hash"),
				ConfigVariables: config.Variables{
					"username":      config.StringVariable(remoteMonitoringUser),
					"password_hash": config.StringVariable(passwordHash),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(systemUserResourceName, "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr(systemUserResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(systemUserResourceName, "id"),
					resource.TestCheckNoResourceAttr(systemUserResourceName, "password"),
					resource.TestCheckResourceAttr(systemUserResourceName, "password_hash", passwordHash),
					checks.CheckUserCanAuthenticate(remoteMonitoringUser, password3),
				),
			},
		},
	})
}

func TestAccResourceSecuritySystemUserNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"username": config.StringVariable("not_system_user"),
					"password": config.StringVariable("new_password"),
				},
				ExpectError: regexp.MustCompile(`System user "not_system_user" not found`),
			},
		},
	})
}

//go:embed testdata/TestAccResourceSecuritySystemUserFromSDK/main.tf
var sdkCreateTestConfig string

func TestAccResourceSecuritySystemUserFromSDK(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create the system user with the last provider version where the system user resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.15",
					},
				},
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
				},
				Config: sdkCreateTestConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(systemUserResourceName, "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr(systemUserResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(systemUserResourceName, "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(systemUserResourceName, "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr(systemUserResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(systemUserResourceName, "id"),
				),
			},
		},
	})
}
