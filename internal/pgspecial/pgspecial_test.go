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

package pgspecial

import "testing"

func TestParseSpecialCommand(t *testing.T) {
	// Test case: command with no arguments
	cmd := ParseSpecialCommand("\\q")
	if cmd.Command != "\\q" || cmd.Args != "" || cmd.Verbose != false {
		t.Errorf("ParseSpecialCommand(\"\\q\") = %v, want {Command: \"\\q\", Args: \"\", Verbose: false}", cmd)
	}

	// Test case: command with arguments
	cmd = ParseSpecialCommand("\\q arg1 arg2")
	if cmd.Command != "\\q" || cmd.Args != "arg1 arg2" || cmd.Verbose != false {
		t.Errorf("ParseSpecialCommand(\"\\q arg1 arg2\") = %v, want {Command: \"\\q\", Args: \"arg1 arg2\", Verbose: false}", cmd)
	}

	// Test case: command with verbose flag
	cmd = ParseSpecialCommand("\\q+")
	if cmd.Command != "\\q" || cmd.Args != "" || cmd.Verbose != true {
		t.Errorf("ParseSpecialCommand(\"\\q+\") = %v, want {Command: \"\\q\", Args: \"\", Verbose: true}", cmd)
	}

	// Test case: command with arguments and verbose flag
	cmd = ParseSpecialCommand("\\q+ arg1 arg2")
	if cmd.Command != "\\q" || cmd.Args != "arg1 arg2" || cmd.Verbose != true {
		t.Errorf("ParseSpecialCommand(\"\\q+ arg1 arg2\") = %v, want {Command: \"\\q\", Args: \"arg1 arg2\", Verbose: true}", cmd)
	}
}
