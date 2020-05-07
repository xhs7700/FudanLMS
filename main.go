package main

import(
    "fmt"
    "database/sql"
    "time"
    "crypto/sha256"
    "encoding/hex"
    "log"
    _ "github.com/go-sql-driver/mysql"
)

var(
    user="root"
    password="(644000)xhs"
    now time.Time
    db *sql.DB
    err error
)

const(
    Admin=iota
    Student
    Suspended
    Guest
)

//error check
func checkErr(err error){if err!=nil{panic(err)}}

//read a complete line from stdin
func ScanLine()string{
    var(
        c byte
        err error
        b []byte
    )
    for err==nil{
        _,err=fmt.Scanf("%c",&c)
        if c!='\n'{b=append(b,c)}else{break}
    }
    return string(b)
}

//create tables in the database
func CreateTable(){
    for _,s:=range rawsql{//execute mysql statements stored in models.go in order
        _,err:=db.Exec(s)
        checkErr(err)
    }
}

//Add a book given its title,author and isbn. Only worked when using administrator account.
func (x *User)AddBook(title,author,isbn string)error{
    if x.Authority!=Admin{return fmt.Errorf("This operation is only executable by a administrator account.")}
    if len(isbn)!=13{return fmt.Errorf("ISBN length incorrect.")}//check whether isbn is correct.
    _,err:=db.Exec("insert into books (isbn,title,author)values(?,?,?)",isbn,title,author)
    if err!=nil{return fmt.Errorf("AddBook %s:%v",title,err)}
    return nil
}

//Remove a book given its isbn. Only worked when using administrator account and this book exists.
func (x *User)RemoveBook(isbn string)error{
    if x.Authority!=Admin{return fmt.Errorf("This operation is only executable by a administrator account.")}
    if len(isbn)!=13{return fmt.Errorf("ISBN length incorrect.")}//check whether isbn is correct.
    title:=""
    err:=db.QueryRow("select title from books where isbn=?;",isbn).Scan(&title)
    if err!=nil{return err}
    _,err=db.Exec("delete from books where isbn=?;",isbn)
    if err!=nil{return fmt.Errorf("RemoveBook %s:%v",title,err)}
    return nil
}

//encrypt plaintext password to a hash value
func hashcode(x string)string{
    h:=sha256.New();h.Write([]byte(x))
    return string(hex.EncodeToString(h.Sum(nil)))
}

//add a student account given its ID(11 bit) and password. Only worked when using administrator account.
func (x *User)Register(id,password string,hash func(string)string)error{
    if x.Authority!=Admin{return fmt.Errorf("This operation is only executable by a administrator account.")}
    if len(id)!=11{return fmt.Errorf("ID length incorrect.")}//check whether id is correct.
    s:=hash(password)//use sha256 to encrypt the password
    _,err:=db.Exec("insert into users(id,password,authority)values(?,?,?)",id,s,1)
    if err!=nil{return fmt.Errorf("Register %s:%v",id,err)}
    return nil
}

func (x *User)QueryBook(Value,Type string)[]Book,error{
    if x.Authority==Guest{return nil,fmt.Errorf("Guest cannot query books.")}
    BookList:=[]Book
    rows,err:=db.Query("select * from books where ? = ?",Type,Value)
    defer rows.Close()
    Error:=func()[]Book,error{return nil,fmt.Errorf("QueryBook type(%s) value(%s):%v",Type,Value,err)}
    if err!=nil{return Error()}
    for rows.Next(){
        var book Book
        err=rows.Scan(&book.ISBN,&book.Title,&book.Author)
        if err!=nil{return Error()}
        BookList=append(BookList,book)
    }
    err=rows.Err()
    if err!=nil{return Error()}
    return BookList,nil
}

func main(){
    db,err=sql.Open("mysql",fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/",user,password))
    defer db.Close()
    checkErr(err)
    if err=db.Ping();err!=nil{log.Fatal(err)}
    db.Exec("create database fudanlms;")
    //defer func(){db.Exec("drop database fudanlms;")}()
    db.Exec("use fudanlms;");
    CreateTable()

    fmt.Println("Done.")
}
