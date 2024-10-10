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
	"io"
	"reflect"
	"strings"
	"testing"
)

func sl(n int, r rune) string {
	z := make([]rune, n)
	for i := 0; i < n; i++ {
		z[i] = r
	}
	return string(z)
}

func TestAppend(t *testing.T) {
	a512 := sl(512, 'a')
	// b1024 := sl(1024, 'b')
	tests := []struct {
		s   []string
		exp string
		l   int
		c   int
	}{
		{[]string{""}, "", 0, 0},
		{[]string{"", ""}, "\n", 1, MinCapIncrease},
		{[]string{"", "", ""}, "\n\n", 2, MinCapIncrease},
		{[]string{"", "", "", ""}, "\n\n\n", 3, MinCapIncrease},
		{[]string{"a", ""}, "a\n", 2, 2}, // 4
		{[]string{"a", "b", ""}, "a\nb\n", 4, MinCapIncrease},
		{[]string{"a", "b", "c", ""}, "a\nb\nc\n", 6, MinCapIncrease},
		{[]string{"", "a", ""}, "\na\n", 3, MinCapIncrease}, // 7
		{[]string{"", "a", "b", ""}, "\na\nb\n", 5, MinCapIncrease},
		{[]string{"", "a", "b", "c", ""}, "\na\nb\nc\n", 7, MinCapIncrease},
		{[]string{"", "foo"}, "\nfoo", 4, MinCapIncrease}, // 10
		{[]string{"", "foo", ""}, "\nfoo\n", 5, MinCapIncrease},
		{[]string{"foo", "", "bar"}, "foo\n\nbar", 8, MinCapIncrease},
		{[]string{"", "foo", "bar"}, "\nfoo\nbar", 8, MinCapIncrease},
		{[]string{a512}, a512, 512, 512}, // 14
		{[]string{a512, a512}, a512 + "\n" + a512, 1025, 5 * MinCapIncrease},
		{[]string{a512, a512, a512}, a512 + "\n" + a512 + "\n" + a512, 1538, 5 * MinCapIncrease},
		{[]string{a512, ""}, a512 + "\n", 513, 2 * MinCapIncrease}, // 17
		{[]string{a512, "", "foo"}, a512 + "\n\nfoo", 517, 2 * MinCapIncrease},
	}
	for i, test := range tests {
		b := new(Stmt)
		for _, s := range test.s {
			b.AppendString(s, "\n")
		}
		if s := b.String(); s != test.exp {
			t.Errorf("test %d expected result of `%s`, got: `%s`", i, test.exp, s)
		}
		if b.Len != test.l {
			t.Errorf("test %d expected resulting len of %d, got: %d", i, test.l, b.Len)
		}
		if c := cap(b.Buf); c != test.c {
			t.Errorf("test %d expected resulting cap of %d, got: %d", i, test.c, c)
		}
		b.Reset(nil)
		if b.Len != 0 {
			t.Errorf("test %d expected after reset len of 0, got: %d", i, b.Len)
		}
		b.AppendString("", "\n")
		if s := b.String(); s != "" {
			t.Errorf("test %d expected after reset appending an empty string would result in empty string, got: `%s`", i, s)
		}
	}
}

func TestVariedSeparator(t *testing.T) {
	b := new(Stmt)
	b.AppendString("foo", "\n")
	b.AppendString("foo", "bar")
	if b.Len != 9 {
		t.Errorf("expected len of 9, got: %d", b.Len)
	}
	if s := b.String(); s != "foobarfoo" {
		t.Errorf("expected `%s`, got: `%s`", "foobarfoo", s)
	}
	if c := cap(b.Buf); c != MinCapIncrease {
		t.Errorf("expected cap of %d, got: %d", MinCapIncrease, c)
	}
}

