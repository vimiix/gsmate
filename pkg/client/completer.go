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

package client

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"gsmate/config"
	"gsmate/internal/errdef"
	"gsmate/internal/logger"
	"gsmate/pkg/client/metadata"

	"github.com/vimiix/go-prompt"
)

const WORD_BREAKS = "\t\n$><=;|&{() "

type caseType bool

var (
	IGNORE_CASE = caseType(true)
	MATCH_CASE  = caseType(false)
)

type CmdCompleter struct {
	client *DBClient
}

func (c *CmdCompleter) Complete() prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		var i int
		start := d.CursorPositionCol()
		preText := []rune(d.TextBeforeCursor())
		for i = start - 1; i >= 0; i-- {
			if strings.ContainsRune(WORD_BREAKS, preText[i]) {
				i++
				break
			}
		}
		if i == -1 {
			i = 0
		}
		previousWords := getPreviousWords(start, preText)
		text := preText[i:start]
		return c.complete(previousWords, text)
	}
}

func (c *CmdCompleter) complete(previousWords []string, text []rune) []prompt.Suggest {
	if len(text) > 0 {
		if len(previousWords) == 0 && text[0] == '\\' {
			/* If current word is a backslash command, offer completions for that */
			return c.completeFromListCase(MATCH_CASE, text, backslashCommands...)
		}
		if text[0] == ':' {
			if len(text) == 1 || text[1] == ':' {
				return nil
			}
			/* If current word is a variable interpolation, handle that case */
			if text[1] == '\'' {
				return c.completeFromVariables(text, ":'", "'", true)
			}
			if text[1] == '"' {
				return c.completeFromVariables(text, ":\"", "\"", true)
			}
			return c.completeFromVariables(text, ":", "", true)
		}

	}

	if len(previousWords) == 0 && len(text) > 0 {
		return c.completeFromListCase(IGNORE_CASE, text, getStartSQLCmdSuggests()...)
	}

	if len(previousWords) == 1 {
		candidates := startSQLCommands[strings.ToUpper(previousWords[0])]
		if candidates != nil {
			return c.completeFromStrList(text, candidates...)
		}
	}

	/* Complete DELETE FROM with a list of tables */
	if TailMatches(IGNORE_CASE, previousWords, "DELETE", "FROM") {
		return c.completeWithUpdatables(text)
	}
	/* Complete DELETE FROM <table> */
	if TailMatches(IGNORE_CASE, previousWords, "DELETE", "FROM", "*") {
		return c.completeFromStrList(text, "USING", "WHERE")
	}

	/* Complete CREATE */
	if TailMatches(IGNORE_CASE, previousWords, "CREATE") {
		return c.completeFromStrList(text, "DATABASE", "SCHEMA", "SEQUENCE", "TABLE", "VIEW", "TEMPORARY")
	}
	if TailMatches(IGNORE_CASE, previousWords, "CREATE", "TEMP|TEMPORARY") {
		return c.completeFromStrList(text, "TABLE", "VIEW")
	}
	if TailMatches(IGNORE_CASE, previousWords, "CREATE", "TABLE", "*") || TailMatches(IGNORE_CASE, previousWords, "CREATE", "TEMP|TEMPORARY", "TABLE", "*") {
		return c.completeFromStrList(text, "(")
	}
	/* Complete INSERT INTO with table names */
	if TailMatches(IGNORE_CASE, previousWords, "INSERT", "INTO") {
		return c.completeWithUpdatables(text)
	}
	/*
	 * Complete INSERT INTO <table> with "(" or "VALUES" or "SELECT" or
	 * "TABLE" or "DEFAULT VALUES" or "OVERRIDING"
	 */
	if TailMatches(IGNORE_CASE, previousWords, "INSERT", "INTO", "*") {
		return c.completeFromStrList(text, "(", "DEFAULT VALUES", "SELECT", "TABLE", "VALUES", "OVERRIDING")
	}
	/*
	 * Complete INSERT INTO <table> (attribs) with "VALUES" or "SELECT" or
	 * "TABLE" or "OVERRIDING"
	 */
	if TailMatches(IGNORE_CASE, previousWords, "INSERT", "INTO", "*", "*") &&
		strings.HasSuffix(previousWords[0], ")") {
		return c.completeFromStrList(text, "SELECT", "TABLE", "VALUES", "OVERRIDING")
	}

	/* Complete OVERRIDING */
	if TailMatches(IGNORE_CASE, previousWords, "OVERRIDING") {
		return c.completeFromStrList(text, "SYSTEM VALUE", "USER VALUE")
	}

	/* Complete after OVERRIDING clause */
	if TailMatches(IGNORE_CASE, previousWords, "OVERRIDING", "*", "VALUE") {
		return c.completeFromStrList(text, "SELECT", "TABLE", "VALUES")
	}

	/* Insert an open parenthesis after "VALUES" */
	if TailMatches(IGNORE_CASE, previousWords, "VALUES") && !TailMatches(IGNORE_CASE, previousWords, "DEFAULT", "VALUES") {
		return c.completeFromStrList(text, "(")
	}
	/* UPDATE --- can be inside EXPLAIN, RULE, etc */
	/* If prev. word is UPDATE suggest a list of tables */
	if TailMatches(IGNORE_CASE, previousWords, "UPDATE") {
		return c.completeWithUpdatables(text)
	}
	/* Complete UPDATE <table> with "SET" */
	if TailMatches(IGNORE_CASE, previousWords, "UPDATE", "*") {
		return c.completeFromStrList(text, "SET")
	}
	/* UPDATE <table> SET <attr> = */
	if TailMatches(IGNORE_CASE, previousWords, "UPDATE", "*", "SET", "!*=") {
		return c.completeFromStrList(text, "=")
	}

	if TailMatches(IGNORE_CASE, previousWords, "SELECT", "*") {
		return c.completeFromStrList(text, "FROM")
	}

	/* ... FROM | JOIN ... */
	if TailMatches(IGNORE_CASE, previousWords, "FROM|JOIN") {
		return c.completeWithUpdatables(text)
	}
	/* TABLE, but not TABLE embedded in other commands */
	if matches(IGNORE_CASE, previousWords, "TABLE") {
		return c.completeWithUpdatables(text)
	}
	/* Backslash commands */
	// if TailMatches(MATCH_CASE, previousWords, `\cd|\e|\edit|\g|\gx|\i|\include|\ir|\include_relative|\o|\out|\s|\w|\write`) {
	// 	return completeFromFiles(text)
	// }
	if TailMatches(MATCH_CASE, previousWords, `\copy`, `*`, `*`) {
		return nil
	}
	// if TailMatches(MATCH_CASE, previousWords, `\da*`) {
	// 	return c.completeWithFunctions(text, []string{"AGGREGATE"})
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\df*`) {
	// 	return c.completeWithFunctions(text, []string{})
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\di*`) {
	// 	return c.completeWithIndexes(text)
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\dn*`) {
	// 	return c.completeWithSchemas(text)
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\ds*`) {
	// 	return c.completeWithSequences(text)
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\dt*`) {
	// 	return c.completeWithTables(text, []string{"TABLE", "BASE TABLE", "SYSTEM TABLE", "SYNONYM", "LOCAL TEMPORARY", "GLOBAL TEMPORARY"})
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\dv*`) {
	// 	return c.completeWithTables(text, []string{"VIEW", "SYSTEM VIEW"})
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\dm*`) {
	// 	return c.completeWithTables(text, []string{"MATERIALIZED VIEW"})
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\d*`) {
	// 	return c.completeWithSelectables(text)
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\l*`) ||
	// 	TailMatches(MATCH_CASE, previousWords, `\lo*`) {
	// 	return c.completeWithCatalogs(text)
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\pset`) {
	// 	return c.completeFromStrList(text, `border`, `columns`, `expanded`, `fieldsep`, `fieldsep_zero`,
	// 		`footer`, `format`, `linestyle`, `null`, `numericlocale`, `pager`, `pager_min_lines`,
	// 		`recordsep`, `recordsep_zero`, `tableattr`, `title`, `title`, `tuples_only`,
	// 		`unicode_border_linestyle`, `unicode_column_linestyle`, `unicode_header_linestyle`)
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\pset`, `expanded`) {
	// 	return c.completeFromStrList(text, "auto", "on", "off")
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\pset`, `pager`) {
	// 	return c.completeFromStrList(text, "always", "on", "off")
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\pset`, `fieldsep_zero|footer|numericlocale|pager|recordsep_zero|tuples_only`) {
	// 	return c.completeFromStrList(text, "on", "off")
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\pset`, `format`) {
	// 	return c.completeFromStrList(text, "unaligned", "aligned", "wrapped", "html", "asciidoc", "latex", "latex-longtable", "troff-ms", "csv", "json", "vertical")
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\pset`, `linestyle`) {
	// 	return c.completeFromStrList(text, "ascii", "old-ascii", "unicode")
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\pset`, `unicode_border_linestyle|unicode_column_linestyle|unicode_header_linestyle`) {
	// 	return c.completeFromStrList(text, "single", "double")
	// }
	// if TailMatches(MATCH_CASE, previousWords, `\pset`, `*`) ||
	// 	TailMatches(MATCH_CASE, previousWords, `\pset`, `*`, `*`) {
	// 	return nil
	// }
	if TailMatches(MATCH_CASE, previousWords, `\?`) {
		return c.completeFromStrList(text, "commands", "options", "variables")
	}
	// is suggesting basic sql commands better than nothing?
	return nil
}

