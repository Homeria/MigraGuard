@echo off
setlocal enabledelayedexpansion
if not exist experiments\reports\metrics mkdir experiments\reports\metrics
for %%f in (experiments\scenarios\*.yaml) do (
    echo 🔄 Exporting CSV for: %%~nxf
    call %~dp0gen-csv-from-scenario.bat %%f experiments\reports\metrics\%%~nf_metrics.csv %1
)
echo ✅ All metrics exported to experiments\reports\metrics\
