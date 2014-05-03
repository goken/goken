package q

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type STATEMENT string

const (
	SELECT STATEMENT = "SELECT"
	FROM             = "FROM"
	WHERE            = "WHERE"
	ORDER            = "ORDER BY"
	JOIN             = "JOIN"
	ON               = "ON"
)

const (
	COMMA = ","
	SPACE = " "
	ASC   = "ASC"
	DESC  = "DESC"
)

func NewDB() *DB {
	return &DB{
		query: new(Query),
	}
}

type DB struct {
	query *Query
}

type Query struct {
	str string
	err error
}

func (q *Query) Start(stmt STATEMENT, literal string) {
	q.str += fmt.Sprintf("%s %s", stmt, literal)
}

func (q *Query) Append(stmt STATEMENT, literal string) {
	q.str += fmt.Sprintf(" %s %s", stmt, literal)
}

func (q *Query) String() string {
	return q.str
}

func (db *DB) Select(str ...string) *DB {
	s := strings.Join(str, COMMA+SPACE)
	db.query.Start(SELECT, s)
	return db
}

func (db *DB) From(str string) *DB {
	db.query.Append(FROM, str)
	return db
}

func escape(parameter string) (string, error) {
	NG := strings.ContainsAny(parameter, " `~!@#$%^&*()-=_+{}[]\\|:;\"'<>,.?/")
	if NG {
		return "", errors.New("invalid parameter")
	}
	return parameter, nil
}

func (db *DB) Where(str string, values ...string) *DB {
	for _, v := range values {
		parameter, err := escape(v)
		if err != nil {
			db.query.err = err
		}
		str = strings.Replace(str, "?", parameter, 1)
	}
	db.query.Append(WHERE, str)
	return db
}

func (db *DB) Join(str string) *DB {
	db.query.Append(JOIN, str)
	return db
}

func (db *DB) On(str string) *DB {
	db.query.Append(ON, str)
	return db
}

func (db *DB) OrderAsc() *DB {
	db.query.Append(ORDER, ASC)
	return db
}

func (db *DB) OrderDesc() *DB {
	db.query.Append(ORDER, DESC)
	return db
}

func (db *DB) Query() (*Query, error) {
	if db.query.err != nil {
		return nil, db.query.err
	}
	return db.query, nil
}
