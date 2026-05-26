@echo off
rem 🛡️ MigraGuard L3 Orchestrator: Global Multi-Scenario Pipeline (CMD Standalone)
rem Usage: .\scripts\cmd\run_global_pipeline.cmd

set "MODULE_DIR=%~dp0"
set "PROJECT_ROOT=%MODULE_DIR%..\.."

echo ==================================================
echo 🛡️  MigraGuard L3 Global Multi-Scenario Orchestrator
echo ==================================================

rem 1. Count and check scenarios
set /a "TOTAL=0"
for %%s in ("%PROJECT_ROOT%\experiments\scenarios\*.yaml") do (
    if exist "%%s" set /a "TOTAL+=1"
)

if %TOTAL%==0 (
    echo ❌ [ERROR] No scenario files found in experiments/scenarios/
    exit /b 1
)

echo 📂 Found %TOTAL% scenario file(s) to process.
echo ==================================================

set "SCENARIO_SCRIPT=%MODULE_DIR%run_scenario_pipeline.cmd"
set /a "COUNT=0"

for %%s in ("%PROJECT_ROOT%\experiments\scenarios\*.yaml") do (
    if exist "%%s" (
        set /a "COUNT+=1"
        for %%i in ("%%s") do set "S_NAME=%%~nxi"
        
        echo.
        echo ==================================================
        echo 👉 [Scenario !COUNT!/%TOTAL%] Processing: !S_NAME!
        rem Use delayed expansion safe workaround call
        call :process_single_scenario "%%s"
    )
)
goto :global_pipeline_complete

:process_single_scenario
set "SCEN_PATH=%~1"
for %%i in ("%SCEN_PATH%") do set "S_NAME=%%~nxi"
echo 👉 [Scenario %COUNT%/%TOTAL%] Processing: %S_NAME%
echo ==================================================

call "%SCENARIO_SCRIPT%" "%SCEN_PATH%"
if errorlevel 1 (
    echo ⚠️  [WARNING] Pipeline execution failed for scenario: %S_NAME%. Continuing to next.
) else (
    echo ✅ [Scenario SUCCESS] Completed scenario: %S_NAME%
)
exit /b

:global_pipeline_complete
echo.
echo ==================================================
echo 🏆 [L3 SUCCESS] Global Multi-Scenario Pipeline Completed!
echo 📂 All archived outputs are located in experiments/reports/batch_runs/
echo ==================================================
exit /b 0
