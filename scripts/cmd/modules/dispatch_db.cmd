@echo off
rem 🛡️ MigraGuard L1 Module: Dispatch DB to Case (CMD Standalone)
rem Usage: .\scripts\cmd\modules\dispatch_db.cmd <seed_db_path> <case_yaml_path> [override_db_path]

set "MODULE_DIR=%~dp0"
set "PROJECT_ROOT=%MODULE_DIR%..\..\.."

rem 1. Guard Clauses
if "%~1"=="" if "%~2"="" (
    echo ❌ [ERROR] Both seed_db_path and case_yaml_path are required.
    echo Usage: %0 ^<seed_db_path^> ^<case_yaml_path^> [override_db_path]
    exit /b 1
)

set "SEED_DB=%~1"
set "CASE_YAML=%~2"
set "OVERRIDE_DB=%~3"

rem Check if paths are rooted
echo "%SEED_DB%" | findstr /R "^[a-zA-Z]:\\" >nul
if errorlevel 1 (set "SEED_DB=%PROJECT_ROOT%\%SEED_DB%")

echo "%CASE_YAML%" | findstr /R "^[a-zA-Z]:\\" >nul
if errorlevel 1 (set "CASE_YAML=%PROJECT_ROOT%\%CASE_YAML%")

if not exist "%SEED_DB%" (
    echo ❌ [ERROR] Seed DB not found: %SEED_DB%
    exit /b 1
)
if not exist "%CASE_YAML%" (
    echo ❌ [ERROR] Case YAML configuration not found: %CASE_YAML%
    exit /b 1
)

rem 2. Extract db_path and case_name dynamically from case configuration metadata block
set "CASE_NAME="
for /f "tokens=2 delims=: " %%a in ('findstr "case_name:" "%CASE_YAML%"') do (
    set "CASE_NAME=%%~a"
)
rem Trim quotes
if defined CASE_NAME (
    set CASE_NAME=%CASE_NAME:'=%
    set CASE_NAME=%CASE_NAME:"=%
)

set "DB_PATH="
if not "%OVERRIDE_DB%"=="" (
    set "DB_PATH=%OVERRIDE_DB%"
) else (
    for /f "tokens=2 delims=: " %%a in ('findstr "db_path:" "%CASE_YAML%"') do (
        set "DB_PATH=%%~a"
    )
)
rem Trim quotes
if defined DB_PATH (
    set DB_PATH=%DB_PATH:'=%
    set DB_PATH=%DB_PATH:"=%
)

if "%DB_PATH%"=="" (
    echo ❌ [ERROR] Could not parse db_path from configuration metadata: %CASE_YAML%
    exit /b 1
)

rem Check if DB_PATH is rooted
echo "%DB_PATH%" | findstr /R "^[a-zA-Z]:\\" >nul
if errorlevel 1 (
    set "DB_PATH=%PROJECT_ROOT%\%DB_PATH%"
)

rem Get destination directory
for %%i in ("%DB_PATH%") do set "DEST_DIR=%%~dpi"

rem 3. Create destination directory and copy database
if not exist "%DEST_DIR%" mkdir "%DEST_DIR%" 2>nul
copy /y "%SEED_DB%" "%DB_PATH%" >nul
if errorlevel 1 (
    echo ❌ [ERROR] Failed to dispatch database to %DB_PATH%
    exit /b 1
)

for %%i in ("%DB_PATH%") do set "DB_NAME=%%~nxi"
echo    -^> [L1 SUCCESS] Dispatched to Case [%CASE_NAME%]: %DB_NAME%
exit /b 0
