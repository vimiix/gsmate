// Copyright 2024 Qian Yao
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import "github.com/vimiix/go-prompt"

// these objects can be create/alter/drop
var operableObj = []string{
	"AGGREGATE",
	"AUDIT POLICY",
	"DATABASE",
	"DATA SOURCE",
	"DEFAULT PRIVILEGES",
	"DIRECTORY",
	"EVENT",
	"EVENT TRIGGER",
	"EXTENSION",
	"FOREIGN DATA WRAPPER",
	"FOREIGN TABLE",
	"FUNCTION",
	"GLOBAL CONFIGURATION",
	"GROUP",
	"INDEX",
	"LANGUAGE",
	"LARGE OBJECT",
	"MASKING POLICY",
	"MATERIALIZED VIEW",
	"OPERATOR",
	"PACKAGE",
	"PROCEDURE",
	"PUBLICATION",
	"RESOURCE LABEL",
	"RESOURCE POOL",
	"ROLE",
	"ROW LEVEL SECURITY POLICY",
	"RULE",
	"SCHEMA",
	"SEQUENCE",
	"SERVER",
	"SUBSCRIPTION",
	"SYNONYM",
	"TABLE",
	"TABLE PARTITION",
	"TABLE SUBPARTITION",
	"TABLESPACE",
	"TEXT SEARCH CONFIGURATION",
	"TEXT SEARCH DICTIONARY",
	"TRIGGER",
	"TYPE",
	"USER",
	"USER MAPPING",
	"VIEW",
}

var startSQLCommands = map[string][]string{
	"ABORT": nil,
	"ALTER": append([]string{
		"SESSION",
		"SYSTEM KILL SESSION",
		"SYSTEM SET",
	}, operableObj...),
	"ANALYZE":    nil,
	"ANALYSE":    nil,
	"BEGIN":      nil,
	"CALL":       nil,
	"CHECKPOINT": nil,
	"CLEAN": {
		"CONNECTION TO",
	},
	"CLOSE":   nil,
	"CLUSTER": nil,
	"COMMENT": nil,
	"COMMIT": {
		"PREPARED",
	},
	"COPY": nil,
	"CREATE": append([]string{
		"SNAPSHOT",
	}, operableObj...),
	"CURSOR":     nil,
	"DEALLOCATE": nil,
	"DECLARE":    nil,
	"DELETE": {
		"FROM",
	},
	"DELIMITER": nil,
	"DO":        nil,
	"DROP":      operableObj,
	"END":       nil,
	"EXECUTE": {
		"DIRECT",
	},
	"EXPLAIN": {
		"PLAIN",
	},
	"FETCH": nil,
	"GRANT": nil,
	"INSERT": {
		"INTO",
	},
	"LOCK": nil,
	"MERGE": {
		"INTO",
	},
	"MOVE": nil,
	"PREDICT": {
		"BY",
	},
	"PREPARE": {
		"TRANSACTION",
	},
	"PURGE": {"SNAPSHOT"},
	"REASSIGN": {
		"OWNED",
	},
	"REFRESH": {
		"MATERIALIZED VIEW",
		"INCREMENTAL MATERIALIZED VIEW",
	},
	"REINDEX": nil,
	"RELEASE": {
		"SAVEPOINT",
	},
	"RESET":  nil,
	"REVOKE": nil,
	"ROLLBACK": {
		"PREPARED",
		"TO SAVEPOINT",
	},
	"SAVEPOINT": nil,
	"SELECT": {
		"INTO",
	},
	"SET": {
		"CONSTRAINTS",
		"ROLE",
		"SESSION AUTHORIZATION",
		"TRANSACTION",
	},
	"SHOW":        {"EVENTS"},
	"SHRINK":      {"TABLE", "INDEX"},
	"SHUTDOWN":    {"fast", "immediate"},
	"START":       {"TRANSACTION"},
	"TIMECAPSULE": {"TABLE"},
	"TRUNCATE":    {"TABLE"},
	"UPDATE":      nil,
	"VACUUM":      nil,
	"VALUES":      nil,
}

func getStartSQLCmdSuggests() []prompt.Suggest {
	r := make([]prompt.Suggest, 0, len(startSQLCommands))
	for k := range startSQLCommands {
		r = append(r, prompt.Suggest{Text: k})
	}
	return r
}

var backslashCommands = []prompt.Suggest{
	{Text: `\!`, Description: "execute command in shell or start interactive shell"},
	{Text: `\?`, Description: "show help on commands"},
	{Text: `\copyright`, Description: "show gsmater copyright information"},
}
