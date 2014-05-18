//
// メイン処理
//
//  DB という struct にメソッドを用意
//  戻り値に 自身のポインターを返すことで、メソッドチェーンを実現
//  可変引数を使って、複数のフィールド(カラム)の処理に対応
//   ※ただし、全てを文字列型として扱ってしまっている。本来は数値などにも対応しなければ
//    → N(Number, Numeric)専用のメソッドを用意することで無理矢理対応
//   ※まだ プレースフォルダー(?) の数値用のものができていない
//


package main

import (
 //"os"
 //"io"
 "fmt"
 "strings"
 //"strconv"
)

// http://play.golang.org/p/zQSEGjZ_3A
type DB struct {
  sql string;
  builtSQL string;
  set bool;
  values bool;
  assign bool;
}

func (q *DB)Build() string {
  if q.assign {
    q.assign = false
    return q.builtSQL
  } else {
    return q.sql
  }
}

func (q *DB)Assign(values ...string) *DB {
  restSQL := ""
  if !q.assign {
    q.assign = true
    restSQL = q.sql
    q.builtSQL = ""
  } else {
    restSQL = q.builtSQL
  }
  
  q.builtSQL = ""
  valueCount := len(values)
  pos := -1

  for i := 0; i < valueCount; i++ {
    pos = strings.Index(restSQL, "?")
    if pos >= 0 {
      q.builtSQL += restSQL[0 : pos]
      q.builtSQL += values[i]
      restSQL = restSQL[pos+1 : len(restSQL)]
    }
  }
  
  q.builtSQL += restSQL
  return q
}

func (q *DB)AssignN(values ...float64) *DB {
  restSQL := ""
  if !q.assign {
    q.assign = true
    restSQL = q.sql
  } else {
    restSQL = q.builtSQL
  }
  
  q.builtSQL = ""
  valueCount := len(values)
  pos := -1

  for i := 0; i < valueCount; i++ {
    pos = strings.Index(restSQL, "?")
    if pos >= 0 {
      q.builtSQL += restSQL[0 : pos]
      q.builtSQL += fmt.Sprint(values[i])
      restSQL = restSQL[pos+1 : len(restSQL)]
    }
  }
  
  q.builtSQL += restSQL
  return q
}


func (q *DB)Select(fields ...string) *DB {
  q.sql = "SELECT "
  q.assign = false
  fieldCount := len(fields)
  for i := 0; i < fieldCount; i++ {
    q.sql += fields[i]
    if i < fieldCount - 1 {
      q.sql += ","
    }
    q.sql += " "
  }
  
  return q
}

func (q *DB)From(tables ...string) *DB {
  q.sql += "FROM "
  tableCount := len(tables)
  for i := 0; i < tableCount; i++ {
    q.sql += tables[i]
    if i < tableCount - 1 {
      q.sql += ","
    }
    q.sql += " "
  }
  
  return q
}

func (q *DB)Where(condition string) *DB {
  q.sql += "WHERE "
  q.sql += condition
  q.sql += " "
  
  return q
}

func (q *DB)Insert(table string) *DB {
  q.values = false
  q.assign = false
  q.sql = "INSERT "
  q.sql += table
  q.sql += " "
  
  return q
}

func (q *DB)Into(fields ...string) *DB {
  q.sql += "INTO ("
  fieldCount := len(fields)
  for i := 0; i < fieldCount; i++ {
    q.sql += fields[i]
    if i < fieldCount - 1 {
      q.sql += ", "
    } else {
      q.sql += ") "
    }
  }
  
  return q
}

func (q *DB)InsertInto(table string, fields ...string) *DB {
  q.values = false
  q.assign = false
  q.sql = "INSERT INTO "
  q.sql += table
  q.sql += " ("
  fieldCount := len(fields)
  for i := 0; i < fieldCount; i++ {
    q.sql += fields[i]
    if i < fieldCount - 1 {
      q.sql += ", "
    } else {
      q.sql += ") "
    }
  }
  
  return q
}

func (q *DB)Values(values ...string) *DB {
  if !q.values {
    q.values = true
    q.sql += "VALUES ("
  } else {
    q.sql = q.sql[0 : len(q.sql)-2]
    q.sql += ", "
  }

  valueCount := len(values)
  for i := 0; i < valueCount; i++ {
    q.sql += ("'" + values[i] + "'")  // use ValuesN for Number
    if i < valueCount - 1 {
      q.sql += ", "
    } else {
      q.sql += ") "
    }
  }
  
  return q
}

func (q *DB)ValuesN(values ...float64) *DB {
  if !q.values {
    q.values = true
    q.sql += "VALUES ("
  } else {
    q.sql = q.sql[0 : len(q.sql)-2]
    q.sql += ", "
  }

  valueCount := len(values)
  for i := 0; i < valueCount; i++ {
    q.sql += fmt.Sprint(values[i])  // no ' for Number
    if i < valueCount - 1 {
      q.sql += ", "
    } else {
      q.sql += ") "
    }
  }
  
  return q
}

