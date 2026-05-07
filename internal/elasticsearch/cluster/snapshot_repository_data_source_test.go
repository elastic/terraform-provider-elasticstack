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

package cluster_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSnapRepoMissing(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "id"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "type"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "url.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "gcs.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "s3.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "azure.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "hdfs.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceSnapRepoFs(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "type", "fs"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "gcs.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "url.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "s3.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "azure.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "hdfs.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.location", "/tmp"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.compress", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_restore_bytes_per_sec", "10mb"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.chunk_size", "1gb"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_snapshot_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_number_of_snapshots", "50"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.readonly", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceSnapRepoURL(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "type", "url"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "s3.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "fs.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "gcs.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "azure.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "hdfs.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.url", "file:/tmp"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.http_max_retries", "3"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.http_socket_timeout", "30s"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.compress", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.max_snapshot_bytes_per_sec", "40mb"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.max_restore_bytes_per_sec", "10mb"),

					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.max_number_of_snapshots", "500"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.chunk_size", "1gb"),
				),
			},
		},
	})
}
