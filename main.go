package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	library "github.com/Nicolai.mordrup/Library/librarystore"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := NewDB("file:librarystorage.db")
	check(err, "failed to open sqlite connection")
	check(EnsureSchema(db), "migration failed")

	myServer := library.NewServer(db)

	// Starts the server and listens on port 8000
	log.Fatal(http.ListenAndServe(":8000", myServer))
}

func check(err error, msg string) {
	if err != nil {
		fmt.Printf("%v, err: %v\n", msg, err)
		os.Exit(1)
	}
}
