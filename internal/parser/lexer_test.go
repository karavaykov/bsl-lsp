package parser

import (
	"testing"
)

func TestLexer_Keywords(t *testing.T) {
	tests := []struct {
		input string
		typ   TokenType
		lit   string
	}{
		{"Процедура", TokenProcedure, "Процедура"},
		{"Функция", TokenFunction, "Функция"},
		{"КонецПроцедуры", TokenEndProcedure, "КонецПроцедуры"},
		{"КонецФункции", TokenEndFunction, "КонецФункции"},
		{"Если", TokenIf, "Если"},
		{"Тогда", TokenThen, "Тогда"},
		{"ИначеЕсли", TokenElseIf, "ИначеЕсли"},
		{"Иначе", TokenElse, "Иначе"},
		{"КонецЕсли", TokenEndIf, "КонецЕсли"},
		{"Для", TokenFor, "Для"},
		{"Пока", TokenWhile, "Пока"},
		{"Цикл", TokenDo, "Цикл"},
		{"КонецЦикла", TokenEndDo, "КонецЦикла"},
		{"ДляКаждого", TokenForEach, "ДляКаждого"},
		{"Из", TokenIn, "Из"},
		{"Продолжить", TokenCycle, "Продолжить"},
		{"Прервать", TokenBreak, "Прервать"},
		{"Попытка", TokenTry, "Попытка"},
		{"Исключение", TokenExcept, "Исключение"},
		{"КонецПопытки", TokenEndTry, "КонецПопытки"},
		{"Возврат", TokenReturn, "Возврат"},
		{"Перем", TokenVar, "Перем"},
		{"Экспорт", TokenExport, "Экспорт"},
		{"Знач", TokenVal, "Знач"},
		{"Новый", TokenNew, "Новый"},
		{"Не", TokenNot, "Не"},
		{"И", TokenAnd, "И"},
		{"Или", TokenOr, "Или"},
		{"Истина", TokenTrue, "Истина"},
		{"Ложь", TokenFalse, "Ложь"},
		{"Неопределено", TokenUndefined, "Неопределено"},
		{"Null", TokenNull, "Null"},
		{"Тип", TokenTypeName, "Тип"},
		{"Перейти", TokenGoTo, "Перейти"},
		{"ПерейтиЕсли", TokenIfGoto, "ПерейтиЕсли"},
	}

	for _, tt := range tests {
		t.Run(tt.lit, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.NextToken()
			if tok.Type != tt.typ {
				t.Errorf("expected type %s, got %s", tt.typ, tok.Type)
			}
			if tok.Literal != tt.lit {
				t.Errorf("expected literal %q, got %q", tt.lit, tok.Literal)
			}
			if tok.Type == TokenEOF {
				t.Error("unexpected EOF")
			}
		})
	}
}

func TestLexer_Directives(t *testing.T) {
	tests := []struct {
		input string
		lit   string
	}{
		{"&НаКлиенте", "&НаКлиенте"},
		{"&НаСервере", "&НаСервере"},
		{"&НаСервереБезКонтекста", "&НаСервереБезКонтекста"},
		{"&НаКлиентеНаСервере", "&НаКлиентеНаСервере"},
		{"&НаКлиентеНаСервереБезКонтекста", "&НаКлиентеНаСервереБезКонтекста"},
	}

	for _, tt := range tests {
		t.Run(tt.lit, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.NextToken()
			if tok.Type != TokenDirective {
				t.Errorf("expected Directive, got %s", tok.Type)
			}
			if tok.Literal != tt.lit {
				t.Errorf("expected %q, got %q", tt.lit, tok.Literal)
			}
		})
	}
}

func TestLexer_Operators(t *testing.T) {
	tests := []struct {
		input string
		typ   TokenType
	}{
		{"+", TokenPlus},
		{"-", TokenMinus},
		{"*", TokenStar},
		{"/", TokenSlash},
		{"%", TokenMod},
		{"=", TokenEqual},
		{"<>", TokenNotEqual},
		{"<", TokenLess},
		{">", TokenGreater},
		{"<=", TokenLessOrEqual},
		{">=", TokenGreaterOrEqual},
		{":=", TokenAssign},
		{"?", TokenQuestion},
		{".", TokenDot},
		{",", TokenComma},
		{";", TokenSemicolon},
		{":", TokenColon},
		{"(", TokenLParen},
		{")", TokenRParen},
		{"[", TokenLBracket},
		{"]", TokenRBracket},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.NextToken()
			if tok.Type != tt.typ {
				t.Errorf("expected %s, got %s", tt.typ, tok.Type)
			}
		})
	}
}

