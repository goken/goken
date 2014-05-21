package orm

import (
	"bytes"
	"fmt"
	"log"
)

func SelectFrom(t Table) postFrom {
	return queryImpl{t, nil, nil}
}

type postFrom interface {
	Where(Criteria) postWhere
	postWhere
}

type postWhere interface {
	OrderBy(...order) query
	query
}

type query interface {
	Exec(dst interface{}) error
}

type order struct {
	column    Column
	orderType orderType
}

func (o order) sql(alias string) string {
	return fmt.Sprintf("%s.%s %s", alias, o.column.Name, o.orderType)
}

type queryImpl struct {
	rootTable Table
	criteria  Criteria
	orders    []order
}

func (q queryImpl) Where(c Criteria) postWhere {
	q.criteria = c
	return q
}

func (q queryImpl) OrderBy(o ...order) query {
	q.orders = o
	return q
}

func (q queryImpl) Exec(dst interface{}) error {
	alias := "a"
	var buf bytes.Buffer

	buf.WriteString("SELECT ")
	for i, c := range q.rootTable.Columns__() {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(alias)
		buf.WriteString(".")
		buf.WriteString(c.Name)
	}

	buf.WriteString(" FROM ")
	buf.WriteString(q.rootTable.Name__(alias))

	if q.criteria != nil {
		buf.WriteString(" WHERE ")
		buf.WriteString(q.criteria.sql(alias))
		buf.WriteString(" ")
		buf.WriteString(alias)
	}

	if len(q.orders) > 0 {
		buf.WriteString(" ORDER BY ")
		for i, o := range q.orders {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(o.sql(alias))
		}
	}

	if q.criteria != nil {
		log.Printf("sql:%s; params:%v", buf.String(), q.criteria.getParams())
	} else {
		log.Printf("sql:%s", buf.String())
	}

	return nil
}

func Update(t Table) updateStatement {
	return updateImpl{}
}

type updateStatement interface {
	Exec(src interface{}) error
}

type updateImpl struct {
}

func (q updateImpl) Exec(src interface{}) error {
	// under construction...
	return nil
}

func InsertInto(t Table) insertStatement {
	return insertImpl{}
}

type insertStatement interface {
	Exec(src interface{}) error
}

type insertImpl struct {
}

func (q insertImpl) Exec(src interface{}) error {
	// under construction...
	return nil
}

func DeleteFrom(t Table) deleteStatement {
	return deleteImpl{}
}

type deleteStatement interface {
	Exec(key interface{}) error
}

type deleteImpl struct {
}

func (q deleteImpl) Exec(key interface{}) error {
	// under construction...
	return nil
}
