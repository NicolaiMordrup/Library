# Library GRPC server and GRPC gateway CRUD api
This is a GRPC server and a GRPC gateway which implements a CRUD 
rest API.


Endpoints: 

| Method        | Endpoint      |  Description             |
| ------------- | ------------- | ------------------------ |
| POST          | /books        | Create a book            |
| GET           | /books         | Get list of all books    |
| GET           | /books/{isbn}   | Get user by isbn           |
| PUT           | /books/{isbn}   | Update user by isbn      |
| DELETE        | /books/{isbn}   | Delete user by isbn      |

## Run locally

- Clone the repository
```
git clone https://github.com/NicolaiMordrup/library.git
```
-  Open a terminal in the root directory of your code and run the following command to start the application.

### Run with local go installtion
```
go run cmd/main.go
```