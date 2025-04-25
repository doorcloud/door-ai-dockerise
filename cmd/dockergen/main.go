package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aliou/dockerfile-gen/internal/loop"
)

func main() {
	// Parse command line flags
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("usage: dockergen <path>")
	}

	// Create filesystem from repository path
	fsys := os.DirFS(flag.Arg(0))

	// Run the generation loop
	dockerfile, err := loop.Run(context.Background(), fsys)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(dockerfile)
}
