package analysis

import (
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

type SymbolKind int

const (
	SymbolVariable  SymbolKind = iota
	SymbolProcedure
	SymbolFunction
	SymbolParameter
)

type Symbol struct {
	Name      string
	Kind      SymbolKind
	Scope     *Scope
	Line      int
	Col       int
	BodyScope *Scope
	Export    bool
}

type Scope struct {
	Parent  *Scope
	symbols map[string]*Symbol
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent:  parent,
		symbols: make(map[string]*Symbol),
	}
}

func (s *Scope) Add(sym *Symbol) {
	sym.Scope = s
	s.symbols[sym.Name] = sym
}

func (s *Scope) Lookup(name string) *Symbol {
	if sym, ok := s.symbols[name]; ok {
		return sym
	}
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return nil
}

func (s *Scope) Symbols() map[string]*Symbol {
	return s.symbols
}

type SymbolTable struct {
	Global  *Scope
	Symbols []*Symbol
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		Global:  NewScope(nil),
		Symbols: make([]*Symbol, 0),
	}
}

func (st *SymbolTable) EnterScope(parent *Scope) *Scope {
	return NewScope(parent)
}

func (st *SymbolTable) Lookup(name string) *Symbol {
	for _, sym := range st.Symbols {
		if sym.Name == name {
			return sym
		}
	}
	return nil
}

func (st *SymbolTable) add(sym *Symbol) {
	st.Symbols = append(st.Symbols, sym)
}

type visitor struct {
	table *SymbolTable
}

func BuildSymbolTable(mod *parser.Module) *SymbolTable {
	t := NewSymbolTable()
	v := &visitor{table: t}
	v.visitModule(mod)
	return t
}

func (v *visitor) addSymbol(name string, kind SymbolKind, scope *Scope, line, col int) *Symbol {
	sym := &Symbol{
		Name: name,
		Kind: kind,
		Line: line,
		Col:  col,
	}
	scope.Add(sym)
	v.table.add(sym)
	return sym
}

func (v *visitor) visitModule(mod *parser.Module) {
	for _, stmt := range mod.Statements {
		v.visitModuleStatement(stmt)
	}
}

func (v *visitor) visitModuleStatement(n parser.Node) {
	switch n := n.(type) {
	case *parser.Procedure:
		line, col := n.Pos()
		bodyScope := v.table.EnterScope(v.table.Global)
		sym := v.addSymbol(n.Name, SymbolProcedure, v.table.Global, line, col)
		sym.BodyScope = bodyScope
		sym.Export = n.Export
		for _, p := range n.Params {
			v.addSymbol(p.Name, SymbolParameter, bodyScope, line, col)
		}
		v.visitBlock(n.Body, bodyScope)

	case *parser.Function:
		line, col := n.Pos()
		bodyScope := v.table.EnterScope(v.table.Global)
		sym := v.addSymbol(n.Name, SymbolFunction, v.table.Global, line, col)
		sym.BodyScope = bodyScope
		sym.Export = n.Export
		for _, p := range n.Params {
			v.addSymbol(p.Name, SymbolParameter, bodyScope, line, col)
		}
		v.visitBlock(n.Body, bodyScope)

	case *parser.VarDeclExpr:
		v.addSymbol(n.Name, SymbolVariable, v.table.Global, n.Line, n.Col)

	case *parser.RegionBlock:
		v.visitBlock(n.Body, v.table.Global)

	case *parser.HashIfBlock:
		v.visitBlock(n.Body, v.table.Global)
		for _, ei := range n.ElseIf {
			v.visitBlock(ei.Body, v.table.Global)
		}
		v.visitBlock(n.ElseBody, v.table.Global)
	}
}

func (v *visitor) visitBlock(stmts []parser.Node, parentScope *Scope) {
	for _, stmt := range stmts {
		v.visitStatement(stmt, parentScope)
	}
}

func (v *visitor) visitStatement(n parser.Node, scope *Scope) {
	switch n := n.(type) {
	case *parser.VarDeclExpr:
		v.addSymbol(n.Name, SymbolVariable, scope, n.Line, n.Col)

	case *parser.IfStmt:
		bodyScope := v.table.EnterScope(scope)
		v.visitBlock(n.Body, bodyScope)
		for _, ei := range n.ElseIf {
			eiScope := v.table.EnterScope(scope)
			v.visitBlock(ei.Body, eiScope)
		}
		elseScope := v.table.EnterScope(scope)
		v.visitBlock(n.ElseBody, elseScope)

	case *parser.WhileStmt:
		bodyScope := v.table.EnterScope(scope)
		v.visitBlock(n.Body, bodyScope)

	case *parser.ForStmt:
		bodyScope := v.table.EnterScope(scope)
		line, col := n.Pos()
		v.addSymbol(n.Var, SymbolVariable, bodyScope, line, col)
		v.visitBlock(n.Body, bodyScope)

	case *parser.ForEachStmt:
		bodyScope := v.table.EnterScope(scope)
		line, col := n.Pos()
		v.addSymbol(n.Var, SymbolVariable, bodyScope, line, col)
		v.visitBlock(n.Body, bodyScope)

	case *parser.TryStmt:
		bodyScope := v.table.EnterScope(scope)
		v.visitBlock(n.Body, bodyScope)
		exceptScope := v.table.EnterScope(scope)
		v.visitBlock(n.Except, exceptScope)

	case *parser.AssignmentStmt:
		v.visitExprForDecl(n.Left, scope)
	}
}

func (v *visitor) visitExprForDecl(n parser.Node, scope *Scope) {
	switch n := n.(type) {
	case *parser.Ident:
		if scope.Lookup(n.Name) == nil {
			v.addSymbol(n.Name, SymbolVariable, scope, n.Line, n.Col)
		}
	}
}

func (k SymbolKind) String() string {
	switch k {
	case SymbolVariable:
		return "Переменная"
	case SymbolProcedure:
		return "Процедура"
	case SymbolFunction:
		return "Функция"
	case SymbolParameter:
		return "Параметр"
	default:
		return "Символ"
	}
}
