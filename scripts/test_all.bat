@echo off
setlocal enabledelayedexpansion

:: Set terminal encoding to UTF-8
chcp 65001 > nul

:: Get the project root directory (one level up from scripts/)
set SCRIPT_DIR=%~dp0
pushd %SCRIPT_DIR%..

set MIGRATIONS_DIR=migrations

echo 🚀 Starting risk analysis for all DDL files...

for %%f in (%MIGRATIONS_DIR%\*.sql) do (
    set FILE_NAME=%%~nxf
    echo.
    echo [Analyzing file: !FILE_NAME!]
    
    :: Run docker command with specified service and internal container path
    docker compose run --rm analyze analyze /app/code/migrations/!FILE_NAME!
)

echo.
echo ✅ All analyses completed.
popd
pause
