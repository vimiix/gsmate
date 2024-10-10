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

package metadata

import (
	"database/sql"
	"strings"

	"gsmate/internal/errdef"
)

type Filter struct {
	// Schema name pattern that objects must belong to;
	// use Name to filter schemas by name
	Schema string
	// Parent name pattern that objects must belong to;
	// does not apply to schema and catalog containing matching objects
	Parent string
	// Reference name pattern of other objects referencing this one,
	Reference string
	// Name pattern that object name must match
	Name string
	// Types of the object
	Types []string
	// WithSystem objects
	WithSystem bool
	// OnlyVisible objects
	OnlyVisible bool
}

type Result interface {
	Values() []interface{}
}

type resultSet struct {
	results    []Result
	columns    []string
	current    int
	filter     func(Result) bool
	scanValues func(Result) []interface{}
}

func (r *resultSet) SetFilter(f func(Result) bool) {
	r.filter = f
}

func (r *resultSet) SetColumns(c []string) {
	r.columns = c
}

func (r *resultSet) SetScanValues(s func(Result) []interface{}) {
	r.scanValues = s
}

func (r *resultSet) Len() int {
	if r.filter == nil {
		return len(r.results)
	}
	len := 0
	for _, rec := range r.results {
		if r.filter(rec) {
			len++
		}
	}
	return len
}

func (r *resultSet) Reset() {
	r.current = 0
}

func (r *resultSet) Next() bool {
	r.current++
	if r.filter != nil {
		for r.current <= len(r.results) && !r.filter(r.results[r.current-1]) {
			r.current++
		}
	}
	return r.current <= len(r.results)
}

func (r resultSet) Columns() ([]string, error) {
	return r.columns, nil
}

func (r resultSet) Scan(dest ...interface{}) error {
	var v []interface{}
	if r.scanValues == nil {
		v = r.results[r.current-1].Values()
	} else {
		v = r.scanValues(r.results[r.current-1])
	}
	if len(v) != len(dest) {
		return errdef.ErrWrongNumberOfArguments
	}
	for i, d := range dest {
		p := d.(*interface{})
		*p = v[i]
	}
	return nil
}

func (r resultSet) Close() error {
	return nil
}

func (r resultSet) Err() error {
	return nil
}

func (r resultSet) NextResultSet() bool {
	return false
}

type CatalogSet struct {
	resultSet
}

func (s CatalogSet) Get() *Catalog {
	return s.results[s.current-1].(*Catalog)
}

func NewCatalogSet(v []Result) *CatalogSet {
	return &CatalogSet{
		resultSet: resultSet{
			results: v,
			columns: []string{"Catalog", "Owner", "Encoding", "Collate", "Ctype"},
		},
	}
}

type Catalog struct {
	Catalog          string
	Owner            string
	Encoding         string
	Collate          string
	Ctype            string
	AccessPrivileges sql.NullString
}

func (s Catalog) Values() []interface{} {
	return []interface{}{s.Catalog, s.Owner, s.Encoding, s.Collate, s.Ctype, s.AccessPrivileges}
}

type SchemaSet struct {
	resultSet
}

func NewSchemaSet(v []Schema) *SchemaSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &SchemaSet{
		resultSet: resultSet{
			results: r,
			columns: []string{"Schema", "Catalog"},
		},
	}
}

func (s SchemaSet) Get() *Schema {
	return s.results[s.current-1].(*Schema)
}

type Schema struct {
	Schema  string
	Catalog string
}

func (s Schema) Values() []interface{} {
	return []interface{}{s.Schema, s.Catalog}
}

type TableSet struct {
	resultSet
}

func NewTableSet(v []Table) *TableSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &TableSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Schema",

				"Name",
				"Type",

				"Rows",
				"Size",
				"Comment",
			},
		},
	}
}

func (t TableSet) Get() *Table {
	return t.results[t.current-1].(*Table)
}

type Table struct {
	Schema  string
	Name    string
	Type    string
	Rows    int64
	Size    string
	Comment sql.NullString
}

func (t Table) Values() []interface{} {
	return []interface{}{
		t.Schema,
		t.Name,
		t.Type,
		t.Rows,
		t.Size,
		t.Comment,
	}
}

type ColumnSet struct {
	resultSet
}

func NewColumnSet(v []Column) *ColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &ColumnSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Table",

				"Name",
				"Type",
				"Nullable",
				"Default",

				"Size",
				"Decimal Digits",
				"Precision Radix",
				"Octet Length",
			},
		},
	}
}

func (c ColumnSet) Get() *Column {
	return c.results[c.current-1].(*Column)
}

