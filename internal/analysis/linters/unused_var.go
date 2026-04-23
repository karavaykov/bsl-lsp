package linters

import (
	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func checkUnusedVariable(mod *parser.Module, st *analysis.SymbolTable) []LintDiagnostic {
	used := collectUsedNames(mod)
	var diags []LintDiagnostic
	for _, sym := range st.Symbols {
		if sym.Kind == analysis.SymbolProcedure || sym.Kind == analysis.SymbolFunction {
			continue
		}
		if !used[sym.Name] {
			length := len(sym.Name)
			if length < 1 {
				length = 1
			}
			diags = append(diags, LintDiagnostic{
				Line:     sym.Line,
				Col:      sym.Col,
				Length:   length,
				Message:  sym.Kind.String() + " \"" + sym.Name + "\" объявлена, но не используется",
				Code:     "unused-variable",
				Severity: SevWarning,
			})
		}
	}
	return diags
}

type nameSet map[string]bool

func collectUsedNames(mod *parser.Module) nameSet {
	used := make(nameSet)
	for _, stmt := range mod.Statements {
		collectUsedInNode(stmt, used)
	}
	return used
}

func collectUsedInNode(n parser.Node, used nameSet) {
	switch n := n.(type) {
	case *parser.Procedure:
		for _, p := range n.Params {
			used[p.Name] = true
		}
		collectUsedInBlock(n.Body, used)
	case *parser.Function:
		for _, p := range n.Params {
			used[p.Name] = true
		}
		collectUsedInBlock(n.Body, used)
	case *parser.VarDeclExpr:
	case *parser.BinaryExpr:
		collectUsedInNode(n.Left, used)
		collectUsedInNode(n.Right, used)
	case *parser.UnaryExpr:
		collectUsedInNode(n.Right, used)
	case *parser.TernaryExpr:
		collectUsedInNode(n.Condition, used)
		collectUsedInNode(n.True, used)
		collectUsedInNode(n.False, used)
	case *parser.AssignmentStmt:
		collectUsedInNode(n.Left, used)
		collectUsedInNode(n.Right, used)
	case *parser.CallStmt:
		for _, arg := range n.Args {
			collectUsedInNode(arg, used)
		}
	case *parser.Ident:
		used[n.Name] = true
	case *parser.ReturnStmt:
		if n.Value != nil {
			collectUsedInNode(n.Value, used)
		}
	case *parser.RaiseStmt:
		if n.Value != nil {
			collectUsedInNode(n.Value, used)
		}
	case *parser.IndexExpr:
		collectUsedInNode(n.Object, used)
		collectUsedInNode(n.Index, used)
	case *parser.FieldAccessExpr:
		collectUsedInNode(n.Object, used)
	case *parser.NewExpr:
		for _, arg := range n.Args {
			collectUsedInNode(arg, used)
		}
	case *parser.ExecuteExpr:
		collectUsedInNode(n.Expr, used)
	case *parser.AddressExpr:
		collectUsedInNode(n.Expr, used)
	case *parser.TypeExpr:
		collectUsedInNode(n.Expr, used)
	case *parser.ValExpr:
		collectUsedInNode(n.Expr, used)
	case *parser.IfStmt:
		collectUsedInNode(n.Condition, used)
		collectUsedInBlock(n.Body, used)
		for _, ei := range n.ElseIf {
			collectUsedInNode(ei.Condition, used)
			collectUsedInBlock(ei.Body, used)
		}
		collectUsedInBlock(n.ElseBody, used)
	case *parser.WhileStmt:
		collectUsedInNode(n.Condition, used)
		collectUsedInBlock(n.Body, used)
	case *parser.ForStmt:
		used[n.Var] = true
		collectUsedInBlock(n.Body, used)
	case *parser.ForEachStmt:
		used[n.Var] = true
		collectUsedInBlock(n.Body, used)
	case *parser.TryStmt:
		collectUsedInBlock(n.Body, used)
		collectUsedInBlock(n.Except, used)
	case *parser.RegionBlock:
		collectUsedInBlock(n.Body, used)
	case *parser.HashIfBlock:
		collectUsedInBlock(n.Body, used)
		for _, ei := range n.ElseIf {
			collectUsedInBlock(ei.Body, used)
		}
		collectUsedInBlock(n.ElseBody, used)
	case *parser.GotoStmt:
	case *parser.LabelStmt:
	default:
	}
}

func collectUsedInBlock(stmts []parser.Node, used nameSet) {
	for _, stmt := range stmts {
		collectUsedInNode(stmt, used)
	}
}
