package client

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

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gsmate/config"
	"gsmate/internal/errdef"
	"gsmate/internal/logger"
	"gsmate/pkg/client/metacmd"
	"gsmate/pkg/client/metadata"
	"gsmate/pkg/version"

	_ "gitee.com/opengauss/openGauss-connector-go-pq"
	"github.com/pkg/errors"
	"github.com/vimiix/go-prompt"
	"github.com/xo/tblfmt"
)

var dummyExecutor = func(string) {}

type DBClient struct {
	cfg          *config.Config
	db           *sql.DB
	tx           *sql.Tx
	version      string
	prompt       *prompt.Prompt
	promptPrefix string
	history      *History
	stmt         *Stmt
}

func New(cfg *config.Config) (*DBClient, error) {
	db, err := sql.Open("opengauss", cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	history, err := NewHistory(cfg.MaxHistory)
	if err != nil {
		return nil, err
	}
	c := &DBClient{
		cfg:     cfg,
		db:      db,
		history: history,
	}

	cc := &CmdCompleter{client: c}

	c.prompt = prompt.New(dummyExecutor,
		cc.Complete(),
		prompt.OptionTitle("gsmate"),
		prompt.OptionHistory(history.Records()),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionLivePrefix(c.LivePrefix()),
	)

	c.stmt = NewStmt(func() ([]rune, error) {
		s, err := c.prompt.Input()
		if err != nil {
			return nil, err
		}
		c.history.Add(s)
		return []rune(s), nil
	})

	err = c.initServerInfo()
	return c, err
}

func (c *DBClient) LivePrefix() func() (string, bool) {
	if c.promptPrefix == "" {
		c.promptPrefix = c.cfg.PromptPrefix()
	}
	return func() (string, bool) {
		status := " => "
		if len(c.stmt.Buf) > 0 && !c.stmt.ready {
			status = " -> "
		}
		return c.promptPrefix + status, true
	}
}

func (c *DBClient) initServerInfo() error {
	var err error
	rows, err := c.DB().Query(queryDBVersion)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		if err = rows.Scan(&c.version); err != nil {
			return err
		}
	}

	logger.Debug("get server version: %s", c.version)
	return err
}

func (c *DBClient) DB() DB {
	if c.tx != nil {
		return c.tx
	}
	return c.db
}

type CloseFunc func()

func (c *DBClient) query(q string, args ...any) (*sql.Rows, CloseFunc, error) {
	logger.Debug("query: %s", q)
	if c.cfg.QueryTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.TODO(), c.cfg.QueryTimeout)
		rows, err := c.DB().QueryContext(ctx, q, args...)
		return rows, func() { cancel(); rows.Close() }, err
	}
	rows, err := c.DB().Query(q, args...)
	return rows, func() { rows.Close() }, err
}

func (c *DBClient) Query(qstr string, conds []string, order string, vals ...any) (*sql.Rows, CloseFunc, error) {
	if len(conds) != 0 {
		qstr += "\nWHERE " + strings.Join(conds, " AND ")
	}
	if order != "" {
		qstr += "\nORDER BY " + order
	}
	return c.query(qstr, vals...)
}

// helpQuitExitRE is a regexp to use to match help, quit, or exit messages.
var helpQuitExitRE = regexp.MustCompile(`(?im)^(help|quit|exit)\s*$`)

// RunCli is the interactive client for db.
func (c *DBClient) Run() error {
	defer func() {
		_ = c.history.Persist()
	}()

	if !c.cfg.LessChatty {
		fmt.Printf("Server: %s\n", c.version)
		fmt.Printf("Client: gsmate %s (%s)\n", version.Version, version.Commit)
		fmt.Println(`Type "help" for more information.`)
		fmt.Println()
	}

	for {
		/*cmd, paramstr,*/ _, _, err := c.stmt.Next(Unquote)
		if err != nil {
			if errors.Is(err, prompt.ErrQuit) {
				return nil
			}
			return err
		}

		var opt metacmd.Option

		// help, exit, quit intercept
		if len(c.stmt.Buf) >= 4 {
			i, first := RunesLastIndex(c.stmt.Buf, '\n'), false
			if i == -1 {
				i, first = 0, true
			}
			if s := strings.ToLower(helpQuitExitRE.FindString(string(c.stmt.Buf[i:]))); s != "" {
				switch s {
				case "help":
					s = `Use \? for help or press ctrl-C to clear the input buffer.`
					if first {
						s = `TODO: help message`
						c.stmt.Reset(nil)
					}
				case "quit", "exit":
					s = `Use \q or ctrl-D to quit.`
					if first {
						return nil
					}
				}
				fmt.Fprintln(os.Stdout, s)
			}
		}

		if opt.Quit {
			return nil
		}

		// FIXME
		if c.stmt.Ready() || opt.Exec != metacmd.ExecNone {
			err = c.doQuery(c.stmt.String())
			if err != nil {
				logger.Error("query error: %v", err)
			} else {
				logger.Debug("reset statement")
			}
			c.stmt.Reset(nil)
		}
	}
}

