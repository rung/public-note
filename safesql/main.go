package main

import (
	"github.com/google/go-safeweb/safesql"
	"os"
)

// database/sqlのwrapper
var db  *safesql.DB


func main() {
	// Normal
	// 	db.Query("SELECT ...", args...)
	//   "SELECT ..."に外からの入力が入ると危ない

	// SafeSQL (pass)
	age := 1
	db.Query(safesql.New("SELECT hoge FROM hoge WHERE age = ?"), age)

	// SafeSQL (Concat) (pass)
	s1 := safesql.New("SELECT hoge")
	s2 := safesql.New(" FROM hoge WHERE age = ?")

	db.Query(safesql.TrustedSQLStringConcat(s1, s2), age)

	// SafeSQL (fail)
	// compile error
	//  Compile時に文字列リテラルであれば、safesql.Newが使えるが、それ以外だと使えない
	s3fromExternal := safesql.New(os.Getenv("HOGE"))
	// Runtime Errorでなく、Compile Errorで防げる...!

}