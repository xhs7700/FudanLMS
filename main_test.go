package main

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

func TestAddBook(t *testing.T) {
	var tests = []struct {
		input Book
		want  error
	}{
		{Book{"1984", "George Orwell", "9787567748996"}, fmt.Errorf("notnil")},
		{Book{"1984", "George Orwell", "9786567748996"}, nil},
		{Book{"孤独深处", "郝景芳", "test"}, fmt.Errorf("notnil")},
	}
	for _, test := range tests {
		book, want := test.input, test.want
		got := AddBook(book.Title, book.Author, book.ISBN)
		if (got == nil) != (want == nil) {
			t.Errorf("AddBook(%s)=%v,want=%v", book, got, want)
		}
	}
}

func TestRemoveBook(t *testing.T) {
	var tests = []struct {
		isbn, reason string
		want         error
	}{
		{"0000298472347", "reason", nil},
		{"9786208060625", "reason", fmt.Errorf("notnil")},
		{"9786208060625", "reason", fmt.Errorf("notnil")},
		{"978720806062", "", fmt.Errorf("notnil")},
	}
	for _, test := range tests {
		isbn, reason, want := test.isbn, test.reason, test.want
		got := RemoveBook(isbn, reason)
		if (got == nil) != (want == nil) {
			t.Errorf("RemoveBook(isbn:%s,reason:%s)=%v,want=%v", isbn, reason, got, want)
		}
	}
}

func TestRegister(t *testing.T) {
	var tests = []struct {
		id, password string
		auth         int
		want         error
	}{
		{"18307130091", "123456", 0, nil},
		{"18307130092", "", 1, fmt.Errorf("notnil")},
		{"1830713009", "123456", 2, fmt.Errorf("notnil")},
		{"18307130090", "123456", 0, fmt.Errorf("notnil")},
	}
	for _, test := range tests {
		id, password, want, auth := test.id, test.password, test.want, test.auth
		got := Register(id, password, auth)
		if (got == nil) != (want == nil) {
			t.Errorf("Register(id:%s,password:%s,auth=%d)=%v,want=%v", id, password, auth, got, want)
		}
	}
}

func TestLogin(t *testing.T) {
	var tests = []struct {
		id, psw string
		user    User
		ok      bool
		err     error
	}{
		{"10000000000", "admin", User{"10000000000", Admin}, true, nil},
		{"1830713009", "123456", EmptyUser, false, fmt.Errorf("ID length incorrect.")},
		{"17307130090", "123456", EmptyUser, false, nil},
		{"10000000000", "admi", EmptyUser, false, nil},
	}
	for _, test := range tests {
		id, psw, user, ok, err := test.id, test.psw, test.user, test.ok, test.err
		gotuser, gotok, goterr := Login(id, psw)
		if gotuser != user || gotok != ok || (goterr == nil) != (err == nil) {
			t.Errorf("Login(id:%s,psw:%s)=(%v,%v,%v),want=(%v,%v,%v)", id, psw, gotuser, gotok, goterr, user, ok, err)
		}
	}
}

func TestResetPassword(t *testing.T) {
	var tests = []struct {
		id, psw string
		err     error
	}{
		{"18307130092", "123456", fmt.Errorf("notnil")},
		{"1830713009", "123456", fmt.Errorf("notnil")},
		{"18307120090", "12345", nil},
	}
	for _, test := range tests {
		id, psw, err := test.id, test.psw, test.err
		got := ResetPassword(id, psw)
		if (got == nil) != (err == nil) {
			t.Errorf("ResetPassword(id:%s,psw:%s)=%v,want=%v", id, psw, got, err)
		}
	}
}