func TestNextResetState(t *testing.T) {
	tests := []struct {
		s     string
		stmts []string
		cmds  []string
		state string
		vars  []string
	}{
		{``, nil, []string{`|`}, `=`, nil}, // 0
		{`;`, []string{`;`}, []string{`|`}, `=`, nil},
		{` ; `, []string{`;`}, []string{`|`, `|`}, `=`, nil},
		{` \v `, nil, []string{`\v| `}, `=`, nil},
		{` \v \p`, nil, []string{`\v| `, `\p|`}, `=`, nil},
		{` \v   foo   \p`, nil, []string{`\v|   foo   `, `\p|`}, `=`, nil}, // 5
		{` \v   foo   bar  \p   zz`, nil, []string{`\v|   foo   bar  `, `\p|   zz`}, `=`, nil},
		{` \very   foo   bar  \print   zz`, nil, []string{`\very|   foo   bar  `, `\print|   zz`}, `=`, nil},
		{`select 1;`, []string{`select 1;`}, []string{`|`}, `=`, nil},
		{`select 1\g`, []string{`select 1`}, []string{`\g|`}, `=`, nil},
		{`select 1 \g`, []string{`select 1 `}, []string{`\g|`}, `=`, nil}, // 10
		{` select 1 \g`, []string{`select 1 `}, []string{`\g|`}, `=`, nil},
		{` select 1   \g  `, []string{`select 1   `}, []string{`\g|  `}, `=`, nil},
		{`select 1; select 1\g`, []string{`select 1;`, `select 1`}, []string{`|`, `\g|`}, `=`, nil},
		{"select 1\n\\g", []string{`select 1`}, []string{`|`, `\g|`}, `=`, nil},
		{"select 1 \\g\n\n\n\n\\v", []string{`select 1 `}, []string{`\g|`, `|`, `|`, `|`, `\v|`}, `=`, nil}, // 15
		{"select 1 \\g\n\n\n\n\\v foob \\p zzz \n\n", []string{`select 1 `}, []string{`\g|`, `|`, `|`, `|`, `\v| foob `, `\p| zzz `, `|`, `|`}, `=`, nil},
		{" select 1 \\g \\p \n select (15)\\g", []string{`select 1 `, `select (15)`}, []string{`\g| `, `\p| `, `\g|`}, `=`, nil},
		{" select 1 (  \\g ) \n ;", []string{"select 1 (  \\g ) \n ;"}, []string{`|`, `|`}, `=`, nil},
		{ // 19
			" select 1\n;select 2\\g  select 3;  \\p   \\z  foo bar ",
			[]string{"select 1\n;", "select 2"},
			[]string{`|`, `|`, `\g|  select 3;  `, `\p|   `, `\z|  foo bar `},
			"=", nil,
		},
		{ // 20
			" select 1\\g\n\n\tselect 2\\g\n select 3;  \\p   \\z  foo bar \\p\\p select * from;  \n\\p",
			[]string{`select 1`, `select 2`, `select 3;`},
			[]string{`\g|`, `|`, `\g|`, `|`, `\p|   `, `\z|  foo bar `, `\p|`, `\p| select * from;  `, `\p|`},
			"=", nil,
		},
		{"select '';", []string{"select '';"}, []string{"|"}, "=", nil}, // 21
		{"select 'a''b\nz';", []string{"select 'a''b\nz';"}, []string{"|", "|"}, "=", nil},
		{"select 'a' 'b\nz';", []string{"select 'a' 'b\nz';"}, []string{"|", "|"}, "=", nil},
		{"select \"\";", []string{"select \"\";"}, []string{"|"}, "=", nil},
		{"select \"\n\";", []string{"select \"\n\";"}, []string{"|", "|"}, "=", nil}, // 25
		{"select $$$$;", []string{"select $$$$;"}, []string{"|"}, "=", nil},
		{"select $$\nfoob(\n$$;", []string{"select $$\nfoob(\n$$;"}, []string{"|", "|", "|"}, "=", nil},
		{"select $tag$$tag$;", []string{"select $tag$$tag$;"}, []string{"|"}, "=", nil},
		{"select $tag$\n\n$tag$;", []string{"select $tag$\n\n$tag$;"}, []string{"|", "|", "|"}, "=", nil},
		{"select $tag$\n(\n$tag$;", []string{"select $tag$\n(\n$tag$;"}, []string{"|", "|", "|"}, "=", nil}, // 30
		{"select $tag$\n\\v(\n$tag$;", []string{"select $tag$\n\\v(\n$tag$;"}, []string{"|", "|", "|"}, "=", nil},
		{"select $tag$\n\\v(\n$tag$\\g", []string{"select $tag$\n\\v(\n$tag$"}, []string{"|", "|", `\g|`}, "=", nil},
		{"select $$\n\\v(\n$tag$$zz$$\\g$$\\g", []string{"select $$\n\\v(\n$tag$$zz$$\\g$$"}, []string{"|", "|", `\g|`}, "=", nil},
		{"select * --\n\\v", nil, []string{"|", `\v|`}, "-", nil}, // 34
		{"select--", nil, []string{"|"}, "-", nil},
		{"select --", nil, []string{"|"}, "-", nil},
		{"select /**/", nil, []string{"|"}, "-", nil},
		{"select/* */", nil, []string{"|"}, "-", nil},
		{"select/*", nil, []string{"|"}, "*", nil},
		{"select /*", nil, []string{"|"}, "*", nil},
		{"select * /**/", nil, []string{"|"}, "-", nil},
		{"select * /* \n\n\n--*/\n;", []string{"select * /* \n\n\n--*/\n;"}, []string{"|", "|", "|", "|", "|"}, "=", nil},
		{"select * /* \n\n\n--*/\n", nil, []string{"|", "|", "|", "|", "|"}, "-", nil}, // 43
		{"select * /* \n\n\n--\n", nil, []string{"|", "|", "|", "|", "|"}, "*", nil},
		{"\\p \\p\nselect (", nil, []string{`\p| `, `\p|`, "|"}, "(", nil}, // 45
		{"\\p \\p\nselect ()", nil, []string{`\p| `, `\p|`, "|"}, "-", nil},
		{"\n             \t\t               \n", nil, []string{"|", "|", "|"}, "=", nil},
		{"\n   foob      \t\t               \n", nil, []string{"|", "|", "|"}, "-", nil},
		{"$$", nil, []string{"|"}, "$", nil},
		{"$$foo", nil, []string{"|"}, "$", nil}, // 50
		{"'", nil, []string{"|"}, "'", nil},
		{"(((()()", nil, []string{"|"}, "(", nil},
		{"\"", nil, []string{"|"}, "\"", nil},
		{"\"foo", nil, []string{"|"}, "\"", nil},
		{":a :b", nil, []string{"|"}, "-", []string{"a", "b"}}, // 55
		{`select :'a b' :"foo bar"`, nil, []string{"|"}, "-", []string{"a b", "foo bar"}},
		{`select :a:b;`, []string{"select :a:b;"}, []string{"|"}, "=", []string{"a", "b"}},
		{"select :'a\n:foo:bar", nil, []string{"|", "|"}, "'", nil}, // 58
		{"select :''\n:foo:bar\\g", []string{"select :''\n:foo:bar"}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :''\n:foo :bar\\g", []string{"select :''\n:foo :bar"}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}}, // 60
		{"select :''\n :foo :bar \\g", []string{"select :''\n :foo :bar "}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :'a\n:'foo':\"bar\"", nil, []string{"|", "|"}, "'", nil}, // 62
		{"select :''\n:'foo':\"bar\"\\g", []string{"select :''\n:'foo':\"bar\""}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :''\n:'foo' :\"bar\"\\g", []string{"select :''\n:'foo' :\"bar\""}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{"select :''\n :'foo' :\"bar\" \\g", []string{"select :''\n :'foo' :\"bar\" "}, []string{"|", `\g|`}, "=", []string{"foo", "bar"}},
		{`select 1\echo 'pg://':foo'/':bar`, nil, []string{`\echo| 'pg://':foo'/':bar`}, "-", nil}, // 66
		{`select :'foo'\echo 'pg://':bar'/' `, nil, []string{`\echo| 'pg://':bar'/' `}, "-", []string{"foo"}},
		{`select 1\g '\g`, []string{`select 1`}, []string{`\g| '\g`}, "=", nil},
		{`select 1\g "\g`, []string{`select 1`}, []string{`\g| "\g`}, "=", nil},
		{"select 1\\g `\\g", []string{`select 1`}, []string{"\\g| `\\g"}, "=", nil}, // 70
		{`select 1\g '\g `, []string{`select 1`}, []string{`\g| '\g `}, "=", nil},
		{`select 1\g "\g `, []string{`select 1`}, []string{`\g| "\g `}, "=", nil},
		{"select 1\\g `\\g ", []string{`select 1`}, []string{"\\g| `\\g "}, "=", nil},
		{"select $$\\g$$\\g", []string{`select $$\g$$`}, []string{`\g|`}, "=", nil},
		{"select $1\\bind a b c\\g", []string{`select $1`}, []string{`\bind| a b c`, `\g|`}, "=", nil},
		{"select $1 \\bind a b c \\g", []string{`select $1 `}, []string{`\bind| a b c `, `\g|`}, "=", nil},
		{"select $2, $a$ foo $a$, $1 \\bind a b \\g", []string{`select $2, $a$ foo $a$, $1 `}, []string{`\bind| a b `, `\g|`}, "=", nil},
	}
	for i, test := range tests {
		b := NewStmt(sp(test.s, "\n"))
		var stmts, cmds, aparams []string
		var vars []*Var
	loop:
		for {
			cmd, params, err := b.Next(Unquote)
			switch {
			case err == io.EOF:
				break loop
			case err != nil:
				t.Fatalf("test %d did not expect error, got: %v", i, err)
			}
			vars = append(vars, b.Vars...)
			if b.Ready() || cmd == `\g` {
				stmts = append(stmts, b.String())
				b.Reset(nil)
			}
			cmds = append(cmds, cmd)
			aparams = append(aparams, params)
		}
		if len(stmts) != len(test.stmts) {
			t.Logf(">> %#v // %#v", test.stmts, stmts)
			t.Fatalf("test %d expected %d statements, got: %d", i, len(test.stmts), len(stmts))
		}
		if !reflect.DeepEqual(stmts, test.stmts) {
			t.Logf(">> %#v // %#v", test.stmts, stmts)
			t.Fatalf("test %d expected statements %s, got: %s", i, jj(test.stmts), jj(stmts))
		}
		if cz := cc(cmds, aparams); !reflect.DeepEqual(cz, test.cmds) {
			t.Logf(">> cmds: %#v, aparams: %#v, cz: %#v, test.cmds: %#v", cmds, aparams, cz, test.cmds)
			t.Fatalf("test %d expected commands %v, got: %v", i, jj(test.cmds), jj(cz))
		}
		if st := b.State(); st != test.state {
			t.Fatalf("test %d expected end parse state `%s`, got: `%s`", i, test.state, st)
		}
		if len(vars) != len(test.vars) {
			t.Fatalf("test %d expected %d vars, got: %d", i, len(test.vars), len(vars))
		}
		for _, n := range test.vars {
			if !hasVar(vars, n) {
				t.Fatalf("test %d missing variable `%s`", i, n)
			}
		}
		b.Reset(nil)
		if len(b.Buf) != 0 {
			t.Fatalf("test %d after reset b.Buf should have len %d, got: %d", i, 0, len(b.Buf))
		}
		if b.Len != 0 {
			t.Fatalf("test %d after reset should have len %d, got: %d", i, 0, b.Len)
		}
		if len(b.Vars) != 0 {
			t.Fatalf("test %d after reset should have len(vars) == 0, got: %d", i, len(b.Vars))
		}
		if b.Prefix != "" {
			t.Fatalf("test %d after reset should have empty prefix, got: %s", i, b.Prefix)
		}
		if b.quote != 0 || b.quoteDollarTag != "" || b.multilineComment || b.balanceCount != 0 {
			t.Fatalf("test %d after reset should have a cleared parse state", i)
		}
		if st := b.State(); st != "=" {
			t.Fatalf("test %d after reset should have state `=`, got: `%s`", i, st)
		}
		if b.ready {
			t.Fatalf("test %d after reset should not be ready", i)
		}
	}
}

func TestEmptyVariablesRawString(t *testing.T) {
	stmt := new(Stmt)
	stmt.AppendString("select ", "\n")
	stmt.Prefix = "SELECT"
	v := &Var{
		I:    7,
		End:  9,
		Name: "a",
		Len:  0,
	}
	stmt.Vars = append(stmt.Vars, v)

	if exp, got := "select ", stmt.RawString(); exp != got {
		t.Fatalf("Defined=false, expected: %s, got: %s", exp, got)
	}

	v.Defined = true
	if exp, got := "select :a", stmt.RawString(); exp != got {
		t.Fatalf("Defined=true, expected: %s, got: %s", exp, got)
	}
}

// cc combines commands with params.
func cc(cmds []string, params []string) []string {
	if len(cmds) == 0 {
		return []string{"|"}
	}
	z := make([]string, len(cmds))
	if len(cmds) != len(params) {
		panic("length of params should be same as cmds")
	}
	for i := 0; i < len(cmds); i++ {
		z[i] = cmds[i] + "|" + params[i]
	}
	return z
}

func jj(s []string) string {
	return "[`" + strings.Join(s, "`,`") + "`]"
}

func sp(a, sep string) func() ([]rune, error) {
	s := strings.Split(a, sep)
	return func() ([]rune, error) {
		if len(s) > 0 {
			z := s[0]
			s = s[1:]
			return []rune(z), nil
		}
		return nil, io.EOF
	}
}

func hasVar(vars []*Var, n string) bool {
	for _, v := range vars {
		if v.Name == n {
			return true
		}
	}
	return false
}

func TestGrab(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp rune
	}{
		{"", 0, 0},
		{"a", 0, 'a'},
		{" a", 0, ' '},
		{"a ", 1, ' '},
		{"a", 1, 0},
	}
	for i, test := range tests {
		z := []rune(test.s)
		r := grab(z, test.i, len(z))
		if r != test.exp {
			t.Errorf("test %d expected %c, got: %c", i, test.exp, r)
		}
	}
}

func TestFindSpace(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp int
		b   bool
	}{
		{"", 0, 0, false},
		{" ", 0, 0, true},
		{"a", 0, 1, false},
		{"a ", 0, 1, true},
		{" a ", 0, 0, true},
		{"aaa", 0, 3, false},
		{" a ", 1, 2, true},
		{"aaa", 1, 3, false},
		{" aaa", 1, 4, false},
	}
	for i, test := range tests {
		z := []rune(test.s)
		n, b := findSpace(z, test.i, len(z))
		if n != test.exp {
			t.Errorf("test %d expected %d, got: %d", i, test.exp, n)
		}
		if b != test.b {
			t.Errorf("test %d expected %t, got: %t", i, test.b, b)
		}
	}
}

func TestFindNonSpace(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp int
		b   bool
	}{
		{"", 0, 0, false},
		{" ", 0, 1, false},
		{"a", 0, 0, true},
		{"a ", 0, 0, true},
		{" a ", 0, 1, true},
		{"    ", 0, 4, false},
		{" a ", 1, 1, true},
		{"aaa", 1, 1, true},
		{" aaa", 1, 1, true},
		{"  aa", 1, 2, true},
		{"    ", 1, 4, false},
	}
	for i, test := range tests {
		z := []rune(test.s)
		n, b := findNonSpace(z, test.i, len(z))
		if n != test.exp {
			t.Errorf("test %d expected %d, got: %d", i, test.exp, n)
		}
		if b != test.b {
			t.Errorf("test %d expected %t, got: %t", i, test.b, b)
		}
	}
}

