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
	"gsmate/internal/pgliterals"
	"regexp"
	"strings"

	"github.com/xwb1989/sqlparser"
)

var whiteSpaceRegex = regexp.MustCompile(`\s+`)

func _compileRegex(keyword string) *regexp.Regexp {
	pattern := `\b` + whiteSpaceRegex.ReplaceAllString(keyword, `\\s+`) + `\b`
	return regexp.MustCompile(`(?mi)` + pattern)
}

var keywords = pgliterals.GetKeywords()
var keywordRegexs = map[string]*regexp.Regexp{}

func init() {
	for key := range keywords {
		keywordRegexs[key] = _compileRegex(key)
	}
}

type PrevalenceCounter struct {
	KeywordCounts map[string]int
	NameCounts    map[string]int // TODO
}

func NewPrevalenceCounter() *PrevalenceCounter {
	return &PrevalenceCounter{
		KeywordCounts: map[string]int{},
		NameCounts:    map[string]int{},
	}
}

func (p *PrevalenceCounter) Update(text string) {
	p.UpdateKeywords(text)
	p.UpdateNames(text)
}

func (p *PrevalenceCounter) UpdateNames(text string) {
	stmt, err := sqlparser.Parse(text)
	if err != nil {
		return
	}
	sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node.(type) {
		case sqlparser.ColIdent, sqlparser.TableIdent:
			s := strings.TrimSpace(sqlparser.String(node))
			if len(s) > 0 {
				p.NameCounts[s]++
			}
		default:
		}
		return true, nil
	}, stmt)
}

func (p *PrevalenceCounter) UpdateKeywords(text string) {
	for key, reg := range keywordRegexs {
		matches := reg.FindAllStringSubmatchIndex(text, -1)
		for range matches {
			p.KeywordCounts[key]++
		}
	}
}

func (p *PrevalenceCounter) ClearNames() {
	p.NameCounts = map[string]int{}
}

func (p *PrevalenceCounter) KeywordCount(key string) int {
	return p.KeywordCounts[key]
}

func (p *PrevalenceCounter) NameCount(name string) int {
	return p.NameCounts[name]
}
