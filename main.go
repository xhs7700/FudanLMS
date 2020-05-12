package main

import(
    "fmt"
    "database/sql"
    "time"
    "crypto/sha256"
    "encoding/hex"
    "log"
    "net/url"
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

//create tables in the database
func CreateTable(){
    SelectDatabase()
    for _,s:=range RawSQLStatement{//execute mysql statements stored in models.go in order
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

func SelectDatabase(){db.Exec("use fudanlms;")}

//Add a book given its title,author and isbn. Only worked when using administrator account.
func AddBook(title,author,isbn string)error{
    var err error
    err=ISBNValidate(isbn);if err!=nil{return err}//check whether isbn is correct.
    sqls:=fmt.Sprintf("insert into books (isbn,title,author)values(%s,%q,%q);",isbn,title,author)
    SelectDatabase();_,err=db.Exec(sqls)
    if err!=nil{return fmt.Errorf("AddBook (title:%s):%v",title,err)}
    return nil
}

func FindUser(id string)(User,bool){
    var err error
    auth:=0;newid:=""
    raw:=fmt.Sprintf("select id,authority from users where id=%s;",id)
    SelectDatabase()
    err=db.QueryRow(raw).Scan(&newid,&auth)
    switch{
    case err==sql.ErrNoRows:
        return User{},false
    case err==nil:
        return User{newid,auth},true
    default:
        panic(fmt.Errorf("FindUser:%v",err))
    }
}

//FindBook returns whether a book exists given its ISBN.
func FindBook(isbn string)(Book,bool){
    var err error
    title,author:="",""
    raw:=fmt.Sprintf("select title,author from books where isbn=%s",isbn)
    SelectDatabase()
    err=db.QueryRow(raw).Scan(&title,&author)
    switch{
    case err==sql.ErrNoRows:
        return Book{},false
    case err==nil:
        return Book{title,author,isbn},true
    default:
        panic(fmt.Errorf("FindBook:%v",err))
    }
}

func FindBorRec(id,isbn string)(BorRec,bool){
    var err error
    var bortime,deadline time.Time
    extendtime:=0
    raw:=fmt.Sprintf("select bortime,deadline,extendtime from borrec where id=%s and isbn=%s;",id,isbn)
    //fmt.Println(sql)
    SelectDatabase()
    err=db.QueryRow(raw).Scan(&bortime,&deadline,&extendtime)
    switch{
    case err==sql.ErrNoRows:
        return BorRec{},false
    case err==nil:
        book,_:=FindBook(isbn)
        return BorRec{id,isbn,book.Title,bortime,deadline,extendtime},true
    default:
        panic(fmt.Errorf("FindBorRec:%v",err))
    }
}

func FindRetRec(id,isbn string)(RetRec,bool){
    var err error
    var bortime,rettime time.Time
    raw:=fmt.Sprintf("select bortime,rettime from retrec where id=%s and isbn=%s;",id,isbn)
    SelectDatabase()
    err=db.QueryRow(raw).Scan(&bortime,&rettime)
    switch{
    case err==sql.ErrNoRows:
        return RetRec{},false
    case err==nil:
        book,_:=FindBook(isbn)
        return RetRec{id,isbn,book.Title,bortime,rettime},true
    default:
        panic(fmt.Errorf("FindRetRec:%v",err))
    }
}

//Remove a book given its isbn. Only worked when using administrator account and this book exists.
func RemoveBook(isbn string)error{
    var err error
    var ok bool
    err=ISBNValidate(isbn);if err!=nil{return err}//check whether isbn is correct.
    if _,ok=FindBook(isbn);ok==false{return fmt.Errorf("book(isbn=%s) not exist.",isbn)}
    sqls:=fmt.Sprintf("delete from books where isbn=%s;",isbn)
    SelectDatabase()
    _,err=db.Exec(sqls)
    if err!=nil{return fmt.Errorf("RemoveBook %s:%v",isbn,err)}
    return nil
}

//encrypt plaintext password to a hash value
func hashcode(x string)string{
    h:=sha256.New();h.Write([]byte(x))
    return string(hex.EncodeToString(h.Sum(nil)))
}

//add a student account given its ID(11 bit) and password. Only worked when using administrator account.
func Register(id,password string,auth int)error{
    var err error
    err=IDValidate(id);if err!=nil{return err}//check whether id is correct.
    s:=hashcode(password)//use sha256 to encrypt the password
    //fmt.Println(s,len(s))
    SelectDatabase()
    sqls:=fmt.Sprintf("insert into users(id,password,authority) values (%s,%q,%d);",id,s,auth)
    //fmt.Println(sqls)
    _,err=db.Exec(sqls)
    if err!=nil{return fmt.Errorf("Register(id:%s,psw:%s):%v",id,password,err)}
    return nil
}

func Login(id,psw string)(User,bool,error){
    var err error
    err=IDValidate(id);if err!=nil{return User{},false,err}
    s:=hashcode(psw)
    SelectDatabase()
    userpsw,auth:="",0;
    raw:=fmt.Sprintf("select password,authority from users where id=%s;",id)
    SelectDatabase()
    err=db.QueryRow(raw).Scan(&userpsw,&auth)
    switch{
    case err==sql.ErrNoRows:
        return User{},false,nil
    case err==nil:
        if userpsw!=s{
            return User{},false,nil
        }else{
            return User{id,auth},true,nil
        }
    default:
        panic(fmt.Errorf("Login:%v",err))
    }
}

func RestorePassword(id,password string)error{
    var err error
    var ok bool
    err=IDValidate(id);if err!=nil{return err}//check whether id is correct
    if _,ok=FindUser(id);ok==false{return fmt.Errorf("RestorePassword(id:%s):id not existed.",id)}
    s:=hashcode(password)
    rawsql:=`
    update users
    set password=%q
    where id=%s
    `
    sqls:=fmt.Sprintf(rawsql,s,id)
    SelectDatabase(); _,err=db.Exec(sqls);if err!=nil{return fmt.Errorf("RestorePassword(id:%s):%v",id,err)}
    return nil
}

func ChangePassword(id,oldpsw,newpsw string)(bool,error){
    var err error
    var ok bool
    _,ok,err=Login(id,oldpsw)
    if err!=nil{return false,fmt.Errorf("ChangePassword-Login(id:%s,oldpsw:%s,newpsw:%s):%v",id,oldpsw,newpsw,err)}
    if ok==false{return false,nil}
    err=RestorePassword(id,newpsw)
    if err!=nil{return false,fmt.Errorf("ChangePassword-Restore(id:%s):%v",id,err)}
    return true,nil
}

//query books by title author or ISBN, guests cannot query books.
func QueryBook(Value,Type string)([]Book,error){
    var err error
    var rows *sql.Rows
    BookList:=[]Book{}
    Error:=func()([]Book,error){return nil,fmt.Errorf("QueryBook type(%s) value(%s):%v",Type,Value,err)}
    sqls:=fmt.Sprintf("select * from books where %s=%q",Type,Value)
    //fmt.Println(sqls)
    SelectDatabase()
    rows,err=db.Query(sqls)
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

//only worked when using student account.
func BorrowBook(id,isbn string,intime time.Time)error{
    var err error
    var ok bool
    err=IDValidate(id);if err!=nil{return err}
    err=ISBNValidate(isbn);if err!=nil{return err}//check whether isbn is correct.
    if _,ok=FindBook(isbn);ok==false{return fmt.Errorf("book(isbn=%s) not exist.",isbn)}
    if _,ok=FindBorRec(id,isbn);ok==true{return fmt.Errorf("Borrow record(id:%s,isbn:%s) already exist.",id,isbn)}
    rawsql:=`
    insert into borrec(id,isbn,bortime,deadline,extendtime)values
        (%s,%s,%q,%q,0);
    `
    now:=intime
    ddl:=now.AddDate(0,0,30)
    sqls:=fmt.Sprintf(rawsql,id,isbn,now.Format(TimeFormat),ddl.Format(TimeFormat))
    //fmt.Println(sqls)
    SelectDatabase()
    _,err=db.Exec(sqls)
    if err!=nil{return fmt.Errorf("User %s borrows book %s:%v",id,isbn,err)}
    return nil
}

func BorrowQuery(id,Type string)([]Book,error){
    var err error
    var ok bool
    var rows *sql.Rows
    Error:=func()([]Book,error){return nil,fmt.Errorf("BorrowQuery(Type:%s):%v",Type,err)}
    err=IDValidate(id);if err!=nil{return nil,err}//check whether id is correct.
    var BookList []Book
    var book Book
    isbn:=""
    sqls:=fmt.Sprintf("select isbn from %s where id=%s",Type,id)
    SelectDatabase()
    rows,err=db.Query(sqls)
    if err!=nil{return Error()}
    defer rows.Close()
    for rows.Next(){
        err=rows.Scan(&isbn);if err!=nil{return Error()}
        //bookList,err=QueryBook(isbn,"isbn");if err!=nil{return Error()}
        book,ok=FindBook(isbn);if ok==false{return nil,fmt.Errorf("BorrowQuery(Type:%s,isbn:%s):cannot find such book.",Type,isbn)}
        BookList=append(BookList,book)
    }
    err=rows.Err();if err!=nil{return Error()}
    return BookList,nil
}

func GetDeadline(id,isbn string)(time.Time,error){
    var err error
    err=IDValidate(id);if err!=nil{return time.Now(),err}//check whether id is correct.
    err=ISBNValidate(isbn);if err!=nil{return time.Now(),err}//check whether isbn is correct.
    borrec,ok:=FindBorRec(id,isbn);if ok==false{return time.Now(),fmt.Errorf("GetDeadline(id:%s,isbn:%s):cannot find such BorRec.",id,isbn)}
    return borrec.Deadline,nil
}

func ExtendDeadline(id,isbn string)(bool,error){
    var err error
    var deadline time.Time
    var extendtime int
    err=IDValidate(id);if err!=nil{return false,err}//check whether id is correct.
    err=ISBNValidate(isbn);if err!=nil{return false,err}//check whether isbn is correct.
    borrec,ok:=FindBorRec(id,isbn);if ok==false{return false,fmt.Errorf("ExtendDeadline(id:%s,isbn:%s):cannot find such BorRec.",id,isbn)}
    extendtime,deadline=borrec.ExtendTime,borrec.Deadline
    if extendtime==3{return false,fmt.Errorf("the deadline of (id:%s,isbn:%s) has been extended for 3 times.",id,isbn)}
    extendtime++;deadline=deadline.AddDate(0,0,7)
    rawsql:=`
    update borrec
    set deadline=%q , extendtime=%d
    where id=%s and isbn=%s;
    `
    sqls:=fmt.Sprintf(rawsql,deadline.Format(TimeFormat),extendtime,id,isbn)
    SelectDatabase()
    _,err=db.Exec(sqls)
    if err!=nil{return false,fmt.Errorf("ExtendDeadline (id:%s,isbn:%s):%v",id,isbn,err)}
    return true,nil
}

func OverdueCheck(id string)([]BorRec,error){
    var err error
    var rows *sql.Rows
    Error:=func()([]BorRec,error){return nil,fmt.Errorf("OverdueCheck(id:%s):%v",id,err)}
    err=IDValidate(id);if err!=nil{return nil,err}//check whether id is correct.
    var BorRecList []BorRec
    sqls:=fmt.Sprintf("select isbn from borrec where id=%s and deadline < %q",id,time.Now().Format(TimeFormat))
    SelectDatabase()
    fmt.Println(sqls)
    rows,err=db.Query(sqls)
    if err!=nil{return Error()}
    defer rows.Close()
    isbn:=""
    for rows.Next(){
        err=rows.Scan(&isbn);if err!=nil{return Error()}
        borrec,_:=FindBorRec(id,isbn)
        BorRecList=append(BorRecList,borrec)
    }
    err=rows.Err();if err!=nil{return Error()}
    return BorRecList,nil
}

func ReturnBook(id,isbn string)error{
    var err error
    err=IDValidate(id);if err!=nil{return err}//check whether id is correct.
    err=ISBNValidate(isbn);if err!=nil{return err}//check whether isbn is correct.
    borrec,ok:=FindBorRec(id,isbn);if ok==false{return fmt.Errorf("ReturnBook(id:%s,isbn:%s):cannot find such BorRec.",id,isbn)}
    sqls:=fmt.Sprintf("delete from borrec where id=%s and isbn=%s;",id,isbn)
    SelectDatabase();_,err=db.Exec(sqls);if err!=nil{return fmt.Errorf("ReturnBook(id:%s,isbn:%s):%v",id,isbn,err)}
    rawsql:="insert into retrec(id,isbn,bortime,rettime) values (%s,%s,%q,%q);"
    sqls=fmt.Sprintf(rawsql,id,isbn,borrec.BorTime.Format(TimeFormat),time.Now().Format(TimeFormat))
    SelectDatabase();_,err=db.Exec(sqls);if err!=nil{return fmt.Errorf("ReturnBook(id:%s,isbn:%s):%v",id,isbn,err)}
    return nil
}

func SuspendCheck(id string)(bool,error){
    var err error
    var BorRecList []BorRec
    BorRecList,err=OverdueCheck(id);if err!=nil{return false,fmt.Errorf("SuspendCheck(id:%s):%v",id,err)}
    if len(BorRecList)>3{return true,nil}else{return false,nil}
}

func main(){
    var err error
    rawsql:=fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/?charset=utf8&loc=%s&parseTime=true",
        user,
        password,
        url.QueryEscape("Asia/Shanghai"))
    db,err=sql.Open("mysql",rawsql)
    defer db.Close()
    checkErr(err)
    if err=db.Ping();err!=nil{log.Fatal(err)}
    db.Exec("create database fudanlms;")
    defer func(){db.Exec("drop database fudanlms;")}()
    SelectDatabase()
    CreateTable()
    err=Register("10000000000","admin",Admin)
    checkErr(err)
    ShellMain()
    //adminUser:=User{"Admin","123456",Admin}
    /*
    stuUser:=User{"18307130090",Student}
    defer func(){SelectDatabase();db.Exec("delete from borrec;");db.Exec("delete from retrec;")}()
    for _,isbn:=range []string{"9787535492821","9787549550166","9787567748996","9787567748997"}{
        err=stuUser.BorrowBook(isbn,time.Now())
        //err=stuUser.BorrowBook(isbn,time.Date(2018,10,29,0,0,0,0,time.Local))
        checkErr(err)
    }
    _,err=ExtendDeadline("18307130090","9787567748996")
    checkErr(err)
    err=ReturnBook("18307130090","9787535492821")
    checkErr(err)
    fmt.Println("Done.")
    */
}
