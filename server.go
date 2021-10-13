package library

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const (
	jsonContentType = "application/json"
)

// Server contains the server stuff.
type Server struct {
	router                    *mux.Router
	store                     DBStorage
	minDurationBetweenUpdates time.Duration
	log                       *zap.SugaredLogger
}

// NewServer creates a new server instance.
func NewServer(
	dataBase *sql.DB,
	logger *zap.SugaredLogger,
	minDurationTimeBetweenUpdates time.Duration,
) *Server {
	s := &Server{}

	router := mux.NewRouter()
	router.HandleFunc("/api/books", s.ListBooks).Methods("GET")
	router.HandleFunc("/api/books/{isbn}", s.GetBook).Methods("GET")
	router.HandleFunc("/api/books/{isbn}", s.CreateBook).Methods("POST")
	router.HandleFunc("/api/books/{isbn}", s.UpdateBook).Methods("PUT")
	router.HandleFunc("/api/books/{isbn}", s.DeleteBook).Methods("DELETE")

	s.router = router
	s.store = DBStorage{db: dataBase, log: logger}
	s.log = logger
	s.minDurationBetweenUpdates = minDurationTimeBetweenUpdates
	return s
}

// ServeHTTP is needed to be implemented when we use the router in the struct.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

// HandleErr for when we get an error.
// If succesfull it writes what type of error in the header we get and then
// display the error message for the user.
func (s *Server) HandleErr(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write([]byte(message))
	if err != nil {
		s.log.Infow("Error", fmt.Sprintf("%v, %v ", message, err))
	}
}

// GetBooks retreives all the books that exists in the library structure.
// if succesfull, it writes the JSON encoding of the books slice to the stream
func (s *Server) ListBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	book := s.store.ReadDatabaseList()

	if err := json.NewEncoder(w).Encode(book); err != nil {
		s.HandleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}
}

// GetBook retreives a specific book that exists in the library structure.
// if succesfull, it writes the JSON encoding of the specific book to the stream
func (s *Server) GetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) // Fetches the parameters of the http.Request URL

	book := s.store.FindSpecificBook(params["isbn"])
	if (Book{} == book) {
		s.HandleErr(w, http.StatusNotFound, "The book did not exist in the library")
		return
	}

	if err := json.NewEncoder(w).Encode(book); err != nil {
		s.HandleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
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
		s.HandleErr(w, http.StatusBadRequest, "Failed to decode book")
		return
	}
	if exists := s.store.FindSpecificBook(book.ISBN); (exists != Book{}) {
		s.HandleErr(w, http.StatusConflict, "A book with this ISBN already exits")
		return
	}
	if !(book.CreateTime.IsZero() && book.UpdateTime.IsZero()) {
		s.HandleErr(w, http.StatusForbidden, "Not allowed to change CreateTime "+
			"or UpdateTime")
		return
	}
	if err := validate(book); err != nil {
		s.HandleErr(w, http.StatusNotAcceptable, err.Error())
		return
	}

	book.CreateTime = time.Now()
	book.UpdateTime = time.Now()
	s.store.InsertIntoDatabase(book)
	if err := json.NewEncoder(w).Encode(book); err != nil {
		s.HandleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}
}

// DeleteBook deletes a book instance from the library.
// if succesfull, it writes the JSON encoding of the new book slice
// without the removed book to the stream
func (s *Server) DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	params := mux.Vars(r)

	if exists := s.store.FindSpecificBook(params["isbn"]); (exists == Book{}) {
		s.HandleErr(w, http.StatusNotFound, "The book did not exist in the "+
			"library or was already deleted")
		return
	}

	s.store.DeleteBookFromDB(params["isbn"])
	books := s.store.ReadDatabaseList()
	if err := json.NewEncoder(w).Encode(books); err != nil {
		s.HandleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
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
	existingBook := s.store.FindSpecificBook(params["isbn"])
	if (existingBook == Book{}) {
		s.HandleErr(w, http.StatusNotFound, "The book did not exist in the library")
		return
	}

	createdTime := existingBook.CreateTime
	updatedTime := existingBook.UpdateTime
	var newBook Book

	if err := json.NewDecoder(r.Body).Decode(&newBook); err != nil {
		s.HandleErr(w, http.StatusBadRequest, "Failed to decode book")
		return
	}
	if newBook.ISBN != params["isbn"] {
		s.HandleErr(w, http.StatusForbidden, "Not allowed to change ISBN")
		return
	}
	if time.Since(updatedTime) < s.minDurationBetweenUpdates {
		s.HandleErr(w, http.StatusTooEarly, "Updated a few seconds ago, "+
			"please wait a moment before updating again")
		return
	}
	if err := validate(newBook); err != nil {
		s.HandleErr(w, http.StatusNotAcceptable, err.Error())
		return
	}

	newBook.CreateTime = createdTime
	newBook.UpdateTime = time.Now()
	s.store.DeleteBookFromDB(newBook.ISBN)
	s.store.InsertIntoDatabase(newBook)

	if err := json.NewEncoder(w).Encode(newBook); err != nil {
		s.HandleErr(w, http.StatusBadRequest, "Failed to Encode the book instance")
		return
	}
}
