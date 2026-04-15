@echo off
setlocal enabledelayedexpansion

:: 터미널 인코딩을 UTF-8로 설정 (한글 깨짐 방지)
chcp 65001 > nul

:: MigraGuard All DDL Test Script (Windows CMD/Batch)

set DB_URL=postgres://user:pass@db:5432/target_db
set SQLITE_PATH=/app/data/migraguard.db
set MIGRATIONS_DIR=migrations

echo 🚀 모든 DDL 파일에 대한 리스크 분석을 시작합니다...

for %%f in (%MIGRATIONS_DIR%\*.sql) do (
    set FILE_NAME=%%~nxf
    echo.
    echo [파일 분석 중: !FILE_NAME!]
    
    :: 도커 명령 실행
    docker compose run --rm analyze-shell analyze /app/code/migrations/!FILE_NAME! --db %DB_URL% --sqlite %SQLITE_PATH%
)

echo.
echo ✅ 모든 분석이 완료되었습니다.
pause
