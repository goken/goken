package main

import (
	"testing"
)

type User struct {
	Id    int
	Name  string
	Email string
	Age   int
}

func TestSelect(t *testing.T) {
	qb := NewQueryBuilder(&User{})
	got := qb.String()
	want := "SELECT Id, Name, Email, Age FROM User"
	if got != want {
		t.Errorf("\nGot  %q\nwant %q", got, want)
	}
}

func TestWhere(t *testing.T) {
	qb := NewQueryBuilder(&User{})
	got := qb.Eq("Id", 1).String()
	want := "SELECT Id, Name, Email, Age FROM User WHERE Id = 1"
	if got != want {
		t.Errorf("\nGot  %q\nwant %q", got, want)
	}

	qb = NewQueryBuilder(&User{})
	got = qb.Eq("Name", "yssk22").String()
	want = "SELECT Id, Name, Email, Age FROM User WHERE Name = 'yssk22'"
	if got != want {
		t.Errorf("\nGot  %q\nwant %q", got, want)
	}

	qb = NewQueryBuilder(&User{})
	got = qb.Eq("Name", nil).String()
	want = "SELECT Id, Name, Email, Age FROM User WHERE Name = NULL"
	if got != want {
		t.Errorf("\nGot  %q\nwant %q", got, want)
	}

	qb = NewQueryBuilder(&User{})
	got = qb.Eq("Name", "foo'; select * from users;").String()
	want = "SELECT Id, Name, Email, Age FROM User WHERE Name = 'foo\\'; select * from users;'"
	if got != want {
		t.Errorf("\nGot  %q\nwant %q", got, want)
	}

}
