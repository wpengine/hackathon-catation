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

package main

import (
	"fmt"
	"html/template"
	"os"

	"github.com/wpengine/hackathon-catation/internal"
)

func main() {
	internal.PrintGPLBanner("catation", "2020")

	hashes := os.Args[1:]
	t, err := template.ParseFiles("template.html")
	if err != nil {
		die(err)
	}

	f, err := os.Create("images.html")
	if err != nil {
		die(err)
	}

	err = t.Execute(f, hashes)

	f.Close()
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