func TestIsEmptyLine(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp bool
	}{
		{"", 0, true},
		{"a", 0, false},
		{" a", 0, false},
		{" a ", 0, false},
		{" \na", 0, false},
		{" \n\ta", 0, false},
		{"a ", 1, true},
		{" a", 1, false},
		{" a ", 1, false},
		{" \na", 1, false},
		{" \n\t ", 1, true},
	}
	for i, test := range tests {
		z := []rune(test.s)
		b := isEmptyLine(z, test.i, len(z))
		if b != test.exp {
			t.Errorf("test %d expected %t, got: %t", i, test.exp, b)
		}
	}
}

func TestReadString(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp string
		ok  bool
	}{
		{`'`, 0, ``, false},
		{` '`, 1, ``, false},
		{`''`, 0, `''`, true},
		{`'foo' `, 0, `'foo'`, true},
		{` 'foo' `, 1, `'foo'`, true},
		{`"foo"`, 0, `"foo"`, true},
		{"`foo`", 0, "`foo`", true},
		{"`'foo'`", 0, "`'foo'`", true},
		{`'foo''foo'`, 0, `'foo''foo'`, true},
		{` 'foo''foo' `, 1, `'foo''foo'`, true},
		{` "foo''foo" `, 1, `"foo''foo"`, true},
		// escaped \" aren't allowed in strings, so the second " would be next
		// double quoted string
		{`"foo\""`, 0, `"foo\"`, true},
		{` "foo\"" `, 1, `"foo\"`, true},
		{`''''`, 0, `''''`, true},
		{` '''' `, 1, `''''`, true},
		{`''''''`, 0, `''''''`, true},
		{` '''''' `, 1, `''''''`, true},
		{`'''`, 0, ``, false},
		{` ''' `, 1, ``, false},
		{`'''''`, 0, ``, false},
		{` ''''' `, 1, ``, false},
		{`"fo'o"`, 0, `"fo'o"`, true},
		{` "fo'o" `, 1, `"fo'o"`, true},
		{`"fo''o"`, 0, `"fo''o"`, true},
		{` "fo''o" `, 1, `"fo''o"`, true},
	}
	for i, test := range tests {
		r := []rune(test.s)
		c, end := rune(strings.TrimSpace(test.s)[0]), len(r)
		if c != '\'' && c != '"' && c != '`' {
			t.Fatalf("test %d incorrect!", i)
		}
		pos, ok := readString(r, test.i+1, end, c, "")
		if ok != test.ok {
			t.Fatalf("test %d expected ok %t, got: %t", i, test.ok, ok)
		}
		if !test.ok {
			continue
		}
		if r[pos] != c {
			t.Fatalf("test %d expected last character to be %c, got: %c", i, c, r[pos])
		}
		v := string(r[test.i : pos+1])
		if n := len(v); n < 2 {
			t.Fatalf("test %d expected result of at least length 2, got: %d", i, n)
		}
		if v != test.exp {
			t.Errorf("test %d expected %q, got: %q", i, test.exp, v)
		}
	}
}

