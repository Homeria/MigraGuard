# MigraGuard 전체 DDL 분석 테스트 스크립트 (Windows PowerShell)

$DB_URL = "postgres://user:pass@db:5432/target_db"
$SQLITE_PATH = "/app/data/migraguard.db"
$MIGRATIONS_DIR = "migrations"

Write-Host "🚀 모든 DDL 파일에 대한 리스크 분석을 시작합니다..." -ForegroundColor Cyan

Get-ChildItem "$MIGRATIONS_DIR/*.sql" | ForEach-Object {
    $fileName = $_.Name
    $containerPath = "/app/code/migrations/$fileName"
    
    Write-Host "`n[파일 분석 중: $fileName]" -ForegroundColor Yellow
    
    # 하나의 컨테이너 안에서 analyze 명령 실행
    docker compose run --rm analyze-shell analyze $containerPath --db $DB_URL --sqlite $SQLITE_PATH
}

Write-Host "`n✅ 모든 분석이 완료되었습니다." -ForegroundColor Green
