package main

import(
    "fmt"
    "strings"
    "bufio"
    "os"
    "time"
    "github.com/howeyc/gopass"
    "strconv"
)

func Readline()string{
    reader:=bufio.NewReader(os.Stdin)
    input,err:=reader.ReadString('\n')
    if err!=nil{fmt.Fprintln(os.Stderr,err)}
    input=strings.TrimSuffix(input,"\n")
    input=strings.TrimSuffix(input,"\r")
    return input
}

func ReadPsw(input string)string{
    fmt.Printf(input)
    pass,err:=gopass.GetPasswdMasked()
    if err!=nil{panic(err)}
    return string(pass)
}

func execLg(args []string)(User,error){
    if len(args)!=1{return User{},fmt.Errorf("Too much arguments:arguments expected 0, have %d",len(args)-1)}
    id,psw:="",""
    fmt.Print("ID:");id=Readline()
    psw=ReadPsw("Password:")
    user,ok,err:=Login(id,psw)
    if err!=nil{return User{},fmt.Errorf("execLg(id:%s,psw:%s):%v",id,psw,err)}
    if ok==false{return User{},fmt.Errorf("Wrong password or ID not exist.")}
    return user,nil
}

func (x User)execChpsw(args []string)(bool,error){
    if len(args)!=1{return false,fmt.Errorf("Too much arguments:arguments expected 0, have %d",len(args)-1)}
    oldpsw,newpsw:="",""
    //fmt.Print("Old Password:");oldpsw=Readline()
    //fmt.Print("New Password:");newpsw=Readline()
    oldpsw=ReadPsw("Old Password:")
    newpsw=ReadPsw("New Password:")
    return ChangePassword(x.ID,oldpsw,newpsw)
}

func execFdbk(args []string)([]Book,error){
    var err error
    var BookList []Book
    length:=len(args)-1
    if length!=2 {return nil,fmt.Errorf("argument number not match:arguments expected 2, have %d",length)}
    source,key,value:=args[1],"",args[2]
    switch source{
    case "-i":
        key="isbn"
    case "-a":
        key="author"
    case "-t":
        key="title"
    default:
        return nil,fmt.Errorf("Invalid arguments.")
    }
    BookList,err=QueryBook(value,key)
    if err!=nil{return nil,fmt.Errorf("execFdbk(source:%s,key:%s,value:%s):%v",source,key,value,err)}
    return BookList,nil
}

func execRg(args []string)error{
    //var err error
    length:=len(args)-1
    if length!=1{return fmt.Errorf("argument number not match:arguments expected 1, have %d",length)}
    id,psw1,psw2,auth:="","","",0
    fmt.Printf("ID:");id=Readline()
    psw1=ReadPsw("Password:")
    psw2=ReadPsw("Repeat Password:")
    if psw1!=psw2{return fmt.Errorf("Two passwords not match.")}
    switch args[1]{
    case "-a":
        auth=Admin
    case "-s":
        auth=Student
    default:
        return fmt.Errorf("Invalid arguments.")
    }
    return Register(id,psw1,auth)
}

func execAd(args []string)error{
    if len(args)!=1{return fmt.Errorf("Too much arguments:arguments expected 0, have %d",len(args)-1)}
    title,author,isbn:="","",""
    fmt.Printf("Title:");title=Readline()
    fmt.Printf("Author:");author=Readline()
    fmt.Printf("ISBN:");isbn=Readline()
    return AddBook(title,author,isbn)
}

func execRm(args []string)error{
    if len(args)!=1{return fmt.Errorf("Too much arguments:arguments expected 0, have %d",len(args)-1)}
    isbn,reason:="",""
    fmt.Printf("ISBN:");isbn=Readline()
    fmt.Printf("Reason:");reason=Readline()
    return RemoveBook(isbn,reason)
}

func (x User)execBorbk(args []string)error{
    if len(args)!=1{return fmt.Errorf("Too much arguments:arguments expected 0, have %d",len(args)-1)}
    isbn:=""
    fmt.Printf("ISBN:");isbn=Readline()
    err:=BorrowBook(x.ID,isbn,time.Now())
    if err!=nil{return err}
    borrec,_:=FindBorRec(x.ID,isbn)
    fmt.Println(borrec)
    return nil
}

