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
	"bytes"
	"regexp"
	"strconv"
	"unicode"
	"unicode/utf8"

	"gsmate/config"
	"gsmate/internal/errdef"
)

// MinCapIncrease is the minimum amount by which to grow a Stmt.Buf.
const MinCapIncrease = 512

// Var holds information about a variable.
type Var struct {
	// I is where the variable starts (ie, ':') in Stmt.Buf.
	I int
	// End is where the variable ends in Stmt.Buf.
	End int
	// Quote is the quote character used if the variable was quoted, 0
	// otherwise.
	Quote rune
	// Name is the actual variable name excluding ':' and any enclosing quote
	// characters.
	Name string
	// Len is the length of the replaced variable.
	Len int
	// Defined indicates whether the variable has been defined.
	Defined bool
}

// String satisfies the fmt.Stringer interface.
func (v *Var) String() string {
	var q string
	switch {
	case v.Quote == '\\':
		return "\\" + v.Name
	case v.Quote != 0:
		q = string(v.Quote)
	}
	return ":" + q + v.Name + q
}

// Stmt is a reusable statement buffer that handles reading and parsing
// SQL-like statements.
type Stmt struct {
	// f is the rune source.
	f func() ([]rune, error)
	// Buf is the statement buffer
	Buf []rune
	// Len is the current len of any statement in Buf.
	Len int
	// Prefix is the detected prefix of the statement.
	Prefix string
	// Vars is the list of encountered variables.
	Vars []*Var
	// r is the unprocessed runes.
	r []rune
	// rlen is the number of unprocessed runes.
	rlen int
	// quote indicates currently parsing a quoted string.
	quote rune
	// quoteDollarTag is the parsed tag of a dollar quoted string
	quoteDollarTag string
	// multilineComment is state of multiline comment processing
	multilineComment bool
	// balanceCount is the balanced paren count
	balanceCount int
	// ready indicates that a complete statement has been parsed
	ready bool
}

// New creates a new Stmt using the supplied rune source f.
func NewStmt(f func() ([]rune, error)) *Stmt {
	b := &Stmt{
		f: f,
	}
	return b
}

// String satisfies fmt.Stringer.
func (b *Stmt) String() string {
	return string(b.Buf)
}

// RawString returns the non-interpolated version of the statement buffer.
func (b *Stmt) RawString() string {
	if b.Len == 0 {
		return ""
	}
	s, z := string(b.Buf), new(bytes.Buffer)
	var i int
	// deinterpolate vars
	for _, v := range b.Vars {
		if !v.Defined {
			continue
		}
		if len(s) > i {
			z.WriteString(s[i:v.I])
		}
		if v.Quote != '\\' {
			z.WriteRune(':')
		}
		if v.Quote != 0 {
			z.WriteRune(v.Quote)
		}
		z.WriteString(v.Name)
		if v.Quote != 0 && v.Quote != '\\' {
			z.WriteRune(v.Quote)
		}
		i = v.I + v.Len
	}
	// add remaining
	if len(s) > i {
		z.WriteString(s[i:])
	}
	return z.String()
}

// Ready returns true when the statement buffer contains a non empty, balanced
// statement that has been properly terminated (ie, ended with a semicolon).
func (b *Stmt) Ready() bool {
	return b.ready
}

// Reset resets the statement buffer.
func (b *Stmt) Reset(r []rune) {
	// reset buf
	b.Buf, b.Len, b.Prefix, b.Vars = nil, 0, "", nil
	// quote state
	b.quote, b.quoteDollarTag = 0, ""
	// multicomment state
	b.multilineComment = false
	// balance state
	b.balanceCount = 0
	// ready state
	b.ready = false
	if r != nil {
		b.r, b.rlen = r, len(r)
	}
}

// lineend is the slice to use when appending a line.
var lineend = []rune{'\n'}

