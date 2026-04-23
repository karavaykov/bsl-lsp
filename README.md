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
| ✅ | **MCP сервер** — Streamable HTTP (POST + SSE), 7 tools, 2 prompts, Resource store |

## Установка

```bash
git clone https://github.com/karavaykov/bsl-lsp.git
cd bsl-lsp
make build
```

Бинарник появится в `./bsl-lsp`.

Или через Docker (без локального Go):

```bash
docker pull ghcr.io/karavaykov/bsl-lsp:latest
```

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

### MCP сервер (для AI-ассистентов)

MCP-сервер реализует протокол [Model Context Protocol](https://modelcontextprotocol.io) через Streamable HTTP транспорт. Позволяет AI-ассистентам (Claude Code, Cursor, VS Code Copilot и др.) анализировать BSL-код.

```bash
# Запуск MCP-сервера на localhost:9090
bsl-lsp-mcp --port 9090 --host localhost

# POST / — JSON-RPC 2.0 endpoint
# GET  /sse — SSE stream для нотификаций
```

**Инструменты (Tools):**

| Tool | Описание |
|------|----------|
| `bsl_parse` | Разобрать BSL-код, вернуть AST и ошибки парсинга |
| `bsl_lint` | Статический анализ (9 правил) |
| `bsl_format` | Автоформатирование BSL |
| `bsl_symbols` | Извлечь символы (процедуры, функции, переменные, параметры) |
| `bsl_define` | Найти определение идентификатора по позиции |
| `bsl_hover` | Получить информацию по идентификатору |
| `bsl_folding_ranges` | Получить folding ranges модуля |

**Промпты (Prompts):**

| Prompt | Описание |
|--------|----------|
| `review_bsl_code` | Шаблон для ревью BSL-кода (встраивает результат `bsl_lint`) |
| `explain_bsl_module` | Шаблон для объяснения структуры модуля (встраивает результат `bsl_symbols`) |

**Ресурсы (Resources):** AST, диагностики и символы кешируются по хешу кода и доступны по URI `bsl://<hash>/<type>`.

**Пример подключения в Claude Desktop (`claude_desktop_config.json`):**
```json
{
  "mcpServers": {
    "bsl": {
      "command": "/path/to/bsl-lsp-mcp",
      "args": ["--port", "9090"]
    }
  }
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

### Docker

```bash
# проверка синтаксиса
docker run --rm -v "$PWD:/work" ghcr.io/karavaykov/bsl-lsp:latest check /work/module.bsl

# форматирование in-place (требует права на запись)
docker run --rm -v "$PWD:/work" -u "$(id -u):$(id -g)" ghcr.io/karavaykov/bsl-lsp:latest format /work/module.bsl

# форматирование с выводом в stdout
docker run --rm -v "$PWD:/work" ghcr.io/karavaykov/bsl-lsp:latest format --stdout /work/module.bsl > clean.bsl
```

### MCP сервер в Docker

```bash
# через docker-compose (порт 9090)
docker compose up

# порт можно поменять в docker-compose.yml
```

Пример `docker-compose.yml` (изменение порта на 8080):
```yaml
services:
  bsl-lsp-mcp:
    image: ghcr.io/karavaykov/bsl-lsp:latest
    entrypoint: ["bsl-lsp-mcp", "--port", "9090", "--host", "0.0.0.0"]
    ports:
      - "9090:9090"    # меняй левый порт при необходимости
    restart: unless-stopped
```

```bash
# или напрямую:
docker run --rm -p 9090:9090 ghcr.io/karavaykov/bsl-lsp:latest bsl-lsp-mcp --port 9090 --host 0.0.0.0
```

Образ автоматически собирается и публикуется в GHCR через GitHub Actions при каждом пуше в `master`.

## Разработка

```bash
make build   # сборка (bsl-lsp + bsl-lsp-mcp)
make test    # все тесты
make vet     # статический анализ
make run     # запуск LSP сервера
make run-mcp # запуск MCP сервера
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

### Ключевые архитектурные решения

- **Zero external dependencies** — только stdlib Go
- **`#Область` внутри процедур** — разрешена (распространённая практика в BSL)
- **`#Если` внутри процедур** — ошибка (невалидный BSL)
- **`_` как LHS присваивания** — ошибка
- **`TokenEqual` на верхнем уровне** — break, чтобы `А = 10` парсилось как `AssignmentStmt`, а не `BinaryExpr`
- **Сравнение литералов** (`5 = 5`) — `BinaryExpr`, не путается с присваиванием
- **CRLF (\r\n)** — `readChar()` в лексере пропускает `\r`. line++ и col=0 только на `\n`. Все `.bsl` файлы работают независимо от line endings
- **Параметры по умолчанию** — `parseParamList()` пропускает `= <expr>` после имени параметра. Поддерживаются многострочные списки параметров

## Лицензия

MIT
