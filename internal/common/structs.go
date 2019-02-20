package common

import (
	"net/http"
)

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
	Body        *string
	Responses   []Response
}

type Param struct {
	Desc     string
	Required bool
}

type Response struct {
	Code int
	Body *string
}

// Get users
// GET /v1/users/{id}
// {id} User id
// (name?) User name
// [Access-Key?] Token access key
// <common.Param>
// 200: string
func test(w http.ResponseWriter, r *http.Request) {
}
