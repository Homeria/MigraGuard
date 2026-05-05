@echo off
setlocal enabledelayedexpansion
set REPORT=experiments\reports\research_results.csv
if not exist experiments\reports mkdir experiments\reports
echo 🚀 Starting Massive Batch Research Analysis...
set /a count=0
for %%b in (experiments\data\*.db) do (
    for %%d in (experiments\ddl\*.sql) do (
        set /a count+=1
        set HEADER_FLAG=
        if !count! GTR 1 set HEADER_FLAG=--no-header
        echo [!count!] %%~nxd @ %%~nxb
        if !count! EQU 1 (
            go run ./cmd/migraguard analyze %%d --sandbox %%b --output csv !HEADER_FLAG! > %REPORT%
        ) else (
            go run ./cmd/migraguard analyze %%d --sandbox %%b --output csv !HEADER_FLAG! >> %REPORT%
        )
    )
)
echo ✅ Research complete. Master report: %REPORT%
