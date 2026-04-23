package parser

type Node interface {
	nodeMarker()
	Pos() (line, col int)
}

type (
	Module struct {
		Directives  []Node
		Statements  []Node
	}

	Procedure struct {
		Directives   []Node
		Name         string
		Params       []*ParamDecl
		Export       bool
		Body         []Node
		Line         int
		Col          int
	}

	Function struct {
		Directives   []Node
		Name         string
		Params       []*ParamDecl
		Export       bool
		Body         []Node
		Line         int
		Col          int
	}

	ParamDecl struct {
		Name  string
		ByVal bool
	}

	IfStmt struct {
		Condition Node
		Body      []Node
		ElseIf    []*ElseIfBranch
		ElseBody  []Node
		Line      int
		Col       int
	}

	ElseIfBranch struct {
		Condition Node
		Body      []Node
	}

	WhileStmt struct {
		Condition Node
		Body      []Node
		Line      int
		Col       int
	}

	ForStmt struct {
		Var   string
		From  Node
		To    Node
		Body  []Node
		Line  int
		Col   int
	}

	ForEachStmt struct {
		Var  string
		In   Node
		Body []Node
		Line int
		Col  int
	}

	TryStmt struct {
		Body    []Node
		Except  []Node
		Line    int
		Col     int
	}

	ReturnStmt struct {
		Value Node
	}

	RaiseStmt struct {
		Value Node
	}

	CycleStmt struct{}
	BreakStmt struct{}

	GotoStmt struct {
		Label string
	}

	VarDeclExpr struct {
		Name string
		Line int
		Col  int
	}

	AssignmentStmt struct {
		Left  Node
		Right Node
	}

	CallStmt struct {
		Function string
		Object   Node
		Args     []Node
	}

	BinaryExpr struct {
		Left  Node
		Op    TokenType
		Right Node
	}

	UnaryExpr struct {
		Op   TokenType
		Right Node
	}

	TernaryExpr struct {
		Condition Node
		True      Node
		False     Node
	}

	Ident struct {
		Name string
		Line int
		Col  int
	}

	NumberLit struct {
		Value string
		Line  int
		Col   int
	}

	StringLit struct {
		Value string
		Line  int
		Col   int
	}

	DateLit struct {
		Value string
		Line  int
		Col   int
	}

	BoolLit struct {
		Value bool
		Line  int
		Col   int
	}

	UndefinedLit struct {
		Line int
		Col  int
	}

	NullLit struct {
		Line int
		Col  int
	}

	IndexExpr struct {
		Object Node
		Index  Node
	}

	FieldAccessExpr struct {
		Object Node
		Field  string
	}

	NewExpr struct {
		TypeName string
		Args     []Node
	}

	ExecuteExpr struct {
		Expr Node
	}

	AddressExpr struct {
		Expr Node
	}

	TypeExpr struct {
		Expr Node
	}

	ValExpr struct {
		Expr Node
	}

	HashIfBlock struct {
		Condition string
		Body      []Node
		ElseIf    []*HashElseIfBranch
		ElseBody  []Node
	}

	HashElseIfBranch struct {
		Condition string
		Body      []Node
	}

	RegionBlock struct {
		Name string
		Body []Node
	}

	CompilerDirective struct {
		Name string
	}

	Comment struct {
		Text string
		Line int
		Col  int
	}
)

