package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Nicolai.mordrup/Library/databaseconnection"
	"github.com/gorilla/mux"
)

func main() {
	// Init the Router
	r := mux.NewRouter()
	//Db, _ := sql.Open("sqlite", "./librarystorage.db") //TODO check if a file already exists
	db, _ := sql.Open("sqlite", "./librarystorage.db") //TODO check if a file already exists

	// First try to connect to database   (succeeded)
	databaseconnection.ConnectToDatabase(db)

	// Creates the Mock data that we play around with
	var MyServer Server
	MyServer.booksList = append(MyServer.booksList, Book{ISBN: "1233211233212",
		Title: "book_1", CreateTime: time.Now().Format("2006-Jan-02 Monday 03:04:05"), Author: &Author{FirstName: "nico",
			LastName: "M"}}, Book{ISBN: "1233211233234",
		Title: "book_2", CreateTime: time.Now().Format("2006-Jan-02 Monday 03:04:05"), Author: &Author{FirstName: "Mico",
			LastName: "N"}})

	MyServer.books = make(map[string]Book)
	MyServer.books["1233211233212"] = Book{ISBN: "1233211233212", Title: "book_1",
		CreateTime: time.Now().Format("2006-Jan-02 Monday 03:04:05"), Author: &Author{FirstName: "nico", LastName: "M"}}
	MyServer.books["1233211233234"] = Book{ISBN: "1233211233234", Title: "book_2",
		CreateTime: time.Now().Format("2006-Jan-02 Monday 03:04:05"), Author: &Author{FirstName: "Mico", LastName: "N"}}

	// Create Route handlers /Endpoints
	r.HandleFunc("/api/books", MyServer.GetBooks).Methods("GET")
	r.HandleFunc("/api/books/{isbn}", MyServer.GetBook).Methods("GET")
	r.HandleFunc("/api/books", MyServer.CreateBook).Methods("POST")
	r.HandleFunc("/api/books/{isbn}", MyServer.UpdateBook).Methods("PUT")
	r.HandleFunc("/api/books/{isbn}", MyServer.DeleteBook).Methods("DELETE")

	// Starts the server and listens on port 8000
	log.Fatal(http.ListenAndServe(":8000", r))
}
