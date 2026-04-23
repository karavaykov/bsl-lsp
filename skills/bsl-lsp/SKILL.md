---
name: bsl-lsp
description: "1C:Enterprise (BSL) syntax checking, static analysis, and code formatting using bsl-lsp — a Go-based LSP server. Use when the user asks to check 1C code, find errors, lint BSL files, or format .bsl/.os code."
---

# BSL-LSP — 1C Syntax Checker & Linter

## Overview

`bsl-lsp` is an LSP server for 1C:Enterprise language (BSL). It provides:
- **Parser diagnostics** — syntax errors
- **Static analysis (9 rules)** — unused variables, empty blocks, unreachable code, magic numbers, too many parameters, nested depth, self-assignment, missing return in functions, global variable assignment inside procedures
- **Formatting** — auto-formatting of BSL code

Binary location: `C:\Users\karavaikov.s\opencodeproj\bsl-lsp\bsl-lsp.exe`

## Commands

### Check syntax & lint a BSL file

Run the bsl-lsp LSP server in one-shot mode. The server reads JSON-RPC from stdin. Below is a PowerShell script that sends a `didOpen` with the file content and captures diagnostics:

```powershell
$bslLsp = "C:\Users\karavaikov.s\opencodeproj\bsl-lsp\bsl-lsp.exe"
$file = "C:\path\to\file.bsl"
$content = Get-Content $file -Raw -Encoding UTF8

# Build JSON-RPC: initialize -> didOpen -> shutdown
$body = @"
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":null,"capabilities":{}}}
{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///$($file.Replace('\','/'))","languageId":"bsl","version":1,"text":"$($content.Replace('"','\"').Replace("`n","\n").Replace("`r",""))"}}}
{"jsonrpc":"2.0","id":2,"method":"shutdown"}
{"jsonrpc":"2.0","method":"exit"}
"@

# Encode as UTF-8 without BOM and send
$utf8 = [System.Text.Encoding]::UTF8.GetBytes($body)
[System.IO.MemoryStream] $stdin = New-Object System.IO.MemoryStream
$stdin.Write($utf8, 0, $utf8.Length)
$stdin.Seek(0, [System.IO.SeekOrigin]::Begin) | Out-Null

$psi = New-Object System.Diagnostics.ProcessStartInfo
$psi.FileName = $bslLsp
$psi.RedirectStandardInput = $true
$psi.RedirectStandardOutput = $true
$psi.UseShellExecute = $false
$psi.CreateNoWindow = $true

$p = [System.Diagnostics.Process]::Start($psi)
$stdin.CopyTo($p.StandardInput.BaseStream)
$p.StandardInput.Close()
$output = $p.StandardOutput.ReadToEnd()
$p.WaitForExit(5000) | Out-Null

# Parse Content-Length headers and extract JSON
$output -split '(?=Content-Length:)' | ForEach-Object {
    if ($_ -match 'Content-Length:\s*\d+\s*\n\n({.*})') {
        $json = $matches[1]
        Write-Output $json
    }
}
```

### Format a BSL file

Use the bsl-lsp `textDocument/formatting` method:

```powershell
$bslLsp = "C:\Users\karavaikov.s\opencodeproj\bsl-lsp\bsl-lsp.exe"
$file = "C:\path\to\file.bsl"
$content = Get-Content $file -Raw -Encoding UTF8

$body = @"
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":null,"capabilities":{}}}
{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///$($file.Replace('\','/'))","languageId":"bsl","version":1,"text":"$($content.Replace('"','\"').Replace("`n","\n").Replace("`r",""))"}}}
{"jsonrpc":"2.0","id":2,"method":"textDocument/formatting","params":{"textDocument":{"uri":"file:///$($file.Replace('\','/'))"},"options":{"tabSize":4,"insertSpaces":true}}}
{"jsonrpc":"2.0","method":"textDocument/didClose","params":{"textDocument":{"uri":"file:///$($file.Replace('\','/'))"}}}
{"jsonrpc":"2.0","id":3,"method":"shutdown"}
{"jsonrpc":"2.0","method":"exit"}
"@

$utf8 = [System.Text.Encoding]::UTF8.GetBytes($body)
[System.IO.MemoryStream] $stdin = New-Object System.IO.MemoryStream
$stdin.Write($utf8, 0, $utf8.Length)
$stdin.Seek(0, [System.IO.SeekOrigin]::Begin) | Out-Null

$psi = New-Object System.Diagnostics.ProcessStartInfo
$psi.FileName = $bslLsp
$psi.RedirectStandardInput = $true
$psi.RedirectStandardOutput = $true
$psi.UseShellExecute = $false
$psi.CreateNoWindow = $true

$p = [System.Diagnostics.Process]::Start($psi)
$stdin.CopyTo($p.StandardInput.BaseStream)
$p.StandardInput.Close()
$output = $p.StandardOutput.ReadToEnd()
$p.WaitForExit(5000) | Out-Null

Write-Output $output
```

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

```powershell
cd C:\Users\karavaikov.s\opencodeproj\bsl-lsp
set PATH=%USERPROFILE%\go\bin;%PATH%
go build -o bsl-lsp.exe ./cmd/bsl-lsp
```

## Test

```powershell
cd C:\Users\karavaikov.s\opencodeproj\bsl-lsp
set PATH=%USERPROFILE%\go\bin;%PATH%
go test ./...
```
