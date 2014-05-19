package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Query builder supports SELECT statement in ORM
type QueryBuilder struct {
	t      reflect.Type
	fields []string
	where  []string
	params []interface{}
}

// Returns a QueryBuilder object.
//
// zeroValue must be a pointer value of a Struct, otherwise it panics.
//
// TODO: Any way to generate zero value from string?
//       NewQueryBuilder("Struct") better than NewQueryBuilder(&Struct{})
func NewQueryBuilder(zeroVal interface{}) *QueryBuilder {
	t := reflect.Indirect(reflect.ValueOf(zeroVal)).Type()
	fields := make([]string, t.NumField())
	for i := 0; i < len(fields); i++ {
		fields[i] = t.Field(i).Name
	}
	return &QueryBuilder{
		t,
		fields,
		make([]string, 0),
		make([]interface{}, 0),
	}
}

func (qb *QueryBuilder) Eq(field string, value interface{}) *QueryBuilder {
	qb.where = append(qb.where, fmt.Sprintf("%s = ?", field))
	qb.params = append(qb.params, value)
	return qb
}

func (qb *QueryBuilder) String() string {
	s := []string{fmt.Sprintf("SELECT %s FROM %s", strings.Join(qb.fields, ", "), qb.t.Name())}
	// Why not using PREPARE / EXEC ?
	if len(qb.where) > 0 {
		s = append(s, "WHERE")
		conds := make([]string, len(qb.where))
		for i, v := range qb.where {
			conds[i] = bind(v, qb.params[i])
		}
		s = append(s, strings.Join(conds, " AND "))
	}
	return strings.Join(s, " ")
}

func bind(where string, v interface{}) string {
	return strings.Replace(where, "?", escape(v), 1)
}

var specialChars = regexp.MustCompile("[\n\r\b\t'\"]")

func escape(v interface{}) string {
	if v == nil {
		return "NULL"
	}

	switch v.(type) {
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		if v.(bool) {
			return "true"
		} else {
			return "false"
		}
		// TODO: Time,
	default:
		s := fmt.Sprintf("%s", v)
		return fmt.Sprintf("'%s'", specialChars.ReplaceAllStringFunc(s, func(c string) string {
			switch c {
			case "\n":
				return "\\n"
			case "\r":
				return "\\r"
			case "\b":
				return "\\b"
			case "\t":
				return "\\t"
			case "\x1a":
				return "\\Z"
			default:
				return fmt.Sprintf("\\%s", c)
			}
		}))
	}
}
