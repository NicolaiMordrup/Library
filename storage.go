package library

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	// sqlite database library
	_ "modernc.org/sqlite"
)

type DBStorage struct {
	db  *sql.DB
	log *zap.SugaredLogger
}

// DatabaseQuery Prepers a database query and executes the query on the
// database. It takes as input a query string and gives as output the rows
func (storage *DBStorage) InsertIntoDatabase(b Book) error {
	stmtL, errL := storage.db.Prepare("INSERT INTO library (isbn,title ,createTime,updateTime, publisher) VALUES(?,?,?,?,?)")
	stmtA, errA := storage.db.Prepare("INSERT INTO author(isbn,firstName, lastName) VALUES(?,?,?)")

	if errL != nil || errA != nil {
		err := errors.New(errL.Error())
		err = fmt.Errorf("%w, %s", err, errA.Error())
		storage.handleErr("Failed to prepare statement", err)
		return err
	}
	_, errA = stmtA.Exec(b.ISBN, b.Author.FirstName, b.Author.LastName)
	_, errL = stmtL.Exec(b.ISBN, b.Title, b.CreateTime, b.UpdateTime, b.Publisher)

	if errL != nil || errA != nil {
		err := errors.New(errL.Error())
		err = fmt.Errorf("%w, %s", err, errA.Error())
		storage.handleErr("Failed to insert into database", err)
		return err
	}

	return nil
}

// ReadDatabase reads the information that we get from the database.
func (storage *DBStorage) ReadDatabaseList() []Book {
	rows, err := storage.db.Query("SELECT library.isbn, library.title, library.createTime,library.updateTime,author.firstName, author.lastName ,library.publisher FROM library INNER JOIN author ON library.isbn = author.isbn;")
	var b []Book
	if err != nil {
		storage.handleErr("Failed to QUERY the statement to the database", err)
		return b
	}
	return storage.ReadRows(rows, b)
}

// Reads from the database and find a specific book that exists.
func (storage *DBStorage) FindSpecificBook(isbnToFind string) Book {
	rows, err := storage.db.Query(fmt.Sprintf("SELECT library.isbn, library.title,library.createTime,library.updateTime,author.firstName, author.lastName ,library.publisher FROM library INNER JOIN author ON library.isbn = author.isbn WHERE library.isbn=%s;", isbnToFind))
	var b []Book
	if err != nil {
		storage.handleErr("Failed to QUERY the statement to the database", err)
		return Book{}
	}
	res := storage.ReadRows(rows, b)
	if len(res) != 0 {
		return res[0]
	}
	return Book{}
}

// ReadRows gets the information from the query and stores it in the Book slice.
func (storage *DBStorage) ReadRows(rows *sql.Rows, b []Book) []Book {
	var isbndb string
	var titledb string
	var createTimedb time.Time
	var updateTimedb time.Time
	var firstNamedb string
	var lastNamedb string
	var publisherdb string

	for rows.Next() {
		err := rows.Scan(
			&isbndb,
			&titledb,
			&createTimedb,
			&updateTimedb,
			&firstNamedb,
			&lastNamedb,
			&publisherdb,
		)
		if err != nil {
			storage.handleErr("Failed to QUERY the statement to the database", err)
			return []Book{}
		}

		b = append(b, Book{ISBN: isbndb, Title: titledb, CreateTime: createTimedb,
			UpdateTime: updateTimedb, Author: Author{FirstName: firstNamedb,
				LastName: lastNamedb}, Publisher: publisherdb})
	}
	return b
}

// Deletes a specific book from the database
func (storage *DBStorage) DeleteBookFromDB(isbn string) error {
	for _, table := range []string{"library", "author"} {
		res, err := storage.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE isbn=%s;",
			table, isbn))
		if err != nil {
			storage.handleErr(fmt.Sprintf("failed to delete %s from database",
				isbn), err)
		}
		if rows, _ := res.RowsAffected(); rows == 0 {
			return errors.New("book did not exist in the library database")
		}
	}
	return nil
}

// Handles the error printing
func (storage *DBStorage) handleErr(errMessage string, err error) {
	storage.log.Infow("starting server", "Error",
		fmt.Errorf("database Error: %s, %v", errMessage, err.Error()))
}
