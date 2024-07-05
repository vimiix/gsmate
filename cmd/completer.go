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

package main

import (
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/urfave/cli/v2"
)

type cmdCompleter struct {
	cmds CommandMap
}

func (c *cmdCompleter) commandSuggestions(arg string) []prompt.Suggest {
	rs := make([]prompt.Suggest, 0)
	for key, cmd := range c.cmds {
		rs = append(rs, prompt.Suggest{Text: key, Description: cmd.Usage})
	}
	return prompt.FilterHasPrefix(rs, arg, false)
}

func (c *cmdCompleter) cmdOptionSuggestions(cmd *cli.Command, arg string) []prompt.Suggest {
	var processFlag = func(arg string, flag cli.Flag) []prompt.Suggest {
		var rs []prompt.Suggest
		for _, name := range flag.Names() {
			text := "-" + name
			if len(name) > 1 {
				text = "--" + name
			}
			rs = append(rs, prompt.Suggest{Text: text})
		}
		return rs
	}

	rs := make([]prompt.Suggest, 0)
	for _, flag := range append(conntionFlags, cmd.Flags...) {
		rs = append(rs, processFlag(arg, flag)...)
	}
	return prompt.FilterHasPrefix(rs, arg, true)
}

func (c *cmdCompleter) PromptPrefix() (string, bool) {
	return "gsmate > ", true
}

func (c *cmdCompleter) Complete(d prompt.Document) []prompt.Suggest {
	preText := strings.TrimSpace(d.TextBeforeCursor())
	if preText == "" {
		return nil
	}
	if len(strings.Fields(preText)) < 2 {
		return c.commandSuggestions(preText)
	}

	words := strings.Split(preText, " ")
	subcmd, ok := c.cmds[words[0]]
	if !ok {
		return nil
	}

	lastWord := d.GetWordBeforeCursor()
	// If word before the cursor starts with "-", returns CLI flag options.
	if strings.HasPrefix(lastWord, "-") {
		return c.cmdOptionSuggestions(subcmd, lastWord)
	}
	return nil
}

func newCmdCompleter() *cmdCompleter {
	return &cmdCompleter{cmds: subCmds}
}