func (x User)execFdrec(args []string)error{
    var err error
    length:=len(args)-1
    if length!=1{return fmt.Errorf("argument number not match:arguments expected 1, have %d",length)}
    id:=x.ID
    if x.Authority==Admin{
        fmt.Printf("ID:")
        id=Readline()
    }
    bor:=func(id string)error{
        BorRecList,err:=BorRecQuery(id);if err!=nil{return err}
        fmt.Println("Borrow Records:")
        for _,borrec:=range(BorRecList){fmt.Println("\t",borrec)}
        return nil
    }
    ret:=func(id string)error{
        RetRecList,err:=RetRecQuery(id);if err!=nil{return err}
        fmt.Println("Returned Records:")
        for _,retrec:=range(RetRecList){fmt.Println("\t",retrec)}
        return nil
    }
    switch args[1]{
    case "-b":
        err=bor(id);if err!=nil{return err}
    case "-r":
        err=ret(id);if err!=nil{return err}
    case "-a":
        err=bor(id);if err!=nil{return err}
        err=ret(id);if err!=nil{return err}
    }
    return nil
}

func (x User)execCkddl(args []string)error{
    if len(args)!=1{return fmt.Errorf("Too much arguments:arguments expected 0, have %d",len(args)-1)}
    id:=x.ID
    if x.Authority==Admin{
        fmt.Printf("ID:")
        id=Readline()
    }
    fmt.Printf("ISBN:");isbn:=Readline()
    borrec,err:=GetDeadline(id,isbn);if err!=nil{return err}
    book,_:=FindBook(isbn)
    output:=fmt.Sprintf("Title:%s\tDeadline:%s\tExtendTime:%d",book.Title,borrec.Deadline.Format(TimeFormat),borrec.ExtendTime)
    fmt.Println(output)
    return nil
}

func (x User)execCkdue(args []string)error{
    if len(args)!=1{return fmt.Errorf("Too much arguments:arguments expected 0, have %d",len(args)-1)}
    id:=x.ID
    if x.Authority==Admin{
        fmt.Printf("ID:")
        id=Readline()
    }
    BorRecList,err:=OverdueCheck(id);if err!=nil{return err}
    for _,borrec:=range(BorRecList){
        book,_:=FindBook(borrec.BookISBN)
        bortime:=borrec.BorTime.Format(TimeFormat)
        deadline:=borrec.Deadline.Format(TimeFormat)
        output:=fmt.Sprintf("Title:%s\tISBN:%s\tBorTime:%s\tDeadline:%s\tExtendTime:%d",book.Title,book.ISBN,bortime,deadline,borrec.ExtendTime)
        fmt.Println(output)
    }
    return nil
}

func (x User)execExt(args []string)error{
    if len(args)!=1{return fmt.Errorf("Too much arguments:arguments expected 0, have %d",len(args)-1)}
    id,times,isbn:=x.ID,1,""
    if x.Authority==Admin{
        fmt.Printf("ID:");id=Readline()
        fmt.Printf("ISBN:");isbn=Readline()
        fmt.Printf("Extend Weeks:");times,_=strconv.Atoi(Readline())
    }else{
        fmt.Printf("ISBN:");isbn=Readline()
    }
    borrec,err:=ExtendDeadline(id,isbn,x.Authority,times)
    if err!=nil{return err}
    book,_:=FindBook(isbn)
    output:=fmt.Sprintf("Title:%s\tDeadline:%s\tExtendTime:%d",book.Title,borrec.Deadline.Format(TimeFormat),borrec.ExtendTime)
    fmt.Println(output)
    return nil
}

func (x User)execRet(args []string)error{
    if len(args)!=1{return fmt.Errorf("Too much arguments:arguments expected 0, have %d",len(args)-1)}
    id,isbn:=x.ID,""
    if x.Authority==Admin{fmt.Printf("ID:");id=Readline()}
    fmt.Printf("ISBN:");isbn=Readline()
    err:=ReturnBook(id,isbn)
    if err!=nil{return err}
    return nil
}

