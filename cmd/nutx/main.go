package main

import (
	"fernandez14/nutx/internal/authorizer"
	"os"
)

func main() {
	// Run authorizer scanner using stdin & stdout.
	authorizer.Scanner(os.Stdin, os.Stdout)
}