// Next reads the next statement from the rune source, returning when either
// the statement has been terminated, or a meta command has been read from the
// rune source. After a call to Next, the collected statement is available in
// Stmt.Buf, or call Stmt.String() to convert it to a string.
//
// After a call to Next, Reset should be called if the extracted statement was
// executed (ie, processed). Note that the rune source supplied to New will be
// called again only after any remaining collected runes have been processed.
//
// Example:
//
//	buf := stmt.New(runeSrc)
//	for {
//	    cmd, params, err := buf.Next(unquoteFunc)
//	    if err { /* ... */ }
//
//	    execute, quit := buf.Ready() || cmd == "g", cmd == "q"
//
//	    // process command ...
//	    switch cmd {
//	        /* ... */
//	    }
//
//	    if quit {
//	        break
//	    }
//
//	    if execute {
//	       s := buf.String()
//	       res, err := db.Query(s)
//	       /* handle database ... */
//	       buf.Reset(nil)
//	    }
//	}
func (b *Stmt) Next(unquote func(string, bool) (bool, string, error)) (string, string, error) {
	var i int
	var err error
	// no runes to process, grab more
	if b.rlen == 0 {
		b.r, err = b.f()
		if err != nil {
			return "", "", err
		}
		b.rlen = len(b.r)
	}
	var cmd, params string
	var ok bool
parse:
	for ; i < b.rlen; i++ {
		// log.Printf(">> (%c) %d", b.r[i], i)
		// grab c, next
		c, next := b.r[i], grab(b.r, i+1, b.rlen)
		switch {
		// find end of string
		case b.quote != 0:
			i, ok = readString(b.r, i, b.rlen, b.quote, b.quoteDollarTag)
			if ok {
				b.quote, b.quoteDollarTag = 0, ""
			}
		// find end of multiline comment
		case b.multilineComment:
			i, ok = readMultilineComment(b.r, i, b.rlen)
			b.multilineComment = !ok
		// start of single or double quoted string
		case c == '\'' || c == '"':
			b.quote = c
		// start of dollar quoted string literal (postgres)
		case c == '$' && (next == '$' || next == '_' || unicode.IsLetter(next)):
			var id string
			id, i, ok = readDollarAndTag(b.r, i, b.rlen)
			if ok {
				b.quote, b.quoteDollarTag = '$', id
			}
		// start of sql comment, skip to end of line
		case c == '-' && next == '-':
			i = b.rlen
		// start of c-style comment, skip to end of line
		case c == '/' && next == '/':
			i = b.rlen
		// start of hash comment, skip to end of line
		case c == '#':
			i = b.rlen
		// start of multiline comment
		case c == '/' && next == '*':
			b.multilineComment = true
			i++
		// variable declaration
		case c == ':' && next != ':':
			if v := readVar(b.r, i, b.rlen); v != nil {
				var q string
				if v.Quote != 0 {
					q = string(v.Quote)
				}
				b.Vars = append(b.Vars, v)
				if ok, z, _ := unquote(q+v.Name+q, true); ok {
					v.Defined = true
					b.r, b.rlen = substituteVar(b.r, v, z)
					i--
				}
				if b.Len != 0 {
					v.I += b.Len + 1
				}
			}
		// unbalance
		case c == '(':
			b.balanceCount++
		// balance
		case c == ')':
			b.balanceCount = max(0, b.balanceCount-1)
		// continue processing quoted string, multiline comment, or unbalanced statements
		case b.quote != 0 || b.multilineComment || b.balanceCount != 0:
		// skip escaped backslash, semicolon, colon
		case c == '\\' && (next == '\\' || next == ';' || next == ':'):
			// FIXME: the below works, but it may not make sense to keep this enabled.
			// FIXME: also, the behavior is slightly different than psql
			v := &Var{
				I:     i,
				End:   i + 2,
				Quote: '\\',
				Name:  string(next),
			}
			b.Vars = append(b.Vars, v)
			b.r, b.rlen = substituteVar(b.r, v, string(next))
			if b.Len != 0 {
				v.I += b.Len + 1
			}
		// start of command
		case c == '\\':
			// parse command and params end positions
			cend, pend := readCommand(b.r, i, b.rlen)
			cmd, params = string(b.r[i:cend]), string(b.r[cend:pend])
			// remove command and params from r
			b.r = append(b.r[:i], b.r[pend:]...)
			b.rlen = len(b.r)
			break parse
		// terminated
		case c == ';':
			b.ready = true
			i++
			break parse
		}
	}
	// fix i -- i will be +1 when passing the length, which is a problem as the
	// '\n' will get copied from the source.
	i = min(i, b.rlen)
	// append line to buf when:
	// 1. in a quoted string (ie, ', ", or $)
	// 2. in a multiline comment
	// 3. line is not empty
	//
	// DO NOT append to buf when:
	// 1. line is empty/whitespace and not in a string/multiline comment
	empty := isEmptyLine(b.r, 0, i)
	appendLine := b.quote != 0 || b.multilineComment || !empty
	if !b.multilineComment && cmd != "" && empty {
		appendLine = false
	}
	if appendLine {
		// skip leading space when empty
		st := 0
		if b.Len == 0 {
			st, _ = findNonSpace(b.r, 0, i)
		}
		b.Append(b.r[st:i], lineend)
	}
	// set prefix
	b.Prefix = findPrefix(b.Buf, prefixCount)
	// reset r
	b.r = b.r[i:]
	b.rlen = len(b.r)
	return cmd, params, nil
}

