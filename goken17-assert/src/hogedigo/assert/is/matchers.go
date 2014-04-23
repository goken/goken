package is

import (
	"fmt"
	"reflect"
	"strings"
)

type Matcher interface {
	Matches(interface{}) bool
	//describeMismatch(item interface{}, mismatchDescription Description)
	DescribeExpected() string
	And(Matcher) Matcher
	Or(Matcher) Matcher
}

type DelegateMatcher struct {
	expectedDescription string
	delegate            func(actual interface{}) bool
}

func newDelegateMatcher(expectedDescription string, f func(actual interface{}) bool) Matcher {
	var m DelegateMatcher
	m.expectedDescription = expectedDescription
	m.delegate = f
	return &m
}

func (m DelegateMatcher) Matches(actual interface{}) bool {
	return m.delegate(actual)
}

func (m DelegateMatcher) DescribeExpected() string {
	return m.expectedDescription
}

func (m DelegateMatcher) And(another Matcher) Matcher {
	return newLogicMatcher(m, another, and)
}

func (m DelegateMatcher) Or(another Matcher) Matcher {
	return newLogicMatcher(m, another, or)
}

const (
	and = "and"
	or  = "or"
)

type LogicMatcher struct {
	a, b  Matcher
	logic string
}

func newLogicMatcher(a, b Matcher, logic string) LogicMatcher {
	var lm LogicMatcher
	lm.a = a
	lm.b = b
	lm.logic = logic
	return lm
}

func (m LogicMatcher) Matches(actual interface{}) bool {
	switch m.logic {
	case and:
		return m.a.Matches(actual) && m.b.Matches(actual)
	case or:
		return m.a.Matches(actual) || m.b.Matches(actual)
	default:
		panic("illegal logic: " + m.logic)
	}
}

func (m LogicMatcher) DescribeExpected() string {
	return fmt.Sprintf("(%s %s %s)", m.a.DescribeExpected(), m.logic, m.b.DescribeExpected())
}

func (m LogicMatcher) And(another Matcher) Matcher {
	return newLogicMatcher(m, another, and)
}

func (m LogicMatcher) Or(another Matcher) Matcher {
	return newLogicMatcher(m, another, or)
}

func EqualTo(o interface{}) Matcher {
	return newDelegateMatcher(fmt.Sprintf(" is equal to %v", o), func(actual interface{}) bool {
		return reflect.DeepEqual(o, actual) ||
			reflect.ValueOf(o) == reflect.ValueOf(actual) ||
			fmt.Sprintf("%#v", o) == fmt.Sprintf("%#v", actual)
	})
}

func GreaterThan(o interface{}) Matcher {
	return newDelegateMatcher(fmt.Sprintf(" is equal to %v", o), func(actual interface{}) bool {
		return reflect.ValueOf(actual).Float() > reflect.ValueOf(o).Float()
	})
}

func LessThan(o interface{}) Matcher {
	return newDelegateMatcher(fmt.Sprintf(" is equal to %v", o), func(actual interface{}) bool {
		return reflect.ValueOf(actual).Float() < reflect.ValueOf(o).Float()
	})
}

func Nil() Matcher {
	return newDelegateMatcher(fmt.Sprintf(" is nil"), func(actual interface{}) bool {
		return actual == nil || reflect.ValueOf(actual).IsNil()
	})
}

func NotNil() Matcher {
	return newDelegateMatcher(fmt.Sprintf(" is not nil"), func(actual interface{}) bool {
		return actual != nil && !reflect.ValueOf(actual).IsNil()
	})
}

func Contains(o interface {}) Matcher {
	return newDelegateMatcher(fmt.Sprintf(" contains %v", o), func(actual interface{}) bool {
			return strings.Contains(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", o))
		})
}
