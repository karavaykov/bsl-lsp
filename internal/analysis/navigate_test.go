package analysis

import (
	"testing"

	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func parseModule(t *testing.T, input string) *parser.Module {
	t.Helper()
	p := parser.NewParser(input)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	return mod
}

func TestFindIdentAtPos_Simple(t *testing.T) {
	input := "Процедура Тест()\n\tА = 1\n\tБ = А\nКонецПроцедуры"
	mod := parseModule(t, input)

	ident := FindIdentAtPos(mod, 3, 6)
	if ident == nil {
		t.Fatal("expected to find ident at (3, 6)")
	}
	if ident.Name != "А" {
		t.Errorf("expected 'А', got %q", ident.Name)
	}
}

func TestFindIdentAtPos_NotFound(t *testing.T) {
	input := "Процедура Тест()\n\tА = 1\nКонецПроцедуры"
	mod := parseModule(t, input)

	ident := FindIdentAtPos(mod, 1, 5)
	if ident != nil {
		t.Errorf("expected nil, got %q", ident.Name)
	}
}

func TestFindIdentAtPos_CallArg(t *testing.T) {
	input := "Процедура Тест()\n\tСообщить(А)\nКонецПроцедуры"
	mod := parseModule(t, input)

	ident := FindIdentAtPos(mod, 2, 11)
	if ident == nil {
		t.Fatal("expected to find ident at (2, 11)")
	}
	if ident.Name != "А" {
		t.Errorf("expected 'А', got %q", ident.Name)
	}
}

func TestFindIdentAtPos_BinaryExpr(t *testing.T) {
	input := "Процедура Тест()\n\tА = Б + В\nКонецПроцедуры"
	mod := parseModule(t, input)

	ident := FindIdentAtPos(mod, 2, 6)
	if ident == nil {
		t.Fatal("expected to find ident at (2, 6)")
	}
	if ident.Name != "Б" {
		t.Errorf("expected 'Б', got %q", ident.Name)
	}
}

func TestFindIdentAtPos_IfCondition(t *testing.T) {
	input := "Процедура Тест()\n\tЕсли А = 1 Тогда\n\t\tБ = 2\n\tКонецЕсли\nКонецПроцедуры"
	mod := parseModule(t, input)

	ident := FindIdentAtPos(mod, 2, 7)
	if ident == nil {
		t.Fatal("expected to find ident at (2, 7)")
	}
	if ident.Name != "А" {
		t.Errorf("expected 'А', got %q", ident.Name)
	}
}

func TestFindIdentAtPos_FieldAccess(t *testing.T) {
	input := "Процедура Тест()\n\tА = Объект.Свойство\nКонецПроцедуры"
	mod := parseModule(t, input)

	ident := FindIdentAtPos(mod, 2, 6)
	if ident == nil {
		t.Fatal("expected to find ident at (2, 6)")
	}
	if ident.Name != "Объект" {
		t.Errorf("expected 'Объект', got %q", ident.Name)
	}
}

func TestFindSymbolAtPos(t *testing.T) {
	input := "Перем Глобал\nПроцедура Тест()\n\tЛокальная = 1\n\tГлобал = Локальная\nКонецПроцедуры"
	mod := parseModule(t, input)
	st := BuildSymbolTable(mod)

	sym := FindSymbolAtPos(st, 1, 7)
	if sym == nil {
		t.Fatal("expected to find symbol at (1, 7)")
	}
	if sym.Name != "Глобал" {
		t.Errorf("expected 'Глобал', got %q", sym.Name)
	}
}

func TestFindDefinition(t *testing.T) {
	input := "Перем Глобал\nПроцедура Тест()\n\tГлобал = 1\nКонецПроцедуры"
	mod := parseModule(t, input)
	st := BuildSymbolTable(mod)

	sym := FindDefinition(st, "Глобал")
	if sym == nil {
		t.Fatal("expected to find 'Глобал'")
	}
	if sym.Kind != SymbolVariable {
		t.Errorf("expected Variable, got %d", sym.Kind)
	}
}

func TestFindIdentAtPos_ForEachVar(t *testing.T) {
	input := "Процедура Тест()\n\tДля Каждого Эл Из Коллекция Цикл\n\t\tСообщить(Эл)\n\tКонецЦикла\nКонецПроцедуры"
	mod := parseModule(t, input)

	ident := FindIdentAtPos(mod, 3, 12)
	if ident == nil {
		t.Fatal("expected to find ident at (3, 12)")
	}
	if ident.Name != "Эл" {
		t.Errorf("expected 'Эл', got %q", ident.Name)
	}
}
