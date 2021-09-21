package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
)

// Struct for the server.
type server struct {
	books     map[string]Book //map of the ISBN as key and a Book instance as value
	booksList []Book          //slice of all the Book instance
}

// Struct for the book properties.
type Book struct {
	ISBN       string    `json:"isbn"` //The identification of the books
	Title      string    `json:"title"`
	CreateTime time.Time `json:"createTime"` //The time of when the book instance was created
	UpdateTime time.Time `json:"updateTime"` // The time of when the book instance was updated
	Author     *Author   `json:"author"`     //Embeded author struct
	//	BookListPlace int       `json:"bookListPlace"` //place of book instance in the booksList   (This is a experiment)
}

// Struct for the books Author properties.
type Author struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// Checks if the user have set the time properties when they created or updated the book instance.
func checkTimeProperties(createTime, updateTime time.Time) bool {
	if createTime.IsZero() && updateTime.IsZero() {
		return true
	} else {
		return false
	}
}

// Handler for when we get an error.
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

//Validates if the given input given is correct.
//if correct we return boolean true, otherwise boolean false.
func validate(b Book) (bool, string) {

	//Todo fixa regex på de resterande

	if matchedISBN, _ := regexp.MatchString(`\d{13}`, b.ISBN); !matchedISBN {
		return false, "isbn"
	} else if matchedTitle, _ := regexp.MatchString(`.`, b.Title); !matchedTitle {
		return false, "title"
	} else if matchedFirstName, _ := regexp.MatchString(`.|^\d*$`, b.Author.FirstName); !matchedFirstName {
		return false, "authors firstname"
	} else if matchedLastName, _ := regexp.MatchString(`^\s*$|^\d*$`, b.Author.LastName); !matchedLastName {
		return false, "authors lastname"
	} else {
		return true, "_"
	}
}

//Retreives all the books that exists in the library structure.
//if succesfull, it writes the JSON encoding of the books slice to the stream
func (s *server) getBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.booksList)
}

//Retreives a specific book that exists in the library structure.
//if succesfull, it writes the JSON encoding of the specific book to the stream
func (s *server) getBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) //fetches the parameters of the http.Request URL

	if val, exists := s.books[params["isbn"]]; exists {
		json.NewEncoder(w).Encode(val)
		return
	}
}

// Creates a Book instance and checks that the right information have been passed
// If the information is validated then we store the information in our local memory
// and it writes the JSON encoding of the specific book to the stream
func (s *server) createBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	var book Book
	err := json.NewDecoder(r.Body).Decode(&book)

	if err != nil {
		handleErr(w, http.StatusBadRequest, "Failed to decode book")
		return
	} else if !checkTimeProperties(book.CreateTime, book.UpdateTime) {
		handleErr(w, http.StatusForbidden, "Not allowed to change CreateTime or UpdateTime")
		return
	} else if _, exists := s.books[book.ISBN]; exists {
		handleErr(w, http.StatusConflict, "A book with this ISBN already exits")
		return
	} else if err, field := validate(book); !err {
		var message string = "The " + field + " field was not correct filed out"
		handleErr(w, http.StatusConflict, message)
		return
	} else {
		book.CreateTime = time.Now()
		s.books[book.ISBN] = book
		s.booksList = append(s.booksList, book)
		json.NewEncoder(w).Encode(book)
	}
}

// Deletes a book instance from the library.
// if succesfull, it writes the JSON encoding of the new book slice
// without the removed book to the stream
func (s *server) deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	params := mux.Vars(r)
	if _, exists := s.books[params["isbn"]]; exists {
		for index, item := range s.booksList {
			if item.ISBN == params["isbn"] {
				s.booksList = append(s.booksList[:index], s.booksList[index+1:]...) //removes the book instance from the book slice
				delete(s.books, params["isbn"])                                     //removes the book instance from map.
				json.NewEncoder(w).Encode(s.booksList)
				return
			}
		}
	} else {
		handleErr(w, http.StatusNotFound, "The book did not exist in the library or was already deleted")
		return
	}

	/*if _, exists := s.books[params["isbn"]]; exists {
		place := s.books[params["isbn"]].BookListPlace
		s.booksList = append(s.booksList[:place], s.booksList[place+1:]...) //removes the book instance from the book slice
		delete(s.books, params["isbn"])                                     //removes the book instance from map.
		json.NewEncoder(w).Encode(s.booksList)
		return
	} else {
		handleErr(w, http.StatusNotFound, "The book did not exist in the library or was already deleted")
		return
	}*/
}

//update a book
func (s *server) updateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	params := mux.Vars(r)

	if _, exists := s.books[params["isbn"]]; exists {
		for index, item := range s.booksList {
			if item.ISBN == params["isbn"] {
				createdTime := s.booksList[index].CreateTime
				//updatedTime := s.booksList[index].UpdateTime
				s.booksList = append(s.booksList[:index], s.booksList[index+1:]...)
				delete(s.books, params["isbn"])
				var book Book
				err := json.NewDecoder(r.Body).Decode(&book)
				if err != nil {
					handleErr(w, http.StatusBadRequest, "Failed to decode book")
					return
				} else if book.ISBN != params["isbn"] {
					handleErr(w, http.StatusForbidden, "Not allowed to change ISBN")
					return
				} else if !checkTimeProperties(book.CreateTime, book.UpdateTime) {
					handleErr(w, http.StatusForbidden, "Not allowed to change CreateTime or UpdateTime")
					return
				} /* else if (time.Now().Unix() - updatedTime.Unix()) < 10 {
					handleErr(w, http.StatusForbidden, "Updated few seconds ago")
					return
				}*/

				book.CreateTime = createdTime
				book.UpdateTime = time.Now()
				s.books[book.ISBN] = book
				s.booksList = append(s.booksList, book)
				json.NewEncoder(w).Encode(book)
				return
			}
		}
	} else {
		handleErr(w, http.StatusNotFound, "The book did not exist in the library")
		return
	}
}

func main() {
	//init the Router
	r := mux.NewRouter()

	//Creates the Mock data that we play around with
	var myServer server
	myServer.booksList = append(myServer.booksList, Book{ISBN: "67894323", Title: "book_1", CreateTime: time.Now(), Author: &Author{FirstName: "nico", LastName: "M"} /*, BookListPlace: 0*/})
	myServer.booksList = append(myServer.booksList, Book{ISBN: "69989432", Title: "book_2", CreateTime: time.Now(), Author: &Author{FirstName: "Mico", LastName: "N"} /*, BookListPlace: 1*/})
	myServer.books = make(map[string]Book)
	myServer.books["67894323"] = Book{ISBN: "67894323", Title: "book_1", CreateTime: time.Now(), Author: &Author{FirstName: "nico", LastName: "M"} /*, BookListPlace: 0*/}
	myServer.books["69989432"] = Book{ISBN: "69989432", Title: "book_2", CreateTime: time.Now(), Author: &Author{FirstName: "Mico", LastName: "N"} /*, BookListPlace: 1*/}

	//create Route handlers /Endpoints
	r.HandleFunc("/api/books", myServer.getBooks).Methods("GET")
	r.HandleFunc("/api/books/{isbn}", myServer.getBook).Methods("GET")
	r.HandleFunc("/api/books", myServer.createBook).Methods("POST")
	r.HandleFunc("/api/books/{isbn}", myServer.updateBook).Methods("PUT")
	r.HandleFunc("/api/books/{isbn}", myServer.deleteBook).Methods("DELETE")

	//Starts the server and listens on port 8000
	log.Fatal(http.ListenAndServe(":8000", r))
}