type Column struct {
	Catalog         string
	Schema          string
	Table           string
	Name            string
	OrdinalPosition int
	DataType        string
	// ScanType        reflect.Type
	Default         string
	ColumnSize      int
	DecimalDigits   int
	NumPrecRadix    int
	CharOctetLength int
	IsNullable      Bool
}

type Bool string

var (
	UNKNOWN Bool = ""
	YES     Bool = "YES"
	NO      Bool = "NO"
)

func (c Column) Values() []interface{} {
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.Table,
		c.Name,
		c.DataType,
		c.IsNullable,
		c.Default,
		c.ColumnSize,
		c.DecimalDigits,
		c.NumPrecRadix,
		c.CharOctetLength,
	}
}

type ColumnStatSet struct {
	resultSet
}

func NewColumnStatSet(v []ColumnStat) *ColumnStatSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &ColumnStatSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Table",
				"Name",

				"Average width",
				"Nulls fraction",
				"Distinct values",
				"Minimum value",
				"Maximum value",
				"Mean value",
				"Top N common values",
				"Top N values freqs",
			},
		},
	}
}

func (c ColumnStatSet) Get() *ColumnStat {
	return c.results[c.current-1].(*ColumnStat)
}

type ColumnStat struct {
	Catalog     string
	Schema      string
	Table       string
	Name        string
	AvgWidth    int
	NullFrac    float64
	NumDistinct int64
	Min         string
	Max         string
	Mean        string
	TopN        []string
	TopNFreqs   []float64
}

func (c ColumnStat) Values() []interface{} {
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.Table,
		c.Name,
		c.AvgWidth,
		c.NullFrac,
		c.NumDistinct,
		c.Min,
		c.Max,
		c.Mean,
		c.TopN,
		c.TopNFreqs,
	}
}

type IndexSet struct {
	resultSet
}

func NewIndexSet(v []Index) *IndexSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &IndexSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",

				"Name",
				"Table",

				"Is primary",
				"Is unique",
				"Type",
			},
		},
	}
}

func (i IndexSet) Get() *Index {
	return i.results[i.current-1].(*Index)
}

type Index struct {
	Catalog   string
	Schema    string
	Table     string
	Name      string
	IsPrimary Bool
	IsUnique  Bool
	Type      string
	Columns   string
}

func (i Index) Values() []interface{} {
	return []interface{}{
		i.Catalog,
		i.Schema,
		i.Name,
		i.Table,
		i.IsPrimary,
		i.IsUnique,
		i.Type,
	}
}

type IndexColumnSet struct {
	resultSet
}

func NewIndexColumnSet(v []IndexColumn) *IndexColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &IndexColumnSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Table",
				"Index name",

				"Name",
				"Data type",
			},
		},
	}
}

func (c IndexColumnSet) Get() *IndexColumn {
	return c.results[c.current-1].(*IndexColumn)
}

type IndexColumn struct {
	Catalog         string
	Schema          string
	Table           string
	IndexName       string
	Name            string
	DataType        string
	OrdinalPosition int
}

func (c IndexColumn) Values() []interface{} {
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.Table,
		c.IndexName,
		c.Name,
		c.DataType,
	}
}

type ConstraintSet struct {
	resultSet
}

func NewConstraintSet(v []Constraint) *ConstraintSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &ConstraintSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Table",
				"Name",

				"Type",
				"Is deferrable",
				"Initially deferred",

				"Foreign catalog",
				"Foreign schema",
				"Foreign table",
				"Foreign name",
				"Match type",
				"Update rule",
				"Delete rule",

				"Check Clause",
			},
		},
	}
}

func (i ConstraintSet) Get() *Constraint {
	return i.results[i.current-1].(*Constraint)
}

type Constraint struct {
	Catalog string
	Schema  string
	Table   string
	Name    string
	Type    string

	IsDeferrable        Bool
	IsInitiallyDeferred Bool

	ForeignCatalog string
	ForeignSchema  string
	ForeignTable   string
	ForeignName    string
	MatchType      string
	UpdateRule     string
	DeleteRule     string

	CheckClause string
}

func (i Constraint) Values() []interface{} {
	return []interface{}{
		i.Catalog,
		i.Schema,
		i.Table,
		i.Name,
		i.Type,
		i.IsDeferrable,
		i.IsInitiallyDeferred,
		i.ForeignCatalog,
		i.ForeignSchema,
		i.ForeignTable,
		i.ForeignName,
		i.MatchType,
		i.UpdateRule,
		i.DeleteRule,
	}
}

type ConstraintColumnSet struct {
	resultSet
}

