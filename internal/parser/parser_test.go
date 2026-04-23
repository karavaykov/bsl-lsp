package parser

import (
	"testing"
)

func TestParser_SimpleProcedure(t *testing.T) {
	input := `Процедура Тест()
КонецПроцедуры`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(mod.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(mod.Statements))
	}

	proc, ok := mod.Statements[0].(*Procedure)
	if !ok {
		t.Fatalf("expected Procedure, got %T", mod.Statements[0])
	}
	if proc.Name != "Тест" {
		t.Errorf("expected name 'Тест', got %q", proc.Name)
	}
}

func TestParser_ProcedureWithParams(t *testing.T) {
	input := `Процедура Тест(Парам1, Знач Парам2) Экспорт
	А = 1
КонецПроцедуры`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	proc := mod.Statements[0].(*Procedure)
	if !proc.Export {
		t.Error("expected export=true")
	}
	if len(proc.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(proc.Params))
	}
	if proc.Params[0].Name != "Парам1" || proc.Params[0].ByVal {
		t.Errorf("param 0: expected Парам1, byVal=false")
	}
	if proc.Params[1].Name != "Парам2" || !proc.Params[1].ByVal {
		t.Errorf("param 1: expected Парам2, byVal=true")
	}
	if len(proc.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(proc.Body))
	}
}

func TestParser_SimpleFunction(t *testing.T) {
	input := `Функция Сложить(А, Б) Экспорт
	Возврат А + Б
КонецФункции`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	fn := mod.Statements[0].(*Function)
	if fn.Name != "Сложить" {
		t.Errorf("expected 'Сложить', got %q", fn.Name)
	}
	if !fn.Export {
		t.Error("expected export=true")
	}
	if len(fn.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fn.Params))
	}
}

func TestParser_IfStatement(t *testing.T) {
	input := `Если А > 0 Тогда
		Х = 1
	ИначеЕсли А = 0 Тогда
		Х = 2
	Иначе
		Х = 3
	КонецЕсли`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(mod.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(mod.Statements))
	}

	ifStmt, ok := mod.Statements[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", mod.Statements[0])
	}

	if len(ifStmt.Body) != 1 {
		t.Errorf("expected 1 body stmt, got %d", len(ifStmt.Body))
	}
	if len(ifStmt.ElseIf) != 1 {
		t.Errorf("expected 1 elseif, got %d", len(ifStmt.ElseIf))
	}
	if len(ifStmt.ElseBody) != 1 {
		t.Errorf("expected 1 else stmt, got %d", len(ifStmt.ElseBody))
	}

	// Проверяем ElseIf
	ei := ifStmt.ElseIf[0]
	if len(ei.Body) != 1 {
		t.Errorf("expected 1 elseif body stmt, got %d", len(ei.Body))
	}
}

func TestParser_WhileStatement(t *testing.T) {
	input := `Пока Индекс > 0 Цикл
		Индекс = Индекс - 1
	КонецЦикла`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	whileStmt, ok := mod.Statements[0].(*WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", mod.Statements[0])
	}

	_ = whileStmt.Condition
	if len(whileStmt.Body) != 1 {
		t.Errorf("expected 1 body stmt, got %d", len(whileStmt.Body))
	}
}

func TestParser_ForStatement(t *testing.T) {
	input := `Для Индекс = 1 По 10 Цикл
		Сумма = Сумма + Индекс
	КонецЦикла`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	forStmt, ok := mod.Statements[0].(*ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", mod.Statements[0])
	}
	if forStmt.Var != "Индекс" {
		t.Errorf("expected var 'Индекс', got %q", forStmt.Var)
	}
	if forStmt.From == nil {
		t.Error("expected From expression")
	}
	if forStmt.To == nil {
		t.Error("expected To expression")
	}
}

