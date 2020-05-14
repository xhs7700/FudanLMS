package main

import (
	"bufio"
	"fmt"
	"github.com/howeyc/gopass"
	"os"
	"strconv"
	"strings"
	"time"
)

//a wrapped function to read one line from terminal
func Readline() string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\r")
	return input
}

//a wrapped function to mask the input password in terminal
func ReadPsw(input string) string {
	fmt.Printf(input)
	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		panic(err)
	}
	return string(pass)
}

//called by case "lg"
func execLg(args []string) (User, error) {
	if len(args) != 1 { //ensure that the input argument is only 'lg'
		return EmptyUser, fmt.Errorf("Too much arguments:arguments expected 0, have %d", len(args)-1)
	}

	//read the ID and Password
	id, psw := "", ""
	fmt.Print("ID:")
	id = Readline()
	psw = ReadPsw("Password:")

	user, ok, err := Login(id, psw) //call the Login function, searching in database to verify the identity
	if err != nil {                 //report error
		return EmptyUser, fmt.Errorf("execLg(id:%s,psw:%s):%v", id, psw, err)
	}
	if ok == false { //report that ID and Password do not match
		return EmptyUser, fmt.Errorf("Wrong password or ID not exist.")
	}
	return user, nil
}

//called by case "chpsw"
func (x User) execChpsw(args []string) (bool, error) {
	if len(args) != 1 { //ensure that the input argument is only "chpsw"
		return false, fmt.Errorf("Too much arguments:arguments expected 0, have %d", len(args)-1)
	}
	oldpsw := ReadPsw("Old Password:") //read password by calling function to ensure that input password is masked in terminal
	newpsw := ReadPsw("New Password:")
	ok, err := ChangePassword(x.ID, oldpsw, newpsw) //call the function to verify and modify data in database
	if err != nil {
		return false, err
	}
	if ok == false {
		return false, fmt.Errorf("Wrong Password.")
	}
	fmt.Println("Success.")
	return true, nil
}

//called by case "fdbk"
func execFdbk(args []string) ([]Book, error) {
	var err error
	var BookList []Book
	length := len(args) - 1
	if length != 2 { //ensure that the argument is like "fdbk -[iat] value"
		return nil, fmt.Errorf("argument number not match:arguments expected 2, have %d", length)
	}
	source, key, value := args[1], "", args[2]
	switch source { //select from different colums in database determined by different arguments
	case "-i": //select from ISBN
		key = "isbn"
	case "-a": //select from Author
		key = "author"
	case "-t": //select from Title
		key = "title"
	default:
		return nil, fmt.Errorf("Invalid arguments.")
	}
	BookList, err = QueryBook(value, key) //call the function to search in database
	if err != nil {
		return nil, fmt.Errorf("execFdbk(source:%s,key:%s,value:%s):%v", source, key, value, err)
	}
	return BookList, nil
}

//called by case "rg"
func execRg(args []string) error {
	length := len(args) - 1
	if length != 1 { //ensure that the argument is like "rg -[as]"
		return fmt.Errorf("argument number not match:arguments expected 1, have %d", length)
	}
	auth := 0

	//read ID and two passwords
	fmt.Printf("ID:")
	id := Readline()
	psw1 := ReadPsw("Password:")
	psw2 := ReadPsw("Repeat Password:")

	if psw1 != psw2 {
		return fmt.Errorf("Two passwords not match.")
	}
	switch args[1] {
	case "-a": //"-a" means the new account authority is Admin
		auth = Admin
	case "-s": //"-s" means the new account authority is Student
		auth = Student
	default:
		return fmt.Errorf("Invalid arguments.")
	}
	err := Register(id, psw1, auth) //call the function to insert data into database
	if err != nil {
		return err
	}
	fmt.Println("Success.")
	return nil
}

//called by case "ad"
func execAd(args []string) error {
	if len(args) != 1 { //ensure that the input argument is only "ad"
		return fmt.Errorf("Too much arguments:arguments expected 0, have %d", len(args)-1)
	}

	//read new book's Title, Author and ISBN
	fmt.Printf("Title:")
	title := Readline()
	fmt.Printf("Author:")
	author := Readline()
	fmt.Printf("ISBN:")
	isbn := Readline()

	err := AddBook(title, author, isbn) //call the function to insert data into database
	if err != nil {
		return err
	}
	fmt.Println("Success.")
	return nil
}

