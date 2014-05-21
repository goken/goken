package ormtest

import (
	"fmt"
	"hogedigo/orm"
	"testing"
)

func TestSelect(_ *testing.T) {
	b := NewBook_meta()

	var v Book

	fmt.Print("\n\n")
	orm.SelectFrom(b).Exec(&v)

	fmt.Print("\n\n")
	orm.SelectFrom(b).
		OrderBy(b.Title.Desc()).Exec(&v)

	fmt.Print("\n\n")
	orm.SelectFrom(b).
		Where(b.Id.Equal(123).And(b.Price.GreatorThanOrEqual(3000))).
		OrderBy(b.Title.Desc()).Exec(&v)

	fmt.Print("\n\n")
	orm.SelectFrom(b).
		Where(
		b.Id.Equal(123).And(b.Price.GreatorThanOrEqual(3000)).And(
			b.Title.In([]string{"aaa", "bbb", "ccc"}))).
		OrderBy(b.Title.Desc()).Exec(&v)
}

func TestUpdate(_ *testing.T) {
	b := NewBook_meta()

	var v Book

	orm.Update(b).Exec(v)
}

func TestInsert(_ *testing.T) {
	b := NewBook_meta()

	var v Book

	orm.InsertInto(b).Exec(v)
}

func TestDelete(_ *testing.T) {
	b := NewBook_meta()

	var v Book

	orm.DeleteFrom(b).Exec(v)
}
