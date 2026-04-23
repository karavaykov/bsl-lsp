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
| 🔲 | Semantic Tokens, CodeLens, Folding, Formatting, SignatureHelp |
| 🔲 | Экспорт/импорт символов между модулями |

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

Подробнее — [AGENTS.md](AGENTS.md) и [ROADMAP.md](ROADMAP.md).

## Архитектура

```
cmd/bsl-lsp/          — точка входа
internal/lsp/         — JSON-RPC 2.0, session, handler
internal/parser/      — лексер + рекурсивный спуск + AST
internal/analysis/    — symbol table, навигация, keywords
internal/workspace/   — thread-safe Document + Manager
pkg/protocol/         — LSP типы
```

## Лицензия

MIT
