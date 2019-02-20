package parser

import (
	"genopi/internal/common"
	"net/http"
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

	var required bool
	if comment[i-1] == '?' {
		i = i - 1
	} else {
		required = true
	}

	return strings.TrimSpace(comment[1:i]), common.Param{
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
