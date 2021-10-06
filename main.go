package main

import (
	"log"
	"net/http"

	library "github.com/Nicolai.mordrup/Library/librarystore"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := library.NewDB("file:librarystorage.db")
	library.Check(err, "failed to open sqlite connection")
	library.Check(library.EnsureSchema(db), "migration failed")

	myServer := library.NewServer(db)

	// Starts the server and listens on port 8000
	log.Fatal(http.ListenAndServe(":8000", myServer))
}
