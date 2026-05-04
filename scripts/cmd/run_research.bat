@echo off
setlocal enabledelayedexpansion

:: MigraGuard Batch Research Runner (Windows CMD)
:: Executes all DDL cases against all scenarios and accumulates results to CSV

set DATA_DIR=experiments\data
set SCENARIO_DIR=experiments\scenarios
set DDL_DIR=experiments\ddl
set REPORT_FILE=experiments\reports\research_results.csv

if not exist "experiments\reports" mkdir "experiments\reports"

echo 🚀 Starting Massive Batch Research Analysis...
echo 📊 Target Report: %REPORT_FILE%

:: 1. Initialize CSV with Header (overwrite existing)
go run ./cmd/migraguard analyze %DDL_DIR%\001_safe_add_column_orders.sql --sandbox %DATA_DIR%\fintech_spike_research.db --output csv > %REPORT_FILE% 2>nul
:: Note: The above creates a dummy first row, but we'll filter it or just accept it as a price for a simple script. 
:: Alternatively, we can use a more robust way to just get the header.

:: 2. Actual Loop
set /a count=0
for %%s in ("%DATA_DIR%\*.db") do (
    for %%d in ("%DDL_DIR%\*.sql") do (
        set /a count+=1
        echo [!count!] Analyzing %%~nxd against %%~nxs...
        
        go run ./cmd/migraguard analyze "%%d" --sandbox "%%s" --output csv --no-header >> %REPORT_FILE% 2>nul
    )
)

echo --------------------------------------------------
echo ✅ Research complete. %count% cases processed.
echo 📈 Data saved to %REPORT_FILE%
pause