//called by case "rm"
func execRm(args []string) error {
	if len(args) != 1 { //ensure that the input argument is only "rm"
		return fmt.Errorf("Too much arguments:arguments expected 0, have %d", len(args)-1)
	}

	//read ISBN and remove reasons
	fmt.Printf("ISBN:")
	isbn := Readline()
	fmt.Printf("Reason:")
	reason := Readline()

	return RemoveBook(isbn, reason) //call the function to move data to another table in database
}

//called by case "borbk"
func (x User) execBorbk(args []string) error {
	if len(args) != 1 { //ensure that the valid input argument is "borbk"
		return fmt.Errorf("Too much arguments:arguments expected 0, have %d", len(args)-1)
	}

	//read the ISBN
	fmt.Printf("ISBN:")
	isbn := Readline()

	err := BorrowBook(x.ID, isbn, time.Now()) //call the function to insert data into borrow records table
	if err != nil {
		return err
	}
	borrec, _ := FindBorRec(x.ID, isbn) //gather the borrow record info
	fmt.Println("Success.")
	fmt.Println(borrec)
	return nil
}

//called by case "fdrec"
func (x User) execFdrec(args []string) error {
	var err error
	length := len(args) - 1
	if length != 1 { //ensure that the valid arguments is like "fdrec -[bar]"
		return fmt.Errorf("argument number not match:arguments expected 1, have %d", length)
	}
	id := x.ID                //student can only query him/herself's record
	if x.Authority == Admin { //administrator can query anyone's record
		fmt.Printf("ID:")
		id = Readline()
	}
	bor := func(id string) error { //wrapped function to query borrow records
		BorRecList, err := BorRecQuery(id) //call the function in order to select info from borrow records table
		if err != nil {
			return err
		}
		fmt.Println("Borrow Records:")
		for _, borrec := range BorRecList {
			fmt.Println("\t", borrec)
		}
		return nil
	}
	ret := func(id string) error { //wrapped function to query returned records
		RetRecList, err := RetRecQuery(id) //call the function in order to select info from returned records table
		if err != nil {
			return err
		}
		fmt.Println("Returned Records:")
		for _, retrec := range RetRecList {
			fmt.Println("\t", retrec)
		}
		return nil
	}
	switch args[1] {
	case "-b": //query borrowed and not returned records
		err = bor(id)
		if err != nil {
			return err
		}
	case "-r": //query returned records
		err = ret(id)
		if err != nil {
			return err
		}
	case "-a": //query all of them
		err = bor(id)
		if err != nil {
			return err
		}
		err = ret(id)
		if err != nil {
			return err
		}
	}
	return nil
}

//called by case "ckddl"
func (x User) execCkddl(args []string) error {
	if len(args) != 1 { //ensure that the valid argument is like "ckddl"
		return fmt.Errorf("Too much arguments:arguments expected 0, have %d", len(args)-1)
	}
	id := x.ID                //student can only query him/herself's record
	if x.Authority == Admin { //administrator can query anyone's record
		fmt.Printf("ID:")
		id = Readline()
	}

	//read ISBN
	fmt.Printf("ISBN:")
	isbn := Readline()
	borrec, err := GetDeadline(id, isbn) //call the function to select info from borrow records table
	if err != nil {
		return err
	}
	output := fmt.Sprintf("Title:%s\tDeadline:%s\tExtendTime:%d", borrec.BookTitle, borrec.Deadline.Format(TimeFormat), borrec.ExtendTime)
	fmt.Println(output)
	return nil
}