func (m *Module) Pos() (int, int)           { return 0, 0 }
func (p *Procedure) Pos() (int, int)         { return p.Line, p.Col }
func (f *Function) Pos() (int, int)          { return f.Line, f.Col }
func (i *IfStmt) Pos() (int, int)            { return i.Line, i.Col }
func (e *ElseIfBranch) Pos() (int, int)      { return e.Condition.Pos() }
func (w *WhileStmt) Pos() (int, int)         { return w.Line, w.Col }
func (f *ForStmt) Pos() (int, int)           { return f.Line, f.Col }
func (f *ForEachStmt) Pos() (int, int)       { return f.Line, f.Col }
func (t *TryStmt) Pos() (int, int)           { return t.Line, t.Col }
func (r *ReturnStmt) Pos() (int, int)        { return 0, 0 }
func (r *RaiseStmt) Pos() (int, int)         { return 0, 0 }
func (c *CycleStmt) Pos() (int, int)         { return 0, 0 }
func (b *BreakStmt) Pos() (int, int)         { return 0, 0 }
func (g *GotoStmt) Pos() (int, int)          { return 0, 0 }
func (v *VarDeclExpr) Pos() (int, int)       { return v.Line, v.Col }
func (a *AssignmentStmt) Pos() (int, int)    { return a.Left.Pos() }
func (c *CallStmt) Pos() (int, int)          { return 0, 0 }
func (b *BinaryExpr) Pos() (int, int)        { return b.Left.Pos() }
func (u *UnaryExpr) Pos() (int, int)         { return 0, 0 }
func (t *TernaryExpr) Pos() (int, int)       { return t.Condition.Pos() }
func (i *Ident) Pos() (int, int)             { return i.Line, i.Col }
func (n *NumberLit) Pos() (int, int)         { return n.Line, n.Col }
func (s *StringLit) Pos() (int, int)         { return s.Line, s.Col }
func (d *DateLit) Pos() (int, int)           { return d.Line, d.Col }
func (b *BoolLit) Pos() (int, int)           { return b.Line, b.Col }
func (u *UndefinedLit) Pos() (int, int)      { return u.Line, u.Col }
func (n *NullLit) Pos() (int, int)           { return n.Line, n.Col }
func (i *IndexExpr) Pos() (int, int)         { return i.Object.Pos() }
func (f *FieldAccessExpr) Pos() (int, int)   { return f.Object.Pos() }
func (n *NewExpr) Pos() (int, int)           { return 0, 0 }
func (e *ExecuteExpr) Pos() (int, int)       { return e.Expr.Pos() }
func (a *AddressExpr) Pos() (int, int)       { return a.Expr.Pos() }
func (t *TypeExpr) Pos() (int, int)          { return t.Expr.Pos() }
func (v *ValExpr) Pos() (int, int)           { return v.Expr.Pos() }
func (h *HashIfBlock) Pos() (int, int)       { return 0, 0 }
func (e *HashElseIfBranch) Pos() (int, int)  { return 0, 0 }
func (r *RegionBlock) Pos() (int, int)       { return 0, 0 }
func (c *CompilerDirective) Pos() (int, int) { return 0, 0 }
func (c *Comment) Pos() (int, int)           { return c.Line, c.Col }

func (_ *Module) nodeMarker()             {}
func (_ *Procedure) nodeMarker()          {}
func (_ *Function) nodeMarker()           {}
func (_ *ParamDecl) nodeMarker()          {}
func (_ *IfStmt) nodeMarker()             {}
func (_ *ElseIfBranch) nodeMarker()       {}
func (_ *WhileStmt) nodeMarker()          {}
func (_ *ForStmt) nodeMarker()            {}
func (_ *ForEachStmt) nodeMarker()        {}
func (_ *TryStmt) nodeMarker()            {}
func (_ *ReturnStmt) nodeMarker()         {}
func (_ *RaiseStmt) nodeMarker()          {}
func (_ *CycleStmt) nodeMarker()          {}
func (_ *BreakStmt) nodeMarker()          {}
func (_ *GotoStmt) nodeMarker()          {}
func (_ *VarDeclExpr) nodeMarker()       {}
func (_ *AssignmentStmt) nodeMarker()     {}
func (_ *CallStmt) nodeMarker()           {}
func (_ *BinaryExpr) nodeMarker()         {}
func (_ *UnaryExpr) nodeMarker()          {}
func (_ *TernaryExpr) nodeMarker()        {}
func (_ *Ident) nodeMarker()              {}
func (_ *NumberLit) nodeMarker()          {}
func (_ *StringLit) nodeMarker()          {}
func (_ *DateLit) nodeMarker()            {}
func (_ *BoolLit) nodeMarker()            {}
func (_ *UndefinedLit) nodeMarker()       {}
func (_ *NullLit) nodeMarker()            {}
func (_ *IndexExpr) nodeMarker()          {}
func (_ *FieldAccessExpr) nodeMarker()    {}
func (_ *NewExpr) nodeMarker()            {}
func (_ *ExecuteExpr) nodeMarker()        {}
func (_ *AddressExpr) nodeMarker()        {}
func (_ *TypeExpr) nodeMarker()           {}
func (_ *ValExpr) nodeMarker()            {}
func (_ *HashIfBlock) nodeMarker()        {}
func (_ *HashElseIfBranch) nodeMarker()   {}
func (_ *RegionBlock) nodeMarker()        {}
func (_ *CompilerDirective) nodeMarker()  {}
func (_ *Comment) nodeMarker()            {}
