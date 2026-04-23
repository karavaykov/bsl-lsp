package linters

import (
	"strconv"

	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func checkMagicNumber(mod *parser.Module, st *analysis.SymbolTable) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range mod.Statements {
		diags = append(diags, checkMagicInNode(stmt, false)...)
	}
	return diags
}

func checkMagicInNode(n parser.Node, inAssignment bool) []LintDiagnostic {
	switch n := n.(type) {
	case *parser.NumberLit:
		if isMagicNumber(n.Value) {
			return []LintDiagnostic{{
				Line:     n.Line,
				Col:      n.Col,
				Length:   len(n.Value),
				Message:  "Магическое число " + n.Value + ". Рекомендуется вынести в константу",
				Code:     "magic-number",
				Severity: SevInfo,
			}}
		}
	case *parser.BinaryExpr:
		var diags []LintDiagnostic
		diags = append(diags, checkMagicInNode(n.Left, false)...)
		diags = append(diags, checkMagicInNode(n.Right, false)...)
		return diags
	case *parser.UnaryExpr:
		return checkMagicInNode(n.Right, false)
	case *parser.AssignmentStmt:
		return checkMagicInNode(n.Right, true)
	case *parser.CallStmt:
		var diags []LintDiagnostic
		for _, arg := range n.Args {
			diags = append(diags, checkMagicInNode(arg, false)...)
		}
		return diags
	case *parser.ReturnStmt:
		if n.Value != nil {
			return checkMagicInNode(n.Value, false)
		}
	case *parser.IfStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkMagicInNode(n.Condition, false)...)
		diags = append(diags, checkMagicInBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkMagicInNode(ei.Condition, false)...)
			diags = append(diags, checkMagicInBlock(ei.Body)...)
		}
		diags = append(diags, checkMagicInBlock(n.ElseBody)...)
		return diags
	case *parser.WhileStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkMagicInNode(n.Condition, false)...)
		diags = append(diags, checkMagicInBlock(n.Body)...)
		return diags
	case *parser.ForStmt:
		var diags []LintDiagnostic
		if n.From != nil {
			diags = append(diags, checkMagicInNode(n.From, false)...)
		}
		if n.To != nil {
			diags = append(diags, checkMagicInNode(n.To, false)...)
		}
		diags = append(diags, checkMagicInBlock(n.Body)...)
		return diags
	case *parser.Procedure:
		return checkMagicInBlock(n.Body)
	case *parser.Function:
		return checkMagicInBlock(n.Body)
	case *parser.TryStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkMagicInBlock(n.Body)...)
		diags = append(diags, checkMagicInBlock(n.Except)...)
		return diags
	case *parser.RegionBlock:
		return checkMagicInBlock(n.Body)
	case *parser.HashIfBlock:
		var diags []LintDiagnostic
		diags = append(diags, checkMagicInBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkMagicInBlock(ei.Body)...)
		}
		diags = append(diags, checkMagicInBlock(n.ElseBody)...)
		return diags
	default:
		return nil
	}
	return nil
}

func checkMagicInBlock(stmts []parser.Node) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range stmts {
		diags = append(diags, checkMagicInNode(stmt, false)...)
	}
	return diags
}

func isMagicNumber(val string) bool {
	n, err := strconv.Atoi(val)
	if err != nil {
		n64, err2 := strconv.ParseInt(val, 10, 64)
		if err2 != nil {
			return false
		}
		n = int(n64)
	}
	return n > 3
}
