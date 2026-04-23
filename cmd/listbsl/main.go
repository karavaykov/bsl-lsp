package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <root>\n", os.Args[0])
		os.Exit(1)
	}
	root := os.Args[1]

	// Use parser from the same module - we'll build a test binary instead
	// This is a helper to list files for the test
	f, err := os.Create("bsl_files_list.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	t0 := time.Now()
	count := 0
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".bsl") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		fmt.Fprintln(f, rel)
		count++
		return nil
	})

	fmt.Fprintf(os.Stderr, "Listed %d files in %v\n", count, time.Since(t0))
}
