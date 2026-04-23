package parser

type TokenType int

const (
	TokenIllegal TokenType = iota
	TokenEOF

	// Literals
	TokenIdent
	TokenNumber
	TokenString
	TokenDate

	// Keywords
	TokenProcedure
	TokenFunction
	TokenEndProcedure
	TokenEndFunction
	TokenIf
	TokenThen
	TokenElseIf
	TokenElse
	TokenEndIf
	TokenFor
	TokenTo
	TokenWhile
	TokenDo
	TokenEndDo
	TokenForEach
	TokenIn
	TokenCycle
	TokenBreak
	TokenTry
	TokenExcept
	TokenEndTry
	TokenReturn
	TokenContinue
	TokenRaise
	TokenVar
	TokenExport
	TokenVal
	TokenExecute
	TokenAddress
	TokenNew
	TokenNot
	TokenAnd
	TokenOr
	TokenTrue
	TokenFalse
	TokenUndefined
	TokenNull
	TokenTypeName
	TokenGoTo
	TokenIfGoto

	// Preprocessor
	TokenHashIf
	TokenHashElseIf
	TokenHashElse
	TokenHashEndIf
	TokenHashRegion
	TokenHashEndRegion
	TokenHashInsert
	TokenHashDelete
	TokenHashArea

	// Compiler directives
	TokenAmpersandAtClient
	TokenAmpersandAtServer
	TokenAmpersandAtServerNoContext
	TokenAmpersandAtClientAtServer
	TokenAmpersandAtClientAtServerNoContext

	// Operators
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenMod
	TokenPower
	TokenEqual
	TokenNotEqual
	TokenLess
	TokenGreater
	TokenLessOrEqual
	TokenGreaterOrEqual
	TokenAssign
	TokenQuestion
	TokenDot
	TokenComma
	TokenSemicolon
	TokenColon

	// Brackets
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket

	// Special
	TokenTilde

	// Comment
	TokenComment
	TokenPreprocessor
	TokenDirective
)

