package common

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
	Structs   []Struct
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
	Type     string
	Desc     string
	Required bool
}

type Response struct {
	Code int
	Type *string
}
