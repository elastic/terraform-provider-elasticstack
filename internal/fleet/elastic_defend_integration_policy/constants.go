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

package elasticdefendintegrationpolicy

// Terraform schema attribute and Elastic Defend policy JSON keys reused
// across the policy schema, mapping helpers, and default-value builders.
const (
	attrMessage             = "message"
	attrEnabled             = "enabled"
	attrMode                = "mode"
	attrBlocklist           = "blocklist"
	attrNetwork             = "network"
	attrFile                = "file"
	attrEvents              = "events"
	attrLogging             = "logging"
	attrMalware             = "malware"
	attrMemoryProtection    = "memory_protection"
	attrBehaviorProtection  = "behavior_protection"
	attrCredentialHardening = "credential_hardening"
	attrPreset              = "preset"
	attrProcess             = "process"
	attrOnWriteScan         = "on_write_scan"
	attrNotifyUser          = "notify_user"
	attrSupported           = "supported"
	attrReputationService   = "reputation_service"
	attrRansomware          = "ransomware"
	attrPopup               = "popup"
	attrValue               = "value"
)

// Description strings reused by attribute schema definitions for repeated
// boolean event-collection toggles and validator help text.
const (
	descCollectProcessEvents = "Collect process events."
	descCollectNetworkEvents = "Collect network events."
	descCollectFileEvents    = "Collect file events."
	descBlocklistEnabled     = "Whether blocklist is enabled."
	descMalwareMode          = "Malware protection mode. Valid values: `\"off\"`, `\"detect\"`, `\"prevent\"`."
	descLoggingFileLevel     = "Log level for file logging. Valid values: `\"info\"`, `\"debug\"`, `\"warning\"`, `\"error\"`, `\"critical\"`."
)
