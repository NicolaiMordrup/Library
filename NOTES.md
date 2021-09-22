# Notes

* Validation, checker functions etc should return only error (Done)
* Collect field errors while validating and return all of them in the error (maybe done?)
* Avoid chaining else if statements (all code after an if is implicitly an else) (Done)
* Move time checker function to where it is used (Done)

## Next steps

### Persistence (Database)

* https://pkg.go.dev/modernc.org/sqlite
* https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/05.3.html (change import to use modernc.org)

### (Later) Configuration

### Testing

* Continue with learn go with tests
* Use https://pkg.go.dev/net/http/httptest to test your API

## Custom errors

A Go error is anything which has a function `.Error() string`.

For example:

```go
type myerror string

func (e myerror) Error() string {
	return string(e)
}

func dostuff() error {
	return myerror("test")
}
```

```go
type ValidationError struct {
	msg           string
	invalidFields []string
}

func (e *ValidationError) Error() string {
	msg := e.msg
	msg += "invalid fields: " + strings.Join(e.invalidFields, ", ")
	return msg
}

func (e *ValidationError) AddFieldViolation(fieldName string) {
	e.invalidFields = append(e.invalidFields, fieldName)
}

func dostuff() error {
	err := &ValidationError{
		msg: "not implemented",
	}
	err.AddFieldViolation("firstName")
	err.AddFieldViolation("lastName")
	return err
}
```
