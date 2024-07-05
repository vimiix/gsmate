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
	"gsmate/src/model"
	"gsmate/src/version"
	"os"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/urfave/cli/v2"
)

var connOpts = &model.ConnectOptions{}

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

var conntionFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "host",
		Aliases:     []string{"h"},
		EnvVars:     []string{"PGHOST"},
		Destination: &connOpts.Host,
		Usage:       "Database server host or socket directory",
	},
	&cli.IntFlag{
		Name:        "port",
		Aliases:     []string{"p"},
		EnvVars:     []string{"PGPORT"},
		Value:       26000,
		Destination: &connOpts.Port,
		Usage:       "Database server port",
	},
	&cli.StringFlag{
		Name:        "username",
		Aliases:     []string{"U"},
		EnvVars:     []string{"PGUSER"},
		Destination: &connOpts.Username,
		Usage:       "Database username",
	},
	&cli.StringFlag{
		Name:        "password",
		Aliases:     []string{"W"},
		EnvVars:     []string{"PGPASSWORD"},
		Destination: &connOpts.Password,
		Usage:       "Connection password",
	},
	&cli.StringFlag{
		Name:        "database",
		Aliases:     []string{"d"},
		EnvVars:     []string{"PGDATABASE"},
		Value:       "postgres",
		Destination: &connOpts.Database,
		Usage:       "Database name to connect to",
	},
	&cli.StringFlag{
		Name:        "appname",
		EnvVars:     []string{"PGAPPNAME"},
		Value:       "gsmate",
		Destination: &connOpts.AppName,
		Usage:       "Custom application name",
	},
	&cli.DurationFlag{
		Name:        "timeout",
		EnvVars:     []string{"PGCONNECT_TIMEOUT"},
		Destination: &connOpts.Timeout,
		Value:       time.Second * 10,
		Usage:       "Connection timeout",
	},
}

func getExecutor(cCtx *cli.Context) func(string) {
	return func(in string) {
		in = strings.TrimSpace(in)
		args := strings.Fields(in)
		switch args[0] {
		case "exit", "\\q":
			os.Exit(0)
		default:
			if _, exist := subCmds[args[0]]; !exist {
				printError(fmt.Errorf("unknown command: %q", args[0]))
				return
			}
			if err := cCtx.App.Run(append([]string{"gsmate"}, args...)); err != nil {
				printError(err)
			}
		}
	}
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
	}
	app.Flags = append(app.Flags, conntionFlags...)
	app.Commands = subCmds.Values()

	app.Action = func(c *cli.Context) error {
		if c.Bool("help") {
			return cli.ShowAppHelp(c)
		}
		if c.Bool("version") {
			cli.ShowVersion(c)
			return nil
		}
		completer := newCmdCompleter()
		prompt.New(
			getExecutor(c),
			completer.Complete,
			prompt.OptionTitle(c.App.Usage),
			prompt.OptionLivePrefix(completer.PromptPrefix),
			prompt.OptionInputTextColor(prompt.Yellow),
			prompt.OptionShowCompletionAtStart(),
		).Run()
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		printError(err)
		os.Exit(1)
	}
}