func (q *DB)ValuesNQ() *DB {
  if !q.values {
    q.values = true
    q.sql += "VALUES ("
  } else {
    q.sql = q.sql[0 : len(q.sql)-2]
    q.sql += ", "
  }

  q.sql += "?) "  // only ? for Number
  
  return q
}


func (q *DB)Update(table string) *DB {
  q.set = false;
  q.assign = false
  q.sql = "UPDATE "
  q.sql += table
  q.sql += " "
  
  return q
}

func (q *DB)Set(field string, value string) *DB {
  if !q.set {
    q.set = true
    q.sql += "SET "
  } else {
    q.sql = q.sql[0 : len(q.sql)-1]
    q.sql += ", "
  }
  
  q.sql += (field + "='" + value + "' ")   // use SetN for Number
  return q
}

func (q *DB)SetN(field string, value float64) *DB {
  if !q.set {
    q.set = true
    q.sql += "SET "
  } else {
    q.sql = q.sql[0 : len(q.sql)-1]
    q.sql += ", "
  }
  
  q.sql += (field + "=" + fmt.Sprint(value) + " ")   // no ' for Number
  return q
}

func (q *DB)SetNQ(field string) *DB {
  if !q.set {
    q.set = true
    q.sql += "SET "
  } else {
    q.sql = q.sql[0 : len(q.sql)-1]
    q.sql += ", "
  }
  
  q.sql += (field + "=? ")   // only ? for Number
  return q
}

func (q *DB)DeleteFrom(table string) *DB {
  q.assign = false
  q.sql = "DELETE FROM "
  q.sql += table
  q.sql += " "
  
  return q
}


// --------------------------------------------------------
func main() {
  fmt.Println("--- goken18 orm query builder ---")
  
  //q := DB{} // 下との違いは？？
  q := &DB{}
  
  // SELECT
  fmt.Println(q.Select("id", "name", "email").Build())
  fmt.Println(q.Select("id", "name", "email").From("users").Build())
  fmt.Println(q.Select("id", "name", "email").From("users", "departments").Build())
  fmt.Println(q.Select("id", "name", "email").From("users", "departments").Where("user.dep = departments.name AND name LIKE 'Mike%'").Build())
  fmt.Println(q.Select("id", "name", "email").From("users").Where("name LIKE 'Mike%'").Build())
  fmt.Println(q.Select("id", "name", "email").From("users").Where("id = '?'").Assign("3").Build())

  // INSERT
  fmt.Println(q.Insert("users").Into("id", "name", "email").Build())
  fmt.Println(q.InsertInto("users", "id", "name", "email").Build())
  fmt.Println(q.Insert("users").Into("id", "name", "email").Values("2123", "Mike Smith", "simth@mail.com").Build())
  fmt.Println(q.InsertInto("users", "id", "name", "email").Values("2123", "Mike Smith", "simth@mail.com").Build())
  fmt.Println(q.InsertInto("users", "id", "name", "email").Values("?", "?", "?").Assign("1356", "Tomy John", "tomy@mail.net").Build())
  fmt.Println(q.InsertInto("users", "id", "name", "email").Values("2123").Values("Mike Smith").Values("smith@mail.com").Build())
  fmt.Println(q.InsertInto("users", "id", "name", "age", "email").Values("2123", "Mike Smith").ValuesN(31).Values("smith@mail.com").Build())
  //  ? for Number
  fmt.Println(q.InsertInto("users", "id", "name", "age", "email").Values("2123", "Mike Smith").ValuesNQ().Values("smith@mail.com").AssignN(34).Build())
  fmt.Println(q.InsertInto("users", "id", "name", "age", "email").Values("?", "?").ValuesNQ().Values("?").Assign("1356", "Tomy John").AssignN(37).Assign("tomy@mail.net").Build())
  
  
  // UPDATE
  fmt.Println(q.Update("users").Build())
  fmt.Println(q.Update("users").Set("name", "Tomy John").Build())
  fmt.Println(q.Update("users").Set("name", "Tomy John").Set("email", "tomy@mail.net").Build())
  fmt.Println(q.Update("users").Set("name", "Tomy John").Set("email", "tomy@mail.net").Where("id = '3'").Build())
  fmt.Println(q.Update("users").Set("name", "?").Set("email", "?").Where("id = '?'").Assign("Tomy John", "tomy@mail.net", "3").Build())
  fmt.Println(q.Update("users").Set("name", "?").Set("email", "?").SetN("age", 35).Where("id = '?'").Assign("Tomy John", "tomy@mail.net", "3").Build())
  // ? for Number
  fmt.Println(q.Update("users").Set("name", "?").SetNQ("age").Set("email", "?").Where("id = '?'").Assign("Tomy John").AssignN(38).Assign("tomy@mail.net", "3").Build())
  // reuse
  fmt.Println(q.Assign("Jef Beck").AssignN(28).Assign("jef@mail.net", "4").Build())
  
    
  // DELETE
  fmt.Println(q.DeleteFrom("users").Build())
  fmt.Println(q.DeleteFrom("users").Where("id = '3'").Build())
  fmt.Println(q.DeleteFrom("users").Where("id = '?'").Assign("3").Build())
  
}
