package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

var MyServer Server

func Router(mockdataFlag, getFlag bool) *mux.Router {
	if mockdataFlag {
		mockData(getFlag) //only creates mock data when we have GET/UPDATE/DELETE methods
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/books", MyServer.GetBooks).Methods("GET")
	router.HandleFunc("/api/books/{isbn}", MyServer.GetBook).Methods("GET")
	router.HandleFunc("/api/books/{isbn}", MyServer.CreateBook).Methods("POST")
	//router.HandleFunc("/api/books/{isbn}", MyServer.UpdateBook).Methods("PUT")
	//router.HandleFunc("/api/books/{isbn}", MyServer.DeleteBook).Methods("DELETE")
	return router
}

func mockData(getFlag bool) {
	if getFlag {
		MyServer.booksList = append(MyServer.booksList, Book{ISBN: "1233211233212",
			Title: "book_1", CreateTime: time.Now().Format("2006-Jan-02 Monday 03:04:05"), Author: &Author{FirstName: "nico",
				LastName: "M"}})

		MyServer.books = make(map[string]Book)
		MyServer.books["1233211233212"] = Book{ISBN: "1233211233212", Title: "book_1",
			CreateTime: time.Now().Format("2006-Jan-02 Monday 03:04:05"), Author: &Author{FirstName: "nico", LastName: "M"}}

		return
	}

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

}

func assertEqual(t *testing.T, got, wanted interface{}, warningMessage string) {
	if !reflect.DeepEqual(got, wanted) {
		t.Errorf("got %v want %v", got, wanted)
	}
}

func assertContentType(t testing.TB, response *httptest.ResponseRecorder, want, warningMessage string) {
	t.Helper()
	if response.Result().Header.Get("content-type") != want {
		t.Errorf("response did not have content-type of %s, got %v", want, response.Result().Header)
	}
}

func assertNoError(t testing.TB, got error, warningMessage string) {
	t.Helper()

	if got != nil {
		t.Errorf("got error %q want nil", got)
	}
}

func assertError(t testing.TB, got error, warningMessage string) {
	t.Helper()
	if got == nil {
		t.Errorf("got error nil want %q", got)
	}
}

func assertStatus(t testing.TB, got, want int, warningMessage string) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func TestGetBooks(t *testing.T) { //List
	t.Run("gets all the books in the library", func(t *testing.T) {
		// Arange
		request, _ := http.NewRequest(http.MethodGet, "/api/books", nil)
		response := httptest.NewRecorder()
		Router(true, false).ServeHTTP(response, request)
		want := MyServer.booksList

		var got []Book
		err := json.NewDecoder(response.Body).Decode(&got) // Act

		//assert
		assertContentType(t, response, jsonContentType, "Should have the json content type application/json")
		assertNoError(t, err, "Should have no errors")
		assertStatus(t, response.Code, http.StatusOK, "Should jave status code 200")
		assertEqual(t, got, want, "Should be equal")
	})

	t.Run("get a specific book in the library", func(t *testing.T) {
		// Arange
		isbn := "1233211233212"
		request, _ := http.NewRequest(http.MethodGet, "/api/books/"+isbn, nil)
		response := httptest.NewRecorder()
		Router(true, true).ServeHTTP(response, request)
		want := MyServer.books[isbn]

		var got Book
		err := json.NewDecoder(response.Body).Decode(&got) // Act

		//assert
		assertContentType(t, response, jsonContentType, "Should have the json content type application/json")
		assertNoError(t, err, "Should have no errors")
		assertStatus(t, response.Code, http.StatusOK, "Should jave status code 200")
		assertEqual(t, got, want, "Should be equal")
	})

	t.Run("get a book that does not exist in the library", func(t *testing.T) {
		// Arange
		isbn := "1233211233216"
		request, _ := http.NewRequest(http.MethodGet, "/api/books/"+isbn, nil)
		response := httptest.NewRecorder()
		Router(true, true).ServeHTTP(response, request)

		var got Book
		err := json.NewDecoder(response.Body).Decode(&got) // Act

		//assert
		assertContentType(t, response, jsonContentType, "Should have the json content type application/json")
		assertError(t, err, "Should have errors")
		assertStatus(t, response.Code, http.StatusNotFound, "Should jave status code 404")
	})
}

func TestCreateBook(t *testing.T) {
	t.Run("Creates a book and stores it in the library", func(t *testing.T) {
		// Arange
		isbn := "1233211233215"
		data := &Book{ISBN: "1233211233215",
			Title: "book_1", Author: &Author{FirstName: "nico",
				LastName: "M"}}

		jsonBytes, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		request, _ := http.NewRequest(http.MethodPost, "/api/books/"+isbn, bytes.NewReader(jsonBytes))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		Router(false, false).ServeHTTP(response, request)
		// Act
		//got := MyServer.books[isbn]

		//assert
		//assertContentType(t, response, jsonContentType, "Should have the json content type application/json")
		assertStatus(t, response.Code, http.StatusOK, "Should jave status code 200")
		//	assertEqual(t, got, want, "Should be equal")
	})

}

/*
func TestUpdateBooks(t *testing.T) {
	t.Run("Updates a specific book which exists in the library", func(t *testing.T) {
		// Arange
		want := &Book{ISBN: "1233211233212",
			Title: "book_1", CreateTime: time.Now().Format("2006-Jan-02 Monday 03:04:05"), UpdateTime: "", Author: &Author{FirstName: "nico",
				LastName: "M"}}
		jsonBook, _ := json.Marshal(want)
		request, _ := http.NewRequest(http.MethodPost, "/api/books", bytes.NewBuffer(jsonBook))
		response := httptest.NewRecorder()
		Router(false, false).ServeHTTP(response, request)

		var got []Book
		err := json.NewDecoder(response.Body).Decode(&got) // Act

		//assert
		assertContentType(t, response, jsonContentType, "Should have the json content type application/json")
		assertNoError(t, err, "Should have no errors")
		assertStatus(t, response.Code, http.StatusOK, "Should jave status code 200")
		assertEqual(t, got, want, "Should be equal")
	})

}*/
