package analysis

import (
	"strings"

	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func FormatDocument(text string, tabSize int, insertSpaces bool) string {
	p := parser.NewParser(text)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		return text
	}

	var out strings.Builder
	indent := 0
	indentStr := "\t"
	if insertSpaces {
		indentStr = strings.Repeat(" ", tabSize)
	}

	for _, stmt := range mod.Statements {
		printModuleStmt(stmt, &indent, &out, indentStr)
	}

	result := out.String()
	if result == "" {
		return text
	}
	return result
}

func printModuleStmt(n parser.Node, indent *int, out *strings.Builder, indentStr string) {
	switch n := n.(type) {
	case *parser.Procedure:
		out.WriteString(strings.Repeat(indentStr, *indent))
		out.WriteString("Процедура ")
		out.WriteString(n.Name)
		out.WriteString("(")
		for i, p := range n.Params {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(p.Name)
		}
		out.WriteString(")")
		if n.Export {
			out.WriteString(" Экспорт")
		}
		out.WriteString("\n")
		*indent++
		printStmtList(n.Body, indent, out, indentStr)
		*indent--
		out.WriteString(strings.Repeat(indentStr, *indent))
		out.WriteString("КонецПроцедуры\n\n")

	case *parser.Function:
		out.WriteString(strings.Repeat(indentStr, *indent))
		out.WriteString("Функция ")
		out.WriteString(n.Name)
		out.WriteString("(")
		for i, p := range n.Params {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(p.Name)
		}
		out.WriteString(")")
		if n.Export {
			out.WriteString(" Экспорт")
		}
		out.WriteString("\n")
		*indent++
		printStmtList(n.Body, indent, out, indentStr)
		*indent--
		out.WriteString(strings.Repeat(indentStr, *indent))
		out.WriteString("КонецФункции\n\n")

	case *parser.VarDeclExpr:
		out.WriteString(strings.Repeat(indentStr, *indent))
		out.WriteString("Перем ")
		out.WriteString(n.Name)
		out.WriteString(";\n")

	case *parser.RegionBlock:
		out.WriteString(strings.Repeat(indentStr, *indent))
		out.WriteString("#Область ")
		out.WriteString(n.Name)
		out.WriteString("\n")
		*indent++
		printStmtList(n.Body, indent, out, indentStr)
		*indent--
		out.WriteString(strings.Repeat(indentStr, *indent))
		out.WriteString("#КонецОбласти\n\n")

	case *parser.HashIfBlock:
		out.WriteString(strings.Repeat(indentStr, *indent))
		out.WriteString("#Если ")
		out.WriteString(n.Condition)
		out.WriteString(" Тогда\n")
		*indent++
		printStmtList(n.Body, indent, out, indentStr)
		*indent--
		for _, ei := range n.ElseIf {
			out.WriteString(strings.Repeat(indentStr, *indent))
			out.WriteString("#ИначеЕсли ")
			out.WriteString(ei.Condition)
			out.WriteString(" Тогда\n")
			*indent++
			printStmtList(ei.Body, indent, out, indentStr)
			*indent--
		}
		if len(n.ElseBody) > 0 {
			out.WriteString(strings.Repeat(indentStr, *indent))
			out.WriteString("#Иначе\n")
			*indent++
			printStmtList(n.ElseBody, indent, out, indentStr)
			*indent--
		}
		out.WriteString(strings.Repeat(indentStr, *indent))
		out.WriteString("#КонецЕсли\n\n")

	default:
		out.WriteString(strings.Repeat(indentStr, *indent))
		out.WriteString(renderExpr(n))
		out.WriteString(";\n")
	}
}

func printStmtList(stmts []parser.Node, indent *int, out *strings.Builder, indentStr string) {
	for _, stmt := range stmts {
		printStmt(stmt, *indent, out, indentStr)
	}
}

