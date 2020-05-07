package main

import(
    "fmt"
    "time"
)


type User struct{
    ID,Password string
    Authority int
}

var AuthorityDict=map[int]string{
    0:"Admin",
    1:"Student",
    2:"Suspended",
    3:"Guest",
}

var TimeFormat="2006-01-02 15:04:05"

func (x User)String()string{return fmt.Sprintf("ID:%s\tAuthority:%s",x.ID,AuthorityDict[x.Authority])}

type Book struct{
    Title,Author,ISBN string
}

func (x Book)String()string{return fmt.Sprintf("Title:%s\tAuthor:%s\tISBN:%s",x.Title,x.Author,x.ISBN)}

type BorRec struct{
    User *User
    Book *Book
    BorTime,Deadline time.Time
    ExtendTime int
}

func (x BorRec)String()string{
    return fmt.Sprintf("UserID:%s\tBookTitle:%s\tBorTime:%s\tDeadline:%s\tExtendTime:%d",x.User.ID,x.Book.Title,x.BorTime.Format(TimeFormat),x.Deadline.Format(TimeFormat),x.ExtendTime)
}

type RetRec struct{
    User *User
    Book *Book
    BorTime,RetTime time.Time
}

func (x RetRec)String()string{
    return fmt.Sprintf("UserID:%s\tBookTitle:%s\tBorTime:%s\tRetTime:%s",x.User.ID,x.Book.Title,x.BorTime.Format(TimeFormat),x.RetTime.Format(TimeFormat))
}

var rawsql=[]string{
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
        foreign key(id)references users(id),
        foreign key(isbn)references books(isbn)
    );
    `,
    "delete from retrec;",
}
