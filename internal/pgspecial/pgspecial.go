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

package pgspecial

import (
	"database/sql"
	"errors"
	"gsmate/internal/pgliterals"
	"gsmate/internal/utils"
	"os"
	"sort"
	"strings"
)

var ErrCommandNotFound = errors.New("command not found")

const (
	PagerOff int = iota
	PagerLongOutput
	PagerAlways
)

var PAGER_MSG = map[int]string{
	PagerOff:        "Pager usage is off.",
	PagerLongOutput: "Pager is used for long output.",
	PagerAlways:     "Pager is always used.",
}

type ParsedSpecialCmd struct {
	Command string // Represents \?, \d, \l ....
	Args    string
	Verbose bool
}

func ParseSpecialCommand(sql string) *ParsedSpecialCmd {
	var c = &ParsedSpecialCmd{}
	parts := strings.SplitN(sql, " ", 2)
	c.Command = parts[0]
	if len(parts) > 1 {
		c.Args = parts[1]
	}
	c.Verbose = strings.Contains(c.Command, "+")
	return c
}

type CmdHandlerParams struct {
	db      *sql.DB
	pattern string
	verbose bool
}

// title, rows, headers, status
type CmdHandlerResult struct {
	Title   string
	Rows    [][]string
	Headers []string
	Status  string
}

type CmdHander func(*CmdHandlerParams) (*CmdHandlerResult, error)

type SpecialCmd struct {
	Handler       CmdHander
	Syntax        string
	Description   string
	Hidden        bool
	CaseSensitive bool
}

type Special struct {
	TimingEnabled  bool
	ExpandedOutput bool
	AutoExpand     bool
	PagerConfig    int
	Pager          string
	CmdMap         map[string]*SpecialCmd
}

func NewSpecial() *Special {
	s := &Special{
		TimingEnabled:  false,
		ExpandedOutput: false,
		AutoExpand:     false,
		PagerConfig:    PagerAlways,
		Pager:          os.Getenv("PAGER"),
		CmdMap:         make(map[string]*SpecialCmd),
	}
	s.registerDefaultCmds()
	return s
}

func (s *Special) Execute(db *sql.DB, sql string) (*CmdHandlerResult, error) {
	sc := ParseSpecialCommand(sql)
	cmd, exist := s.CmdMap[sc.Command]
	if !exist {
		cmd, exist = s.CmdMap[strings.ToLower(sc.Command)]
		if !exist || cmd.CaseSensitive {
			return nil, ErrCommandNotFound
		}
	}

	return cmd.Handler(&CmdHandlerParams{db: db, pattern: sc.Args, verbose: sc.Verbose})
}

func (s *Special) registerDefaultCmds() {
	s.Register(s.ShowCommandHelpListing, "\\?", "\\?", "Show Commands.")
}

func (s *Special) ShowCommandHelpListing(_ *CmdHandlerParams) (*CmdHandlerResult, error) {
	var cmds []string
	for cmd := range pgliterals.PGCommands {
		cmds = append(cmds, cmd)
	}
	sort.Strings(cmds)
	rows := utils.Chunks(cmds)
	return &CmdHandlerResult{Rows: rows}, nil
}

func (s *Special) ShowCommandHelp(pattern string) {
	cmd := strings.ToUpper(strings.TrimSpace(pattern))
	if cmd == "" {
		s.ShowCommandHelpListing(nil)
		return
	}
}

type Option struct {
	Hidden        bool
	CaseSensitive bool
	Aliases       []string
}

func NewOption() *Option {
	return &Option{
		Hidden:        false,
		CaseSensitive: true,
		Aliases:       []string{},
	}
}

type OptionFunc func(*Option)

func WithHidden(hidden bool) OptionFunc {
	return func(opt *Option) {
		opt.Hidden = hidden
	}
}

func WithCaseSensitive(caseSensitive bool) OptionFunc {
	return func(opt *Option) {
		opt.CaseSensitive = caseSensitive
	}
}

func WithAliases(aliases []string) OptionFunc {
	return func(opt *Option) {
		opt.Aliases = aliases
	}
}

func (s *Special) Register(
	hander CmdHander,
	command string,
	syntax string,
	description string,
	optFuncs ...OptionFunc,
) {
	opt := NewOption()
	for _, f := range optFuncs {
		f(opt)
	}
	if !opt.CaseSensitive {
		command = strings.ToLower(command)
	}
	s.CmdMap[command] = &SpecialCmd{
		Handler:       hander,
		Syntax:        syntax,
		Description:   description,
		Hidden:        opt.Hidden,
		CaseSensitive: opt.CaseSensitive,
	}
	for _, alias := range opt.Aliases {
		if !opt.CaseSensitive {
			alias = strings.ToLower(alias)
		}
		s.CmdMap[alias] = &SpecialCmd{
			Handler:       hander,
			Syntax:        syntax,
			Description:   description,
			Hidden:        true,
			CaseSensitive: opt.CaseSensitive,
		}
	}
}
