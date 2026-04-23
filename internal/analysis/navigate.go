package analysis

import (
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func FindIdentAtPos(mod *parser.Module, line, col int) *parser.Ident {
	for _, stmt := range mod.Statements {
		if ident := findIdentInNode(stmt, line, col); ident != nil {
			return ident
		}
	}
	return nil
}

func findIdentInNode(n parser.Node, line, col int) *parser.Ident {
	switch n := n.(type) {
	case *parser.Ident:
		if n.Line == line && col >= n.Col && col < n.Col+len(n.Name) {
			return n
		}
	case *parser.CallStmt:
		if n.Object != nil {
			if ident := findIdentInNode(n.Object, line, col); ident != nil {
				return ident
			}
		}
		for _, arg := range n.Args {
			if ident := findIdentInNode(arg, line, col); ident != nil {
				return ident
			}
		}
	case *parser.BinaryExpr:
		if ident := findIdentInNode(n.Left, line, col); ident != nil {
			return ident
		}
		if ident := findIdentInNode(n.Right, line, col); ident != nil {
			return ident
		}
	case *parser.UnaryExpr:
		return findIdentInNode(n.Right, line, col)
	case *parser.TernaryExpr:
		if ident := findIdentInNode(n.Condition, line, col); ident != nil {
			return ident
		}
		if ident := findIdentInNode(n.True, line, col); ident != nil {
			return ident
		}
		return findIdentInNode(n.False, line, col)
	case *parser.AssignmentStmt:
		if ident := findIdentInNode(n.Left, line, col); ident != nil {
			return ident
		}
		return findIdentInNode(n.Right, line, col)
	case *parser.ReturnStmt:
		if n.Value != nil {
			return findIdentInNode(n.Value, line, col)
		}
	case *parser.RaiseStmt:
		if n.Value != nil {
			return findIdentInNode(n.Value, line, col)
		}
	case *parser.IndexExpr:
		if ident := findIdentInNode(n.Object, line, col); ident != nil {
			return ident
		}
		return findIdentInNode(n.Index, line, col)
	case *parser.FieldAccessExpr:
		return findIdentInNode(n.Object, line, col)
	case *parser.NewExpr:
		return nil
	case *parser.ExecuteExpr:
		return findIdentInNode(n.Expr, line, col)
	case *parser.AddressExpr:
		return findIdentInNode(n.Expr, line, col)
	case *parser.TypeExpr:
		return findIdentInNode(n.Expr, line, col)
	case *parser.ValExpr:
		return findIdentInNode(n.Expr, line, col)
	case *parser.Procedure:
		return findIdentInBlock(n.Body, line, col)
	case *parser.Function:
		return findIdentInBlock(n.Body, line, col)
	case *parser.IfStmt:
		if ident := findIdentInNode(n.Condition, line, col); ident != nil {
			return ident
		}
		if ident := findIdentInBlock(n.Body, line, col); ident != nil {
			return ident
		}
		for _, ei := range n.ElseIf {
			if ident := findIdentInNode(ei.Condition, line, col); ident != nil {
				return ident
			}
			if ident := findIdentInBlock(ei.Body, line, col); ident != nil {
				return ident
			}
		}
		return findIdentInBlock(n.ElseBody, line, col)
	case *parser.WhileStmt:
		if ident := findIdentInNode(n.Condition, line, col); ident != nil {
			return ident
		}
		return findIdentInBlock(n.Body, line, col)
	case *parser.ForStmt:
		if ident := findIdentInBlock(n.Body, line, col); ident != nil {
			return ident
		}
	case *parser.ForEachStmt:
		if ident := findIdentInNode(n.In, line, col); ident != nil {
			return ident
		}
		return findIdentInBlock(n.Body, line, col)
	case *parser.TryStmt:
		if ident := findIdentInBlock(n.Body, line, col); ident != nil {
			return ident
		}
		return findIdentInBlock(n.Except, line, col)
	case *parser.VarDeclExpr:
		return nil
	case *parser.GotoStmt:
		return nil
	case *parser.RegionBlock:
		return findIdentInBlock(n.Body, line, col)
	case *parser.HashIfBlock:
		if ident := findIdentInBlock(n.Body, line, col); ident != nil {
			return ident
		}
		for _, ei := range n.ElseIf {
			if ident := findIdentInBlock(ei.Body, line, col); ident != nil {
				return ident
			}
		}
		return findIdentInBlock(n.ElseBody, line, col)
	}
	return nil
}

func findIdentInBlock(stmts []parser.Node, line, col int) *parser.Ident {
	for _, stmt := range stmts {
		if ident := findIdentInNode(stmt, line, col); ident != nil {
			return ident
		}
	}
	return nil
}

func FindSymbolAtPos(table *SymbolTable, line, col int) *Symbol {
	for _, sym := range table.Symbols {
		if sym.Line == line && col >= sym.Col && col < sym.Col+len(sym.Name) {
			return sym
		}
	}
	return nil
}

func FindDefinition(table *SymbolTable, name string) *Symbol {
	return table.Lookup(name)
}