func (x User)execInput(input string)(User,error){
    var err error
    var args []string
    input=strings.TrimSuffix(input,"\n")
    input=strings.TrimSuffix(input,"\r")
    rawargs:=strings.Split(input," ")
    for _,arg:=range(rawargs){
        if arg==""{continue}
        args=append(args,arg)
    }

    switch args[0]{
    case "exit"://quit the shell program
        os.Exit(0)
    case "quit"://quit the shell program
        os.Exit(0)
    case "lg"://login the account
        var newuser User
        newuser,err=execLg(args)//call the login-interacting function
        if err!=nil{return x,fmt.Errorf("execInput-lg:%v",err)}
        return newuser,nil//switch to the new account
    case "fdbk":
        var BookList []Book
        BookList,err=execFdbk(args)
        if err!=nil{return x,fmt.Errorf("execInput-fdbk:%v",err)}
        fmt.Println("Search Result:")
        for _,book:=range(BookList){fmt.Println(book)}
    case "chpsw":
        var ok bool
        if x.Authority==Guest{return x,fmt.Errorf("Please login.")}
        ok,err=x.execChpsw(args)
        if err!=nil{return x,err}
        if ok==false{return x,fmt.Errorf("Old password incorrect.")}
    case "rg":
        if x.Authority!=Admin{return x,fmt.Errorf("Only administrator account can register new account.")}
        err=execRg(args)
        if err!=nil{return x,err}
    case "ad":
        if x.Authority!=Admin{return x,fmt.Errorf("Only administrator account can register new account.")}
        err=execAd(args)
        if err!=nil{return x,err}
    case "rm":
        if x.Authority!=Admin{return x,fmt.Errorf("Only administrator account can register new account.")}
        err=execRm(args)
        if err!=nil{return x,err}
    case "borbk":
        switch x.Authority{
        case Guest:
            return x,fmt.Errorf("Please login.")
        case Suspended:
            return x,fmt.Errorf("Your account is suspended. Please return overdue books first.")
        }
        err=x.execBorbk(args)
        if err!=nil{return x,err}
    case "fdrec":
        if x.Authority==Guest{return x,fmt.Errorf("Please login.")}
        err=x.execFdrec(args)
        if err!=nil{return x,err}
    case "ckddl":
        if x.Authority==Guest{return x,fmt.Errorf("Please login.")}
        err=x.execCkddl(args)
        if err!=nil{return x,err}
    case "ckdue":
        if x.Authority==Guest{return x,fmt.Errorf("Please login.")}
        err=x.execCkdue(args)
        if err!=nil{return x,err}
    case "ext":
        switch x.Authority{
        case Guest:
            return x,fmt.Errorf("Please login.")
        case Suspended:
            return x,fmt.Errorf("Your account is suspended. Please return overdue books first.")
        }
        err=x.execExt(args)
        if err!=nil{return x,err}
    case "ret":
        if x.Authority==Guest{return x,fmt.Errorf("Please login.")}
        err=x.execRet(args)
        if err!=nil{return x,err}
    default:
        return x,fmt.Errorf("Undefined Operation.")
    }

    return x,nil
}

func (x User)HeaderPrint(){
    switch x.Authority{
    case Admin:
        fmt.Print(fmt.Sprintf("FudanLMS %s(Admin) >",x.ID))
    case Student:
        fmt.Print(fmt.Sprintf("FudanLMS %s >",x.ID))
    case Suspended:
        fmt.Print(fmt.Sprintf("FudanLMS %s(Suspended) >",x.ID))
    case Guest:
        fmt.Print("FudanLMS >")
    default:
        panic(fmt.Errorf("HeaderPrint:Invalid authority code."))
    }
}

func ShellMain(){
    var err error
    input:=""
    reader:=bufio.NewReader(os.Stdin)
    user:=User{"20000000000",Guest}
    for{
        user.HeaderPrint()
        input,err=reader.ReadString('\n')
        if err!=nil{fmt.Fprintln(os.Stderr,err)}
        user,err=user.execInput(input);if err!=nil{fmt.Println(err)}
    }
}
