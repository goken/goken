/*
- Fluent interfaceの様な感じにしたかった
- パラメータは文字列で渡すと割り切ったうえに、文字列の検査はしていない
- 時間がなかったのでSelect文だけ。。。
- 他の文にも対応しようとすると設計が破綻するかも
*/
package main

import (
	"fmt"
	"strings"
)

type Clause interface {
	Build() string
}

type SelectStatement struct {
	columns []string
}

func Select(columns ...string) *SelectStatement {
	statement := &SelectStatement{
		make([]string, len(columns)),
	}
	copy(statement.columns, columns)

	return statement
}

func (statement *SelectStatement) From(table string) *FromClause {
	clause := &FromClause{statement, table}

	return clause
}

func (statement *SelectStatement) build() string {
	return fmt.Sprintf("SELECT %s", strings.Join(statement.columns, ", "))
}

type FromClause struct {
	statement *SelectStatement
	table     string
}

func (clause *FromClause) Build() string {
	elements := make([]string, 0, 2)
	elements = append(elements, clause.statement.build())
	elements = append(elements, fmt.Sprintf("FROM %s", clause.table))

	return strings.Join(elements, " ")
}

func (clause *FromClause) Where(condition string) *WhereClause {
	where := &WhereClause{clause, condition}

	return where
}

type WhereClause struct {
	clause    *FromClause
	condition string
}

func (clause *WhereClause) Build() string {
	elements := make([]string, 0, 2)
	elements = append(elements, clause.clause.Build())
	elements = append(elements, fmt.Sprintf("WHERE %s", clause.condition))

	return strings.Join(elements, " ")
}
