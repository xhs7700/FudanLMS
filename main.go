package main

import(
    "fmt"
    "database/sql"
    "time"
    "crypto/sha256"
    "encoding/hex"
    "log"
    //"strings"
    _ "github.com/go-sql-driver/mysql"
)

var(
    user="root"
    password="(644000)xhs"
    db *sql.DB
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

func IDValidate(id string)error{
    if len(id)!=11{
        return fmt.Errorf("ID length incorrect.")
    }
    return nil
}

func ISBNValidate(isbn string)error{
    if len(isbn)!=13{
        if len(isbn)!=13{return fmt.Errorf("ISBN length incorrect.")}
    }
    return nil
}

func (x *User)UserValidate(Type int,Value string)error{
    if x.Authority!=Type{
        return fmt.Errorf("This operation is only executable by a %s account.",Value)
    }
    return nil
}

func (x *User)UserMultiValidate(Type []int,Value string)error{
    for _,t:=range Type{
        if x.Authority==t{return nil}
    }
    return fmt.Errorf("This operation is only executable by a %s account.",Value)
}

func SelectDababase(){db.Exec("use fudanlms;")}

//Add a book given its title,author and isbn. Only worked when using administrator account.
func (x *User)AddBook(title,author,isbn string)error{
    var err error
    err=x.UserValidate(Admin,"admin");if err!=nil{return err}
    err=ISBNValidate(isbn);if err!=nil{return err}//check whether isbn is correct.
    _,err=db.Exec("insert into books (isbn,title,author)values(?,?,?)",isbn,title,author)
    if err!=nil{return fmt.Errorf("AddBook %s:%v",title,err)}
    return nil
}

func IsFind(err error,name string)bool{
    switch{
    case err==sql.ErrNoRows:
        return false
    case err==nil:
        return true
    default:
        panic(fmt.Errorf("%s:%v",name,err))
    }
}

func FindStudent(id string)bool{
    var err error
    auth:=0
    err=db.QueryRow("select authority from users where id=?;",id).Scan(&auth)
    return IsFind(err,"FindStudent")
}

//FindBook returns whether a book exists given its ISBN.
func FindBook(isbn string)bool{
    var err error
    title:=""
    err=db.QueryRow("select title from books where isbn=?;",isbn).Scan(&title)
    return IsFind(err,"FindBook")
}

func FindRec(id,isbn,recname string)bool{
    var err error
    find_id:=""
    sql:=fmt.Sprintf("select id from %s where id=%s and isbn=%s;",recname,id,isbn)
    //fmt.Println(sql)
    err=db.QueryRow(sql).Scan(&find_id)
    return IsFind(err,"Find"+recname)
}

//Remove a book given its isbn. Only worked when using administrator account and this book exists.
func (x *User)RemoveBook(isbn string)error{
    var err error
    err=x.UserValidate(Admin,"admin");if err!=nil{return err}
    err=ISBNValidate(isbn);if err!=nil{return err}//check whether isbn is correct.
    if FindBook(isbn)==false{return fmt.Errorf("book(isbn=%s) not exist.",isbn)}
    _,err=db.Exec("delete from books where isbn=?;",isbn)
    if err!=nil{return fmt.Errorf("RemoveBook %s:%v",isbn,err)}
    return nil
}

//encrypt plaintext password to a hash value
func hashcode(x string)string{
    h:=sha256.New();h.Write([]byte(x))
    return string(hex.EncodeToString(h.Sum(nil)))
}

//add a student account given its ID(11 bit) and password. Only worked when using administrator account.
func (x *User)Register(id,password string,hash func(string)string)error{
    var err error
    err=x.UserValidate(Admin,"admin");if err!=nil{return err}
    err=IDValidate(id);if err!=nil{return err}//check whether id is correct.
    s:=hash(password)//use sha256 to encrypt the password
    db.Exec("use fudanlms;")
    _,err=db.Exec("insert into users(id,password,authority)values(?,?,?)",id,s,1)
    if err!=nil{return fmt.Errorf("Register %s:%v",id,err)}
    return nil
}

//query books by title author or ISBN
func (x *User)QueryBook(Value,Type string)([]Book,error){
    var err error
    var rows *sql.Rows
    err=x.UserMultiValidate([]int{Admin,Student,Suspended},"admin/stu/suspended");if err!=nil{return nil,err}
    BookList:=[]Book{}
    Error:=func()([]Book,error){return nil,fmt.Errorf("QueryBook type(%s) value(%s):%v",Type,Value,err)}
    sql:=fmt.Sprintf("select * from books where %s=%s",Type,Value)
    db.Exec("use fudanlms;")
    rows,err=db.Query(sql)
    if err!=nil{return Error()}
    defer rows.Close()
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

func (x *User)BorrowBook(isbn string)error{
    var err error
    err=x.UserValidate(Student,"stu");if err!=nil{return err}
    err=ISBNValidate(isbn);if err!=nil{return err}//check whether isbn is correct.
    if FindBook(isbn)==false{return fmt.Errorf("book(isbn=%s) not exist.",isbn)}
    if FindRec(x.ID,isbn,"borrec")==true{return fmt.Errorf("Borrow record(id:%s,isbn:%s) already exist.",x.ID,isbn)}
    rawsql:=`
    insert into borrec(id,isbn,bortime,deadline,extendtime)values
        (%s,%s,%q,%q,0);
    `
    now:=time.Now()
    ddl:=now.AddDate(0,0,30)
    sql:=fmt.Sprintf(rawsql,x.ID,isbn,now.Format(TimeFormat),ddl.Format(TimeFormat))
    //fmt.Println(sql)
    _,err=db.Exec(sql)
    if err!=nil{return fmt.Errorf("User %s borrows book %s:%v",x.ID,isbn,err)}
    return nil
}

func (x *User)BorrowQuery(Type string)([]Book,error){
    var err error
    var rows *sql.Rows
    Error:=func()([]Book,error){return nil,fmt.Errorf("BorrowQuery(Type:%s):%v",Type,err)}
    err=x.UserMultiValidate([]int{Student,Suspended},"stu/suspended");if err!=nil{return Error()}
    var BookList,bookList []Book
    isbn:=""
    sql:=fmt.Sprintf("select isbn from %s where id=%s",Type,x.ID)
    db.Exec("use fudanlms;")
    rows,err=db.Query(sql)
    if err!=nil{return Error()}
    defer rows.Close()
    for rows.Next(){
        err=rows.Scan(&isbn);if err!=nil{return Error()}
        bookList,err=x.QueryBook(isbn,"isbn");if err!=nil{return Error()}
        BookList=append(BookList,bookList...)
    }
    err=rows.Err();if err!=nil{return Error()}
    return BookList,nil
}

func main(){
    var err error
    db,err=sql.Open("mysql",fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/",user,password))
    defer db.Close()
    checkErr(err)
    if err=db.Ping();err!=nil{log.Fatal(err)}
    //db.Exec("create database fudanlms;")
    //defer func(){db.Exec("drop database fudanlms;")}()
    db.Exec("use fudanlms;");
    //CreateTable()
    //adminUser:=User{"Admin","123456",Admin}
    stuUser:=User{"18307130090","(644000)xhs",Student}
    for _,isbn:=range []string{"9787535492821","9787549550166","9787567748996","9787567748997"}{
        err=stuUser.BorrowBook(isbn)
        checkErr(err)
    }
    var BookList []Book
    BookList,err=stuUser.BorrowQuery("borrec")
    checkErr(err)
    db.Exec("use fudanlms;")
    _,err=db.Exec("delete from borrec;")
    checkErr(err)
    fmt.Println(BookList)
    fmt.Println("Done.")
}
