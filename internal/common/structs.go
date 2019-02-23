package common

import (
	"net/http"
)

type Method struct {
	Package  string
	Name     string
	Comments []string
}

type Struct struct {
	Package string
	Name    string
	Fields  []Field
}

type Field struct {
	Name     string
	Type     string
	Optional bool
	Array    bool
}

type Api struct {
	Status    Status
	Endpoints []Endpoint
}

type Status struct {
	Title   string
	Version string
	URL     string
}

type Endpoint struct {
	Name        string
	Method      string
	Path        string
	PathParams  map[string]Param
	QueryParams map[string]Param
	Headers     map[string]Param
	Body        *Struct
	Responses   []Response
}

type Param struct {
	Type     string
	Desc     string
	Required bool
}

type Response struct {
	Code int
	Body string
}

type body struct {
	Type user `json:"type"`
	Desc []user
}

type user struct {
	Type string `json:"type"`
	Desc []string
}

// Get users
// GET /v1/users/{id}
// {id:uuid} User id
// (name?:string) User name
// [Access-Key?:string] Token access key
// <common.body>
// 200 OK
// 400 If missing params
func test(w http.ResponseWriter, r *http.Request) {
}
