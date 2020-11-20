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
	"os"

	"github.com/wpengine/hackathon-catation/cmd/shortener/bitly"
	"github.com/wpengine/hackathon-catation/internal"
)

func main() {
	internal.PrintGPLBanner("catation", "2020")

	hash := os.Args[1]

	bitly := bitly.API{
		Key: os.Getenv("BITLY_API_KEY"),
	}

	link, err := bitly.Shorten("http://ipfs.io/ipfs/" + hash)
	if err != nil {
		die(err)
	}

	fmt.Println(link)
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
