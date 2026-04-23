package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealProjectBSLFiles(t *testing.T) {
	root := `/mnt/c/Users/karavaikov.s/AI 1C`

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Errorf("walk error for %s: %v", path, err)
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".bsl") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read %s: %v", path, err)
			return nil
		}

		rel, _ := filepath.Rel(root, path)

		t.Run(rel, func(t *testing.T) {
			p := NewParser(string(data))
			p.ParseModule()
			errs := p.Errors()
			if len(errs) > 0 {
				for _, e := range errs {
					t.Errorf("%s: line %d:%d: %s", rel, e.Line, e.Col, e.Message)
				}
			} else {
				t.Logf("%s: OK", rel)
			}
		})

		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
}

func TestRealProjectBSLFilesDetailed(t *testing.T) {
	files := []string{
		`Catalogs/АР_ОбъектыАренды/Ext/ManagerModule.bsl`,
		`Catalogs/АР_ОбъектыАренды/Ext/ObjectModule.bsl`,
		`Catalogs/АР_ОбъектыАренды/Forms/ФормаЭлемента/Ext/Form/Module.bsl`,
		`Catalogs/ДоговорыКонтрагентов/Forms/ФормаЭлемента/Ext/Form/Module.bsl`,
		`CommonModules/ДФИ/Ext/Module.bsl`,
		`CommonModules/ИнтеграцияGLPI/Ext/Module.bsl`,
		`CommonModules/ПовторяющиесяМетодыНаСервере/Ext/Module.bsl`,
		`DataProcessors/ПроверкаGLPI/Forms/Форма/Ext/Form/Module.bsl`,
		`DataProcessors/ЭФ_ЗагрузкаДанныхСдачиОтчетности/Forms/Форма/Ext/Form/Module.bsl`,
		`DataProcessors/ЭФ_ФормулаЭффективности/Forms/Форма/Ext/Form/Module.bsl`,
		`Documents/ЭФ_ЧекЛистПоУА/Ext/ObjectModule.bsl`,
		`Documents/ЭФ_ЧекЛистПоУА/Forms/ФормаДокумента/Ext/Form/Module.bsl`,
		`Ext/SessionModule.bsl`,
		`InformationRegisters/ЭФ_ЗадачиGLPI/Forms/ФормаЗаписи/Ext/Form/Module.bsl`,
		`InformationRegisters/ЭФ_ЗадачиGLPI/Forms/ФормаСписка/Ext/Form/Module.bsl`,
	}

	root := `/mnt/c/Users/karavaikov.s/AI 1C`

	for _, f := range files {
		f := f
		t.Run(f, func(t *testing.T) {
			path := root + `/` + f
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}

			p := NewParser(string(data))
			p.ParseModule()
			errs := p.Errors()
			if len(errs) > 0 {
				for _, e := range errs {
					t.Errorf("line %d:%d: %s", e.Line, e.Col, e.Message)
				}
			}
		})
	}
}
