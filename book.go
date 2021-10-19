package library

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	librarypb "github.com/NicolaiMordrup/library/gen/proto/go/librarypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Struct for the book properties.
type Book struct {
	ISBN       string    `json:"isbn"` // The identification of the books
	Title      string    `json:"title"`
	CreateTime time.Time `json:"createTime"` // The time of creation of book instance
	UpdateTime time.Time `json:"updateTime"` // The time of update for book instance
	Publisher  string    `json:"publisher"`
	Author     Author    `json:"author"` // Embedded author struct
}

// Struct for the books Author properties.
type Author struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// The regex patterns for the validate function
var (
	isbnPattern      = regexp.MustCompile(`^\d{13}$`)
	titlePattern     = regexp.MustCompile(`.`)
	firstNamePattern = regexp.MustCompile(`^[a-zA-Z]+(?:\s+[a-zA-Z]+)*$`)
	LastNamePattern  = regexp.MustCompile(`^[a-zA-Z]+(?:\s+[a-zA-Z]+)*$`)
	publisherPattern = regexp.MustCompile(`^[a-zA-Z]+(?:\s+[a-zA-Z]+)*$`)
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
	if matchedPublisher := publisherPattern.MatchString(b.Publisher); !matchedPublisher {
		fieldErrors = append(fieldErrors, " Publishers name")
	}

	if len(fieldErrors) != 0 {
		return fmt.Errorf("validation failed, field error(s):%v."+
			" Fix these error before proceeding",
			strings.Join(fieldErrors, ", "))
	}
	return nil
}

// NewBookFromProto converts a *librarypb.Book which is on the proto fromat
// to the Book instance such that the database can deal with it.
func NewBookFromProto(b *librarypb.Book) Book {
	return Book{
		ISBN:       b.GetName(),
		Title:      b.GetTitle(),
		Publisher:  b.GetPublisher(),
		CreateTime: b.GetCreateTime().AsTime(),
		UpdateTime: b.GetUpdateTime().AsTime(),
		Author: Author{
			FirstName: b.Author.GetFirstName(),
			LastName:  b.Author.GetLastName(),
		},
	}
}

// AsProto converts a Book instance to the *librarypb.Book which is of proto
// instance such that the response can deal with it.
func (b *Book) AsProto() *librarypb.Book {
	return &librarypb.Book{
		Name:       b.ISBN,
		Title:      b.Title,
		Publisher:  b.Publisher,
		CreateTime: timestamppb.New(b.CreateTime),
		UpdateTime: timestamppb.New(b.UpdateTime),
		Author: &librarypb.Author{
			FirstName: b.Author.FirstName,
			LastName:  b.Author.LastName,
		},
	}
}
