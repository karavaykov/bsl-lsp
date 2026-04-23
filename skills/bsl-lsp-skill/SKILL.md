---
name: bsl-lsp
description: "1C:Enterprise (BSL) syntax checking, static analysis, and code formatting using bsl-lsp — a Go-based LSP server. Use when the user asks to check 1C code, find errors, lint BSL files, or format .bsl/.os code."
---

# BSL-LSP — 1C Syntax Checker & Linter

## Overview

`bsl-lsp` provides:
- **Parser diagnostics** — syntax errors
- **Static analysis (9 rules)** — unused variables, empty blocks, unreachable code, magic numbers, too many parameters, nested depth, self-assignment, missing return in functions, global variable assignment inside procedures
- **Formatting** — auto-formatting of BSL code

Binary location on this machine: `C:\Users\karavaikov.s\opencodeproj\bsl-lsp\bsl-lsp.exe` (Windows) or `/mnt/c/Users/karavaikov.s/opencodeproj/bsl-lsp/bsl-lsp` (WSL/Linux).

Also available as Docker image on GitHub Container Registry:
```
ghcr.io/karavaykov/bsl-lsp:latest
```

## CLI Commands

The bsl-lsp binary has built-in CLI mode (no LSP handshake needed).

### Check syntax & lint a BSL file

```bash
bsl-lsp check <file.bsl>
```

Output format:
```
file.bsl:line:col: parse error: <message>
file.bsl:line:col: [warning/unused-variable] <message>
file.bsl:line:col: [info/empty-block] <message>
```

Exit code: 1 if parse errors or warnings found, 0 otherwise.

### Format a BSL file

In-place:
```bash
bsl-lsp format <file.bsl>
```

To stdout (preview):
```bash
bsl-lsp format --stdout <file.bsl> > formatted.bsl
```

### Using Docker (portable, no local Go needed)

```bash
docker run --rm -v "$PWD:/work" ghcr.io/karavaykov/bsl-lsp:latest check /work/module.bsl
docker run --rm -v "$PWD:/work" ghcr.io/karavaykov/bsl-lsp:latest format /work/module.bsl
docker run --rm -v "$PWD:/work" ghcr.io/karavaykov/bsl-lsp:latest format --stdout /work/module.bsl
```

### Start LSP server (for editors)

```bash
bsl-lsp lsp
```

Legacy: `bsl-lsp` with no args also starts LSP server (backward compatible).

## Linter rules (9 checks)

| Code | Check | Severity |
|---|---|---|
| `unused-variable` | Variables/parameters declared but never used | Warning |
| `empty-block` | Empty procedure/function/if/loop/try bodies | Warning |
| `unreachable-code` | Code after Return/Raise/Break/Continue | Warning |
| `magic-number` | Numeric literals > 3 | Info |
| `too-many-params` | More than 7 parameters | Warning |
| `nested-depth` | Nesting depth > 5 levels | Warning |
| `suspicious-assignment` | Self-assignment (`a = a`) | Warning |
| `missing-return` | Function without Return in all branches | Warning |
| `global-var-in-proc` | Global variable assignment inside procedure | Info |

## Build

```bash
cd /mnt/c/Users/karavaikov.s/opencodeproj/bsl-lsp
export PATH="$HOME/go/bin:$PATH"
go build -o bsl-lsp ./cmd/bsl-lsp
```

## Test

```bash
cd /mnt/c/Users/karavaikov.s/opencodeproj/bsl-lsp
export PATH="$HOME/go/bin:$PATH"
go test ./...
```
