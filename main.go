package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Nicolai.mordrup/Library/databaseconnection"

	"github.com/gorilla/mux"
)

// Struct for the server.
type server struct {
	books     map[string]Book // Map of the ISBN as key and a Book instance as value
	booksList []Book          // Slice of all the Book instance
}

// Struct for the book properties.
type Book struct {
	ISBN       string    `json:"isbn"` // The identification of the books
	Title      string    `json:"title"`
	CreateTime time.Time `json:"createTime"` // The time of creation of book instance
	UpdateTime time.Time `json:"updateTime"` // The time of update for book instance
	Author     *Author   `json:"author"`     // Embedded author struct
}

// Struct for the books Author properties.
type Author struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// handleErr for when we get an error.
// If succesfull it writes what type of error in the header we get and then
// display the error message for the user.
func handleErr(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write([]byte(message))
	if err != nil {
		log.Printf("%v, %v \n", message, err)
	}
}

// The regex patterns for the validate function
var (
	isbnPattern      = regexp.MustCompile(`^\d{13}$`)
	titlePattern     = regexp.MustCompile(`.`)
	firstNamePattern = regexp.MustCompile(`^[a-zA-Z]+(?:\s+[a-zA-Z]+)*$`)
	LastNamePattern  = regexp.MustCompile(`^[a-zA-Z]+(?:\s+[a-zA-Z]+)*$`)
)

// validate if the given input given is correct.
// if correct we return boolean true, otherwise boolean false.
func validate(b Book) error {
	var fieldErrors []string

	if matchedISBN := isbnPattern.MatchString(b.ISBN); !matchedISBN {
		fieldErrors = append(fieldErrors, " isbn ")
	}
	if matchedTitle := titlePattern.MatchString(b.Title); !matchedTitle {
		fieldErrors = append(fieldErrors, " title ")
	}
	if matchedFirstName := firstNamePattern.MatchString(b.Author.FirstName); !matchedFirstName {
		fieldErrors = append(fieldErrors, " authors firstname ")
	}
	if matchedLastName := LastNamePattern.MatchString(b.Author.LastName); !matchedLastName {
		fieldErrors = append(fieldErrors, " authors lastname ")
	}

	if len(fieldErrors) != 0 {
		return fmt.Errorf("validation failed, field error(s): %v . Fix these error before proceeding",
			strings.Join(fieldErrors, ", "))
	}
	return nil
}

// GetBooks retreives all the books that exists in the library structure.
// if succesfull, it writes the JSON encoding of the books slice to the stream
func (s *server) GetBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.booksList); err != nil {
		handleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}
}

// GetBook retreives a specific book that exists in the library structure.
// if succesfull, it writes the JSON encoding of the specific book to the stream
func (s *server) GetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) // Fetches the parameters of the http.Request URL

	val, exists := s.books[params["isbn"]]

	if !exists {
		handleErr(w, http.StatusNotFound, "The book did not exist in the library")
		return
	}
	if err := json.NewEncoder(w).Encode(val); err != nil {
		handleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}
}

// CreateBook creates a Book instance and checks that the right information have
// been passed If the information is validated then we store the information in
// our local memory and it writes the JSON encoding of the specific book to the
// stream
func (s *server) CreateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		handleErr(w, http.StatusBadRequest, "Failed to decode book")
		return
	}
	if _, exists := s.books[book.ISBN]; exists {
		handleErr(w, http.StatusConflict, "A book with this ISBN already exits")
		return
	}
	if !(book.CreateTime.IsZero() && book.UpdateTime.IsZero()) {
		handleErr(w, http.StatusForbidden, "Not allowed to change CreateTime or UpdateTime")
		return
	}
	if err := validate(book); err != nil {
		handleErr(w, http.StatusNotAcceptable, err.Error())
		return
	}

	book.CreateTime = time.Now()
	s.books[book.ISBN] = book
	s.booksList = append(s.booksList, book)

	if err := json.NewEncoder(w).Encode(book); err != nil {
		handleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}
}

