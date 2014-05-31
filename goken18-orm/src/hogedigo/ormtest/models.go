package ormtest

// +ormgen
type Book struct {
	Id     int
	Title  string
	Author string
	Price  int
}
