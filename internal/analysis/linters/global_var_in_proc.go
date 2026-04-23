package linters

import (
	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func checkGlobalVarInProc(mod *parser.Module, st *analysis.SymbolTable) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range mod.Statements {
		diags = append(diags, checkGlobalVarInNode(stmt, st)...)
	}
	return diags
}

func checkGlobalVarInNode(n parser.Node, st *analysis.SymbolTable) []LintDiagnostic {
	switch n := n.(type) {
	case *parser.Procedure:
		return checkGlobalVarInBlock(n.Body, st, true)
	case *parser.Function:
		return checkGlobalVarInBlock(n.Body, st, true)
	case *parser.IfStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkGlobalVarInBlock(n.Body, st, false)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkGlobalVarInBlock(ei.Body, st, false)...)
		}
		diags = append(diags, checkGlobalVarInBlock(n.ElseBody, st, false)...)
		return diags
	case *parser.WhileStmt:
		return checkGlobalVarInBlock(n.Body, st, false)
	case *parser.ForStmt:
		return checkGlobalVarInBlock(n.Body, st, false)
	case *parser.ForEachStmt:
		return checkGlobalVarInBlock(n.Body, st, false)
	case *parser.TryStmt:
		var diags []LintDiagnostic
		diags = append(diags, checkGlobalVarInBlock(n.Body, st, false)...)
		diags = append(diags, checkGlobalVarInBlock(n.Except, st, false)...)
		return diags
	case *parser.RegionBlock:
		return checkGlobalVarInBlock(n.Body, st, false)
	case *parser.HashIfBlock:
		var diags []LintDiagnostic
		diags = append(diags, checkGlobalVarInBlock(n.Body, st, false)...)
		for _, ei := range n.ElseIf {
			diags = append(diags, checkGlobalVarInBlock(ei.Body, st, false)...)
		}
		diags = append(diags, checkGlobalVarInBlock(n.ElseBody, st, false)...)
		return diags
	default:
		return nil
	}
}

func checkGlobalVarInBlock(stmts []parser.Node, st *analysis.SymbolTable, insideProc bool) []LintDiagnostic {
	var diags []LintDiagnostic
	for _, stmt := range stmts {
		if insideProc {
			if as, ok := stmt.(*parser.AssignmentStmt); ok {
				if ident, ok2 := as.Left.(*parser.Ident); ok2 {
					sym := st.Global.Lookup(ident.Name)
					if sym != nil && sym.Kind == analysis.SymbolVariable && sym.Scope == st.Global {
						diags = append(diags, LintDiagnostic{
							Line:     ident.Line,
							Col:      ident.Col,
							Length:   len(ident.Name),
							Message:  "Присваивание глобальной переменной \"" + ident.Name + "\" внутри процедуры/функции",
							Code:     "global-var-in-proc",
							Severity: SevInfo,
						})
					}
				}
			}
		}
		diags = append(diags, checkGlobalVarInNode(stmt, st)...)
	}
	return diags
}
