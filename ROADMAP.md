# Дорожная карта bsl-lsp

## ✅ Готово
- **Парсер** — лексер (UTF-8, кириллица), AST (все конструкции), рекурсивный спуск (If/While/For/Try/выражения и т.д.), error recovery, препроцессор/области/директивы, 7 интеграционных тестов на реальных `.bsl` файлах
- **Фундамент LSP** — JSON-RPC 2.0 через stdin/stdout, handshake, textDocument sync (didOpen/Change/Close/Save), publishDiagnostics, Document+Manager (thread-safe), 0-based LSP позиционирование
- **Symbol table** — глобальные/локальные области видимости, разрешение имён (переменные, процедуры, функции, параметры), поддержка `Перем`, автообъявление через присваивание
- **DocumentSymbol** — структура модуля в LSP outline (процедуры/функции + локальные символы)

## ✅ Фаза A — остальные LSP возможности
- [x] **GoToDefinition** — поиск идентификатора в AST + lookup в symbol table → Location
- [x] **Hover** — возвращает markdown с типом (Процедура/Функция/Переменная/Параметр)
- [x] **Completion** — BSL keywords + глобальные методы 1С + локальные/глобальные символы + методы после `.`
- [ ] **Экспорт/импорт символов** между модулями (требует cross-module analysis, Фаза A+)

## 🔲 Фаза B — продвинутые LSP
- [ ] Semantic Tokens, CodeLens, Folding Ranges, Formatting, SignatureHelp

## 🔲 Фаза C — промышленное качество
- [ ] Бенчмарки (10k+ строк), тестовая база (50+ реальных файлов), zero-alloc paths, структурированные логи, конфигурация LSP через initialize, CI/CD
