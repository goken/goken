package ormtest

import (
	"hogedigo/orm"
)

type Book_meta struct {
	Id, Title, Author, Price orm.Column
}

func NewBook_meta() Book_meta {
	return Book_meta{
		orm.Column{"id"},
		orm.Column{"title"},
		orm.Column{"author"},
		orm.Column{"price"},
	}
}

func (m Book_meta) Columns__() []orm.Column {
	return []orm.Column{
		m.Id,
		m.Title,
		m.Author,
		m.Price,
	}
}

func (m Book_meta) Name__(alias string) string {
	return "book " + alias
}
