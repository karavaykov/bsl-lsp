package main

import (
	"fmt"
	"log"
	"os"

	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/analysis/linters"
	"github.com/karavaikov/bsl-lsp/internal/lsp"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

const usage = `Usage: bsl-lsp <command> [options] [files...]

Commands:
  lsp                          Start LSP server (default, stdin/stdout)
  check <file.bsl>...          Check syntax and run static analysis
  format <file.bsl>...         Format BSL files (in-place)
  format --stdout <file.bsl>   Format BSL file to stdout

Examples:
  bsl-lsp check module.bsl
  bsl-lsp format module.bsl
  bsl-lsp format --stdout module.bsl > formatted.bsl
`

func main() {
	log.SetFlags(log.Lshortfile)

	args := os.Args[1:]
	if len(args) == 0 || (len(args) == 1 && (args[0] == "lsp" || args[0] == "help" || args[0] == "--help" || args[0] == "-h")) {
		if len(args) > 0 && args[0] == "lsp" {
			runLSP()
			return
		}
		fmt.Fprint(os.Stderr, usage)
		os.Exit(0)
	}

	switch args[0] {
	case "lsp":
		runLSP()
	case "check":
		runCheck(args[1:])
	case "format":
		runFormat(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n%s", args[0], usage)
		os.Exit(1)
	}
}

func runLSP() {
	log.Println("bsl-lsp starting...")
	if err := lsp.Run(); err != nil {
		log.Fatalf("bsl-lsp error: %v", err)
	}
}

func runCheck(files []string) {
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: bsl-lsp check <file.bsl>...")
		os.Exit(1)
	}

	exitCode := 0
	for _, file := range files {
		text, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: error reading file: %v\n", file, err)
			exitCode = 1
			continue
		}

		p := parser.NewParser(string(text))
		mod := p.ParseModule()

		parseErrors := p.Errors()
		for _, e := range parseErrors {
			fmt.Printf("%s:%d:%d: parse error: %s\n", file, e.Line, e.Col, e.Message)
			exitCode = 1
		}

		symbols := analysis.BuildSymbolTable(mod)
		diags := linters.RunAll(mod, symbols)

		severityNames := map[int]string{
			linters.SevWarning: "warning",
			linters.SevInfo:    "info",
		}

		for _, d := range diags {
			sev := severityNames[d.Severity]
			if sev == "" {
				sev = fmt.Sprintf("severity=%d", d.Severity)
			}
			fmt.Printf("%s:%d:%d: [%s/%s] %s\n", file, d.Line, d.Col, sev, d.Code, d.Message)
			if d.Severity == linters.SevWarning {
				exitCode = 1
			}
		}

		if len(parseErrors) == 0 && len(diags) == 0 {
			fmt.Printf("%s: OK\n", file)
		}
	}
	os.Exit(exitCode)
}

func runFormat(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: bsl-lsp format [--stdout] <file.bsl>...")
		os.Exit(1)
	}

	stdout := false
	files := args
	if args[0] == "--stdout" {
		stdout = true
		files = args[1:]
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: bsl-lsp format [--stdout] <file.bsl>...")
		os.Exit(1)
	}

	exitCode := 0
	for _, file := range files {
		text, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: error reading file: %v\n", file, err)
			exitCode = 1
			continue
		}

		formatted := analysis.FormatDocument(string(text), 4, true)

		if stdout {
			fmt.Print(formatted)
		} else {
			info, err := os.Stat(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: error stat: %v\n", file, err)
				exitCode = 1
				continue
			}
			if err := os.WriteFile(file, []byte(formatted), info.Mode()); err != nil {
				fmt.Fprintf(os.Stderr, "%s: error writing file: %v\n", file, err)
				exitCode = 1
				continue
			}
			fmt.Printf("%s: formatted\n", file)
		}
	}
	os.Exit(exitCode)
}

func init() {
	log.SetFlags(0)
}