func NewConstraintColumnSet(v []ConstraintColumn) *ConstraintColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &ConstraintColumnSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Table",
				"Constraint",
				"Name",
				"Foreign Catalog",
				"Foreign Schema",
				"Foreign Table",
				"Foreign Constraint",
				"Foreign Name",
			},
		},
	}
}

func (c ConstraintColumnSet) Get() *ConstraintColumn {
	return c.results[c.current-1].(*ConstraintColumn)
}

type ConstraintColumn struct {
	Catalog         string
	Schema          string
	Table           string
	Constraint      string
	Name            string
	OrdinalPosition int

	ForeignCatalog    string
	ForeignSchema     string
	ForeignTable      string
	ForeignConstraint string
	ForeignName       string
}

func (c ConstraintColumn) Values() []interface{} {
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.Table,
		c.Constraint,
		c.Name,
		c.ForeignCatalog,
		c.ForeignSchema,
		c.ForeignTable,
		c.ForeignConstraint,
		c.ForeignName,
	}
}

type FunctionSet struct {
	resultSet
}

func NewFunctionSet(v []Function) *FunctionSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &FunctionSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",

				"Name",
				"Result data type",
				"Argument data types",
				"Type",

				"Volatility",
				"Security",
				"Language",
				"Source code",
			},
		},
	}
}

func (f FunctionSet) Get() *Function {
	return f.results[f.current-1].(*Function)
}

type Function struct {
	Catalog    string
	Schema     string
	Name       string
	ResultType string
	ArgTypes   string
	Type       string
	Volatility string
	Security   string
	Language   string
	Source     string

	SpecificName string
}

func (f Function) Values() []interface{} {
	return []interface{}{
		f.Catalog,
		f.Schema,
		f.Name,
		f.ResultType,
		f.ArgTypes,
		f.Type,
		f.Volatility,
		f.Security,
		f.Language,
		f.Source,
	}
}

type FunctionColumnSet struct {
	resultSet
}

func NewFunctionColumnSet(v []FunctionColumn) *FunctionColumnSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &FunctionColumnSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Catalog",
				"Schema",
				"Function name",

				"Name",
				"Type",
				"Data type",

				"Size",
				"Decimal Digits",
				"Precision Radix",
				"Octet Length",
			},
		},
	}
}

func (c FunctionColumnSet) Get() *FunctionColumn {
	return c.results[c.current-1].(*FunctionColumn)
}

type FunctionColumn struct {
	Catalog         string
	Schema          string
	Table           string
	Name            string
	FunctionName    string
	OrdinalPosition int
	Type            string
	DataType        string
	// ScanType        reflect.Type
	ColumnSize      int
	DecimalDigits   int
	NumPrecRadix    int
	CharOctetLength int
}

func (c FunctionColumn) Values() []interface{} {
	return []interface{}{
		c.Catalog,
		c.Schema,
		c.FunctionName,
		c.Name,
		c.Type,
		c.DataType,
		c.ColumnSize,
		c.DecimalDigits,
		c.NumPrecRadix,
		c.CharOctetLength,
	}
}

type SequenceSet struct {
	resultSet
}

func NewSequenceSet(v []Sequence) *SequenceSet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &SequenceSet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Type",
				"Start",
				"Min",
				"Max",
				"Increment",
				"Cycles?",
			},
		},
	}
}

func (s SequenceSet) Get() *Sequence {
	return s.results[s.current-1].(*Sequence)
}

type Sequence struct {
	Catalog   string
	Schema    string
	Name      string
	DataType  string
	Start     string
	Min       string
	Max       string
	Increment string
	Cycles    Bool
}

func (s Sequence) Values() []interface{} {
	return []interface{}{
		s.DataType,
		s.Start,
		s.Min,
		s.Max,
		s.Increment,
		s.Cycles,
	}
}

type PrivilegeSummarySet struct {
	resultSet
}

func NewPrivilegeSummarySet(v []PrivilegeSummary) *PrivilegeSummarySet {
	r := make([]Result, len(v))
	for i := range v {
		r[i] = &v[i]
	}
	return &PrivilegeSummarySet{
		resultSet: resultSet{
			results: r,
			columns: []string{
				"Schema",
				"Name",
				"Type",
				"Access privileges",
				"Column privileges",
			},
		},
	}
}

func (s PrivilegeSummarySet) Get() *PrivilegeSummary {
	return s.results[s.current-1].(*PrivilegeSummary)
}

// PrivilegeSummary summarizes the privileges granted on a database object
type PrivilegeSummary struct {
	Catalog          string
	Schema           string
	Name             string
	ObjectType       string
	ObjectPrivileges ObjectPrivileges
	ColumnPrivileges ColumnPrivileges
}

