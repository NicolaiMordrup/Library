package databaseconnection

import (
	"database/sql"
	"fmt"

	// Import sqlite driver
	_ "modernc.org/sqlite"
	//"github.com/Nicolai.mordrup/Library/"
)

var Db *sql.DB

// ConnectToDatabase connects to the database an checks if the file is already
// created or not.
//TODO fix such that the database handles errors correct.
func ConnectToDatabase(db *sql.DB) {
	Db = db
	CreateDatabase(db)

}

// DatabaseQuery Prepers a database query and executes the query on the
// database. It takes as input a query string and gives as output the rows
/*func DatabaseQuery(b Book) {
	stmt, _ := Db.Prepare("INSERT INTO library (isbn,title ,createTime, updateTime) VALUES(?,?,?,?)")
	stmt.Exec(b.ISBN, )
}*/

// ReadDatabase reads the information that we get from the database.
func ReadDatabase(db *sql.DB) {
	rows, _ := db.Query("SELECT isbn, title ,createTime, updateTime FROM library")

	var isbn string
	var title string
	var createTime string
	var updateTime string

	for rows.Next() {
		rows.Scan(&isbn, &title, &createTime, &updateTime)
		fmt.Println(isbn + ": " + title + ": " + createTime + ": " + updateTime)
	}
}

// DeleteDatabase deletes the database. Takes a database as input
func DeleteDatabase(db *sql.DB) {
	stmt, err := db.Prepare("DROP TABLE ./librarystorage.db")
	if err != nil {
		handleErr("Failed  to delete the database")
		return
	}
	stmt.Exec()
}

// CreateDatabase creates the table if it is not already created
func CreateDatabase(db *sql.DB) {
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS library(isbn INTEGER PRIMARY KEY, title TEXT, createTime TEXT, UpdateTime text)")
	if err != nil {
		handleErr("Failed to create the database ")
		return
	}
	stmt.Exec()

}

//Handles the error printing
func handleErr(errMessage string) {
	fmt.Errorf(errMessage)
}
