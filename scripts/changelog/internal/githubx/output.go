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

package githubx

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

// AppendGitHubOutput appends name=value pairs to path (the file named by
// GITHUB_OUTPUT). Single-line values use "name=value\n". Values containing
// newlines use GitHub Actions heredoc framing with a random delimiter token.
func AppendGitHubOutput(path, name, value string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("missing GITHUB_OUTPUT")
	}
	if name == "" {
		return errors.New("output name must be non-empty")
	}

	var fragment string
	if strings.ContainsAny(value, "\r\n") {
		delim, err := randomDelimiterToken()
		if err != nil {
			return err
		}
		var b strings.Builder
		fmt.Fprintf(&b, "%s<<%s\n", name, delim)
		b.WriteString(value)
		if len(value) == 0 || !strings.HasSuffix(value, "\n") {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "%s\n", delim)
		fragment = b.String()
	} else {
		fragment = fmt.Sprintf("%s=%s\n", name, value)
	}

	return appendToOutputFile(path, fragment)
}

func randomDelimiterToken() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate output delimiter: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}

func appendToOutputFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return fmt.Errorf("write failed and close failed: %w", errors.Join(err, closeErr))
		}
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}
