# Genopi

Generate OpenAPI 3.0 documentation from Go src documentation.

### Code template

Comment functions in the source code with the following template:

```go
// Function summary
// METHOD /path/{path}
// {path:type} This is a path param, must exist in path above
// (query:type) This is a query param
// [header:type] This is a header param
// <package.BodyStruct>
// RESPONSE_CODE package.ResponseStruct
func rest(w http.ResponseWriter, r *http.Request) {
    ...
}
```

The first line must be a function summary, followed by a method and path. The rest is optional. If there are path 
variables, they must exist in the path. Structs must have package names and be separated by a dot. Structs must exist 
in the working directory, or a subdirectory, from where `genopi` is run so make sure to create a vendor directory if 
using external structs. Params can be optional using `?` and for structs they are optional if they are pointers, 
see examples below. If struct variables are annotated with `json:"name"`, name will be used instead of the variables name.

Some special types are allowed:

* `time.Time` becomes type string with format date
* `uuid.UUID` becomes type string with format uuid

### Examples

Minimal

```go
// Ping the server
// GET /ping
func ping(w http.ResponseWriter, r *http.Request) {
    ...
}
```

Get with path param

```go
package myapp

type User struct {
    ID uuid.UUID `json:"id"`
}

type Message struct {
    Created time.Time `json:"created"`
    Message string    `json:"message"`
}

// Get user by id
// GET /users/{user_id}
// {user_id:uuid} User id path param
// (optional_param?:string) Optional query param
// 200 myapp.User
// 400 myapp.Message
func getUser(w http.ResponseWriter, r *http.Request) {
    ...
}
```

Post with body and optional params. Skip struct if response code provides empty response.

```go
package myapp

type Body struct {
    Required string
    Optional *string
}

// Add user info
// POST /users/{user_id}/info
// <myapp.Body>
// 204
// 400
func addUserInfo(w http.ResponseWriter, r *http.Request) {
    ...
}
```


### Installing and running

Flags:

- `-t` Api title
- `-u` Api base url
- `-v` Api version
- `-o` Path to store generated api documentation  

Generating api documentation:

1. `go get github.com/plankt/genopi/cmd/genopi`
2. `cd $GOPATH/src/your/project`
3. `genopi -t "Api title" -u "https://api.url" -v "1.0" -o "docs/api.yaml"`

### Roadmap

- [x] Parse api from comments in source code
- [x] Generate OpenAPI 3.0 documentation
- [ ] Provide api tests using `golang.org/pkg/testing/quick`
