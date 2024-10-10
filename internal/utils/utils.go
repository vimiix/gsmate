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
	"fmt"
	"os"
	"strings"
	"unicode"

	"gsmate/internal/logger"

	"github.com/fatih/color"
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

// PrintError prints an error message to the standard error stream in red color.
//
// It takes an error object as a parameter and prints it along with the "error:" prefix.
func PrintError(err any) {
	fmt.Fprintln(os.Stderr, color.RedString("error: %v", err))
}

// Getenv gets the value of one or more environment variables.
//
// It takes one or more environment variable names as parameters and returns
// the value of the first environment variable that is set. If none of the
// environment variables are set, it returns an empty string along with a
// false boolean value.
//
// If multiple keys are given, it will return the value of the first key that
// is set.
func Getenv(keys ...string) (string, bool) {
	m := make(map[string]string)
	for _, v := range os.Environ() {
		if i := strings.Index(v, "="); i != -1 {
			m[v[:i]] = v[i+1:]
		}
	}
	for _, key := range keys {
		if v, ok := m[key]; ok {
			return v, true
		}
	}
	return "", false
}

// Grab grabs i from r, or returns 0 if i >= end.
func Grab(r []rune, i, end int) rune {
	if i < end {
		return r[i]
	}
	return 0
}

// EmptyStr reports whether s contains at least one printable, non-space character.
func EmptyStr(s string) bool {
	i := strings.IndexFunc(s, func(r rune) bool {
		return unicode.IsPrint(r) && !unicode.IsSpace(r)
	})
	return i == -1
}
