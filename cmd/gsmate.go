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
	"time"

	"gsmate/config"
	"gsmate/internal/logger"
	"gsmate/internal/utils"
	"gsmate/pkg/client"
	"gsmate/pkg/version"

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

func main() {
	connArgs := &config.Connection{}
	app := cli.NewApp()
	app.Name = "gsmate"
	app.Usage = "Yet another openGauss client and status collector"
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
			Name:        "host",
			Aliases:     []string{"h"},
			EnvVars:     []string{"PGHOST"},
			Destination: &connArgs.Host,
			Usage:       "Database server host or socket directory",
			Required:    true,
		},
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			EnvVars:     []string{"PGPORT"},
			Value:       5432,
			Destination: &connArgs.Port,
			Usage:       "Database server port",
			Action: func(ctx *cli.Context, v int) error {
				if v > 65535 || v < 0 {
					return fmt.Errorf("flag port value %v out of range[0-65535]", v)
				}
				return nil
			},
		},
		&cli.StringFlag{
			Name:        "user",
			Aliases:     []string{"U"},
			EnvVars:     []string{"PGUSER"},
			Destination: &connArgs.Username,
			Usage:       "Database username",
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "password",
			Aliases:     []string{"W"},
			EnvVars:     []string{"PGPASSWORD"},
			Destination: &connArgs.Password,
			Usage:       "Connection password",
		},
		&cli.StringFlag{
			Name:        "dbname",
			Aliases:     []string{"d"},
			EnvVars:     []string{"PGDATABASE"},
			Value:       "postgres",
			Destination: &connArgs.DBName,
			Usage:       "Database name to connect to",
		},
		&cli.StringFlag{
			Name:        "appname",
			EnvVars:     []string{"PGAPPNAME"},
			Value:       "gsmate",
			Destination: &connArgs.AppName,
			Usage:       "Custom application name",
		},
		&cli.DurationFlag{
			Name:        "timeout",
			EnvVars:     []string{"PGCONNECT_TIMEOUT"},
			Destination: &connArgs.ConnTimeout,
			Value:       time.Second * 10,
			Usage:       "Connection timeout",
		},
	}

	app.Action = func(c *cli.Context) error {
		if c.Bool("help") {
			return cli.ShowAppHelp(c)
		}
		if c.Bool("version") {
			fmt.Println(version.GetVersionDetail())
			return nil
		}

		if err := config.Init(); err != nil {
			return err
		}

		cfg := config.Get()
		cfg.Connection.Merge(connArgs)

		if cfg.Silence {
			logger.MuteLogger()
		} else {
			logger.SetLogLevelByString(cfg.LogLevel)
		}

		dbcli, err := client.New(cfg)
		if err != nil {
			return err
		}

		return dbcli.Run()
	}
	if err := app.Run(os.Args); err != nil {
		utils.PrintError(err)
	}
}