func TestLexer_String(t *testing.T) {
	input := `"Hello, World!"`
	l := NewLexer(input)
	tok := l.NextToken()

	if tok.Type != TokenString {
		t.Errorf("expected String, got %s", tok.Type)
	}
	if tok.Literal != "Hello, World!" {
		t.Errorf("expected %q, got %q", "Hello, World!", tok.Literal)
	}
}

func TestLexer_StringWithDoubleQuote(t *testing.T) {
	input := `"ООО ""Ромашка"""`
	l := NewLexer(input)
	tok := l.NextToken()

	if tok.Type != TokenString {
		t.Errorf("expected String, got %s", tok.Type)
	}
	if tok.Literal != `ООО "Ромашка"` {
		t.Errorf("expected %q, got %q", `ООО "Ромашка"`, tok.Literal)
	}
}

func TestLexer_Number(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"123", "123"},
		{"3.14", "3.14"},
		{"0", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.NextToken()
			if tok.Type != TokenNumber {
				t.Errorf("expected Number, got %s", tok.Type)
			}
			if tok.Literal != tt.value {
				t.Errorf("expected %q, got %q", tt.value, tok.Literal)
			}
		})
	}
}

func TestLexer_Date(t *testing.T) {
	input := `'20240101'`
	l := NewLexer(input)
	tok := l.NextToken()

	if tok.Type != TokenDate {
		t.Errorf("expected Date, got %s", tok.Type)
	}
	if tok.Literal != "20240101" {
		t.Errorf("expected %q, got %q", "20240101", tok.Literal)
	}
}

func TestLexer_Comment(t *testing.T) {
	input := "// это комментарий"
	l := NewLexer(input)
	tok := l.NextToken()

	if tok.Type != TokenComment {
		t.Errorf("expected Comment, got %s", tok.Type)
	}
	if tok.Literal != " это комментарий" {
		t.Errorf("expected %q, got %q", " это комментарий", tok.Literal)
	}
}

func TestLexer_Ident(t *testing.T) {
	input := "МояПеременная"
	l := NewLexer(input)
	tok := l.NextToken()

	if tok.Type != TokenIdent {
		t.Errorf("expected Ident, got %s", tok.Type)
	}
	if tok.Literal != "МояПеременная" {
		t.Errorf("expected %q, got %q", "МояПеременная", tok.Literal)
	}
}

func TestLexer_IdentWithUnderscore(t *testing.T) {
	input := "_myVar"
	l := NewLexer(input)
	tok := l.NextToken()

	if tok.Type != TokenIdent {
		t.Errorf("expected Ident, got %s", tok.Type)
	}
	if tok.Literal != "_myVar" {
		t.Errorf("expected %q, got %q", "_myVar", tok.Literal)
	}
}

func TestLexer_LineCol(t *testing.T) {
	input := "Процедура\n  Перем"
	l := NewLexer(input)

	tok1 := l.NextToken()
	if tok1.Line != 1 || tok1.Col != 1 {
		t.Errorf("expected line=1 col=1, got line=%d col=%d", tok1.Line, tok1.Col)
	}

	tok2 := l.NextToken()
	if tok2.Line != 2 || tok2.Col != 3 {
		t.Errorf("expected line=2 col=3, got line=%d col=%d", tok2.Line, tok2.Col)
	}
}

func TestLexer_FullProcedure(t *testing.T) {
	input := `Процедура Тест(Парам1, Парам2) Экспорт
	// тело
	Возврат
КонецПроцедуры`

	l := NewLexer(input)
	tokens := l.Tokenize()

	expectedTypes := []TokenType{
		TokenProcedure,
		TokenIdent,
		TokenLParen,
		TokenIdent,
		TokenComma,
		TokenIdent,
		TokenRParen,
		TokenExport,
		TokenComment,
		TokenReturn,
		TokenEndProcedure,
		TokenEOF,
	}

	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d", len(expectedTypes), len(tokens))
	}

	for i, tt := range expectedTypes {
		if tokens[i].Type != tt {
			t.Errorf("token[%d]: expected %s, got %s (%q)",
				i, tt, tokens[i].Type, tokens[i].Literal)
		}
	}
}