//called by case "ckdue"
func (x User) execCkdue(args []string) error {
	if len(args) != 1 { //ensure that the valid argument is like "ckdue"
		return fmt.Errorf("Too much arguments:arguments expected 0, have %d", len(args)-1)
	}
	id := x.ID                //student can only query him/herself's record
	if x.Authority == Admin { //administrator can query anyone's record
		fmt.Printf("ID:")
		id = Readline()
	}
	BorRecList, err := OverdueCheck(id) //call this function to select info from borrow records table
	if err != nil {
		return err
	}
	fmt.Println("Overdue Records:")
	for _, borrec := range BorRecList {
		bortime := borrec.BorTime.Format(TimeFormat)
		deadline := borrec.Deadline.Format(TimeFormat)
		output := fmt.Sprintf("Title:%s\tISBN:%s\tBorTime:%s\tDeadline:%s\tExtendTime:%d", borrec.BookTitle, borrec.BookISBN, bortime, deadline, borrec.ExtendTime)
		fmt.Println(output)
	}
	return nil
}

//called by case "ext"
func (x User) execExt(args []string) error {
	if len(args) != 1 { //ensure that the valid argument is like "ext"
		return fmt.Errorf("Too much arguments:arguments expected 0, have %d", len(args)-1)
	}
	id, times, isbn := x.ID, 1, "" //student can only extend him/herself's record three times, each time a week
	if x.Authority == Admin {      //administrator can extend anyone's record, and can extend any times,each time any weeks
		fmt.Printf("ID:")
		id = Readline()
		fmt.Printf("ISBN:")
		isbn = Readline()
		fmt.Printf("Extend Weeks:")
		times, _ = strconv.Atoi(Readline())
	} else {
		fmt.Printf("ISBN:")
		isbn = Readline()
	}
	borrec, err := ExtendDeadline(id, isbn, x.Authority, times) //call this function to update data in borrow records table
	if err != nil {
		return err
	}
	output := fmt.Sprintf("Title:%s\tDeadline:%s\tExtendTime:%d", borrec.BookTitle, borrec.Deadline.Format(TimeFormat), borrec.ExtendTime)
	fmt.Println("Success.")
	fmt.Println(output)
	return nil
}

//called by case "ret"
func (x User) execRet(args []string) error {
	if len(args) != 1 { //ensure that the valid argument is like "ret"
		return fmt.Errorf("Too much arguments:arguments expected 0, have %d", len(args)-1)
	}
	id, isbn := x.ID, ""      //student can only return him/herself's borrowed books
	if x.Authority == Admin { //administrator can return anyone's borrowed books
		fmt.Printf("ID:")
		id = Readline()
	}

	//read ISBN
	fmt.Printf("ISBN:")
	isbn = Readline()
	err := ReturnBook(id, isbn) //call this function to move data from borrow records table to returned records table
	if err != nil {
		return err
	}
	fmt.Println("Success.")
	return nil
}

func execRes(args []string) error {
	if len(args) != 1 { //ensure that the valid argument is like "ret"
		return fmt.Errorf("Too much arguments:arguments expected 0, have %d", len(args)-1)
	}

	//read ID and new password
	fmt.Printf("ID:")
	id := Readline()
	psw := ReadPsw("New Password:")

	err := ResetPassword(id, psw) //call the function to update data in users table
	if err != nil {
		return err
	}
	fmt.Println("Success.")
	return nil
}