// Append appends r to b.Buf separated by sep when b.Buf is not already empty.
//
// Dynamically grows b.Buf as necessary to accommodate r and the separator.
// Specifically, when b.Buf is not empty, b.Buf will grow by increments of
// MinCapIncrease.
//
// After a call to Append, b.Len will be len(b.Buf)+len(sep)+len(r). Call Reset
// to reset the Buf.
func (b *Stmt) Append(r, sep []rune) {
	rlen := len(r)
	// initial
	if b.Buf == nil {
		b.Buf, b.Len = r, rlen
		return
	}
	blen, seplen := b.Len, len(sep)
	tlen := blen + rlen + seplen
	// grow
	if bcap := cap(b.Buf); tlen > bcap {
		n := tlen + 2*rlen
		n += MinCapIncrease - (n % MinCapIncrease)
		z := make([]rune, blen, n)
		copy(z, b.Buf)
		b.Buf = z
	}
	b.Buf = b.Buf[:tlen]
	copy(b.Buf[blen:], sep)
	copy(b.Buf[blen+seplen:], r)
	b.Len = tlen
}

// AppendString is a util func wrapping Append.
func (b *Stmt) AppendString(s, sep string) {
	b.Append([]rune(s), []rune(sep))
}

// State returns a string representing the state of statement parsing.
func (b *Stmt) State() string {
	switch {
	case b.quote != 0:
		return string(b.quote)
	case b.multilineComment:
		return "*"
	case b.balanceCount != 0:
		return "("
	case b.Len != 0:
		return "-"
	}
	return "="
}

// IsSpaceOrControl is a special test for either a space or a control (ie, \b)
// characters.
func IsSpaceOrControl(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r)
}

// RunesLastIndex returns the last index in r of needle, or -1 if not found.
func RunesLastIndex(r []rune, needle rune) int {
	i := len(r) - 1
	for ; i >= 0; i-- {
		if r[i] == needle {
			return i
		}
	}
	return i
}

// prefixCount is the number of words to extract from a prefix.
const prefixCount = 6

// grab grabs i from r, or returns 0 if i >= end.
func grab(r []rune, i, end int) rune {
	if i < end {
		return r[i]
	}
	return 0
}

// findSpace finds first space rune in r, returning end if not found.
func findSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if IsSpaceOrControl(r[i]) {
			return i, true
		}
	}
	return i, false
}

// findNonSpace finds first non space rune in r, returning end if not found.
func findNonSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if !IsSpaceOrControl(r[i]) {
			return i, true
		}
	}
	return i, false
}

// findRune finds the next rune c in r, returning end if not found.
func findRune(r []rune, i, end int, c rune) (int, bool) {
	for ; i < end; i++ {
		if r[i] == c {
			return i, true
		}
	}
	return i, false
}

// isEmptyLine returns true when r is empty or composed of only whitespace.
func isEmptyLine(r []rune, i, end int) bool {
	_, ok := findNonSpace(r, i, end)
	return !ok
}

// identifierRE is a regexp that matches dollar tag identifiers ($tag$).
var identifierRE = regexp.MustCompile(`(?i)^[a-z_][a-z0-9_]{0,127}$`)

// readDollarAndTag reads a dollar style $tag$ in r, starting at i, returning
// the enclosed "tag" and position, or -1 if the dollar and tag was invalid.
func readDollarAndTag(r []rune, i, end int) (string, int, bool) {
	start, found := i, false
	i++
	for ; i < end; i++ {
		if r[i] == '$' {
			found = true
			break
		}
		if i-start > 128 {
			break
		}
	}
	if !found {
		return "", i, false
	}
	// check valid identifier
	id := string(r[start+1 : i])
	if id != "" && !identifierRE.MatchString(id) {
		return "", i, false
	}
	return id, i, true
}

// readString seeks to the end of a string returning the position and whether
// or not the string's end was found.
//
// If the string's terminator was not found, then the result will be the passed
// end.
func readString(r []rune, i, end int, quote rune, tag string) (int, bool) {
	var prev, c, next rune
	for ; i < end; i++ {
		c, next = r[i], grab(r, i+1, end)
		switch {
		case quote == '\'' && c == '\\':
			i++
			prev = 0
			continue
		case quote == '\'' && c == '\'' && next == '\'':
			i++
			continue
		case quote == '\'' && c == '\'' && prev != '\'',
			quote == '"' && c == '"',
			quote == '`' && c == '`':
			return i, true
		case quote == '$' && c == '$':
			if id, pos, ok := readDollarAndTag(r, i, end); ok && tag == id {
				return pos, true
			}
		}
		prev = c
	}
	return end, false
}

