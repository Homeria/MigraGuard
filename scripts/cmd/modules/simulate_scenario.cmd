@echo off
rem 🛡️ MigraGuard L1 Module: Simulate Scenario (CMD Standalone)
rem Usage: .\scripts\cmd\modules\simulate_scenario.cmd <scenario_yaml_path>

set "MODULE_DIR=%~dp0"
set "PROJECT_ROOT=%MODULE_DIR%..\..\.."

rem 1. Guard Clauses
if "%~1"=="" (
    echo ❌ [ERROR] Scenario YAML path is required.
    echo Usage: %0 ^<scenario_yaml_path^>
    exit /b 1
)

set "S_PATH=%~1"
rem Check if path is rooted
echo "%S_PATH%" | findstr /R "^[a-zA-Z]:\\" >nul
if errorlevel 1 (
    set "S_PATH=%PROJECT_ROOT%\%S_PATH%"
)

if not exist "%S_PATH%" (
    echo ❌ [ERROR] Scenario file not found: %S_PATH%
    exit /b 1
)

rem Get filename
for %%i in ("%S_PATH%") do set "S_NAME=%%~nxi"

rem 2. Run simulation
echo 🔄 [L1] Simulating Scenario: %S_NAME%
pushd "%PROJECT_ROOT%"

set "BINARY_PATH=build\migraguard.exe"
if not exist "%BINARY_PATH%" (
    set "BINARY_PATH=build\migraguard"
)

"%BINARY_PATH%" simulate --scenario "%S_PATH%" --force
set "EXIT_CODE=%errorlevel%"

popd

if %EXIT_CODE% neq 0 (
    echo ❌ [ERROR] Simulation execution failed.
    exit /b 1
)

echo ✅ [L1 SUCCESS] Simulation complete for %S_NAME%
exit /b 0
