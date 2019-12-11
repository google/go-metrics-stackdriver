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
	stackdriver "github.com/google/go-metrics-stackdriver"
)

func TestExtractor(t *testing.T) {
	tests := []struct {
		desc       string
		in         []string
		wantKey    []string
		wantLabels []metrics.Label
	}{
		// https://www.vaultproject.io/docs/internals/telemetry.html#audit-metrics
		{
			desc:    "vault.audit.log_request",
			in:      []string{"vault", "audit", "log_request"},
			wantKey: []string{"vault", "audit", "log_request"},
		},
		{
			desc:    "vault.audit.log_response",
			in:      []string{"vault", "audit", "log_response"},
			wantKey: []string{"vault", "audit", "log_response"},
		},
		{
			desc:    "vault.audit.log_request_failure",
			in:      []string{"vault", "audit", "log_request_failure"},
			wantKey: []string{"vault", "audit", "log_request_failure"},
		},
		{
			desc:    "vault.audit.log_response_failure",
			in:      []string{"vault", "audit", "log_response_failure"},
			wantKey: []string{"vault", "audit", "log_response_failure"},
		},
		{
			desc:    "vault.audit.file.log_request",
			in:      []string{"vault", "audit", "file", "log_request"},
			wantKey: []string{"vault", "audit", "log_request"},
			wantLabels: []metrics.Label{
				{
					Name:  "type",
					Value: "file",
				},
			},
		},
		{
			desc:    "vault.audit.file.log_response",
			in:      []string{"vault", "audit", "file", "log_response"},
			wantKey: []string{"vault", "audit", "log_response"},
			wantLabels: []metrics.Label{
				{
					Name:  "type",
					Value: "file",
				},
			},
		},
		{
			desc:    "vault.audit.file.log_request_failure",
			in:      []string{"vault", "audit", "file", "log_request_failure"},
			wantKey: []string{"vault", "audit", "log_request_failure"},
			wantLabels: []metrics.Label{
				{
					Name:  "type",
					Value: "file",
				},
			},
		},
		{
			desc:    "vault.audit.file.log_response_failure",
			in:      []string{"vault", "audit", "file", "log_response_failure"},
			wantKey: []string{"vault", "audit", "log_response_failure"},
			wantLabels: []metrics.Label{
				{
					Name:  "type",
					Value: "file",
				},
			},
		},
		{
			desc:    "vault.audit.syslog.log_request",
			in:      []string{"vault", "audit", "syslog", "log_request"},
			wantKey: []string{"vault", "audit", "log_request"},
			wantLabels: []metrics.Label{
				{
					Name:  "type",
					Value: "syslog",
				},
			},
		},
		{
			desc:    "vault.audit.syslog.log_response",
			in:      []string{"vault", "audit", "syslog", "log_response"},
			wantKey: []string{"vault", "audit", "log_response"},
			wantLabels: []metrics.Label{
				{
					Name:  "type",
					Value: "syslog",
				},
			},
		},
		{
			desc:    "vault.audit.syslog.log_request_failure",
			in:      []string{"vault", "audit", "syslog", "log_request_failure"},
			wantKey: []string{"vault", "audit", "log_request_failure"},
			wantLabels: []metrics.Label{
				{
					Name:  "type",
					Value: "syslog",
				},
			},
		},
		{
			desc:    "vault.audit.syslog.log_response_failure",
			in:      []string{"vault", "audit", "syslog", "log_response_failure"},
			wantKey: []string{"vault", "audit", "log_response_failure"},
			wantLabels: []metrics.Label{
				{
					Name:  "type",
					Value: "syslog",
				},
			},
		},
		// https://www.vaultproject.io/docs/internals/telemetry.html#core-metrics
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
			desc:    "vault.barrier.put",
			in:      []string{"vault", "barrier", "put"},
			wantKey: []string{"vault", "barrier"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "put",
				},
			},
		},
		{
			desc:    "vault.barrier.delete",
			in:      []string{"vault", "barrier", "delete"},
			wantKey: []string{"vault", "barrier"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "delete",
				},
			},
		},
		{
			desc:    "vault.barrier.list",
			in:      []string{"vault", "barrier", "list"},
			wantKey: []string{"vault", "barrier"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "list",
				},
			},
		},
		{
			desc:    "vault.core.check_token",
			in:      []string{"vault", "core", "check_token"},
			wantKey: []string{"vault", "core", "check_token"},
		},
		{
			desc:    "vault.core.fetch_acl_and_token",
			in:      []string{"vault", "core", "fetch_acl_and_token"},
			wantKey: []string{"vault", "core", "fetch_acl_and_token"},
		},
		{
			desc:    "vault.core.handle_request",
			in:      []string{"vault", "core", "handle_request"},
			wantKey: []string{"vault", "core", "handle_request"},
		},
		{
			desc:    "vault.core.handle_login_request",
			in:      []string{"vault", "core", "handle_login_request"},
			wantKey: []string{"vault", "core", "handle_login_request"},
		},
		{
			desc:    "vault.core.leadership_setup_failed",
			in:      []string{"vault", "core", "leadership_setup_failed"},
			wantKey: []string{"vault", "core", "leadership_setup_failed"},
		},
		{
			desc:    "vault.core.leadership_lost",
			in:      []string{"vault", "core", "leadership_lost"},
			wantKey: []string{"vault", "core", "leadership_lost"},
		},
		{
			desc:    "vault.core.post_unseal",
			in:      []string{"vault", "core", "post_unseal"},
			wantKey: []string{"vault", "core", "post_unseal"},
		},
		{
			desc:    "vault.core.pre_seal",
			in:      []string{"vault", "core", "pre_seal"},
			wantKey: []string{"vault", "core", "pre_seal"},
		},
		{
			desc:    "vault.core.seal-with-request",
			in:      []string{"vault", "core", "seal-with-request"},
			wantKey: []string{"vault", "core", "seal-with-request"},
		},
		{
			desc:    "vault.core.seal",
			in:      []string{"vault", "core", "seal"},
			wantKey: []string{"vault", "core", "seal"},
		},
		{
			desc:    "vault.core.seal-internal",
			in:      []string{"vault", "core", "seal-internal"},
			wantKey: []string{"vault", "core", "seal-internal"},
		},
		{
			desc:    "vault.core.step_down",
			in:      []string{"vault", "core", "step_down"},
			wantKey: []string{"vault", "core", "step_down"},
		},
		{
			desc:    "vault.core.unseal",
			in:      []string{"vault", "core", "unseal"},
			wantKey: []string{"vault", "core", "unseal"},
		},
		// https://www.vaultproject.io/docs/internals/telemetry.html#runtime-metrics
		{
			desc:    "vault.runtime.alloc_bytes",
			in:      []string{"vault", "runtime", "alloc_bytes"},
			wantKey: []string{"vault", "runtime", "alloc_bytes"},
		},
		{
			desc:    "vault.runtime.free_count",
			in:      []string{"vault", "runtime", "free_count"},
			wantKey: []string{"vault", "runtime", "free_count"},
		},
		{
			desc:    "vault.runtime.heap_objects",
			in:      []string{"vault", "runtime", "heap_objects"},
			wantKey: []string{"vault", "runtime", "heap_objects"},
		},
		{
			desc:    "vault.runtime.malloc_count",
			in:      []string{"vault", "runtime", "malloc_count"},
			wantKey: []string{"vault", "runtime", "malloc_count"},
		},
		{
			desc:    "vault.runtime.num_goroutines",
			in:      []string{"vault", "runtime", "num_goroutines"},
			wantKey: []string{"vault", "runtime", "num_goroutines"},
		},
		{
			desc:    "vault.runtime.sys_bytes",
			in:      []string{"vault", "runtime", "sys_bytes"},
			wantKey: []string{"vault", "runtime", "sys_bytes"},
		},
		{
			desc:    "vault.runtime.total_gc_pause_ns",
			in:      []string{"vault", "runtime", "total_gc_pause_ns"},
			wantKey: []string{"vault", "runtime", "total_gc_pause_ns"},
		},
		{
			desc:    "vault.runtime.total_gc_runs",
			in:      []string{"vault", "runtime", "total_gc_runs"},
			wantKey: []string{"vault", "runtime", "total_gc_runs"},
		},
		// https://www.vaultproject.io/docs/internals/telemetry.html#policy-and-token-metrics
		{
			desc:    "vault.expire.fetch-lease-times",
			in:      []string{"vault", "expire", "fetch-lease-times"},
			wantKey: []string{"vault", "expire", "fetch-lease-times"},
		},
		{
			desc:    "vault.expire.fetch-lease-times-by-token",
			in:      []string{"vault", "expire", "fetch-lease-times-by-token"},
			wantKey: []string{"vault", "expire", "fetch-lease-times-by-token"},
		},
		{
			desc:    "vault.expire.num_leases",
			in:      []string{"vault", "expire", "num_leases"},
			wantKey: []string{"vault", "expire", "num_leases"},
		},
		{
			desc:    "vault.expire.revoke",
			in:      []string{"vault", "expire", "revoke"},
			wantKey: []string{"vault", "expire", "revoke"},
		},
		{
			desc:    "vault.expire.revoke-force",
			in:      []string{"vault", "expire", "revoke-force"},
			wantKey: []string{"vault", "expire", "revoke-force"},
		},
		{
			desc:    "vault.expire.revoke-prefix",
			in:      []string{"vault", "expire", "revoke-prefix"},
			wantKey: []string{"vault", "expire", "revoke-prefix"},
		},
		{
			desc:    "vault.expire.revoke-by-token",
			in:      []string{"vault", "expire", "revoke-by-token"},
			wantKey: []string{"vault", "expire", "revoke-by-token"},
		},
		{
			desc:    "vault.expire.renew",
			in:      []string{"vault", "expire", "renew"},
			wantKey: []string{"vault", "expire", "renew"},
		},
		{
			desc:    "vault.expire.renew-token",
			in:      []string{"vault", "expire", "renew-token"},
			wantKey: []string{"vault", "expire", "renew-token"},
		},
		{
			desc:    "vault.expire.register",
			in:      []string{"vault", "expire", "register"},
			wantKey: []string{"vault", "expire", "register"},
		},
		{
			desc:    "vault.expire.register-auth",
			in:      []string{"vault", "expire", "register-auth"},
			wantKey: []string{"vault", "expire", "register-auth"},
		},
		{
			desc:    "vault.policy.get_policy",
			in:      []string{"vault", "policy", "get_policy"},
			wantKey: []string{"vault", "policy"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "get_policy",
				},
			},
		},
		{
			desc:    "vault.policy.list_policy",
			in:      []string{"vault", "policy", "list_policy"},
			wantKey: []string{"vault", "policy"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "list_policy",
				},
			},
		},
		{
			desc:    "vault.policy.delete_policy",
			in:      []string{"vault", "policy", "delete_policy"},
			wantKey: []string{"vault", "policy"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "delete_policy",
				},
			},
		},
		{
			desc:    "vault.policy.set_policy",
			in:      []string{"vault", "policy", "set_policy"},
			wantKey: []string{"vault", "policy"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "set_policy",
				},
			},
		},
		{
			desc:    "vault.token.create",
			in:      []string{"vault", "token", "create"},
			wantKey: []string{"vault", "token"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "create",
				},
			},
		},
		{
			desc:    "vault.token.createAccessor",
			in:      []string{"vault", "token", "createAccessor"},
			wantKey: []string{"vault", "token"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "createAccessor",
				},
			},
		},
		{
			desc:    "vault.token.lookup",
			in:      []string{"vault", "token", "lookup"},
			wantKey: []string{"vault", "token"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "lookup",
				},
			},
		},
		{
			desc:    "vault.token.revoke",
			in:      []string{"vault", "token", "revoke"},
			wantKey: []string{"vault", "token"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "revoke",
				},
			},
		},
		{
			desc:    "vault.token.revoke-tree",
			in:      []string{"vault", "token", "revoke-tree"},
			wantKey: []string{"vault", "token"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "revoke-tree",
				},
			},
		},
		{
			desc:    "vault.token.store",
			in:      []string{"vault", "token", "store"},
			wantKey: []string{"vault", "token"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "store",
				},
			},
		},
		// https://www.vaultproject.io/docs/internals/telemetry.html#auth-methods-metrics
		{
			desc:    "vault.rollback.attempt.auth-token",
			in:      []string{"vault", "rollback", "attempt", "auth-token"},
			wantKey: []string{"vault", "rollback", "attempt"},
			wantLabels: []metrics.Label{
				{
					Name:  "mount",
					Value: "auth-token",
				},
			},
		},
		{
			desc:    "vault.rollback.attempt.auth-ldap",
			in:      []string{"vault", "rollback", "attempt", "auth-ldap"},
			wantKey: []string{"vault", "rollback", "attempt"},
			wantLabels: []metrics.Label{
				{
					Name:  "mount",
					Value: "auth-ldap",
				},
			},
		},
		{
			desc:    "vault.rollback.attempt.cubbyhole",
			in:      []string{"vault", "rollback", "attempt", "cubbyhole"},
			wantKey: []string{"vault", "rollback", "attempt"},
			wantLabels: []metrics.Label{
				{
					Name:  "mount",
					Value: "cubbyhole",
				},
			},
		},
		{
			desc:    "vault.rollback.attempt.secret",
			in:      []string{"vault", "rollback", "attempt", "secret"},
			wantKey: []string{"vault", "rollback", "attempt"},
			wantLabels: []metrics.Label{
				{
					Name:  "mount",
					Value: "secret",
				},
			},
		},
		{
			desc:    "vault.rollback.attempt.sys",
			in:      []string{"vault", "rollback", "attempt", "sys"},
			wantKey: []string{"vault", "rollback", "attempt"},
			wantLabels: []metrics.Label{
				{
					Name:  "mount",
					Value: "sys",
				},
			},
		},
		{
			desc:    "vault.route.rollback.auth-token",
			in:      []string{"vault", "route", "rollback", "auth-token"},
			wantKey: []string{"vault", "route"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "rollback",
				},
				{
					Name:  "mount",
					Value: "auth-token",
				},
			},
		},
		{
			desc:    "vault.route.rollback.auth-ldap",
			in:      []string{"vault", "route", "rollback", "auth-ldap"},
			wantKey: []string{"vault", "route"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "rollback",
				},
				{
					Name:  "mount",
					Value: "auth-ldap",
				},
			},
		},
		{
			desc:    "vault.route.rollback.cubbyhole",
			in:      []string{"vault", "route", "rollback", "cubbyhole"},
			wantKey: []string{"vault", "route"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "rollback",
				},
				{
					Name:  "mount",
					Value: "cubbyhole",
				},
			},
		},
		{
			desc:    "vault.route.rollback.secret",
			in:      []string{"vault", "route", "rollback", "secret"},
			wantKey: []string{"vault", "route"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "rollback",
				},
				{
					Name:  "mount",
					Value: "secret",
				},
			},
		},
		{
			desc:    "vault.route.rollback.sys",
			in:      []string{"vault", "route", "rollback", "sys"},
			wantKey: []string{"vault", "route"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "rollback",
				},
				{
					Name:  "mount",
					Value: "sys",
				},
			},
		},
		// https://www.vaultproject.io/docs/internals/telemetry.html#merkle-tree-and-write-ahead-log-metrics
		{
			desc:    "vault.merkle_flushdirty",
			in:      []string{"vault", "merkle_flushdirty"},
			wantKey: []string{"vault", "merkle_flushdirty"},
		},
		{
			desc:    "vault.merkle_savecheckpoint",
			in:      []string{"vault", "merkle_savecheckpoint"},
			wantKey: []string{"vault", "merkle_savecheckpoint"},
		},
		{
			desc:    "vault.wal_deletewals",
			in:      []string{"vault", "wal_deletewals"},
			wantKey: []string{"vault", "wal_deletewals"},
		},
		{
			desc:    "vault.wal_gc_deleted",
			in:      []string{"vault", "wal_gc_deleted"},
			wantKey: []string{"vault", "wal_gc_deleted"},
		},
		{
			desc:    "vault.wal_gc_total",
			in:      []string{"vault", "wal_gc_total"},
			wantKey: []string{"vault", "wal_gc_total"},
		},
		{
			desc:    "vault.wal_loadWAL",
			in:      []string{"vault", "wal_loadWAL"},
			wantKey: []string{"vault", "wal_loadWAL"},
		},
		{
			desc:    "vault.wal_persistwals",
			in:      []string{"vault", "wal_persistwals"},
			wantKey: []string{"vault", "wal_persistwals"},
		},
		{
			desc:    "vault.wal_flushready",
			in:      []string{"vault", "wal_flushready"},
			wantKey: []string{"vault", "wal_flushready"},
		},
		// https://www.vaultproject.io/docs/internals/telemetry.html#replication-metrics
		// TODO
		// https://www.vaultproject.io/docs/internals/telemetry.html#secrets-engines-metrics
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
			desc:    "database.Close",
			in:      []string{"database", "Close"},
			wantKey: []string{"database"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "Close",
				},
			},
		},
		{
			desc:    "database.Close.error",
			in:      []string{"database", "Close", "error"},
			wantKey: []string{"database", "error"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "Close",
				},
			},
		},
		{
			desc:    "database.foo.Close",
			in:      []string{"database", "foo", "Close"},
			wantKey: []string{"database"},
			wantLabels: []metrics.Label{
				{
					Name:  "name",
					Value: "foo",
				},
				{
					Name:  "method",
					Value: "Close",
				},
			},
		},
		{
			desc:    "database.foo.Close.error",
			in:      []string{"database", "foo", "Close", "error"},
			wantKey: []string{"database", "error"},
			wantLabels: []metrics.Label{
				{
					Name:  "name",
					Value: "foo",
				},
				{
					Name:  "method",
					Value: "Close",
				},
			},
		},
		{
			desc:    "database.CreateUser",
			in:      []string{"database", "CreateUser"},
			wantKey: []string{"database"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "CreateUser",
				},
			},
		},
		{
			desc:    "database.CreateUser.error",
			in:      []string{"database", "CreateUser", "error"},
			wantKey: []string{"database", "error"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "CreateUser",
				},
			},
		},
		{
			desc:    "database.foo.CreateUser",
			in:      []string{"database", "foo", "CreateUser"},
			wantKey: []string{"database"},
			wantLabels: []metrics.Label{
				{
					Name:  "name",
					Value: "foo",
				},
				{
					Name:  "method",
					Value: "CreateUser",
				},
			},
		},
		{
			desc:    "database.foo.CreateUser.error",
			in:      []string{"database", "foo", "CreateUser", "error"},
			wantKey: []string{"database", "error"},
			wantLabels: []metrics.Label{
				{
					Name:  "name",
					Value: "foo",
				},
				{
					Name:  "method",
					Value: "CreateUser",
				},
			},
		},
		{
			desc:    "database.RenewUser",
			in:      []string{"database", "RenewUser"},
			wantKey: []string{"database"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "RenewUser",
				},
			},
		},
		{
			desc:    "database.RenewUser.error",
			in:      []string{"database", "RenewUser", "error"},
			wantKey: []string{"database", "error"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "RenewUser",
				},
			},
		},
		{
			desc:    "database.foo.RenewUser",
			in:      []string{"database", "foo", "RenewUser"},
			wantKey: []string{"database"},
			wantLabels: []metrics.Label{
				{
					Name:  "name",
					Value: "foo",
				},
				{
					Name:  "method",
					Value: "RenewUser",
				},
			},
		},
		{
			desc:    "database.foo.RenewUser.error",
			in:      []string{"database", "foo", "RenewUser", "error"},
			wantKey: []string{"database", "error"},
			wantLabels: []metrics.Label{
				{
					Name:  "name",
					Value: "foo",
				},
				{
					Name:  "method",
					Value: "RenewUser",
				},
			},
		},
		{
			desc:    "database.RevokeUser",
			in:      []string{"database", "RevokeUser"},
			wantKey: []string{"database"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "RevokeUser",
				},
			},
		},
		{
			desc:    "database.RevokeUser.error",
			in:      []string{"database", "RevokeUser", "error"},
			wantKey: []string{"database", "error"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "RevokeUser",
				},
			},
		},
		{
			desc:    "database.foo.RevokeUser",
			in:      []string{"database", "foo", "RevokeUser"},
			wantKey: []string{"database"},
			wantLabels: []metrics.Label{
				{
					Name:  "name",
					Value: "foo",
				},
				{
					Name:  "method",
					Value: "RevokeUser",
				},
			},
		},
		{
			desc:    "database.foo.RevokeUser.error",
			in:      []string{"database", "foo", "RevokeUser", "error"},
			wantKey: []string{"database", "error"},
			wantLabels: []metrics.Label{
				{
					Name:  "name",
					Value: "foo",
				},
				{
					Name:  "method",
					Value: "RevokeUser",
				},
			},
		},
		// https://www.vaultproject.io/docs/internals/telemetry.html#storage-backend-metrics
		{
			desc:    "vault.consul.put",
			in:      []string{"vault", "consul", "put"},
			wantKey: []string{"vault", "consul"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "put",
				},
			},
		},
		{
			desc:    "vault.consul.get",
			in:      []string{"vault", "consul", "get"},
			wantKey: []string{"vault", "consul"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "get",
				},
			},
		},
		{
			desc:    "vault.consul.delete",
			in:      []string{"vault", "consul", "delete"},
			wantKey: []string{"vault", "consul"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "delete",
				},
			},
		},
		{
			desc:    "vault.consul.list",
			in:      []string{"vault", "consul", "list"},
			wantKey: []string{"vault", "consul"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "list",
				},
			},
		},
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
			desc:    "vault.gcs.get",
			in:      []string{"vault", "gcs", "get"},
			wantKey: []string{"vault", "gcs"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "get",
				},
			},
		},
		{
			desc:    "vault.gcs.delete",
			in:      []string{"vault", "gcs", "delete"},
			wantKey: []string{"vault", "gcs"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "delete",
				},
			},
		},
		{
			desc:    "vault.gcs.list",
			in:      []string{"vault", "gcs", "list"},
			wantKey: []string{"vault", "gcs"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "list",
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
			desc:    "vault.spanner.put",
			in:      []string{"vault", "spanner", "put"},
			wantKey: []string{"vault", "spanner"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "put",
				},
			},
		},
		{
			desc:    "vault.spanner.get",
			in:      []string{"vault", "spanner", "get"},
			wantKey: []string{"vault", "spanner"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "get",
				},
			},
		},
		{
			desc:    "vault.spanner.delete",
			in:      []string{"vault", "spanner", "delete"},
			wantKey: []string{"vault", "spanner"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "delete",
				},
			},
		},
		{
			desc:    "vault.spanner.list",
			in:      []string{"vault", "spanner", "list"},
			wantKey: []string{"vault", "spanner"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "list",
				},
			},
		},
		{
			desc:    "vault.spanner.lock.lock",
			in:      []string{"vault", "spanner", "lock", "lock"},
			wantKey: []string{"vault", "spanner", "lock"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "lock",
				},
			},
		},
		{
			desc:    "vault.spanner.lock.unlock",
			in:      []string{"vault", "spanner", "lock", "unlock"},
			wantKey: []string{"vault", "spanner", "lock"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "unlock",
				},
			},
		},
		{
			desc:    "vault.spanner.lock.value",
			in:      []string{"vault", "spanner", "lock", "value"},
			wantKey: []string{"vault", "spanner", "lock"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "value",
				},
			},
		},
		// EXTRA UNDOCUMENTED
		{
			desc:    "vault.route.create.kv-",
			in:      []string{"vault", "route", "create", "kv-"},
			wantKey: []string{"vault", "route"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "create",
				},
				{
					Name:  "mount",
					Value: "kv-",
				},
			},
		},
		{
			desc:    "vault.route.update.kv-",
			in:      []string{"vault", "route", "update", "kv-"},
			wantKey: []string{"vault", "route"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "update",
				},
				{
					Name:  "mount",
					Value: "kv-",
				},
			},
		},
		{
			desc:    "vault.route.read.kv-",
			in:      []string{"vault", "route", "read", "kv-"},
			wantKey: []string{"vault", "route"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "read",
				},
				{
					Name:  "mount",
					Value: "kv-",
				},
			},
		},
		{
			desc:    "vault.route.delete.kv-",
			in:      []string{"vault", "route", "delete", "kv-"},
			wantKey: []string{"vault", "route"},
			wantLabels: []metrics.Label{
				{
					Name:  "method",
					Value: "delete",
				},
				{
					Name:  "mount",
					Value: "kv-",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			key, labels, _ := Extractor(tc.in)
			if diff := cmp.Diff(tc.wantKey, key); diff != "" {
				t.Errorf("Extractor(%s) mismatch key (-want +got):\n%s", tc.in, diff)
			}
			if diff := cmp.Diff(tc.wantLabels, labels); diff != "" {
				t.Errorf("Extractor(%s) mismatch labels (-want +got):\n%s", tc.in, diff)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	c := &stackdriver.Config{
		Extractor: Extractor,
		Bucketer:  Bucketer,
	}
}
