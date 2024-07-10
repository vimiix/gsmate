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
	"os/exec"
	"strings"
	"time"

	"gsmate/cmd/subcmd"
	"gsmate/internal/logger"
	"gsmate/pkg/version"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var (
	authors = []*cli.Author{
		{Name: "Vimiix", Email: "i@vimiix.com"},
	}
	copyright = func() string {
		yearRange := "2024"
		nowYear := time.Now().Year()
		if nowYear > 2024 {
			yearRange = fmt.Sprintf("2024-%d", nowYear)
		}
		return fmt.Sprintf("Copyright (C) %s Vimiix", yearRange)
	}
)

var interactiveCmds = []struct {
	Names []string
	Usage string
}{
	{
		Names: []string{"help", "\\h"},
		Usage: "show all available commands in interactive mode",
	},
	{
		Names: []string{"exit", "\\q"},
		Usage: "exit interactive mode",
	},
	{
		Names: []string{"shell", "\\!"},
		Usage: `run shell command, such as: "\! ls -l"`,
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "gsmate"
	app.Usage = "Yet another openGauss client and collector"
	app.Version = version.Version
	app.HideVersion = true // self control version flag to ensure help massage style is consistent
	app.Authors = authors
	app.Copyright = copyright()
	app.EnableBashCompletion = true
	app.UseShortOptionHandling = true
	app.HideHelp = true
	app.Suggest = true
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:               "help",
			Aliases:            []string{"?"},
			Usage:              "Show help information",
			DisableDefaultText: true,
		},
		&cli.BoolFlag{
			Name:               "version",
			Aliases:            []string{"v"},
			Usage:              "Print the version",
			DisableDefaultText: true,
		},
		&cli.StringFlag{
			Name:  "log-level",
			Usage: "Log level (case-insensitive), support:\ndebug, info, warn, error, fatal",
			Value: "info",
		},
		&cli.BoolFlag{
			Name:               "silence",
			Usage:              "Mute logger",
			DisableDefaultText: true,
		},
	}
	app.Commands = subcmd.GetSubCmds().Values()

	app.Action = func(c *cli.Context) error {
		if c.Bool("help") {
			return cli.ShowAppHelp(c)
		}
		if c.Bool("version") {
			cli.ShowVersion(c)
			return nil
		}
		if c.Bool("silence") {
			logger.MuteLogger()
		} else {
			logger.SetLogLevelByString(c.String("log-level"))
		}
		prompt.New(
			executor(c),
			completer,
			prompt.OptionTitle(c.App.Usage),
			prompt.OptionLivePrefix(promptPrefix),
			prompt.OptionInputTextColor(prompt.Yellow),
			prompt.OptionShowCompletionAtStart(),
		).Run()
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		subcmd.PrintError(err)
		os.Exit(1)
	}
}

func executor(cCtx *cli.Context) func(string) {
	return func(in string) {
		in = strings.TrimSpace(in)
		args := strings.Fields(in)
		switch args[0] {
		case "exit", "\\q":
			color.Green("Bye!")
			os.Exit(0)
		case "help", "\\h":
			fmt.Println("Gsmate Commands:")
			for _, cmd := range cCtx.App.Commands {
				subcmd.PrintCommand(cmd.Name, cmd.Usage)
			}
			fmt.Println("Type '<command> -?' for more information.")
			fmt.Println()
			fmt.Println("Interactive Commands:")
			for _, c := range interactiveCmds {
				subcmd.PrintCommand(strings.Join(c.Names, ", "), c.Usage)
			}
		case "shell", "\\!":
			command := exec.Command("sh", "-c", strings.Join(args[1:], " "))
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			command.Stdin = os.Stdin
			command.Env = os.Environ()
			if err := command.Run(); err != nil {
				subcmd.PrintError(err)
			}
		default:
			_, exist := subcmd.GetSubCmds().Get(args[0])
			if !exist {
				subcmd.PrintError(fmt.Errorf("gsmate: '%s' is not a gsmate command", args[0]))
				println("Type 'help' or '\\h' to see all available commands.")
				return
			}
			if err := cCtx.App.Run(append([]string{"gsmate"}, args...)); err != nil {
				subcmd.PrintError(err)
			}
		}
	}
}

// completer implements the prompt.Completer interface.
func completer(d prompt.Document) []prompt.Suggest {
	preText := strings.TrimSpace(d.TextBeforeCursor())
	if preText == "" {
		return nil
	}
	if len(strings.Fields(preText)) < 2 {
		return commandSuggestions(preText)
	}

	words := strings.Split(preText, " ")
	subcmd, exist := subcmd.GetSubCmds().Get(words[0])
	if !exist {
		return nil
	}

	lastWord := d.GetWordBeforeCursor()
	// If word before the cursor starts with "-", returns CLI flag options.
	if strings.HasPrefix(lastWord, "-") {
		return cmdOptionSuggestions(subcmd, lastWord)
	}
	return nil
}

func commandSuggestions(arg string) []prompt.Suggest {
	var rs []prompt.Suggest
	subcmd.GetSubCmds().Range(func(key string, cmd *cli.Command) {
		if strings.HasPrefix(key, arg) {
			rs = append(rs, prompt.Suggest{Text: key, Description: cmd.Usage})
		}
	})
	for _, c := range interactiveCmds {
		for _, name := range c.Names {
			if strings.HasPrefix(name, arg) {
				rs = append(rs, prompt.Suggest{Text: name, Description: c.Usage})
			}
		}

	}
	return rs
}

func cmdOptionSuggestions(cmd *cli.Command, arg string) []prompt.Suggest {
	rs := make([]prompt.Suggest, 0)
	unique := make(map[string]bool)
	for _, flag := range cmd.Flags {
		for _, name := range flag.Names() {
			text := "-" + name
			if len(name) > 1 {
				text = "--" + name
			}
			if strings.HasPrefix(text, arg) && !unique[text] {
				unique[text] = true
				rs = append(rs, prompt.Suggest{Text: text})
			}
		}
	}
	return rs
}

func promptPrefix() (string, bool) {
	return "gsmate> ", true
}
