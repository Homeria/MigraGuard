@echo off
REM 🛡️ MigraGuard: Predictive 24-Hour Forecast Workflow
REM Usage: .\scripts\cmd\analyze\run-predictive-forecast.bat <db_path> <ddl_path>

if "%~1"=="" goto usage
if "%~2"=="" goto usage

set DB_PATH=%1
set DDL_PATH=%2
set CSV_NAME=predictive_forecast.csv
set IMG_NAME=predictive_risk_heatmap.png

echo --------------------------------------------------------
echo [1/2] Running Predictive Risk Analysis (Go Engine)...
echo --------------------------------------------------------
go run ./cmd/migraguard analyze %DDL_PATH% --sandbox %DB_PATH% --forecast --output console

if %ERRORLEVEL% neq 0 (
    echo [ERROR] Analysis failed. Skipping visualization.
    exit /b %ERRORLEVEL%
)

echo.
echo --------------------------------------------------------
echo [2/2] Generating Risk Heatmap Visualization (Python)...
echo --------------------------------------------------------
if not exist %CSV_NAME% (
    echo [ERROR] Forecast CSV (%CSV_NAME%) was not generated.
    exit /b 1
)

python ./tools/visualization/analyze/plot_predictive_heatmap.py %CSV_NAME%

if %ERRORLEVEL% equ 0 (
    echo [SUCCESS] Analysis complete!
    echo Report Location: %IMG_NAME%
) else (
    echo [ERROR] Visualization failed.
    exit /b %ERRORLEVEL%
)
goto :eof

:usage
echo Usage: %0 ^<db_path^> ^<ddl_path^>
echo Example: %0 experiments\data\exp_01_steady_normal.db experiments\ddl\011_danger_rewrite_order_no.sql
exit /b 1
