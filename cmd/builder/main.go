package main

import (
	"fmt"
	"html/template"
	"os"
)

func main() {
	hashes := os.Args
	imgs := make([]template.HTML, len(hashes)-1)
	for i := 0; i < len(hashes)-1; i++ {
		imgs[i] = template.HTML(fmt.Sprintf(`<img src="/ipfs/%s">`, hashes[i+1]))
	}
	t, err := template.ParseFiles("template.html")
	if err != nil {
		die(err)
	}

	f, err := os.Create("images.html")
	if err != nil {
		die(err)
	}

	err = t.Execute(f, imgs)

	f.Close()
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
