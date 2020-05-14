package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"time"
	_ "github.com/go-sql-driver/mysql"
)

var (
	user     = "root"
	password = "(644000)xhs"
	db       *sql.DB
	EmptyUser=User{"20000000000",Guest}
)

const (
	Admin = iota
	Student
	Suspended
	Guest
)

//error check
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

//create tables in the database
func CreateTable() {
	SelectDatabase()
	for _, s := range CreateTableSQL { //execute mysql statements stored in models.go in order
		db.Exec(s)
		//checkErr(err)
	}
}

func InsertData(){
	var err error
	SelectDatabase()
	for _,user:=range(InsertUserData){
		_,err=db.Exec(user.String())
		if err!=nil{checkErr(err)}
	}
	for _,book:=range(InsertBookData){
		_,err=db.Exec(book.String())
		if err!=nil{checkErr(err)}
	}
	for _,borrec:=range(InsertBorRecData){
		_,err=db.Exec(borrec.String())
		if err!=nil{checkErr(err)}
	}
}

func DropDatabase(){
	var err error
	SelectDatabase()
	_,err=db.Exec("drop database fudanlms;");checkErr(err)
}

//check whether given ID is valid(11 bit)
func IDValidate(id string) error {
	if len(id) != 11 {
		return fmt.Errorf("ID length incorrect.")
	}
	return nil
}

//check whether given ISBN is valid(13 bit)
func ISBNValidate(isbn string) error {
	if len(isbn) != 13 {
		return fmt.Errorf("ISBN length incorrect.")
	}
	return nil
}

func SelectDatabase() { db.Exec("use fudanlms;") }

//called by function execAd in order to insert data into database
func AddBook(title, author, isbn string) error {
	var err error
	err = ISBNValidate(isbn) //check whether given ISBN is valid(13 bit)
	if err != nil {
		return err
	}
	sqls := fmt.Sprintf("insert into books (isbn,title,author)values(%s,%q,%q);", isbn, title, author)
	SelectDatabase()
	_, err = db.Exec(sqls)
	if err != nil {
		return fmt.Errorf("AddBook (title:%s):%v", title, err)
	}
	return nil
}

//FindUser returns a user structure given its UserID
func FindUser(id string) (User, bool) {
	var err error
	auth := 0
	newid := ""
	raw := fmt.Sprintf("select id,authority from users where id=%s;", id)
	SelectDatabase()
	err = db.QueryRow(raw).Scan(&newid, &auth)
	switch {
	case err == sql.ErrNoRows:
		return EmptyUser, false
	case err == nil:
		return User{newid, auth}, true
	default:
		panic(fmt.Errorf("FindUser:%v", err))
	}
}

//FindBook returns a book structure given its ISBN.
func FindBook(isbn string) (Book, bool) {
	var err error
	title, author := "", ""
	raw := fmt.Sprintf("select title,author from books where isbn=%s", isbn)
	SelectDatabase()
	err = db.QueryRow(raw).Scan(&title, &author)
	switch {
	case err == sql.ErrNoRows:
		return Book{}, false
	case err == nil:
		return Book{title, author, isbn}, true
	default:
		panic(fmt.Errorf("FindBook:%v", err))
	}
}

//FindBorRec returns a borrow record structure given its UserID and BookISBN
func FindBorRec(id, isbn string) (BorRec, bool) {
	var err error
	var bortime, deadline time.Time
	extendtime := 0
	raw := fmt.Sprintf("select bortime,deadline,extendtime from borrec where id=%s and isbn=%s;", id, isbn)
	//fmt.Println(sql)
	SelectDatabase()
	err = db.QueryRow(raw).Scan(&bortime, &deadline, &extendtime)
	switch {
	case err == sql.ErrNoRows:
		return BorRec{}, false
	case err == nil:
		book, _ := FindBook(isbn) //find the book's title
		return BorRec{id, isbn, book.Title, bortime, deadline, extendtime}, true
	default:
		panic(fmt.Errorf("FindBorRec:%v", err))
	}
}