func TestChangePassword(t *testing.T) {
	var tests = []struct {
		id, oldpsw, newpsw string
		ok                 bool
		err                error
	}{
		{"18307110089", "123456", "1234567", false, nil},
		{"1830711008", "123456", "1234567", false, fmt.Errorf("notnil")},
		{"18307110090", "12345", "1234567", false, nil},
		{"18307110090", "123456", "1234567", true, nil},
	}
	for _, test := range tests {
		id, oldpsw, newpsw, ok, err := test.id, test.oldpsw, test.newpsw, test.ok, test.err
		gotok, goterr := ChangePassword(id, oldpsw, newpsw)
		if gotok != ok || (goterr == nil) != (err == nil) {
			t.Errorf("ChangePassword(id:%s,oldpsw:%s,newpsw:%s)=%v,%v;want=%v,%v", id, oldpsw, newpsw, gotok, goterr, ok, err)
		}
	}
}

func TestQueryBook(t *testing.T) {
	var tests = []struct {
		value, Type string
		books       []Book
		err         error
	}{
		{
			"追风筝的人", "title",
			[]Book{
				{"追风筝的人", "卡勒德·胡赛尼", "9787208061644"},
				{"追风筝的人", "卡勒德.胡赛尼", "9787208060625"},
			},
			nil,
		},
		{"abcd", "isbn", nil, fmt.Errorf("notnil")},
		{
			"George_Orwell", "author",
			[]Book{
				{"1984", "George_Orwell", "9787567748996"},
				{"1984", "George_Orwell", "9787567748997"},
				{"Animal_Farm", "George_Orwell", "9787567743908"},
			},
			nil,
		},
	}
	for _, test := range tests {
		value, Type, books, err := test.value, test.Type, test.books, test.err
		gotbooks, goterr := QueryBook(value, Type)
		//fmt.Println(gotbooks)
		//fmt.Println(books)
		var isequal bool
		if len(gotbooks) != len(books) {
			isequal = false
		} else {
			sort.Slice(books, func(i, j int) bool {
				if books[i].ISBN < books[j].ISBN {
					return true
				}
				return false
			})
			sort.Slice(gotbooks, func(i, j int) bool {
				if gotbooks[i].ISBN < gotbooks[j].ISBN {
					return true
				}
				return false
			})
			isequal = true
			for i := range books {
				if books[i] != gotbooks[i] {
					isequal = false
					break
				}
			}
		}
		if isequal == false || (goterr == nil) != (err == nil) {
			t.Errorf("QueryBook(Value:%s,Type:%s)=(%v,%v);want=(%v,%v)", value, Type, gotbooks, goterr, books, err)
		}
	}
}

func TestBorrowBook(t *testing.T) {
	var tests = []struct {
		id, isbn string
		intime   time.Time
		err      error
	}{
		{"18307130090", "9787567748997", time.Now(), fmt.Errorf("notnil")},
		{"18307130090", "9787567748996", time.Now(), nil},
		{"1830713009", "9787567748997", time.Now(), fmt.Errorf("notnil")},
		{"18307130090", "test", time.Now(), fmt.Errorf("notnil")},
		{"18307130090", "0001567748997", time.Now(), fmt.Errorf("notnil")},
	}
	for _, test := range tests {
		id, isbn, intime, err := test.id, test.isbn, test.intime, test.err
		goterr := BorrowBook(id, isbn, intime)
		if (goterr == nil) != (err == nil) {
			strtime := intime.Format(TimeFormat)
			t.Errorf("BorrowBook(id:%s,isbn:%s,intime:%s)=%v;want=%v", id, isbn, strtime, goterr, err)
		}
		if err == nil {
			err2 := ReturnBook(id, isbn)
			checkErr(err2)
		}
	}
}

