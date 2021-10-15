package main

import (
	"context"
	"fmt"
	"log"

	librarypb "github.com/NicolaiMordrup/library/gen/proto/go"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	connectTo := "127.0.0.1:8080"
	conn, err := grpc.Dial(connectTo, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to libraryService on %s: %w",
			connectTo, err)
	}
	log.Println("Connected to", connectTo)
	libraryStore := librarypb.NewLibraryServiceClient(conn)

	// Creates a book
	if _, err := libraryStore.CreateBook(context.Background(),
		&librarypb.CreateBookRequest{
			Book: &librarypb.Book{
				Isbn:      "1234567891123",
				Title:     "F1",
				Publisher: "svenP",
				Author: &librarypb.Author{
					FirstName: "alfred",
					LastName:  "gsadasd",
				},
			},
		}); err != nil {
		return fmt.Errorf("failed to create book: %w", err)
	}
	log.Println("Successfully created book")

	//Updates a book
	res, err := libraryStore.ReadBook(context.Background(),
		&librarypb.ReadBookRequest{Isbn: "1234567891123"})
	if err != nil {
		return fmt.Errorf("failed to fetch a book: %w", err)
	}

	log.Println("Successfully fetched a  book, %v", res.GetBook())

	//updated a book
	resUpdate, errUpdate := libraryStore.UpdateBook(context.Background(),
		&librarypb.UpdateBookRequest{Book: &librarypb.Book{
			Isbn:      "1234567891123",
			Title:     "F1 memories",
			Publisher: "svenPersson",
			Author: &librarypb.Author{
				FirstName: "alfred",
				LastName:  "gsadasd",
			},
		},
		})
	if errUpdate != nil {
		return fmt.Errorf("failed to update book: %w", errUpdate)
	}

	log.Println("Successfully updated a  book, %v", resUpdate.GetBook())

	/*
		//list all the books in the library
		resList, errList := libraryStore.ListBook(context.Background(),
			&librarypb.ListBookRequest{})

		if errList != nil {
			return fmt.Errorf("failed to fetch a book: %w", errList)
		}

		//for book, _ := range resList.B {

		//}

		log.Println("Successfully fetched a  book, %v", resList.GetBook())
	*/
	//Deletes a book
	_, errDelete := libraryStore.DeleteBook(context.Background(),
		&librarypb.DeleteBookRequest{Isbn: "1234567891123"})
	if errDelete != nil {
		return fmt.Errorf("failed to fetch a book: %w", errDelete)
	}
	fmt.Println("successfully deleted a book")

	return nil

}
