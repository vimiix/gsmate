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

package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ConnectOptions struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	AppName  string
	Timeout  time.Duration
	DSN      string
}

func (o *ConnectOptions) Tidy() {
	if o.Username == "" && o.Database != "" {
		// work as psql
		o.Username = o.Database
	}
}

func (o *ConnectOptions) Address() string {
	return fmt.Sprintf("%s@%s:%d/%s", o.Username, o.Host, o.Port, o.Database)
}

// GetDSN returns the DSN string for connecting to the database server.
func (o *ConnectOptions) GetDSN() string {
	if o.DSN != "" {
		return o.DSN
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable application_name=%s",
		o.Host, o.Port, o.Username, o.Password, o.Database, o.AppName)

	if o.Timeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", int(o.Timeout.Seconds()))
	}

	return dsn
}

type NotifyCallback func(string)

type Connection struct {
	DB             *sql.DB
	Opt            *ConnectOptions
	ServerVersion  string
	ServerPid      int
	IsSuperuser    bool
	NotifyCallback NotifyCallback
}

func NewConnection(ctx context.Context, db *sql.DB, opt *ConnectOptions) (*Connection, error) {

	c := &Connection{
		DB:  db,
		Opt: opt,
	}
	_, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return c, nil
}
