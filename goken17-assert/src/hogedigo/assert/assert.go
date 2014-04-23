package assert

import (
	"hogedigo/assert/is"
	"testing"
	"runtime"
	"strings"
	"fmt"
)

// JunitのassertThatを意識してみました
// usage: assert.That(actual, is.EqualTo(expected))
//
// 論理判定はmethod chainにしてみたけど読みにくくなったかも。
// AndとOrの優先順位もGoのoperatorと異なっていてわかりにくい
//
// 肝心のassertionはちょっとテキトー。
// 考慮しないといけないこといろいろありそう。桁溢れとか浮動小数比較時のdeltaとか
//
// その他工夫した点 - 行番号表示など

func That(t *testing.T, actual interface{}, m is.Matcher) {
	if !m.Matches(actual) {
		desc := location() + "\n Expected: " + m.DescribeExpected() +
			"\n     but: was %v"

		t.Errorf(desc, actual)
	}
}

func location() string {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		if index := strings.LastIndex(file, "/"); index >= 0 {
			file = file[index+1:]
		} else if index = strings.LastIndex(file, "\\"); index >= 0 {
			file = file[index+1:]
		}
	} else {
		file = "???"
		line = -1
	}
	return fmt.Sprintf("%s:%d", file, line)
}
