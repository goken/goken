package orm

import (
	"reflect"
	"strings"
)

type orderType string

const (
	asc  orderType = "ASC"
	desc orderType = "DESC"
)

type Table interface {
	Name__(alias string) string
	Columns__() []Column
}

type Column struct {
	Name string
}

func (c Column) Equal(p interface{}) Criteria {
	return Criterion{c, "%s = ?", []interface{}{p}}
}

func (c Column) NotEqual(p interface{}) Criteria {
	return Criterion{c, "%s <> ?", []interface{}{p}}
}

func (c Column) GreatorThan(p interface{}) Criteria {
	return Criterion{c, "%s > ?", []interface{}{p}}
}

func (c Column) LessThan(p interface{}) Criteria {
	return Criterion{c, "%s < ?", []interface{}{p}}
}

func (c Column) GreatorThanOrEqual(p interface{}) Criteria {
	return Criterion{c, "%s >= ?", []interface{}{p}}
}

func (c Column) LessThanOrEqual(p interface{}) Criteria {
	return Criterion{c, "%s <= ?", []interface{}{p}}
}

func (c Column) In(p interface{}) Criteria {
	v := reflect.ValueOf(p)
	if v.Kind() != reflect.Slice {
		return InvalidCriterion{
			Criterion: Criterion{c, "%s IN {ERROR}", []interface{}{p}},
			err:       "p must be a slice.",
		}
	}
	paramLen := v.Len()
	if paramLen == 0 {
		return InvalidCriterion{
			Criterion: Criterion{c, "%s IN {ERROR}", []interface{}{p}},
			err:       "p must be a slice.",
		}
	}

	placeHolders := strings.Repeat(", ?", paramLen)[2:]

	params := make([]interface{}, paramLen)
	for i := 0; i < paramLen; i++ {
		params[i] = v.Index(i).Interface()
	}

	return Criterion{c, "%s IN (" + placeHolders + ")", params}
}

func (c Column) IsNull() Criteria {
	return Criterion{c, "%s IS NULL", []interface{}{}}
}

func (c Column) IsNotNull() Criteria {
	return Criterion{c, "%s IS NOT NULL", []interface{}{}}
}

func (c Column) Asc() order {
	return order{c, asc}
}

func (c Column) Desc() order {
	return order{c, desc}
}
