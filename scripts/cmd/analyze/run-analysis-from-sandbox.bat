@echo off
if "%~1"=="" echo Usage: %0 ^<db_name^> ^<sql_name^> [output_type] & exit /b 1
set D=experiments\data\%~1
if not exist !D! set D=%~1
set Q=experiments\ddl\%~2
if not exist !Q! set Q=%~2
set O=%~3
if "!O!"=="" set O=console
go run ./cmd/migraguard analyze !Q! --sandbox !D! --output !O!
