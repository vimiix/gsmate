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

package login

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"gsmate/internal/logger"
	"gsmate/internal/model"
	"gsmate/pkg/version"

	_ "gitee.com/opengauss/openGauss-connector-go-pq"
)

type Option struct {
	ConnOpts *model.ConnectOptions
	Prompt   string
	PingExit bool
	Pager    string
}

// Main is the entry point of the login functionality.
//
// It returns an error if there is any issue opening the database connection.
// Otherwise, it will start a interactive client for db.
func Main(ctx context.Context, opt *Option) error {
	logger.Debugf("connecting to %s ...", opt.ConnOpts.Address())
	db, err := sql.Open("opengauss", opt.ConnOpts.GetDSN())
	if err != nil {
		return err
	}
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	logger.Debug("connection established")
	defer func() {
		logger.Debug("close connection")
		_ = db.Close()
	}()

	if opt.PingExit {
		logger.Info("%s: PONG", opt.ConnOpts.Address())
		return nil
	}

	c, err := NewDBCli(db, opt)
	if err != nil {
		return err
	}
	return c.RunCli(ctx)
}

type DBCli struct {
	db            *sql.DB
	opt           *Option
	serverVersion string
	timing        bool
	hisory        []Query
}

func NewDBCli(db *sql.DB, opt *Option) (*DBCli, error) {
	c := &DBCli{db: db, opt: opt}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *DBCli) init() error {
	if err := c.queryRow(queryDBVersion, &c.serverVersion); err != nil {
		return err
	}
	logger.Debug("server version: %s", c.serverVersion)
	c.setDefaultPager()

	return nil
}

func (c *DBCli) setDefaultPager() {
	osEnvPager := os.Getenv("PAGER")
	if c.opt.Pager != "" {
		logger.Debug("default pager found in option: %s", c.opt.Pager)
		os.Setenv("PAGER", c.opt.Pager)
	} else if osEnvPager != "" {
		logger.Debug("default pager found in env: %s", osEnvPager)
		os.Setenv("PAGER", osEnvPager)
	} else {
		logger.Debug("default pager not found. Using os default pager")
	}
	if os.Getenv("LESS") == "" {
		os.Setenv("LESS", "-SRXF")
	}
}

func (c *DBCli) queryRow(q string, dest ...any) error {
	logger.Debug("query: %s", q)
	return c.db.QueryRow(q).Scan(dest...)
}

// RunCli is the interactive client for db.
func (c *DBCli) RunCli(ctx context.Context) error {
	fmt.Printf("Server: %s\n", c.serverVersion)
	fmt.Printf("Client: gsmate %s\n", version.Version)
	fmt.Println("Type \"help\" for more information.")
	defer func() {
		fmt.Println("Bye!")
	}()
	return nil
}
