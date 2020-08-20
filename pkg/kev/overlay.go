/**
 * Copyright 2020 Appvia Ltd <info@appvia.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kev

import (
	"encoding/json"
	"io"

	"github.com/r3labs/diff"
)

// MarshalYAML makes Services implement yaml.Marshaller
func (s Services) MarshalYAML() (interface{}, error) {
	services := map[string]ServiceConfig{}
	for _, service := range s {
		services[service.Name] = service
	}
	return services, nil
}

// MarshalJSON makes Services implement json.Marshaler
func (s Services) MarshalJSON() ([]byte, error) {
	data, err := s.MarshalYAML()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "  ")
}

// GetLabels gets a service's labels
func (sc ServiceConfig) GetLabels() map[string]string {
	return sc.Labels
}

func (o *composeOverlay) diff(other *composeOverlay) (changeset, error) {
	d, _ := diff.NewDiffer()
	clog, err := d.Diff(other, o)
	if err != nil {
		return changeset{}, err
	}
	return newChangeset(clog)
}

func (o *composeOverlay) patch(cset changeset, reporter io.Writer) {
	cset.applyVersionPatchesIfAny(o, reporter)
	cset.applyServicesPatchesIfAny(o, reporter)
	cset.applyVolumesPatchesIfAny(o, reporter)
}