func TestParser_ForEachStatement(t *testing.T) {
	input := `Для Каждого Элемент Из Коллекция Цикл
		Сообщить(Элемент)
	КонецЦикла`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	fe, ok := mod.Statements[0].(*ForEachStmt)
	if !ok {
		t.Fatalf("expected ForEachStmt, got %T", mod.Statements[0])
	}
	if fe.Var != "Элемент" {
		t.Errorf("expected var 'Элемент', got %q", fe.Var)
	}
}

func TestParser_TryStatement(t *testing.T) {
	input := `Попытка
		А = 1 / Б
	Исключение
		А = 0
	КонецПопытки`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	tryStmt, ok := mod.Statements[0].(*TryStmt)
	if !ok {
		t.Fatalf("expected TryStmt, got %T", mod.Statements[0])
	}
	if len(tryStmt.Body) != 1 {
		t.Errorf("expected 1 body stmt, got %d", len(tryStmt.Body))
	}
	if len(tryStmt.Except) != 1 {
		t.Errorf("expected 1 except stmt, got %d", len(tryStmt.Except))
	}
}

func TestParser_ReturnStatement(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		hasValue  bool
	}{
		{"return without value", "Возврат", false},
		{"return with value", "Возврат 42", true},
		{"return with ident", "Возврат Результат", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			mod := p.ParseModule()
			if len(p.Errors()) > 0 {
				t.Fatalf("unexpected errors: %v", p.Errors())
			}
			ret, ok := mod.Statements[0].(*ReturnStmt)
			if !ok {
				t.Fatalf("expected ReturnStmt, got %T", mod.Statements[0])
			}
			if tt.hasValue && ret.Value == nil {
				t.Error("expected return value")
			}
			if !tt.hasValue && ret.Value != nil {
				t.Error("expected no return value")
			}
		})
	}
}

func TestParser_CallAndExpression(t *testing.T) {
	input := `Сообщить("Hello")`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	call, ok := mod.Statements[0].(*CallStmt)
	if !ok {
		t.Fatalf("expected CallStmt, got %T", mod.Statements[0])
	}
	if call.Function != "Сообщить" {
		t.Errorf("expected 'Сообщить', got %q", call.Function)
	}
	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}
}

func TestParser_Assignment(t *testing.T) {
	input := `Х = Х + 1`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	assign, ok := mod.Statements[0].(*AssignmentStmt)
	if !ok {
		t.Fatalf("expected AssignmentStmt, got %T", mod.Statements[0])
	}

	ident, ok := assign.Left.(*Ident)
	if !ok {
		t.Fatalf("expected Ident on left, got %T", assign.Left)
	}
	if ident.Name != "Х" {
		t.Errorf("expected 'Х', got %q", ident.Name)
	}

	bin, ok := assign.Right.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr on right, got %T", assign.Right)
	}
	if bin.Op != TokenPlus {
		t.Errorf("expected + op, got %s", bin.Op)
	}
}

func TestParser_BinaryPrecedence(t *testing.T) {
	input := `А + Б * В`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	// А + (Б * В)
	bin, ok := mod.Statements[0].(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", mod.Statements[0])
	}
	if bin.Op != TokenPlus {
		t.Errorf("expected + at top, got %s", bin.Op)
	}

	// Правая часть — Б * В
	right, ok := bin.Right.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr on right, got %T", bin.Right)
	}
	if right.Op != TokenStar {
		t.Errorf("expected * on right, got %s", right.Op)
	}
}

func TestParser_NewExpression(t *testing.T) {
	input := `Объект = Новый Структура("Ключ", Значение)`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	assign := mod.Statements[0].(*AssignmentStmt)
	newExpr, ok := assign.Right.(*NewExpr)
	if !ok {
		t.Fatalf("expected NewExpr, got %T", assign.Right)
	}
	if newExpr.TypeName != "Структура" {
		t.Errorf("expected 'Структура', got %q", newExpr.TypeName)
	}
	if len(newExpr.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(newExpr.Args))
	}
}

