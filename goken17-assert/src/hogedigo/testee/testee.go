package testee

func IntValue(v int64) int64 {
	return v
}

func FloatValue(v float64) float64 {
	return v
}

func StrValue(v string) string {
	return v
}

func StructValue(a, b string) struct{ a, b string } {
	return struct{ a, b string }{a, b}
}
