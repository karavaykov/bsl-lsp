# bsl-lsp

> Go находится в `~/go/bin/go` (go1.21.13). Перед запуском Go-команд экспортируй PATH: `export PATH="$HOME/go/bin:$PATH"`
>
> **Правило:** после завершения каждого пункта дорожной карты (ROADMAP.md) запускай `go build ./... && go test ./... && go vet ./...` для верификации. Перед запуском — **спрашивай подтверждение у пользователя**.

> **Важные решения:**
> - `bug_comment_after_stmt.bsl` → `module_comment_stmt.bsl` (это корректный BSL, не баг)
> - `bug_directive_inside_proc.bsl` → `module_directive_inside_proc.bsl` (`#Область` внутри процедуры разрешена)
> - `bug_hashif_at_module_level.bsl` → `module_hashif_at_module_level.bsl` (`#Если`/`#КонецЕсли` на уровне модуля — валидный BSL)
> - `bug_nested_unclosed_blocks.bsl` → `module_nested_unclosed_blocks.bsl` (не баг парсера, проблема validator/error-recovery)
> - `module_edge.bsl`: невалидный `Для Каждого Из Таблица Цикл` удалён (BSL требует переменную цикла)
> - `TokenEqual` на `precLowest` → break, чтобы `А = 10` парсилось как `AssignmentStmt`, а не `BinaryExpr`
> - `#Если`/`#ИначеЕсли`/`#Иначе`/`#КонецЕсли` внутри процедуры/блока → ошибка (невалидный BSL)
> - `internal/analysis/linters/` — 9 правил статического анализа, интегрированных в `publishDiagnostics`. При добавлении нового правила: создай файл в пакете `linters`, добавь функцию `checkXxx` в `rules` в `lint.go`, напиши тесты.
> - **CRLF (`\r\n`)** — `readChar()` в лексере пропускает `\r` (line++ и col=0 только на `\n`). Все `.bsl` файлы корректно обрабатываются независимо от line endings (LF или CRLF).
> - **Параметры со значениями по умолчанию** — `parseParamList()` пропускает `= <expr>` после имени параметра. BSL допускает `Знач Парам = Значение` и многострочные списки параметров.
> - **MCP сервер (`internal/mcp/`)** — Streamable HTTP транспорт (POST `/` + SSE `/sse`), 7 tools (`bsl_parse`, `bsl_lint`, `bsl_format`, `bsl_symbols`, `bsl_define`, `bsl_hover`, `bsl_folding_ranges`), 2 prompts (`review_bsl_code`, `explain_bsl_module`), Resource store (hash-based кэш). При добавлении нового tool: зарегистрируй в `registerTools()` в `tools.go`, создай `handleXxx` метод в той же папке, добавь тест в `mcp_test.go`. Все инструменты переиспользуют существующие API из `parser`, `analysis`, `linters`.



LSP сервер для языка 1С (BSL). Чистый Go, zero external dependencies.

## Команды

```bash
go build -o bsl-lsp ./cmd/bsl-lsp           # LSP сервер
go build -o bsl-lsp-mcp ./cmd/bsl-lsp-mcp   # MCP сервер
go test -v ./...                            # все тесты (включая интеграционные на real .bsl)
go vet ./...                                # статический анализ
```

Либо через `Makefile`: `make build`, `make test`, `make vet`, `make run`, `make run-mcp`, `make clean`.

## Архитектура

- **`cmd/bsl-lsp/main.go`** — точка входа LSP сервера, вызывает `lsp.Run()`
- **`cmd/bsl-lsp-mcp/main.go`** — точка входа MCP сервера (`--host`, `--port`), запускает HTTP
- **`internal/lsp/`** — JSON-RPC 2.0 через stdin/stdout (`Content-Length` заголовки), session, handler
- **`internal/parser/`** — лексер (rune-based, UTF-8, полная кириллица) + рекурсивный спуск + AST
- **`internal/analysis/`**:
  - `symbol.go` — symbol table: `BuildSymbolTable(mod)` обходит AST, строит области видимости
  - `navigate.go` — `FindIdentAtPos(mod, line, col)` ищет идентификатор в AST по позиции (GoToDefinition, Hover)
  - `keywords.go` — `BSLKeywords` (34 keyword), `BSLGlobalMethods` (глобальные функции 1С)
  - `formatter.go` — `FormatDocument` — полное форматирование BSL
  - `semantic.go` — `CollectSemanticTokens`, `CollectFoldingRanges`, `FindCallAtPos`
- **`internal/analysis/linters/`**:
  - `lint.go` — `RunAll(mod, st)` — запускает все 9 правил статического анализа
  - `unused_var.go` — неиспользуемые переменные/параметры
  - `empty_block.go` — пустые блоки (процедура/функция/если/цикл/попытка)
  - `unreachable.go` — код после `Возврат`/`ВызватьИсключение`/`Прервать`/`Продолжить`
  - `magic_number.go` — магические числа (>3)
  - `too_many_params.go` — >7 параметров
  - `nested_depth.go` — глубина вложенности >5
  - `suspicious_assign.go` — самоприсваивание (`a = a`)
  - `missing_return.go` — функция без `Возврат` в некоторых ветках
  - `global_var_in_proc.go` — присваивание глобальной переменной внутри процедуры