var tokenNames = map[TokenType]string{
	TokenIllegal:                      "ILLEGAL",
	TokenEOF:                          "EOF",
	TokenIdent:                        "IDENTIFIER",
	TokenNumber:                       "NUMBER",
	TokenString:                       "STRING",
	TokenDate:                         "DATE",
	TokenProcedure:                    "ПРОЦЕДУРА",
	TokenFunction:                     "ФУНКЦИЯ",
	TokenEndProcedure:                 "КОНЕЦПРОЦЕДУРЫ",
	TokenEndFunction:                  "КОНЕЦФУНКЦИИ",
	TokenIf:                           "ЕСЛИ",
	TokenThen:                         "ТОГДА",
	TokenElseIf:                       "ИНАЧЕЕСЛИ",
	TokenElse:                         "ИНАЧЕ",
	TokenEndIf:                        "КОНЕЦЕСЛИ",
	TokenFor:                          "ДЛЯ",
	TokenTo:                           "ПО",
	TokenWhile:                        "ПОКА",
	TokenDo:                           "ЦИКЛ",
	TokenEndDo:                        "КОНЕЦЦИКЛА",
	TokenForEach:                      "ДЛЯКАЖДОГО",
	TokenIn:                           "ИЗ",
	TokenCycle:                        "ПРОДОЛЖИТЬ",
	TokenBreak:                        "ПРЕРВАТЬ",
	TokenTry:                          "ПОПЫТКА",
	TokenExcept:                       "ИСКЛЮЧЕНИЕ",
	TokenEndTry:                       "КОНЕЦПОПЫТКИ",
	TokenReturn:                       "ВОЗВРАТ",
	TokenContinue:                     "CONTINUE",
	TokenRaise:                        "ВЫЗВАТЬИСКЛЮЧЕНИЕ",
	TokenVar:                          "ПЕРЕМ",
	TokenExport:                       "ЭКСПОРТ",
	TokenVal:                          "ЗНАЧ",
	TokenExecute:                      "ВЫПОЛНИТЬ",
	TokenAddress:                      "АДРЕС",
	TokenNew:                          "НОВЫЙ",
	TokenNot:                          "НЕ",
	TokenAnd:                          "И",
	TokenOr:                           "ИЛИ",
	TokenTrue:                         "ИСТИНА",
	TokenFalse:                        "ЛОЖЬ",
	TokenUndefined:                    "НЕОПРЕДЕЛЕНО",
	TokenNull:                         "NULL",
	TokenTypeName:                     "ТИП",
	TokenGoTo:                         "ПЕРЕЙТИ",
	TokenIfGoto:                       "ПЕРЕЙТИЕСЛИ",
	TokenHashIf:                       "#ЕСЛИ",
	TokenHashElseIf:                   "#ИНАЧЕЕСЛИ",
	TokenHashElse:                     "#ИНАЧЕ",
	TokenHashEndIf:                    "#КОНЕЦЕСЛИ",
	TokenHashRegion:                   "#ОБЛАСТЬ",
	TokenHashEndRegion:                "#КОНЕЦОБЛАСТИ",
	TokenHashInsert:                   "#ВСТАВИТЬ",
	TokenHashDelete:                   "#УДАЛИТЬ",
	TokenHashArea:                     "#ОБЛАСТЬ",
	TokenAmpersandAtClient:            "&НаКлиенте",
	TokenAmpersandAtServer:            "&НаСервере",
	TokenAmpersandAtServerNoContext:   "&НаСервереБезКонтекста",
	TokenAmpersandAtClientAtServer:    "&НаКлиентеНаСервере",
	TokenAmpersandAtClientAtServerNoContext: "&НаКлиентеНаСервереБезКонтекста",
	TokenPlus:                         "+",
	TokenMinus:                        "-",
	TokenStar:                         "*",
	TokenSlash:                        "/",
	TokenMod:                          "%",
	TokenPower:                        "^",
	TokenEqual:                        "=",
	TokenNotEqual:                     "<>",
	TokenLess:                         "<",
	TokenGreater:                      ">",
	TokenLessOrEqual:                  "<=",
	TokenGreaterOrEqual:               ">=",
	TokenAssign:                       ":=",
	TokenQuestion:                     "?",
	TokenDot:                          ".",
	TokenComma:                        ",",
	TokenSemicolon:                    ";",
	TokenColon:                        ":",
	TokenLParen:                       "(",
	TokenRParen:                       ")",
	TokenLBracket:                     "[",
	TokenRBracket:                     "]",
	TokenComment:                      "//",
	TokenPreprocessor:                 "PREPROCESSOR",
	TokenDirective:                    "DIRECTIVE",
	TokenTilde:                        "~",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

func NewToken(typ TokenType, literal string, line, col int) Token {
	return Token{
		Type:    typ,
		Literal: literal,
		Line:    line,
		Col:     col,
	}
}

var keywordsRu = map[string]TokenType{
	"процедура":         TokenProcedure,
	"функция":           TokenFunction,
	"конецпроцедуры":    TokenEndProcedure,
	"конецфункции":      TokenEndFunction,
	"если":              TokenIf,
	"тогда":             TokenThen,
	"иначеесли":         TokenElseIf,
	"иначе":             TokenElse,
	"конецесли":         TokenEndIf,
	"для":               TokenFor,
	"по":                TokenTo,
	"пока":              TokenWhile,
	"цикл":              TokenDo,
	"конеццикла":        TokenEndDo,
	"длякаждого":        TokenForEach,
	"из":                TokenIn,
	"продолжить":        TokenCycle,
	"прервать":          TokenBreak,
	"попытка":           TokenTry,
	"исключение":        TokenExcept,
	"конецпопытки":      TokenEndTry,
	"возврат":           TokenReturn,
	"вызватьисключение": TokenRaise,
	"перем":             TokenVar,
	"экспорт":           TokenExport,
	"знач":              TokenVal,
	"выполнить":         TokenExecute,
	"адрес":             TokenAddress,
	"новый":             TokenNew,
	"не":                TokenNot,
	"и":                 TokenAnd,
	"или":               TokenOr,
	"истина":            TokenTrue,
	"ложь":              TokenFalse,
	"неопределено":      TokenUndefined,
	"null":              TokenNull,
	"тип":               TokenTypeName,
	"перейти":           TokenGoTo,
	"перейтиесли":       TokenIfGoto,
}

var directives = map[string]TokenType{
	"&наклиенте":                      TokenAmpersandAtClient,
	"&насервере":                      TokenAmpersandAtServer,
	"&насерверебезконтекста":          TokenAmpersandAtServerNoContext,
	"&наклиентенасервере":             TokenAmpersandAtClientAtServer,
	"&наклиентенасерверебезконтекста": TokenAmpersandAtClientAtServerNoContext,
}

func LookupIdent(ident string) TokenType {
	lower := toLower(ident)
	if t, ok := keywordsRu[lower]; ok {
		return t
	}
	return TokenIdent
}

func LookupDirective(ident string) TokenType {
	lower := toLower(ident)
	if t, ok := directives[lower]; ok {
		return t
	}
	return TokenDirective
}

func toLower(s string) string {
	var result []rune
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else if r >= 'А' && r <= 'Я' {
			result = append(result, r+32)
		} else if r == 'Ё' {
			result = append(result, 'ё')
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func equalsFold(a, b string) bool {
	return toLower(a) == toLower(b)
}
