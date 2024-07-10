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
	"gsmate/internal/model"
	"gsmate/pkg/login"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	connOpts      = &model.ConnectOptions{}
	conntionFlags = []cli.Flag{
		&cli.StringFlag{
			Name:        "host",
			Aliases:     []string{"h"},
			EnvVars:     []string{"PGHOST"},
			Destination: &connOpts.Host,
			Usage:       "Database server host or socket directory",
			Required:    true,
		},
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			EnvVars:     []string{"PGPORT"},
			Value:       5432,
			Destination: &connOpts.Port,
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
			Destination: &connOpts.Username,
			Usage:       "Database username",
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "password",
			Aliases:     []string{"W"},
			EnvVars:     []string{"PGPASSWORD"},
			Destination: &connOpts.Password,
			Usage:       "Connection password",
		},
		&cli.StringFlag{
			Name:        "dbname",
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
)

func newLoginCmd() *cli.Command {
	opt := &login.Option{
		ConnOpts: connOpts,
	}
	cmd := newDefaultCmd()
	cmd.Name = "login"
	cmd.Usage = "Login to the database server"
	cmd.Flags = append(cmd.Flags, conntionFlags...)
	cmd.Flags = append(cmd.Flags, &cli.StringFlag{
		Name: "prompt",
		Usage: "Prompt format, support macros: \n" +
			"{host}, {user}, {db}, {port}, {schema}, {client_pid}, {server_pid}\n",
		Value:       "{user}@{host}/{db}> ",
		Destination: &opt.Prompt,
	}, &cli.BoolFlag{
		Name:        "ping",
		Usage:       "Check database connectivity, then exit",
		Destination: &opt.PingExit,
	})
	cmd.Action = func(c *cli.Context) error {
		connOpts.Tidy()
		return login.Main(c.Context, opt)
	}
	return cmd
}