func (c *CmdCompleter) completeFromStrList(text []rune, options ...string) []prompt.Suggest {
	candidates := make([]prompt.Suggest, 0, len(options))
	for _, o := range options {
		candidates = append(candidates, prompt.Suggest{Text: o})
	}
	return c.completeFromListCase(IGNORE_CASE, text, candidates...)
}

func (c *CmdCompleter) completeFromListCase(ct caseType, text []rune, options ...prompt.Suggest) []prompt.Suggest {
	if len(options) == 0 {
		return nil
	}
	prefix := string(text)
	if ct == IGNORE_CASE {
		prefix = strings.ToUpper(prefix)
	}
	result := make([]prompt.Suggest, 0, len(options))
	for _, o := range options {
		if (ct == IGNORE_CASE && !strings.HasPrefix(strings.ToUpper(o.Text), prefix)) ||
			(ct == MATCH_CASE && !strings.HasPrefix(o.Text, prefix)) {
			continue
		}
		result = append(result, o)
	}
	return result
}

func (c *CmdCompleter) completeFromVariables(text []rune, prefix, suffix string, needValue bool) []prompt.Suggest {
	cfgs := config.GetConfigMap()
	names := make([]prompt.Suggest, 0, len(cfgs))
	for name, value := range cfgs {
		if needValue && value == "" {
			continue
		}
		names = append(names, prompt.Suggest{
			Text: fmt.Sprintf("%s%s%s", prefix, name, suffix),
		})
	}
	return c.completeFromListCase(MATCH_CASE, text, names...)
}

