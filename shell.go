package main

import(
    "fmt"
    "strings"
    "bufio"
    "os"
    "github.com/howeyc/gopass"
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
    //var err error
    if len(args)!=1{return User{},fmt.Errorf("Too much arguments:arguments expected 0, have %d",len(args)-1)}
    id,psw:="",""
    fmt.Print("ID:")
    id=Readline()
    //fmt.Print("Password:");psw=Readline()
    psw=ReadPsw("Password:")
    user,ok,err:=Login(id,psw)
    if err!=nil{return User{},fmt.Errorf("execLg(id:%s,psw:%s):%v",id,psw,err)}
    if ok==false{return User{},fmt.Errorf("execLg(id:%s,psw:%s):Wrong password or ID not exist.",id,psw)}
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
    case "exit":
        os.Exit(0)
    case "lg":
        var newuser User
        newuser,err=execLg(args)
        if err!=nil{return x,fmt.Errorf("execInput-lg:%v",err)}
        return newuser,nil
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
    default:
        return x,fmt.Errorf("Undefined Operation.")
    }

    return x,nil
}

func (x User)HeaderPrint(){
    switch x.Authority{
    case Admin:
        fmt.Print("FudanLMS %s(Admin) >",x.ID)
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