//Remove a book given its isbn. Only worked when using administrator account and this book exists.
func RemoveBook(isbn, reason string) error {
	var err error
	err = ISBNValidate(isbn) //check whether given ISBN is valid(13 bit)
	if err != nil {
		return err
	}
	book, ok := FindBook(isbn) //check whether this book exist
	if ok == false {
		return fmt.Errorf("book(isbn=%s) not exist.", isbn)
	}
	sqls := fmt.Sprintf("delete from books where isbn=%s;", isbn)
	SelectDatabase()
	_, err = db.Exec(sqls) //remove data from books table
	if err != nil {
		return fmt.Errorf("RemoveBook-delete %s:%v", isbn, err)
	}
	rawsql := "insert into rmrec values(%s,%q,%q,%q,%q)"
	now := time.Now().Format(TimeFormat)
	sqls = fmt.Sprintf(rawsql, book.ISBN, book.Title, book.Author, now, reason)
	SelectDatabase()
	_, err = db.Exec(sqls) //insert data into remove records table
	if err != nil {
		return fmt.Errorf("RemoveBook-insert %s:%v", isbn, err)
	}
	return nil
}

//encrypt plaintext password to a hash value
func hashcode(x string) string {
	h := sha256.New()
	h.Write([]byte(x))
	return string(hex.EncodeToString(h.Sum(nil)))
}

//called by function execRg in order to insert data into database
func Register(id, password string, auth int) error {
	var err error
	if password==""{
		return fmt.Errorf("Password cannot be empty.")
	}
	err = IDValidate(id) //check whether ID is valid(11 bit)
	if err != nil {
		return err
	}
	s := hashcode(password) //use sha256 to encrypt the password
	SelectDatabase()
	sqls := fmt.Sprintf("insert into users(id,password,authority) values (%s,%q,%d);", id, s, auth)
	_, err = db.Exec(sqls)
	if err != nil {
		return fmt.Errorf("Register(id:%s,psw:%s):%v", id, password, err)
	}
	return nil
}

//called by function execLg in order to verify the identity
func Login(id, psw string) (User, bool, error) {
	var err error
	err = IDValidate(id) //check if ID is valid(11 bit)
	if err != nil {
		return EmptyUser, false, err
	}
	s := hashcode(psw) //use SHA256 to encrypt the password stored in database
	userpsw, auth := "", 0
	raw := fmt.Sprintf("select password,authority from users where id=%s;", id)
	SelectDatabase()
	err = db.QueryRow(raw).Scan(&userpsw, &auth) //execute the sql statement
	switch {
	case err == sql.ErrNoRows: //ID not found in database
		return EmptyUser, false, nil
	case err == nil:
		if userpsw != s { //Input Password do not match
			return EmptyUser, false, nil
		} else { //Identity is verified.
			return User{id, auth}, true, nil
		}
	default: //error reported
		panic(fmt.Errorf("Login:%v", err))
	}
}

//called by function execRes in order to force set the password
func ResetPassword(id, password string) error {
	var err error
	err = IDValidate(id) //check whether id is correct
	if err != nil {
		return err
	}
	if _, ok := FindUser(id); ok == false { //check whether id exist in database
		return fmt.Errorf("RestorePassword(id:%s):id not existed.", id)
	}
	s := hashcode(password) //use SHA256 to encrypt the given password
	rawsql := `
    update users
    set password=%q
    where id=%s
    `
	sqls := fmt.Sprintf(rawsql, s, id)
	SelectDatabase()
	_, err = db.Exec(sqls)
	if err != nil {
		return fmt.Errorf("RestorePassword(id:%s):%v", id, err)
	}
	return nil
}

//called by function execChpsw in order to verify and modify data in database
func ChangePassword(id, oldpsw, newpsw string) (bool, error) {
	var err error
	var ok bool
	_, ok, err = Login(id, oldpsw) //call Login to verify
	if err != nil {
		return false, fmt.Errorf("ChangePassword-Login(id:%s,oldpsw:%s,newpsw:%s):%v", id, oldpsw, newpsw, err)
	}
	if ok == false {
		return false, nil
	}
	err = ResetPassword(id, newpsw) //call ResetPassword to set new password
	if err != nil {
		return false, fmt.Errorf("ChangePassword-Restore(id:%s):%v", id, err)
	}
	return true, nil
}