func printStmt(n parser.Node, indent int, out *strings.Builder, indentStr string) {
	switch n := n.(type) {
	case *parser.VarDeclExpr:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("Перем ")
		out.WriteString(n.Name)
		out.WriteString(";\n")

	case *parser.IfStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("Если ")
		out.WriteString(renderExpr(n.Condition))
		out.WriteString(" Тогда\n")
		indent++
		printStmtList(n.Body, &indent, out, indentStr)
		indent--
		for _, ei := range n.ElseIf {
			out.WriteString(strings.Repeat(indentStr, indent))
			out.WriteString("ИначеЕсли ")
			out.WriteString(renderExpr(ei.Condition))
			out.WriteString(" Тогда\n")
			indent++
			printStmtList(ei.Body, &indent, out, indentStr)
			indent--
		}
		if len(n.ElseBody) > 0 {
			out.WriteString(strings.Repeat(indentStr, indent))
			out.WriteString("Иначе\n")
			indent++
			printStmtList(n.ElseBody, &indent, out, indentStr)
			indent--
		}
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("КонецЕсли;\n")

	case *parser.WhileStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("Пока ")
		out.WriteString(renderExpr(n.Condition))
		out.WriteString(" Цикл\n")
		indent++
		printStmtList(n.Body, &indent, out, indentStr)
		indent--
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("КонецЦикла;\n")

	case *parser.ForStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("Для ")
		out.WriteString(n.Var)
		out.WriteString(" = ")
		out.WriteString(renderExpr(n.From))
		out.WriteString(" По ")
		out.WriteString(renderExpr(n.To))
		out.WriteString(" Цикл\n")
		indent++
		printStmtList(n.Body, &indent, out, indentStr)
		indent--
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("КонецЦикла;\n")

	case *parser.ForEachStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("Для Каждого ")
		out.WriteString(n.Var)
		out.WriteString(" Из ")
		out.WriteString(renderExpr(n.In))
		out.WriteString(" Цикл\n")
		indent++
		printStmtList(n.Body, &indent, out, indentStr)
		indent--
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("КонецЦикла;\n")

	case *parser.TryStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("Попытка\n")
		indent++
		printStmtList(n.Body, &indent, out, indentStr)
		indent--
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("Исключение\n")
		indent++
		printStmtList(n.Except, &indent, out, indentStr)
		indent--
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("КонецПопытки;\n")

	case *parser.ReturnStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		if n.Value != nil {
			out.WriteString("Возврат ")
			out.WriteString(renderExpr(n.Value))
		} else {
			out.WriteString("Возврат")
		}
		out.WriteString(";\n")

	case *parser.RaiseStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		if n.Value != nil {
			out.WriteString("ВызватьИсключение ")
			out.WriteString(renderExpr(n.Value))
		} else {
			out.WriteString("ВызватьИсключение")
		}
		out.WriteString(";\n")

	case *parser.CycleStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("Продолжить;\n")

	case *parser.BreakStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("Прервать;\n")

	case *parser.GotoStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString("Перейти ")
		out.WriteString(n.Label)
		out.WriteString(";\n")

	case *parser.AssignmentStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString(renderExpr(n.Left))
		out.WriteString(" = ")
		out.WriteString(renderExpr(n.Right))
		out.WriteString(";\n")

	case *parser.CallStmt:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString(renderExpr(n))
		out.WriteString(";\n")

	default:
		out.WriteString(strings.Repeat(indentStr, indent))
		out.WriteString(renderExpr(n))
		out.WriteString(";\n")
	}
}

func renderExpr(n parser.Node) string {
	if n == nil {
		return ""
	}
	switch n := n.(type) {
	case *parser.Ident:
		return n.Name
	case *parser.NumberLit:
		return n.Value
	case *parser.StringLit:
		return `"` + n.Value + `"`
	case *parser.DateLit:
		return "'" + n.Value + "'"
	case *parser.BoolLit:
		if n.Value {
			return "Истина"
		}
		return "Ложь"
	case *parser.UndefinedLit:
		return "Неопределено"
	case *parser.NullLit:
		return "Null"
	case *parser.BinaryExpr:
		return renderExpr(n.Left) + " " + opStr(n.Op) + " " + renderExpr(n.Right)
	case *parser.UnaryExpr:
		return opStr(n.Op) + renderExpr(n.Right)
	case *parser.TernaryExpr:
		return "?( " + renderExpr(n.Condition) + ", " + renderExpr(n.True) + ", " + renderExpr(n.False) + " )"
	case *parser.CallStmt:
		if n.Object != nil {
			return renderExpr(n.Object) + "." + n.Function + "(" + renderArgs(n.Args) + ")"
		}
		return n.Function + "(" + renderArgs(n.Args) + ")"
	case *parser.IndexExpr:
		return renderExpr(n.Object) + "[" + renderExpr(n.Index) + "]"
	case *parser.FieldAccessExpr:
		return renderExpr(n.Object) + "." + n.Field
	case *parser.NewExpr:
		return "Новый " + n.TypeName + "(" + renderArgs(n.Args) + ")"
	case *parser.ExecuteExpr:
		return "Выполнить(" + renderExpr(n.Expr) + ")"
	case *parser.AddressExpr:
		return "Адрес(" + renderExpr(n.Expr) + ")"
	case *parser.TypeExpr:
		return "Тип(" + renderExpr(n.Expr) + ")"
	case *parser.ValExpr:
		return "Значение(" + renderExpr(n.Expr) + ")"
	case *parser.VarDeclExpr:
		return "Перем " + n.Name
	default:
		return ""
	}
}

func renderArgs(args []parser.Node) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = renderExpr(a)
	}
	return strings.Join(parts, ", ")
}

func opStr(t parser.TokenType) string {
	switch t {
	case parser.TokenPlus:
		return "+"
	case parser.TokenMinus:
		return "-"
	case parser.TokenStar:
		return "*"
	case parser.TokenSlash:
		return "/"
	case parser.TokenMod:
		return "%"
	case parser.TokenEqual:
		return "="
	case parser.TokenNotEqual:
		return "<>"
	case parser.TokenLess:
		return "<"
	case parser.TokenGreater:
		return ">"
	case parser.TokenLessOrEqual:
		return "<="
	case parser.TokenGreaterOrEqual:
		return ">="
	case parser.TokenAnd:
		return "И"
	case parser.TokenOr:
		return "Или"
	case parser.TokenNot:
		return "Не"
	default:
		return "?"
	}
}