func TestBorRecQuery(t *testing.T) {
	var tests = []struct {
		id      string
		borrecs []BorRec
		err     error
	}{
		//{"1830713009",nil,fmt.Errorf("notnil")},
		//{"30000000000",nil,fmt.Errorf("notnil")},
		{
			"18307130090",
			[]BorRec{
				{"18307130090", "9787567748997", "1984", time.Date(2020, 5, 12, 17, 30, 0, 0, time.Local), time.Date(2020, 6, 11, 17, 30, 0, 0, time.Local), 0},
				{"18307130090", "9787208061644", "追风筝的人", time.Date(2020, 5, 11, 16, 18, 37, 0, time.Local), time.Date(2020, 7, 1, 16, 18, 37, 0, time.Local), 3},
			},
			nil,
		},
	}
	for _, test := range tests {
		id, borrecs, err := test.id, test.borrecs, test.err
		gotborrecs, goterr := BorRecQuery(id)
		var isequal bool
		if len(gotborrecs) != len(borrecs) {
			isequal = false
		} else {
			sort.Slice(borrecs, func(i, j int) bool {
				if borrecs[i].BookISBN < borrecs[j].BookISBN {
					return true
				}
				return false
			})
			sort.Slice(gotborrecs, func(i, j int) bool {
				if gotborrecs[i].BookISBN < gotborrecs[j].BookISBN {
					return true
				}
				return false
			})
			isequal = true
			for i := range borrecs {
				if borrecs[i].UserID != gotborrecs[i].UserID || borrecs[i].BookISBN != gotborrecs[i].BookISBN {
					isequal = false
					break
				}
			}
		}
		//fmt.Println(isequal)
		//fmt.Println(borrecs)
		//fmt.Println(gotborrecs)
		if isequal == false || (goterr == nil) != (err == nil) {
			t.Errorf("BorRecQuery(id:%s)=(%v,%v);want=(%v,%v)", id, gotborrecs, goterr, borrecs, err)
		}
	}
}

func TestGetDeadline(t *testing.T) {
	var tests = []struct {
		id, isbn string
		borrec   BorRec
		err      error
	}{
		{"1830713009", "9787567748997", BorRec{}, fmt.Errorf("notnil")},
		{"18307130090", "test", BorRec{}, fmt.Errorf("notnil")},
		{"10000000000", "9787535492821", BorRec{}, fmt.Errorf("notnil")},
		{
			"18307130012", "9787567743908",
			BorRec{"18307130012", "9787567743908", "Animal_Farm", time.Date(2020, 1, 22, 10, 0, 0, 0, time.Local), time.Date(2020, 4, 8, 0, 0, 0, 0, time.Local), 2},
			nil,
		},
	}
	for _, test := range tests {
		id, isbn, borrec, err := test.id, test.isbn, test.borrec, test.err
		gotborrec, goterr := GetDeadline(id, isbn)
		if borrec.UserID != gotborrec.UserID || borrec.BookISBN != gotborrec.BookISBN || (goterr == nil) != (err == nil) {
			t.Errorf("GetDeadline(id:%s,isbn:%s)=(%v,%v);want=(%v,%v)", id, isbn, gotborrec, goterr, borrec, err)
		}
	}
}

func TestExtendDeadline(t *testing.T) {
	var tests = []struct {
		id, isbn    string
		auth, weeks int
		borrec      BorRec
		err         error
	}{
		{"1830713009", "9787567748996", 1, 1, BorRec{}, fmt.Errorf("notnil")},
		{"18307130090", "test", 1, 1, BorRec{}, fmt.Errorf("notnil")},
		{"18300130090", "9787567748997", 1, 1, BorRec{}, fmt.Errorf("notnil")},
		{"18307130090", "0002567748997", 1, 1, BorRec{}, fmt.Errorf("notnil")},
		{"18307130090", "9787208061644", 1, 1, BorRec{}, fmt.Errorf("notnil")},
		{
			"10000000000", "9787567748996", 0, 1,
			BorRec{"10000000000", "9787567748996", "1984", time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2038, 1, 26, 3, 14, 8, 0, time.Local), 5},
			nil,
		},
	}
	for _, test := range tests {
		id, isbn, auth, weeks, borrec, err := test.id, test.isbn, test.auth, test.weeks, test.borrec, test.err
		gotborrec, goterr := ExtendDeadline(id, isbn, auth, weeks)
		//fmt.Println(gotborrec)
		//fmt.Println(borrec)
		if gotborrec.UserID != borrec.UserID || gotborrec.BookISBN != borrec.BookISBN || (goterr == nil) != (err == nil) {
			t.Errorf("ExtendDeadline(id:%s,isbn:%s,auth:%d,weeks:%d)=(%v,%v);want=(%v,%v)", id, isbn, auth, weeks, gotborrec, goterr, borrec, err)
		}
	}
}