// readMultilineComment finds the end of a multiline comment (ie, '*/').
func readMultilineComment(r []rune, i, end int) (int, bool) {
	i++
	for ; i < end; i++ {
		if r[i-1] == '*' && r[i] == '/' {
			return i, true
		}
	}
	return end, false
}

// readStringVar reads a string quoted variable.
func readStringVar(r []rune, i, end int) *Var {
	start, q := i, grab(r, i+1, end)
	for i += 2; i < end; i++ {
		c := grab(r, i, end)
		if c == q {
			if i-start < 3 {
				return nil
			}
			return &Var{
				I:     start,
				End:   i + 1,
				Quote: q,
				Name:  string(r[start+2 : i]),
			}
		}
		/*
			// this is commented out, because need to determine what should be
			// the "right" behavior ... should we only allow "identifiers"?
			else if c != '_' && !unicode.IsLetter(c) && !unicode.IsNumber(c) {
				return nil
			}
		*/
	}
	return nil
}

// readVar reads variable from r.
func readVar(r []rune, i, end int) *Var {
	if grab(r, i, end) != ':' || grab(r, i+1, end) == ':' {
		return nil
	}
	if end-i < 2 {
		return nil
	}
	if c := grab(r, i+1, end); c == '"' || c == '\'' {
		return readStringVar(r, i, end)
	}
	start := i
	i++
	for ; i < end; i++ {
		if c := grab(r, i, end); c != '_' && !unicode.IsLetter(c) && !unicode.IsNumber(c) {
			break
		}
	}
	if i-start < 2 {
		return nil
	}
	return &Var{
		I:    start,
		End:  i,
		Name: string(r[start+1 : i]),
	}
}

// readCommand reads the command and any parameters from r, returning the
// offset from i for the end of command, and the end of the command parameters.
//
// A command is defined as the first non-blank text after \, followed by
// parameters up to either the next \ or a control character (for example, \n):
func readCommand(r []rune, i, end int) (int, int) {
command:
	// find end of command
	for ; i < end; i++ {
		next := grab(r, i+1, end)
		switch {
		case next == 0:
			return end, end
		case next == '\\' || unicode.IsControl(next):
			i++
			return i, i
		case unicode.IsSpace(next):
			i++
			break command
		}
	}
	cmd, quote := i, rune(0)
params:
	// find end of params
	for ; i < end; i++ {
		c, next := r[i], grab(r, i+1, end)
		switch {
		case next == 0:
			return cmd, end
		case quote == 0 && (c == '\'' || c == '"' || c == '`'):
			quote = c
		case quote != 0 && c == quote:
			quote = 0
		// skip escaped
		case quote != 0 && c == '\\' && (next == quote || next == '\\'):
			i++
		case quote == 0 && (c == '\\' || unicode.IsControl(c)):
			break params
		}
	}
	// log.Printf(">>> params: %q remaining: %q", string(r[cmd:i]), string(r[i:end]))
	return cmd, i
}

// findPrefix finds the prefix in r up to n words.
func findPrefix(r []rune, n int) string {
	var s []rune
	var words int
loop:
	for i, end := 0, len(r); i < end; i++ {
		// skip space + control characters
		if j, _ := findNonSpace(r, i, end); i != j {
			r, end, i = r[j:], end-j, 0
		}
		// grab current and next character
		c, next := grab(r, i, end), grab(r, i+1, end)
		switch {
		// do nothing
		case c == 0:
		// statement terminator
		case c == ';':
			break loop
		// single line comments '--' and '//'
		case c == '-' && next == '-', c == '/' && next == '/', c == '#':
			if i != 0 {
				s, words = appendUpperRunes(s, r[:i], ' '), words+1
			}
			// find line end
			if i, _ = findRune(r, i, end, '\n'); i >= end {
				break
			}
			r, end, i = r[i+1:], end-i-1, -1
		// multiline comments '/*' '*/'
		case c == '/' && next == '*':
			if i != 0 {
				s, words = appendUpperRunes(s, r[:i]), words+1
			}
			// find comment end '*/'
			for i += 2; i < end; i++ {
				if grab(r, i, end) == '*' && grab(r, i+1, end) == '/' {
					r, end, i = r[i+2:], end-i-2, -1
					break
				}
			}
			// add space when remaining runes begin with space, and previous
			// captured word did not
			if sl := len(s); end > 0 && sl != 0 && IsSpaceOrControl(r[0]) && !IsSpaceOrControl(s[sl-1]) {
				s = append(s, ' ')
			}
		// end of statement, max words, or punctuation that can be ignored
		case words == n || !unicode.IsLetter(c):
			break loop
		// ignore remaining, as no prefix can come after
		case next != '/' && !unicode.IsLetter(next):
			s, words = appendUpperRunes(s, r[:i+1], ' '), words+1
			if next == 0 {
				break
			}
			if next == ';' {
				break loop
			}
			r, end, i = r[i+2:], end-i-2, -1
		}
	}
	// trim right ' ', if any
	if sl := len(s); sl != 0 && s[sl-1] == ' ' {
		return string(s[:sl-1])
	}
	return string(s)
}

