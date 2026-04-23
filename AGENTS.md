# bsl-lsp

LSP сервер для языка 1С (BSL). Чистый Go, zero external dependencies.

## Команды

```bash
go build -o bsl-lsp ./cmd/bsl-lsp   # бинарник
go test -v ./...                    # все тесты (включая интеграционные на real .bsl)
go vet ./...                        # статический анализ
```

Либо через `Makefile`: `make build`, `make test`, `make vet`, `make run`, `make clean`.

## Архитектура

- **`cmd/bsl-lsp/main.go`** — точка входа, вызывает `lsp.Run()`
- **`internal/lsp/`** — JSON-RPC 2.0 через stdin/stdout (`Content-Length` заголовки), session, handler
- **`internal/parser/`** — лексер (rune-based, UTF-8, полная кириллица) + рекурсивный спуск + AST
- **`internal/analysis/`**:
  - `symbol.go` — symbol table: `BuildSymbolTable(mod)` обходит AST, строит области видимости
  - `navigate.go` — `FindIdentAtPos(mod, line, col)` ищет идентификатор в AST по позиции (GoToDefinition, Hover)
  - `keywords.go` — `BSLKeywords` (34 keyword), `BSLGlobalMethods` (глобальные функции 1С)
- **`internal/workspace/document.go`** — thread-safe Document + Manager
- **`pkg/protocol/types.go`** — LSP типы

## LSP протокол

- Диагностики публикуются после каждого `didOpen`/`didChange` через `textDocument/publishDiagnostics`
- Ошибки парсера конвертируются в 0-based LSP позиции
- Неизвестные методы возвращают JSON-RPC error `-32601`
- `textDocument/didSave` — no-op
- Symbol table перестраивается при каждом didOpen/didChange (хранится в `Handler.documents[uri]`)
- `textDocument/definition` — `FindIdentAtPos` + `Lookup` → Location
- `textDocument/hover` — `FindIdentAtPos` + `Lookup` → markdown (kind + name)
- `textDocument/completion` — keywords + global methods + symbols + dot-context completion

## Парсер

Разбирает: процедуры/функции (параметры, Знач, Экспорт), If/ElseIf/Else, While, For/To, ForEach/In, Try/Except, Return/Raise, Goto, Break/Continue, присваивания, вызовы, бинарные/унарные/тернарные выражения, field access/index, New/Execute/Address/Type/Val, `#Если`/`#Область`, директивы компиляции, комментарии. Error recovery через `syncToStmt`.

## Тестирование

- Модульные тесты: `TestLexer_*`, `TestParser_*`
- Интеграционные: `TestRealBSLParseFiles` — парсит все `.bsl` файлы из `internal/parser/testdata/real_bsl/`
- Symbol table: `TestBuildSymbolTable_*` в `internal/analysis/symbol_test.go`
- Тесты в `package parser` (white-box)

## Соглашения

- Нет внешних Go-зависимостей (только stdlib)
- `skipComments()` пропускает комментарии и препроцессор при разборе statement'ов
- `consumeSemicolon` в `parseBlock`, `nextToken` в Break/Cycle
- AST узлы Procedure/Function/If/While/For/ForEach/Try хранят `Line`/`Col` позиции (исправлено, не через `Directives[0]`)
