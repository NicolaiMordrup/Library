package library

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type BookErr string

//TODO fixa så att dessa stämmer
const (
	jsonContentType = "application/json"
	ErrEncodeFail   = BookErr("Failed to Encode the book instance")
	ErrDidNotExist  = BookErr("The book did not exist in the library")
)

func (e BookErr) Error() string {
	return string(e)
}

// Struct for the Server.
type Server struct {
	router *mux.Router
	db     *sql.DB
}

//NewServer creates a new server instance
func NewServer(datab *sql.DB) *Server {
	s := &Server{}

	router := mux.NewRouter()
	router.HandleFunc("/api/books", s.GetBooks).Methods("GET")
	router.HandleFunc("/api/books/{isbn}", s.GetBook).Methods("GET")
	router.HandleFunc("/api/books/{isbn}", s.CreateBook).Methods("POST")
	router.HandleFunc("/api/books/{isbn}", s.UpdateBook).Methods("PUT")
	router.HandleFunc("/api/books/{isbn}", s.DeleteBook).Methods("DELETE")

	s.router = router
	s.db = datab
	//ConnectToDatabase(s.db) //connects to database
	return s
}

//ServeHTTP is needed to be implemented when we use the router in the struct
func (r *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// HandleErr for when we get an error.
// If succesfull it writes what type of error in the header we get and then
// display the error message for the user.
func HandleErr(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write([]byte(message))
	if err != nil {
		log.Printf("%v, %v \n", message, err)
	}
}

// GetBooks retreives all the books that exists in the library structure.
// if succesfull, it writes the JSON encoding of the books slice to the stream
func (s *Server) GetBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	book := ReadDatabaseList(s.db)

	if err := json.NewEncoder(w).Encode(book); err != nil {
		HandleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}
}

// GetBook retreives a specific book that exists in the library structure.
// if succesfull, it writes the JSON encoding of the specific book to the stream
func (s *Server) GetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) // Fetches the parameters of the http.Request URL

	book := FindSpecificBook(s.db, params["isbn"])
	if (Book{} == book) {
		HandleErr(w, http.StatusNotFound, "The book did not exist in the library")
		return
	}

	if err := json.NewEncoder(w).Encode(book); err != nil {
		HandleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}
}

// CreateBook creates a Book instance and checks that the right information have
// been passed If the information is validated then we store the information in
// our local memory and it writes the JSON encoding of the specific book to the
// stream
func (s *Server) CreateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		HandleErr(w, http.StatusBadRequest, "Failed to decode book")
		return
	}
	if exists := FindSpecificBook(s.db, book.ISBN); (exists != Book{}) {
		HandleErr(w, http.StatusConflict, "A book with this ISBN already exits")
		return
	}
	if !(book.CreateTime.IsZero() && book.UpdateTime.IsZero()) {
		HandleErr(w, http.StatusForbidden, "Not allowed to change CreateTime or UpdateTime")
		return
	}
	if err := validate(book); err != nil {
		HandleErr(w, http.StatusNotAcceptable, err.Error())
		return
	}

	book.CreateTime = time.Now()
	InsertIntoDatabase(s.db, book)
	if err := json.NewEncoder(w).Encode(book); err != nil {
		HandleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}

}

// DeleteBook deletes a book instance from the library.
// if succesfull, it writes the JSON encoding of the new book slice
// without the removed book to the stream
func (s *Server) DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	params := mux.Vars(r)

	if exists := FindSpecificBook(s.db, params["isbn"]); (exists == Book{}) {
		HandleErr(w, http.StatusNotFound, "The book did not exist in the library or was already deleted")
		return
	}

	DeleteBookFromDB(s.db, params["isbn"])
	books := ReadDatabaseList(s.db)
	if err := json.NewEncoder(w).Encode(books); err != nil {
		HandleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}
}

// UpdateBook updates a book instance and checks that the right information have
// been passed If the information is validated then we store the information in
// our local memory and it writes the JSON encoding of the specific book to the
// stream
func (s *Server) UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	params := mux.Vars(r)
	exists := FindSpecificBook(s.db, params["isbn"])
	if (exists == Book{}) {
		HandleErr(w, http.StatusNotFound, "The book did not exist in the library")
		return
	}

	createdTime := exists.CreateTime
	updatedTime := exists.UpdateTime
	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		HandleErr(w, http.StatusBadRequest, "Failed to decode book")
		return
	}
	if book.ISBN != params["isbn"] {
		HandleErr(w, http.StatusForbidden, "Not allowed to change ISBN")
		return
	}
	if (time.Now().Unix() - updatedTime.Unix()) < 10 {
		HandleErr(w, http.StatusTooEarly, "Updated a few seconds ago, please wait a moment before updating again")
		return
	}
	if err := validate(book); err != nil {
		HandleErr(w, http.StatusNotAcceptable, err.Error())
		return
	}

	book.CreateTime = createdTime
	book.UpdateTime = time.Now()
	DeleteBookFromDB(s.db, exists.ISBN)
	InsertIntoDatabase(s.db, book)

	if err := json.NewEncoder(w).Encode(book); err != nil {
		HandleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}

}
