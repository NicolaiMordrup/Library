package library

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	// Import sqlite driver
	_ "modernc.org/sqlite"
)

// DatabaseQuery Prepers a database query and executes the query on the
// database. It takes as input a query string and gives as output the rows
func InsertIntoDatabase(db *sql.DB, b Book) {
	stmtL, errL := db.Prepare("INSERT INTO library (isbn,title ,createTime,updateTime, publisher) VALUES(?,?,?,?,?)")
	stmtA, errA := db.Prepare("INSERT INTO author(isbn,firstName, lastName) VALUES(?,?,?)")

	if errL != nil || errA != nil {
		err := errors.New(errL.Error())
		err = fmt.Errorf("%w, %s", err, errA.Error())
		handleErr("Failed to insert into database", err)
		return
	}
	stmtA.Exec(b.ISBN, b.Author.FirstName, b.Author.LastName)
	stmtL.Exec(b.ISBN, b.Title, b.CreateTime, b.UpdateTime, b.Publisher)
}

// ReadDatabase reads the information that we get from the database.
func ReadDatabaseList(db *sql.DB) []Book {
	rows, err := db.Query("SELECT library.isbn, library.title, library.createTime,library.updateTime,author.firstName, author.lastName ,library.publisher FROM library INNER JOIN author ON library.isbn = author.isbn;")
	var b []Book
	if err != nil {
		handleErr("Failed to QUERY the statment to the database", err)
		return b
	}
	return ReadRows(rows, b)
}

//Reads from the database and find a specific book that exists.
func FindSpecificBook(db *sql.DB, isbnToFind string) Book {
	rows, err := db.Query(fmt.Sprintf("SELECT library.isbn, library.title,library.createTime,library.updateTime,author.firstName, author.lastName ,library.publisher FROM library INNER JOIN author ON library.isbn = author.isbn WHERE library.isbn=%s;", isbnToFind))
	var b []Book
	if err != nil {
		handleErr("Failed to QUERY the statment to the database", err)
		return Book{}
	}
	res := ReadRows(rows, b)
	if len(res) != 0 {
		return res[0]
	}
	return Book{}
}

//ReadRows gets the information from the query and stores it in the Book slice.
func ReadRows(rows *sql.Rows, b []Book) []Book {
	var isbndb string
	var titledb string
	var createTimedb time.Time
	var updateTimedb time.Time
	var firstNamedb string
	var lastNamedb string
	var publisherdb string

	for rows.Next() {
		rows.Scan(
			&isbndb,
			&titledb,
			&createTimedb,
			&updateTimedb,
			&firstNamedb,
			&lastNamedb,
			&publisherdb,
		)
		b = append(b, Book{ISBN: isbndb, Title: titledb, CreateTime: createTimedb,
			UpdateTime: updateTimedb, Author: &Author{FirstName: firstNamedb,
				LastName: lastNamedb}, Publisher: publisherdb})
	}
	return b
}

//Deletes a specific book from the database
func DeleteBookFromDB(db *sql.DB, isbn string) {
	for _, table := range []string{"library", "author"} {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE isbn=%s;", table, isbn))
		if err != nil {
			handleErr(fmt.Sprintf("failed to delete %s from database", isbn), err)
		}
	}
}

//Handles the error printing
func handleErr(errMessage string, err error) {
	fmt.Println(fmt.Errorf("database Error: %s, %s", errMessage, err.Error()))
}
