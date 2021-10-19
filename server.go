package library

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	librarypb "github.com/NicolaiMordrup/library/gen/proto/go/librarypb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// checks if we have implementat all the different functions
var _ librarypb.LibraryServiceServer = new(libraryServiceServer)

// libraryServiceServer for the grpc server struct
type libraryServiceServer struct {
	librarypb.UnsafeLibraryServiceServer
	store                     DBStorage
	minDurationBetweenUpdates time.Duration
	log                       *zap.SugaredLogger
}

// NewServer creates a new GRPC server instance.
func NewServer(
	dataBase *sql.DB,
	logger *zap.SugaredLogger,
	minDurationTimeBetweenUpdates time.Duration,
) *libraryServiceServer {

	s := &libraryServiceServer{}

	s.store = DBStorage{db: dataBase, log: logger}
	s.log = logger
	s.minDurationBetweenUpdates = minDurationTimeBetweenUpdates
	return s
}

// RunGRPCServer initializes and Starts the grpc server. Here we listen on the
// port given and then if successful we register a library server service.
func (s *libraryServiceServer) RunGRPCServer(addr string) error {
	listenOn := addr
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenOn, err)
	}

	server := grpc.NewServer()
	librarypb.RegisterLibraryServiceServer(server, s)
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

// CreateBook creates a Book instance and checks that the right information have
// been passed. If the information is validated then we store the information in
// our database and sends the successfully added book back as response.
func (s *libraryServiceServer) CreateBook(ctx context.Context,
	req *librarypb.CreateBookRequest) (*librarypb.Book, error) {

	newBook := NewBookFromProto(req.Book) // creates a Book instance

	if !(req.Book.GetCreateTime().AsTime().Unix() == 0 &&
		req.Book.GetUpdateTime().AsTime().Unix() == 0) {
		return nil, status.Errorf(codes.PermissionDenied,
			"not allowed to change CreateTime or UpdateTime")
	}
	if err := validate(newBook); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	newBook.CreateTime = time.Now()
	newBook.UpdateTime = time.Now()

	if err := s.store.InsertIntoDatabase(newBook); err != nil {
		return nil, status.Errorf(codes.AlreadyExists,
			"the book with this isbn already existed")
	}

	return newBook.AsProto(), nil
}

// GetBook retreives a specific book that exists in the library structure.
func (s *libraryServiceServer) GetBook(ctx context.Context,
	req *librarypb.GetBookRequest) (*librarypb.Book, error) {

	isbnPath := req.GetName()
	bookIsbn := strings.Split(isbnPath, "/")[1]

	book := s.store.FindSpecificBook(bookIsbn)
	if (Book{} == book) {
		return nil, status.Errorf(codes.AlreadyExists,
			"the book did not exist in the library")
	}

	return book.AsProto(), nil
}

// UpdateBook updates a book instance and checks that the right information have
// been passed. If the information is validated then we store the information in
// the database and sends back which book we updated as response
func (s *libraryServiceServer) UpdateBook(ctx context.Context,
	req *librarypb.UpdateBookRequest) (*librarypb.Book, error) {
	bookIsbn := req.Book.GetName()

	existingBook := s.store.FindSpecificBook(bookIsbn)
	if (existingBook == Book{}) {
		return nil, status.Errorf(codes.NotFound,
			"the book did not exist in the library")
	}

	newBook := NewBookFromProto(req.Book)
	newBook.ISBN = bookIsbn

	createdTime := existingBook.CreateTime
	updatedTime := existingBook.UpdateTime

	if existingBook.ISBN != newBook.ISBN {
		return nil, status.Errorf(codes.PermissionDenied,
			"not allowed to chang the ISBN")
	}
	if time.Since(updatedTime) < s.minDurationBetweenUpdates {
		return nil, status.Errorf(codes.Internal, "updated a few seconds ago, "+
			"please wait a moment before updating again")
	}
	if err := validate(newBook); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	newBook.CreateTime = createdTime
	newBook.UpdateTime = time.Now()

	if err := s.store.DeleteBookFromDB(bookIsbn); err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}
	if err := s.store.InsertIntoDatabase(newBook); err != nil {
		return nil, status.Errorf(codes.AlreadyExists, err.Error())
	}

	return newBook.AsProto(), nil
}

// DeleteBook deletes a book instance from the library database.
// If successful it sends back which book we deleted as response
func (s *libraryServiceServer) DeleteBook(ctx context.Context,
	req *librarypb.DeleteBookRequest) (*librarypb.Book, error) {

	isbnPath := req.GetName()
	bookIsbn := strings.Split(isbnPath, "/")[1]

	exists := s.store.FindSpecificBook(bookIsbn)

	if err := s.store.DeleteBookFromDB(bookIsbn); err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}
	return exists.AsProto(), nil
}

// ListBooks retreives all the books that exists in the library structure
// database. if successful, it sends all the book instances as a response to the
// GRPC gateway.
func (s *libraryServiceServer) ListBooks(ctx context.Context,
	req *librarypb.ListBooksRequest) (*librarypb.ListBooksResponse, error) {

	Books := s.store.ReadDatabaseList() // reads all the books from database

	var BooksConvert []*librarypb.Book

	for _, book := range Books {
		BooksConvert = append(BooksConvert, book.AsProto())
	}

	booksResp := &librarypb.ListBooksResponse{
		Book: BooksConvert,
	}
	return booksResp, nil
}