- **`internal/mcp/`** — MCP сервер:
  - `server.go` — `Server` struct, dispatch MCP методов (`initialize`, `tools/list/call`, `resources/list/read`, `prompts/list/get`)
  - `transport.go` — Streamable HTTP: POST `/` → JSON-RPC, GET `/sse` → SSE stream
  - `tools.go` — 7 инструментов (обёртки над `parser`, `analysis`, `linters`)
  - `resources.go` — `ResourceStore` (hash-based кэш AST/диагностик/символов)
  - `prompts.go` — 2 промпта: `review_bsl_code`, `explain_bsl_module`
  - `types.go` — MCP-specific типы (`Tool`, `Resource`, `Prompt`, `ToolCallResult`, и т.д.)
- **`internal/workspace/document.go`** — thread-safe Document + Manager
- **`pkg/protocol/types.go`** — LSP типы

## MCP протокол

- Transport: Streamable HTTP (`POST /` — JSON-RPC 2.0, `GET /sse` — Server-Sent Events)
- `initialize` → `InitializeResult` с capabilities (tools + resources + prompts)
- `tools/list` → список 7 инструментов
- `tools/call` → вызов инструмента по имени с аргументами
- `resources/list` → список закешированных ресурсов
- `resources/read` → чтение ресурса по URI (`bsl://<hash>/<type>`)
- `prompts/list` → список 2 промптов
- `prompts/get` → получение промпта по имени (сервер подставляет результаты анализа)
- Неизвестные методы возвращают JSON-RPC error `-32601`
- Нотификации через SSE при изменении списка tools/resources

## LSP протокол

- Диагностики публикуются после каждого `didOpen`/`didChange` через `textDocument/publishDiagnostics`
- Ошибки парсера конвертируются в 0-based LSP позиции
- Неизвестные методы возвращают JSON-RPC error `-32601`
- `textDocument/didSave` — no-op
- Symbol table перестраивается при каждом didOpen/didChange (хранится в `Handler.documents[uri]`)
- `textDocument/definition` — `FindIdentAtPos` + `Lookup` → Location (с cross-module fallback через `ProjectAnalysis`)
- `textDocument/hover` — `FindIdentAtPos` + `Lookup` → markdown (kind + name + export tag), cross-module
- `textDocument/completion` — keywords + global methods + symbols + dot-context completion + exported cross-module symbols
- `textDocument/semanticTokens/full` — семантическая подсветка через `CollectSemanticTokens`
- `textDocument/codeLens` — `Экспорт`/`Локальная` над процедурами/функциями
- `textDocument/foldingRange` — сворачивание процедур/функций через `CollectFoldingRanges`
- `textDocument/formatting` — полное форматирование BSL через `FormatDocument`
- `textDocument/signatureHelp` — подсказки параметров через `FindCallAtPos`

## Парсер

Разбирает: процедуры/функции (параметры, Знач, Экспорт), If/ElseIf/Else, While, For/To, ForEach/In, Try/Except, Return/Raise, Goto, Break/Continue, присваивания, вызовы, бинарные/унарные/тернарные выражения, field access/index, New/Execute/Address/Type/Val, `#Если`/`#Область`, директивы компиляции, комментарии. Error recovery через `syncToStmt`.

### Особенности реализации
- `#Область`/`#КонецОбласти` внутри процедур — разрешено (молча пропускается)
- `#Если`/`#ИначеЕсли`/`#Иначе`/`#КонецЕсли` внутри процедур/блоков — error "preprocessor conditional directives are not allowed inside procedures and blocks"
- `_` как левая часть присваивания (`_ = 1`) — error "bare underscore is not a valid identifier"
- `Для Каждого <переменная> Из <коллекция> Цикл` — переменная цикла обязательна
- `5 = 5` (сравнение литералов) — парсится как `BinaryExpr`, НЕ как `AssignmentStmt`

## Тестирование

- Модульные тесты: `TestLexer_*`, `TestParser_*`
- Интеграционные: `TestRealBSLParseFiles` — парсит все `.bsl` файлы из `internal/parser/testdata/real_bsl/`
  - ~60 `module_*.bsl` (должны парситься без ошибок)
  - ~18 `bug_*.bsl` (должны выдавать ошибки парсинга)
- Symbol table: `TestBuildSymbolTable_*` в `internal/analysis/symbol_test.go`
- Тесты в `package parser` (white-box)
- Проверка: `go build ./... && go test ./... && go vet ./...`

## Соглашения

- Нет внешних Go-зависимостей (только stdlib)
- `skipComments()` пропускает комментарии и препроцессор при разборе statement'ов
- `consumeSemicolon` в `parseBlock`, `nextToken` в Break/Cycle
- AST узлы Procedure/Function/If/While/For/ForEach/Try хранят `Line`/`Col` позиции (исправлено, не через `Directives[0]`)
- MCP tools переиспользуют существующие API: `parser.NewParser`, `analysis.BuildSymbolTable`, `linters.RunAll`, `analysis.FormatDocument`, `analysis.FindIdentAtPos`, `analysis.CollectFoldingRanges`
- При добавлении нового MCP tool: зарегистрируй в `registerTools()` в `tools.go`, реализуй `handleXxx`, добавь тест в `mcp_test.go`