//called by function execFdbk in order to query books by title author or ISBN
func QueryBook(Value, Type string) ([]Book, error) { //Type is in {isbn,author,title}
	var err error
	var rows *sql.Rows
	//fmt.Println(Value,Type)
	BookList := []Book{} //store the selected books

	//Define the standard error-report function
	Error := func() ([]Book, error) { return nil, fmt.Errorf("QueryBook type(%s) value(%s):%v", Type, Value, err) }

	if Type == "isbn" && Value != "*" { //value "*" means select all the books in database
		err = ISBNValidate(Value) //check whether ISBN is valid(13 bit)
		if err != nil {
			return nil, err
		}
	}
	var sqls string
	if Value == "*" { //value "*" means select all the books in database
		sqls = fmt.Sprintf("select * from books;")
	} else {
		sqls = fmt.Sprintf("select * from books where %s=%q;", Type, Value)
	}
	//fmt.Println(sqls)
	SelectDatabase()
	rows, err = db.Query(sqls)
	if err != nil {
		return Error()
	}
	defer rows.Close()
	for rows.Next() { //read books info from returned query info
		var book Book
		err = rows.Scan(&book.ISBN, &book.Title, &book.Author)
		//fmt.Println(book)
		if err != nil {
			return Error()
		}
		BookList = append(BookList, book)
	}
	err = rows.Err()
	if err != nil {
		return Error()
	}
	//fmt.Println(BookList)
	//fmt.Println("")
	return BookList, nil
}

//called by function execBorbk in order to borrow books by ID and ISBN
func BorrowBook(id, isbn string, intime time.Time) error {
	var err error
	var ok bool
	err = IDValidate(id) //check whether given ID is valid(11 bit)
	if err != nil {
		return err
	}
	err = ISBNValidate(isbn) //check whether given ISBN is valid(13 bit)
	if err != nil {
		return err
	}
	if _, ok = FindBook(isbn); ok == false { //check whether this book exist
		return fmt.Errorf("book(isbn=%s) not exist.", isbn)
	}
	if _, ok = FindBorRec(id, isbn); ok == true { //check whether this book has been borrowed by this user and not returned yet
		return fmt.Errorf("Borrow record(id:%s,isbn:%s) already exist.", id, isbn)
	}
	rawsql := `
    insert into borrec(id,isbn,bortime,deadline,extendtime)values
        (%s,%s,%q,%q,0);
    `
	now := intime
	ddl := now.AddDate(0, 0, 30) //Deadline is one month afterwards
	sqls := fmt.Sprintf(rawsql, id, isbn, now.Format(TimeFormat), ddl.Format(TimeFormat))
	//fmt.Println(sqls)
	SelectDatabase()
	_, err = db.Exec(sqls)
	if err != nil {
		return fmt.Errorf("User %s borrows book %s:%v", id, isbn, err)
	}
	return nil
}

//called by function execFdrec in order to query borrow records
func BorRecQuery(id string) ([]BorRec, error) {
	var err error
	var rows *sql.Rows

	//a wrapped function that returns the standard error
	Error := func() ([]BorRec, error) { return nil, fmt.Errorf("BorRecQuery(id:%s):%v", id, err) }

	err = IDValidate(id) //check whether given ID is valid(11 bit)
	if err != nil {
		return nil, err
	}
	if _, ok := FindUser(id); ok == false { //check whether id exist in database
		return nil,fmt.Errorf("RestorePassword(id:%s):id not existed.", id)
	}
	var BorRecList []BorRec
	var bortime, deadline time.Time
	isbn, extendtime := "", 0
	sqls := fmt.Sprintf("select isbn,bortime,deadline,extendtime from borrec where id=%s", id)
	SelectDatabase()
	rows, err = db.Query(sqls)
	if err != nil {
		return Error()
	}
	defer rows.Close()

	for rows.Next() { //read record info
		err = rows.Scan(&isbn, &bortime, &deadline, &extendtime)
		if err != nil {
			return Error()
		}
		book, _ := FindBook(isbn) //gather the book's complete info to make a borrow record structure
		borrec := BorRec{id, isbn, book.Title, bortime, deadline, extendtime}
		BorRecList = append(BorRecList, borrec)
	}
	err = rows.Err()
	if err != nil {
		return Error()
	}
	return BorRecList, nil
}

