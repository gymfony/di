package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/gymfony/di/digen"
)

func main() {
	fs := flag.NewFlagSet("digen", flag.ContinueOnError)
	path := fs.String("project-path", "./", "The path to your project where dependencies need to be generated")
	verbose := fs.Bool("v", false, "Enable verbose (debug) logging")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse flags: %v\n", err)
		os.Exit(1)
	}

	if *path == "" {
		fmt.Fprintln(os.Stderr, "Error: project path cannot be empty")
		os.Exit(1)
	}

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	parser := digen.NewParser(log)

	graph, err := parser.Parse(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	//TODO temporary output for GYM-9 stage
	fmt.Println(graph)
}
