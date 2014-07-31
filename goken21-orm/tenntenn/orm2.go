package main

import (
	"fmt"
	"strings"
)

type DB struct {
}

func NewDB() *DB {
	return &DB{}
}

func (db *DB) Select(cols ...string) *Select {
	return &Select{
		cols: cols,
	}
}

type Select struct {
	cols []string
}

func (s *Select) String() string {
	return fmt.Sprintf("SELECT %s", strings.Join(s.cols, ", "))
}

func (s *Select) From(table string) *From {
	return &From{
		base:  s.String(),
		table: table,
	}
}

type From struct {
	base  string
	table string
}

func (f *From) String() string {
	return fmt.Sprintf("FROM %s", f.table)
}

func (f *From) Where(cond Cond) *Where {
	return &Where{
		base: fmt.Sprintf("%s %s", f.base, f.String()),
		cond: cond,
	}
}

type Where struct {
	base string
	cond Cond
}

func (w *Where) Build() string {
	return fmt.Sprintf("%s Where %s", w.base, w.cond.String())
}

type Cond fmt.Stringer

type AND []Cond

func (a AND) String() string {
	s := make([]string, 0, len(a))
	for _, c := range a {
		s = append(s, c.String())
	}
	return strings.Join(s, " AND ")
}

type OR []Cond

func (o OR) String() string {
	s := make([]string, len(o))
	for _, c := range o {
		s = append(s, c.String())
	}
	return strings.Join(s, " OR ")
}

type S string

func (s S) String() string {
	return string(s)
}

type Eq struct {
	Key   string
	Value fmt.Stringer
}

func (eq Eq) String() string {
	return fmt.Sprintf("%s = %s", eq.Key, eq.Value.String())
}

func main() {
	sql := NewDB().Select("id", "name").From("person").Where(AND{
		S("age >= 18"),
		S("id > 100"),
	}).Build()
	fmt.Println(sql)
}
