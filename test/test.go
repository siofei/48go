package main

import (
	"fmt"
	"os"
)

func main() {
	_, err := os.Open("slkdf.txt")
	fmt.Fprintf(nil, "%T\n", err)
}
