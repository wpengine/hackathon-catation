package main

import (
	"fmt"
	"os"
)

func main() {
	fn := os.Args[1]
	fh, err := os.Open(fn)
	if err != nil {
		die(err)
	}
	defer fh.Close()
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