// DeleteBook deletes a book instance from the library.
// if succesfull, it writes the JSON encoding of the new book slice
// without the removed book to the stream
func (s *server) DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	params := mux.Vars(r)

	if _, exists := s.books[params["isbn"]]; !exists {
		handleErr(w, http.StatusNotFound, "The book did not exist in the library or was already deleted")
		return
	}

	for index, item := range s.booksList {
		if item.ISBN != params["isbn"] {
			continue
		}
		s.booksList = append(s.booksList[:index], s.booksList[index+1:]...) // Removes the book instance from the book slice
		delete(s.books, params["isbn"])                                     // Removes the book instance from map.
		if err := json.NewEncoder(w).Encode(s.booksList); err != nil {
			handleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
			return
		}
		return
	}
}

// UpdateBook updates a book instance and checks that the right information have
// been passed If the information is validated then we store the information in
// our local memory and it writes the JSON encoding of the specific book to the
// stream
func (s *server) UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	params := mux.Vars(r)

	if _, exists := s.books[params["isbn"]]; !exists {
		handleErr(w, http.StatusNotFound, "The book did not exist in the library")
		return
	}

	for index, item := range s.booksList {
		if item.ISBN != params["isbn"] {
			continue
		}
		createdTime := s.booksList[index].CreateTime
		updatedTime := s.booksList[index].UpdateTime
		var book Book

		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			handleErr(w, http.StatusBadRequest, "Failed to decode book")
			return
		}
		if book.ISBN != params["isbn"] {
			handleErr(w, http.StatusForbidden, "Not allowed to change ISBN")
			return
		}
		if (time.Now().Unix() - updatedTime.Unix()) < 10 {
			handleErr(w, http.StatusTooEarly, "Updated a few seconds ago, please wait a moment before updating again")
			return
		}
		if err := validate(book); err != nil {
			handleErr(w, http.StatusNotAcceptable, err.Error())
			return
		}

		s.booksList = append(s.booksList[:index], s.booksList[index+1:]...) // Removes the book instance from the book slice
		delete(s.books, params["isbn"])                                     // Removes the book instance from map.

		book.CreateTime = createdTime
		book.UpdateTime = time.Now()
		s.books[book.ISBN] = book
		s.booksList = append(s.booksList, book)
		if err := json.NewEncoder(w).Encode(book); err != nil {
			handleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
			return
		}
		return
	}
}

func main() {
	// Init the Router
	r := mux.NewRouter()

	// First try to connect to database   (succeeded)
	databaseconnection.ConnectToDatabase()

	// Creates the Mock data that we play around with
	var myServer server
	myServer.booksList = append(myServer.booksList, Book{ISBN: "1233211233212",
		Title: "book_1", CreateTime: time.Now(), Author: &Author{FirstName: "nico",
			LastName: "M"}}, Book{ISBN: "1233211233234",
		Title: "book_2", CreateTime: time.Now(), Author: &Author{FirstName: "Mico",
			LastName: "N"}})

	myServer.books = make(map[string]Book)
	myServer.books["1233211233212"] = Book{ISBN: "1233211233212", Title: "book_1",
		CreateTime: time.Now(), Author: &Author{FirstName: "nico", LastName: "M"}}
	myServer.books["1233211233234"] = Book{ISBN: "1233211233234", Title: "book_2",
		CreateTime: time.Now(), Author: &Author{FirstName: "Mico", LastName: "N"}}

	// Create Route handlers /Endpoints
	r.HandleFunc("/api/books", myServer.GetBooks).Methods("GET")
	r.HandleFunc("/api/books/{isbn}", myServer.GetBook).Methods("GET")
	r.HandleFunc("/api/books", myServer.CreateBook).Methods("POST")
	r.HandleFunc("/api/books/{isbn}", myServer.UpdateBook).Methods("PUT")
	r.HandleFunc("/api/books/{isbn}", myServer.DeleteBook).Methods("DELETE")

	// Starts the server and listens on port 8000
	log.Fatal(http.ListenAndServe(":8000", r))
}