func TestParser_FieldAccess(t *testing.T) {
	input := `Результат = Объект.Свойство.Метод(Парам)`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	assign := mod.Statements[0].(*AssignmentStmt)

	// Должен быть: (Объект.Свойство).Метод(Парам)
	call, ok := assign.Right.(*CallStmt)
	if !ok {
		t.Fatalf("expected CallStmt, got %T", assign.Right)
	}
	if call.Function != "Метод" {
		t.Errorf("expected method 'Метод', got %q", call.Function)
	}
	if len(call.Args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(call.Args))
	}

	fa, ok := call.Args[0].(*Ident)
	if !ok || fa.Name != "Парам" {
		t.Errorf("expected arg 'Парам', got %T %+v", call.Args[0], call.Args[0])
	}
}

func TestParser_EmptyModule(t *testing.T) {
	input := ""
	p := NewParser(input)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(mod.Statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(mod.Statements))
	}
}

func TestParser_ProcedureWithBody(t *testing.T) {
	input := `Процедура Выполнить()
	Если Условие Тогда
		Для Сч = 1 По 10 Цикл
			Попытка
				Вызов()
			Исключение
				Запись = Ложь
			КонецПопытки
		КонецЦикла
	КонецЕсли
КонецПроцедуры`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	proc := mod.Statements[0].(*Procedure)
	if len(proc.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(proc.Body))
	}

	ifStmt, ok := proc.Body[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", proc.Body[0])
	}
	if len(ifStmt.Body) != 1 {
		t.Fatalf("expected 1 if body stmt, got %d", len(ifStmt.Body))
	}

	forStmt, ok := ifStmt.Body[0].(*ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", ifStmt.Body[0])
	}
	if len(forStmt.Body) != 1 {
		t.Fatalf("expected 1 for body stmt, got %d", len(forStmt.Body))
	}

	tryStmt, ok := forStmt.Body[0].(*TryStmt)
	if !ok {
		t.Fatalf("expected TryStmt, got %T", forStmt.Body[0])
	}
	if len(tryStmt.Body) != 1 {
		t.Errorf("expected 1 try body, got %d", len(tryStmt.Body))
	}
	if len(tryStmt.Except) != 1 {
		t.Errorf("expected 1 except body, got %d", len(tryStmt.Except))
	}
}

func TestParser_ErrorRecovery(t *testing.T) {
	// Незавершённая конструкция — нет Тогда после Если
	input := `Процедура Тест()
	Если 1
		А = 1;
	КонецЕсли;
	Б = 2;
КонецПроцедуры`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) == 0 {
		t.Error("expected errors, got none")
	}
	if len(mod.Statements) != 1 {
		t.Fatalf("expected 1 procedure, got %d", len(mod.Statements))
	}
	proc := mod.Statements[0].(*Procedure)
	if len(proc.Body) < 2 {
		t.Fatalf("expected at least 2 body stmts (IfStmt + Б=2), got %d", len(proc.Body))
	}
}

func TestParser_ErrorRecovery_MultipleErrors(t *testing.T) {
	// Бинарный оператор без правого операнда: `А = 1 + ;`
	input := `Процедура Тест()
	А = 1 + ;
	Б = 1;
	В = 1 * ;
	Г = 2;
КонецПроцедуры`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) == 0 {
		t.Error("expected errors, got none")
	}
	if len(mod.Statements) != 1 {
		t.Fatalf("expected 1 procedure, got %d", len(mod.Statements))
	}
	proc := mod.Statements[0].(*Procedure)
	if len(proc.Body) < 2 {
		t.Fatalf("expected at least 2 body stmts, got %d", len(proc.Body))
	}
}

func TestParser_IndexAccess(t *testing.T) {
	input := `Значение = Массив[0]`

	p := NewParser(input)
	mod := p.ParseModule()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	assign := mod.Statements[0].(*AssignmentStmt)
	_ = assign.Right.(*IndexExpr)
}
