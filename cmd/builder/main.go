package main

import (
	"fmt"
	"html/template"
	"os"
)

func main() {
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
