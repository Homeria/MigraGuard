@echo off
setlocal enabledelayedexpansion
if "%~1"=="" echo Usage: %0 ^<sql_name^> [db_url] [sqlite_path] [output_type] & exit /b 1
set Q=experiments\ddl\%~1
if not exist !Q! set Q=%~1
set DB_URL=%~2
set SQLITE=%~3
set OUT_TYPE=%~4
if "!OUT_TYPE!"=="" set OUT_TYPE=console

set ARGS=analyze !Q!
if not "!DB_URL!"=="" set ARGS=!ARGS! --db !DB_URL!
if not "!SQLITE!"=="" set ARGS=!ARGS! --sqlite !SQLITE!
if not "!OUT_TYPE!"=="console" set ARGS=!ARGS! --output !OUT_TYPE!

go run ./cmd/migraguard !ARGS!