func (x User) execInput(input string) (User, error) {
	var err error
	var args []string
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\r")
	rawargs := strings.Split(input, " ")
	for _, arg := range rawargs {
		if arg == "" {
			continue
		}
		args = append(args, arg)
	}
	if args == nil {
		return x, nil
	}
	switch args[0] {
	case "exit": //quit the shell program
		os.Exit(0)
	case "quit": //quit the shell program
		os.Exit(0)
	case "lg": //login the account
		var newuser User
		newuser, err = execLg(args) //call the login-execute function
		if err != nil {
			return x, fmt.Errorf("execInput-lg:%v", err)
		}
		return newuser, nil //switch to the new account
	case "fdbk": //search for books by their authors, titles or ISBN.
		var BookList []Book            //store the selected books
		BookList, err = execFdbk(args) //call the findBook-execute function
		if err != nil {
			return x, fmt.Errorf("execInput-fdbk:%v", err)
		}
		fmt.Println("Search Result:")
		for _, book := range BookList {
			fmt.Println(book)
		}
	case "chpsw": //change current user's password
		var ok bool
		if x.Authority == Guest { //guests do not have accounts
			return x, fmt.Errorf("Please login.")
		}
		ok, err = x.execChpsw(args) //call the changePassword-execute function
		if err != nil {
			return x, err
		}
		if ok == false {
			return x, fmt.Errorf("Old password incorrect.")
		}
	case "rg": //register new account
		if x.Authority != Admin { //only permitted by administrator account
			return x, fmt.Errorf("Only administrator account can register new account.")
		}
		err = execRg(args) //call register-execute function
		if err != nil {
			return x, err
		}
	case "ad": //add new books
		if x.Authority != Admin { //only permitted by administrator account
			return x, fmt.Errorf("Only administrator account can register new account.")
		}
		err = execAd(args) //call addBook-execute function
		if err != nil {
			return x, err
		}
	case "rm": //remove books with reasons
		if x.Authority != Admin { //only permitted by administrator account
			return x, fmt.Errorf("Only administrator account can register new account.")
		}
		err = execRm(args) //call the removeBook-execute function
		if err != nil {
			return x, err
		}
	case "borbk": //borrow one book
		switch x.Authority {
		case Guest: //guests do not have accounts
			return x, fmt.Errorf("Please login.")
		case Suspended: //suspended students cannot borrow books
			return x, fmt.Errorf("Your account is suspended. Please return overdue books first.")
		}
		err = x.execBorbk(args) //call borrowBook-execute function
		if err != nil {
			return x, err
		}
	case "fdrec": //query borrow/returned records
		if x.Authority == Guest { //guests do not have account
			return x, fmt.Errorf("Please login.")
		}
		err = x.execFdrec(args) //call findRecords-execute function
		if err != nil {
			return x, err
		}
	case "ckddl": //query one borrowed book's deadline
		if x.Authority == Guest { //guests do not have accounts
			return x, fmt.Errorf("Please login.")
		}
		err = x.execCkddl(args) //call the checkDeadline-execute function
		if err != nil {
			return x, err
		}
	case "ckdue": //check whether one user has overdue books
		if x.Authority == Guest { //guests do not have account
			return x, fmt.Errorf("Please login.")
		}
		err = x.execCkdue(args) //call the checkOverdue-execute function
		if err != nil {
			return x, err
		}
	case "ext": //extend one borrow record's deadline
		switch x.Authority {
		case Guest: //guests do not have account
			return x, fmt.Errorf("Please login.")
		case Suspended: //suspend students cannot extend Deadline
			return x, fmt.Errorf("Your account is suspended. Please return overdue books first.")
		}
		err = x.execExt(args) //call the extendDeadline-execute function
		if err != nil {
			return x, err
		}
	case "ret": //return one borrowed book
		if x.Authority == Guest { //guests do not have account
			return x, fmt.Errorf("Please login.")
		}
		err = x.execRet(args) //call the returnBook-execute function
		if err != nil {
			return x, err
		}
	case "res": //reset user's password
		if x.Authority != Admin { //only permitted by administrator account
			return x, fmt.Errorf("Only administrator account can force set password.")
		}
		err = execRes(args) //call the resetPassword-execute function
		if err != nil {
			return x, err
		}
	case "help":
		fmt.Println(HelpText)
	default:
		return x, fmt.Errorf("Undefined Operation.")
	}

	return x, nil
}

func (x User) HeaderPrint() {
	switch x.Authority {
	case Admin:
		fmt.Print(fmt.Sprintf("FudanLMS %s(Admin) >", x.ID))
	case Student:
		fmt.Print(fmt.Sprintf("FudanLMS %s >", x.ID))
	case Suspended:
		fmt.Print(fmt.Sprintf("FudanLMS %s(Suspended) >", x.ID))
	case Guest:
		fmt.Print("FudanLMS >")
	default:
		panic(fmt.Errorf("HeaderPrint:Invalid authority code."))
	}
}

func ShellMain() {
	var err error
	fmt.Println(WelcomeText)
	input := ""
	reader := bufio.NewReader(os.Stdin)
	user := User{"20000000000", Guest}
	for {
		user.HeaderPrint() //print the header of shell
		input, err = reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		user, err = user.execInput(input)
		if err != nil {
			fmt.Println(err)
		}
		user, err = user.SuspendCheck()
		if err != nil {
			fmt.Println(err)
		}
	}
}
