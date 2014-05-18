package db

func Select(fields ...string) string {
//func Select() string {
  s := "select "
  fieldCount := len(fields)
  for i := 0; i < fieldCount; i++ {
    s = s + fields[i]
    if i < fieldCount - 1 {
      s = s + ","
    }
    s = s + " "
  }
  return s
}