func TestReadCommand(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp string
	}{
		{`\c foo bar z`, 0, `\c| foo bar z|`}, // 0
		{`\c foo bar z `, 0, `\c| foo bar z |`},
		{`\c foo bar z  `, 0, `\c| foo bar z  |`},
		{`\c    foo    bar    z  `, 0, `\c|    foo    bar    z  |`},
		{`\c    pg://blah    bar    z  `, 0, `\c|    pg://blah    bar    z  |`},
		{`\foo    pg://blah    bar    z  `, 0, `\foo|    pg://blah    bar    z  |`}, // 5
		{`\a\b`, 0, `\a||\b`},
		{`\a \b`, 0, `\a| |\b`},
		{"\\a \n\\b", 0, "\\a| |\n\\b"},
		{` \ab \bc \cd `, 5, `\bc| |\cd `},
		{`\p foo \p`, 0, `\p| foo |\p`}, // 10
		{`\p foo   \p bar`, 0, `\p| foo   |\p bar`},
		{`\p\p`, 0, `\p||\p`},
		{`\p \r foo`, 0, `\p| |\r foo`},
		{`\print   \reset    foo`, 0, `\print|   |\reset    foo`},
		{`\print   \reset    foo`, 9, `\reset|    foo|`}, // 15
		{`\print   \reset    foo  `, 9, `\reset|    foo  |`},
		{`\print   \reset    foo  bar  `, 9, `\reset|    foo  bar  |`},
		{`\c 'foo bar' z`, 0, `\c| 'foo bar' z|`},
		{`\c foo "bar " z `, 0, `\c| foo "bar " z |`},
		{"\\c `foo bar z  `  ", 0, "\\c| `foo bar z  `  |"}, // 20
		{`\c 'foob':foo:bar'test'  `, 0, `\c| 'foob':foo:bar'test'  |`},
		{"\\a \n\\b\\c\n", 0, "\\a| |\n\\b\\c\n"},
		{`\a'foob' \b`, 0, `\a'foob'| |\b`},
		{`\foo 'test' "bar"\print`, 0, `\foo| 'test' "bar"|\print`}, // 25
		{`\foo 'test' "bar"  \print`, 0, `\foo| 'test' "bar"  |\print`},
		{`\afoob' \b`, 0, `\afoob'| |\b`},
		{`\afoob' '\b  `, 0, `\afoob'| '\b  |`},
		{`\afoob' '\b  '\print`, 0, `\afoob'| '\b  '|\print`},
		{`\afoob' '\b  ' \print`, 0, `\afoob'| '\b  ' |\print`}, // 30
		{`\afoob' '\b  ' \print `, 0, `\afoob'| '\b  ' |\print `},
		{"\\foo `foob'foob'\\print", 0, "\\foo| `foob'foob'\\print|"},
		{"\\foo `foob'foob'  \\print", 0, "\\foo| `foob'foob'  \\print|"},
		{`\foo "foob'foob'\\print`, 0, `\foo| "foob'foob'\\print|`},
		{`\foo "foob'foob'  \\print`, 0, `\foo| "foob'foob'  \\print|`}, // 35
		{`\foo "\""\print`, 0, `\foo| "\""|\print`},
		{`\foo "\"'"\print`, 0, `\foo| "\"'"|\print`},
		{`\foo "\"''"\print`, 0, `\foo| "\"''"|\print`},
	}
	for i, test := range tests {
		z := []rune(test.s)
		if !strings.Contains(test.exp, "|") {
			t.Fatalf("test %d expected value is invalid (missing |): %q", i, test.exp)
		}
		v := strings.Split(test.exp, "|")
		if len(v) != 3 {
			t.Fatalf("test %d should have 3 expected values, has: %d", i, len(v))
		}
		cmd, params := readCommand(z, test.i, len(z))
		if s := string(z[test.i:cmd]); s != v[0] {
			t.Errorf("test %d expected command to be `%s`, got: `%s` [%d, %d]", i, v[0], s, cmd, params)
		}
		if s := string(z[cmd:params]); s != v[1] {
			t.Errorf("test %d expected params to be `%s`, got: `%s` [%d, %d]", i, v[1], s, cmd, params)
		}
		if s := string(z[params:]); s != v[2] {
			t.Errorf("test %d expected remaining to be `%s`, got: `%s`", i, v[2], s)
		}
	}
}

