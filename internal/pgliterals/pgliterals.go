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

package pgliterals

import (
	"embed"
	"encoding/json"
	"gsmate/internal/logger"
)

type PGLiterals struct {
	KeyWords  map[string][]string `json:"keywords"`
	Functions []string            `json:"functions"`
	DataTypes []string            `json:"datatypes"`
	Reserved  []string            `json:"reserved"`
}

var literals = &PGLiterals{}

//go:embed pgliterals.json
var pgLiteralsFile embed.FS

func init() {
	bs, err := pgLiteralsFile.ReadFile("pgliterals.json")
	if err != nil {
		logger.Fatal("read pgliterals.json error: %v", err)
	}
	if err := json.Unmarshal(bs, literals); err != nil {
		logger.Fatal("unmarshal pgliterals.json error: %v", err)
	}
}

func GetKeywords() map[string][]string {
	return literals.KeyWords
}

func GetFunctions() []string {
	return literals.Functions
}

func GetDatatypes() []string {
	return literals.DataTypes
}

func GetReserved() []string {
	return literals.Reserved
}
