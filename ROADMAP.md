# Дорожная карта bsl-lsp

## ✅ Готово
- **Парсер** — лексер (UTF-8, кириллица), AST (все конструкции: If/While/For/Try/ForEach/Return/Raise/Goto/Break/Continue/выражения/New/Execute/Address/Type/Val/тернарный), рекурсивный спуск, error recovery (`syncToStmt`), препроцессор/области/директивы
- **Фундамент LSP** — JSON-RPC 2.0 через stdin/stdout, handshake, textDocument sync (didOpen/Change/Close/Save), publishDiagnostics, Document+Manager (thread-safe), 0-based LSP позиционирование
- **Symbol table** — глобальные/локальные области видимости, разрешение имён, поддержка `Перем`, автообъявление через присваивание
- **DocumentSymbol** — структура модуля в LSP outline

## ✅ Фаза A — LSP возможности
- [x] GoToDefinition — поиск идентификатора в AST + lookup в symbol table → Location
- [x] Hover — markdown с типом (Процедура/Функция/Переменная/Параметр)
- [x] Completion — BSL keywords + глобальные методы 1С + локальные/глобальные символы + методы после `.`
- [x] Экспорт/импорт символов между модулями — `ProjectAnalysis`, cross-module lookup

## ✅ Фаза B — продвинутые LSP
- [x] Semantic Tokens — `textDocument/semanticTokens/full`
- [x] CodeLens — `Экспорт`/`Локальная` над процедурами/функциями
- [x] Folding Ranges — сворачивание процедур/функций
- [x] Formatting — автоформатирование BSL
- [x] SignatureHelp — подсказки параметров

## ✅ Фаза C — тестовая база (выполнено с учётом эвристик)
- [x] ~60 `module_*.bsl` интеграционных тестов (0 ошибок парсинга)
- [x] ~18 `bug_*.bsl` тестов (ожидаемые ошибки парсинга)
- [x] `#Область` внутри процедур — разрешена (молча пропускается)
- [x] `#Если`/`#ИначеЕсли`/`#Иначе`/`#КонецЕсли` внутри процедур/блоков — ошибка
- [x] `_` как LHS присваивания — ошибка
- [x] `Для Каждого` без переменной — ошибка (синтаксическая неоднозначность)
- [x] `5 = 5` (сравнение литералов) — корректный `BinaryExpr`, НЕ путается с присваиванием
- [x] `TokenEqual` break на `precLowest` — хак для различия присваивания и сравнения

## ✅ Фаза D — статический анализ (BSL-HC)
- [x] **Пакет `internal/analysis/linters/`** — 9 правил: `unused-variable`, `empty-block`, `unreachable-code`, `magic-number`, `too-many-params`, `nested-depth`, `suspicious-assignment`, `missing-return`, `global-var-in-proc`
- [x] Интеграция в `publishDiagnostics` — диагностики линтеров публикуются вместе с ошибками парсера
- [x] Тесты для каждого правила (позитивные + негативные)
- [x] Сохранение zero external dependencies

## ✅ Фаза E — CLI режим
- [x] Subcommands `check` и `format` в `cmd/bsl-lsp/main.go`
- [x] `bsl-lsp check <file>` — синтаксис + статический анализ (9 правил), human-readable вывод
- [x] `bsl-lsp format <file>` — автоформатирование in-place
- [x] `bsl-lsp format --stdout <file>` — вывод результата в stdout
- [x] `bsl-lsp lsp` / `bsl-lsp` (без аргументов) — LSP-сервер (обратная совместимость)
- [x] `bsl-lsp --help` — справка
- [x] SKILL.md обновлён на прямой вызов CLI (без LSP-обёрток)
- [x] Сохранение zero external dependencies

## Critical Context (реализовано)

> CLI-режим добавлен: `bsl-lsp check <file>` и `bsl-lsp format <file>`.  
> SKILL.md использует прямой вызов CLI вместо LSP-обёрток.  
> Выбран вариант **A** — subcommands в `cmd/bsl-lsp/` с ручным dispatch через `os.Args`.

## 🔲 Фаза F — промышленное качество
- [ ] Бенчмарки (10k+ строк), zero-alloc paths
- [ ] Структурированные логи (slog, уровни, structured)
- [ ] Конфигурация LSP через `initialize` (опции клиента)
- [ ] CI/CD (GitHub Actions: lint, test, build, release)
- [ ] Покрытие error recovery: больше `bug_*.bsl` кейсов
- [ ] Тесты на анализ (symbol table + navigation) на реальных модулях
- [ ] Fuzz-тесты лексера/парсера
