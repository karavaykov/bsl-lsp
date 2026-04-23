package analysis

type ModuleSymbols struct {
	URI     string
	Table   *SymbolTable
	Exports map[string]*Symbol
}

func CollectExports(st *SymbolTable, uri string) *ModuleSymbols {
	ms := &ModuleSymbols{
		URI:     uri,
		Table:   st,
		Exports: make(map[string]*Symbol),
	}
	for _, sym := range st.Symbols {
		if sym.Export && sym.Scope == st.Global {
			ms.Exports[sym.Name] = sym
		}
	}
	return ms
}

type ProjectAnalysis struct {
	Modules map[string]*ModuleSymbols
}

func NewProjectAnalysis() *ProjectAnalysis {
	return &ProjectAnalysis{
		Modules: make(map[string]*ModuleSymbols),
	}
}

func (pa *ProjectAnalysis) UpdateModule(uri string, st *SymbolTable) {
	pa.Modules[uri] = CollectExports(st, uri)
}

func (pa *ProjectAnalysis) RemoveModule(uri string) {
	delete(pa.Modules, uri)
}

func (pa *ProjectAnalysis) LookupSymbol(name string) (foundURI string, sym *Symbol) {
	for uri, ms := range pa.Modules {
		if s, ok := ms.Exports[name]; ok {
			return uri, s
		}
	}
	return "", nil
}
