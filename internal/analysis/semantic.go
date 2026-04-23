package analysis

import (
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

type FoldingRange struct {
	StartLine int
	EndLine   int
	Kind      string
}

type CallInfo struct {
	Name        string
	ActiveParam int
}

func CollectSemanticTokens(mod *parser.Module, st *SymbolTable) []int {
	var data []int
	prevLine, prevCol := 0, 0

	emit := func(line, col, length, tokenType int) {
		if line < prevLine || (line == prevLine && col < prevCol) {
			return
		}
		deltaLine := line - prevLine
		var deltaCol int
		if deltaLine == 0 {
			deltaCol = col - prevCol
		} else {
			deltaCol = col
		}
		data = append(data, deltaLine, deltaCol, length, tokenType, 0)
		prevLine = line
		prevCol = col + length
	}

	collectNodeList(mod.Statements, st, emit)
	return data
}

func collectNode(n parser.Node, st *SymbolTable, emit func(line, col, length, tokenType int)) {
	switch n := n.(type) {
	case *parser.Procedure:
		emit(n.Line, n.Col, len(n.Name), 3)
		for _, p := range n.Params {
			collectNode(p, st, emit)
		}
		collectNodeList(n.Body, st, emit)

	case *parser.Function:
		emit(n.Line, n.Col, len(n.Name), 3)
		for _, p := range n.Params {
			collectNode(p, st, emit)
		}
		collectNodeList(n.Body, st, emit)

	case *parser.ParamDecl:
	case *parser.CallStmt:
		emit(n.Line, n.Col, len(n.Function), 0)
		if n.Object != nil {
			line, col := n.Object.Pos()
			emit(line, col, 0, 0)
		}

	case *parser.Ident:
		sym := st.Lookup(n.Name)
		tokType := 0
		if sym != nil {
			switch sym.Kind {
			case SymbolVariable:
				tokType = 2
			case SymbolProcedure:
				tokType = 3
			case SymbolFunction:
				tokType = 4
			case SymbolParameter:
				tokType = 5
			}
		}
		emit(n.Line, n.Col, len(n.Name), tokType)

	case *parser.StringLit:
		emit(n.Line, n.Col, len(n.Value), 6)

	case *parser.NumberLit:
		emit(n.Line, n.Col, len(n.Value), 7)

	case *parser.Comment:
		emit(n.Line, n.Col, len(n.Text), 8)

	case *parser.VarDeclExpr:
		line, col := n.Pos()
		emit(line, col, len(n.Name), 2)

	case *parser.IfStmt:
		collectNodeList(n.Body, st, emit)
		for _, ei := range n.ElseIf {
			collectNodeList(ei.Body, st, emit)
		}
		collectNodeList(n.ElseBody, st, emit)

	case *parser.WhileStmt:
		collectNodeList(n.Body, st, emit)

	case *parser.ForStmt:
		collectNodeList(n.Body, st, emit)

	case *parser.ForEachStmt:
		collectNodeList(n.Body, st, emit)

	case *parser.TryStmt:
		collectNodeList(n.Body, st, emit)
		collectNodeList(n.Except, st, emit)

	case *parser.BinaryExpr:
		collectNode(n.Left, st, emit)
		collectNode(n.Right, st, emit)

	case *parser.UnaryExpr:
		collectNode(n.Right, st, emit)

	case *parser.TernaryExpr:
		collectNode(n.Condition, st, emit)
		collectNode(n.True, st, emit)
		collectNode(n.False, st, emit)

	case *parser.AssignmentStmt:
		collectNode(n.Left, st, emit)
		collectNode(n.Right, st, emit)

	case *parser.IndexExpr:
		collectNode(n.Object, st, emit)
		collectNode(n.Index, st, emit)

	case *parser.FieldAccessExpr:
		collectNode(n.Object, st, emit)

	case *parser.NewExpr:
		collectNodeList(n.Args, st, emit)

	case *parser.RegionBlock:
		collectNodeList(n.Body, st, emit)

	case *parser.HashIfBlock:
		collectNodeList(n.Body, st, emit)
		for _, ei := range n.ElseIf {
			collectNodeList(ei.Body, st, emit)
		}
		collectNodeList(n.ElseBody, st, emit)
	}
}

func collectNodeList(nodes []parser.Node, st *SymbolTable, emit func(line, col, length, tokenType int)) {
	for _, n := range nodes {
		collectNode(n, st, emit)
	}
}

func FindCallAtPos(mod *parser.Module, line, col int) *CallInfo {
	var found *CallInfo
	walkForCall(mod.Statements, line, col, &found)
	return found
}

func walkForCall(nodes []parser.Node, line, col int, found **CallInfo) {
	if *found != nil {
		return
	}
	for _, n := range nodes {
		if *found != nil {
			return
		}
		switch n := n.(type) {
		case *parser.CallStmt:
			if n.ParenLine+1 == line && col > n.ParenCol {
				activeParam := countCommasBefore(n.Args, line, col)
				*found = &CallInfo{Name: n.Function, ActiveParam: activeParam}
				return
			}
		case *parser.IfStmt:
			walkForCall(n.Body, line, col, found)
			for _, ei := range n.ElseIf {
				walkForCall(ei.Body, line, col, found)
			}
			walkForCall(n.ElseBody, line, col, found)
		case *parser.WhileStmt:
			walkForCall(n.Body, line, col, found)
		case *parser.ForStmt:
			walkForCall(n.Body, line, col, found)
		case *parser.ForEachStmt:
			walkForCall(n.Body, line, col, found)
		case *parser.TryStmt:
			walkForCall(n.Body, line, col, found)
			walkForCall(n.Except, line, col, found)
		case *parser.RegionBlock:
			walkForCall(n.Body, line, col, found)
		case *parser.HashIfBlock:
			walkForCall(n.Body, line, col, found)
			for _, ei := range n.ElseIf {
				walkForCall(ei.Body, line, col, found)
			}
			walkForCall(n.ElseBody, line, col, found)
		}
	}
}

func countCommasBefore(args []parser.Node, line, col int) int {
	count := 0
	for _, arg := range args {
		al, ac := arg.Pos()
		if al+1 > line || (al+1 == line && ac+1 >= col) {
			break
		}
		count++
	}
	return count
}

func CollectFoldingRanges(mod *parser.Module, st *SymbolTable) []FoldingRange {
	var ranges []FoldingRange

	for _, sym := range st.Symbols {
		if sym.Scope != st.Global {
			continue
		}
		if sym.Kind != SymbolProcedure && sym.Kind != SymbolFunction {
			continue
		}
		if sym.BodyScope == nil {
			continue
		}
		endLine := findBlockEndLine(mod, sym)
		if endLine > sym.Line {
			ranges = append(ranges, FoldingRange{
				StartLine: sym.Line,
				EndLine:   endLine,
				Kind:      "region",
			})
		}
	}

	return ranges
}

func findBlockEndLine(mod *parser.Module, sym *Symbol) int {
	maxLine := sym.Line
	for _, s := range mod.Statements {
		l, _ := s.Pos()
		if l > maxLine {
			maxLine = l
		}
	}
	return maxLine
}
