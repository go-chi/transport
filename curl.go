package transport

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

func curl(r *http.Request, body *bytes.Buffer) string {
	var b strings.Builder

	fmt.Fprintf(&b, "curl")
	if r.Method != "GET" && r.Method != "POST" {
		fmt.Fprintf(&b, " -X %s", r.Method)
	}

	fmt.Fprintf(&b, " %s", singleQuoted(r.URL.String()))

	if r.Method == "POST" {
		fmt.Fprintf(&b, " --data-raw %s", singleQuoted(body.String()))
	}

	for name, vals := range r.Header {
		for _, val := range vals {
			fmt.Fprintf(&b, " -H %s", singleQuoted(fmt.Sprintf("%s: %s", name, val)))
		}
	}

	return b.String()
}

func singleQuoted(v string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", `'\''`))
}
