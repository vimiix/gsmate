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

package config

import (
	"fmt"
	"time"
)

type Connection struct {
	Host         string        `ini:"host,omitempty"`
	Port         int           `ini:"port,omitempty"`
	Username     string        `ini:"user,omitempty"`
	Password     string        `ini:"password,omitempty"`
	DBName       string        `ini:"dbname,omitempty"`
	AppName      string        `ini:"application_name,omitempty"`
	ConnTimeout  time.Duration `ini:"connect_timeout,omitempty"`
	QueryTimeout time.Duration `ini:"query_timeout,omitempty"`
}

func (c *Connection) Merge(other *Connection) {
	if other == nil {
		return
	}
	if other.Host != "" {
		c.Host = other.Host
	}
	if other.Port != 0 {
		c.Port = other.Port
	}
	if other.Username != "" {
		c.Username = other.Username
	}
	if other.Password != "" {
		c.Password = other.Password
	}
	if other.DBName != "" {
		c.DBName = other.DBName
	}
	if other.AppName != "" {
		c.AppName = other.AppName
	}
}

func (c *Connection) Tidy() {
	if c.Username == "" && c.DBName != "" {
		// work as psql
		c.Username = c.DBName
	}
}

func (c *Connection) Address() string {
	return fmt.Sprintf("%s@%s:%d/%s", c.Username, c.Host, c.Port, c.DBName)
}

// GetDSN returns the DSN string for connecting to the database server.
func (c *Connection) GetDSN() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable application_name=%s",
		c.Host, c.Port, c.Username, c.Password, c.DBName, c.AppName)

	if c.ConnTimeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", int(c.ConnTimeout.Seconds()))
	}

	return dsn
}
