// Copyright (C) 2020  WPEngine
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
        {{range .}}<img src="/ipfs/{{.}}" style="max-width:100%; max-height:100vh; margin:auto" />
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
