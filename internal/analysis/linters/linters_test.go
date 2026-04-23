package linters

import (
	"testing"

	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func parseAndBuild(t *testing.T, input string) (*parser.Module, *analysis.SymbolTable) {
	t.Helper()
	p := parser.NewParser(input)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	st := analysis.BuildSymbolTable(mod)
	return mod, st
}

func countByCode(diags []LintDiagnostic, code string) int {
	n := 0
	for _, d := range diags {
		if d.Code == code {
			n++
		}
	}
	return n
}

// --- unused-variable ---

func TestUnusedVariable_Diagnosed(t *testing.T) {
	input := `Процедура Тест()
	Перем А
	А = 1
	Перем Б
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkUnusedVariable(mod, st)
	if n := countByCode(diags, "unused-variable"); n != 1 {
		t.Errorf("expected 1 unused variable diag (Б), got %d", n)
	}
}

func TestUnusedVariable_AllUsed(t *testing.T) {
	input := `Процедура Тест()
	А = 1
	Б = А + 1
	Сообщить(Б)
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkUnusedVariable(mod, st)
	if n := countByCode(diags, "unused-variable"); n != 0 {
		t.Errorf("expected 0 unused variable diags, got %d", n)
	}
}

// --- empty-block ---

func TestEmptyBlock_Diagnosed(t *testing.T) {
	input := `Процедура Тест()
	Если Истина Тогда
	КонецЕсли
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkEmptyBlock(mod, st)
	if n := countByCode(diags, "empty-block"); n < 1 {
		t.Errorf("expected at least 1 empty-block diag, got %d", n)
	}
}

func TestEmptyBlock_NonEmpty(t *testing.T) {
	input := `Процедура Тест()
	А = 1
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkEmptyBlock(mod, st)
	if n := countByCode(diags, "empty-block"); n != 0 {
		t.Errorf("expected 0 empty-block diags, got %d", n)
	}
}

// --- unreachable-code ---

func TestUnreachableCode_Diagnosed(t *testing.T) {
	input := `Процедура Тест()
	Возврат;
	А = 1
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkUnreachableCode(mod, st)
	if n := countByCode(diags, "unreachable-code"); n != 1 {
		t.Errorf("expected 1 unreachable-code diag, got %d", n)
	}
}

func TestUnreachableCode_NoUnreachable(t *testing.T) {
	input := `Процедура Тест()
	А = 1
	Возврат
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkUnreachableCode(mod, st)
	if n := countByCode(diags, "unreachable-code"); n != 0 {
		t.Errorf("expected 0 unreachable-code diags, got %d", n)
	}
}

// --- magic-number ---

func TestMagicNumber_Diagnosed(t *testing.T) {
	input := `Процедура Тест()
	А = 42
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkMagicNumber(mod, st)
	if n := countByCode(diags, "magic-number"); n != 1 {
		t.Errorf("expected 1 magic-number diag, got %d", n)
	}
}

func TestMagicNumber_SmallNumbers(t *testing.T) {
	input := `Процедура Тест()
	А = 0
	Б = 1
	В = 2
	Г = 3
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkMagicNumber(mod, st)
	if n := countByCode(diags, "magic-number"); n != 0 {
		t.Errorf("expected 0 magic-number diags for small numbers, got %d", n)
	}
}

// --- too-many-params ---

func TestTooManyParams_Diagnosed(t *testing.T) {
	input := `Процедура Тест(П1, П2, П3, П4, П5, П6, П7, П8)
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkTooManyParams(mod, st)
	if n := countByCode(diags, "too-many-params"); n != 1 {
		t.Errorf("expected 1 too-many-params diag, got %d", n)
	}
}

func TestTooManyParams_Ok(t *testing.T) {
	input := `Процедура Тест(П1, П2, П3, П4, П5, П6, П7)
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkTooManyParams(mod, st)
	if n := countByCode(diags, "too-many-params"); n != 0 {
		t.Errorf("expected 0 too-many-params diags, got %d", n)
	}
}

// --- nested-depth ---

func TestNestedDepth_Diagnosed(t *testing.T) {
	input := `Процедура Тест()
	Если 1 = 1 Тогда
		Если 2 = 2 Тогда
			Если 3 = 3 Тогда
				Если 4 = 4 Тогда
					Если 5 = 5 Тогда
						Если 6 = 6 Тогда
							А = 1
						КонецЕсли
					КонецЕсли
				КонецЕсли
			КонецЕсли
		КонецЕсли
	КонецЕсли
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkNestedDepth(mod, st)
	if n := countByCode(diags, "nested-depth"); n < 1 {
		t.Errorf("expected at least 1 nested-depth diag, got %d", n)
	}
}

func TestNestedDepth_Shallow(t *testing.T) {
	input := `Процедура Тест()
	Если 1 = 1 Тогда
		Если 2 = 2 Тогда
			А = 1
		КонецЕсли
	КонецЕсли
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkNestedDepth(mod, st)
	if n := countByCode(diags, "nested-depth"); n != 0 {
		t.Errorf("expected 0 nested-depth diags, got %d", n)
	}
}

// --- suspicious-assignment ---

func TestSuspiciousAssignment_Diagnosed(t *testing.T) {
	input := `Процедура Тест()
	А = А
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkSuspiciousAssignment(mod, st)
	if n := countByCode(diags, "suspicious-assignment"); n != 1 {
		t.Errorf("expected 1 suspicious-assignment diag, got %d", n)
	}
}

func TestSuspiciousAssignment_Normal(t *testing.T) {
	input := `Процедура Тест()
	А = 1
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkSuspiciousAssignment(mod, st)
	if n := countByCode(diags, "suspicious-assignment"); n != 0 {
		t.Errorf("expected 0 suspicious-assignment diags, got %d", n)
	}
}

// --- missing-return ---

func TestMissingReturn_Diagnosed(t *testing.T) {
	input := `Функция Тест()
	Если Истина Тогда
		Возврат 1
	КонецЕсли
КонецФункции`
	mod, st := parseAndBuild(t, input)
	diags := checkMissingReturn(mod, st)
	if n := countByCode(diags, "missing-return"); n != 1 {
		t.Errorf("expected 1 missing-return diag, got %d", n)
	}
}

func TestMissingReturn_HasReturn(t *testing.T) {
	input := `Функция Тест()
	Возврат 1
КонецФункции`
	mod, st := parseAndBuild(t, input)
	diags := checkMissingReturn(mod, st)
	if n := countByCode(diags, "missing-return"); n != 0 {
		t.Errorf("expected 0 missing-return diags, got %d", n)
	}
}

// --- global-var-in-proc ---

func TestGlobalVarInProc_Diagnosed(t *testing.T) {
	input := `Перем Глобал

Процедура Тест()
	Глобал = 1
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkGlobalVarInProc(mod, st)
	if n := countByCode(diags, "global-var-in-proc"); n != 1 {
		t.Errorf("expected 1 global-var-in-proc diag, got %d", n)
	}
}

func TestGlobalVarInProc_LocalOnly(t *testing.T) {
	input := `Процедура Тест()
	Локальная = 1
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := checkGlobalVarInProc(mod, st)
	if n := countByCode(diags, "global-var-in-proc"); n != 0 {
		t.Errorf("expected 0 global-var-in-proc diags, got %d", n)
	}
}

// --- RunAll smoke test ---

func TestRunAll_NoPanic(t *testing.T) {
	input := `Процедура Тест()
	Сообщить("привет")
	Возврат
КонецПроцедуры`
	mod, st := parseAndBuild(t, input)
	diags := RunAll(mod, st)
	// just ensure no panic — nil or empty slice both acceptable
	_ = diags
}
