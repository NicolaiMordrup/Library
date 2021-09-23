package databaseconnection

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	// Import sqlite driver
	_ "modernc.org/sqlite"
)

// ConnectToDatabase connects to the database an checks if the file is already
// created or not.
//TODO fix such that the database handles errors correct.
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

	ReadDatabase(rows)

}

// DatabaseQuery Prepers a database query and executes the query on the
// database. It takes as input a query string and gives as output the rows
func DatabaseQuery() /* *sql.Rows */ {
	//TODO create the code for this
}

// ReadDatabase reads the information that we get from the database.
func ReadDatabase(rows *sql.Rows) {
	var isbn int
	var title string
	var createTime string
	var updateTime string

	for rows.Next() {
		rows.Scan(&isbn, &title, &createTime, &updateTime)
		fmt.Println(strconv.Itoa(isbn) + ": " + title + ": " + createTime + ": " + updateTime)
	}
}

// DeleteDatabase deletes the database. Takes a database as input
func DeleteDatabase(db *sql.DB) {
	//TODO create the code for this
}

// CreateDatabase creates the table if it is not already created
func CreateDatabase() {
	//TODO create the code for this
}