func (c *DBClient) doQuery(q string, args ...any) error {
	rows, closeFunc, err := c.query(q, args...)
	if err != nil {
		return err
	}
	defer closeFunc()
	params := config.GetPrintConfig()
	resultSet := tblfmt.ResultSet(rows)
	return tblfmt.EncodeAll(os.Stdout, resultSet, params)
}

func (c *DBClient) Catalogs(f metadata.Filter) (*metadata.CatalogSet, error) {
	qstr := `SELECT d.datname as "Name",
       pg_catalog.pg_get_userbyid(d.datdba) as "Owner",
       pg_catalog.pg_encoding_to_char(d.encoding) as "Encoding",
       d.datcollate as "Collate",
       d.datctype as "Ctype",
       COALESCE(pg_catalog.array_to_string(d.datacl, E'\n'),'') AS "Access privileges"
FROM pg_catalog.pg_database d`
	rows, closeFunc, err := c.Query(qstr, []string{}, "1")
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	var results []metadata.Result
	for rows.Next() {
		rec := metadata.Catalog{}
		err = rows.Scan(&rec.Catalog, &rec.Owner, &rec.Encoding, &rec.Collate, &rec.Ctype, &rec.AccessPrivileges)
		if err != nil {
			return nil, err
		}
		results = append(results, &rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewCatalogSet(results), nil
}

func (c *DBClient) Tables(f metadata.Filter) (*metadata.TableSet, error) {
	qstr := `SELECT n.nspname as "Schema",
  c.relname as "Name",
  CASE c.relkind WHEN 'r' THEN 'table' WHEN 'v' THEN 'view' WHEN 'm' THEN 'materialized view' WHEN 'i' THEN 'index' WHEN 'S' THEN 'sequence' WHEN 's' THEN 'special' WHEN 'f' THEN 'foreign table' WHEN 'p' THEN 'partitioned table' WHEN 'I' THEN 'partitioned index' ELSE 'unknown' END as "Type",
  COALESCE((c.reltuples / NULLIF(c.relpages, 0)) * (pg_catalog.pg_relation_size(c.oid) / current_setting('block_size')::int), 0)::bigint as "Rows",
  pg_catalog.pg_size_pretty(pg_catalog.pg_table_size(c.oid)) as "Size",
  COALESCE(pg_catalog.obj_description(c.oid, 'pg_class'), '') as "Description"
FROM pg_catalog.pg_class c
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
`
	conds := []string{"n.nspname !~ '^pg_toast' AND c.relkind != 'c'"}
	vals := []interface{}{}
	if f.OnlyVisible {
		conds = append(conds, "pg_catalog.pg_table_is_visible(c.oid)")
	}
	if !f.WithSystem {
		conds = append(conds, "n.nspname NOT IN ('pg_catalog', 'information_schema')")
	}
	if f.Schema != "" {
		vals = append(vals, f.Schema)
		conds = append(conds, fmt.Sprintf("n.nspname LIKE $%d", len(vals)))
	}
	if f.Name != "" {
		vals = append(vals, f.Name)
		conds = append(conds, fmt.Sprintf("c.relname LIKE $%d", len(vals)))
	}
	if len(f.Types) != 0 {
		tableTypes := map[string][]rune{
			"TABLE":             {'r', 'p', 's', 'f'},
			"VIEW":              {'v'},
			"MATERIALIZED VIEW": {'m'},
			"SEQUENCE":          {'S'},
		}
		pholders := []string{"''"}
		for _, t := range f.Types {
			for _, k := range tableTypes[t] {
				vals = append(vals, string(k))
				pholders = append(pholders, fmt.Sprintf("$%d", len(vals)))
			}
		}
		conds = append(conds, fmt.Sprintf("c.relkind IN (%s)", strings.Join(pholders, ", ")))
	}
	rows, closeFunc, err := c.Query(qstr, conds, "1, 3, 2", vals...)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.NewTableSet([]metadata.Table{}), nil
		}
		return nil, err
	}
	defer closeFunc()

	results := []metadata.Table{}
	for rows.Next() {
		rec := metadata.Table{}
		err = rows.Scan(&rec.Schema, &rec.Name, &rec.Type, &rec.Rows, &rec.Size, &rec.Comment)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return metadata.NewTableSet(results), nil
}

func (c *DBClient) Functions(f metadata.Filter) (*metadata.FunctionSet, error) {
	return nil, errdef.ErrNotSupported
}
