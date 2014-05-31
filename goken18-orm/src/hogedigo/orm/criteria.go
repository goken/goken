package orm

import (
	"fmt"
)

type logicalOperator string

const (
	and logicalOperator = "and"
	or  logicalOperator = "or"
)

type Criteria interface {
	And(Criteria) Criteria
	Or(Criteria) Criteria
	sql(alias string) string
	getParams() []interface{}
}

type Criterion struct {
	Column Column
	Format string
	Params []interface{}
}

func (c Criterion) And(right Criteria) Criteria {
	return LogicCriteria{c, right, and}
}

func (c Criterion) Or(right Criteria) Criteria {
	return LogicCriteria{c, right, or}
}

func (c Criterion) sql(alias string) string {
	return fmt.Sprintf(c.Format, alias+"."+c.Column.Name)
}

func (c Criterion) getParams() []interface{} {
	return c.Params
}

type InvalidCriterion struct {
	Criterion
	err string
}

type LogicCriteria struct {
	left, right Criteria
	operator    logicalOperator
}

func (c LogicCriteria) And(right Criteria) Criteria {
	return LogicCriteria{c, right, and}
}

func (c LogicCriteria) Or(right Criteria) Criteria {
	return LogicCriteria{c, right, or}
}

func (c LogicCriteria) sql(alias string) string {
	return fmt.Sprintf("%s %s (%s)", c.left.sql(alias), c.operator, c.right.sql(alias))
}

func (c LogicCriteria) getParams() []interface{} {
	return append(c.left.getParams(), c.right.getParams()...)
}
