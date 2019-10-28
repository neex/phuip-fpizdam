package main

import (
	"net/http"
	"strings"
)

type DetectMethod struct {
	PHPOptionEnable  string
	PHPOptionDisable string
	Check            func(resp *http.Response, data []byte) bool
}

// TODO: we need detection method with .php at the end of each option
var Methods = map[string]*DetectMethod{
	"session.auto_start": {
		PHPOptionEnable:  "session.auto_start=1;;;",
		PHPOptionDisable: "session.auto_start=0;;;",
		Check: func(resp *http.Response, _ []byte) bool {
			return strings.Contains(resp.Header.Get("set-cookie"), "PHPSESSID")
		},
	},
}
