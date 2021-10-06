package librarystore

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Struct for the book properties.
type Book struct {
	ISBN       string    `json:"isbn"` // The identification of the books
	Title      string    `json:"title"`
	CreateTime time.Time `json:"createTime"` // The time of creation of book instance
	UpdateTime time.Time `json:"updateTime"` // The time of update for book instance
	Publisher  string    `json:"publisher"`
	Author     *Author   `json:"author"` // Embedded author struct
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
		return fmt.Errorf("validation failed, field error(s):%v. Fix these error before proceeding",
			strings.Join(fieldErrors, ", "))
	}
	return nil
}
