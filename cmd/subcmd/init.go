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

package subcmd

import (
	"fmt"
	"gsmate/internal/orderedmap"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var subcmds = orderedmap.NewOrderedMap[string, *cli.Command]()

func init() {
	subcmds.Set("login", newLoginCmd())
	subcmds.Set("version", newVersionCmd())
}

func GetSubCmds() *orderedmap.OrderedMap[string, *cli.Command] {
	return subcmds
}

// newDefaultCmd creates a new cli.Command with default settings.
//
// It returns a pointer to a new cli.Command with the following settings:
//   - HideHelp: true
//   - UseShortOptionHandling: true
//
// Returns:
//   - *cli.Command
func newDefaultCmd() *cli.Command {
	cmd := &cli.Command{
		HideHelp:               true,
		UseShortOptionHandling: true,
	}
	cmd.Flags = append(cmd.Flags, &cli.BoolFlag{
		Name:               "help",
		Aliases:            []string{"?"},
		Usage:              "Show help information",
		DisableDefaultText: true,
	})
	return cmd
}

// PrintError prints an error message to the standard error stream in red color.
//
// It takes an error object as a parameter and prints it along with the "error:" prefix.
func PrintError(err any) {
	fmt.Fprintln(os.Stderr, color.RedString("error: %v", err))
}

func PrintCommand(cmd, desc string) {
	fmt.Fprintf(os.Stdout, "%-30s: %s\n", color.GreenString(cmd), desc)
}
