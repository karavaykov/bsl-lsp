package linters

import (
	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func checkMissingReturn(mod *parser.Module, st *analysis.SymbolTable) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range mod.Statements {
		diags = append(diags, checkMissingReturnInNode(stmt)...)
	}
	return diags
}

func checkMissingReturnInNode(n parser.Node) []LintDiagnostic {
	switch n := n.(type) {
	case *parser.Function:
		if !hasReturnInAllPaths(n.Body) {
			return []LintDiagnostic{{
				Line:     n.Line,
				Col:      n.Col,
				Length:   len(n.Name),
				Message:  "Функция \"" + n.Name + "\" — не во всех ветках есть Возврат",
				Code:     "missing-return",
				Severity: SevWarning,
			}}
		}
		return checkMissingReturnInBlock(n.Body)
	case *parser.Procedure:
		return checkMissingReturnInBlock(n.Body)
	case *parser.IfStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkMissingReturnInBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkMissingReturnInBlock(ei.Body)...)
		}
		diags = append(diags, checkMissingReturnInBlock(n.ElseBody)...)
		return diags
	case *parser.WhileStmt:
		return checkMissingReturnInBlock(n.Body)
	case *parser.ForStmt:
		return checkMissingReturnInBlock(n.Body)
	case *parser.ForEachStmt:
		return checkMissingReturnInBlock(n.Body)
	case *parser.TryStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkMissingReturnInBlock(n.Body)...)
		diags = append(diags, checkMissingReturnInBlock(n.Except)...)
		return diags
	case *parser.RegionBlock:
		return checkMissingReturnInBlock(n.Body)
	case *parser.HashIfBlock:
		var diags []LintDiagnostic
		diags = append(diags, checkMissingReturnInBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkMissingReturnInBlock(ei.Body)...)
		}
		diags = append(diags, checkMissingReturnInBlock(n.ElseBody)...)
		return diags
	default:
		return nil
	}
}

func checkMissingReturnInBlock(stmts []parser.Node) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range stmts {
		diags = append(diags, checkMissingReturnInNode(stmt)...)
	}
	return diags
}

func hasReturnInAllPaths(stmts []parser.Node) bool {
	for i, stmt := range stmts {
		switch s := stmt.(type) {
		case *parser.ReturnStmt:
			return true
		case *parser.RaiseStmt:
			return true
		case *parser.IfStmt:
			hasIfBody := hasReturnInAllPaths(s.Body)
			hasElseBody := hasReturnInAllPaths(s.ElseBody)
			for _, ei := range s.ElseIf {
				if !hasReturnInAllPaths(ei.Body) {
					hasElseBody = false
				}
			}
			if hasIfBody && hasElseBody {
				if i+1 < len(stmts) {
					if _, ok := stmts[i+1].(*parser.LabelStmt); ok {
						continue
					}
				}
				return true
			}
			return false
		case *parser.TryStmt:
			if hasReturnInAllPaths(s.Body) || hasReturnInAllPaths(s.Except) {
				if i+1 < len(stmts) {
					if _, ok := stmts[i+1].(*parser.LabelStmt); ok {
						continue
					}
				}
				return true
			}
			return false
		case *parser.WhileStmt, *parser.ForStmt, *parser.ForEachStmt:
			return false
		}
	}
	return false
}
