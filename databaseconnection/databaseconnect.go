package databaseconnection

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	// Import sqlite driver
	_ "modernc.org/sqlite"
)

func ConnectToDatabase() {
	db, _ := sql.Open("sqlite", "./librarystorage.db")

	stmt, _ := db.Prepare("CREATE TABLE IF NOT EXISTS library(isbn INTEGER PRIMARY KEY, title TEXT, createTime TEXT, UpdateTime text)")
	//res, _ := stmt.Exec()
	stmt.Exec()
	stmt, _ = db.Prepare("INSERT INTO library (title ,createTime, updateTime) VALUES(?,?,?)")
	var createdTime string = time.Now().String()
	var updatedTime string = time.Now().String()
	stmt.Exec("book_2", createdTime, updatedTime)
	rows, _ := db.Query("SELECT isbn, title ,createTime, updateTime FROM library")
	var isbn int
	var title string
	var createTime string
	var updateTime string

	for rows.Next() {
		rows.Scan(&isbn, &title, &createTime, &updateTime)
		fmt.Println(strconv.Itoa(isbn) + ": " + title + ": " + createTime + ": " + updateTime)
	}

}
