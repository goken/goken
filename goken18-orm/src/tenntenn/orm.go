package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Query map[string]interface{}

func (query Query) Where() (string, []interface{}) {
	placeholders := make([]string, 0, len(query))
	values := make([]interface{}, 0, len(query))
	for col, value := range query {
		placeholders = append(placeholders, fmt.Sprintf("%s = ?", col))
		values = append(values, value)
	}

	return strings.Join(placeholders, " AND "), values
}

type Db struct {
	conn *sql.DB
}

func (db *Db) Get(tableName string, query Query, model interface{}) error {
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("model must be pointer")
	}

	pv := v.Elem()
	cols := make([]string, 0, pv.NumField())
	for i := 0; i < pv.NumField(); i++ {
		field := pv.Type().Field(i)
		col := field.Tag.Get("col")
		if col != "" {
			cols = append(cols, col)
		}
	}
	placeholders, values := query.Where()
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s", strings.Join(cols, ","), tableName, placeholders)
	fmt.Println(q)

	rows, err := db.conn.Query(q, values...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		dest := make([]interface{}, len(cols))
		for i := 0; i < len(cols); i++ {
			dest[i] = pv.Field(i).Addr().Interface()
		}
		if err := rows.Scan(dest...); err != nil {
			log.Fatal(err)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return nil
}

type Person struct {
	Id          int    `col:"id"`
	Name        string `col:"name"`
	_OnlyObject int
}

func main() {
	conn, _ := sql.Open("sqlite3", "./sample.db")
	db := &Db{conn}
	var person Person
	db.Get("person", Query{"id": 1}, &person)
	fmt.Println(person)
}
