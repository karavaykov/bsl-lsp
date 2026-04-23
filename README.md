# bsl-lsp

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

LSP-сервер для языка **1С:Предприятие (BSL)**. Чистый Go, zero external dependencies.

Обеспечивает интеллектуальное редактирование BSL-кода в любом редакторе, поддерживающем LSP (VS Code, Neovim, Emacs, JetBrains и др.).

## Возможности

| Статус | Возможность |
|--------|-------------|
| ✅ | Парсер — лексер (UTF-8, кириллица), AST (все конструкции), рекурсивный спуск, error recovery |
| ✅ | JSON-RPC 2.0 (stdin/stdout), `textDocument` sync |
| ✅ | Symbol table — области видимости, разрешение имён, `Перем`, автообъявление |
| ✅ | DocumentSymbol — outline модуля (процедуры/функции + локальные символы) |
| ✅ | GoToDefinition — переход к определению |
| ✅ | Hover — подсказка по символу (тип, описание) |
| ✅ | Completion — ключевые слова BSL, глобальные методы, локальные переменные, dot-контекст |
| ✅ | Semantic Tokens — семантическая подсветка кода |
| ✅ | CodeLens — подписи `Экспорт`/`Локальная` над процедурами/функциями |
| ✅ | Folding Ranges — сворачивание блоков |
| ✅ | Formatting — автоформатирование BSL |
| ✅ | SignatureHelp — подсказки параметров |
| ✅ | Экспорт/импорт символов между модулями |
| ✅ | **Статический анализ (BSL-HC)** — 9 правил: неиспользуемые переменные, пустые блоки, недостижимый код, магические числа (>3), >7 параметров, глубина вложенности >5, самоприсваивание, пропущенный `Возврат` в функциях, присваивание глобальных переменных внутри процедур |

## Установка

```bash
git clone https://github.com/karavaykov/bsl-lsp.git
cd bsl-lsp
make build
```

Бинарник появится в `./bsl-lsp`.

## Использование

Сервер работает через stdin/stdout. Подключите его как LSP-сервер в вашем редакторе:

```json
{
  "command": "/path/to/bsl-lsp",
  "args": []
}
```

### VS Code

Установите расширение BSL и укажите путь к серверу в `settings.json`:

```json
{
  "bsl.lsp.server": "/path/to/bsl-lsp"
}
```

### Neovim (lspconfig)

```lua
require('lspconfig').bsl_lsp.setup {
  cmd = { '/path/to/bsl-lsp' }
}
```

## Разработка

```bash
make build   # сборка
make test    # все тесты
make vet     # статический анализ
make run     # запуск сервера
make clean   # очистка
```

### Валидация изменений

После любых изменений в парсере обязательно:

```bash
go build ./... && go test ./... && go vet ./...
```

Все интеграционные тесты (`TestRealBSLParseFiles`) должны проходить:
- `module_*.bsl` — 0 ошибок
- `bug_*.bsl` — ожидаемые ошибки

Подробнее — [AGENTS.md](AGENTS.md) и [ROADMAP.md](ROADMAP.md).

## Архитектура

```
cmd/bsl-lsp/          — точка входа
internal/lsp/         — JSON-RPC 2.0, session, handler
internal/parser/      — лексер + рекурсивный спуск + AST
internal/analysis/    — symbol table, навигация, keywords, formatter
internal/analysis/linters/ — статический анализ BSL (9 правил)
internal/workspace/   — thread-safe Document + Manager
pkg/protocol/         — LSP типы
```

### Ключевые архитектурные решения

- **Zero external dependencies** — только stdlib Go
- **`#Область` внутри процедур** — разрешена (распространённая практика в BSL)
- **`#Если` внутри процедур** — ошибка (невалидный BSL)
- **`_` как LHS присваивания** — ошибка
- **`TokenEqual` на верхнем уровне** — break, чтобы `А = 10` парсилось как `AssignmentStmt`, а не `BinaryExpr`
- **Сравнение литералов** (`5 = 5`) — `BinaryExpr`, не путается с присваиванием

## Лицензия

MIT
