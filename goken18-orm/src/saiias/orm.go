package main

/*
* goken-18課題
*
* 設計方針
* 自分のポインタを返してメソッドチェーンする.
*
* 本来はもっと厳密に型を定義して，invalidなsqlが発行されないようにすべき
*/

import (
	"strings"
)

type DB struct{
	Sql string;	
}

func (db *DB) Select(params ...string) *DB{
	column := strings.Join(params,", ")
	db.Sql += "SELECT " + column
	return db
}

func (db *DB) From(value string) *DB{
	db.Sql += " FROM " + value
	return db
}

func (db *DB) Where(cond string) *DB{
	db.Sql += " WHERE " + cond
	return db
}

func (db *DB) End() string{
	return db.Sql
}



