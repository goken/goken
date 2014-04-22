package assert

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func structfmt(v reflect.Value) (str string) {
	typ := v.Type()
	nf := typ.NumField()
	str += "\n{\n"
	for i := 0; i < nf; i++ {
		sf := typ.Field(i)
		fv := v.Field(i)
		str += "    "
		str += fmt.Sprintf("%s:\t%s", sf.Name, format(fv))
		str += "\n"
	}
	str += "}\n"
	return str
}

func strfmt(v reflect.Value) string {
	return fmt.Sprintf("%s(%s)", v.String(), v.Type().String())
}

func intfmt(v reflect.Value) string {
	return fmt.Sprintf("%d(%s)", v.Int(), v.Type().String())
}

func format(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Int:
		return intfmt(v)
	case reflect.String:
		return strfmt(v)
	case reflect.Struct:
		return structfmt(v)
	}
	return ""
}

func Equal(t *testing.T, actual, expected interface{}) {
	if reflect.DeepEqual(actual, expected) {
		// Do Nothing while its went well.
	} else {
		av := reflect.ValueOf(actual)
		ev := reflect.ValueOf(expected)

		message := "\n"
		message += fmt.Sprintf("actual  : %s\n", format(av))
		message += fmt.Sprintf("expected: %s\n", format(ev))
		t.Error(message)
	}
}
