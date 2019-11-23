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
package vault

import (
	"testing"

	"github.com/armon/go-metrics"
	"github.com/google/go-cmp/cmp"
)

func TestExtractor(t *testing.T) {
	tests := []struct {
		desc       string
		in         []string
		wantKey    []string
		wantLabels []metrics.Label
	}{
		{
			desc:    "vault.gcs.put",
			in:      []string{"vault", "gcs", "put"},
			wantKey: []string{"vault", "gcs"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "put",
				},
			},
		},
		{
			desc:    "vault.gcs.lock.lock",
			in:      []string{"vault", "gcs", "lock", "lock"},
			wantKey: []string{"vault", "gcs", "lock"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "lock",
				},
			},
		},
		{
			desc:    "vault.gcs.lock.unlock",
			in:      []string{"vault", "gcs", "lock", "unlock"},
			wantKey: []string{"vault", "gcs", "lock"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "unlock",
				},
			},
		},
		{
			// TODO: this should maybe be its own metric, "value" isnt a method
			desc:    "vault.gcs.lock.value",
			in:      []string{"vault", "gcs", "lock", "value"},
			wantKey: []string{"vault", "gcs", "lock"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "value",
				},
			},
		},
		{
			desc:    "database.Initialize",
			in:      []string{"database", "Initialize"},
			wantKey: []string{"database"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "Initialize",
				},
			},
		},
		{
			desc:    "database.Initialize.error",
			in:      []string{"database", "Initialize", "error"},
			wantKey: []string{"database", "error"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "Initialize",
				},
			},
		},
		{
			desc:    "database.foo.Initialize",
			in:      []string{"database", "foo", "Initialize"},
			wantKey: []string{"database"},
			wantLabels: []metrics.Label{
				{
					Name:  "name",
					Value: "foo",
				},
				{
					Name:  "method",
					Value: "Initialize",
				},
			},
		},
		{
			desc:    "database.foo.Initialize.error",
			in:      []string{"database", "foo", "Initialize", "error"},
			wantKey: []string{"database", "error"},
			wantLabels: []metrics.Label{
				{
					Name:  "name",
					Value: "foo",
				},
				{
					Name:  "method",
					Value: "Initialize",
				},
			},
		},
		{
			desc:    "vault.policy.get",
			in:      []string{"vault", "policy", "get"},
			wantKey: []string{"vault", "policy"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "get",
				},
			},
		},
		{
			desc:    "vault.barrier.get",
			in:      []string{"vault", "barrier", "get"},
			wantKey: []string{"vault", "barrier"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "get",
				},
			},
		},
		{
			desc:    "vault.token.get",
			in:      []string{"vault", "token", "get"},
			wantKey: []string{"vault", "token"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "get",
				},
			},
		},
		{
			desc:    "vault.route.get.kv-",
			in:      []string{"vault", "route", "get", "kv-"},
			wantKey: []string{"vault", "route"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "get",
				},
				{
					Name:  "mount",
					Value: "kv-",
				},
			},
		},
		{
			desc:    "vault.rollback.attempt.kv-",
			in:      []string{"vault", "rollback", "attempt", "kv-"},
			wantKey: []string{"vault", "rollback", "attempt"},
			wantLabels: []metrics.Label{
				{
					Name:  "mount",
					Value: "kv-",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			key, labels := Extractor(tc.in)
			if diff := cmp.Diff(tc.wantKey, key); diff != "" {
				t.Errorf("Extractor(%s) mismatch key (-want +got):\n%s", tc.in, diff)
			}
			if diff := cmp.Diff(tc.wantLabels, labels); diff != "" {
				t.Errorf("Extractor(%s) mismatch labels (-want +got):\n%s", tc.in, diff)
			}
		})
	}

}
