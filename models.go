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

type UserDB struct {
	User
	Password string
}

func (x UserDB) String() string {
	return fmt.Sprintf("insert into users values(%q,%q,%d);", x.ID, x.Password, x.Authority)
}

type BookDB Book

func (x BookDB) String() string {
	return fmt.Sprintf("insert into books values(%q,%q,%q);", x.ISBN, x.Title, x.Author)
}

type BorRecDB BorRec

func (x BorRecDB) String() string {
	bortime := x.BorTime.Format(TimeFormat)
	deadline := x.Deadline.Format(TimeFormat)
	return fmt.Sprintf("insert into borrec values(%q,%q,%q,%q,%d);", x.UserID, x.BookISBN, bortime, deadline, x.ExtendTime)
}

type RetRecDB RetRec

func (x RetRecDB) String() string {
	bortime := x.BorTime.Format(TimeFormat)
	rettime := x.RetTime.Format(TimeFormat)
	return fmt.Sprintf("insert into retrec values(%q,%q,%q,%q);", x.UserID, x.BookISBN, bortime, rettime)
}

var CreateTableSQL = []string{
	"delete from rmrec;",
	"delete from borrec;",
	"delete from retrec;",
	"delete from users;",
	"delete from books;",
	`
    create table if not exists users(
        id char(11) primary key,
        password char(64) not null,
        authority tinyint
    );
    `,
	`
    create table if not exists books(
        isbn char(13) primary key,
        title varchar(64),
        author varchar(64)
    );
    `,
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
	`
    create table if not exists rmrec(
        isbn char(13),
        title varchar(64),
        author varchar(64),
        removetime datetime,
        reason varchar(128) not null,
		primary key(isbn,removetime)
    );
    `,
}

var InsertUserData = []UserDB{
	{User{"10000000000", Admin}, "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918"},
	{User{"18307130090", Student}, "6077bcd15894379cd66224eb4053d033416d6e931edfb5bd21d3338536beb18b"},
	{User{"18307130012", Suspended}, "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"},
	{User{"20000000000", Guest}, "84983c60f7daadc1cb8698621f802c0d9f9a3c3c295c810748fb048115c186ec"},
	{User{"18307120090", Student}, "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"},
	{User{"18307110090", Student}, "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"},
}

var InsertBookData = []BookDB{
	{"DaoYuDuBai", "Jiang Xun", "9787535492821"},
	{"YeHuoJi", "Long Yingtai", "9787549550166"},
	{"1984", "George_Orwell", "9787567748996"},
	{"1984", "George_Orwell", "9787567748997"},
	{"Animal_Farm", "George_Orwell", "9787567743908"},
	{"追风筝的人", "卡勒德·胡赛尼", "9787208061644"},
	{"追风筝的人", "卡勒德.胡赛尼", "9787208060625"},
	{"删除测试", "删除", "0000298472347"},
}

var InsertBorRecData = []BorRecDB{
	{"18307130090", "9787567748997", "", time.Date(2020, 5, 12, 17, 30, 0, 0, time.Local), time.Date(2020, 6, 11, 17, 30, 0, 0, time.Local), 0},
	{"18307130090", "9787208061644", "", time.Date(2020, 5, 11, 16, 18, 37, 0, time.Local), time.Date(2020, 7, 1, 16, 18, 37, 0, time.Local), 3},
	{"10000000000", "9787567748996", "", time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2038, 1, 19, 3, 14, 8, 0, time.Local), 5},
	{"18307130012", "9787535492821", "", time.Date(2019, 12, 31, 23, 59, 59, 0, time.Local), time.Date(2020, 1, 30, 23, 59, 59, 0, time.Local), 0},
	{"18307130012", "9787549550166", "", time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2019, 2, 14, 0, 0, 0, 0, time.Local), 2},
	{"18307130012", "9787567748996", "", time.Date(2018, 6, 12, 18, 0, 0, 0, time.Local), time.Date(2018, 7, 12, 18, 0, 0, 0, time.Local), 0},
	{"18307130012", "9787567743908", "", time.Date(2020, 1, 22, 10, 0, 0, 0, time.Local), time.Date(2020, 4, 8, 0, 0, 0, 0, time.Local), 2},
}

const WelcomeText string = `Welcom to use Fudan University Library Management System(FudanLMS).
This system is based on Go and MySQL.
You can type "help" for help.`

const HelpText string = `List of all FudanLMS commands:
help		display this help text
exit		quit the shell program
quit		same as exit
lg		login the account
fdbk		search for books by their authors, titles or ISBN
chpsw		change current user's password
rg		register new account
ad		add new books
rm		remove books with reasons
borbk		borrow one book
fdrec		query borrow/returned records
ckddl		query one borrowed book's deadline
ckdue		check whether one user has overdue books
ext		extend one borrow record's deadline
ret		return one borrowed book
res		reset user's password`
