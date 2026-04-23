package linters

import (
	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func checkNestedDepth(mod *parser.Module, st *analysis.SymbolTable) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range mod.Statements {
		diags = append(diags, checkDepthInNode(stmt, 0)...)
	}
	return diags
}

func checkDepthInNode(n parser.Node, depth int) []LintDiagnostic {
	if depth > 5 {
		l, c := n.Pos()
		return []LintDiagnostic{{
			Line:     l,
			Col:      c,
			Length:   1,
			Message:  "Глубина вложенности превышает 5 уровней",
			Code:     "nested-depth",
			Severity: SevWarning,
		}}
	}
	switch n := n.(type) {
	case *parser.Procedure:
		return checkDepthInBlock(n.Body, depth)
	case *parser.Function:
		return checkDepthInBlock(n.Body, depth)
	case *parser.IfStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkDepthInBlock(n.Body, depth+1)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkDepthInBlock(ei.Body, depth+1)...)
		}
		diags = append(diags, checkDepthInBlock(n.ElseBody, depth+1)...)
		return diags
	case *parser.WhileStmt:
		return checkDepthInBlock(n.Body, depth+1)
	case *parser.ForStmt:
		return checkDepthInBlock(n.Body, depth+1)
	case *parser.ForEachStmt:
		return checkDepthInBlock(n.Body, depth+1)
	case *parser.TryStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkDepthInBlock(n.Body, depth+1)...)
		diags = append(diags, checkDepthInBlock(n.Except, depth+1)...)
		return diags
	case *parser.RegionBlock:
		return checkDepthInBlock(n.Body, depth)
	case *parser.HashIfBlock:
		var diags []LintDiagnostic
		diags = append(diags, checkDepthInBlock(n.Body, depth)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkDepthInBlock(ei.Body, depth)...)
		}
		diags = append(diags, checkDepthInBlock(n.ElseBody, depth)...)
		return diags
	default:
		return nil
	}
}

func checkDepthInBlock(stmts []parser.Node, depth int) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range stmts {
		diags = append(diags, checkDepthInNode(stmt, depth)...)
	}
	return diags
}