//called by function execFdrec in order to query returned records
func RetRecQuery(id string) ([]RetRec, error) {
	var err error
	var rows *sql.Rows

	//a wrapped function that returns the standard error
	Error := func() ([]RetRec, error) { return nil, fmt.Errorf("RetRecQuery(id:%s):%v", id, err) }

	err = IDValidate(id) //check whether given ID is valid(11 bit)
	if err != nil {
		return nil, err
	}
	var RetRecList []RetRec
	var bortime, rettime time.Time
	isbn := ""
	sqls := fmt.Sprintf("select isbn,bortime,rettime from retrec where id=%s", id)
	SelectDatabase()
	rows, err = db.Query(sqls)
	if err != nil {
		return Error()
	}
	defer rows.Close()

	for rows.Next() { //read record info
		err = rows.Scan(&isbn, &bortime, &rettime)
		if err != nil {
			return Error()
		}
		book, _ := FindBook(isbn) //gather this book's complete info to make a returned record structure
		retrec := RetRec{id, isbn, book.Title, bortime, rettime}
		RetRecList = append(RetRecList, retrec)
	}
	err = rows.Err()
	if err != nil {
		return Error()
	}
	return RetRecList, nil
}

//called by function execCkddl in order to query one borrow record's deadline
func GetDeadline(id, isbn string) (BorRec, error) {
	var err error
	err = IDValidate(id) //check whether ID is valid(11 bit)
	if err != nil {
		return BorRec{}, err
	}
	err = ISBNValidate(isbn) //check whether ISBN is valid(13 bit)
	if err != nil {
		return BorRec{}, err
	}
	borrec, ok := FindBorRec(id, isbn) //call the function to find the corresponding borrow record
	if ok == false {
		return BorRec{}, fmt.Errorf("GetDeadline(id:%s,isbn:%s):cannot find such BorRec.", id, isbn)
	}
	return borrec, nil
}

//called by function execExt in order to extend one borrow record's deadline
func ExtendDeadline(id, isbn string, auth, weeks int) (BorRec, error) {
	var err error
	var deadline time.Time
	var extendtime int
	err = IDValidate(id) //check whether ID is valid(11 bit)
	if err != nil {
		return BorRec{}, err
	}
	err = ISBNValidate(isbn) //check whether ISBN is valid(13 bit)
	if err != nil {
		return BorRec{}, err
	}
	borrec, ok := FindBorRec(id, isbn) //call the function to find the corresponding borrow record
	if ok == false {
		return BorRec{}, fmt.Errorf("ExtendDeadline(id:%s,isbn:%s):cannot find such BorRec.", id, isbn)
	}
	extendtime, deadline = borrec.ExtendTime, borrec.Deadline
	if extendtime >= 3 && auth != Admin { //Student can only extend up to 3 times
		return BorRec{}, fmt.Errorf("the deadline of (id:%s,isbn:%s) has been extended for 3 times.", id, isbn)
	}
	if auth == Admin {
		deadline = deadline.AddDate(0, 0, 7*weeks) //administrator can extend any weeks once
	} else {
		deadline = deadline.AddDate(0, 0, 7) //student can only extend a week once
		extendtime++
	}
	rawsql := `
    update borrec
    set deadline=%q , extendtime=%d
    where id=%s and isbn=%s;
    `
	sqls := fmt.Sprintf(rawsql, deadline.Format(TimeFormat), extendtime, id, isbn)
	SelectDatabase()
	_, err = db.Exec(sqls)
	if err != nil {
		return BorRec{}, fmt.Errorf("ExtendDeadline (id:%s,isbn:%s):%v", id, isbn, err)
	}
	borrec.ExtendTime, borrec.Deadline = extendtime, deadline //update the borrow record structure
	return borrec, nil
}

