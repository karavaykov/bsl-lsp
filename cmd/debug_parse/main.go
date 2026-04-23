package main

import (
	"fmt"
	"os"
	"time"

	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func main() {
	data, err := os.ReadFile("/mnt/c/Users/karavaikov.s/AI 1C/CommonModules/ИнтеграцияGLPI/Ext/Module.bsl")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	input := string(data)
	
	// Binary search: try parsing first N lines
	for _, n := range []int{100, 200, 300, 500, 800, 1000, 1200, 1500} {
		lineCount := 0
		pos := 0
		for i, ch := range input {
			if ch == '\n' {
				lineCount++
				if lineCount == n {
					pos = i
					break
				}
			}
		}
		if pos == 0 {
			pos = len(input)
		}
		snippet := input[:pos]

		done := make(chan bool)
		var errs []parser.ParseError

		go func() {
			p := parser.NewParser(snippet)
			p.ParseModule()
			errs = p.Errors()
			done <- true
		}()

		select {
		case <-done:
			fmt.Printf("Lines 1-%d: %d errors\n", n, len(errs))
			for i, e := range errs {
				if i >= 5 {
					fmt.Printf("  ... and %d more\n", len(errs)-i)
					break
				}
				fmt.Printf("  line %d:%d: %s\n", e.Line, e.Col, e.Message)
			}
		case <-time.After(3 * time.Second):
			fmt.Printf("Lines 1-%d: TIMEOUT\n", n)
		}
	}
}
