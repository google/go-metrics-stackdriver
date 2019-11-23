// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package vault provides helper functions to improve the go-metrics to stackdriver metric
// conversions specific to HashiCorp Vault.
package vault

import "github.com/armon/go-metrics"

// Extractor extracts known patterns from the key into metrics.Label for better metric grouping
// and to help avoid the limit of 500 custom metric descriptors per project
// (https://cloud.google.com/monitoring/quotas).
func Extractor(key []string) ([]string, []metrics.Label) {
	if len(key) == 2 {
		if key[0] == "database" {
			return key[:1], []metrics.Label{
				{
					Name:  "method",
					Value: key[1],
				},
			}
		}
	}
	if len(key) == 3 {
		if key[0] == "vault" && (key[1] == "barrier" || key[1] == "token" || key[1] == "policy") {
			// vault.barrier.<method>
			// vault.token.<method>
			// vault.policy.<method>
			return key[:2], []metrics.Label{
				{
					Name:  "method",
					Value: key[2],
				},
			}
		}
		if key[0] == "vault" && (key[2] == "put" || key[2] == "get" || key[2] == "delete" || key[2] == "list") {
			// vault.<backend>.<method>
			return key[:2], []metrics.Label{
				{
					Name:  "method",
					Value: key[2],
				},
			}
		}
		if key[0] == "database" && key[2] != "error" {
			// database.<name>.<method>
			// note: there are database.<method>.error counters
			return key[:1], []metrics.Label{
				{
					Name:  "name",
					Value: key[1],
				},
				{
					Name:  "method",
					Value: key[2],
				},
			}

		}
		if key[0] == "database" && key[2] == "error" {
			// database.<method>.error
			return []string{"database", "error"}, []metrics.Label{
				{
					Name:  "method",
					Value: key[1],
				},
			}
		}
	}
	if len(key) == 4 {
		if key[0] == "vault" && key[1] == "route" {
			// vault.route.<method>.<mount>
			return key[:2], []metrics.Label{
				{
					Name:  "method",
					Value: key[2],
				},
				{
					Name:  "mount",
					Value: key[3],
				},
			}
		}
		if key[0] == "vault" && key[1] == "rollback" && key[2] == "attempt" {
			// vault.rollback.attempt.<mount>
			return key[:3], []metrics.Label{
				{
					Name:  "mount",
					Value: key[3],
				},
			}
		}
		if key[0] == "vault" && key[2] == "lock" {
			// vault.<backend>.lock.<method>
			return key[:3], []metrics.Label{
				{
					Name:  "method",
					Value: key[3],
				},
			}
		}
		if key[0] == "database" && key[3] == "error" {
			// database.<name>.<method>.error
			return []string{key[0], key[3]}, []metrics.Label{
				{
					Name:  "name",
					Value: key[1],
				},
				{
					Name:  "method",
					Value: key[2],
				},
			}
		}
	}
	return key, nil
}
