package q

import (
	"errors"
	"jxck/assert"
	"testing"
)

func TestSelect(t *testing.T) {
	db := NewDB()
	query, err := db.Select("name", "age", "email").From("users").Query()
	if err != nil {
		t.Fatal(err)
	}
	var actual string = query.String()
	var expected string = "SELECT name, age, email FROM users"
	assert.Equal(t, actual, expected)
}

func TestWhere(t *testing.T) {
	db := NewDB()
	query, err := db.Select("name").From("users").Where("id = ? and age > ?", "1", "20").Query()
	if err != nil {
		t.Fatal(err)
	}
	var actual string = query.String()
	var expected string = "SELECT name FROM users WHERE id = 1 and age > 20"
	assert.Equal(t, actual, expected)
}

func TestOrderDesc(t *testing.T) {
	db := NewDB()
	query, err := db.Select("name").From("users").Where("id = 1").OrderDesc().Query()
	if err != nil {
		t.Fatal(err)
	}
	var actual string = query.String()
	var expected string = "SELECT name FROM users WHERE id = 1 ORDER BY DESC"
	assert.Equal(t, actual, expected)
}

func TestJoin(t *testing.T) {
	db := NewDB()
	query, err := db.Select("name").From("users").Where("id = 1").Join("depts").On("users.dept_id = depts.id").OrderDesc().Query()
	if err != nil {
		t.Fatal(err)
	}
	var actual string = query.String()
	var expected string = "SELECT name FROM users WHERE id = 1 JOIN depts ON users.dept_id = depts.id ORDER BY DESC"
	assert.Equal(t, actual, expected)
}

func TestInvalidParameter(t *testing.T) {
	db := NewDB()
	query, err := db.Select("name").From("users").Where("id = ?", "1 or where 1=1; select * from user;").Query()
	if query != nil {
		t.Fatal(err)
	}
	assert.Equal(t, err, errors.New("invalid parameter"))
}
