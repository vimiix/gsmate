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

package utils

import (
	"gsmate/internal/logger"

	"golang.org/x/term"
)

var getWindowSize = term.GetSize

func Chunks(vals []string) [][]string {
	w, _, err := getWindowSize(0)
	if err != nil {
		logger.Debug("failed to get terminal size, set to 80 default: %s", err)
		w = 80
	}

	var max int
	for _, v := range vals {
		if len(v) > max {
			max = len(v)
		}
	}

	cols := w / (max + 2)
	if cols == 0 {
		cols = 1
	}

	var rs [][]string
	for i := 0; i < len(vals); i += cols {
		end := i + cols
		if end >= len(vals) {
			end = len(vals)
		}
		rs = append(rs, vals[i:end])
	}
	return rs
}
