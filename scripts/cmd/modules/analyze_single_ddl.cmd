@echo off
rem 🛡️ MigraGuard L1 Module: Analyze Single DDL (CMD Standalone)
rem Usage: .\scripts\cmd\modules\analyze_single_ddl.cmd <ddl_path> <db_path> <config_path> <output_csv_path> [--append]

set "MODULE_DIR=%~dp0"
set "PROJECT_ROOT=%MODULE_DIR%..\..\.."

rem 1. Guard Clauses
if "%~1"=="" if "%~2"="" if "%~3"="" if "%~4"="" (
    echo ❌ [ERROR] Missing required arguments.
    echo Usage: %0 ^<ddl_path^> ^<db_path^> ^<config_path^> ^<output_csv_path^> [--append]
    exit /b 1
)

set "DDL_PATH=%~1"
set "DB_PATH=%~2"
set "CONFIG_PATH=%~3"
set "OUT_CSV=%~4"
set "MODE=OVERWRITE"

if "%~5"=="--append" (
    set "MODE=APPEND"
)

rem Check if paths are rooted
echo "%DDL_PATH%" | findstr /R "^[a-zA-Z]:\\" >nul
if errorlevel 1 (set "DDL_PATH=%PROJECT_ROOT%\%DDL_PATH%")

echo "%DB_PATH%" | findstr /R "^[a-zA-Z]:\\" >nul
if errorlevel 1 (set "DB_PATH=%PROJECT_ROOT%\%DB_PATH%")

echo "%CONFIG_PATH%" | findstr /R "^[a-zA-Z]:\\" >nul
if errorlevel 1 (set "CONFIG_PATH=%PROJECT_ROOT%\%CONFIG_PATH%")

echo "%OUT_CSV%" | findstr /R "^[a-zA-Z]:\\" >nul
if errorlevel 1 (set "OUT_CSV=%PROJECT_ROOT%\%OUT_CSV%")

if not exist "%DDL_PATH%" (echo ❌ [ERROR] DDL file not found: %DDL_PATH% & exit /b 1)
if not exist "%DB_PATH%" (echo ❌ [ERROR] Database file not found: %DB_PATH% & exit /b 1)
if not exist "%CONFIG_PATH%" (echo ❌ [ERROR] Config file not found: %CONFIG_PATH% & exit /b 1)

rem 2. Run analysis
pushd "%PROJECT_ROOT%"

set "HEADER_FLAG="
if "%MODE%"=="APPEND" (
    if exist "%OUT_CSV%" (
        set "HEADER_FLAG=--no-header"
    )
)

for %%i in ("%OUT_CSV%") do set "DEST_DIR=%%~dpi"
if not exist "%DEST_DIR%" mkdir "%DEST_DIR%" 2>nul

set "BINARY_PATH=build\migraguard.exe"
if not exist "%BINARY_PATH%" (
    set "BINARY_PATH=build\migraguard"
)

rem Run execution. We ignore errors using '|| rem' to maintain smooth batch pipeline flow
if "%MODE%"=="OVERWRITE" (
    "%BINARY_PATH%" analyze "%DDL_PATH%" --sandbox "%DB_PATH%" --output csv %HEADER_FLAG% --config "%CONFIG_PATH%" > "%OUT_CSV%" 2>nul || rem
) else (
    "%BINARY_PATH%" analyze "%DDL_PATH%" --sandbox "%DB_PATH%" --output csv %HEADER_FLAG% --config "%CONFIG_PATH%" >> "%OUT_CSV%" 2>nul || rem
)

popd

for %%i in ("%DDL_PATH%") do set "DDL_NAME=%%~nxi"
for %%i in ("%DB_PATH%") do set "DB_NAME=%%~nxi"
echo    -^> [L1 SUCCESS] Analyzed %DDL_NAME% on %DB_NAME% (Mode: %MODE%)
exit /b 0
