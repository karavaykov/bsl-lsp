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
- **`internal/workspace/document.go`** — thread-safe Document + Manager
- **`pkg/protocol/types.go`** — LSP типы

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
