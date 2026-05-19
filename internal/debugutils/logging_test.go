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
	"testing"
)

func TestIsDebugOrHigher(t *testing.T) {
	tests := []struct {
		name       string
		tfLog      string
		unsetTFLog bool
		want       bool
	}{
		{name: "TF_LOG unset", unsetTFLog: true, want: false},
		{name: "TF_LOG empty", tfLog: "", want: false},
		{name: "TF_LOG=DEBUG", tfLog: "DEBUG", want: true},
		{name: "TF_LOG=debug", tfLog: "debug", want: true},
		{name: "TF_LOG=TRACE", tfLog: "TRACE", want: true},
		{name: "TF_LOG=trace", tfLog: "trace", want: true},
		{name: "TF_LOG=INFO", tfLog: "INFO", want: false},
		{name: "TF_LOG=WARN", tfLog: "WARN", want: false},
		{name: "TF_LOG=ERROR", tfLog: "ERROR", want: false},
		{name: "TF_LOG=NOTAREALLEVEL", tfLog: "NOTAREALLEVEL", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.unsetTFLog {
				t.Setenv(envLog, "")
			} else {
				t.Setenv(envLog, tt.tfLog)
			}
			if got := IsDebugOrHigher(); got != tt.want {
				t.Errorf("IsDebugOrHigher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSensitiveInSchema(t *testing.T) {
	tests := []struct {
		name  string
		tfLog string
		tfAcc string
		want  bool
	}{
		{name: "TF_LOG unset and TF_ACC unset", tfLog: "", tfAcc: "", want: true},
		{name: "TF_LOG=DEBUG and TF_ACC unset", tfLog: "DEBUG", tfAcc: "", want: false},
		{name: "TF_LOG unset and TF_ACC=1", tfLog: "", tfAcc: "1", want: false},
		{name: "TF_LOG=DEBUG and TF_ACC=1", tfLog: "DEBUG", tfAcc: "1", want: false},
		{name: "TF_LOG=INFO and TF_ACC=0", tfLog: "INFO", tfAcc: "0", want: true},
		{name: "TF_LOG=INFO and TF_ACC empty", tfLog: "INFO", tfAcc: "", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envLog, tt.tfLog)
			t.Setenv("TF_ACC", tt.tfAcc)
			if got := IsSensitiveInSchema(); got != tt.want {
				t.Errorf("IsSensitiveInSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}
