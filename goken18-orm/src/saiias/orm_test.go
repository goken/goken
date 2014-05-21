package main

import (
	"testing"
)

func TestSelect(t *testing.T){
	db := &DB{}
	want := "SELECT id, name, email FROM User WHERE id = 1"
	accuary := db.Select("id","name","email").From("User").Where("id = 1").End()

	if accuary != want {
		t.Errorf("\nGot  %q\nwant %q", accuary, want)
	}
}


