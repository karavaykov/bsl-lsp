package parser

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	input   string
	pos     int
	line    int
	col     int
	readPos int
	ch      rune
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
	}
	l.skipBOM()
	l.readChar()
	return l
}

func (l *Lexer) skipBOM() {
	if len(l.input) >= 3 && l.input[:3] == "\xef\xbb\xbf" {
		l.pos = 3
		l.readPos = 3
	}
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
		l.pos = l.readPos
		return
	}

	r, size := utf8.DecodeRuneInString(l.input[l.readPos:])
	l.ch = r
	l.pos = l.readPos
	l.readPos += size

	if r == '\r' {
		// CRLF: if next char is \n, skip \r and read \n next call
		l.col++
		return
	}

	if r == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
	return r
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	line, col := l.line, l.col

	switch l.ch {
	case 0:
		tok = NewToken(TokenEOF, "", line, col)

	case '+':
		tok = NewToken(TokenPlus, string(l.ch), line, col)
		l.readChar()

	case '-':
		if l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			text := l.readComment()
			tok = NewToken(TokenComment, text, line, col)
		} else {
			tok = NewToken(TokenMinus, string(l.ch), line, col)
			l.readChar()
		}

	case '*':
		tok = NewToken(TokenStar, string(l.ch), line, col)
		l.readChar()

	case '/':
		if l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			text := l.readComment()
			tok = NewToken(TokenComment, text, line, col)
		} else if l.peekChar() == '*' {
			l.readChar()
			l.readChar()
			text, ok := l.readBlockComment()
			if !ok {
				tok = NewToken(TokenIllegal, text+" (unclosed block comment)", line, col)
			} else {
				tok = NewToken(TokenComment, text, line, col)
			}
		} else {
			tok = NewToken(TokenSlash, string(l.ch), line, col)
			l.readChar()
		}

	case '%':
		tok = NewToken(TokenMod, string(l.ch), line, col)
		l.readChar()

	case '^':
		tok = NewToken(TokenPower, string(l.ch), line, col)
		l.readChar()

	case '=':
		tok = NewToken(TokenEqual, string(l.ch), line, col)
		l.readChar()

	case '<':
		if l.peekChar() == '>' {
			l.readChar()
			tok = NewToken(TokenNotEqual, "<>", line, col)
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(TokenLessOrEqual, "<=", line, col)
		} else {
			tok = NewToken(TokenLess, string(l.ch), line, col)
		}
		l.readChar()

	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(TokenGreaterOrEqual, ">=", line, col)
		} else {
			tok = NewToken(TokenGreater, string(l.ch), line, col)
		}
		l.readChar()

	case ':':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(TokenAssign, ":=", line, col)
		} else {
			tok = NewToken(TokenColon, string(l.ch), line, col)
		}
		l.readChar()

	case '?':
		tok = NewToken(TokenQuestion, string(l.ch), line, col)
		l.readChar()

	case '.':
		tok = NewToken(TokenDot, string(l.ch), line, col)
		l.readChar()

	case ',':
		tok = NewToken(TokenComma, string(l.ch), line, col)
		l.readChar()

	case ';':
		tok = NewToken(TokenSemicolon, string(l.ch), line, col)
		l.readChar()

	case '(':
		tok = NewToken(TokenLParen, string(l.ch), line, col)
		l.readChar()

	case ')':
		tok = NewToken(TokenRParen, string(l.ch), line, col)
		l.readChar()

	case '[':
		tok = NewToken(TokenLBracket, string(l.ch), line, col)
		l.readChar()

	case ']':
		tok = NewToken(TokenRBracket, string(l.ch), line, col)
		l.readChar()

	case '~':
		tok = NewToken(TokenTilde, string(l.ch), line, col)
		l.readChar()

	case '"':
		val, ok := l.readString()
		if !ok {
			tok = NewToken(TokenIllegal, val+" (unclosed string)", line, col)
		} else {
			tok = NewToken(TokenString, val, line, col)
		}

	case '\'':
		val, ok := l.readDate()
		if !ok {
			tok = NewToken(TokenIllegal, val+" (invalid date literal)", line, col)
		} else {
			tok = NewToken(TokenDate, val, line, col)
		}

	case '|':
		l.readChar()
		text := l.readComment()
		tok = NewToken(TokenComment, text, line, col)

	case '#':
		val := l.readPreprocessor()
		tok = NewToken(l.lookupPreprocessor(val), val, line, col)

	case '&':
		val := l.readDirective()
		tok = NewToken(TokenDirective, val, line, col)

	default:
		if isLetter(l.ch) || isRussianRune(l.ch) {
			ident := l.readIdent()
			tokType := LookupIdent(ident)
			if tokType == TokenIdent {
				tok = NewToken(TokenIdent, ident, line, col)
			} else {
				tok = NewToken(tokType, ident, line, col)
			}
			return tok
		} else if isDigit(l.ch) {
			val := l.readNumber()
			tok = NewToken(TokenNumber, val, line, col)
			return tok
		} else {
			tok = NewToken(TokenIllegal, string(l.ch), line, col)
			l.readChar()
		}
	}

	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readComment() string {
	start := l.pos
	for l.ch != '\n' && l.ch != 0 {
		if l.ch == '\r' {
			l.readChar()
			continue
		}
		l.readChar()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readBlockComment() (string, bool) {
	start := l.pos
	for {
		if l.ch == 0 {
			return l.input[start:l.pos], false
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			break
		}
		l.readChar()
	}
	return l.input[start:l.pos], true
}

func (l *Lexer) readString() (string, bool) {
	var result []rune
	l.readChar()
	for {
		if l.ch == '"' {
			if l.peekChar() == '"' {
				result = append(result, '"')
				l.readChar()
				l.readChar()
				continue
			}
			l.readChar()
			return string(result), true
		}
		if l.ch == 0 {
			return string(result), false
		}
		result = append(result, l.ch)
		l.readChar()
	}
}

func (l *Lexer) readDate() (string, bool) {
	l.readChar()
	start := l.pos
	for l.ch != '\'' && l.ch != 0 {
		l.readChar()
	}
	val := l.input[start:l.pos]
	ok := true
	if l.ch == '\'' {
		l.readChar()
	}
	// Проверяем формат: только цифры и не пустая
	for _, r := range val {
		if r < '0' || r > '9' {
			ok = false
			break
		}
	}
	if len(val) == 0 {
		ok = false
	}
	return val, ok
}

func (l *Lexer) readIdent() string {
	start := l.pos
	for l.ch != 0 && (isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || isRussianRune(l.ch)) {
		l.readChar()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readNumber() string {
	start := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readPreprocessor() string {
	start := l.pos
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[start:l.pos]
}

var preprocessorKeywords = map[string]TokenType{
	"#если":          TokenHashIf,
	"#иначеесли":     TokenHashElseIf,
	"#иначе":         TokenHashElse,
	"#конецесли":     TokenHashEndIf,
	"#область":       TokenHashRegion,
	"#конецобласти":  TokenHashEndRegion,
	"#вставка":       TokenHashInsert,
	"#удалить":       TokenHashDelete,
}

func (l *Lexer) lookupPreprocessor(val string) TokenType {
	trimmed := val
	for i, ch := range trimmed {
		if ch == ' ' || ch == '\t' || ch == '\n' {
			trimmed = trimmed[:i]
			break
		}
	}
	if t, ok := preprocessorKeywords[toLower(trimmed)]; ok {
		return t
	}
	return TokenPreprocessor
}

func (l *Lexer) readDirective() string {
	start := l.pos
	l.readChar()
	for l.ch != 0 && (isRussianRune(l.ch) || isLetter(l.ch)) {
		l.readChar()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isRussianRune(ch rune) bool {
	return unicode.Is(unicode.Cyrillic, ch)
}

func (l *Lexer) Pos() (line, col int) {
	return l.line, l.col
}

func (l *Lexer) Err(format string, args ...interface{}) error {
	return fmt.Errorf("line %d:%d: %s", l.line, l.col, fmt.Sprintf(format, args...))
}
