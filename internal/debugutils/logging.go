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

package debugutils

import (
	"os"
	"strings"
)

const envLog = "TF_LOG"

const (
	logLevelTrace = "TRACE"
	logLevelDebug = "DEBUG"
	logLevelInfo  = "INFO"
	logLevelWarn  = "WARN"
	logLevelError = "ERROR"
)

var validLogLevels = []string{logLevelTrace, logLevelDebug, logLevelInfo, logLevelWarn, logLevelError}

func logLevel() string {
	envLevel := os.Getenv(envLog)
	if envLevel == "" {
		return ""
	}

	for _, l := range validLogLevels {
		if strings.EqualFold(envLevel, l) {
			return strings.ToUpper(envLevel)
		}
	}

	// Mirror terraform-plugin-sdk/v2/helper/logging: invalid TF_LOG defaults to TRACE.
	return logLevelTrace
}

// IsDebugOrHigher reports whether TF_LOG is set to DEBUG or TRACE (case-insensitive).
// Unrecognized non-empty TF_LOG values are treated as TRACE, matching
// terraform-plugin-sdk/v2/helper/logging.IsDebugOrHigher SDK compatibility behavior.
func IsDebugOrHigher() bool {
	level := logLevel()
	return level == logLevelDebug || level == logLevelTrace
}

// IsSensitiveInSchema reports whether integration-policy variable attributes should be
// marked sensitive in the resource schema (masked unless debug logging or acceptance tests).
func IsSensitiveInSchema() bool {
	return !IsDebugOrHigher() && os.Getenv("TF_ACC") != "1"
}
