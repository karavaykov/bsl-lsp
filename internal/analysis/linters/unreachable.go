package linters

import (
	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func checkUnreachableCode(mod *parser.Module, st *analysis.SymbolTable) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range mod.Statements {
		diags = append(diags, checkUnreachableInNode(stmt)...)
	}
	return diags
}

func checkUnreachableInNode(n parser.Node) []LintDiagnostic {
	switch n := n.(type) {
	case *parser.Procedure:
		return checkUnreachableBlock(n.Body)
	case *parser.Function:
		return checkUnreachableBlock(n.Body)
	case *parser.IfStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkUnreachableBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkUnreachableBlock(ei.Body)...)
		}
		diags = append(diags, checkUnreachableBlock(n.ElseBody)...)
		return diags
	case *parser.WhileStmt:
		return checkUnreachableBlock(n.Body)
	case *parser.ForStmt:
		return checkUnreachableBlock(n.Body)
	case *parser.ForEachStmt:
		return checkUnreachableBlock(n.Body)
	case *parser.TryStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkUnreachableBlock(n.Body)...)
		diags = append(diags, checkUnreachableBlock(n.Except)...)
		return diags
	case *parser.RegionBlock:
		return checkUnreachableBlock(n.Body)
	case *parser.HashIfBlock:
		var diags []LintDiagnostic
		diags = append(diags, checkUnreachableBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkUnreachableBlock(ei.Body)...)
		}
		diags = append(diags, checkUnreachableBlock(n.ElseBody)...)
		return diags
	default:
		return nil
	}
}

func checkUnreachableBlock(stmts []parser.Node) []LintDiagnostic {
	var diags []LintDiagnostic
	for i := 0; i < len(stmts); i++ {
		stmt := stmts[i]
		if isTerminalStmt(stmt) {
			if i+1 < len(stmts) {
				next := stmts[i+1]
				if _, ok := next.(*parser.LabelStmt); ok {
					continue
				}
				l, c := next.Pos()
				diags = append(diags, LintDiagnostic{
					Line:     l,
					Col:      c,
					Length:   1,
					Message:  "Недостижимый код после " + terminalName(stmt),
					Code:     "unreachable-code",
					Severity: SevWarning,
				})
			}
		}
		diags = append(diags, checkUnreachableInNode(stmt)...)
	}
	return diags
}

func isTerminalStmt(n parser.Node) bool {
	switch n.(type) {
	case *parser.ReturnStmt, *parser.RaiseStmt, *parser.BreakStmt, *parser.CycleStmt, *parser.GotoStmt:
		return true
	}
	return false
}

func terminalName(n parser.Node) string {
	switch n.(type) {
	case *parser.ReturnStmt:
		return "Возврат"
	case *parser.RaiseStmt:
		return "ВызватьИсключение"
	case *parser.BreakStmt:
		return "Прервать"
	case *parser.CycleStmt:
		return "Продолжить"
	case *parser.GotoStmt:
		return "Перейти"
	}
	return ""
}
