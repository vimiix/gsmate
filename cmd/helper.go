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
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

type CommandMap map[string]*cli.Command

func (c CommandMap) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (c CommandMap) Values() []*cli.Command {
	values := make([]*cli.Command, 0, len(c))
	for _, v := range c {
		values = append(values, v)
	}
	return values
}

var subCmds = CommandMap{
	"version": newVersionCmd(),
}

func newDefaultCmd() *cli.Command {
	return &cli.Command{
		HideHelp:               true,
		UseShortOptionHandling: true,
	}
}

func printError(err any) {
	fmt.Fprintln(os.Stderr, color.RedString("error: %v", err))
}
