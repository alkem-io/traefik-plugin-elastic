// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Code generated from the elasticsearch-specification DO NOT EDIT.
// https://github.com/elastic/elasticsearch-specification/tree/363111664e81786557afe06e68221018847b3676

package types

import (
	"bytes"
	"errors"
	"io"

	"encoding/json"
)

// WhitespaceAnalyzer type.
//
// https://github.com/elastic/elasticsearch-specification/blob/363111664e81786557afe06e68221018847b3676/specification/_types/analysis/analyzers.ts#L108-L111
type WhitespaceAnalyzer struct {
	Type    string  `json:"type,omitempty"`
	Version *string `json:"version,omitempty"`
}

func (s *WhitespaceAnalyzer) UnmarshalJSON(data []byte) error {

	dec := json.NewDecoder(bytes.NewReader(data))

	for {
		t, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch t {

		case "type":
			if err := dec.Decode(&s.Type); err != nil {
				return err
			}

		case "version":
			if err := dec.Decode(&s.Version); err != nil {
				return err
			}

		}
	}
	return nil
}

// NewWhitespaceAnalyzer returns a WhitespaceAnalyzer.
func NewWhitespaceAnalyzer() *WhitespaceAnalyzer {
	r := &WhitespaceAnalyzer{}

	r.Type = "whitespace"

	return r
}
