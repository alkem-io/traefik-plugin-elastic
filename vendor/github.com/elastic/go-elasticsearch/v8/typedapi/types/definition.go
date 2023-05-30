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

// Definition type.
//
// https://github.com/elastic/elasticsearch-specification/blob/363111664e81786557afe06e68221018847b3676/specification/ml/put_trained_model/types.ts#L24-L29
type Definition struct {
	// Preprocessors Collection of preprocessors
	Preprocessors []Preprocessor `json:"preprocessors,omitempty"`
	// TrainedModel The definition of the trained model.
	TrainedModel TrainedModel `json:"trained_model"`
}

// NewDefinition returns a Definition.
func NewDefinition() *Definition {
	r := &Definition{}

	return r
}
