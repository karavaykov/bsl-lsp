package linters

import (
	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func checkEmptyBlock(mod *parser.Module, st *analysis.SymbolTable) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range mod.Statements {
		diags = append(diags, checkEmptyInNode(stmt)...)
	}
	return diags
}

func checkEmptyInNode(n parser.Node) []LintDiagnostic {
	switch n := n.(type) {
	case *parser.Procedure:
		if isEmptyBlock(n.Body) {
			return []LintDiagnostic{{
				Line:     n.Line,
				Col:      n.Col,
				Length:   len(n.Name),
				Message:  "Процедура \"" + n.Name + "\" имеет пустое тело",
				Code:     "empty-block",
				Severity: SevWarning,
			}}
		}
		return checkEmptyInBlock(n.Body)
	case *parser.Function:
		if isEmptyBlock(n.Body) {
			return []LintDiagnostic{{
				Line:     n.Line,
				Col:      n.Col,
				Length:   len(n.Name),
				Message:  "Функция \"" + n.Name + "\" имеет пустое тело",
				Code:     "empty-block",
				Severity: SevWarning,
			}}
		}
		return checkEmptyInBlock(n.Body)
	case *parser.IfStmt:
		var diags []LintDiagnostic
		if isEmptyBlock(n.Body) {
			diags = append(diags, LintDiagnostic{
				Line:     n.Line,
				Col:      n.Col,
				Length:   1,
				Message:  "Пустое тело условия Если",
				Code:     "empty-block",
				Severity: SevInfo,
			})
		}
		for _, ei := range n.ElseIf {
			if isEmptyBlock(ei.Body) {
				cl, _ := ei.Condition.Pos()
				diags = append(diags, LintDiagnostic{
					Line:     cl,
					Col:      0,
					Length:   1,
					Message:  "Пустое тело условия ИначеЕсли",
					Code:     "empty-block",
					Severity: SevInfo,
				})
			}
			diags = append(diags, checkEmptyInBlock(ei.Body)...)
		}
		if n.ElseBody != nil && isEmptyBlock(n.ElseBody) {
			diags = append(diags, LintDiagnostic{
				Line:     n.Line,
				Col:      0,
				Length:   1,
				Message:  "Пустое тело блока Иначе",
				Code:     "empty-block",
				Severity: SevInfo,
			})
		}
		if n.ElseBody != nil {
			diags = append(diags, checkEmptyInBlock(n.ElseBody)...)
		}
		return diags
	case *parser.WhileStmt:
		if isEmptyBlock(n.Body) {
			return []LintDiagnostic{{
				Line:     n.Line,
				Col:      n.Col,
				Length:   1,
				Message:  "Пустое тело цикла Пока",
				Code:     "empty-block",
				Severity: SevInfo,
			}}
		}
		return checkEmptyInBlock(n.Body)
	case *parser.ForStmt:
		if isEmptyBlock(n.Body) {
			return []LintDiagnostic{{
				Line:     n.Line,
				Col:      n.Col,
				Length:   1,
				Message:  "Пустое тело цикла Для",
				Code:     "empty-block",
				Severity: SevInfo,
			}}
		}
		return checkEmptyInBlock(n.Body)
	case *parser.ForEachStmt:
		if isEmptyBlock(n.Body) {
			return []LintDiagnostic{{
				Line:     n.Line,
				Col:      n.Col,
				Length:   1,
				Message:  "Пустое тело цикла Для Каждого",
				Code:     "empty-block",
				Severity: SevInfo,
			}}
		}
		return checkEmptyInBlock(n.Body)
	case *parser.TryStmt:
		var diags []LintDiagnostic
		if isEmptyBlock(n.Body) {
			diags = append(diags, LintDiagnostic{
				Line:     n.Line,
				Col:      n.Col,
				Length:   1,
				Message:  "Пустое тело Попытка",
				Code:     "empty-block",
				Severity: SevWarning,
			})
		}
		diags = append(diags, checkEmptyInBlock(n.Body)...)
		diags = append(diags, checkEmptyInBlock(n.Except)...)
		return diags
	case *parser.RegionBlock:
		return checkEmptyInBlock(n.Body)
	case *parser.HashIfBlock:
		var diags []LintDiagnostic
		diags = append(diags, checkEmptyInBlock(n.Body)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkEmptyInBlock(ei.Body)...)
		}
		diags = append(diags, checkEmptyInBlock(n.ElseBody)...)
		return diags
	default:
		return nil
	}
}

func checkEmptyInBlock(stmts []parser.Node) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range stmts {
		diags = append(diags, checkEmptyInNode(stmt)...)
	}
	return diags
}

func isEmptyBlock(stmts []parser.Node) bool {
	for _, s := range stmts {
		switch s.(type) {
		case *parser.Comment:
			continue
		default:
			return false
		}
	}
	return true
}