func (s PrivilegeSummary) Values() []interface{} {
	return []interface{}{
		s.Catalog,
		s.Schema,
		s.Name,
		s.ObjectType,
		s.ObjectPrivileges,
		s.ColumnPrivileges,
	}
}

// ObjectPrivilege represents a privilege granted on a database object.
type ObjectPrivilege struct {
	Grantee       string
	Grantor       string
	PrivilegeType string
	IsGrantable   bool
}

// ColumnPrivilege represents a privilege granted on a column.
type ColumnPrivilege struct {
	Column        string
	Grantee       string
	Grantor       string
	PrivilegeType string
	IsGrantable   bool
}

// ObjectPrivileges represents privileges granted on a database object.
// The privileges are assumed to be sorted. Otherwise the
// String() method will fail.
type ObjectPrivileges []ObjectPrivilege

// ColumnPrivileges represents privileges granted on a column.
// The privileges are assumed to be sorted. Otherwise the
// String() method will fail.
type ColumnPrivileges []ColumnPrivilege

func (p ObjectPrivileges) Len() int      { return len(p) }
func (p ObjectPrivileges) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p ObjectPrivileges) Less(i, j int) bool {
	switch {
	case p[i].Grantee != p[j].Grantee:
		return p[i].Grantee < p[j].Grantee
	case p[i].Grantor != p[j].Grantor:
		return p[i].Grantor < p[j].Grantor
	}
	return p[i].PrivilegeType < p[j].PrivilegeType
}

// String returns a string representation of ObjectPrivileges.
// Assumes the ObjectPrivileges to be sorted.
func (p ObjectPrivileges) String() string {
	if len(p) == 0 {
		return ""
	}

	lines := []string{}
	types := []string{}
	for i := range p {
		switch {
		// Is last privilege or next privilege has new grantee or grantor; finalize line
		case i == len(p)-1 || p[i].Grantee != p[i+1].Grantee || p[i].Grantor != p[i+1].Grantor:
			types = append(types, typeStr(p[i].PrivilegeType, p[i].IsGrantable))
			lines = append(lines, lineStr(p[i].Grantee, p[i].Grantor, types))
			types = types[:0]
		default:
			types = append(types, typeStr(p[i].PrivilegeType, p[i].IsGrantable))
		}
	}
	return strings.Join(lines, "\n")
}

func (p ColumnPrivileges) Len() int      { return len(p) }
func (p ColumnPrivileges) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p ColumnPrivileges) Less(i, j int) bool {
	switch {
	case p[i].Column != p[j].Column:
		return p[i].Column < p[j].Column
	case p[i].Grantee != p[j].Grantee:
		return p[i].Grantee < p[j].Grantee
	case p[i].Grantor != p[j].Grantor:
		return p[i].Grantor < p[j].Grantor
	}
	return p[i].PrivilegeType < p[j].PrivilegeType
}

// String returns a string representation of ColumnPrivileges.
// Assumes the ColumnPrivileges to be sorted.
func (p ColumnPrivileges) String() string {
	if len(p) == 0 {
		return ""
	}

	colBlocks := []string{}
	lines := []string{}
	types := []string{}
	for i := range p {
		switch {
		// Is last privilege or next privilege has new column; finalize column block
		case i == len(p)-1 || p[i].Column != p[i+1].Column:
			types = append(types, typeStr(p[i].PrivilegeType, p[i].IsGrantable))
			lines = append(lines, "  "+lineStr(p[i].Grantee, p[i].Grantor, types))
			colBlocks = append(colBlocks, p[i].Column+":\n"+strings.Join(lines, "\n"))
			lines = lines[:0]
			types = types[:0]
		// Next privilege has new grantee or grantor; finalize line
		case p[i].Grantee != p[i+1].Grantee || p[i].Grantor != p[i+1].Grantor:
			types = append(types, typeStr(p[i].PrivilegeType, p[i].IsGrantable))
			lines = append(lines, "  "+lineStr(p[i].Grantee, p[i].Grantor, types))
			types = types[:0]
		default:
			types = append(types, typeStr(p[i].PrivilegeType, p[i].IsGrantable))
		}
	}
	return strings.Join(colBlocks, "\n")
}

// typeStr appends an asterisk suffix to grantable privileges
func typeStr(privilege string, grantable bool) string {
	if grantable {
		return privilege + "*"
	} else {
		return privilege
	}
}

// lineStr compiles grantee, grantor and privilege types into a line of output
func lineStr(grantee, grantor string, types []string) string {
	if grantor != "" {
		return grantee + "=" + strings.Join(types, ",") + "/" + grantor
	} else {
		return grantee + "=" + strings.Join(types, ",")
	}
}
