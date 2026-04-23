package linters

import (
	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func checkSuspiciousAssignment(mod *parser.Module, st *analysis.SymbolTable) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range mod.Statements {
		diags = append(diags, checkSuspiciousInNode(stmt)...)
	}
	return diags
}

func checkSuspiciousInNode(n parser.Node) []LintDiagnostic {
	switch n := n.(type) {
	case *parser.AssignmentStmt:
		if isSelfAssign(n.Left, n.Right) {
			leftName := nameOfIdent(n.Left)
			l, c := n.Left.Pos()
			return []LintDiagnostic{{
				Line:     l,
				Col:      c,
				Length:   len(leftName),
				Message:  "Присваивание переменной \"" + leftName + "\" самой себе",
				Code:     "suspicious-assignment",
				Severity: SevWarning,
			}}
		}
		var diags []LintDiagnostic
		diags = append(diags, checkSuspiciousInNode(n.Right)...)
		return diags
	case *parser.IfStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkSuspiciousInBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkSuspiciousInBlock(ei.Body)...)
		}
		diags = append(diags, checkSuspiciousInBlock(n.ElseBody)...)
		return diags
	case *parser.WhileStmt:
		return checkSuspiciousInBlock(n.Body)
	case *parser.ForStmt:
		return checkSuspiciousInBlock(n.Body)
	case *parser.ForEachStmt:
		return checkSuspiciousInBlock(n.Body)
	case *parser.TryStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkSuspiciousInBlock(n.Body)...)
		diags = append(diags, checkSuspiciousInBlock(n.Except)...)
		return diags
	case *parser.Procedure:
		return checkSuspiciousInBlock(n.Body)
	case *parser.Function:
		return checkSuspiciousInBlock(n.Body)
	case *parser.RegionBlock:
		return checkSuspiciousInBlock(n.Body)
	case *parser.HashIfBlock:
		var diags []LintDiagnostic
		diags = append(diags, checkSuspiciousInBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkSuspiciousInBlock(ei.Body)...)
		}
		diags = append(diags, checkSuspiciousInBlock(n.ElseBody)...)
		return diags
	default:
		return nil
	}
}

func checkSuspiciousInBlock(stmts []parser.Node) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range stmts {
		diags = append(diags, checkSuspiciousInNode(stmt)...)
	}
	return diags
}

func isSelfAssign(left, right parser.Node) bool {
	leftName := nameOfIdent(left)
	if leftName == "" {
		return false
	}
	rightName := nameOfIdent(right)
	if rightName == "" {
		return false
	}
	return leftName == rightName
}

func nameOfIdent(n parser.Node) string {
	switch n := n.(type) {
	case *parser.Ident:
		return n.Name
	case *parser.FieldAccessExpr:
		return nameOfIdent(n.Object) + "." + n.Field
	}
	return ""
}