func TestFindPrefix(t *testing.T) {
	tests := []struct {
		s   string
		w   int
		exp string
	}{
		{"", 4, ""},
		{"  ", 4, ""},
		{"  ", 4, ""},
		{" select ", 4, "SELECT"},
		{" select to ", 4, "SELECT TO"},
		{" select to ", 4, "SELECT TO"}, // 5
		{" select   to   ", 4, "SELECT TO"},
		{"select into from", 2, "SELECT INTO"},
		{"select into * from", 4, "SELECT INTO"},
		{" select into  *   from  ", 4, "SELECT INTO"},
		{" select   \t  into \n *  \t\t\n\n\n  from     ", 4, "SELECT INTO"}, // 10
		{"  select\n\n\tb\t\tzfrom j\n\n  ", 2, "SELECT B"},
		{"select/* foob  */into", 4, "SELECTINTO"}, // 12
		{"select/* foob  */\tinto", 4, "SELECT INTO"},
		{"select/* foob  */ into", 4, "SELECT INTO"},
		{"select/* foob  */ into ", 4, "SELECT INTO"},
		{"select /* foob  */ into ", 4, "SELECT INTO"},
		{"   select /* foob  */ into ", 4, "SELECT INTO"},
		{" select * --test\n from where \n\nfff", 4, "SELECT"},
		{"/*idreamedital*/foo//bar\n/*  nothing */test\n\n\nwe made /*\n\n\n\n*/   \t   it    ", 5, "FOO TEST WE MADE IT"},
		{" --yes\n//no\n\n\t/*whatever*/ ", 4, ""}, // 20
		{"/*/*test*/*/ select ", 4, ""},
		{"/*/*test*/*/ select ", 4, ""},
		{"//", 4, ""},
		{"-", 4, ""},
		{"* select", 4, ""},
		{"/**/", 4, ""},
		{"--\n\t\t\thello,\t--", 4, "HELLO"},
		{"/*   */\n\n\n\tselect/*--\n*/\t\b\bzzz", 4, "SELECT ZZZ"}, // 28
		{"n\nn\n\nn\tn", 7, "N N N N"},
		{"n\nn\n\nn\tn", 1, "N"},
		{"--\n/* */n/* */\nn\n--\nn\tn", 7, "N N N N"},
		{"--\n/* */n\n/* */\nn\n--\nn\tn", 7, "N N N N"},
		{"\n\n/* */\nn n", 7, "N N"},
		{"\n\n/* */\nn/* */n", 7, "NN"},
		{"\n\n/* */\nn /* */n", 7, "N N"},
		{"\n\n/* */\nn/* */\nn", 7, "N N"},
		{"\n\n/* */\nn/* */ n", 7, "N N"},
		{"*/foob", 7, ""},
		{"*/ \n --\nfoob", 7, ""},
		{"--\n\n--\ntest", 7, "TEST"}, // 40
		{"\b\btest", 7, "TEST"},
		{"select/*\r\n\r\n*/blah", 7, "SELECTBLAH"},
		{"\r\n\r\nselect from where", 8, "SELECT FROM WHERE"},
		{"\r\n\b\bselect 1;create 2;", 8, "SELECT"},
		{"\r\n\bbegin transaction;\ncreate x where;", 8, "BEGIN TRANSACTION"}, // 45
		{"begin;test;create;awesome", 3, "BEGIN"},
		{" /* */ ; begin; ", 5, ""},
		{" /* foo */ test; test", 5, "TEST"},
		{";test", 5, ""},
		{"\b\b\t;test", 5, ""},
		{"\b\t; test", 5, ""},
		{"\b\tfoob; test", 5, "FOOB"},
		{"  TEST /*\n\t\b*/\b\t;foob", 10, "TEST"},
		{"begin transaction\n\tinsert into x;\ncommit;", 6, "BEGIN TRANSACTION INSERT INTO X"},
		{"--\nbegin /* */transaction/* */\n/* */\tinsert into x;--/* */\ncommit;", 6, "BEGIN TRANSACTION INSERT INTO X"},
		{"#\nbegin /* */transaction/* */\n/* */\t#\ninsert into x;#\n--/* */\ncommit;", 6, "BEGIN TRANSACTION INSERT INTO X"},
	}
	for i, test := range tests {
		if p := findPrefix([]rune(test.s), test.w); p != test.exp {
			t.Errorf("test %d %q expected %q, got: %q", i, test.s, test.exp, p)
		}
	}
}

