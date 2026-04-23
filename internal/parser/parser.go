package parser

import (
	"fmt"
)

type ParseError struct {
	Line    int
	Col     int
	Message string
}

type Parser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
	errors    []ParseError
}

func NewParser(input string) *Parser {
	p := &Parser{
		lexer: NewLexer(input),
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.errorExpected(t)
	return false
}

func (p *Parser) errorExpected(t TokenType) {
	p.errors = append(p.errors, ParseError{
		Line:    p.peekToken.Line,
		Col:     p.peekToken.Col,
		Message: fmt.Sprintf("expected %s, got %s (%s)", t, p.peekToken.Type, p.peekToken.Literal),
	})
}

func (p *Parser) error(format string, args ...interface{}) {
	p.errors = append(p.errors, ParseError{
		Line:    p.curToken.Line,
		Col:     p.curToken.Col,
		Message: fmt.Sprintf(format, args...),
	})
}

func (p *Parser) Errors() []ParseError {
	return p.errors
}

// skipComments — пропускает комментарии, препроцессор и токены, не относящиеся к statements
func (p *Parser) skipComments() {
	for p.curTokenIs(TokenComment) || p.curTokenIs(TokenPreprocessor) {
		p.nextToken()
	}
}

// ParseModule — модуль верхнего уровня
func (p *Parser) ParseModule() *Module {
	mod := &Module{}

	for !p.curTokenIs(TokenEOF) {
		p.skipComments()

		if p.curTokenIs(TokenEOF) {
			break
		}

		stmt := p.parseModuleStatement()
		if stmt != nil {
			mod.Statements = append(mod.Statements, stmt)
			continue
		}

		// Не удалось разобрать — пропускаем
		p.syncToStmt(nil)
	}

	return mod
}

// parseModuleStatement — парсит statement на уровне модуля
func (p *Parser) parseModuleStatement() Node {
	switch p.curToken.Type {
	case TokenProcedure:
		return p.parseProcedure()
	case TokenFunction:
		return p.parseFunction()
	case TokenVar:
		return p.parseVarDecl()
	case TokenHashRegion, TokenHashArea:
		return p.parseRegion()
	case TokenDirective:
		return p.parseCompilerDirective()
	case TokenHashIf:
		return p.parseHashIf()
	default:
		return p.parseStatement()
	}
}

// parseStatement — парсит statement внутри процедуры/функции/блока
func (p *Parser) parseStatement() Node {
	switch p.curToken.Type {
	case TokenIf:
		return p.parseIfStmt()
	case TokenWhile:
		return p.parseWhileStmt()
	case TokenFor:
		return p.parseForOrForEach()
	case TokenTry:
		return p.parseTryStmt()
	case TokenReturn:
		return p.parseReturnStmt()
	case TokenRaise:
		return p.parseRaiseStmt()
	case TokenCycle:
		p.nextToken()
		return &CycleStmt{}
	case TokenBreak:
		p.nextToken()
		return &BreakStmt{}
	case TokenGoTo:
		return p.parseGotoStmt()
	case TokenVar:
		return p.parseVarDecl()
	case TokenTilde:
		return p.parseLabelDef()
	case TokenSemicolon:
		p.nextToken()
		return nil
	case TokenComment, TokenPreprocessor, TokenDirective:
		p.nextToken()
		return nil
	case TokenHashIf, TokenHashElseIf, TokenHashElse, TokenHashEndIf:
		p.error("preprocessor conditional directives are not allowed inside procedures and blocks")
		p.nextToken()
		return nil
	default:
		return p.parseExpressionStmt()
	}
}

func isStmtStart(t TokenType) bool {
	switch t {
	case TokenIf, TokenWhile, TokenFor, TokenTry,
		TokenReturn, TokenRaise, TokenCycle, TokenBreak, TokenGoTo, TokenVar,
		TokenComment, TokenPreprocessor, TokenDirective, TokenTilde,
		TokenHashRegion, TokenHashEndRegion, TokenHashArea,
		TokenHashIf, TokenHashElseIf, TokenHashElse, TokenHashEndIf,
		TokenIdent, TokenNumber, TokenString, TokenDate,
		TokenTrue, TokenFalse, TokenUndefined, TokenNull,
		TokenMinus, TokenNot, TokenNew, TokenExecute,
		TokenAddress, TokenTypeName, TokenVal, TokenLParen:
		return true
	}
	return false
}

// parseBlock — парсит последовательность statement'ов до заданного endToken
func (p *Parser) parseBlock(endTokens ...TokenType) []Node {
	var stmts []Node
	for !p.curTokenIs(TokenEOF) {
		p.skipComments()

		for _, et := range endTokens {
			if p.curTokenIs(et) {
				return stmts
			}
		}

		if p.curTokenIs(TokenIllegal) {
			p.error("illegal token: %s", p.curToken.Literal)
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
			p.consumeSemicolon()
			continue
		}

		// Не удалось разобрать statement — синхронизация
		p.syncToStmt(endTokens)
	}
	return stmts
}

// syncToStmt — пропускает токены до возобновляющего токена или endToken
func (p *Parser) syncToStmt(endTokens []TokenType) {
	for !p.curTokenIs(TokenEOF) {
		for _, et := range endTokens {
			if p.curTokenIs(et) {
				return
			}
		}
		if isStmtStart(p.curToken.Type) || p.curTokenIs(TokenSemicolon) {
			return
		}
		p.nextToken()
	}
}

// ============================================================
// Процедуры, функции, параметры
// ============================================================

func (p *Parser) parseProcedure() *Procedure {
	proc := &Procedure{
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}
	p.nextToken()
	p.skipComments()

	if p.isIdentToken() {
		proc.Name = p.curToken.Literal
		p.nextToken()
	} else {
		p.error("expected procedure name")
	}

	p.skipComments()

	if p.curTokenIs(TokenLParen) {
		proc.Params = p.parseParamList()
	}

	p.skipComments()

	if p.curTokenIs(TokenExport) {
		proc.Export = true
		p.nextToken()
	}

	proc.Body = p.parseBlock(TokenEndProcedure)
	p.skipComments()
	if p.curTokenIs(TokenEndProcedure) {
		p.nextToken()
	}

	return proc
}

func (p *Parser) parseFunction() *Function {
	fn := &Function{
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}
	p.nextToken()
	p.skipComments()

	if p.isIdentToken() {
		fn.Name = p.curToken.Literal
		p.nextToken()
	} else {
		p.error("expected function name")
	}

	p.skipComments()

	if p.curTokenIs(TokenLParen) {
		fn.Params = p.parseParamList()
	}

	p.skipComments()

	if p.curTokenIs(TokenExport) {
		fn.Export = true
		p.nextToken()
	}

	fn.Body = p.parseBlock(TokenEndFunction)
	p.skipComments()
	if p.curTokenIs(TokenEndFunction) {
		p.nextToken()
	}

	return fn
}

func (p *Parser) parseParamList() []*ParamDecl {
	var params []*ParamDecl
	p.nextToken()

	for !p.curTokenIs(TokenRParen) && !p.curTokenIs(TokenEOF) {
		param := &ParamDecl{}

		if p.curTokenIs(TokenVal) {
			param.ByVal = true
			p.nextToken()
		}

		if p.curTokenIs(TokenIdent) {
			param.Name = p.curToken.Literal
			p.nextToken()
		} else {
			p.error("expected parameter name")
		}

		params = append(params, param)

		p.skipComments()

		if p.curTokenIs(TokenComma) {
			p.nextToken()
		}

		p.skipComments()
	}

	if p.curTokenIs(TokenRParen) {
		p.nextToken()
	}

	return params
}

// ============================================================
// If / ElsIf / Else / EndIf
// ============================================================

func (p *Parser) parseIfStmt() *IfStmt {
	stmt := &IfStmt{
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}
	p.nextToken() // 'Если'

	p.skipComments()
	stmt.Condition = p.parseCompareExpr()
	p.skipComments()

	if !p.curTokenIs(TokenThen) {
		p.error("expected ТОГДА")
	}
	p.nextToken()

	stmt.Body = p.parseBlock(TokenElseIf, TokenElse, TokenEndIf)

	for p.curTokenIs(TokenElseIf) {
		p.nextToken()
		ei := &ElseIfBranch{}
		p.skipComments()
		ei.Condition = p.parseCompareExpr()
		p.skipComments()
		if !p.curTokenIs(TokenThen) {
			p.error("expected ТОГДА")
		}
		p.nextToken()
		ei.Body = p.parseBlock(TokenElseIf, TokenElse, TokenEndIf)
		stmt.ElseIf = append(stmt.ElseIf, ei)
	}

	if p.curTokenIs(TokenElse) {
		p.nextToken()
		stmt.ElseBody = p.parseBlock(TokenEndIf)
	}

	if p.curTokenIs(TokenEndIf) {
		p.nextToken()
	}

	return stmt
}

// ============================================================
// While / Do / EndDo
// ============================================================

func (p *Parser) parseWhileStmt() *WhileStmt {
	stmt := &WhileStmt{
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}
	p.nextToken() // 'Пока'

	p.skipComments()
	stmt.Condition = p.parseCompareExpr()
	p.skipComments()

	if !p.curTokenIs(TokenDo) {
		p.error("expected ЦИКЛ")
	} else {
		p.nextToken()
	}

	stmt.Body = p.parseBlock(TokenEndDo)

	if p.curTokenIs(TokenEndDo) {
		p.nextToken()
	}

	return stmt
}

// ============================================================
// For / To / Do / EndDo
// ============================================================

func (p *Parser) parseForStmt() *ForStmt {
	stmt := &ForStmt{
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}
	p.nextToken() // 'Для'

	if p.curTokenIs(TokenIdent) {
		stmt.Var = p.curToken.Literal
		p.nextToken()
	} else {
		p.error("expected loop variable")
	}

	if p.curTokenIs(TokenEqual) {
		p.nextToken()
	}

	p.skipComments()
	stmt.From = p.parseExpression()
	p.skipComments()

	if p.curTokenIs(TokenTo) {
		p.nextToken()
	}

	p.skipComments()
	stmt.To = p.parseExpression()
	p.skipComments()

	if p.curTokenIs(TokenDo) {
		p.nextToken()
	}

	stmt.Body = p.parseBlock(TokenEndDo)

	if p.curTokenIs(TokenEndDo) {
		p.nextToken()
	}

	return stmt
}

// ============================================================
// ForEach / In / Do / EndDo
// ============================================================

func (p *Parser) parseForEachStmt() *ForEachStmt {
	stmt := &ForEachStmt{
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}
	p.nextToken() // 'Для'

	if p.curTokenIs(TokenIdent) && equalsFold(p.curToken.Literal, "каждого") {
		p.nextToken()
	}

	if p.curTokenIs(TokenIdent) {
		stmt.Var = p.curToken.Literal
		p.nextToken()
	} else {
		p.error("expected loop variable")
	}

	if p.curTokenIs(TokenIn) {
		p.nextToken()
	}

	p.skipComments()
	stmt.In = p.parseExpression()
	p.skipComments()

	if p.curTokenIs(TokenDo) {
		p.nextToken()
	}

	stmt.Body = p.parseBlock(TokenEndDo)

	if p.curTokenIs(TokenEndDo) {
		p.nextToken()
	}

	return stmt
}

// ============================================================
// Try / Except / EndTry
// ============================================================

func (p *Parser) parseTryStmt() *TryStmt {
	stmt := &TryStmt{
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}
	p.nextToken() // 'Попытка'

	stmt.Body = p.parseBlock(TokenExcept, TokenEndTry)

	if p.curTokenIs(TokenExcept) {
		p.nextToken()
		stmt.Except = p.parseBlock(TokenEndTry)
	} else if p.curTokenIs(TokenEndTry) {
		p.error("expected ИСКЛЮЧЕНИЕ")
	}

	if p.curTokenIs(TokenEndTry) {
		p.nextToken()
	}

	return stmt
}

// ============================================================
// Return / Raise / Goto
// ============================================================

func (p *Parser) parseReturnStmt() *ReturnStmt {
	stmt := &ReturnStmt{}
	p.nextToken() // 'Возврат'

	if p.isExprStart() {
		stmt.Value = p.parseExpression()
	}

	return stmt
}

func (p *Parser) parseRaiseStmt() *RaiseStmt {
	stmt := &RaiseStmt{}
	p.nextToken()

	if p.isExprStart() {
		stmt.Value = p.parseExpression()
	}

	return stmt
}

func (p *Parser) parseGotoStmt() *GotoStmt {
	stmt := &GotoStmt{}
	p.nextToken()

	if p.curTokenIs(TokenTilde) {
		p.nextToken()
	}

	if p.curTokenIs(TokenIdent) {
		stmt.Label = p.curToken.Literal
		p.nextToken()
	} else {
		p.error("expected label after ПЕРЕЙТИ")
	}

	return stmt
}

func (p *Parser) parseLabelDef() Node {
	tok := p.curToken
	p.nextToken()
	if p.curTokenIs(TokenIdent) {
		label := &LabelStmt{
			Label: p.curToken.Literal,
			Line:  tok.Line,
			Col:   tok.Col,
		}
		p.nextToken()
		if p.curTokenIs(TokenColon) {
			p.nextToken()
		}
		return label
	}
	p.error("expected label name after ~")
	return &LabelStmt{Label: "", Line: tok.Line, Col: tok.Col}
}

// ============================================================
// Переменные (Перем)
// ============================================================

func (p *Parser) parseVarDecl() Node {
	p.nextToken() // Перем

	p.skipComments()

	if !p.curTokenIs(TokenIdent) {
		return nil
	}

	decl := &VarDeclExpr{
		Name: p.curToken.Literal,
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}
	p.nextToken()

	// skip optional Экспорт
	if p.curTokenIs(TokenExport) {
		p.nextToken()
	}

	return decl
}

// ============================================================
// Директивы компиляции &НаКлиенте и т.д.
// ============================================================

func (p *Parser) parseCompilerDirective() Node {
	dir := &CompilerDirective{Name: p.curToken.Literal}
	p.nextToken()
	return dir
}

// ============================================================
// #Если / #ИначеЕсли / #Иначе / #КонецЕсли
// ============================================================

func (p *Parser) parseHashIf() Node {
	block := &HashIfBlock{}
	block.Condition = p.curToken.Literal
	p.nextToken()

	block.Body = p.parseBlock(TokenHashElseIf, TokenHashElse, TokenHashEndIf)

	for p.curTokenIs(TokenHashElseIf) {
		ei := &HashElseIfBranch{Condition: p.curToken.Literal}
		p.nextToken()
		ei.Body = p.parseBlock(TokenHashElseIf, TokenHashElse, TokenHashEndIf)
		block.ElseIf = append(block.ElseIf, ei)
	}

	if p.curTokenIs(TokenHashElse) {
		p.nextToken()
		block.ElseBody = p.parseBlock(TokenHashEndIf)
	}

	if p.curTokenIs(TokenHashEndIf) {
		p.nextToken()
	}

	return block
}

// ============================================================
// #Область / #КонецОбласти
// ============================================================

func (p *Parser) parseRegion() Node {
	region := &RegionBlock{Name: p.curToken.Literal}
	p.nextToken()
	region.Body = p.parseBlock(TokenHashEndRegion)
	if p.curTokenIs(TokenHashEndRegion) {
		p.nextToken()
	}
	return region
}

func (p *Parser) parseExpressionStmt() Node {
	if p.curTokenIs(TokenRParen) || p.curTokenIs(TokenRBracket) || p.curTokenIs(TokenColon) || p.curTokenIs(TokenComma) {
		p.error("unexpected token %s (%s)", p.curToken.Type, p.curToken.Literal)
		p.nextToken()
		return nil
	}

	expr := p.parseExpression()
	if expr == nil {
		return nil
	}

	if be, ok := expr.(*BinaryExpr); ok && be.Op == TokenEqual {
		if ident, ok := be.Left.(*Ident); ok && ident.Name == "_" {
			p.error("bare underscore is not a valid identifier")
			return nil
		}
		// Конвертируем сравнение '=' обратно в присваивание
		if isAssignLHS(be.Left) && !p.curTokenIs(TokenEqual) {
			stmt := &AssignmentStmt{Left: be.Left, Right: be.Right}
			return stmt
		}
	}

	if p.curTokenIs(TokenAssign) || (p.curTokenIs(TokenEqual) && isAssignLHS(expr)) {
		if ident, ok := expr.(*Ident); ok && ident.Name == "_" {
			p.error("bare underscore is not a valid identifier")
			p.nextToken()
			p.parseExpression()
			return nil
		}
		stmt := &AssignmentStmt{Left: expr}
		p.nextToken()
		stmt.Right = p.parseExpression()
		return stmt
	}

	if call, ok := expr.(*CallStmt); ok {
		return call
	}

	switch expr.(type) {
	case *NumberLit, *StringLit, *DateLit, *BoolLit, *UndefinedLit, *NullLit:
		p.error("expression is not a valid statement")
	}

	return expr
}

func isAssignLHS(n Node) bool {
	switch n.(type) {
	case *Ident, *FieldAccessExpr, *IndexExpr:
		return true
	}
	return false
}

// ============================================================
// Expressions (recursive descent, precedence climbing)
// ============================================================

type precedence int

const (
	precLowest    precedence = iota
	precAssign               // :=
	precTernary              // ?
	precComparison           // =, <>, <, >, <=, >=
	precNot                  // Не
	precAnd                  // И
	precOr                   // Или
	precAdd                  // +, -
	precMul                  // *, /, %
	precPower                // ^
	precPrefix               // - (unary), Не
	precPostfix              // ., [ ]
	precCall                 // (
)

var precedences = map[TokenType]precedence{
	TokenOr:             precOr,
	TokenAnd:            precAnd,
	TokenNot:            precNot,
	TokenEqual:          precComparison,
	TokenNotEqual:       precComparison,
	TokenLess:           precComparison,
	TokenGreater:        precComparison,
	TokenLessOrEqual:    precComparison,
	TokenGreaterOrEqual: precComparison,
	TokenPlus:           precAdd,
	TokenMinus:          precAdd,
	TokenStar:           precMul,
	TokenSlash:          precMul,
	TokenMod:            precMul,
	TokenPower:          precPower,
	TokenQuestion:       precTernary,
	TokenAssign:         precAssign,
}

func (p *Parser) peekPrecedence() precedence {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return precLowest
}

func (p *Parser) curPrecedence() precedence {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return precLowest
}

func (p *Parser) parseExpression() Node {
	return p.parseBinaryOrTernary(precLowest)
}

// parseCompareExpr парсит выражение, включая операторы сравнения (=, <>, <, >, <=, >=)
func (p *Parser) parseCompareExpr() Node {
	return p.parseBinaryOrTernary(precComparison)
}

func (p *Parser) parseBinaryOrTernary(prec precedence) Node {
	left := p.parsePrefix()
	if left == nil {
		return nil
	}

	for !p.curTokenIs(TokenSemicolon) && !p.curTokenIs(TokenEOF) {
		p.skipComments()

		if p.curTokenIs(TokenQuestion) && p.peekTokenIs(TokenColon) {
			// Тернарный оператор: условие ? (выражение) : (выражение)
			tern := &TernaryExpr{Condition: left}
			p.nextToken() // ?
			p.nextToken() // :
			tern.True = p.parseExpression()

			if p.curTokenIs(TokenColon) {
				p.nextToken()
				tern.False = p.parseExpression()
			}

			left = tern
			continue
		}

		nextPrec := p.curPrecedence()
		if nextPrec < prec {
			break
		}

			// Не-тернарный бинарный оператор
		if p.isBinaryOp() || p.curTokenIs(TokenEqual) {
			op := p.curToken.Type
			p.nextToken()
			right := p.parseBinaryOrTernary(nextPrec + 1)
			if right == nil {
				p.error("expected expression after %s", op)
			}
			left = &BinaryExpr{Left: left, Op: op, Right: right}
			continue
		}

		break
	}

	return left
}

func (p *Parser) isBinaryOp() bool {
	switch p.curToken.Type {
	case TokenPlus, TokenMinus, TokenStar, TokenSlash, TokenMod, TokenPower,
		TokenNotEqual, TokenLess, TokenGreater,
		TokenLessOrEqual, TokenGreaterOrEqual,
		TokenAnd, TokenOr:
		return true
	}
	return false
}

func (p *Parser) parsePrefix() Node {
	p.skipComments()

	switch p.curToken.Type {
	case TokenMinus:
		return p.parseUnaryExpr()
	case TokenNot:
		return p.parseUnaryExpr()
	case TokenQuestion:
		if p.peekTokenIs(TokenLParen) {
			return p.parseTernaryFunc()
		}
		tok := p.curToken
		p.nextToken()
		return &Ident{Name: tok.Literal, Line: tok.Line, Col: tok.Col}
	case TokenNew:
		return p.parseNewExpr()
	case TokenExecute:
		if p.peekTokenIs(TokenLParen) {
			return p.parseIdentOrCall()
		}
		return p.parseExecuteExpr()
	case TokenAddress:
		return p.parseAddressExpr()
	case TokenTypeName:
		return p.parseTypeExpr()
	case TokenVal:
		return p.parseValExpr()
	case TokenTrue:
		tok := p.curToken
		p.nextToken()
		return &BoolLit{Value: true, Line: tok.Line, Col: tok.Col}
	case TokenFalse:
		tok := p.curToken
		p.nextToken()
		return &BoolLit{Value: false, Line: tok.Line, Col: tok.Col}
	case TokenUndefined:
		tok := p.curToken
		p.nextToken()
		return &UndefinedLit{Line: tok.Line, Col: tok.Col}
	case TokenNull:
		tok := p.curToken
		p.nextToken()
		return &NullLit{Line: tok.Line, Col: tok.Col}
	case TokenNumber:
		tok := p.curToken
		p.nextToken()
		return &NumberLit{Value: tok.Literal, Line: tok.Line, Col: tok.Col}
	case TokenString:
		tok := p.curToken
		p.nextToken()
		return &StringLit{Value: tok.Literal, Line: tok.Line, Col: tok.Col}
	case TokenDate:
		tok := p.curToken
		p.nextToken()
		return &DateLit{Value: tok.Literal, Line: tok.Line, Col: tok.Col}
	case TokenIdent:
		return p.parseIdentOrCall()
	case TokenLParen:
		p.nextToken()
		expr := p.parseCompareExpr()
		if p.curTokenIs(TokenRParen) {
			p.nextToken()
		}
		return expr
	case TokenDirective:
		return p.parseCompilerDirective()
	default:
		if p.curTokenIs(TokenEOF) {
			return nil
		}
		if p.curTokenIs(TokenIllegal) {
			p.error("illegal token: %s", p.curToken.Literal)
			p.nextToken()
			return nil
		}
		// Токены, которые не могут начинать выражение — молча пропускаем
		// (бинарные операторы, скобки `)`, `]`, `;`, `:`, `,` и т.д.)
		if !p.isExprStart() {
			p.nextToken()
			return nil
		}
		tok := p.curToken
		p.error("unexpected token %s (%s)", tok.Type, tok.Literal)
		p.nextToken()
		return nil
	}
}

func (p *Parser) parseUnaryExpr() *UnaryExpr {
	expr := &UnaryExpr{Op: p.curToken.Type}
	p.nextToken()
	expr.Right = p.parsePrefix()
	return expr
}

func (p *Parser) parseTernaryFunc() Node {
	tok := p.curToken
	p.nextToken()
	p.nextToken()
	var args []Node
	for !p.curTokenIs(TokenRParen) && !p.curTokenIs(TokenEOF) {
		arg := p.parseExpression()
		args = append(args, arg)
		p.skipComments()
		if p.curTokenIs(TokenComma) {
			p.nextToken()
		}
		p.skipComments()
	}
	if p.curTokenIs(TokenRParen) {
		p.nextToken()
	}
	return &CallStmt{
		Function:  "?",
		Args:      args,
		Line:      tok.Line,
		Col:       tok.Col,
	}
}

func (p *Parser) parseIdentOrCall() Node {
	tok := p.curToken
	p.nextToken()

	var expr Node = &Ident{Name: tok.Literal, Line: tok.Line, Col: tok.Col}

	for {
		p.skipComments()

		if p.curTokenIs(TokenDot) {
			p.nextToken()
			if p.isFieldNameToken() {
				expr = &FieldAccessExpr{Object: expr, Field: p.curToken.Literal}
				p.nextToken()
				continue
			}
			break
		}

		if p.curTokenIs(TokenLBracket) {
			p.nextToken()
			index := p.parseExpression()
			expr = &IndexExpr{Object: expr, Index: index}
			if p.curTokenIs(TokenRBracket) {
				p.nextToken()
			}
			continue
		}

		if p.curTokenIs(TokenLParen) {
			p.nextToken()
			var args []Node
			for !p.curTokenIs(TokenRParen) && !p.curTokenIs(TokenEOF) {
				arg := p.parseExpression()
				args = append(args, arg)
				p.skipComments()
				if p.curTokenIs(TokenComma) {
					p.nextToken()
				}
				p.skipComments()
			}
			if p.curTokenIs(TokenRParen) {
				p.nextToken()
			}
			// Если expr — FieldAccess, создаем вызов метода
			if fa, ok := expr.(*FieldAccessExpr); ok {
				expr = &CallStmt{Function: fa.Field, Object: fa.Object, Args: args}
			} else {
				expr = &CallStmt{Function: tok.Literal, Args: args}
			}
			continue
		}

		break
	}

	return expr
}

func (p *Parser) parseNewExpr() *NewExpr {
	expr := &NewExpr{}
	p.nextToken()
	if p.curTokenIs(TokenIdent) {
		expr.TypeName = p.curToken.Literal
		p.nextToken()
	} else if !p.curTokenIs(TokenLParen) {
		p.error("expected type name after НОВЫЙ")
	}
	if p.curTokenIs(TokenLParen) {
		p.nextToken()
		for !p.curTokenIs(TokenRParen) && !p.curTokenIs(TokenEOF) {
			arg := p.parseExpression()
			expr.Args = append(expr.Args, arg)
			if p.curTokenIs(TokenComma) {
				p.nextToken()
			}
		}
		if p.curTokenIs(TokenRParen) {
			p.nextToken()
		}
	}
	return expr
}

func (p *Parser) parseExecuteExpr() *ExecuteExpr {
	expr := &ExecuteExpr{}
	p.nextToken()
	if p.curTokenIs(TokenLParen) {
		p.nextToken()
	}
	expr.Expr = p.parseExpression()
	if p.curTokenIs(TokenRParen) {
		p.nextToken()
	}
	return expr
}

func (p *Parser) parseAddressExpr() *AddressExpr {
	expr := &AddressExpr{}
	p.nextToken()
	if p.curTokenIs(TokenLParen) {
		p.nextToken()
	}
	expr.Expr = p.parseExpression()
	if p.curTokenIs(TokenRParen) {
		p.nextToken()
	}
	return expr
}

func (p *Parser) parseTypeExpr() *TypeExpr {
	expr := &TypeExpr{}
	p.nextToken()
	if p.curTokenIs(TokenLParen) {
		p.nextToken()
	}
	expr.Expr = p.parseExpression()
	if p.curTokenIs(TokenRParen) {
		p.nextToken()
	}
	return expr
}

func (p *Parser) parseValExpr() *ValExpr {
	expr := &ValExpr{}
	p.nextToken()
	if p.curTokenIs(TokenLParen) {
		p.nextToken()
	}
	expr.Expr = p.parseExpression()
	if p.curTokenIs(TokenRParen) {
		p.nextToken()
	}
	return expr
}

func (p *Parser) parseForOrForEach() Node {
	if !p.curTokenIs(TokenFor) {
		return nil
	}
	if p.peekTokenIs(TokenIdent) && equalsFold(p.peekToken.Literal, "каждого") {
		return p.parseForEachStmt()
	}
	return p.parseForStmt()
}

func (p *Parser) isExprStart() bool {
	switch p.curToken.Type {
	case TokenIdent, TokenNumber, TokenString, TokenDate,
		TokenTrue, TokenFalse, TokenUndefined, TokenNull,
		TokenMinus, TokenNot, TokenNew, TokenExecute,
		TokenAddress, TokenTypeName, TokenVal,
		TokenLParen, TokenDirective:
		return true
	}
	return false
}

// isIdentToken — может ли токен быть именем процедуры/функции/переменной
func (p *Parser) isIdentToken() bool {
	switch p.curToken.Type {
	case TokenIdent, TokenExecute, TokenAddress, TokenTypeName, TokenVal:
		return true
	}
	return false
}

func containsToken(tokens []TokenType, t TokenType) bool {
	for _, tok := range tokens {
		if tok == t {
			return true
		}
	}
	return false
}

func (p *Parser) isFieldNameToken() bool {
	switch p.curToken.Type {
	case TokenIdent, TokenNew, TokenExecute, TokenAddress, TokenTypeName, TokenVal,
		TokenTrue, TokenFalse, TokenUndefined, TokenNull:
		return true
	}
	return false
}

func (p *Parser) consumeSemicolon() {
	if p.curTokenIs(TokenSemicolon) {
		p.nextToken()
	}
}
