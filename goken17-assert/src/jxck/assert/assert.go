package assert

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func structfmt(v reflect.Value) (str string) {
	typ := v.Type()
	nf := typ.NumField()
	str += "\n{\n"

	inner := "\t"
	for i := 0; i < nf; i++ {
		tf := typ.Field(i)
		fv := v.Field(i)
		inner += fmt.Sprintf("%s:\t%s\n", tf.Name, format(fv))
	}
	inner = strings.Replace(inner, "\n", "\n\t", -1)
	inner = strings.TrimSuffix(inner, "\t")
	str += inner

	str += "}\n"
	return str
}

func strfmt(v reflect.Value) string {
	return fmt.Sprintf("%q(%s)", v.String(), v.Type().String())
}

func intfmt(v reflect.Value) string {
	return fmt.Sprintf("%d(%s)", v.Int(), v.Type().String())
}

func boolfmt(v reflect.Value) string {
	return fmt.Sprintf("%t(%s)", v.Bool(), v.Type().String())
}

func slicefmt(v reflect.Value) string {
	length := v.Len()
	slice := v.Slice(0, length)

	str := "["
	for i := 0; i < length; i += 1 {
		str = fmt.Sprintf("%s%v, ", str, format(slice.Index(i)))
	}
	str += "\b\b]"
	return fmt.Sprintf("%v(%s[%d])", str, v.Type().String(), length)
}

func format(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Int:
		return intfmt(v)
	case reflect.String:
		return strfmt(v)
	case reflect.Bool:
		return boolfmt(v)
	case reflect.Slice:
		return slicefmt(v)
	case reflect.Struct:
		return structfmt(v)
	}
	return ""
}

func getInfo() string {
		_, file, line, _ := runtime.Caller(2)
		file = filepath.Base(file)
		return fmt.Sprintf("%s:%d", file, line)
}

func Equal(t *testing.T, actual, expected interface{}) {
	if reflect.DeepEqual(actual, expected) {
		// Do Nothing while its went well.
	} else {
		av := reflect.ValueOf(actual)
		ev := reflect.ValueOf(expected)

		message := "\n"
		message += getInfo() + "\n"
		message += fmt.Sprintf("[actual]  :%s\n", format(av))
		message += fmt.Sprintf("[expected]:%s\n", format(ev))

		t.Error(message)
	}
}
