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

import "encoding/json"

// MarshalYAML makes Services implement yaml.Marshaller.
func (s Services) MarshalYAML() (interface{}, error) {
	services := map[string]ServiceConfig{}
	for _, service := range s {
		services[service.Name] = service
	}
	return services, nil
}

// MarshalJSON makes Services implement json.Marshaler.
func (s Services) MarshalJSON() ([]byte, error) {
	data, err := s.MarshalYAML()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "  ")
}

// Map converts services to a map.
func (s Services) Map() map[string]ServiceConfig {
	out := map[string]ServiceConfig{}
	for _, service := range s {
		out[service.Name] = service
	}
	return out
}

// Set converts services to a set.
func (s Services) Set() map[string]bool {
	out := map[string]bool{}
	for _, service := range s {
		out[service.Name] = true
	}
	return out
}

// GetLabels gets a service's labels
func (sc ServiceConfig) GetLabels() map[string]string {
	return sc.Labels
}

// minusEnvVars returns a copy of the ServiceConfig with blank env vars
func (sc ServiceConfig) minusEnvVars() ServiceConfig {
	return ServiceConfig{
		Name:        sc.Name,
		Labels:      sc.Labels,
		Environment: map[string]*string{},
	}
}

// condenseLabels returns a copy of the ServiceConfig with only condensed base service labels
func (sc ServiceConfig) condenseLabels(labels []string) ServiceConfig {
	for key := range sc.GetLabels() {
		if !contains(labels, key) {
			delete(sc.Labels, key)
		}
	}

	return ServiceConfig{
		Name:        sc.Name,
		Labels:      sc.Labels,
		Environment: sc.Environment,
	}
}