// TailMatches when last words match all patterns
func TailMatches(ct caseType, words []string, patterns ...string) bool {
	if len(words) < len(patterns) {
		return false
	}
	for i, p := range patterns {
		if !wordMatches(ct, p, words[len(patterns)-i-1]) {
			return false
		}
	}
	return true
}

func wordMatches(ct caseType, pattern, word string) bool {
	if pattern == "*" {
		return true
	}

	if pattern[0] == '!' {
		return !wordMatches(ct, pattern[1:], word)
	}

	cmp := func(a, b string) bool { return a == b }
	if ct == IGNORE_CASE {
		cmp = strings.EqualFold
	}

	for _, p := range strings.Split(pattern, "|") {
		star := strings.IndexByte(p, '*')
		if star == -1 {
			if cmp(p, word) {
				return true
			}
		} else {
			if len(word) >= len(p)-1 && cmp(p[0:star], word[0:star]) && (star >= len(p) || cmp(p[star+1:], word[len(word)-len(p)+star+1:])) {
				return true
			}
		}
	}

	return false
}

func matches(ct caseType, words []string, patterns ...string) bool {
	if len(words) != len(patterns) {
		return false
	}
	for i, p := range patterns {
		if !wordMatches(ct, p, words[len(patterns)-i-1]) {
			return false
		}
	}
	return true
}

