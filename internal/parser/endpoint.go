package parser

import (
	"genopi/internal/common"
	"net/http"
	"strconv"
	"strings"
)

func tryMethod(comment string) (method, path string, ok bool) {
	if strings.HasPrefix(comment, http.MethodGet) {
		method = "get"
	} else if strings.HasPrefix(comment, http.MethodPost) {
		method = "post"
	} else if strings.HasPrefix(comment, http.MethodPut) {
		method = "put"
	} else if strings.HasPrefix(comment, http.MethodDelete) {
		method = "delete"
	} else {
		return "", "", false
	}

	path = strings.TrimSpace(comment[len(method):])

	return method, path, true
}

func tryParam(comment, left, right string) (key string, value common.Param, ok bool) {
	if !strings.HasPrefix(comment, left) {
		return "", common.Param{}, false
	}

	i := strings.Index(comment, right)
	if i == -1 {
		return "", common.Param{}, false
	}
	j := i + 1

	prefix := comment[1:i]
	ps := strings.Split(prefix, ":")
	ident := strings.TrimSpace(ps[0])
	typ := strings.TrimSpace(ps[1])

	var required bool
	if strings.HasSuffix(ident, "?") {
		ident = ident[:len(ident)-1]
	} else {
		required = true
	}

	return ident, common.Param{
		Type:     typ,
		Desc:     strings.TrimSpace(comment[j:]),
		Required: required,
	}, true
}

func tryBody(comment string) (body string, ok bool) {
	if strings.HasPrefix(comment, "<") && strings.HasSuffix(comment, ">") {
		return strings.TrimSpace(comment[1 : len(comment)-1]), true
	}
	return "", false
}

func tryResponse(comment string) (common.Response, bool) {
	if len(comment) < 3 {
		return common.Response{}, false
	}

	code, err := strconv.ParseInt(comment[0:3], 10, 32)
	if err != nil {
		return common.Response{}, false
	}

	body := strings.TrimSpace(comment[3:])
	return common.Response{
		Code: int(code),
		Body: body,
	}, true
}
