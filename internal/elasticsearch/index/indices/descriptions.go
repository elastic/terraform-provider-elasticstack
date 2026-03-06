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

package indices

import _ "embed"

//go:embed descriptions/target.md
var targetDescription string

//go:embed descriptions/codec.md
var codecDescription string

//go:embed descriptions/shard_check_on_startup.md
var shardCheckOnStartupDescription string

//go:embed descriptions/final_pipeline.md
var finalPipelineDescription string

//go:embed descriptions/indexing_slowlog_source.md
var indexingSlowlogSourceDescription string

//go:embed descriptions/deletion_protection.md
var deletionProtectionDescription string

//go:embed descriptions/wait_for_active_shards.md
var waitForActiveShardsDescription string

//go:embed descriptions/master_timeout.md
var masterTimeoutDescription string

//go:embed descriptions/mappings.md
var mappingsDescription string