func getPreviousWords(point int, buf []rune) []string {
	var i int

	/*
	 * Allocate a slice of strings (rune slices). The worst case is that the line contains only
	 * non-whitespace WORD_BREAKS characters, making each one a separate word.
	 * This is usually much more space than we need, but it's cheaper than
	 * doing a separate malloc() for each word.
	 */
	previousWords := make([]string, 0, point*2)

	/*
	 * First we look for a non-word char before the current point.  (This is
	 * probably useless, if readline is on the same page as we are about what
	 * is a word, but if so it's cheap.)
	 */
	for i = point - 1; i >= 0; i-- {
		if strings.ContainsRune(WORD_BREAKS, buf[i]) {
			break
		}
	}
	point = i

	/*
	 * Now parse words, working backwards, until we hit start of line.  The
	 * backwards scan has some interesting but intentional properties
	 * concerning parenthesis handling.
	 */
	for point >= 0 {
		var start, end int
		inquotes := false
		parentheses := 0

		/* now find the first non-space which then constitutes the end */
		end = -1
		for i = point; i >= 0; i-- {
			if !unicode.IsSpace(buf[i]) {
				end = i
				break
			}
		}
		/* if no end found, we're done */
		if end < 0 {
			break
		}

		/*
		 * Otherwise we now look for the start.  The start is either the last
		 * character before any word-break character going backwards from the
		 * end, or it's simply character 0.  We also handle open quotes and
		 * parentheses.
		 */
		for start = end; start > 0; start-- {
			if buf[start] == '"' {
				inquotes = !inquotes
			}
			if inquotes {
				continue
			}
			if buf[start] == ')' {
				parentheses++
			} else if buf[start] == '(' {
				parentheses -= 1
				if parentheses <= 0 {
					break
				}
			} else if parentheses == 0 && strings.ContainsRune(WORD_BREAKS, buf[start-1]) {
				break
			}
		}

		/* Return the word located at start to end inclusive */
		i = end - start + 1
		previousWords = append(previousWords, string(buf[start:start+i]))

		/* Continue searching */
		point = start - 1
	}

	return previousWords
}

func (c *CmdCompleter) completeWithUpdatables(text []rune) []prompt.Suggest {
	filter := parseIdentifier(string(text))

	// exclude materialized views, sequences, system tables, synonyms
	filter.Types = []string{"TABLE", "BASE TABLE", "LOCAL TEMPORARY", "GLOBAL TEMPORARY", "VIEW"}
	names := c.getNames(
		func() (iterator, error) {
			return c.client.Tables(filter)
		},
		func(res interface{}) string {
			t := res.(*metadata.TableSet).Get()
			return qualifiedIdentifier(filter, t.Schema, t.Name)
		},
	)
	sort.Strings(names)
	return c.completeFromStrList(text, names...)
}

type iterator interface {
	Next() bool
	Close() error
}

func (c *CmdCompleter) getNames(query func() (iterator, error), mapper func(interface{}) string) []string {
	res, err := query()
	if err != nil {
		if err != errdef.ErrNotSupported {
			logger.Error("complete error: %s", err)
		}
		return nil
	}
	defer res.Close()

	// there can be duplicates if names are not qualified
	values := make(map[string]struct{}, 10)
	for res.Next() {
		values[mapper(res)] = struct{}{}
	}
	result := make([]string, 0, len(values))
	for v := range values {
		result = append(result, v)
	}
	return result
}

func qualifiedIdentifier(filter metadata.Filter, schema, name string) string {
	// TODO handle quoted identifiers
	if filter.Schema != "" {
		return schema + "." + name
	}
	return name
}

// parseIdentifier into catalog, schema and name
func parseIdentifier(name string) metadata.Filter {
	result := metadata.Filter{}
	if !strings.ContainsRune(name, '.') {
		result.Name = name + "%"
		result.OnlyVisible = true
	} else {
		parts := strings.SplitN(name, ".", 2)
		result.Schema = parts[0]
		result.Name = parts[1] + "%"
	}

	if result.Schema != "" || len(result.Name) > 3 {
		result.WithSystem = true
	}
	return result
}
