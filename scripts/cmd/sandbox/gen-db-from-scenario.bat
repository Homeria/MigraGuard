@echo off
setlocal enabledelayedexpansion
if "%~1"=="" echo Usage: %0 ^<scenario_name^> [--force] & exit /b 1
set S=experiments\scenarios\%~1
if not exist !S! set S=%~1
go run ./cmd/migraguard simulate --scenario !S! %2
for /f "tokens=2 delims=: " %%a in ('findstr "experiment_name:" !S!') do (
    set N=%%a
    set N=!N:'=!
    set N=!N:"=!
)
if exist !N!.db move /y !N!.db experiments\data\ >nul & echo [OK] Generated: experiments\data\!N!.db
