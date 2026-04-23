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

### LSP-сервер (для редакторов)

Сервер работает через stdin/stdout. Подключите его как LSP-сервер в вашем редакторе:

```json
{
  "command": "/path/to/bsl-lsp",
  "args": []
}
```

### CLI (проверка и форматирование)

```bash
bsl-lsp check <file.bsl>...          # синтаксис + статический анализ
bsl-lsp format <file.bsl>...         # автоформатирование (in-place)
bsl-lsp format --stdout <file.bsl>   # вывод результата в stdout
```

Пример проверки:
```text
$ bsl-lsp check module.bsl
module.bsl:10:0: [info/empty-block] Пустое тело блока Иначе
module.bsl:7:1: [warning/missing-return] Функция "ПолучитьИзКэша" — не во всех ветках есть Возврат
```

Пример форматирования:
```bash
bsl-lsp format --stdout messy.bsl > clean.bsl
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

## OpenCode Skill

Проект включает [SKILL.md](SKILL.md) — инструкцию для ИИ-агента [OpenCode](https://opencode.ai), добавляющую два инструмента:

| Инструмент | Описание |
|------------|----------|
| `check_bsl` | Проверка синтаксиса + статический анализ BSL-файла (9 правил) |
| `format_bsl` | Автоформатирование BSL-файла |

Skill устанавливается копированием в `~/.cursor/skills/bsl-lsp/`:
```bash
cp -r skills/bsl-lsp-skill ~/.cursor/skills/bsl-lsp
```

## Лицензия

MIT
