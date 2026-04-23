package analysis

import (
	"testing"

	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func checkSymbol(t *testing.T, st *SymbolTable, name string, kind SymbolKind) *Symbol {
	t.Helper()
	sym := st.Lookup(name)
	if sym == nil {
		t.Errorf("symbol %q not found", name)
		return nil
	}
	if sym.Kind != kind {
		t.Errorf("symbol %q: expected kind %d, got %d", name, kind, sym.Kind)
	}
	return sym
}

func TestBuildSymbolTable_Procedure(t *testing.T) {
	input := `Процедура Тест()
	А = 1
КонецПроцедуры`

	p := parser.NewParser(input)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	st := BuildSymbolTable(mod)
	checkSymbol(t, st, "Тест", SymbolProcedure)
}

func TestBuildSymbolTable_Params(t *testing.T) {
	input := `Процедура Тест(Парам1, Знач Парам2) Экспорт
	А = Парам1
КонецПроцедуры`

	p := parser.NewParser(input)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	st := BuildSymbolTable(mod)
	checkSymbol(t, st, "Тест", SymbolProcedure)
	checkSymbol(t, st, "Парам1", SymbolParameter)
	checkSymbol(t, st, "Парам2", SymbolParameter)
}

func TestBuildSymbolTable_GlobalVars(t *testing.T) {
	input := `Перем ГлобальнаяПеременная Экспорт
Процедура Тест()
КонецПроцедуры`

	p := parser.NewParser(input)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	st := BuildSymbolTable(mod)
	checkSymbol(t, st, "Тест", SymbolProcedure)
	checkSymbol(t, st, "ГлобальнаяПеременная", SymbolVariable)
}

func TestBuildSymbolTable_LocalVars(t *testing.T) {
	input := `Процедура Тест()
	Перем Локальная
	Локальная = 1
КонецПроцедуры`

	p := parser.NewParser(input)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	st := BuildSymbolTable(mod)
	checkSymbol(t, st, "Тест", SymbolProcedure)
	checkSymbol(t, st, "Локальная", SymbolVariable)
}

func TestBuildSymbolTable_AutoDeclVar(t *testing.T) {
	input := `Процедура Тест()
	НоваяПеременная = 42
КонецПроцедуры`

	p := parser.NewParser(input)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	st := BuildSymbolTable(mod)
	checkSymbol(t, st, "Тест", SymbolProcedure)
	checkSymbol(t, st, "НоваяПеременная", SymbolVariable)
}

func TestBuildSymbolTable_ScopeIsolation(t *testing.T) {
	input := `Процедура Внешняя()
	Если Истина Тогда
		ВнутриЕсли = 1
	КонецЕсли
	// ВнутриЕсли не должна быть видна на этом уровне
КонецПроцедуры`

	p := parser.NewParser(input)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	st := BuildSymbolTable(mod)
	checkSymbol(t, st, "Внешняя", SymbolProcedure)
	// Локальная в процедуре — не должна быть глобальной
	checkSymbol(t, st, "ВнутриЕсли", SymbolVariable)
}

func TestBuildSymbolTable_GlobalVarsAreGlobal(t *testing.T) {
	input := `Перем МодульнаяПеременная

Процедура Тест()
	МодульнаяПеременная = 1
КонецПроцедуры`

	p := parser.NewParser(input)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	st := BuildSymbolTable(mod)
	checkSymbol(t, st, "МодульнаяПеременная", SymbolVariable)
	checkSymbol(t, st, "Тест", SymbolProcedure)
}