func TestOverdueCheck(t *testing.T) {
	var tests = []struct {
		id      string
		borrecs []BorRec
		err     error
	}{
		{"1830713009", nil, fmt.Errorf("notnil")},
		{
			"18307130012",
			[]BorRec{
				{"18307130012", "9787535492821", "DaoYuDuBai", time.Date(2019, 12, 31, 23, 59, 59, 0, time.Local), time.Date(2020, 1, 30, 23, 59, 59, 0, time.Local), 0},
				{"18307130012", "9787549550166", "YeHuoJi", time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2019, 2, 14, 0, 0, 0, 0, time.Local), 2},
				{"18307130012", "9787567748996", "1984", time.Date(2018, 6, 12, 18, 0, 0, 0, time.Local), time.Date(2018, 7, 12, 18, 0, 0, 0, time.Local), 0},
				{"18307130012", "9787567743908", "Animal_Farm", time.Date(2020, 1, 22, 10, 0, 0, 0, time.Local), time.Date(2020, 4, 8, 0, 0, 0, 0, time.Local), 2},
			},
			nil,
		},
		{"18307130090", nil, nil},
	}
	for _, test := range tests {
		id, borrecs, err := test.id, test.borrecs, test.err
		gotborrecs, goterr := OverdueCheck(id)
		var isequal bool
		if gotborrecs == nil && borrecs == nil {
			isequal = true
		} else if len(gotborrecs) != len(borrecs) {
			isequal = false
		} else {
			sort.Slice(borrecs, func(i, j int) bool {
				if borrecs[i].BookISBN < borrecs[j].BookISBN {
					return true
				}
				return false
			})
			sort.Slice(gotborrecs, func(i, j int) bool {
				if gotborrecs[i].BookISBN < gotborrecs[j].BookISBN {
					return true
				}
				return false
			})
			isequal = true
			for i := range borrecs {
				if borrecs[i].UserID != gotborrecs[i].UserID || borrecs[i].BookISBN != gotborrecs[i].BookISBN {
					isequal = false
					break
				}
			}
		}
		if isequal == false || (goterr == nil) != (err == nil) {
			t.Errorf("OverdueCheck(id:%s)=(%v,%v);want=(%v,%v)", id, gotborrecs, goterr, borrecs, err)
		}
	}
}

func TestReturnBook(t *testing.T) {
	var tests = []struct {
		id, isbn string
		err      error
	}{
		{"1830713009", "9787208060625", fmt.Errorf("notnil")},
		{"18307130090", "test", fmt.Errorf("notnil")},
		{"18307130090", "9787567743908", fmt.Errorf("notnil")},
		{"18307130090", "9787208061644", nil},
	}
	for _, test := range tests {
		id, isbn, err := test.id, test.isbn, test.err
		goterr := ReturnBook(id, isbn)
		if (goterr == nil) != (err == nil) {
			t.Errorf("ReturnBook(id:%s,isbn:%s)=%v;want=%v", id, isbn, goterr, err)
		}
	}
}

func TestSuspendCheck(t *testing.T) {
	var tests = []struct {
		userin, userout User
		err             error
	}{
		{
			User{"10000000000", Admin},
			User{"10000000000", Admin},
			nil,
		},
		{
			User{"1830713009", Student},
			User{"1830713009", Student},
			fmt.Errorf("notnil"),
		},
		{
			User{"18307130090", Student},
			User{"18307130090", Student},
			nil,
		},
		{
			User{"18307130090", Suspended},
			User{"18307130090", Student},
			nil,
		},
		{
			User{"18307130012", Suspended},
			User{"18307130012", Suspended},
			nil,
		},
		{
			User{"18307130012", Student},
			User{"18307130012", Suspended},
			nil,
		},
	}
	for _, test := range tests {
		userin, userout, err := test.userin, test.userout, test.err
		gotuser, goterr := userin.SuspendCheck()
		if gotuser != userout || (goterr == nil) != (err == nil) {
			t.Errorf("(%v).SuspendCheck()=(%v,%v);want=(%v,%v)", userin, gotuser, goterr, userout, err)
		}
	}
}
