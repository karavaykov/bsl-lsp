package linters

import (
	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func checkTooManyParams(mod *parser.Module, st *analysis.SymbolTable) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range mod.Statements {
		diags = append(diags, checkParamsInNode(stmt)...)
	}
	return diags
}

func checkParamsInNode(n parser.Node) []LintDiagnostic {
	switch n := n.(type) {
	case *parser.Procedure:
		if len(n.Params) > 7 {
			return []LintDiagnostic{{
				Line:     n.Line,
				Col:      n.Col,
				Length:   len(n.Name),
				Message:  "Процедура \"" + n.Name + "\" имеет " + itoa(len(n.Params)) + " параметров (максимум 7)",
				Code:     "too-many-params",
				Severity: SevWarning,
			}}
		}
	case *parser.Function:
		if len(n.Params) > 7 {
			return []LintDiagnostic{{
				Line:     n.Line,
				Col:      n.Col,
				Length:   len(n.Name),
				Message:  "Функция \"" + n.Name + "\" имеет " + itoa(len(n.Params)) + " параметров (максимум 7)",
				Code:     "too-many-params",
				Severity: SevWarning,
			}}
		}
	case *parser.IfStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkParamsInBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkParamsInBlock(ei.Body)...)
		}
		diags = append(diags, checkParamsInBlock(n.ElseBody)...)
		return diags
	case *parser.WhileStmt:
		return checkParamsInBlock(n.Body)
	case *parser.ForStmt:
		return checkParamsInBlock(n.Body)
	case *parser.ForEachStmt:
		return checkParamsInBlock(n.Body)
	case *parser.TryStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkParamsInBlock(n.Body)...)
		diags = append(diags, checkParamsInBlock(n.Except)...)
		return diags
	case *parser.RegionBlock:
		return checkParamsInBlock(n.Body)
	case *parser.HashIfBlock:
		var diags []LintDiagnostic
		diags = append(diags, checkParamsInBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkParamsInBlock(ei.Body)...)
		}
		diags = append(diags, checkParamsInBlock(n.ElseBody)...)
		return diags
	}
	return nil
}

func checkParamsInBlock(stmts []parser.Node) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range stmts {
		diags = append(diags, checkParamsInNode(stmt)...)
	}
	return diags
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
