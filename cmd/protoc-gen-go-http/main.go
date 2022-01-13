package main

import (
	"flag"
	"fmt"
)

const version = "v0.0.1"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")

	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-http %v\n", version)
		return
	}

}
