package parser

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestRealBSLParseFiles(t *testing.T) {
	entries, err := os.ReadDir("testdata/real_bsl")
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".bsl") {
			continue
		}

		path := "testdata/real_bsl/" + entry.Name()

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		t.Run(entry.Name(), func(t *testing.T) {
			input := string(data)
			p := NewParser(input)
			p.ParseModule()
			errs := p.Errors()
			if strings.HasPrefix(entry.Name(), "bug_") {
				if len(errs) == 0 {
					t.Errorf("BUG: parser did not report error for %s", entry.Name())
				}
			} else {
				if len(errs) > 0 {
					for _, e := range errs {
						t.Errorf("line %d:%d: %s", e.Line, e.Col, e.Message)
					}
					fmt.Fprintf(os.Stderr, "=== %s ===\n%s\n", entry.Name(), input)
				}
			}
		})
	}
}
