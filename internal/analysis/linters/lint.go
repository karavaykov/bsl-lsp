package linters

import (
	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

type LintDiagnostic struct {
	Line     int
	Col      int
	Length   int
	Message  string
	Code     string
	Severity int
}

const (
	SevWarning = 2
	SevInfo    = 3
)

type ruleFunc func(*parser.Module, *analysis.SymbolTable) []LintDiagnostic

var rules = []ruleFunc{
	checkUnusedVariable,
	checkEmptyBlock,
	checkUnreachableCode,
	checkMagicNumber,
	checkTooManyParams,
	checkNestedDepth,
	checkSuspiciousAssignment,
	checkMissingReturn,
	checkGlobalVarInProc,
}

func RunAll(mod *parser.Module, st *analysis.SymbolTable) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, rule := range rules {
		diags = append(diags, rule(mod, st)...)
	}
	return diags
}
