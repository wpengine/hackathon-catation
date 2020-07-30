package build

import (
	"bytes"
	"fmt"
	"html/template"
)

const html = `
<html>
    <head>
        <title>Images</title>
    </head>
    <body>
        <h1>Images</h1>
        {{range .}}<img src="/ipfs/{{.}}" />
        {{end}}
    </body>
</html>
`

func IndexHTML(hashes ...string) ([]byte, error) {
	t, err := template.New("index.html").Parse(html)
	if err != nil {
		return nil, fmt.Errorf("internal error: cannot parse default index.html template: %w", err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, hashes)
	if err != nil {
		return nil, fmt.Errorf("building index.html: %w", err)
	}
	return buf.Bytes(), nil
}
