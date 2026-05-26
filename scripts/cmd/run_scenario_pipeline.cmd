@echo off
rem 🛡️ MigraGuard L2 Orchestrator: Single Scenario Pipeline (CMD Standalone)
rem Usage: .\scripts\cmd\run_scenario_pipeline.cmd <scenario_yaml_path>

set "MODULE_DIR=%~dp0"
set "PROJECT_ROOT=%MODULE_DIR%..\.."

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

for %%i in ("%S_PATH%") do set "S_NAME=%%~nxi"

echo ==================================================
echo 🛡️  MigraGuard L2 Scenario Pipeline: %S_NAME%
echo ==================================================

rem 0. Generate Unique timestamp (Safe formatting for spaces and colons)
set "t=%time: =0%"
set "TIMESTAMP=%date:~0,4%-%date:~5,2%-%date:~8,2%_%t:~0,2%-%t:~3,2%-%t:~6,2%"
set "RUN_DIR=%PROJECT_ROOT%\experiments\reports\batch_runs\%TIMESTAMP%"
set "RUN_DATA_DIR=%RUN_DIR%\data"

if not exist "%RUN_DATA_DIR%" mkdir "%RUN_DATA_DIR%" 2>nul

rem 0. Clean old py scripts
echo 🧹 [0/5] Cleaning temporary Python files inside reports...
del /s /q "%PROJECT_ROOT%\experiments\reports\*.py" >nul 2>&1

rem 1. Simulate seed
set "SIMULATE_SCRIPT=%MODULE_DIR%modules\simulate_scenario.cmd"
call "%SIMULATE_SCRIPT%" "%S_PATH%"
if errorlevel 1 (
    echo ❌ [ERROR] Simulation stage failed.
    exit /b 1
)

rem Extract seed DB name from scenario config
set "EXP_NAME="
for /f "tokens=2 delims=: " %%a in ('findstr "experiment_name:" "%S_PATH%"') do (
    set "EXP_NAME=%%~a"
)
if defined EXP_NAME (
    set EXP_NAME=%EXP_NAME:'=%
    set EXP_NAME=%EXP_NAME:"=%
)

set "SEED_DB=%PROJECT_ROOT%\%EXP_NAME%.db"
if not exist "%SEED_DB%" (
    set "SEED_DB=%PROJECT_ROOT%\experiments\data\%EXP_NAME%.db"
)

if not exist "%SEED_DB%" (
    echo ❌ [ERROR] Seed database not found: %SEED_DB%
    exit /b 1
)

rem 2. Dispatch DB to each case nested inside the runs pack
echo 📂 [2/5] Dispatching databases to capacity cases nested in runs pack...
set "DISPATCH_SCRIPT=%MODULE_DIR%modules\dispatch_db.cmd"

for %%c in ("%PROJECT_ROOT%\experiments\configs\cases\*.yaml") do (
    if exist "%%c" (
        set "CASE_NAME="
        for /f "tokens=2 delims=: " %%a in ('findstr "case_name:" "%%c"') do (
            set "CASE_NAME=%%~a"
        )
        rem Remove quotes
        if defined CASE_NAME (
            set CASE_NAME=!CASE_NAME:'=!
            set CASE_NAME=!CASE_NAME:"=!
        )
        
        rem Use delayed expansion fallback safely
        call :get_case_name "%%c"
    )
)
goto :after_dispatch

:get_case_name
set "C_PATH=%~1"
set "C_NAME="
for /f "tokens=2 delims=: " %%a in ('findstr "case_name:" "%C_PATH%"') do (set "C_NAME=%%~a")
if defined C_NAME (
    set C_NAME=%C_NAME:'=%
    set C_NAME=%C_NAME:"=%
)
set "DEST_DB=%RUN_DATA_DIR%\%C_NAME%.db"
call "%DISPATCH_SCRIPT%" "%SEED_DB%" "%C_PATH%" "%DEST_DB%"
exit /b

:after_dispatch

rem Cleanup root seed db copy if it exists to keep workspace tidy
if exist "%PROJECT_ROOT%\%EXP_NAME%.db" del /f /q "%PROJECT_ROOT%\%EXP_NAME%.db" >nul 2>&1

rem 3. Batch Analyze (All Case configs x All DDL sqls) targeting runs DBs and local runs CSV inside data/
echo 🚀 [3/5] Executing dynamic batch analyses...
set "REPORT_CSV=%RUN_DATA_DIR%\research_results.csv"
if exist "%REPORT_CSV%" del /f /q "%REPORT_CSV%" >nul 2>&1

set "ANALYZE_SCRIPT=%MODULE_DIR%modules\analyze_single_ddl.cmd"
set /a "COUNT=0"

for %%c in ("%PROJECT_ROOT%\experiments\configs\cases\*.yaml") do (
    if exist "%%c" (
        for %%s in ("%PROJECT_ROOT%\experiments\ddl\*.sql") do (
            if exist "%%s" (
                set /a "COUNT+=1"
                call :run_single_analysis "%%c" "%%s"
            )
        )
    )
)
goto :after_analysis

:run_single_analysis
set "C_PATH=%~1"
set "S_PATH_DDL=%~2"
set "C_NAME="
for /f "tokens=2 delims=: " %%a in ('findstr "case_name:" "%C_PATH%"') do (set "C_NAME=%%~a")
if defined C_NAME (
    set C_NAME=%C_NAME:'=%
    set C_NAME=%C_NAME:"=%
)
set "DB_PATH=%RUN_DATA_DIR%\%C_NAME%.db"

if %COUNT% gtr 1 (
    call "%ANALYZE_SCRIPT%" "%S_PATH_DDL%" "%DB_PATH%" "%C_PATH%" "%REPORT_CSV%" --append
) else (
    call "%ANALYZE_SCRIPT%" "%S_PATH_DDL%" "%DB_PATH%" "%C_PATH%" "%REPORT_CSV%"
)
exit /b

:after_analysis

rem 4. Generate heatmaps directly inside the runs pack
echo 📊 [4/5] Plotting 24h Individual heatmaps directly inside runs pack...
python3 "%PROJECT_ROOT%\tools\visualization\analyze\run_massive_forecast_plots.py" "%RUN_DIR%"

rem 5. Plot Master report directly inside the data/ folder of runs pack
echo 🎨 [5/5] Generating Master Heatmap Report directly inside data/ folder...
python3 "%PROJECT_ROOT%\tools\visualization\analyze\plot_research_report.py" "%REPORT_CSV%"

rem Sync results to artifact path for system integration
copy /y "%RUN_DATA_DIR%\research_results_analysis.png" "\home\gyeongho\.gemini\antigravity-cli\brain\fc61e4a8-7c53-4959-8257-7d147633cdc3\" >nul 2>&1

echo ==================================================
echo ✅ [L2 SUCCESS] Scenario Pipeline Complete!
echo 📍 Master Chart: %RUN_DATA_DIR%\research_results_analysis.png
echo ==================================================
exit /b 0
