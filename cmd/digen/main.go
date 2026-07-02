package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gymfony/di/digen"
)

func main() {
	logger := digen.NewParserLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	parser := digen.NewParser(logger)
	graph, err := parser.Parse(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	fmt.Println(graph)
}
