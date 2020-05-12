package main

import (
	"fmt"
	"time"
)

type User struct {
	ID        string
	Authority int
}

var AuthorityDict = map[int]string{
	0: "Admin",
	1: "Student",
	2: "Suspended",
	3: "Guest",
}

var TimeFormat = "2006-01-02 15:04:05"

func (x User) String() string {
	return fmt.Sprintf("ID:%s\tAuthority:%s", x.ID, AuthorityDict[x.Authority])
}

type Book struct {
	Title, Author, ISBN string
}

func (x Book) String() string {
	return fmt.Sprintf("Title:%s\tAuthor:%s\tISBN:%s", x.Title, x.Author, x.ISBN)
}

type BorRec struct {
	UserID, BookISBN, BookTitle string
	BorTime, Deadline           time.Time
	ExtendTime                  int
}

func (x BorRec) String() string {
	return fmt.Sprintf("UserID:%s\tBookTitle:%s\tBookISBN:%s\tBorTime:%s\tDeadline:%s\tExtendTime:%d", x.UserID, x.BookTitle, x.BookISBN, x.BorTime.Format(TimeFormat), x.Deadline.Format(TimeFormat), x.ExtendTime)
}

type RetRec struct {
	UserID, BookISBN, BookTitle string
	BorTime, RetTime            time.Time
}

func (x RetRec) String() string {
	return fmt.Sprintf("UserID:%s\tBookTitle:%s\tBookISBN:%s\tBorTime:%s\tRetTime:%s", x.UserID, x.BookTitle, x.BookISBN, x.BorTime.Format(TimeFormat), x.RetTime.Format(TimeFormat))
}

var RawSQLStatement = []string{
	`
    create table if not exists users(
        id char(11) primary key,
        password char(64) not null,
        authority tinyint
    );
    `,
	"delete from users;",
	`
    create table if not exists books(
        isbn char(13) primary key,
        title varchar(64),
        author varchar(64)
    );
    `,
	"delete from books;",
	`
    create table if not exists borrec(
        id char(11),
        isbn char(13),
        bortime datetime,
        deadline datetime,
        extendtime tinyint,
        foreign key(id)references users(id),
        foreign key(isbn)references books(isbn)
    );
    `,
	"delete from borrec;",
	`
    create table if not exists retrec(
        id char(11),
        isbn char(13),
        bortime datetime,
        rettime datetime,
		primary key(id,isbn,bortime),
        foreign key(id)references users(id),
        foreign key(isbn)references books(isbn)
    );
    `,
	"delete from retrec;",
	`
    create table if not exists rmrec(
        isbn char(13) primary key,
        title varchar(64),
        author varchar(64),
        removetime datetime primary key,
        reason varchar(128)
    );
    `,
	"delete from rmrec;",
}