// FindPrefix finds the first 6 prefix words in s.
func FindPrefix(s string) string {
	return findPrefix([]rune(s), prefixCount)
}

// substitute substitutes n runes in r starting at i with the runes in s.
// Dynamically grows r if necessary.
func substitute(r []rune, i, end, n int, s string) ([]rune, int) {
	sr, rcap := []rune(s), cap(r)
	sn := len(sr)
	// grow ...
	tlen := len(r) + sn - n
	if tlen > rcap {
		z := make([]rune, tlen)
		copy(z, r)
		r = z
	} else {
		r = r[:rcap]
	}
	// substitute
	copy(r[i+sn:], r[i+n:])
	copy(r[i:], sr)
	return r[:tlen], tlen
}

// substituteVar substitutes part of r, based on v, with s.
func substituteVar(r []rune, v *Var, s string) ([]rune, int) {
	sr, rcap := []rune(s), cap(r)
	v.Len = len(sr)
	// grow ...
	tlen := len(r) + v.Len - (v.End - v.I)
	if tlen > rcap {
		z := make([]rune, tlen)
		copy(z, r)
		r = z
	} else {
		r = r[:rcap]
	}
	// substitute
	copy(r[v.I+v.Len:], r[v.End:])
	copy(r[v.I:v.I+v.Len], sr)
	return r[:tlen], tlen
}

// appendUpperRunes creates a new []rune from s, with the runes in r on the end
// converted to upper case. extra runes will be appended to the final []rune.
func appendUpperRunes(s []rune, r []rune, extra ...rune) []rune {
	sl, rl, el := len(s), len(r), len(extra)
	sre := make([]rune, sl+rl+el)
	copy(sre[:sl], s)
	for i := 0; i < rl; i++ {
		sre[sl+i] = unicode.ToUpper(r[i])
	}
	copy(sre[sl+rl:], extra)
	return sre
}

func getConfig(s string) (bool, string, error) {
	q, n := "", s
	if c := s[0]; c == '\'' || c == '"' {
		var err error
		if n, err = Dequote(s, c); err != nil {
			return false, "", err
		}
		q = string(c)
	}
	if val, ok := config.GetConfigMap()[n]; ok {
		return true, q + val + q, nil
	}
	return false, s, nil
}

func Unquote(s string, isConfig bool) (bool, string, error) {
	if isConfig {
		return getConfig(s)
	}
	if len(s) < 2 {
		return false, "", errdef.ErrInvalidQuotedString
	}
	c := s[0]
	z, err := Dequote(s, c)
	if err != nil {
		return false, "", err
	}
	if c == '\'' || c == '"' {
		return true, z, nil
	}
	if c != '`' {
		return false, "", errdef.ErrInvalidQuotedString
	}
	return true, z, nil
}

var cleanDoubleRE = regexp.MustCompile(`(^|[^\\])''`)

// Dequote unquotes a string.
func Dequote(s string, quote byte) (string, error) {
	if len(s) < 2 || s[len(s)-1] != quote {
		return "", errdef.ErrUnterminatedQuotedString
	}
	s = s[1 : len(s)-1]
	if quote == '\'' {
		s = cleanDoubleRE.ReplaceAllString(s, "$1\\'")
	}

	// this is the last part of strconv.Unquote
	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	for len(s) > 0 {
		c, multibyte, ss, err := strconv.UnquoteChar(s, quote)
		switch {
		case err != nil && err == strconv.ErrSyntax:
			return "", errdef.ErrInvalidQuotedString
		case err != nil:
			return "", err
		}
		s = ss
		if c < utf8.RuneSelf || !multibyte {
			buf = append(buf, byte(c))
		} else {
			n := utf8.EncodeRune(runeTmp[:], c)
			buf = append(buf, runeTmp[:n]...)
		}
	}
	return string(buf), nil
}