func TestReadVar(t *testing.T) {
	tests := []struct {
		s   string
		i   int
		exp *Var
	}{
		{``, 0, nil},
		{`:`, 0, nil},
		{` :`, 0, nil},
		{`a:`, 0, nil},
		{`a:a`, 0, nil},
		{`: `, 0, nil},
		{`: a `, 0, nil},
		{`:a`, 0, v(0, 2, `a`)}, // 7
		{`:ab`, 0, v(0, 3, `ab`)},
		{`:a `, 0, v(0, 2, `a`)},
		{`:a_ `, 0, v(0, 3, `a_`)},
		{":a_\t ", 0, v(0, 3, `a_`)},
		{":a_\n ", 0, v(0, 3, `a_`)},
		{`:a9`, 0, v(0, 3, `a9`)}, // 13
		{`:ab9`, 0, v(0, 4, `ab9`)},
		{`:a 9`, 0, v(0, 2, `a`)},
		{`:a_9 `, 0, v(0, 4, `a_9`)},
		{":a_9\t ", 0, v(0, 4, `a_9`)},
		{":a_9\n ", 0, v(0, 4, `a_9`)},
		{`:a_;`, 0, v(0, 3, `a_`)}, // 19
		{`:a_\`, 0, v(0, 3, `a_`)},
		{`:a_$`, 0, v(0, 3, `a_`)},
		{`:a_'`, 0, v(0, 3, `a_`)},
		{`:a_"`, 0, v(0, 3, `a_`)},
		{`:ab `, 0, v(0, 3, `ab`)}, // 24
		{`:ab123 `, 0, v(0, 6, `ab123`)},
		{`:ab123`, 0, v(0, 6, `ab123`)},
		{`:'`, 0, nil}, // 27
		{`:' `, 0, nil},
		{`:' a`, 0, nil},
		{`:' a `, 0, nil},
		{`:"`, 0, nil},
		{`:" `, 0, nil},
		{`:" a`, 0, nil},
		{`:" a `, 0, nil},
		{`:''`, 0, nil}, // 35
		{`:'' `, 0, nil},
		{`:'' a`, 0, nil},
		{`:""`, 0, nil},
		{`:"" `, 0, nil},
		{`:"" a`, 0, nil},
		{`:'     `, 0, nil}, // 41
		{`:'       `, 0, nil},
		{`:"     `, 0, nil},
		{`:"       `, 0, nil},
		{`:'a'`, 0, v(0, 4, `a`, `'`)}, // 45
		{`:'a' `, 0, v(0, 4, `a`, `'`)},
		{`:'ab'`, 0, v(0, 5, `ab`, `'`)},
		{`:'ab' `, 0, v(0, 5, `ab`, `'`)},
		{`:'ab  ' `, 0, v(0, 7, `ab  `, `'`)},
		{`:"a"`, 0, v(0, 4, `a`, `"`)}, // 50
		{`:"a" `, 0, v(0, 4, `a`, `"`)},
		{`:"ab"`, 0, v(0, 5, `ab`, `"`)},
		{`:"ab" `, 0, v(0, 5, `ab`, `"`)},
		{`:"ab  " `, 0, v(0, 7, `ab  `, `"`)},
		{`:型`, 0, v(0, 2, "型")}, // 55
		{`:'型'`, 0, v(0, 4, "型", `'`)},
		{`:"型"`, 0, v(0, 4, "型", `"`)},
		{` :型 `, 1, v(1, 3, "型")},
		{` :'型' `, 1, v(1, 5, "型", `'`)},
		{` :"型" `, 1, v(1, 5, "型", `"`)},
		{`:型示師`, 0, v(0, 4, "型示師")}, // 61
		{`:'型示師'`, 0, v(0, 6, "型示師", `'`)},
		{`:"型示師"`, 0, v(0, 6, "型示師", `"`)},
		{` :型示師 `, 1, v(1, 5, "型示師")},
		{` :'型示師' `, 1, v(1, 7, "型示師", `'`)},
		{` :"型示師" `, 1, v(1, 7, "型示師", `"`)},
	}
	for i, test := range tests {
		z := []rune(test.s)
		v := readVar(z, test.i, len(z))
		if !reflect.DeepEqual(v, test.exp) {
			t.Errorf("test %d expected %#v, got: %#v", i, test.exp, v)
		}
		if test.exp != nil && v != nil {
			n := string(z[v.I+1 : v.End])
			if v.Quote != 0 {
				if c := rune(n[0]); c != v.Quote {
					t.Errorf("test %d expected var to start with quote %c, got: %c", i, c, v.Quote)
				}
				if c := rune(n[len(n)-1]); c != v.Quote {
					t.Errorf("test %d expected var to end with quote %c, got: %c", i, c, v.Quote)
				}
				n = n[1 : len(n)-1]
			}
			if n != test.exp.Name {
				t.Errorf("test %d expected var name of `%s`, got: `%s`", i, test.exp.Name, n)
			}
		}
	}
}

func TestSubstitute(t *testing.T) {
	a512 := sl(512, 'a')
	b512 := sl(512, 'a')
	b512 = b512[:1] + "b" + b512[2:]
	if len(b512) != 512 {
		t.Fatalf("b512 should be length 512, is: %d", len(b512))
	}
	tests := []struct {
		s   string
		i   int
		n   int
		t   string
		exp string
	}{
		{"", 0, 0, "", ""},
		{"a", 0, 1, "b", "b"},
		{"ab", 1, 1, "cd", "acd"},
		{"", 0, 0, "ab", "ab"},
		{"abc", 1, 2, "d", "ad"},
		{a512, 1, 1, "b", b512},
		{"foo", 0, 1, "bar", "baroo"},
	}
	for i, test := range tests {
		r := []rune(test.s)
		r, rlen := substitute(r, test.i, len(r), test.n, test.t)
		if rlen != len(test.exp) {
			t.Errorf("test %d expected length %d, got: %d", i, len(test.exp), rlen)
		}
		if s := string(r); s != test.exp {
			t.Errorf("test %d expected %q, got %q", i, test.exp, s)
		}
	}
}

func TestSubstituteVar(t *testing.T) {
	a512 := sl(512, 'a')
	tests := []struct {
		s   string
		v   *Var
		sub string
		exp string
	}{
		{`:a`, v(0, 2, `a`), `x`, `x`},
		{` :a`, v(1, 3, `a`), `x`, ` x`},
		{`:a `, v(0, 2, `a`), `x`, `x `},
		{` :a `, v(1, 3, `a`), `x`, ` x `},
		{` :'a' `, v(1, 5, `a`, `'`), `'x'`, ` 'x' `},
		{` :"a" `, v(1, 5, "a", `"`), `"x"`, ` "x" `},
		{`:a`, v(0, 2, `a`), ``, ``}, // 6
		{` :a`, v(1, 3, `a`), ``, ` `},
		{`:a `, v(0, 2, `a`), ``, ` `},
		{` :a `, v(1, 3, `a`), ``, `  `},
		{` :'a' `, v(1, 5, `a`, `'`), ``, `  `},
		{` :"a" `, v(1, 5, "a", `"`), "", `  `},
		{` :aaa `, v(1, 5, "aaa"), "", "  "}, // 12
		{` :aaa `, v(1, 5, "aaa"), a512, " " + a512 + " "},
		{` :` + a512 + ` `, v(1, len(a512)+2, a512), "", "  "},
		{`:foo`, v(0, 4, "foo"), "这是一个", `这是一个`}, // 15
		{`:foo `, v(0, 4, "foo"), "这是一个", `这是一个 `},
		{` :foo`, v(1, 5, "foo"), "这是一个", ` 这是一个`},
		{` :foo `, v(1, 5, "foo"), "这是一个", ` 这是一个 `},
		{`:'foo'`, v(0, 6, `foo`, `'`), `'这是一个'`, `'这是一个'`}, // 19
		{`:'foo' `, v(0, 6, `foo`, `'`), `'这是一个'`, `'这是一个' `},
		{` :'foo'`, v(1, 7, `foo`, `'`), `'这是一个'`, ` '这是一个'`},
		{` :'foo' `, v(1, 7, `foo`, `'`), `'这是一个'`, ` '这是一个' `},
		{`:"foo"`, v(0, 6, `foo`, `"`), `"这是一个"`, `"这是一个"`}, // 23
		{`:"foo" `, v(0, 6, `foo`, `"`), `"这是一个"`, `"这是一个" `},
		{` :"foo"`, v(1, 7, `foo`, `"`), `"这是一个"`, ` "这是一个"`},
		{` :"foo" `, v(1, 7, `foo`, `"`), `"这是一个"`, ` "这是一个" `},
		{`:型`, v(0, 2, `型`), `x`, `x`}, // 27
		{` :型`, v(1, 3, `型`), `x`, ` x`},
		{`:型 `, v(0, 2, `型`), `x`, `x `},
		{` :型 `, v(1, 3, `型`), `x`, ` x `},
		{` :'型' `, v(1, 5, `型`, `'`), `'x'`, ` 'x' `},
		{` :"型" `, v(1, 5, "型", `"`), `"x"`, ` "x" `},
		{`:型`, v(0, 2, `型`), ``, ``}, // 33
		{` :型`, v(1, 3, `型`), ``, ` `},
		{`:型 `, v(0, 2, `型`), ``, ` `},
		{` :型 `, v(1, 3, `型`), ``, `  `},
		{` :'型' `, v(1, 5, `型`, `'`), ``, `  `},
		{` :"型" `, v(1, 5, "型", `"`), "", `  `},
		{`:型示師`, v(0, 4, `型示師`), `本門台初埼本門台初埼`, `本門台初埼本門台初埼`}, // 39
		{` :型示師`, v(1, 5, `型示師`), `本門台初埼本門台初埼`, ` 本門台初埼本門台初埼`},
		{`:型示師 `, v(0, 4, `型示師`), `本門台初埼本門台初埼`, `本門台初埼本門台初埼 `},
		{` :型示師 `, v(1, 5, `型示師`), `本門台初埼本門台初埼`, ` 本門台初埼本門台初埼 `},
		{` :型示師 `, v(1, 5, `型示師`), `本門台初埼本門台初埼`, ` 本門台初埼本門台初埼 `},
		{` :'型示師' `, v(1, 7, `型示師`), `'本門台初埼本門台初埼'`, ` '本門台初埼本門台初埼' `},
		{` :"型示師" `, v(1, 7, `型示師`), `"本門台初埼本門台初埼"`, ` "本門台初埼本門台初埼" `},
	}
	for i, test := range tests {
		z := []rune(test.s)
		y, l := substituteVar(z, test.v, test.sub)
		if sl := len([]rune(test.sub)); test.v.Len != sl {
			t.Errorf("test %d, expected v.Len to be %d, got: %d", i, sl, test.v.Len)
		}
		if el := len([]rune(test.exp)); l != el {
			t.Errorf("test %d expected l==%d, got: %d", i, el, l)
		}
		if s := string(y); s != test.exp {
			t.Errorf("test %d expected `%s`, got: `%s`", i, test.exp, s)
		}
	}
}

func v(i, end int, n string, x ...string) *Var {
	z := &Var{
		I:    i,
		End:  end,
		Name: n,
	}
	if len(x) != 0 {
		z.Quote = []rune(x[0])[0]
	}
	return z
}
