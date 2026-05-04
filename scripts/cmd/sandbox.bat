@echo off
setlocal enabledelayedexpansion

:: MigraGuard One-Click Sandbox Script (Windows CMD)
:: Usage: scripts\sandbox.bat <scenario_name> <ddl_name> [--docker] [--force]

set SCENARIO_IN=%1
set SQL_IN=%2
set DOCKER=false
set FORCE=false

shift
shift
:parse_args
if "%1"=="" goto end_parse
if "%1"=="--docker" set DOCKER=true
if "%1"=="--force" set FORCE=true
shift
goto parse_args
:end_parse

if "%SCENARIO_IN%"=="" goto usage
if "%SQL_IN%"=="" goto usage

:: Resolve Paths
set SCENARIO=experiments\scenarios\%SCENARIO_IN%
if not exist "%SCENARIO%" set SCENARIO=%SCENARIO_IN%

set SQL=experiments\ddl\%SQL_IN%
if not exist "%SQL%" set SQL=%SQL_IN%

:: Extract Experiment Name from YAML
for /f "tokens=2 delims=: " %%a in ('findstr "experiment_name" %SCENARIO%') do (
    set DB_NAME=%%a
    set DB_NAME=!DB_NAME:"=!
    set DB_NAME=!DB_NAME:'=!
)
set DB_PATH=experiments\data\!DB_NAME!.db

echo --------------------------------------------------
echo 🚀 MigraGuard Sandbox Runner (CMD)
echo Scenario: %SCENARIO%
echo SQL:      %SQL%
echo Sandbox:  %DB_PATH%
echo --------------------------------------------------

if "%DOCKER%"=="true" (
    set FORCE_FLAG=
    if "%FORCE%"=="true" set FORCE_FLAG=--force
    docker compose run --rm analyze-shell sh -c "go run ./cmd/migraguard simulate --scenario %SCENARIO% !FORCE_FLAG! && go run ./cmd/migraguard analyze %SQL% --sandbox %DB_PATH%"
) else (
    set SIM_CMD=go run ./cmd/migraguard simulate --scenario %SCENARIO%
    if "%FORCE%"=="true" set SIM_CMD=!SIM_CMD! --force
    %SIM_CMD%
    if %ERRORLEVEL% equ 0 (
        go run ./cmd/migraguard analyze %SQL% --sandbox %DB_PATH%
    )
)
goto :eof

:usage
echo Usage: %0 ^<scenario_name^> ^<ddl_name^> [--docker] [--force]
exit /b 1