//called by function execCkdue in order to query one user's overdue records
func OverdueCheck(id string) ([]BorRec, error) {
	var err error
	var rows *sql.Rows
	Error := func() ([]BorRec, error) { return nil, fmt.Errorf("OverdueCheck(id:%s):%v", id, err) }
	err = IDValidate(id) //check whether the ID is valid(11 bit)
	if err != nil {
		return nil, err
	}
	var BorRecList []BorRec
	sqls := fmt.Sprintf("select isbn from borrec where id=%s and deadline < %q", id, time.Now().Format(TimeFormat))
	SelectDatabase()
	//fmt.Println(sqls)
	rows, err = db.Query(sqls)
	if err != nil {
		return Error()
	}
	defer rows.Close()
	isbn := ""

	//read info
	for rows.Next() {
		err = rows.Scan(&isbn)
		if err != nil {
			return Error()
		}
		borrec, _ := FindBorRec(id, isbn) //call the function to gather the record's complete info
		BorRecList = append(BorRecList, borrec)
	}
	err = rows.Err()
	if err != nil {
		return Error()
	}
	return BorRecList, nil
}

//called by function execRet in order to return one borrowed book
func ReturnBook(id, isbn string) error {
	var err error
	err = IDValidate(id) //check whether ID is valid(11 bit)
	if err != nil {
		return err
	}
	err = ISBNValidate(isbn) //check whether ISBN is valid(13 bit)
	if err != nil {
		return err
	}
	borrec, ok := FindBorRec(id, isbn) //call the function to find the corresponding borrow record
	if ok == false {
		return fmt.Errorf("ReturnBook(id:%s,isbn:%s):cannot find such BorRec.", id, isbn)
	}
	sqls := fmt.Sprintf("delete from borrec where id=%s and isbn=%s;", id, isbn)
	SelectDatabase()
	_, err = db.Exec(sqls) //delete data from borrow records table
	if err != nil {
		return fmt.Errorf("ReturnBook(id:%s,isbn:%s):%v", id, isbn, err)
	}
	rawsql := "insert into retrec(id,isbn,bortime,rettime) values (%s,%s,%q,%q);"
	sqls = fmt.Sprintf(rawsql, id, isbn, borrec.BorTime.Format(TimeFormat), time.Now().Format(TimeFormat))
	SelectDatabase()
	_, err = db.Exec(sqls) //insert data from returned records table
	if err != nil {
		return fmt.Errorf("ReturnBook(id:%s,isbn:%s):%v", id, isbn, err)
	}
	return nil
}

//called by ShellMain at the end of every loop to check whether the account should be suspended.
func (x User) SuspendCheck() (User, error) {
	var err error
	var BorRecList []BorRec
	if x.Authority == Admin || x.Authority == Guest {
		return x, nil
	}
	//fmt.Printf("authority=%s\n",AuthorityDict[x.Authority])
	BorRecList, err = OverdueCheck(x.ID)
	if err != nil {
		return x, fmt.Errorf("SuspendCheck(id:%s):%v", x.ID, err)
	}
	auth := 0
	if len(BorRecList) > 3 {
		if x.Authority == Suspended {
			return x, nil
		}
		auth = Suspended
	} else {
		if x.Authority == Student {
			return x, nil
		}
		auth = Student
	}
	rawsql := `
	update users
	set authority=%d
	where id=%s
	`
	sqls := fmt.Sprintf(rawsql, auth, x.ID)
	_, err = db.Exec(sqls)
	if err != nil {
		return x, fmt.Errorf("SuspendCheck(id:%s):%v", x.ID, err)
	}
	x.Authority = auth
	return x, nil
}

func init(){
	var err error
	rawsql := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/?charset=utf8&loc=%s&parseTime=true",
		user,
		password,
		url.QueryEscape("Asia/Shanghai"))
	db, err = sql.Open("mysql", rawsql)
	checkErr(err)
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	db.Exec("create database fudanlms;")
	CreateTable()
	InsertData()
}

func main() {
	defer db.Close()
	//defer DropDatabase()
	ShellMain()
}
