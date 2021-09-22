package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func ConnectToDatabase() {
	db, _ := sql.Open("sqlite", "/.librarystorage.db")

	stmt, _ := db.Prepare("INSERT INTO userinfo(username, departname, created) values(?,?,?)")
	res, _ := stmt.Exec("Nicolai", "securitas", "2021-09-21")

	id, _ := res.LastInsertId()
	fmt.Println(id)
}
